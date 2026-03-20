package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stripe/stripe-go/v76"
	"github.com/yourusername/clawhost/hosting-service/billing"
	"github.com/yourusername/clawhost/hosting-service/models"
	"gorm.io/gorm"
)

type Server struct {
	router              *gin.Engine
	logger              *slog.Logger
	db                  *gorm.DB
	billing             *billing.BillingService
	stripeWebhookSecret string
	jwtSecret           string
	coreAPIURL          string
	allowDevAuth        bool
}

func NewServer(db *gorm.DB, logger *slog.Logger, stripeKey, stripeWebhookSecret, jwtSecret, coreAPIURL string, allowDevAuth bool) *Server {
	if logger == nil {
		logger = slog.Default()
	}

	s := &Server{
		router:              gin.New(),
		logger:              logger,
		db:                  db,
		stripeWebhookSecret: strings.TrimSpace(stripeWebhookSecret),
		jwtSecret:           strings.TrimSpace(jwtSecret),
		coreAPIURL:          strings.TrimRight(strings.TrimSpace(coreAPIURL), "/"),
		allowDevAuth:        allowDevAuth,
	}

	s.router.Use(gin.Recovery())

	if strings.TrimSpace(stripeKey) != "" {
		s.billing = billing.NewBillingService(stripeKey)
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "service": "clawhost-hosting-service", "version": "1.0.0"})
	})

	if s.allowDevAuth {
		s.router.POST("/api/v1/auth/dev-token", s.createDevToken)
	}

	s.router.POST("/webhooks/stripe", s.handleStripeWebhook)

	v1 := s.router.Group("/api/v1")
	v1.Use(s.jwtAuthMiddleware())
	{
		v1.GET("/dashboard", s.getDashboard)

		v1.GET("/instances", s.listInstances)
		v1.POST("/instances", s.createInstance)
		v1.GET("/instances/:id", s.getInstance)
		v1.PUT("/instances/:id/config", s.updateInstanceConfig)
		v1.POST("/instances/:id/restart", s.restartInstance)
		v1.GET("/instances/:id/provision-status", s.getInstanceProvisionStatus)

		v1.POST("/support/tickets", s.createTicket)
		v1.GET("/support/tickets", s.getCustomerTickets)
		v1.GET("/support/tickets/:id", s.getTicket)
		v1.POST("/support/tickets/:id/messages", s.addTicketMessage)
		v1.GET("/support/stats", s.ticketStats)

		v1.POST("/billing/customer", s.createOrUpdateBillingCustomer)
		v1.POST("/billing/subscriptions", s.createSubscription)
		v1.GET("/billing/subscriptions/:id", s.getSubscription)
		v1.POST("/billing/subscriptions/:id/cancel", s.cancelSubscription)
	}
}

func (s *Server) Start(port string) error {
	if strings.TrimSpace(port) == "" {
		port = "8090"
	}
	s.logger.Info("hosting_service_starting", slog.String("port", port))
	return s.router.Run(":" + port)
}

func (s *Server) jwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if s.jwtSecret == "" {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "auth not configured (missing HOSTING_JWT_SECRET)"})
			return
		}

		auth := strings.TrimSpace(c.GetHeader("Authorization"))
		if !strings.HasPrefix(auth, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}

		tokenString := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(s.jwtSecret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			return
		}

		customerID, parseErr := parseCustomerIDClaim(claims["customer_id"])
		if parseErr != nil || customerID == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing customer_id claim"})
			return
		}

		c.Set("customer_id", customerID)
		c.Next()
	}
}

func parseCustomerIDClaim(raw any) (uint, error) {
	switch typed := raw.(type) {
	case float64:
		if typed <= 0 {
			return 0, errors.New("invalid id")
		}
		return uint(typed), nil
	case string:
		n, err := strconv.ParseUint(strings.TrimSpace(typed), 10, 64)
		if err != nil || n == 0 {
			return 0, errors.New("invalid id")
		}
		return uint(n), nil
	default:
		return 0, errors.New("unsupported claim type")
	}
}

type devTokenRequest struct {
	Email       string `json:"email"`
	Name        string `json:"name"`
	CompanyName string `json:"company_name"`
}

func (s *Server) createDevToken(c *gin.Context) {
	if !s.allowDevAuth {
		c.JSON(http.StatusForbidden, gin.H{"error": "dev auth disabled"})
		return
	}
	if s.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not configured"})
		return
	}
	if s.jwtSecret == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "auth not configured (missing HOSTING_JWT_SECRET)"})
		return
	}

	var req devTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}
	if strings.TrimSpace(req.Email) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}

	customer := models.Customer{}
	if err := s.db.Where("email = ?", strings.TrimSpace(req.Email)).First(&customer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			customer = models.Customer{Email: strings.TrimSpace(req.Email), Name: strings.TrimSpace(req.Name), CompanyName: strings.TrimSpace(req.CompanyName)}
			if createErr := s.db.Create(&customer).Error; createErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": createErr.Error()})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	claims := jwt.MapClaims{
		"customer_id": customer.ID,
		"email":       customer.Email,
		"exp":         time.Now().Add(24 * time.Hour).Unix(),
		"iat":         time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString, "customer": customer})
}

func mustCustomerID(c *gin.Context) uint {
	raw, _ := c.Get("customer_id")
	if id, ok := raw.(uint); ok {
		return id
	}
	return 0
}

func (s *Server) getDashboard(c *gin.Context) {
	if s.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not configured"})
		return
	}
	customerID := mustCustomerID(c)

	var totalInstances int64
	var runningInstances int64
	var openTickets int64
	var subscriptions int64

	_ = s.db.Model(&models.Instance{}).Where("customer_id = ?", customerID).Count(&totalInstances).Error
	_ = s.db.Model(&models.Instance{}).Where("customer_id = ? AND status = ?", customerID, "running").Count(&runningInstances).Error
	_ = s.db.Model(&models.SupportTicket{}).Where("customer_id = ? AND status IN ?", customerID, []string{"open", "in_progress", "waiting_customer"}).Count(&openTickets).Error
	_ = s.db.Model(&models.Subscription{}).Where("customer_id = ?", customerID).Count(&subscriptions).Error

	var instances []models.Instance
	_ = s.db.Where("customer_id = ?", customerID).Order("updated_at DESC").Limit(10).Find(&instances).Error

	c.JSON(http.StatusOK, gin.H{
		"overview": gin.H{
			"total_instances":  totalInstances,
			"active_instances": runningInstances,
			"open_tickets":     openTickets,
			"subscriptions":    subscriptions,
		},
		"instances": instances,
	})
}

type createInstanceRequest struct {
	Name        string `json:"name"`
	Provider    string `json:"provider"`
	ServerType  string `json:"server_type"`
	Location    string `json:"location"`
	OpenClawURL string `json:"openclaw_url"`
}

func (s *Server) listInstances(c *gin.Context) {
	if s.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not configured"})
		return
	}
	customerID := mustCustomerID(c)
	var instances []models.Instance
	if err := s.db.Where("customer_id = ?", customerID).Order("updated_at DESC").Find(&instances).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"instances": instances})
}

func (s *Server) createInstance(c *gin.Context) {
	if s.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not configured"})
		return
	}
	customerID := mustCustomerID(c)

	var req createInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	jobID, err := s.requestProvisionFromCore(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	instance := models.Instance{
		CustomerID:    customerID,
		Name:          strings.TrimSpace(req.Name),
		Status:        "provisioning",
		Provider:      strings.TrimSpace(req.Provider),
		ServerType:    strings.TrimSpace(req.ServerType),
		Location:      strings.TrimSpace(req.Location),
		OpenClawURL:   strings.TrimSpace(req.OpenClawURL),
		ExternalJobID: jobID,
	}

	if err := s.db.Create(&instance).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"instance": instance, "provision_job_id": jobID})
}

func (s *Server) requestProvisionFromCore(req createInstanceRequest) (string, error) {
	if s.coreAPIURL == "" {
		return "", errors.New("CORE_API_URL is not configured")
	}

	payload := map[string]any{
		"provider":    req.Provider,
		"server_type": req.ServerType,
		"location":    req.Location,
		"openclaw_config": map[string]any{
			"openclaw_url": req.OpenClawURL,
		},
	}
	body, _ := json.Marshal(payload)

	endpoint := s.coreAPIURL + "/api/v1/provision"
	httpReq, _ := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("core provision request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("core provision request returned %d: %s", resp.StatusCode, string(respBody))
	}

	var parsed map[string]any
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("invalid core response: %w", err)
	}

	jobID, _ := parsed["job_id"].(string)
	if strings.TrimSpace(jobID) == "" {
		return "", errors.New("core response missing job_id")
	}
	return jobID, nil
}

func (s *Server) getInstance(c *gin.Context) {
	if s.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not configured"})
		return
	}

	instanceID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || instanceID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid instance id"})
		return
	}

	customerID := mustCustomerID(c)
	instance := models.Instance{}
	if err := s.db.Where("id = ? AND customer_id = ?", uint(instanceID), customerID).First(&instance).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"instance": instance})
}

type updateConfigRequest struct {
	Configuration map[string]any `json:"configuration"`
	OpenClawURL   string         `json:"openclaw_url"`
}

func (s *Server) updateInstanceConfig(c *gin.Context) {
	if s.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not configured"})
		return
	}

	instanceID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || instanceID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid instance id"})
		return
	}

	var req updateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	customerID := mustCustomerID(c)
	instance := models.Instance{}
	if err := s.db.Where("id = ? AND customer_id = ?", uint(instanceID), customerID).First(&instance).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
		return
	}

	if req.Configuration != nil {
		cfg, _ := json.Marshal(req.Configuration)
		instance.Configuration = string(cfg)
	}
	if strings.TrimSpace(req.OpenClawURL) != "" {
		instance.OpenClawURL = strings.TrimSpace(req.OpenClawURL)
	}

	if err := s.db.Save(&instance).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"instance": instance, "message": "configuration updated"})
}

func (s *Server) restartInstance(c *gin.Context) {
	if s.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not configured"})
		return
	}

	instanceID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || instanceID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid instance id"})
		return
	}

	customerID := mustCustomerID(c)
	instance := models.Instance{}
	if err := s.db.Where("id = ? AND customer_id = ?", uint(instanceID), customerID).First(&instance).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
		return
	}

	instance.Status = "restarting"
	if err := s.db.Save(&instance).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "restart requested", "instance": instance})
}

func (s *Server) getInstanceProvisionStatus(c *gin.Context) {
	if s.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not configured"})
		return
	}

	instanceID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || instanceID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid instance id"})
		return
	}

	customerID := mustCustomerID(c)
	instance := models.Instance{}
	if err := s.db.Where("id = ? AND customer_id = ?", uint(instanceID), customerID).First(&instance).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
		return
	}

	if strings.TrimSpace(instance.ExternalJobID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "instance has no external provisioning job"})
		return
	}
	if s.coreAPIURL == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "CORE_API_URL is not configured"})
		return
	}

	endpoint := fmt.Sprintf("%s/api/v1/provision/status/%s", s.coreAPIURL, instance.ExternalJobID)
	resp, reqErr := (&http.Client{Timeout: 10 * time.Second}).Get(endpoint)
	if reqErr != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": reqErr.Error()})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= http.StatusBadRequest {
		c.JSON(http.StatusBadGateway, gin.H{"error": string(body)})
		return
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err == nil {
		if status, ok := payload["status"].(string); ok && strings.TrimSpace(status) != "" {
			instance.Status = status
			_ = s.db.Save(&instance).Error
		}
	}

	c.Data(http.StatusOK, "application/json", body)
}

type createTicketRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Priority    string `json:"priority"`
}

func (s *Server) createTicket(c *gin.Context) {
	if s.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not configured"})
		return
	}

	customerID := mustCustomerID(c)
	var req createTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	ticket := models.SupportTicket{
		CustomerID:  customerID,
		Title:       strings.TrimSpace(req.Title),
		Description: strings.TrimSpace(req.Description),
		Category:    strings.TrimSpace(req.Category),
		Priority:    strings.TrimSpace(req.Priority),
		Status:      "open",
	}
	if ticket.Priority == "" {
		ticket.Priority = "medium"
	}

	if err := s.db.Create(&ticket).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ticket)
}

func (s *Server) getCustomerTickets(c *gin.Context) {
	if s.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not configured"})
		return
	}

	customerID := mustCustomerID(c)
	var tickets []models.SupportTicket
	if err := s.db.Where("customer_id = ?", customerID).Order("updated_at DESC").Find(&tickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tickets": tickets})
}

func (s *Server) getTicket(c *gin.Context) {
	if s.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not configured"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket id"})
		return
	}

	customerID := mustCustomerID(c)
	ticket := models.SupportTicket{}
	if err := s.db.Preload("Messages").Where("id = ? AND customer_id = ?", uint(id), customerID).First(&ticket).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ticket)
}

type addMessageRequest struct {
	Message string `json:"message"`
}

func (s *Server) addTicketMessage(c *gin.Context) {
	if s.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not configured"})
		return
	}

	ticketID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || ticketID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket id"})
		return
	}

	var req addMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	customerID := mustCustomerID(c)
	ticket := models.SupportTicket{}
	if err := s.db.Where("id = ? AND customer_id = ?", uint(ticketID), customerID).First(&ticket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
		return
	}

	msg := models.SupportMessage{
		TicketID:   uint(ticketID),
		AuthorID:   customerID,
		AuthorType: "customer",
		Message:    strings.TrimSpace(req.Message),
		IsInternal: false,
	}
	if err := s.db.Create(&msg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if ticket.Status == "waiting_customer" {
		ticket.Status = "in_progress"
		_ = s.db.Save(&ticket).Error
	}

	c.JSON(http.StatusCreated, msg)
}

func (s *Server) ticketStats(c *gin.Context) {
	if s.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not configured"})
		return
	}

	customerID := mustCustomerID(c)
	var total int64
	var open int64
	var inProgress int64
	var resolved int64
	var closed int64

	_ = s.db.Model(&models.SupportTicket{}).Where("customer_id = ?", customerID).Count(&total).Error
	_ = s.db.Model(&models.SupportTicket{}).Where("customer_id = ? AND status = ?", customerID, "open").Count(&open).Error
	_ = s.db.Model(&models.SupportTicket{}).Where("customer_id = ? AND status = ?", customerID, "in_progress").Count(&inProgress).Error
	_ = s.db.Model(&models.SupportTicket{}).Where("customer_id = ? AND status = ?", customerID, "resolved").Count(&resolved).Error
	_ = s.db.Model(&models.SupportTicket{}).Where("customer_id = ? AND status = ?", customerID, "closed").Count(&closed).Error

	stats := map[string]int64{"total": total, "open": open, "in_progress": inProgress, "resolved": resolved, "closed": closed}

	c.JSON(http.StatusOK, stats)
}

type createBillingCustomerRequest struct {
	Email       string `json:"email"`
	Name        string `json:"name"`
	CompanyName string `json:"company_name"`
}

func (s *Server) createOrUpdateBillingCustomer(c *gin.Context) {
	if s.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not configured"})
		return
	}
	if s.billing == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "billing not configured (missing STRIPE_SECRET_KEY)"})
		return
	}

	var req createBillingCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	customerID := mustCustomerID(c)
	customerModel := models.Customer{}
	if err := s.db.First(&customerModel, customerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "customer not found"})
		return
	}

	if strings.TrimSpace(req.Email) != "" {
		customerModel.Email = strings.TrimSpace(req.Email)
	}
	if strings.TrimSpace(req.Name) != "" {
		customerModel.Name = strings.TrimSpace(req.Name)
	}
	if strings.TrimSpace(req.CompanyName) != "" {
		customerModel.CompanyName = strings.TrimSpace(req.CompanyName)
	}

	stripeCustomer, err := s.billing.CreateCustomer(billing.CustomerData{
		Email:       customerModel.Email,
		Name:        customerModel.Name,
		CompanyName: customerModel.CompanyName,
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	customerModel.StripeCustomerID = stripeCustomer.ID
	if err := s.db.Save(&customerModel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"customer": customerModel, "stripe_customer_id": stripeCustomer.ID})
}

type createSubscriptionRequest struct {
	PriceID  string `json:"price_id"`
	PlanName string `json:"plan_name"`
}

func (s *Server) createSubscription(c *gin.Context) {
	if s.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not configured"})
		return
	}
	if s.billing == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "billing not configured (missing STRIPE_SECRET_KEY)"})
		return
	}

	var req createSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	customerID := mustCustomerID(c)
	customerModel := models.Customer{}
	if err := s.db.First(&customerModel, customerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "customer not found"})
		return
	}
	if strings.TrimSpace(customerModel.StripeCustomerID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "customer has no stripe_customer_id; call /billing/customer first"})
		return
	}

	priceID := strings.TrimSpace(req.PriceID)
	if priceID == "" && strings.TrimSpace(req.PlanName) != "" {
		priceID = s.billing.GetPriceIDForPlan(req.PlanName)
	}
	if strings.TrimSpace(priceID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "price_id or known plan_name is required"})
		return
	}

	sub, err := s.billing.CreateSubscription(billing.SubscriptionData{
		CustomerID: customerModel.StripeCustomerID,
		PriceID:    priceID,
		PlanName:   req.PlanName,
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	stored := models.Subscription{
		CustomerID:           customerID,
		PlanName:             req.PlanName,
		Status:               string(sub.Status),
		StripeSubscriptionID: sub.ID,
		StripePriceID:        priceID,
	}
	if sub.CurrentPeriodEnd > 0 {
		stored.CurrentPeriodEnd = time.Unix(sub.CurrentPeriodEnd, 0).UTC()
	}
	if err := s.db.Create(&stored).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"subscription": stored, "stripe_subscription": sub})
}

func (s *Server) getSubscription(c *gin.Context) {
	if s.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not configured"})
		return
	}

	customerID := mustCustomerID(c)
	id := strings.TrimSpace(c.Param("id"))
	stored := models.Subscription{}
	if err := s.db.Where("customer_id = ? AND stripe_subscription_id = ?", customerID, id).First(&stored).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"subscription": stored})
}

func (s *Server) cancelSubscription(c *gin.Context) {
	if s.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not configured"})
		return
	}
	if s.billing == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "billing not configured (missing STRIPE_SECRET_KEY)"})
		return
	}

	customerID := mustCustomerID(c)
	id := strings.TrimSpace(c.Param("id"))
	stored := models.Subscription{}
	if err := s.db.Where("customer_id = ? AND stripe_subscription_id = ?", customerID, id).First(&stored).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}

	sub, err := s.billing.CancelSubscription(id)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	stored.Status = string(sub.Status)
	if err := s.db.Save(&stored).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"subscription": stored})
}

func (s *Server) handleStripeWebhook(c *gin.Context) {
	if s.billing == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "billing not configured"})
		return
	}
	if s.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not configured"})
		return
	}
	if s.stripeWebhookSecret == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "STRIPE_WEBHOOK_SECRET not configured"})
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	sig := c.GetHeader("Stripe-Signature")
	event, err := s.billing.ConstructWebhookEvent(body, sig, s.stripeWebhookSecret)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	switch event.Type {
	case "customer.subscription.created", "customer.subscription.updated", "customer.subscription.deleted":
		var sub stripe.Subscription
		if unmarshalErr := json.Unmarshal(event.Data.Raw, &sub); unmarshalErr == nil {
			stored := models.Subscription{}
			if findErr := s.db.Where("stripe_subscription_id = ?", sub.ID).First(&stored).Error; findErr == nil {
				stored.Status = string(sub.Status)
				if sub.CurrentPeriodEnd > 0 {
					stored.CurrentPeriodEnd = time.Unix(sub.CurrentPeriodEnd, 0).UTC()
				}
				_ = s.db.Save(&stored).Error
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"received": true})
}

func (s *Server) DebugRoutes() []string {
	out := make([]string, 0, len(s.router.Routes()))
	for _, route := range s.router.Routes() {
		out = append(out, fmt.Sprintf("%s %s", route.Method, route.Path))
	}
	return out
}

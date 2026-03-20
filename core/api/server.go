package api

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ProvisionRequest is the payload for POST /api/v1/provision
type ProvisionRequest struct {
	Provider       string `json:"provider"` // "local" | "digitalocean" | "hetzner"
	OpenClawRepo   string `json:"openclaw_repo_url"`
	DBPassword     string `json:"db_password"`
	WebUISecretKey string `json:"webui_secret_key"`
	OpenAIAPIKey   string `json:"openai_api_key"`
	OpenAIBaseURL  string `json:"openai_api_base_url"`
}

// ProvisionJob tracks the state of an async provisioning job
type ProvisionJob struct {
	ID          string
	Status      string // "running" | "completed" | "failed"
	CurrentStep string
	Progress    int
	Error       string
	Log         string
	StartedAt   time.Time
	CompletedAt *time.Time
}

// CoreServer represents the core ClawHost API server
type CoreServer struct {
	router *gin.Engine
	port   string
	logger *slog.Logger

	startedAt      time.Time
	lastRestartAt  time.Time
	activityMu     sync.RWMutex
	recentActivity []string

	jobsMu sync.RWMutex
	jobs   map[string]*ProvisionJob
}

func NewRouter(logger *slog.Logger) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(requestLoggerMiddleware(logger))
	return router
}

func requestLoggerMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		latency := time.Since(start)
		logger.Info("http_request",
			slog.String("method", c.Request.Method),
			slog.String("path", c.FullPath()),
			slog.String("raw_path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("latency", latency),
			slog.String("client_ip", c.ClientIP()),
		)
	}
}

// NewCoreServer creates a new core API server instance
func NewCoreServer(router *gin.Engine, logger *slog.Logger, port string) *CoreServer {
	if port == "" {
		port = "8080"
	}
	if logger == nil {
		logger = slog.Default()
	}
	if router == nil {
		router = NewRouter(logger)
	}

	return &CoreServer{
		router:         router,
		port:           port,
		logger:         logger,
		startedAt:      time.Now().UTC(),
		recentActivity: make([]string, 0, 50),
		jobs:           make(map[string]*ProvisionJob),
	}
}

// SetupRoutes configures the core API routes
func (s *CoreServer) SetupRoutes() {
	// Local dashboard UI
	s.router.GET("/", s.renderLocalDashboard)
	s.router.GET("/dashboard", s.renderLocalDashboard)
	s.router.POST("/dashboard/actions/restart", s.restartServicesAction)

	// Health check endpoint
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "clawhost-core",
			"version": "1.0.0",
		})
	})

	// Core API routes
	v1 := s.router.Group("/api/v1")
	{
		// Instance management (open source core)
		v1.GET("/instances/:id/status", s.getInstanceStatus)
		v1.GET("/instances/:id/metrics", s.getInstanceMetrics)
		v1.GET("/instances/:id/health", s.getInstanceHealth)
		v1.GET("/instances/:id/logs", s.getInstanceLogs)

		// Provisioning
		v1.POST("/provision", s.provisionInstance)
		v1.GET("/provision/status/:job_id", s.getProvisioningStatus)
	}

	s.addActivity("System started")
}

// Instance status endpoint
func (s *CoreServer) getInstanceStatus(c *gin.Context) {
	instanceID := c.Param("id")

	// Mock response - in reality, this would check actual instance status
	c.JSON(http.StatusOK, gin.H{
		"instance_id": instanceID,
		"status":      "running",
		"uptime":      "2h 15m",
		"version":     "openclaw-1.0.0",
	})
}

// Instance metrics endpoint
func (s *CoreServer) getInstanceMetrics(c *gin.Context) {
	instanceID := c.Param("id")

	// Mock metrics response
	c.JSON(http.StatusOK, gin.H{
		"instance_id":  instanceID,
		"cpu_usage":    45.2,
		"memory_usage": 67.8,
		"disk_usage":   23.1,
		"network_in":   1024,
		"network_out":  512,
		"timestamp":    "2026-03-18T10:30:00Z",
	})
}

// Instance health endpoint
func (s *CoreServer) getInstanceHealth(c *gin.Context) {
	instanceID := c.Param("id")

	// Mock health response
	c.JSON(http.StatusOK, gin.H{
		"instance_id": instanceID,
		"healthy":     true,
		"last_check":  "2026-03-18T10:30:00Z",
		"checks": gin.H{
			"http_endpoint":    true,
			"database":         true,
			"disk_space":       true,
			"memory_available": true,
		},
	})
}

// Instance logs endpoint
func (s *CoreServer) getInstanceLogs(c *gin.Context) {
	instanceID := c.Param("id")
	limit := c.DefaultQuery("limit", "100")

	// Mock logs response
	c.JSON(http.StatusOK, gin.H{
		"instance_id": instanceID,
		"limit":       limit,
		"logs": []gin.H{
			{
				"timestamp": "2026-03-18T10:30:00Z",
				"level":     "info",
				"message":   "OpenClaw instance started",
				"source":    "openclaw",
			},
			{
				"timestamp": "2026-03-18T10:29:45Z",
				"level":     "info",
				"message":   "Database connection established",
				"source":    "system",
			},
		},
	})
}

// Provisioning status endpoint
func (s *CoreServer) getProvisioningStatus(c *gin.Context) {
	jobID := c.Param("job_id")

	s.jobsMu.RLock()
	job, ok := s.jobs[jobID]
	s.jobsMu.RUnlock()

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}

	res := gin.H{
		"job_id":       job.ID,
		"status":       job.Status,
		"progress":     job.Progress,
		"current_step": job.CurrentStep,
		"started_at":   job.StartedAt,
	}
	if job.CompletedAt != nil {
		res["completed_at"] = job.CompletedAt
	}
	if job.Error != "" {
		res["error"] = job.Error
	}
	if job.Log != "" {
		res["log"] = job.Log
	}
	c.JSON(http.StatusOK, res)
}

// provisionInstance handles POST /api/v1/provision
func (s *CoreServer) provisionInstance(c *gin.Context) {
	var req ProvisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	if req.Provider == "" {
		req.Provider = "local"
	}

	if req.OpenClawRepo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "openclaw_repo_url is required"})
		return
	}
	if req.DBPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "db_password is required"})
		return
	}
	if req.WebUISecretKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "webui_secret_key is required"})
		return
	}

	if req.Provider != "local" {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("provider %q not yet supported; use \"local\"", req.Provider)})
		return
	}

	jobID := fmt.Sprintf("job-%d", time.Now().UnixNano())
	job := &ProvisionJob{
		ID:          jobID,
		Status:      "running",
		CurrentStep: "starting",
		Progress:    0,
		StartedAt:   time.Now().UTC(),
	}
	s.jobsMu.Lock()
	s.jobs[jobID] = job
	s.jobsMu.Unlock()

	s.addActivity(fmt.Sprintf("Provision started: %s (local)", jobID))
	go s.runLocalProvision(job, req)

	c.JSON(http.StatusAccepted, gin.H{
		"job_id":     jobID,
		"status":     "running",
		"status_url": "/api/v1/provision/status/" + jobID,
	})
}

// runLocalProvision installs OpenClaw on the same machine via docker compose
func (s *CoreServer) runLocalProvision(job *ProvisionJob, req ProvisionRequest) {
	installDir := "/opt/clawhost/app"
	openclawiDir := "/opt/openclaw"

	setStep := func(step string, progress int) {
		s.jobsMu.Lock()
		job.CurrentStep = step
		job.Progress = progress
		s.jobsMu.Unlock()
		s.logger.Info("provision_step", slog.String("job", job.ID), slog.String("step", step))
	}

	fail := func(err string) {
		now := time.Now().UTC()
		s.jobsMu.Lock()
		job.Status = "failed"
		job.Error = err
		job.CompletedAt = &now
		s.jobsMu.Unlock()
		s.logger.Error("provision_failed", slog.String("job", job.ID), slog.String("error", err))
	}

	appendLog := func(line string) {
		s.jobsMu.Lock()
		job.Log += line + "\n"
		s.jobsMu.Unlock()
	}

	run := func(name string, args ...string) error {
		cmd := exec.Command(name, args...)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		err := cmd.Run()
		appendLog(out.String())
		return err
	}

	// Step 1: clone or update OpenClaw source
	setStep("cloning openclaw repo", 10)
	if _, err := os.Stat(openclawiDir + "/.git"); err == nil {
		if err := run("git", "-C", openclawiDir, "pull"); err != nil {
			fail("git pull failed: " + err.Error())
			return
		}
	} else {
		if err := run("git", "clone", req.OpenClawRepo, openclawiDir); err != nil {
			fail("git clone failed: " + err.Error())
			return
		}
	}

	// Step 2: build Docker image
	setStep("building docker image (this may take a few minutes)", 30)
	if err := run("docker", "build", "-t", "openclaw:latest", openclawiDir); err != nil {
		fail("docker build failed: " + err.Error())
		return
	}

	// Step 3: write .env file
	setStep("writing environment config", 70)
	envContent := fmt.Sprintf(
		"DB_PASSWORD=%s\nWEBUI_SECRET_KEY=%s\nOPENAI_API_KEY=%s\nOPENAI_API_BASE_URL=%s\n",
		req.DBPassword,
		req.WebUISecretKey,
		req.OpenAIAPIKey,
		func() string {
			if req.OpenAIBaseURL != "" {
				return req.OpenAIBaseURL
			}
			return "https://api.openai.com/v1"
		}(),
	)
	if err := os.WriteFile(installDir+"/.env", []byte(envContent), 0600); err != nil {
		fail("writing .env failed: " + err.Error())
		return
	}

	// Step 4: docker compose up
	setStep("starting openclaw stack", 85)
	cmd := exec.Command(
		"docker", "compose",
		"-f", installDir+"/infra/docker/openclaw.yml",
		"--env-file", installDir+"/.env",
		"up", "-d",
	)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		appendLog(out.String())
		fail("docker compose up failed: " + err.Error())
		return
	}
	appendLog(out.String())

	now := time.Now().UTC()
	s.jobsMu.Lock()
	job.Status = "completed"
	job.CurrentStep = "OpenClaw is running"
	job.Progress = 100
	job.CompletedAt = &now
	s.jobsMu.Unlock()
	s.addActivity(fmt.Sprintf("Provision completed: %s", job.ID))
}

// Start starts the core API server
func (s *CoreServer) Start() error {
	s.SetupRoutes()

	s.logger.Info("core_api_starting", slog.String("port", s.port))
	return s.router.Run(":" + s.port)
}

// GetRouter returns the gin router for testing
func (s *CoreServer) GetRouter() *gin.Engine {
	return s.router
}

func (s *CoreServer) addActivity(message string) {
	s.activityMu.Lock()
	defer s.activityMu.Unlock()

	entry := time.Now().Format("15:04") + " - " + message
	s.recentActivity = append([]string{entry}, s.recentActivity...)
	if len(s.recentActivity) > 25 {
		s.recentActivity = s.recentActivity[:25]
	}
}

func (s *CoreServer) getRecentActivity(limit int) []string {
	s.activityMu.RLock()
	defer s.activityMu.RUnlock()

	if limit <= 0 || limit > len(s.recentActivity) {
		limit = len(s.recentActivity)
	}
	out := make([]string, limit)
	copy(out, s.recentActivity[:limit])
	return out
}

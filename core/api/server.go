package api

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// CoreServer represents the core ClawHost API server
type CoreServer struct {
	router *gin.Engine
	port   string
	logger *slog.Logger

	startedAt      time.Time
	lastRestartAt  time.Time
	activityMu     sync.RWMutex
	recentActivity []string
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

		// Provisioning status (core functionality)
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

	// Mock provisioning status
	c.JSON(http.StatusOK, gin.H{
		"job_id":       jobID,
		"status":       "completed",
		"progress":     100,
		"current_step": "Instance ready",
		"started_at":   "2026-03-18T10:00:00Z",
		"completed_at": "2026-03-18T10:15:00Z",
	})
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

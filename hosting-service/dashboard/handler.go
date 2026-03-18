package dashboard

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DashboardHandler handles customer dashboard requests
type DashboardHandler struct {
	db *gorm.DB
}

func NewDashboardHandler(db *gorm.DB) *DashboardHandler {
	return &DashboardHandler{db: db}
}

// GetDashboard returns dashboard overview for logged-in customer
func (h *DashboardHandler) GetDashboard(c *gin.Context) {
	customerID, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Customer not authenticated",
		})
		return
	}

	// In a real implementation, this would fetch actual data from database
	dashboardData := gin.H{
		"customer_id": customerID,
		"overview": gin.H{
			"total_instances":   2,
			"active_instances":  2,
			"monthly_requests": 45230,
			"uptime":           "99.95%",
		},
		"instances": []gin.H{
			{
				"id":           "inst-1",
				"name":         "Production Bot",
				"status":       "running",
				"url":          "https://prod-bot.clawhost.com",
				"created_at":   "2026-03-01T00:00:00Z",
				"last_activity": "2026-03-18T10:25:00Z",
			},
			{
				"id":           "inst-2",
				"name":         "Development Bot",
				"status":       "running",
				"url":          "https://dev-bot.clawhost.com",
				"created_at":   "2026-03-10T00:00:00Z",
				"last_activity": "2026-03-18T09:15:00Z",
			},
		},
		"recent_activity": []gin.H{
			{
				"timestamp": "2026-03-18T10:25:00Z",
				"type":      "message_processed",
				"instance":  "Production Bot",
				"details":   "Processed WhatsApp message",
			},
			{
				"timestamp": "2026-03-18T10:20:00Z",
				"type":      "instance_restart",
				"instance":  "Development Bot",
				"details":   "Instance restarted successfully",
			},
		},
	}

	c.JSON(http.StatusOK, dashboardData)
}

// GetInstanceDetail returns detailed information about a specific instance
func (h *DashboardHandler) GetInstanceDetail(c *gin.Context) {
	customerID, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Customer not authenticated",
		})
		return
	}

	instanceID := c.Param("id")

	// Mock detailed instance data
	instanceDetail := gin.H{
		"customer_id": customerID,
		"instance": gin.H{
			"id":               instanceID,
			"name":             "Production Bot",
			"status":           "running",
			"url":              "https://prod-bot.clawhost.com",
			"server_type":      "cx21",
			"location":         "Nuremberg",
			"public_ip":        "1.2.3.4",
			"openclaw_version": "1.0.0",
			"created_at":       "2026-03-01T00:00:00Z",
			"last_backup":      "2026-03-18T02:00:00Z",
		},
		"configuration": gin.H{
			"llm_provider":       "openai",
			"llm_model":          "gpt-4",
			"personality_name":   "Customer Support Bot",
			"whatsapp_enabled":   true,
			"telegram_enabled":   false,
			"discord_enabled":    true,
		},
		"metrics": gin.H{
			"cpu_usage":      25.5,
			"memory_usage":   42.3,
			"disk_usage":     18.7,
			"requests_today": 1247,
			"uptime":         "17 days, 3 hours",
		},
	}

	c.JSON(http.StatusOK, instanceDetail)
}

// UpdateInstanceConfig updates instance configuration
func (h *DashboardHandler) UpdateInstanceConfig(c *gin.Context) {
	customerID, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Customer not authenticated",
		})
		return
	}

	instanceID := c.Param("id")

	var configUpdate map[string]interface{}
	if err := c.ShouldBindJSON(&configUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid configuration data",
			"details": err.Error(),
		})
		return
	}

	// In reality, this would update the instance configuration in the database
	// and potentially restart/reconfigure the OpenClaw instance

	c.JSON(http.StatusOK, gin.H{
		"message":     "Configuration updated successfully",
		"customer_id": customerID,
		"instance_id": instanceID,
		"updated_config": configUpdate,
	})
}

// RestartInstance initiates an instance restart
func (h *DashboardHandler) RestartInstance(c *gin.Context) {
	customerID, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Customer not authenticated",
		})
		return
	}

	instanceID := c.Param("id")

	// In reality, this would trigger an actual instance restart
	// For now, we'll simulate it

	c.JSON(http.StatusAccepted, gin.H{
		"message":     "Instance restart initiated",
		"customer_id": customerID,
		"instance_id": instanceID,
		"status":      "restarting",
		"estimated_time": "2-3 minutes",
	})
}
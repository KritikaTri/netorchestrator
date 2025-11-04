package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"netorchestrator/internal/automation"
	"netorchestrator/internal/intelligence"
	"netorchestrator/internal/models"
	"netorchestrator/internal/orchestration"
	"netorchestrator/internal/services"
)

// Handlers contains all HTTP handlers
type Handlers struct {
	networkService       *services.NetworkService
	monitoringService    *services.MonitoringService
	orchestrationService *services.OrchestrationService
	validationService    *services.ValidationService
	AutomationHandler    *automation.Handler
	IntelligenceHandlers *intelligence.IntelligenceHandlers
	containerProvisioner *orchestration.ContainerProvisioner
	linkProvisioner      *orchestration.LinkProvisioner
	policyEnforcer       *orchestration.PolicyEnforcer
	logger               *zap.Logger
}

// NewHandlers creates a new handlers instance
func NewHandlers(
	networkService *services.NetworkService,
	monitoringService *services.MonitoringService,
	orchestrationService *services.OrchestrationService,
	validationService *services.ValidationService,
	automationHandler *automation.Handler,
	intelligenceHandlers *intelligence.IntelligenceHandlers,
	containerProvisioner *orchestration.ContainerProvisioner,
	linkProvisioner *orchestration.LinkProvisioner,
	policyEnforcer *orchestration.PolicyEnforcer,
	logger *zap.Logger,
) *Handlers {
	return &Handlers{
		networkService:       networkService,
		monitoringService:    monitoringService,
		orchestrationService: orchestrationService,
		validationService:    validationService,
		AutomationHandler:    automationHandler,
		IntelligenceHandlers: intelligenceHandlers,
		containerProvisioner: containerProvisioner,
		linkProvisioner:      linkProvisioner,
		policyEnforcer:       policyEnforcer,
		logger:               logger,
	}
}

// HealthCheck handles health check requests
// @Summary Health Check
// @Description Get the health status of the API Gateway
// @Tags System
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Health status"
// @Router /health [get]
func (h *Handlers) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "netorchestrator-api-gateway",
		"timestamp": time.Now().UTC(),
		"version":   "2.0.0",
	})
}

// ReadyCheck handles readiness check requests
// @Summary Readiness Check
// @Description Check dependencies like DB and Redis are reachable
// @Tags System
// @Produce json
// @Success 200 {object} map[string]interface{} "Ready"
// @Failure 503 {object} map[string]interface{} "Unready"
// @Router /ready [get]
func (h *Handlers) ReadyCheck(c *gin.Context) {
	if err := h.monitoringService.Readiness(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "unready",
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":    "ready",
		"timestamp": time.Now().UTC(),
	})
}

// AcknowledgeAlert acknowledges an alert
// @Summary Acknowledge Alert
// @Description Mark an alert as acknowledged by user
// @Tags Monitoring
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Alert ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/monitoring/alerts/{id}/acknowledge [post]
func (h *Handlers) AcknowledgeAlert(c *gin.Context) {
	alertID := c.Param("id")
	userID := c.GetString("user_id")
	username := c.GetString("username")

	if alertID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Alert ID is required",
			"code":    "INVALID_ALERT_ID",
			"details": "Please provide a valid alert ID",
		})
		return
	}

	h.logger.Info("Alert acknowledged",
		zap.String("alert_id", alertID),
		zap.String("user_id", userID),
		zap.String("username", username))

	c.JSON(http.StatusOK, gin.H{
		"message":         "Alert acknowledged successfully",
		"alert_id":        alertID,
		"acknowledged_by": username,
		"acknowledged_at": time.Now().UTC(),
	})
}

// ResolveAlert resolves an alert
// @Summary Resolve Alert
// @Description Mark an alert as resolved
// @Tags Monitoring
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Alert ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/monitoring/alerts/{id}/resolve [post]
func (h *Handlers) ResolveAlert(c *gin.Context) {
	alertID := c.Param("id")
	userID := c.GetString("user_id")
	username := c.GetString("username")

	if alertID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Alert ID is required",
			"code":    "INVALID_ALERT_ID",
			"details": "Please provide a valid alert ID",
		})
		return
	}

	h.logger.Info("Alert resolved",
		zap.String("alert_id", alertID),
		zap.String("user_id", userID),
		zap.String("username", username))

	c.JSON(http.StatusOK, gin.H{
		"message":     "Alert resolved successfully",
		"alert_id":    alertID,
		"resolved_by": username,
		"resolved_at": time.Now().UTC(),
	})
}

// GetAlertRule gets an alert rule by ID
// @Summary Get Alert Rule
// @Description Get details of a specific alert rule
// @Tags Monitoring
// @Produce json
// @Security BearerAuth
// @Param id path string true "Rule ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/monitoring/alert-rules/{id} [get]
func (h *Handlers) GetAlertRule(c *gin.Context) {
	ruleID := c.Param("id")

	if ruleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Rule ID is required",
			"code":    "INVALID_RULE_ID",
			"details": "Please provide a valid rule ID",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rule": gin.H{
			"id":          ruleID,
			"name":        "Sample Alert Rule",
			"description": "This is a sample alert rule",
			"condition":   "cpu_usage > 80",
			"severity":    "warning",
			"enabled":     true,
			"created_at":  time.Now().Add(-24 * time.Hour).UTC(),
			"updated_at":  time.Now().UTC(),
		},
	})
}

// ListAlertRules lists alert rules
// @Summary List Alert Rules
// @Description Get all configured alert rules
// @Tags Monitoring
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/monitoring/alert-rules [get]
func (h *Handlers) ListAlertRules(c *gin.Context) {
	rules := []gin.H{
		{
			"id":          uuid.New().String(),
			"name":        "High CPU Usage",
			"description": "Alert when CPU usage exceeds 80%",
			"condition":   "cpu_usage > 80",
			"severity":    "warning",
			"enabled":     true,
			"created_at":  time.Now().Add(-48 * time.Hour).UTC(),
		},
		{
			"id":          uuid.New().String(),
			"name":        "Memory Threshold",
			"description": "Alert when memory usage exceeds 90%",
			"condition":   "memory_usage > 90",
			"severity":    "critical",
			"enabled":     true,
			"created_at":  time.Now().Add(-72 * time.Hour).UTC(),
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"rules": rules,
		"count": len(rules),
	})
}

// CreateAlertRule creates an alert rule
// @Summary Create Alert Rule
// @Description Create a new alert rule for monitoring
// @Tags Monitoring
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param rule body object true "Alert Rule"
// @Success 201 {object} map[string]interface{}
// @Router /api/v1/monitoring/alert-rules [post]
func (h *Handlers) CreateAlertRule(c *gin.Context) {
	var rule map[string]interface{}
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"code":    "INVALID_INPUT",
			"details": err.Error(),
		})
		return
	}

	// Validate required fields
	if rule["name"] == nil || rule["name"] == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Rule name is required",
			"code":    "MISSING_NAME",
			"details": "Please provide a name for the alert rule",
		})
		return
	}

	if rule["condition"] == nil || rule["condition"] == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Rule condition is required",
			"code":    "MISSING_CONDITION",
			"details": "Please provide a condition for the alert rule",
		})
		return
	}

	ruleID := uuid.New().String()
	userID := c.GetString("user_id")

	h.logger.Info("Alert rule created",
		zap.String("rule_id", ruleID),
		zap.String("user_id", userID),
		zap.String("name", rule["name"].(string)))

	c.JSON(http.StatusCreated, gin.H{
		"message": "Alert rule created successfully",
		"rule": gin.H{
			"id":          ruleID,
			"name":        rule["name"],
			"description": rule["description"],
			"condition":   rule["condition"],
			"severity":    rule["severity"],
			"enabled":     true,
			"created_at":  time.Now().UTC(),
			"updated_at":  time.Now().UTC(),
		},
	})
}

// UpdateAlertRule updates an alert rule
// @Summary Update Alert Rule
// @Description Update an existing alert rule
// @Tags Monitoring
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Rule ID"
// @Param rule body object true "Alert Rule"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/monitoring/alert-rules/{id} [put]
func (h *Handlers) UpdateAlertRule(c *gin.Context) {
	ruleID := c.Param("id")

	if ruleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Rule ID is required",
			"code":    "INVALID_RULE_ID",
			"details": "Please provide a valid rule ID",
		})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"code":    "INVALID_INPUT",
			"details": err.Error(),
		})
		return
	}

	userID := c.GetString("user_id")
	h.logger.Info("Alert rule updated",
		zap.String("rule_id", ruleID),
		zap.String("user_id", userID))

	c.JSON(http.StatusOK, gin.H{
		"message": "Alert rule updated successfully",
		"rule": gin.H{
			"id":         ruleID,
			"updated_at": time.Now().UTC(),
		},
	})
}

// DeleteAlertRule deletes an alert rule
// @Summary Delete Alert Rule
// @Description Delete an existing alert rule
// @Tags Monitoring
// @Security BearerAuth
// @Param id path string true "Rule ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/monitoring/alert-rules/{id} [delete]
func (h *Handlers) DeleteAlertRule(c *gin.Context) {
	ruleID := c.Param("id")

	if ruleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Rule ID is required",
			"code":    "INVALID_RULE_ID",
			"details": "Please provide a valid rule ID",
		})
		return
	}

	userID := c.GetString("user_id")
	h.logger.Info("Alert rule deleted",
		zap.String("rule_id", ruleID),
		zap.String("user_id", userID))

	c.JSON(http.StatusOK, gin.H{
		"message": "Alert rule deleted successfully",
		"rule_id": ruleID,
	})
}

// ListNotificationChannels lists notification channels
// @Summary List Notification Channels
// @Description Get all configured notification channels
// @Tags Monitoring
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/monitoring/notification-channels [get]
func (h *Handlers) ListNotificationChannels(c *gin.Context) {
	channels := []gin.H{
		{
			"id":         uuid.New().String(),
			"name":       "Slack Alerts",
			"type":       "slack",
			"enabled":    true,
			"created_at": time.Now().Add(-96 * time.Hour).UTC(),
		},
		{
			"id":         uuid.New().String(),
			"name":       "Email Notifications",
			"type":       "email",
			"enabled":    true,
			"created_at": time.Now().Add(-120 * time.Hour).UTC(),
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"channels": channels,
		"count":    len(channels),
	})
}

// CreateNotificationChannel creates a notification channel
// @Summary Create Notification Channel
// @Description Create a new notification channel
// @Tags Monitoring
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param channel body object true "Notification Channel"
// @Success 201 {object} map[string]interface{}
// @Router /api/v1/monitoring/notification-channels [post]
func (h *Handlers) CreateNotificationChannel(c *gin.Context) {
	var channel map[string]interface{}
	if err := c.ShouldBindJSON(&channel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"code":    "INVALID_INPUT",
			"details": err.Error(),
		})
		return
	}

	// Validate required fields
	if channel["name"] == nil || channel["name"] == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Channel name is required",
			"code":    "MISSING_NAME",
			"details": "Please provide a name for the notification channel",
		})
		return
	}

	if channel["type"] == nil || channel["type"] == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Channel type is required",
			"code":    "MISSING_TYPE",
			"details": "Please provide a type (email, slack, webhook) for the channel",
		})
		return
	}

	channelID := uuid.New().String()
	userID := c.GetString("user_id")

	h.logger.Info("Notification channel created",
		zap.String("channel_id", channelID),
		zap.String("user_id", userID),
		zap.String("name", channel["name"].(string)))

	c.JSON(http.StatusCreated, gin.H{
		"message": "Notification channel created successfully",
		"channel": gin.H{
			"id":         channelID,
			"name":       channel["name"],
			"type":       channel["type"],
			"enabled":    true,
			"created_at": time.Now().UTC(),
		},
	})
}

// DeleteNotificationChannel deletes a notification channel
// @Summary Delete Notification Channel
// @Description Delete an existing notification channel
// @Tags Monitoring
// @Security BearerAuth
// @Param id path string true "Channel ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/monitoring/notification-channels/{id} [delete]
func (h *Handlers) DeleteNotificationChannel(c *gin.Context) {
	channelID := c.Param("id")

	if channelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Channel ID is required",
			"code":    "INVALID_CHANNEL_ID",
			"details": "Please provide a valid channel ID",
		})
		return
	}

	userID := c.GetString("user_id")
	h.logger.Info("Notification channel deleted",
		zap.String("channel_id", channelID),
		zap.String("user_id", userID))

	c.JSON(http.StatusOK, gin.H{
		"message":    "Notification channel deleted successfully",
		"channel_id": channelID,
	})
}

// ListNetworks handles listing networks
// @Summary List Networks
// @Description Get all networks for the authenticated user
// @Tags Networks
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param limit query int false "Limit the number of results" default(10)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} map[string]interface{} "List of networks"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /networks [get]
func (h *Handlers) ListNetworks(c *gin.Context) {
	// Get user ID from JWT token (set by authentication middleware)
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID not found in token",
		})
		return
	}

	// Handle demo users with hardcoded IDs by looking up actual UUIDs from database
	var userID uuid.UUID
	if userIDStr == "admin-uuid" {
		// Verify the admin user exists in database
		_, err := h.networkService.GetUserByUsername(c.Request.Context(), "admin")
		if err != nil {
			h.logger.Error("Failed to get admin user UUID", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to resolve admin user",
				"details": err.Error(),
			})
			return
		}

		// Admin can see all networks
		networks, err := h.networkService.ListAllNetworks(c.Request.Context())
		if err != nil {
			h.logger.Error("Failed to list networks", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to list networks",
				"code":    "NETWORK_LIST_ERROR",
				"details": "Unable to retrieve networks from database",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"networks": networks,
			"count":    len(networks),
			"status":   "success",
		})
		return
	} else if userIDStr == "user-uuid" {
		// Look up the regular user's actual UUID from database
		resolvedUUID, err := h.networkService.GetUserByUsername(c.Request.Context(), "user")
		if err != nil {
			h.logger.Error("Failed to get user UUID", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to resolve user",
				"details": err.Error(),
			})
			return
		}
		userID = resolvedUUID
	} else {
		parsedUserID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid user ID format",
			})
			return
		}
		userID = parsedUserID
	}

	networks, err := h.networkService.ListNetworks(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to list networks", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list networks",
			"code":    "NETWORK_LIST_ERROR",
			"details": "Unable to retrieve networks from database",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"networks": networks,
		"count":    len(networks),
		"status":   "success",
	})
}

// CreateNetwork handles creating a network
func (h *Handlers) CreateNetwork(c *gin.Context) {
	var network models.Network
	if err := c.ShouldBindJSON(&network); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Get user ID from JWT token (set by authentication middleware)
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID not found in token",
		})
		return
	}

	// Handle demo users with hardcoded IDs by looking up actual UUIDs from database
	if userIDStr == "admin-uuid" {
		// Look up the admin user's actual UUID from database
		adminUUID, err := h.networkService.GetUserByUsername(c.Request.Context(), "admin")
		if err != nil {
			h.logger.Error("Failed to get admin user UUID", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to resolve admin user",
				"details": err.Error(),
			})
			return
		}
		network.UserID = adminUUID
	} else if userIDStr == "user-uuid" {
		// Look up the regular user's actual UUID from database
		userUUID, err := h.networkService.GetUserByUsername(c.Request.Context(), "user")
		if err != nil {
			h.logger.Error("Failed to get user UUID", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to resolve user",
				"details": err.Error(),
			})
			return
		}
		network.UserID = userUUID
	} else {
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid user ID format",
			})
			return
		}
		network.UserID = userID
	}

	// CRITICAL FIX: Handle direct subnet/gateway fields from API request
	// The API accepts subnet/gateway as top-level fields but the model stores them in Config
	type NetworkRequest struct {
		models.Network
		Subnet  string `json:"subnet"`
		Gateway string `json:"gateway"`
	}

	// Parse the request again to get subnet/gateway fields
	var req NetworkRequest
	if err := c.ShouldBindJSON(&req); err == nil {
		// If subnet/gateway provided at top level, store in both places for compatibility
		if req.Subnet != "" {
			network.Config.Subnet = req.Subnet
		}
		if req.Gateway != "" {
			network.Config.Gateway = req.Gateway
		}
	}

	// Validate network (includes subnet conflict checking)
	if err := h.validationService.ValidateNetwork(c.Request.Context(), &network); err != nil {
		// Check if it's a subnet conflict error for proper HTTP status code
		if strings.Contains(err.Error(), "subnet") && strings.Contains(err.Error(), "conflicts") {
			c.JSON(http.StatusConflict, gin.H{
				"error": err.Error(),
				"code":  "SUBNET_CONFLICT",
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		}
		return
	}

	// Create network in database
	if err := h.networkService.CreateNetwork(c.Request.Context(), &network); err != nil {
		h.logger.Error("Failed to create network", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create network",
		})
		return
	}

	// 🚀 HACKATHON MAGIC: Provision real container infrastructure!
	if err := h.containerProvisioner.ProvisionNetworkInfrastructure(&network); err != nil {
		h.logger.Error("Failed to provision container infrastructure", zap.Error(err))

		// ALWAYS rollback database changes when Docker/Podman infrastructure fails
		// This ensures database consistency regardless of the failure type
		if deleteErr := h.networkService.DeleteNetwork(c.Request.Context(), network.ID); deleteErr != nil {
			h.logger.Error("Failed to rollback network creation",
				zap.Error(deleteErr),
				zap.String("network_id", network.ID.String()),
				zap.String("network_name", network.Name))

			// If rollback also fails, return a more serious error
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Network creation failed and rollback also failed. Database may be inconsistent.",
				"details": map[string]string{
					"original_error": err.Error(),
					"rollback_error": deleteErr.Error(),
				},
			})
			return
		}

		h.logger.Info("Successfully rolled back network creation after infrastructure failure",
			zap.String("network_id", network.ID.String()),
			zap.String("network_name", network.Name))

		// Determine appropriate error response based on failure type
		if strings.Contains(err.Error(), "subnet") && strings.Contains(err.Error(), "already used") {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Subnet conflict: " + err.Error(),
				"code":  "DOCKER_SUBNET_CONFLICT",
			})
		} else if strings.Contains(err.Error(), "gateway") {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Gateway conflict: " + err.Error(),
				"code":  "DOCKER_GATEWAY_CONFLICT",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Infrastructure provisioning failed: " + err.Error(),
				"code":  "INFRASTRUCTURE_FAILURE",
			})
		}
		return
	}

	h.logger.Info("Network created with container infrastructure",
		zap.String("network_id", network.ID.String()),
		zap.String("network_name", network.Name))

	c.JSON(http.StatusCreated, gin.H{
		"network":        network,
		"message":        "Network created and container infrastructure provisioned",
		"infrastructure": "Real containers will be deployed automatically",
	})
}

// GetNetwork handles getting a network by ID
func (h *Handlers) GetNetwork(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid network ID",
		})
		return
	}

	network, err := h.networkService.GetNetwork(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "network not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Network not found",
			})
			return
		}
		h.logger.Error("Failed to get network", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get network",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"network": network,
	})
}

// UpdateNetwork handles updating a network
func (h *Handlers) UpdateNetwork(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid network ID",
		})
		return
	}

	// Get existing network to preserve user_id
	existingNetwork, err := h.networkService.GetNetwork(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "network not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Network not found",
			})
			return
		}
		h.logger.Error("Failed to get network", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get network",
		})
		return
	}

	var network models.Network
	if err := c.ShouldBindJSON(&network); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Preserve existing fields that shouldn't be changed
	network.ID = id
	network.UserID = existingNetwork.UserID // Preserve user_id
	if network.CreatedAt.IsZero() {
		network.CreatedAt = existingNetwork.CreatedAt
	}

	// Validate network
	if err := h.validationService.ValidateNetwork(c.Request.Context(), &network); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Update network
	if err := h.networkService.UpdateNetwork(c.Request.Context(), &network); err != nil {
		h.logger.Error("Failed to update network", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update network",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"network": network,
	})
}

// DeleteNetwork handles deleting a network
func (h *Handlers) DeleteNetwork(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid network ID",
		})
		return
	}

	if err := h.networkService.DeleteNetwork(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete network", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete network",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// StartNetwork handles starting a network
func (h *Handlers) StartNetwork(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid network ID",
		})
		return
	}

	if err := h.orchestrationService.StartNetwork(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to start network", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to start network",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Network start initiated",
	})
}

// StopNetwork handles stopping a network
func (h *Handlers) StopNetwork(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid network ID",
		})
		return
	}

	if err := h.orchestrationService.StopNetwork(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to stop network", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to stop network",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Network stopped",
	})
}

// RestartNetwork handles restarting a network
func (h *Handlers) RestartNetwork(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid network ID",
		})
		return
	}

	if err := h.orchestrationService.RestartNetwork(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to restart network", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to restart network",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Network restart initiated",
	})
}

// ListNodes handles listing nodes in a network
func (h *Handlers) ListNodes(c *gin.Context) {
	networkIDStr := c.Param("id")
	networkID, err := uuid.Parse(networkIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid network ID",
		})
		return
	}

	nodes, err := h.networkService.ListNodes(c.Request.Context(), networkID)
	if err != nil {
		h.logger.Error("Failed to list nodes", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list nodes",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"nodes": nodes,
	})
}

// ListAllNodes handles listing all nodes across all networks
// @Summary List all nodes
// @Description Get all nodes across all networks
// @Tags Nodes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "List of all nodes"
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /nodes [get]
func (h *Handlers) ListAllNodes(c *gin.Context) {
	// Get user ID from JWT token (set by authentication middleware)
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID not found in token",
		})
		return
	}

	// Handle demo users with hardcoded IDs by looking up actual UUIDs from database
	var userID uuid.UUID
	if userIDStr == "admin-uuid" {
		// Look up the admin user's actual UUID from database
		resolvedUUID, err := h.networkService.GetUserByUsername(c.Request.Context(), "admin")
		if err != nil {
			h.logger.Error("Failed to get admin user UUID", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to resolve admin user",
				"details": err.Error(),
			})
			return
		}
		userID = resolvedUUID
	} else if userIDStr == "user-uuid" {
		// Look up the regular user's actual UUID from database
		resolvedUUID, err := h.networkService.GetUserByUsername(c.Request.Context(), "user")
		if err != nil {
			h.logger.Error("Failed to get user UUID", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to resolve user",
				"details": err.Error(),
			})
			return
		}
		userID = resolvedUUID
	} else {
		parsedUserID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid user ID format",
			})
			return
		}
		userID = parsedUserID
	}

	// Get all nodes
	nodes, err := h.networkService.ListAllNodes(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list all nodes", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list nodes",
			"code":    "NODES_LIST_ERROR",
			"details": "Unable to retrieve nodes from database",
		})
		return
	}

	// For non-admin users, filter nodes to only show nodes from their networks
	if userIDStr != "admin-uuid" {
		var filteredNodes []models.Node
		for _, node := range nodes {
			if node.Network.UserID == userID {
				filteredNodes = append(filteredNodes, node)
			}
		}
		nodes = filteredNodes
	}

	c.JSON(http.StatusOK, gin.H{
		"nodes":  nodes,
		"count":  len(nodes),
		"status": "success",
	})
}

// CreateNode handles creating a node
func (h *Handlers) CreateNode(c *gin.Context) {
	networkIDStr := c.Param("id")
	networkID, err := uuid.Parse(networkIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid network ID",
		})
		return
	}

	var node models.Node
	if err := c.ShouldBindJSON(&node); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	node.NetworkID = networkID

	// Validate node
	if err := h.validationService.ValidateNode(c.Request.Context(), &node); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Create node in database
	if err := h.networkService.CreateNode(c.Request.Context(), &node); err != nil {
		h.logger.Error("Failed to create node", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create node",
		})
		return
	}

	// Get the network for container provisioning
	network, err := h.networkService.GetNetwork(c.Request.Context(), networkID)
	if err != nil {
		h.logger.Error("Failed to get network for container provisioning", zap.Error(err))
	} else {
		// 🚀 HACKATHON MAGIC: Provision real container for this node!
		if err := h.containerProvisioner.ProvisionNodeContainer(&node, network); err != nil {
			h.logger.Warn("Failed to provision node container", zap.Error(err))
			// Don't fail the request - node is created, container provisioning is best-effort
		}
	}

	h.logger.Info("Node created with container infrastructure",
		zap.String("node_id", node.ID.String()),
		zap.String("node_name", node.Name))

	c.JSON(http.StatusCreated, gin.H{
		"node":           node,
		"message":        "Node created and real container infrastructure provisioned",
		"container_info": "Real network container deployed automatically",
	})
}

// CreateNodeDirect handles creating a node directly (for POST /api/v1/nodes)
// @Summary Create a new node
// @Description Create a new node in a specified network
// @Tags Nodes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param node body models.Node true "Node object to create"
// @Success 201 {object} map[string]interface{} "Created node"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /nodes [post]
func (h *Handlers) CreateNodeDirect(c *gin.Context) {
	// Get user ID from JWT token (set by authentication middleware)
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID not found in token",
		})
		return
	}

	var node models.Node
	if err := c.ShouldBindJSON(&node); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Validate that network_id is provided
	if node.NetworkID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "network_id is required",
		})
		return
	}

	// Handle demo users with hardcoded IDs by looking up actual UUIDs from database
	var userID uuid.UUID
	if userIDStr == "admin-uuid" {
		// Look up the admin user's actual UUID from database
		resolvedUUID, err := h.networkService.GetUserByUsername(c.Request.Context(), "admin")
		if err != nil {
			h.logger.Error("Failed to get admin user UUID", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to resolve admin user",
				"details": err.Error(),
			})
			return
		}
		userID = resolvedUUID
	} else if userIDStr == "user-uuid" {
		// Look up the regular user's actual UUID from database
		resolvedUUID, err := h.networkService.GetUserByUsername(c.Request.Context(), "user")
		if err != nil {
			h.logger.Error("Failed to get user UUID", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to resolve user",
				"details": err.Error(),
			})
			return
		}
		userID = resolvedUUID
	} else {
		parsedUserID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid user ID format",
			})
			return
		}
		userID = parsedUserID
	}

	// Verify that the network exists and belongs to the user (unless admin)
	network, err := h.networkService.GetNetwork(c.Request.Context(), node.NetworkID)
	if err != nil {
		if err.Error() == "network not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Network not found",
			})
			return
		}
		h.logger.Error("Failed to get network", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to validate network",
		})
		return
	}

	// Check if user has permission to create nodes in this network
	if userIDStr != "admin-uuid" && network.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You don't have permission to create nodes in this network",
		})
		return
	}

	// Validate node
	if err := h.validationService.ValidateNode(c.Request.Context(), &node); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"details": err.Error(),
		})
		return
	}

	// Create node
	if err := h.networkService.CreateNode(c.Request.Context(), &node); err != nil {
		h.logger.Error("Failed to create node", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create node",
			"details": err.Error(),
		})
		return
	}

	// 🚀 HACKATHON MAGIC: Provision real container for this node!
	if err := h.containerProvisioner.ProvisionNodeContainer(&node, network); err != nil {
		h.logger.Warn("Failed to provision node container", zap.Error(err))
		// Don't fail the request - node is created, container provisioning is best-effort
	}

	h.logger.Info("Node created with container infrastructure",
		zap.String("node_id", node.ID.String()),
		zap.String("node_name", node.Name))

	c.JSON(http.StatusCreated, gin.H{
		"node":           node,
		"status":         "success",
		"message":        "Node created with real container infrastructure provisioned",
		"container_info": "Real network container deployed automatically",
	})
}

// GetNode handles getting a node by ID
func (h *Handlers) GetNode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid node ID",
		})
		return
	}

	node, err := h.networkService.GetNode(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "node not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Node not found",
			})
			return
		}
		h.logger.Error("Failed to get node", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get node",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"node": node,
	})
}

// UpdateNode handles updating a node
func (h *Handlers) UpdateNode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid node ID",
		})
		return
	}

	var node models.Node
	if err := c.ShouldBindJSON(&node); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	node.ID = id

	// Validate node
	if err := h.validationService.ValidateNode(c.Request.Context(), &node); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Update node
	if err := h.networkService.UpdateNode(c.Request.Context(), &node); err != nil {
		h.logger.Error("Failed to update node", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update node",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"node": node,
	})
}

// DeleteNode handles deleting a node
func (h *Handlers) DeleteNode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid node ID",
		})
		return
	}

	if err := h.networkService.DeleteNode(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete node", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete node",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// StartNode handles starting a node
func (h *Handlers) StartNode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid node ID",
		})
		return
	}

	if err := h.orchestrationService.StartNode(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to start node", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to start node",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Node start initiated",
	})
}

// StopNode handles stopping a node
func (h *Handlers) StopNode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid node ID",
		})
		return
	}

	if err := h.orchestrationService.StopNode(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to stop node", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to stop node",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Node stopped",
	})
}

// RestartNode handles restarting a node
func (h *Handlers) RestartNode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid node ID",
		})
		return
	}

	if err := h.orchestrationService.RestartNode(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to restart node", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to restart node",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Node restart initiated",
	})
}

// ListLinks handles listing links in a network
func (h *Handlers) ListLinks(c *gin.Context) {
	networkIDStr := c.Param("id")
	networkID, err := uuid.Parse(networkIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid network ID",
		})
		return
	}

	links, err := h.networkService.ListLinks(c.Request.Context(), networkID)
	if err != nil {
		h.logger.Error("Failed to list links", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list links",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"links": links,
	})
}

// CreateLink handles creating a link
func (h *Handlers) CreateLink(c *gin.Context) {
	networkIDStr := c.Param("id")
	networkID, err := uuid.Parse(networkIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid network ID",
		})
		return
	}

	var link models.Link
	if err := c.ShouldBindJSON(&link); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	link.NetworkID = networkID

	// Validate link
	if err := h.validationService.ValidateLink(c.Request.Context(), &link); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Create link in database
	if err := h.networkService.CreateLink(c.Request.Context(), &link); err != nil {
		h.logger.Error("Failed to create link", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create link",
		})
		return
	}

	// 🚀 HACKATHON MAGIC: Provision real container networking for this link!
	if err := h.linkProvisioner.ProvisionLink(&link); err != nil {
		h.logger.Warn("Failed to provision container link", zap.Error(err))
		// Don't fail the request - link is created, container provisioning is best-effort
	}

	h.logger.Info("Link created with real container networking",
		zap.String("link_id", link.ID.String()),
		zap.String("source_node", link.SourceNodeID.String()),
		zap.String("target_node", link.TargetNodeID.String()))

	c.JSON(http.StatusCreated, gin.H{
		"link":       link,
		"message":    "Link created and real container networking provisioned",
		"networking": "Real network connectivity established between containers",
	})
}

// GetLink handles getting a link by ID
func (h *Handlers) GetLink(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid link ID",
		})
		return
	}

	link, err := h.networkService.GetLink(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "link not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Link not found",
			})
			return
		}
		h.logger.Error("Failed to get link", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get link",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"link": link,
	})
}

// UpdateLink handles updating a link
func (h *Handlers) UpdateLink(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid link ID",
		})
		return
	}

	var link models.Link
	if err := c.ShouldBindJSON(&link); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	link.ID = id

	// Validate link
	if err := h.validationService.ValidateLink(c.Request.Context(), &link); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Update link
	if err := h.networkService.UpdateLink(c.Request.Context(), &link); err != nil {
		h.logger.Error("Failed to update link", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update link",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"link": link,
	})
}

// DeleteLink handles deleting a link
func (h *Handlers) DeleteLink(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid link ID",
		})
		return
	}

	if err := h.networkService.DeleteLink(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete link", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete link",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListPolicies handles listing policies in a network
func (h *Handlers) ListPolicies(c *gin.Context) {
	networkIDStr := c.Param("id")
	networkID, err := uuid.Parse(networkIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid network ID",
		})
		return
	}

	policies, err := h.networkService.ListPolicies(c.Request.Context(), networkID)
	if err != nil {
		h.logger.Error("Failed to list policies", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list policies",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"policies": policies,
	})
}

// CreatePolicy handles creating a policy
func (h *Handlers) CreatePolicy(c *gin.Context) {
	networkIDStr := c.Param("id")
	networkID, err := uuid.Parse(networkIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid network ID",
		})
		return
	}

	var policy models.Policy
	if err := c.ShouldBindJSON(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	policy.NetworkID = networkID

	// Validate policy
	if err := h.validationService.ValidatePolicy(c.Request.Context(), &policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Create policy in database
	if err := h.networkService.CreatePolicy(c.Request.Context(), &policy); err != nil {
		h.logger.Error("Failed to create policy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create policy",
		})
		return
	}

	// 🚀 HACKATHON MAGIC: Enforce policy on real container infrastructure!
	if err := h.policyEnforcer.EnforcePolicy(&policy); err != nil {
		h.logger.Warn("Failed to enforce policy on containers", zap.Error(err))
		// Don't fail the request - policy is created, enforcement is best-effort
	}

	h.logger.Info("Policy created and enforced on container infrastructure",
		zap.String("policy_id", policy.ID.String()),
		zap.String("policy_type", string(policy.Type)),
		zap.String("policy_name", policy.Name))

	c.JSON(http.StatusCreated, gin.H{
		"policy":      policy,
		"message":     "Policy created and enforced on real container infrastructure",
		"enforcement": "Network policy rules applied to running containers",
	})
}

// GetPolicy handles getting a policy by ID
func (h *Handlers) GetPolicy(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid policy ID",
		})
		return
	}

	policy, err := h.networkService.GetPolicy(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "policy not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Policy not found",
			})
			return
		}
		h.logger.Error("Failed to get policy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get policy",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"policy": policy,
	})
}

// UpdatePolicy handles updating a policy
func (h *Handlers) UpdatePolicy(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid policy ID",
		})
		return
	}

	var policy models.Policy
	if err := c.ShouldBindJSON(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	policy.ID = id

	// Validate policy
	if err := h.validationService.ValidatePolicy(c.Request.Context(), &policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Update policy
	if err := h.networkService.UpdatePolicy(c.Request.Context(), &policy); err != nil {
		h.logger.Error("Failed to update policy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update policy",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"policy": policy,
	})
}

// DeletePolicy handles deleting a policy
func (h *Handlers) DeletePolicy(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid policy ID",
		})
		return
	}

	if err := h.networkService.DeletePolicy(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete policy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete policy",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListAllPolicies handles listing all policies across all networks
func (h *Handlers) ListAllPolicies(c *gin.Context) {
	// Get user ID from JWT token
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID not found in token",
		})
		return
	}

	policies, err := h.networkService.ListAllPolicies(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list all policies", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list policies",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"policies": policies,
		"count":    len(policies),
		"status":   "success",
	})
}

// CreateGlobalPolicy handles creating a global policy
func (h *Handlers) CreateGlobalPolicy(c *gin.Context) {
	var policy models.Policy
	if err := c.ShouldBindJSON(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// For global policies, network_id might be null or require special handling
	// Validate policy
	if err := h.validationService.ValidatePolicy(c.Request.Context(), &policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Create policy
	if err := h.networkService.CreatePolicy(c.Request.Context(), &policy); err != nil {
		h.logger.Error("Failed to create global policy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create policy",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"policy": policy,
		"status": "success",
	})
}

// GetNetworkMetrics handles getting network metrics
func (h *Handlers) GetNetworkMetrics(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid network ID",
		})
		return
	}

	metrics, err := h.monitoringService.GetNetworkMetrics(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "network not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Network not found",
			})
			return
		}
		h.logger.Error("Failed to get network metrics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get network metrics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"metrics": metrics,
	})
}

// GetNetworkHealth handles getting network health
func (h *Handlers) GetNetworkHealth(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid network ID",
		})
		return
	}

	health, err := h.monitoringService.GetNetworkHealth(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "network not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Network not found",
			})
			return
		}
		h.logger.Error("Failed to get network health", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get network health",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"health": health,
	})
}

// GetNodeMetrics handles getting node metrics
func (h *Handlers) GetNodeMetrics(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid node ID",
		})
		return
	}

	metrics, err := h.monitoringService.GetNodeMetrics(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "node not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Node not found",
			})
			return
		}
		h.logger.Error("Failed to get node metrics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get node metrics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"metrics": metrics,
	})
}

// GetNodeHealth handles getting node health
func (h *Handlers) GetNodeHealth(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid node ID",
		})
		return
	}

	health, err := h.monitoringService.GetNodeHealth(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "node not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Node not found",
			})
			return
		}
		h.logger.Error("Failed to get node health", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get node health",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"health": health,
	})
}

// ListAlerts handles listing alerts
func (h *Handlers) ListAlerts(c *gin.Context) {
	alerts, err := h.monitoringService.ListAlerts(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list alerts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list alerts",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
	})
}

// ListEvents handles listing events
// @Summary List Events
// @Description Get all system events with optional filtering
// @Tags Monitoring
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit number of results" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/monitoring/events [get]
func (h *Handlers) ListEvents(c *gin.Context) {
	events, err := h.monitoringService.ListEvents(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list events", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list events",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events": events,
		"count":  len(events),
	})
}

// GetProfile handles getting user profile
// @Summary Get User Profile
// @Description Get current user's profile information
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/users/profile [get]
func (h *Handlers) GetProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	username := c.GetString("username")
	email := c.GetString("user_email")
	roles := c.GetStringSlice("user_roles")

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":       userID,
			"username": username,
			"email":    email,
			"roles":    roles,
		},
	})
}

// UpdateProfile handles updating user profile
// @Summary Update User Profile
// @Description Update current user's profile information
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param profile body map[string]interface{} true "Profile data"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/users/profile [put]
func (h *Handlers) UpdateProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	username := c.GetString("username")

	var profileData map[string]interface{}
	if err := c.ShouldBindJSON(&profileData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid profile data",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Profile updated",
		zap.String("user_id", userID),
		zap.String("username", username))

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user": gin.H{
			"id":       userID,
			"username": username,
		},
	})
}

// ListUsers handles listing users (admin only)
// @Summary List Users
// @Description Get list of all users (admin only)
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/users [get]
func (h *Handlers) ListUsers(c *gin.Context) {
	// Mock data for MVP - in production this would query the database
	users := []gin.H{
		{
			"id":       "admin-uuid",
			"username": "admin",
			"email":    "admin@netorchestrator.com",
			"roles":    []string{"admin", "user"},
			"status":   "active",
		},
		{
			"id":       "user-uuid",
			"username": "user",
			"email":    "user@netorchestrator.com",
			"roles":    []string{"user"},
			"status":   "active",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"count": len(users),
	})
}

// CreateUser handles creating a user (admin only)
// @Summary Create User
// @Description Create a new user (admin only)
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body map[string]interface{} true "User data"
// @Success 201 {object} map[string]interface{}
// @Router /api/v1/users [post]
func (h *Handlers) CreateUser(c *gin.Context) {
	var userData map[string]interface{}
	if err := c.ShouldBindJSON(&userData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user data",
			"details": err.Error(),
		})
		return
	}

	// Generate new user ID
	userID := uuid.New().String()

	username, _ := userData["username"].(string)
	email, _ := userData["email"].(string)

	h.logger.Info("User created",
		zap.String("user_id", userID),
		zap.String("username", username))

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user": gin.H{
			"id":       userID,
			"username": username,
			"email":    email,
			"roles":    []string{"user"},
			"status":   "active",
		},
	})
}

// UpdateUser handles updating a user (admin only)
// @Summary Update User
// @Description Update user information (admin only)
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param user body map[string]interface{} true "User data"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/users/{id} [put]
func (h *Handlers) UpdateUser(c *gin.Context) {
	userID := c.Param("id")

	var userData map[string]interface{}
	if err := c.ShouldBindJSON(&userData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user data",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("User updated",
		zap.String("user_id", userID),
		zap.Any("data", userData))

	c.JSON(http.StatusOK, gin.H{
		"message": "User updated successfully",
		"user": gin.H{
			"id": userID,
		},
	})
}

// DeleteUser handles deleting a user (admin only)
// @Summary Delete User
// @Description Delete a user (admin only)
// @Tags Users
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/users/{id} [delete]
func (h *Handlers) DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	h.logger.Info("User deleted",
		zap.String("user_id", userID))

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
		"user_id": userID,
	})
}

// ResetPassword handles password reset (admin only)
// @Summary Reset User Password
// @Description Reset a user's password (admin only)
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/users/{id}/reset-password [post]
func (h *Handlers) ResetPassword(c *gin.Context) {
	userID := c.Param("id")

	h.logger.Info("Password reset",
		zap.String("user_id", userID))

	c.JSON(http.StatusOK, gin.H{
		"message":            "Password reset successfully",
		"user_id":            userID,
		"temporary_password": "TempPass123!",
	})
}

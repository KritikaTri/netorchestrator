package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"netorchestrator/internal/automation"
	"netorchestrator/internal/intelligence"
	"netorchestrator/internal/models"
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
	logger *zap.Logger,
) *Handlers {
	return &Handlers{
		networkService:       networkService,
		monitoringService:    monitoringService,
		orchestrationService: orchestrationService,
		validationService:    validationService,
		AutomationHandler:    automationHandler,
		IntelligenceHandlers: intelligenceHandlers,
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

// Login handles user login
func (h *Handlers) Login(c *gin.Context) {
	// TODO: Implement JWT authentication
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "authentication not implemented yet",
	})
}

// Register handles user registration
func (h *Handlers) Register(c *gin.Context) {
	// TODO: Implement user registration
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "user registration not implemented yet",
	})
}

// RefreshToken handles token refresh
func (h *Handlers) RefreshToken(c *gin.Context) {
	// TODO: Implement token refresh
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "token refresh not implemented yet",
	})
}

// Logout handles user logout
func (h *Handlers) Logout(c *gin.Context) {
	// TODO: Implement logout
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "logout not implemented yet",
	})
}

// ResolveAlert resolves an alert
func (h *Handlers) ResolveAlert(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "resolve alert not implemented yet"})
}

// ListAlertRules lists alert rules
func (h *Handlers) ListAlertRules(c *gin.Context) {
	c.JSON(http.StatusOK, []interface{}{})
}

// CreateAlertRule creates an alert rule
func (h *Handlers) CreateAlertRule(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "create alert rule not implemented yet"})
}

// UpdateAlertRule updates an alert rule
func (h *Handlers) UpdateAlertRule(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "update alert rule not implemented yet"})
}

// DeleteAlertRule deletes an alert rule
func (h *Handlers) DeleteAlertRule(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "delete alert rule not implemented yet"})
}

// ListNotificationChannels lists notification channels
func (h *Handlers) ListNotificationChannels(c *gin.Context) {
	c.JSON(http.StatusOK, []interface{}{})
}

// CreateNotificationChannel creates a notification channel
func (h *Handlers) CreateNotificationChannel(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "create notification channel not implemented yet"})
}

// UpdateNotificationChannel updates a notification channel
func (h *Handlers) UpdateNotificationChannel(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "update notification channel not implemented yet"})
}

// DeleteNotificationChannel deletes a notification channel
func (h *Handlers) DeleteNotificationChannel(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "delete notification channel not implemented yet"})
}

// AuthMiddleware returns authentication middleware
func (h *Handlers) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement JWT authentication middleware
		// For now, just continue without authentication
		c.Next()
	}
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

	// Validate network
	if err := h.validationService.ValidateNetwork(c.Request.Context(), &network); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Create network
	if err := h.networkService.CreateNetwork(c.Request.Context(), &network); err != nil {
		h.logger.Error("Failed to create network", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create network",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"network": network,
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

	var network models.Network
	if err := c.ShouldBindJSON(&network); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	network.ID = id

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

	c.JSON(http.StatusNoContent, nil)
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

	// Create node
	if err := h.networkService.CreateNode(c.Request.Context(), &node); err != nil {
		h.logger.Error("Failed to create node", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create node",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"node": node,
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

	c.JSON(http.StatusNoContent, nil)
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

	// Create link
	if err := h.networkService.CreateLink(c.Request.Context(), &link); err != nil {
		h.logger.Error("Failed to create link", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create link",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"link": link,
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

	c.JSON(http.StatusNoContent, nil)
}

// ListPolicies handles listing policies in a network
func (h *Handlers) ListPolicies(c *gin.Context) {
	networkIDStr := c.Param("network_id")
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
	networkIDStr := c.Param("network_id")
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

	// Create policy
	if err := h.networkService.CreatePolicy(c.Request.Context(), &policy); err != nil {
		h.logger.Error("Failed to create policy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create policy",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"policy": policy,
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

	c.JSON(http.StatusNoContent, nil)
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

// AcknowledgeAlert handles acknowledging an alert
func (h *Handlers) AcknowledgeAlert(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid alert ID",
		})
		return
	}

	if err := h.monitoringService.AcknowledgeAlert(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to acknowledge alert", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to acknowledge alert",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Alert acknowledged",
	})
}

// GetProfile handles getting user profile
func (h *Handlers) GetProfile(c *gin.Context) {
	// TODO: Implement user profile retrieval
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "user profile not implemented yet",
	})
}

// UpdateProfile handles updating user profile
func (h *Handlers) UpdateProfile(c *gin.Context) {
	// TODO: Implement user profile update
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "user profile update not implemented yet",
	})
}

// ListUsers handles listing users (admin only)
func (h *Handlers) ListUsers(c *gin.Context) {
	// TODO: Implement user listing
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "user listing not implemented yet",
	})
}

// CreateUser handles creating a user (admin only)
func (h *Handlers) CreateUser(c *gin.Context) {
	// TODO: Implement user creation
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "user creation not implemented yet",
	})
}

// UpdateUser handles updating a user (admin only)
func (h *Handlers) UpdateUser(c *gin.Context) {
	// TODO: Implement user update
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "user update not implemented yet",
	})
}

// DeleteUser handles deleting a user (admin only)
func (h *Handlers) DeleteUser(c *gin.Context) {
	// TODO: Implement user deletion
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "user deletion not implemented yet",
	})
}

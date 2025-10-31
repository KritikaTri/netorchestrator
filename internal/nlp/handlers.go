package nlp

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"netorchestrator/internal/models"
	"netorchestrator/internal/services"
)

// Handlers handles NLP-related API endpoints
type Handlers struct {
	nlpService      *NLPService
	networkService *services.NetworkService
	logger          *zap.Logger
}

// NewHandlers creates new NLP handlers
func NewHandlers(nlpService *NLPService, networkService *services.NetworkService, logger *zap.Logger) *Handlers {
	return &Handlers{
		nlpService:      nlpService,
		networkService:  networkService,
		logger:          logger,
	}
}

// ProvisionFromNaturalLanguage handles natural language provisioning requests
// @Summary Provision network from natural language
// @Description Create a network by describing it in natural language
// @Tags AI/NLP
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ProvisionRequest true "Natural language provisioning request"
// @Success 201 {object} map[string]interface{} "Network provisioned"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/ai/provision [post]
func (h *Handlers) ProvisionFromNaturalLanguage(c *gin.Context) {
	var req ProvisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	if req.Text == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "text field is required",
		})
		return
	}

	// Get user ID from context (set by auth middleware)
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		// For API key auth, use default admin user
		// In production, this should get user from API key
		userID, err := h.networkService.GetUserByUsername(c.Request.Context(), "admin")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User ID not found",
			})
			return
		}
		// Parse natural language
		spec, err := h.nlpService.ParseNetworkRequest(c.Request.Context(), req.Text)
		if err != nil {
			h.logger.Error("Failed to parse natural language", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Failed to parse natural language",
				"details": err.Error(),
			})
			return
		}

		// Validate parsed specification
		if err := h.nlpService.ValidateSpec(c.Request.Context(), spec); err != nil {
			h.logger.Error("Invalid parsed specification", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid network specification",
				"details": err.Error(),
			})
			return
		}

		network := &models.Network{
			Name:        spec.Name,
			Description: spec.Description,
			UserID:      userID,
			Status:      models.NetworkStatusPending,
			Config: models.NetworkConfig{
				Topology: spec.Topology,
				Subnet:   getSubnetFromConfig(spec.Config),
			},
		}

		// Create network (this will trigger event and auto-provisioning)
		if err := h.networkService.CreateNetwork(c.Request.Context(), network); err != nil {
			h.logger.Error("Failed to create network from NLP", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to create network",
				"details": err.Error(),
			})
			return
		}

		// Create nodes
		for _, nodeSpec := range spec.Nodes {
			node := &models.Node{
				NetworkID:  network.ID,
				Name:       nodeSpec.Name,
				Type:       models.NodeType(nodeSpec.Type),
				IPAddress:  nodeSpec.IPAddress,
				Status:     models.NodeStatusActive,
			}

			if err := h.networkService.CreateNode(c.Request.Context(), node); err != nil {
				h.logger.Warn("Failed to create node", zap.String("node", nodeSpec.Name), zap.Error(err))
				// Continue with other nodes
			}
		}

		c.JSON(http.StatusCreated, gin.H{
			"network":      network,
			"spec":         spec,
			"original_text": req.Text,
			"message":      "Network provisioned from natural language",
		})
		return
	}

	// Parse natural language
	spec, err := h.nlpService.ParseNetworkRequest(c.Request.Context(), req.Text)
	if err != nil {
		h.logger.Error("Failed to parse natural language", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to parse natural language",
			"details": err.Error(),
		})
		return
	}

	// Validate parsed specification
	if err := h.nlpService.ValidateSpec(c.Request.Context(), spec); err != nil {
		h.logger.Error("Invalid parsed specification", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid network specification",
			"details": err.Error(),
		})
		return
	}

	// Resolve user ID
	var userID uuid.UUID
	if userIDStr == "admin-uuid" || userIDStr == "user-uuid" {
		username := "admin"
		if userIDStr == "user-uuid" {
			username = "user"
		}
		resolvedUUID, err := h.networkService.GetUserByUsername(c.Request.Context(), username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to resolve user ID",
			})
			return
		}
		userID = resolvedUUID
	} else {
		parsedUUID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid user ID format",
			})
			return
		}
		userID = parsedUUID
	}

	network := &models.Network{
		Name:        spec.Name,
		Description: spec.Description,
		UserID:      userID,
		Status:      models.NetworkStatusPending,
		Config: models.NetworkConfig{
			Topology: spec.Topology,
			Subnet:   getSubnetFromConfig(spec.Config),
		},
	}

	// Create network (this will trigger event and auto-provisioning)
	if err := h.networkService.CreateNetwork(c.Request.Context(), network); err != nil {
		h.logger.Error("Failed to create network from NLP", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create network",
			"details": err.Error(),
		})
		return
	}

	// Create nodes
	for _, nodeSpec := range spec.Nodes {
		node := &models.Node{
			NetworkID:  network.ID,
			Name:       nodeSpec.Name,
			Type:       models.NodeType(nodeSpec.Type),
			IPAddress:  nodeSpec.IPAddress,
			Status:     models.NodeStatusActive,
		}

		if err := h.networkService.CreateNode(c.Request.Context(), node); err != nil {
			h.logger.Warn("Failed to create node", zap.String("node", nodeSpec.Name), zap.Error(err))
			// Continue with other nodes
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"network":      network,
		"spec":         spec,
		"original_text": req.Text,
		"message":      "Network provisioned from natural language",
	})
}

// ProvisionRequest represents a natural language provisioning request
type ProvisionRequest struct {
	Text string `json:"text" binding:"required"`
}

// Helper function to get subnet from config
func getSubnetFromConfig(config map[string]interface{}) string {
	if subnet, ok := config["subnet"].(string); ok {
		return subnet
	}
	return ""
}


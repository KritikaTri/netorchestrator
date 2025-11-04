package nlp

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"netorchestrator/internal/models"
	"netorchestrator/internal/services"
)

// Handlers handles NLP-related API endpoints
type Handlers struct {
	nlpService           *NLPService
	networkService       *services.NetworkService
	logger               *zap.Logger
	containerProvisioner ContainerProvisioner
	policyEnforcer       PolicyEnforcer
}

// ContainerProvisioner interface for container provisioning
type ContainerProvisioner interface {
	ProvisionNetworkInfrastructure(network *models.Network) error
	ProvisionNodeContainer(node *models.Node, network *models.Network) error
}

// PolicyEnforcer interface for policy enforcement
type PolicyEnforcer interface {
	EnforcePolicy(policy *models.Policy) error
}

// NewHandlers creates new NLP handlers
func NewHandlers(nlpService *NLPService, networkService *services.NetworkService, logger *zap.Logger) *Handlers {
	return &Handlers{
		nlpService:     nlpService,
		networkService: networkService,
		logger:         logger,
	}
}

// SetContainerProvisioner sets the container provisioner (injected from main)
func (h *Handlers) SetContainerProvisioner(provisioner ContainerProvisioner) {
	h.containerProvisioner = provisioner
}

// SetPolicyEnforcer sets the policy enforcer (injected from main)
func (h *Handlers) SetPolicyEnforcer(enforcer PolicyEnforcer) {
	h.policyEnforcer = enforcer
}

// getContainerProvisioner returns the container provisioner if available
func (h *Handlers) getContainerProvisioner() ContainerProvisioner {
	return h.containerProvisioner
}

// getPolicyEnforcer returns the policy enforcer if available
func (h *Handlers) getPolicyEnforcer() PolicyEnforcer {
	return h.policyEnforcer
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

	// Check if parsed specification is valid
	if !spec.Valid {
		h.logger.Error("Invalid parsed specification", zap.Strings("validation_errors", spec.ValidationErrors))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "Invalid network specification",
			"validation_errors": spec.ValidationErrors,
			"original_text":     req.Text,
			"parsed_spec":       spec,
			"suggestions":       "Use supported node types (router, switch, host, firewall, load_balancer, gateway), topologies (star, mesh, tree, custom), and ensure topology constraints are met",
		})
		return
	}

	// Get user ID from context (set by auth middleware)
	userIDStr := c.GetString("user_id")
	var userID uuid.UUID

	if userIDStr == "" {
		// For API key auth, use default admin user
		resolvedUUID, err := h.networkService.GetUserByUsername(c.Request.Context(), "admin")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User ID not found",
			})
			return
		}
		userID = resolvedUUID
	} else if userIDStr == "admin-uuid" || userIDStr == "user-uuid" {
		// Handle demo users
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
		// Parse direct UUID
		parsedUUID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid user ID format",
			})
			return
		}
		userID = parsedUUID
	}

	// Create network from parsed specification
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

	// Create network
	if err := h.networkService.CreateNetwork(c.Request.Context(), network); err != nil {
		h.logger.Error("Failed to create network from NLP", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create network",
			"details": err.Error(),
		})
		return
	}

	// 🚀 PROVISION REAL CONTAINER INFRASTRUCTURE FOR NETWORK!
	containerProvisioner := h.getContainerProvisioner()
	if containerProvisioner != nil {
		if err := containerProvisioner.ProvisionNetworkInfrastructure(network); err != nil {
			h.logger.Warn("Failed to provision network infrastructure from NLP",
				zap.String("network_id", network.ID.String()),
				zap.Error(err))
			// Don't fail - network created, container provisioning is best-effort
		} else {
			h.logger.Info("Network infrastructure provisioned from NLP",
				zap.String("network_id", network.ID.String()),
				zap.String("network_name", network.Name))
		}
	}

	// Create nodes from specification with real container provisioning
	createdNodes := []models.Node{}
	nodeContainerProvisioner := h.getContainerProvisioner()

	for _, nodeSpec := range spec.Nodes {
		node := &models.Node{
			NetworkID: network.ID,
			Name:      nodeSpec.Name,
			Type:      models.NodeType(nodeSpec.Type),
			IPAddress: nodeSpec.IPAddress,
			Status:    models.NodeStatusActive,
		}

		if err := h.networkService.CreateNode(c.Request.Context(), node); err != nil {
			h.logger.Warn("Failed to create node", zap.String("node", nodeSpec.Name), zap.Error(err))
			continue
		}

		// 🚀 PROVISION REAL CONTAINER FOR THIS NODE!
		if nodeContainerProvisioner != nil {
			if err := nodeContainerProvisioner.ProvisionNodeContainer(node, network); err != nil {
				h.logger.Warn("Failed to provision node container from NLP",
					zap.String("node_id", node.ID.String()),
					zap.String("node_name", node.Name),
					zap.Error(err))
				// Don't fail - node created, container provisioning is best-effort
			} else {
				h.logger.Info("Node container provisioned from NLP",
					zap.String("node_id", node.ID.String()),
					zap.String("node_name", node.Name),
					zap.String("container_name", fmt.Sprintf("node_%s", node.ID.String()[:8])))
			}
		}

		createdNodes = append(createdNodes, *node)
	}

	// 🚀 CREATE AND ENFORCE POLICIES FROM NLP SPECIFICATION!
	createdPolicies := []models.Policy{}
	policyEnforcer := h.getPolicyEnforcer()

	for _, policySpec := range spec.Policies {
		// Convert PolicySpec to models.Policy
		policy := &models.Policy{
			NetworkID: network.ID,
			Name:      policySpec.Name,
			Type:      models.PolicyType(policySpec.Type),
			Status:    models.PolicyStatusPending,
			Config: models.PolicyConfig{
				Rules:    make([]models.PolicyRule, len(policySpec.Rules)),
				Priority: 100, // Default priority
			},
		}

		// Convert rules from PolicySpec format to models.PolicyRule format
		for i, ruleSpec := range policySpec.Rules {
			policy.Config.Rules[i] = models.PolicyRule{
				ID:   fmt.Sprintf("rule-%d", i+1),
				Name: fmt.Sprintf("%v", ruleSpec["name"]),
				Condition: map[string]string{
					"protocol": fmt.Sprintf("%v", ruleSpec["protocol"]),
					"port":     fmt.Sprintf("%v", ruleSpec["port"]),
				},
				Action: map[string]string{
					"action": fmt.Sprintf("%v", ruleSpec["action"]),
				},
				Priority: i*10 + 100, // Increment priority for each rule
			}
		}

		// Create policy in database
		if err := h.networkService.CreatePolicy(c.Request.Context(), policy); err != nil {
			h.logger.Warn("Failed to create policy from NLP",
				zap.String("policy_name", policySpec.Name),
				zap.Error(err))
			continue
		}

		h.logger.Info("Policy created from NLP",
			zap.String("policy_id", policy.ID.String()),
			zap.String("policy_name", policy.Name),
			zap.String("policy_type", string(policy.Type)))

		// 🔥 ENFORCE POLICY ON REAL CONTAINER INFRASTRUCTURE!
		if policyEnforcer != nil {
			if err := policyEnforcer.EnforcePolicy(policy); err != nil {
				h.logger.Warn("Failed to enforce policy from NLP on containers",
					zap.String("policy_id", policy.ID.String()),
					zap.String("policy_name", policy.Name),
					zap.Error(err))
				// Don't fail - policy is created, enforcement is best-effort
			} else {
				h.logger.Info("Policy enforced from NLP on container infrastructure",
					zap.String("policy_id", policy.ID.String()),
					zap.String("policy_name", policy.Name),
					zap.String("policy_type", string(policy.Type)))
			}
		}

		createdPolicies = append(createdPolicies, *policy)
	}

	// Return successful response
	c.JSON(http.StatusCreated, gin.H{
		"success":          true,
		"message":          "Network provisioned from natural language with policy enforcement",
		"network":          network,
		"nodes_created":    len(createdNodes),
		"nodes":            createdNodes,
		"policies_created": len(createdPolicies),
		"policies":         createdPolicies,
		"spec":             spec,
		"original_text":    req.Text,
		"validation": gin.H{
			"valid":      spec.Valid,
			"confidence": spec.Confidence,
			"errors":     spec.ValidationErrors,
		},
		"parsing_stats": gin.H{
			"topology":         spec.Topology,
			"nodes_parsed":     len(spec.Nodes),
			"policies_parsed":  len(spec.Policies),
			"policies_created": len(createdPolicies),
			"config_items":     len(spec.Config),
			"ai_enhanced":      spec.Confidence > 0.7,
		},
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

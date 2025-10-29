package security

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type APIKeyHandlers struct {
	apiKeyManager *APIKeyManager
	auditLogger   *AuditLogger
}

func NewAPIKeyHandlers(apiKeyManager *APIKeyManager, auditLogger *AuditLogger) *APIKeyHandlers {
	return &APIKeyHandlers{
		apiKeyManager: apiKeyManager,
		auditLogger:   auditLogger,
	}
}

// CreateAPIKeyRequest represents the request to create an API key
type CreateAPIKeyRequest struct {
	Name        string    `json:"name" binding:"required"`
	Permissions []string  `json:"permissions"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

// CreateAPIKey creates a new API key
// @Summary Create API Key
// @Description Create a new API key for authentication
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateAPIKeyRequest true "API Key creation request"
// @Success 201 {object} APIKey
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/auth/api-keys [post]
func (h *APIKeyHandlers) CreateAPIKey(c *gin.Context) {
	var req CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from JWT claims
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Set default permissions if none provided
	if len(req.Permissions) == 0 {
		req.Permissions = []string{
			PermissionReadNetworks,
			PermissionReadWorkflows,
			PermissionReadMetrics,
		}
	}

	apiKey, err := h.apiKeyManager.GenerateAPIKey(
		req.Name,
		userID.(string),
		req.Permissions,
		req.ExpiresAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create API key"})
		return
	}

	// Log the API key creation
	h.auditLogger.LogSecurityEvent(
		userID.(string),
		c.GetString("username"),
		"api_key_created",
		"API key created: "+apiKey.Name,
		c.ClientIP(),
		AuditInfo,
	)

	c.JSON(http.StatusCreated, apiKey)
}

// ListAPIKeys returns all API keys for the authenticated user
// @Summary List API Keys
// @Description Get all API keys for the authenticated user
// @Tags Authentication
// @Produce json
// @Security BearerAuth
// @Success 200 {array} APIKey
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/auth/api-keys [get]
func (h *APIKeyHandlers) ListAPIKeys(c *gin.Context) {
	// Get user ID from JWT claims
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Mock response for demo - in production this would query the database
	mockAPIKeys := []APIKey{
		{
			ID:          "api-key-1",
			Name:        "Production API Key",
			UserID:      userID.(string),
			Permissions: []string{PermissionReadNetworks, PermissionReadWorkflows},
			IsActive:    true,
			CreatedAt:   time.Now().Add(-24 * time.Hour),
			UpdatedAt:   time.Now().Add(-24 * time.Hour),
		},
		{
			ID:          "api-key-2",
			Name:        "Development API Key",
			UserID:      userID.(string),
			Permissions: []string{PermissionReadNetworks, PermissionWriteNetworks},
			IsActive:    true,
			CreatedAt:   time.Now().Add(-72 * time.Hour),
			UpdatedAt:   time.Now().Add(-72 * time.Hour),
		},
	}

	c.JSON(http.StatusOK, mockAPIKeys)
}

// RevokeAPIKey revokes an API key
// @Summary Revoke API Key
// @Description Revoke an API key by ID
// @Tags Authentication
// @Produce json
// @Security BearerAuth
// @Param id path string true "API Key ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/auth/api-keys/{id} [delete]
func (h *APIKeyHandlers) RevokeAPIKey(c *gin.Context) {
	keyID := c.Param("id")
	if keyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "API key ID is required"})
		return
	}

	// Get user ID from JWT claims for audit logging
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	err := h.apiKeyManager.RevokeAPIKey(keyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke API key"})
		return
	}

	// Log the API key revocation
	h.auditLogger.LogSecurityEvent(
		userID.(string),
		c.GetString("username"),
		"api_key_revoked",
		"API key revoked: "+keyID,
		c.ClientIP(),
		AuditInfo,
	)

	c.JSON(http.StatusOK, gin.H{"message": "API key revoked successfully"})
}

// RegisterAPIKeyRoutes registers API key management routes
func RegisterAPIKeyRoutes(router *gin.RouterGroup, apiKeyManager *APIKeyManager, auditLogger *AuditLogger, authMiddleware gin.HandlerFunc) {
	handlers := NewAPIKeyHandlers(apiKeyManager, auditLogger)
	
	apiKeys := router.Group("/api-keys")
	apiKeys.Use(authMiddleware) // Require authentication for all API key operations
	
	apiKeys.POST("", handlers.CreateAPIKey)
	apiKeys.GET("", handlers.ListAPIKeys)
	apiKeys.DELETE("/:id", handlers.RevokeAPIKey)
}
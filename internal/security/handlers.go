package security

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthHandlers struct {
	securityManager *SecurityManager
}

func NewAuthHandlers(securityManager *SecurityManager) *AuthHandlers {
	return &AuthHandlers{
		securityManager: securityManager,
	}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	User         UserInfo  `json:"user"`
}

type UserInfo struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type APIKeyRequest struct {
	Name        string    `json:"name" binding:"required"`
	Permissions []string  `json:"permissions"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

type APIKeyResponse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Key         string     `json:"key"` // Only returned once during creation
	Permissions []string   `json:"permissions"`
	ExpiresAt   *time.Time `json:"expires_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

// @Summary User login
// @Description Authenticate user and return JWT tokens
// @Tags authentication
// @Accept json
// @Produce json
// @Param login body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
func (h *AuthHandlers) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// In production, this would validate against the database
	// For demo purposes, we'll accept admin/admin123 and user/user123
	var userID, email string
	var roles []string

	switch req.Username {
	case "admin":
		if req.Password != "admin123" {
			h.securityManager.auditLogger.LogAuthentication("", req.Username, c.ClientIP(), false, "Invalid password")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		userID = "admin-uuid"
		email = "admin@netorchestrator.com"
		roles = []string{"admin", "user"}

	case "user":
		if req.Password != "user123" {
			h.securityManager.auditLogger.LogAuthentication("", req.Username, c.ClientIP(), false, "Invalid password")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		userID = "user-uuid"
		email = "user@netorchestrator.com"
		roles = []string{"user"}

	default:
		h.securityManager.auditLogger.LogAuthentication("", req.Username, c.ClientIP(), false, "User not found")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate tokens
	accessToken, err := h.securityManager.authService.GenerateToken(userID, req.Username, email, roles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	refreshToken, err := h.securityManager.authService.GenerateToken(userID, req.Username, email, roles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	h.securityManager.auditLogger.LogAuthentication(userID, req.Username, c.ClientIP(), true, "")

	c.JSON(http.StatusOK, LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    86400, // 24 hours in seconds
		User: UserInfo{
			ID:       userID,
			Username: req.Username,
			Email:    email,
			Roles:    roles,
		},
	})
}

// @Summary Refresh token
// @Description Generate new access token using refresh token
// @Tags authentication
// @Accept json
// @Produce json
// @Param refresh body RefreshRequest true "Refresh token"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/refresh [post]
func (h *AuthHandlers) RefreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newToken, err := h.securityManager.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": newToken,
		"token_type":   "Bearer",
		"expires_in":   86400,
	})
}

// @Summary Get current user profile
// @Description Get authenticated user's profile information
// @Tags authentication
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserInfo
// @Failure 401 {object} map[string]string
// @Router /auth/profile [get]
func (h *AuthHandlers) Profile(c *gin.Context) {
	userID := c.GetString("user_id")
	username := c.GetString("username")
	email := c.GetString("user_email")
	roles := c.GetStringSlice("user_roles")

	c.JSON(http.StatusOK, UserInfo{
		ID:       userID,
		Username: username,
		Email:    email,
		Roles:    roles,
	})
}

// @Summary Create API key
// @Description Generate a new API key for programmatic access
// @Tags authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param apikey body APIKeyRequest true "API key configuration"
// @Success 201 {object} APIKeyResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/api-keys [post]
func (h *AuthHandlers) CreateAPIKey(c *gin.Context) {
	var req APIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	username := c.GetString("username")

	// Default permissions if none provided
	if len(req.Permissions) == 0 {
		req.Permissions = []string{PermissionReadNetworks, PermissionReadWorkflows, PermissionReadMetrics}
	}

	apiKey, err := h.securityManager.apiKeyMgr.GenerateAPIKey(req.Name, userID, req.Permissions, req.ExpiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate API key"})
		return
	}

	h.securityManager.auditLogger.LogSecurityEvent(userID, username, "api_key_created", "New API key created: "+req.Name, c.ClientIP(), AuditInfo)

	// Return the API key (this is the only time the raw key is shown)
	c.JSON(http.StatusCreated, APIKeyResponse{
		ID:          apiKey.ID,
		Name:        apiKey.Name,
		Key:         apiKey.Key,
		Permissions: apiKey.Permissions,
		ExpiresAt:   apiKey.ExpiresAt,
		CreatedAt:   apiKey.CreatedAt,
	})
}

// @Summary Logout
// @Description Logout user and invalidate token (placeholder for token blacklisting)
// @Tags authentication
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Router /auth/logout [post]
func (h *AuthHandlers) Logout(c *gin.Context) {
	userID := c.GetString("user_id")
	username := c.GetString("username")

	// In production, this would add the token to a blacklist
	h.securityManager.auditLogger.LogAuthentication(userID, username, c.ClientIP(), true, "User logged out")

	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}
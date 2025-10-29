package security

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type SecurityManager struct {
	authService  *AuthService
	rateLimiter  *RateLimiter
	apiKeyMgr    *APIKeyManager
	auditLogger  *AuditLogger
}

func NewSecurityManager(secretKey string) *SecurityManager {
	authService := NewAuthService(secretKey)
	return &SecurityManager{
		authService: authService,
		rateLimiter: NewRateLimiter(StandardRateLimit, StandardBurst),
		apiKeyMgr:   NewAPIKeyManager(authService),
		auditLogger: NewAuditLogger(),
	}
}

// APIKeyManager returns the API key manager
func (sm *SecurityManager) APIKeyManager() *APIKeyManager {
	return sm.apiKeyMgr
}

// AuditLogger returns the audit logger
func (sm *SecurityManager) AuditLogger() *AuditLogger {
	return sm.auditLogger
}

// JWTAuthMiddleware validates JWT tokens
func (sm *SecurityManager) JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			sm.auditLogger.LogSecurityEvent("", "", "missing_auth_header", "No authorization header provided", c.ClientIP(), AuditWarning)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenParts := strings.SplitN(authHeader, " ", 2)
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			sm.auditLogger.LogSecurityEvent("", "", "invalid_auth_format", "Invalid authorization header format", c.ClientIP(), AuditWarning)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		claims, err := sm.authService.ValidateToken(tokenParts[1])
		if err != nil {
			sm.auditLogger.LogSecurityEvent("", "", "invalid_token", err.Error(), c.ClientIP(), AuditWarning)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Add user info to context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("user_email", claims.Email)
		c.Set("user_roles", claims.Roles)

		c.Next()
	}
}

// APIKeyAuthMiddleware validates API keys as alternative to JWT
func (sm *SecurityManager) APIKeyAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			// Try query parameter as fallback
			apiKey = c.Query("api_key")
		}

		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API key required"})
			c.Abort()
			return
		}

		// Validate API key (this would check against database in production)
		// For demo, we'll accept any key starting with "neto_"
		if !strings.HasPrefix(apiKey, "neto_") {
			sm.auditLogger.LogSecurityEvent("", "", "invalid_api_key", "Invalid API key format", c.ClientIP(), AuditWarning)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
			c.Abort()
			return
		}

		// Mock user for API key (in production, this would be looked up from database)
		c.Set("user_id", "api_user")
		c.Set("username", "API User")
		c.Set("auth_method", "api_key")

		c.Next()
	}
}

// RateLimitMiddleware applies rate limiting
func (sm *SecurityManager) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		limiter := sm.rateLimiter.GetLimiter(c.ClientIP())
		if !limiter.Allow() {
			sm.auditLogger.LogSecurityEvent(
				c.GetString("user_id"),
				c.GetString("username"),
				"rate_limit_exceeded",
				"Rate limit exceeded for IP",
				c.ClientIP(),
				AuditWarning,
			)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":     "Rate limit exceeded",
				"retry_after": "60s",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// StrictRateLimitMiddleware applies stricter rate limiting for sensitive endpoints
func (sm *SecurityManager) StrictRateLimitMiddleware() gin.HandlerFunc {
	strictLimiter := NewRateLimiter(StrictRateLimit, StrictBurst)
	
	return func(c *gin.Context) {
		limiter := strictLimiter.GetLimiter(c.ClientIP())
		if !limiter.Allow() {
			sm.auditLogger.LogSecurityEvent(
				c.GetString("user_id"),
				c.GetString("username"),
				"strict_rate_limit_exceeded",
				"Strict rate limit exceeded for sensitive endpoint",
				c.ClientIP(),
				AuditError,
			)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"retry_after": "60s",
				"message":     "This endpoint has strict rate limiting due to security sensitivity",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// AuditMiddleware logs all API requests for audit purposes
func (sm *SecurityManager) AuditMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Log the request for audit purposes
		sm.auditLogger.LogAPIAccess(
			param.Keys["user_id"].(string),
			param.Keys["username"].(string),
			param.Method,
			param.Path,
			param.ClientIP,
			param.Keys["user_agent"].(string),
			param.StatusCode,
			param.Latency,
		)
		
		// Return empty string since we're doing custom logging
		return ""
	})
}

// RequireRole middleware ensures user has required role
func (sm *SecurityManager) RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoles, exists := c.Get("user_roles")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "No roles found"})
			c.Abort()
			return
		}

		roles, ok := userRoles.([]string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid role format"})
			c.Abort()
			return
		}

		hasRole := false
		for _, role := range roles {
			if role == requiredRole || role == "admin" { // admin has all permissions
				hasRole = true
				break
			}
		}

		if !hasRole {
			sm.auditLogger.LogSecurityEvent(
				c.GetString("user_id"),
				c.GetString("username"),
				"insufficient_permissions",
				"User lacks required role: "+requiredRole,
				c.ClientIP(),
				AuditWarning,
			)
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SecurityHeadersMiddleware adds security headers
func (sm *SecurityManager) SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Next()
	}
}
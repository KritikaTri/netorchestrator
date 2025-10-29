package security

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

type APIKey struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Key         string    `json:"key" gorm:"unique;not null"`
	HashedKey   string    `json:"-" gorm:"not null"`
	UserID      string    `json:"user_id" gorm:"not null"`
	Permissions []string  `json:"permissions" gorm:"type:json"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	ExpiresAt   *time.Time `json:"expires_at"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type APIKeyManager struct {
	authService *AuthService
}

func NewAPIKeyManager(authService *AuthService) *APIKeyManager {
	return &APIKeyManager{
		authService: authService,
	}
}

// GenerateAPIKey creates a new API key
func (m *APIKeyManager) GenerateAPIKey(name, userID string, permissions []string, expiresAt *time.Time) (*APIKey, error) {
	// Generate random key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	key := "neto_" + hex.EncodeToString(keyBytes)
	
	// Hash the key for storage
	hashedKey, err := m.authService.HashPassword(key)
	if err != nil {
		return nil, fmt.Errorf("failed to hash API key: %w", err)
	}

	apiKey := &APIKey{
		ID:          generateID(),
		Name:        name,
		Key:         key, // This will be returned once, then removed
		HashedKey:   hashedKey,
		UserID:      userID,
		Permissions: permissions,
		IsActive:    true,
		ExpiresAt:   expiresAt,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return apiKey, nil
}

// ValidateAPIKey checks if an API key is valid
func (m *APIKeyManager) ValidateAPIKey(key string) (*APIKey, error) {
	// In a real implementation, this would query the database
	// For now, we'll return a mock validation
	return nil, fmt.Errorf("API key validation not implemented - requires database integration")
}

// RevokeAPIKey deactivates an API key
func (m *APIKeyManager) RevokeAPIKey(keyID string) error {
	// Implementation would update the database to set IsActive = false
	return fmt.Errorf("API key revocation not implemented - requires database integration")
}

func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// Common API key permissions
var (
	PermissionReadNetworks    = "networks:read"
	PermissionWriteNetworks   = "networks:write"
	PermissionDeleteNetworks  = "networks:delete"
	PermissionReadWorkflows   = "workflows:read"
	PermissionWriteWorkflows  = "workflows:write"
	PermissionExecuteWorkflows = "workflows:execute"
	PermissionReadMetrics     = "metrics:read"
	PermissionManageUsers     = "users:manage"
	PermissionSystemAdmin     = "system:admin"
)
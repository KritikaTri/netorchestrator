package security

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ErrInvalidAPIKey is returned for unknown, revoked or expired keys. The
// message is deliberately uniform so callers cannot probe which case applied.
var ErrInvalidAPIKey = errors.New("invalid api key")

func digestKey(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}

type APIKey struct {
	ID          string     `json:"id" gorm:"primaryKey"`
	Name        string     `json:"name" gorm:"not null"`
	Key         string     `json:"key" gorm:"unique;not null"`
	HashedKey   string     `json:"-" gorm:"not null"`
	UserID      string     `json:"user_id" gorm:"not null"`
	Permissions []string   `json:"permissions" gorm:"type:json"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	ExpiresAt   *time.Time `json:"expires_at"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type APIKeyManager struct {
	authService *AuthService

	mu     sync.RWMutex
	byHash map[string]*APIKey // sha256(key) -> record (plaintext never stored)
	byID   map[string]*APIKey
}

func NewAPIKeyManager(authService *AuthService) *APIKeyManager {
	return &APIKeyManager{
		authService: authService,
		byHash:      map[string]*APIKey{},
		byID:        map[string]*APIKey{},
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

	// API keys are 256-bit random values, so a fast SHA-256 digest is enough
	// and lets us look keys up by digest (bcrypt could not be indexed).
	hashedKey := digestKey(key)

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

	// Keep only the digest server-side; the plaintext is returned once to the caller.
	stored := *apiKey
	stored.Key = ""
	m.mu.Lock()
	m.byHash[hashedKey] = &stored
	m.byID[stored.ID] = &stored
	m.mu.Unlock()

	return apiKey, nil
}

// ValidateAPIKey returns the key record if key is known, active and unexpired.
func (m *APIKeyManager) ValidateAPIKey(key string) (*APIKey, error) {
	d := digestKey(key)
	m.mu.Lock()
	defer m.mu.Unlock()
	rec, ok := m.byHash[d]
	// constant-time compare as defence in depth against map-lookup timing
	if !ok || subtle.ConstantTimeCompare([]byte(rec.HashedKey), []byte(d)) != 1 {
		return nil, ErrInvalidAPIKey
	}
	now := time.Now()
	if !rec.IsActive || (rec.ExpiresAt != nil && !now.Before(*rec.ExpiresAt)) {
		return nil, ErrInvalidAPIKey
	}
	rec.LastUsedAt = &now
	cp := *rec
	return &cp, nil
}

// RevokeAPIKey deactivates an API key by ID.
func (m *APIKeyManager) RevokeAPIKey(keyID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	rec, ok := m.byID[keyID]
	if !ok {
		return fmt.Errorf("api key %q not found", keyID)
	}
	rec.IsActive = false
	rec.UpdatedAt = time.Now()
	return nil
}

func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// Common API key permissions
var (
	PermissionReadNetworks     = "networks:read"
	PermissionWriteNetworks    = "networks:write"
	PermissionDeleteNetworks   = "networks:delete"
	PermissionReadWorkflows    = "workflows:read"
	PermissionWriteWorkflows   = "workflows:write"
	PermissionExecuteWorkflows = "workflows:execute"
	PermissionReadMetrics      = "metrics:read"
	PermissionManageUsers      = "users:manage"
	PermissionSystemAdmin      = "system:admin"
)

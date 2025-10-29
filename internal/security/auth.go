package security

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

type AuthService struct {
	secretKey     []byte
	tokenExpiry   time.Duration
	refreshExpiry time.Duration
}

func NewAuthService(secretKey string) *AuthService {
	return &AuthService{
		secretKey:     []byte(secretKey),
		tokenExpiry:   time.Hour * 24,     // 24 hours
		refreshExpiry: time.Hour * 24 * 7, // 7 days
	}
}

// GenerateToken creates a new JWT token for a user
func (a *AuthService) GenerateToken(userID, username, email string, roles []string) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.tokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "netorchestrator",
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.secretKey)
}

// ValidateToken parses and validates a JWT token
func (a *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return a.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// HashPassword creates a bcrypt hash of the password
func (a *AuthService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

// VerifyPassword checks if password matches the hash
func (a *AuthService) VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// RefreshToken generates a new token from a valid refresh token
func (a *AuthService) RefreshToken(refreshToken string) (string, error) {
	claims, err := a.ValidateToken(refreshToken)
	if err != nil {
		return "", err
	}

	// Generate new access token with same user info
	return a.GenerateToken(claims.UserID, claims.Username, claims.Email, claims.Roles)
}
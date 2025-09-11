package infrastructure

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"go-distributed-hashcat/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTService handles JWT token operations
type JWTService struct {
	secretKey     []byte
	tokenDuration time.Duration
}

// NewJWTService creates a new JWT service
func NewJWTService() *JWTService {
	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		secretKey = "your-secret-key-change-this-in-production" // Default fallback
	}

	// Get token duration from environment variable (in hours)
	tokenDurationHours := 24 // Default 24 hours
	if hoursStr := os.Getenv("JWT_TOKEN_DURATION_HOURS"); hoursStr != "" {
		if hours, err := strconv.Atoi(hoursStr); err == nil {
			tokenDurationHours = hours
		}
	}

	return &JWTService{
		secretKey:     []byte(secretKey),
		tokenDuration: time.Duration(tokenDurationHours) * time.Hour,
	}
}

// GenerateToken generates a JWT token for a user
func (j *JWTService) GenerateToken(user *domain.User) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(j.tokenDuration)

	claims := &domain.JWTClaims{
		UserID:   user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "go-distributed-hashcat",
			Subject:   user.ID.String(),
			Audience:  []string{"hashcat-users"},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secretKey)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, expiresAt, nil
}

// ValidateToken validates a JWT token and returns the claims
func (j *JWTService) ValidateToken(tokenString string) (*domain.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &domain.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*domain.JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Check if token is expired
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("token has expired")
	}

	return claims, nil
}

// RefreshToken generates a new token for the same user
func (j *JWTService) RefreshToken(tokenString string) (string, time.Time, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("invalid token for refresh: %w", err)
	}

	// Parse user ID from string to UUID
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("invalid user ID in token: %w", err)
	}

	// Create a new user object from claims for token generation
	user := &domain.User{
		ID:       userID,
		Username: claims.Username,
		Email:    claims.Email,
		Role:     claims.Role,
	}

	return j.GenerateToken(user)
}

// GetTokenDuration returns the configured token duration
func (j *JWTService) GetTokenDuration() time.Duration {
	return j.tokenDuration
}

package usecase

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"go-distributed-hashcat/internal/domain"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthUsecase interface {
	Login(username, password string) (*domain.LoginResponse, error)
	Register(req *domain.CreateUserRequest) (*domain.User, error)
	ValidateToken(token string) (*domain.User, error)
	ChangePassword(userID uuid.UUID, req *domain.ChangePasswordRequest) error
	GetCurrentUser(userID uuid.UUID) (*domain.User, error)
}

type authUsecase struct {
	userRepo domain.UserRepository
	tokens   map[string]tokenData // Simple in-memory token store
}

type tokenData struct {
	UserID    uuid.UUID
	ExpiresAt time.Time
}

func NewAuthUsecase(userRepo domain.UserRepository) AuthUsecase {
	return &authUsecase{
		userRepo: userRepo,
		tokens:   make(map[string]tokenData),
	}
}

func (a *authUsecase) Login(username, password string) (*domain.LoginResponse, error) {
	// Get user by username
	user, err := a.userRepo.GetByUsername(username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if !user.IsActive {
		return nil, errors.New("account is disabled")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Generate token
	token, err := a.generateToken()
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	// Store token with expiration (24 hours)
	expiresAt := time.Now().Add(24 * time.Hour)
	a.tokens[token] = tokenData{
		UserID:    user.ID,
		ExpiresAt: expiresAt,
	}

	// Update last login
	if err := a.userRepo.UpdateLastLogin(user.ID); err != nil {
		// Log error but don't fail login
	}

	return &domain.LoginResponse{
		Token:     token,
		User:      *user,
		ExpiresAt: expiresAt,
	}, nil
}

func (a *authUsecase) Register(req *domain.CreateUserRequest) (*domain.User, error) {
	// Check if username already exists
	if existingUser, _ := a.userRepo.GetByUsername(req.Username); existingUser != nil {
		return nil, errors.New("username already exists")
	}

	// Check if email already exists
	if existingUser, _ := a.userRepo.GetByEmail(req.Email); existingUser != nil {
		return nil, errors.New("email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Create user
	user := &domain.User{
		ID:        uuid.New(),
		Username:  req.Username,
		Email:     req.Email,
		Password:  string(hashedPassword),
		Role:      req.Role,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := a.userRepo.Create(user); err != nil {
		return nil, errors.New("failed to create user")
	}

	// Don't return password in response
	user.Password = ""
	return user, nil
}

func (a *authUsecase) ValidateToken(token string) (*domain.User, error) {
	// Check if token exists
	tokenData, exists := a.tokens[token]
	if !exists {
		return nil, errors.New("invalid token")
	}

	// Check if token is expired
	if time.Now().After(tokenData.ExpiresAt) {
		delete(a.tokens, token)
		return nil, errors.New("token expired")
	}

	// Get user
	user, err := a.userRepo.GetByID(tokenData.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if !user.IsActive {
		return nil, errors.New("account is disabled")
	}

	// Don't return password
	user.Password = ""
	return user, nil
}

func (a *authUsecase) ChangePassword(userID uuid.UUID, req *domain.ChangePasswordRequest) error {
	// Get user
	user, err := a.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		return errors.New("invalid old password")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash new password")
	}

	// Update password
	user.Password = string(hashedPassword)
	user.UpdatedAt = time.Now()

	return a.userRepo.Update(user)
}

func (a *authUsecase) GetCurrentUser(userID uuid.UUID) (*domain.User, error) {
	user, err := a.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Don't return password
	user.Password = ""
	return user, nil
}

func (a *authUsecase) generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Cleanup expired tokens periodically
func (a *authUsecase) CleanupExpiredTokens() {
	for token, data := range a.tokens {
		if time.Now().After(data.ExpiresAt) {
			delete(a.tokens, token)
		}
	}
}

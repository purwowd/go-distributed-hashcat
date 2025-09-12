package usecase

import (
	"context"
	"fmt"
	"time"

	"go-distributed-hashcat/internal/domain"
	"go-distributed-hashcat/internal/infrastructure"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type authUsecase struct {
	userRepo   domain.UserRepository
	jwtService *infrastructure.JWTService
}

// NewAuthUsecase creates a new authentication usecase
func NewAuthUsecase(userRepo domain.UserRepository, jwtService *infrastructure.JWTService) domain.AuthUsecase {
	return &authUsecase{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

// Login authenticates a user and returns a JWT token
func (u *authUsecase) Login(ctx context.Context, req *domain.LoginRequest) (*domain.LoginResponse, error) {
	// Get user by username
	user, err := u.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		// Check if it's a "not found" error vs other database errors
		if _, ok := err.(*domain.UserNotFoundError); ok {
			return nil, &domain.InvalidCredentialsError{}
		}
		return nil, &domain.InvalidCredentialsError{}
	}

	// Check if user is active
	if !user.IsActive {
		return nil, &domain.AuthenticationError{Message: "account is deactivated"}
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, &domain.InvalidCredentialsError{}
	}

	// Generate JWT token
	token, expiresAt, err := u.jwtService.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Update last login time
	now := time.Now()
	user.LastLogin = &now
	user.UpdatedAt = now
	err = u.userRepo.UpdateLastLogin(ctx, user.ID)
	if err != nil {
		// Log error but don't fail the login
		// In production, you might want to use a proper logger
		fmt.Printf("Warning: failed to update last login time for user %s: %v\n", user.Username, err)
	}

	return &domain.LoginResponse{
		Token:     token,
		User:      *user,
		ExpiresAt: expiresAt,
	}, nil
}

// Logout invalidates a JWT token (in a simple implementation, we just validate the token)
func (u *authUsecase) Logout(ctx context.Context, token string) error {
	// Validate token to ensure it's valid before logout
	_, err := u.jwtService.ValidateToken(token)
	if err != nil {
		return &domain.AuthenticationError{Message: "invalid token"}
	}

	// In a more sophisticated implementation, you might:
	// 1. Add the token to a blacklist
	// 2. Store logout time in database
	// 3. Implement token revocation

	// For now, we just validate the token
	return nil
}

// ValidateToken validates a JWT token and returns the claims
func (u *authUsecase) ValidateToken(ctx context.Context, token string) (*domain.JWTClaims, error) {
	claims, err := u.jwtService.ValidateToken(token)
	if err != nil {
		return nil, &domain.AuthenticationError{Message: "invalid token"}
	}

	return claims, nil
}

// RefreshToken generates a new token for the same user
func (u *authUsecase) RefreshToken(ctx context.Context, token string) (*domain.LoginResponse, error) {
	// Validate the current token
	claims, err := u.jwtService.ValidateToken(token)
	if err != nil {
		return nil, &domain.AuthenticationError{Message: "invalid token"}
	}

	// Get user to ensure they still exist and are active
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, &domain.AuthenticationError{Message: "invalid user ID in token"}
	}

	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, &domain.AuthenticationError{Message: "user not found"}
	}

	if !user.IsActive {
		return nil, &domain.AuthenticationError{Message: "account is deactivated"}
	}

	// Generate new token
	newToken, expiresAt, err := u.jwtService.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new token: %w", err)
	}

	return &domain.LoginResponse{
		Token:     newToken,
		User:      *user,
		ExpiresAt: expiresAt,
	}, nil
}

// CreateUser creates a new user
func (u *authUsecase) CreateUser(ctx context.Context, req *domain.CreateUserRequest) (*domain.User, error) {
	// Check if username already exists
	_, err := u.userRepo.GetByUsername(ctx, req.Username)
	if err == nil {
		return nil, &domain.UserAlreadyExistsError{Username: req.Username, Email: req.Email}
	}

	// Check if email already exists
	_, err = u.userRepo.GetByEmail(ctx, req.Email)
	if err == nil {
		return nil, &domain.UserAlreadyExistsError{Username: req.Username, Email: req.Email}
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Set default role if not provided
	role := req.Role
	if role == "" {
		role = "user"
	}

	// Create user
	user := &domain.User{
		ID:        uuid.New(),
		Username:  req.Username,
		Email:     req.Email,
		Password:  string(hashedPassword),
		Role:      role,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = u.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUser retrieves a user by ID
func (u *authUsecase) GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return u.userRepo.GetByID(ctx, id)
}

// UpdateUser updates a user
func (u *authUsecase) UpdateUser(ctx context.Context, id uuid.UUID, req *domain.UpdateUserRequest) (*domain.User, error) {
	// Get existing user
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Username != nil {
		// Check if new username is already taken by another user
		existingUser, err := u.userRepo.GetByUsername(ctx, *req.Username)
		if err == nil && existingUser.ID != id {
			return nil, &domain.UserAlreadyExistsError{Username: *req.Username, Email: user.Email}
		}
		user.Username = *req.Username
	}

	if req.Email != nil {
		// Check if new email is already taken by another user
		existingUser, err := u.userRepo.GetByEmail(ctx, *req.Email)
		if err == nil && existingUser.ID != id {
			return nil, &domain.UserAlreadyExistsError{Username: user.Username, Email: *req.Email}
		}
		user.Email = *req.Email
	}

	if req.Password != nil {
		// Hash new password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		user.Password = string(hashedPassword)
	}

	if req.Role != nil {
		user.Role = *req.Role
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	user.UpdatedAt = time.Now()

	err = u.userRepo.Update(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// DeleteUser deletes a user
func (u *authUsecase) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return u.userRepo.Delete(ctx, id)
}

// GetAllUsers retrieves all users
func (u *authUsecase) GetAllUsers(ctx context.Context) ([]domain.User, error) {
	return u.userRepo.GetAll(ctx)
}

// CheckUsernameExists checks if a username exists in the database
func (u *authUsecase) CheckUsernameExists(ctx context.Context, username string) (bool, error) {
	_, err := u.userRepo.GetByUsername(ctx, username)
	if err != nil {
		// If user not found, return false (username doesn't exist)
		if _, ok := err.(*domain.UserNotFoundError); ok {
			return false, nil
		}
		// For other errors, return the error
		return false, err
	}
	// User found, username exists
	return true, nil
}

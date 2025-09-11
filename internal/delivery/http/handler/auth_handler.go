package handler

import (
	"net/http"
	"strconv"

	"go-distributed-hashcat/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	authUsecase domain.AuthUsecase
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authUsecase domain.AuthUsecase) *AuthHandler {
	return &AuthHandler{
		authUsecase: authUsecase,
	}
}

// Login handles user login
// @Summary Login user
// @Description Authenticate user with username/email and password
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body domain.LoginRequest true "Login credentials"
// @Success 200 {object} domain.LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	response, err := h.authUsecase.Login(c.Request.Context(), &req)
	if err != nil {
		switch err.(type) {
		case *domain.InvalidCredentialsError:
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid username or password",
				"code":  "INVALID_CREDENTIALS",
			})
		case *domain.AuthenticationError:
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
				"code":  "AUTH_ERROR",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
				"code":  "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, response)
}

// Logout handles user logout
// @Summary Logout user
// @Description Invalidate user token
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body domain.LogoutRequest true "Logout request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req domain.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	err := h.authUsecase.Logout(c.Request.Context(), req.Token)
	if err != nil {
		switch err.(type) {
		case *domain.AuthenticationError:
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully logged out",
	})
}

// RefreshToken handles token refresh
// @Summary Refresh token
// @Description Generate new token using existing valid token
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body domain.LogoutRequest true "Token to refresh"
// @Success 200 {object} domain.LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req domain.LogoutRequest // Reusing LogoutRequest struct for token
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	response, err := h.authUsecase.RefreshToken(c.Request.Context(), req.Token)
	if err != nil {
		switch err.(type) {
		case *domain.AuthenticationError:
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
		}
		return
	}

	c.JSON(http.StatusOK, response)
}

// ValidateToken validates a token
// @Summary Validate token
// @Description Validate if token is still valid
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body domain.LogoutRequest true "Token to validate"
// @Success 200 {object} domain.JWTClaims
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/auth/validate [post]
func (h *AuthHandler) ValidateToken(c *gin.Context) {
	var req domain.LogoutRequest // Reusing LogoutRequest struct for token
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	claims, err := h.authUsecase.ValidateToken(c.Request.Context(), req.Token)
	if err != nil {
		switch err.(type) {
		case *domain.AuthenticationError:
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
		}
		return
	}

	c.JSON(http.StatusOK, claims)
}

// CreateUser handles user creation (admin only)
// @Summary Create user
// @Description Create a new user account
// @Tags users
// @Accept json
// @Produce json
// @Param request body domain.CreateUserRequest true "User data"
// @Success 201 {object} domain.User
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/v1/users [post]
func (h *AuthHandler) CreateUser(c *gin.Context) {
	var req domain.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	user, err := h.authUsecase.CreateUser(c.Request.Context(), &req)
	if err != nil {
		switch err.(type) {
		case *domain.UserAlreadyExistsError:
			c.JSON(http.StatusConflict, gin.H{
				"error": err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, user)
}

// GetUser handles getting user by ID
// @Summary Get user
// @Description Get user information by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} domain.User
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/users/{id} [get]
func (h *AuthHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	user, err := h.authUsecase.GetUser(c.Request.Context(), id)
	if err != nil {
		switch err.(type) {
		case *domain.UserNotFoundError:
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateUser handles updating user
// @Summary Update user
// @Description Update user information
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body domain.UpdateUserRequest true "User update data"
// @Success 200 {object} domain.User
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/v1/users/{id} [put]
func (h *AuthHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	var req domain.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	user, err := h.authUsecase.UpdateUser(c.Request.Context(), id, &req)
	if err != nil {
		switch err.(type) {
		case *domain.UserNotFoundError:
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
		case *domain.UserAlreadyExistsError:
			c.JSON(http.StatusConflict, gin.H{
				"error": err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

// DeleteUser handles deleting user
// @Summary Delete user
// @Description Delete user account
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/users/{id} [delete]
func (h *AuthHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	err = h.authUsecase.DeleteUser(c.Request.Context(), id)
	if err != nil {
		switch err.(type) {
		case *domain.UserNotFoundError:
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}

// GetAllUsers handles getting all users
// @Summary Get all users
// @Description Get list of all users
// @Tags users
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {array} domain.User
// @Failure 500 {object} map[string]string
// @Router /api/v1/users [get]
func (h *AuthHandler) GetAllUsers(c *gin.Context) {
	// Parse pagination parameters
	page := 1
	limit := 10

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	users, err := h.authUsecase.GetAllUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
		return
	}

	// Simple pagination (in production, you might want to implement this in the repository)
	start := (page - 1) * limit
	end := start + limit

	if start >= len(users) {
		users = []domain.User{}
	} else if end > len(users) {
		users = users[start:]
	} else {
		users = users[start:end]
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": len(users),
		},
	})
}

// CheckUsernameExists handles checking if a username exists
// @Summary Check username exists
// @Description Check if a username exists in the database
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body map[string]string true "Username to check"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} map[string]string
// @Router /api/v1/auth/check-username [post]
func (h *AuthHandler) CheckUsernameExists(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	exists, err := h.authUsecase.CheckUsernameExists(c.Request.Context(), req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"exists": exists,
	})
}

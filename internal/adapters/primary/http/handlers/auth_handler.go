package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/notinoteapp/internal/adapters/primary/http/dto"
	appdto "github.com/yourusername/notinoteapp/internal/application/dto"
	"github.com/yourusername/notinoteapp/internal/application/services"
	"github.com/yourusername/notinoteapp/internal/core/domain"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handles user registration with email/password
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}

	// Register user
	authResp, err := h.authService.Register(c.Request.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		status := http.StatusInternalServerError
		message := "Failed to register user"

		switch err {
		case domain.ErrUserAlreadyExists:
			status = http.StatusConflict
			message = "User with this email already exists"
		case domain.ErrInvalidEmail, domain.ErrInvalidName, domain.ErrPasswordTooWeak:
			status = http.StatusBadRequest
			message = err.Error()
		}

		c.JSON(status, dto.ErrorResponse{
			Success: false,
			Error:   message,
		})
		return
	}

	// Build response
	resp := h.buildAuthResponse(authResp)
	c.JSON(http.StatusCreated, resp)
}

// Login handles user login with email/password
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}

	// Login user
	authResp, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		status := http.StatusInternalServerError
		message := "Failed to login"

		switch err {
		case domain.ErrInvalidCredentials:
			status = http.StatusUnauthorized
			message = "Invalid email or password"
		case domain.ErrUserInactive:
			status = http.StatusForbidden
			message = "Account is inactive"
		}

		c.JSON(status, dto.ErrorResponse{
			Success: false,
			Error:   message,
		})
		return
	}

	// Build response
	resp := h.buildAuthResponse(authResp)
	c.JSON(http.StatusOK, resp)
}

// RefreshToken handles token refresh
// POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}

	// Refresh token
	authResp, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		status := http.StatusUnauthorized
		message := "Invalid or expired refresh token"

		c.JSON(status, dto.ErrorResponse{
			Success: false,
			Error:   message,
		})
		return
	}

	// Build response
	resp := h.buildAuthResponse(authResp)
	c.JSON(http.StatusOK, resp)
}

// Logout handles user logout
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// In a stateless JWT system, logout is handled client-side by removing the token
	// For additional security, you could implement token blacklisting using Redis

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logged out successfully",
	})
}

// GetCurrentUser returns the current authenticated user's profile
// GET /api/v1/me
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Success: false,
			Error:   "User not authenticated",
		})
		return
	}

	// Get user from database
	user, err := h.authService.GetUserByID(c.Request.Context(), userID.(int64))
	if err != nil {
		status := http.StatusInternalServerError
		message := "Failed to get user profile"

		if err == domain.ErrUserNotFound {
			status = http.StatusNotFound
			message = "User not found"
		}

		c.JSON(status, dto.ErrorResponse{
			Success: false,
			Error:   message,
		})
		return
	}

	// Build response
	c.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
		Data:    dto.NewUserResponse(user),
	})
}

// VerifyGoogleToken verifies Google ID token from frontend
// POST /api/v1/auth/google/verify
func (h *AuthHandler) VerifyGoogleToken(c *gin.Context) {
	var req dto.GoogleTokenRequest
	fmt.Println("test", req)
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}

	// Verify token and authenticate user
	authResp, err := h.authService.VerifyGoogleToken(c.Request.Context(), req.IDToken)
	if err != nil {
		status := http.StatusUnauthorized
		message := "Failed to verify Google token"

		switch err {
		case domain.ErrOAuthUserInfo:
			message = "Failed to get user info from Google"
		case domain.ErrUserInactive:
			status = http.StatusForbidden
			message = "Account is inactive"
		}

		c.JSON(status, dto.ErrorResponse{
			Success: false,
			Error:   message,
		})
		return
	}

	// Build response
	resp := h.buildAuthResponse(authResp)
	c.JSON(http.StatusOK, resp)
}

// VerifyFacebookToken verifies Facebook access token from frontend
// POST /api/v1/auth/facebook/verify
func (h *AuthHandler) VerifyFacebookToken(c *gin.Context) {
	var req dto.FacebookTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}

	// Verify token and authenticate user
	authResp, err := h.authService.VerifyFacebookToken(c.Request.Context(), req.AccessToken)
	if err != nil {
		status := http.StatusUnauthorized
		message := "Failed to verify Facebook token"

		switch err {
		case domain.ErrOAuthUserInfo:
			message = "Failed to get user info from Facebook"
		case domain.ErrUserInactive:
			status = http.StatusForbidden
			message = "Account is inactive"
		}

		c.JSON(status, dto.ErrorResponse{
			Success: false,
			Error:   message,
		})
		return
	}

	// Build response
	resp := h.buildAuthResponse(authResp)
	c.JSON(http.StatusOK, resp)
}

// buildAuthResponse builds the authentication response
func (h *AuthHandler) buildAuthResponse(authResp *appdto.AuthResponse) dto.AuthResponse {
	// 24 hours in seconds
	expiresIn := 86400
	return dto.NewAuthResponse(authResp, expiresIn)
}

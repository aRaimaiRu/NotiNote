package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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

// RegisterRequest represents the registration request body
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required,min=1,max=255"`
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RefreshTokenRequest represents the refresh token request body
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    *struct {
		User struct {
			ID        int64                `json:"id"`
			Email     string               `json:"email"`
			Name      string               `json:"name"`
			Provider  domain.AuthProvider  `json:"provider"`
			AvatarURL string               `json:"avatar_url,omitempty"`
			CreatedAt time.Time            `json:"created_at"`
		} `json:"user"`
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"` // seconds
	} `json:"data,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// Register handles user registration with email/password
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
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

		c.JSON(status, ErrorResponse{
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
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
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

		c.JSON(status, ErrorResponse{
			Success: false,
			Error:   message,
		})
		return
	}

	// Build response
	resp := h.buildAuthResponse(authResp)
	c.JSON(http.StatusOK, resp)
}

// GoogleLogin initiates Google OAuth login
// GET /api/v1/auth/google
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	authURL, err := h.authService.GetOAuthURL(c.Request.Context(), domain.AuthProviderGoogle)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Error:   "Failed to generate Google login URL",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"auth_url": authURL,
		},
	})
}

// GoogleCallback handles Google OAuth callback
// GET /api/v1/auth/google/callback
func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Missing code or state parameter",
		})
		return
	}

	// Handle OAuth callback
	authResp, err := h.authService.HandleOAuthCallback(c.Request.Context(), domain.AuthProviderGoogle, code, state)
	if err != nil {
		status := http.StatusInternalServerError
		message := "Failed to authenticate with Google"

		switch err {
		case domain.ErrOAuthStateMismatch:
			status = http.StatusBadRequest
			message = "Invalid state parameter - possible CSRF attack"
		case domain.ErrUserInactive:
			status = http.StatusForbidden
			message = "Account is inactive"
		}

		c.JSON(status, ErrorResponse{
			Success: false,
			Error:   message,
		})
		return
	}

	// Build response
	resp := h.buildAuthResponse(authResp)
	c.JSON(http.StatusOK, resp)
}

// FacebookLogin initiates Facebook OAuth login
// GET /api/v1/auth/facebook
func (h *AuthHandler) FacebookLogin(c *gin.Context) {
	authURL, err := h.authService.GetOAuthURL(c.Request.Context(), domain.AuthProviderFacebook)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Error:   "Failed to generate Facebook login URL",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"auth_url": authURL,
		},
	})
}

// FacebookCallback handles Facebook OAuth callback
// GET /api/v1/auth/facebook/callback
func (h *AuthHandler) FacebookCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Error:   "Missing code or state parameter",
		})
		return
	}

	// Handle OAuth callback
	authResp, err := h.authService.HandleOAuthCallback(c.Request.Context(), domain.AuthProviderFacebook, code, state)
	if err != nil {
		status := http.StatusInternalServerError
		message := "Failed to authenticate with Facebook"

		switch err {
		case domain.ErrOAuthStateMismatch:
			status = http.StatusBadRequest
			message = "Invalid state parameter - possible CSRF attack"
		case domain.ErrUserInactive:
			status = http.StatusForbidden
			message = "Account is inactive"
		}

		c.JSON(status, ErrorResponse{
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
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
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

		c.JSON(status, ErrorResponse{
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

// buildAuthResponse builds the authentication response
func (h *AuthHandler) buildAuthResponse(authResp *services.AuthResponse) AuthResponse {
	resp := AuthResponse{
		Success: true,
		Data: &struct {
			User struct {
				ID        int64               `json:"id"`
				Email     string              `json:"email"`
				Name      string              `json:"name"`
				Provider  domain.AuthProvider `json:"provider"`
				AvatarURL string              `json:"avatar_url,omitempty"`
				CreatedAt time.Time           `json:"created_at"`
			} `json:"user"`
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
			TokenType    string `json:"token_type"`
			ExpiresIn    int    `json:"expires_in"`
		}{},
	}

	resp.Data.User.ID = authResp.User.ID
	resp.Data.User.Email = authResp.User.Email
	resp.Data.User.Name = authResp.User.Name
	resp.Data.User.Provider = authResp.User.Provider
	resp.Data.User.AvatarURL = authResp.User.AvatarURL
	resp.Data.User.CreatedAt = authResp.User.CreatedAt

	resp.Data.AccessToken = authResp.AccessToken
	resp.Data.RefreshToken = authResp.RefreshToken
	resp.Data.TokenType = "Bearer"
	resp.Data.ExpiresIn = 86400 // 24 hours in seconds

	return resp
}

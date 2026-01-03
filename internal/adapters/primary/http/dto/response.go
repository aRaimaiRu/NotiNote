package dto

import (
	"time"

	appdto "github.com/yourusername/notinoteapp/internal/application/dto"
	"github.com/yourusername/notinoteapp/internal/core/domain"
)

// AuthResponse represents the authentication response sent to clients
type AuthResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    *struct {
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
		ExpiresIn    int    `json:"expires_in"` // seconds
	} `json:"data,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// SuccessResponse represents a generic success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// UserResponse represents a user profile response
type UserResponse struct {
	ID        int64               `json:"id"`
	Email     string              `json:"email"`
	Name      string              `json:"name"`
	Provider  domain.AuthProvider `json:"provider"`
	AvatarURL string              `json:"avatar_url,omitempty"`
	IsActive  bool                `json:"is_active"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

// NewAuthResponse creates an HTTP AuthResponse from application layer AuthResponse
func NewAuthResponse(appResp *appdto.AuthResponse, expiresIn int) AuthResponse {
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

	resp.Data.User.ID = appResp.User.ID
	resp.Data.User.Email = appResp.User.Email
	resp.Data.User.Name = appResp.User.Name
	resp.Data.User.Provider = appResp.User.Provider
	resp.Data.User.AvatarURL = appResp.User.AvatarURL
	resp.Data.User.CreatedAt = appResp.User.CreatedAt

	resp.Data.AccessToken = appResp.AccessToken
	resp.Data.RefreshToken = appResp.RefreshToken
	resp.Data.TokenType = "Bearer"
	resp.Data.ExpiresIn = expiresIn

	return resp
}

// NewUserResponse creates a UserResponse from domain User
func NewUserResponse(user *domain.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Provider:  user.Provider,
		AvatarURL: user.AvatarURL,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

package dto

import (
	"time"

	"github.com/yourusername/notinoteapp/internal/core/domain"
)

// AuthResponse represents the authentication response returned by the service layer
// This is used by all authentication operations (login, register, OAuth, refresh)
type AuthResponse struct {
	User         *UserDTO `json:"user"`
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresAt    int64    `json:"expires_at"` // Unix timestamp
}

// UserDTO represents user data returned in responses
type UserDTO struct {
	ID        int64                `json:"id"`
	Email     string               `json:"email"`
	Name      string               `json:"name"`
	Provider  domain.AuthProvider  `json:"provider"`
	AvatarURL string               `json:"avatar_url,omitempty"`
	IsActive  bool                 `json:"is_active"`
	CreatedAt time.Time            `json:"created_at"`
	UpdatedAt time.Time            `json:"updated_at"`
}

// LoginInput represents the input for login operation
type LoginInput struct {
	Email    string
	Password string
}

// RegisterInput represents the input for registration operation
type RegisterInput struct {
	Email    string
	Password string
	Name     string
}

// RefreshTokenInput represents the input for token refresh operation
type RefreshTokenInput struct {
	RefreshToken string
}

// OAuthCallbackInput represents the input for OAuth callback handling
type OAuthCallbackInput struct {
	Provider domain.AuthProvider
	Code     string
	State    string
}

// VerifyTokenInput represents the input for OAuth token verification (frontend SDK)
type VerifyTokenInput struct {
	Provider domain.AuthProvider
	Token    string // ID token for Google, access token for Facebook
}

// ToUserDTO converts a domain User to UserDTO
func ToUserDTO(user *domain.User) *UserDTO {
	if user == nil {
		return nil
	}

	return &UserDTO{
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

// NewAuthResponse creates an AuthResponse from domain user and tokens
func NewAuthResponse(user *domain.User, accessToken, refreshToken string, expiresAt int64) *AuthResponse {
	return &AuthResponse{
		User:         ToUserDTO(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}
}
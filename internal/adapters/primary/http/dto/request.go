package dto

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

// GoogleTokenRequest represents the Google ID token verification request
type GoogleTokenRequest struct {
	IDToken string `json:"id_token" binding:"required"`
}

// FacebookTokenRequest represents the Facebook access token verification request
type FacebookTokenRequest struct {
	AccessToken string `json:"access_token" binding:"required"`
}

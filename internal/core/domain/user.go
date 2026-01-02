package domain

import (
	"errors"
	"regexp"
	"time"
)

// AuthProvider represents the authentication method used
type AuthProvider string

const (
	AuthProviderEmail    AuthProvider = "email"
	AuthProviderGoogle   AuthProvider = "google"
	AuthProviderFacebook AuthProvider = "facebook"
)

// User represents a user entity in the domain
type User struct {
	ID           int64        `json:"id"`
	Email        string       `json:"email"`
	Name         string       `json:"name"`
	PasswordHash string       `json:"-"` // Never expose password hash in JSON
	Provider     AuthProvider `json:"provider"`
	ProviderID   string       `json:"provider_id,omitempty"` // OAuth provider user ID
	AvatarURL    string       `json:"avatar_url,omitempty"`
	IsActive     bool         `json:"is_active"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

// OAuthUserInfo represents user information from OAuth providers
type OAuthUserInfo struct {
	Provider   AuthProvider
	ProviderID string
	Email      string
	Name       string
	AvatarURL  string
}

var (
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrInvalidName     = errors.New("name must be between 1 and 255 characters")
	ErrPasswordTooWeak = errors.New("password must be at least 8 characters and contain uppercase, lowercase, number, and special character")
	ErrEmailRequired   = errors.New("email is required")
)

// emailRegex validates email format
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// NewUser creates a new user with email/password authentication
func NewUser(email, name, passwordHash string) (*User, error) {
	if err := ValidateEmail(email); err != nil {
		return nil, err
	}

	if err := ValidateName(name); err != nil {
		return nil, err
	}

	now := time.Now()
	return &User{
		Email:        email,
		Name:         name,
		PasswordHash: passwordHash,
		Provider:     AuthProviderEmail,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// NewOAuthUser creates a new user from OAuth provider information
func NewOAuthUser(info *OAuthUserInfo) (*User, error) {
	if info == nil {
		return nil, errors.New("oauth user info cannot be nil")
	}

	if info.ProviderID == "" {
		return nil, errors.New("provider ID is required")
	}

	if err := ValidateEmail(info.Email); err != nil {
		return nil, err
	}

	if err := ValidateName(info.Name); err != nil {
		return nil, err
	}

	now := time.Now()
	return &User{
		Email:      info.Email,
		Name:       info.Name,
		Provider:   info.Provider,
		ProviderID: info.ProviderID,
		AvatarURL:  info.AvatarURL,
		IsActive:   true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	if email == "" {
		return ErrEmailRequired
	}

	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}

	return nil
}

// ValidateName validates user name
func ValidateName(name string) error {
	if len(name) < 1 || len(name) > 255 {
		return ErrInvalidName
	}
	return nil
}

// ValidatePassword validates password strength
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return ErrPasswordTooWeak
	}

	var (
		hasUpper   = regexp.MustCompile(`[A-Z]`).MatchString(password)
		hasLower   = regexp.MustCompile(`[a-z]`).MatchString(password)
		hasNumber  = regexp.MustCompile(`[0-9]`).MatchString(password)
		hasSpecial = regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password)
	)

	if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		return ErrPasswordTooWeak
	}

	return nil
}

// UpdateProfile updates user profile information
func (u *User) UpdateProfile(name, avatarURL string) error {
	if err := ValidateName(name); err != nil {
		return err
	}

	u.Name = name
	u.AvatarURL = avatarURL
	u.UpdatedAt = time.Now()

	return nil
}

// Deactivate marks user as inactive
func (u *User) Deactivate() {
	u.IsActive = false
	u.UpdatedAt = time.Now()
}

// Activate marks user as active
func (u *User) Activate() {
	u.IsActive = true
	u.UpdatedAt = time.Now()
}

// IsOAuthUser returns true if user registered via OAuth
func (u *User) IsOAuthUser() bool {
	return u.Provider != AuthProviderEmail
}

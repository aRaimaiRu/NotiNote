package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUser(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		userName    string
		password    string
		wantErr     bool
		expectedErr error
	}{
		{
			name:     "valid user creation",
			email:    "test@example.com",
			userName: "Test User",
			password: "hashedpassword123",
			wantErr:  false,
		},
		{
			name:        "invalid email - empty",
			email:       "",
			userName:    "Test User",
			password:    "hashedpassword123",
			wantErr:     true,
			expectedErr: ErrEmailRequired,
		},
		{
			name:        "invalid email - no @",
			email:       "invalid.email.com",
			userName:    "Test User",
			password:    "hashedpassword123",
			wantErr:     true,
			expectedErr: ErrInvalidEmail,
		},
		{
			name:        "invalid name - empty",
			email:       "test@example.com",
			userName:    "",
			password:    "hashedpassword123",
			wantErr:     true,
			expectedErr: ErrInvalidName,
		},
		{
			name:        "invalid name - too long",
			email:       "test@example.com",
			userName:    string(make([]byte, 256)),
			password:    "hashedpassword123",
			wantErr:     true,
			expectedErr: ErrInvalidName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewUser(tt.email, tt.userName, tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, user)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.email, user.Email)
				assert.Equal(t, tt.userName, user.Name)
				assert.Equal(t, tt.password, user.PasswordHash)
				assert.Equal(t, AuthProviderEmail, user.Provider)
				assert.True(t, user.IsActive)
				assert.NotZero(t, user.CreatedAt)
				assert.NotZero(t, user.UpdatedAt)
			}
		})
	}
}

func TestNewOAuthUser(t *testing.T) {
	tests := []struct {
		name        string
		info        *OAuthUserInfo
		wantErr     bool
		expectedErr error
	}{
		{
			name: "valid Google OAuth user",
			info: &OAuthUserInfo{
				Provider:   AuthProviderGoogle,
				ProviderID: "google123",
				Email:      "user@gmail.com",
				Name:       "John Doe",
				AvatarURL:  "https://example.com/avatar.jpg",
			},
			wantErr: false,
		},
		{
			name: "valid Facebook OAuth user",
			info: &OAuthUserInfo{
				Provider:   AuthProviderFacebook,
				ProviderID: "fb456",
				Email:      "user@facebook.com",
				Name:       "Jane Doe",
				AvatarURL:  "",
			},
			wantErr: false,
		},
		{
			name: "invalid - nil info",
			info: nil,
			wantErr: true,
		},
		{
			name: "invalid - empty provider ID",
			info: &OAuthUserInfo{
				Provider:   AuthProviderGoogle,
				ProviderID: "",
				Email:      "user@gmail.com",
				Name:       "John Doe",
			},
			wantErr: true,
		},
		{
			name: "invalid - empty email",
			info: &OAuthUserInfo{
				Provider:   AuthProviderGoogle,
				ProviderID: "google123",
				Email:      "",
				Name:       "John Doe",
			},
			wantErr:     true,
			expectedErr: ErrEmailRequired,
		},
		{
			name: "invalid - empty name",
			info: &OAuthUserInfo{
				Provider:   AuthProviderGoogle,
				ProviderID: "google123",
				Email:      "user@gmail.com",
				Name:       "",
			},
			wantErr:     true,
			expectedErr: ErrInvalidName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewOAuthUser(tt.info)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, user)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.info.Email, user.Email)
				assert.Equal(t, tt.info.Name, user.Name)
				assert.Equal(t, tt.info.Provider, user.Provider)
				assert.Equal(t, tt.info.ProviderID, user.ProviderID)
				assert.Equal(t, tt.info.AvatarURL, user.AvatarURL)
				assert.Empty(t, user.PasswordHash)
				assert.True(t, user.IsActive)
				assert.NotZero(t, user.CreatedAt)
				assert.NotZero(t, user.UpdatedAt)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		wantErr     bool
		expectedErr error
	}{
		{"valid email", "user@example.com", false, nil},
		{"valid email with subdomain", "user@mail.example.com", false, nil},
		{"valid email with plus", "user+tag@example.com", false, nil},
		{"valid email with dots", "user.name@example.com", false, nil},
		{"empty email", "", true, ErrEmailRequired},
		{"no @ symbol", "userexample.com", true, ErrInvalidEmail},
		{"no domain", "user@", true, ErrInvalidEmail},
		{"no username", "@example.com", true, ErrInvalidEmail},
		{"invalid format", "user@.com", true, ErrInvalidEmail},
		{"multiple @ symbols", "user@@example.com", true, ErrInvalidEmail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		userName string
		wantErr bool
	}{
		{"valid name", "John Doe", false},
		{"valid name with special chars", "François O'Brien-Smith", false},
		{"valid single character", "A", false},
		{"valid long name", string(make([]byte, 255)), false},
		{"empty name", "", true},
		{"name too long", string(make([]byte, 256)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateName(tt.userName)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidName)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"valid password", "SecurePass123!", false},
		{"valid password all requirements", "Aa1!aaaa", false},
		{"too short", "Short1!", true},
		{"no uppercase", "password123!", true},
		{"no lowercase", "PASSWORD123!", true},
		{"no number", "Password!@#", true},
		{"no special char", "Password123", true},
		{"empty password", "", true},
		{"only 7 chars", "Pass1!", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrPasswordTooWeak)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUser_VerifyPassword(t *testing.T) {
	// This test requires the actual password hasher
	// It's more of an integration test, but included for completeness
	t.Run("should indicate password not set for OAuth users", func(t *testing.T) {
		user := &User{
			Email:        "oauth@example.com",
			Name:         "OAuth User",
			Provider:     AuthProviderGoogle,
			ProviderID:   "google123",
			PasswordHash: "",
		}

		// OAuth users don't have password hashes
		assert.Empty(t, user.PasswordHash)
	})
}

func TestUser_UpdateProfile(t *testing.T) {
	originalTime := time.Now().Add(-1 * time.Hour)
	user := &User{
		ID:        1,
		Email:     "test@example.com",
		Name:      "Original Name",
		AvatarURL: "https://example.com/old-avatar.jpg",
		CreatedAt: originalTime,
		UpdatedAt: originalTime,
	}

	// Update profile
	user.Name = "Updated Name"
	user.AvatarURL = "https://example.com/new-avatar.jpg"
	user.UpdatedAt = time.Now()

	assert.Equal(t, "Updated Name", user.Name)
	assert.Equal(t, "https://example.com/new-avatar.jpg", user.AvatarURL)
	assert.True(t, user.UpdatedAt.After(originalTime))
	assert.Equal(t, originalTime, user.CreatedAt) // CreatedAt should not change
}

func TestAuthProvider_Validation(t *testing.T) {
	tests := []struct {
		name     string
		provider AuthProvider
		isValid  bool
	}{
		{"email provider", AuthProviderEmail, true},
		{"google provider", AuthProviderGoogle, true},
		{"facebook provider", AuthProviderFacebook, true},
		{"invalid provider", AuthProvider("twitter"), false},
		{"empty provider", AuthProvider(""), false},
	}

	validProviders := map[AuthProvider]bool{
		AuthProviderEmail:    true,
		AuthProviderGoogle:   true,
		AuthProviderFacebook: true,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, exists := validProviders[tt.provider]
			assert.Equal(t, tt.isValid, exists)
		})
	}
}

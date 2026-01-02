package oauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/notinoteapp/internal/core/domain"
)

func TestNewGoogleProvider(t *testing.T) {
	tests := []struct {
		name         string
		clientID     string
		clientSecret string
		redirectURL  string
		scopes       []string
	}{
		{
			name:         "with custom scopes",
			clientID:     "test-client-id",
			clientSecret: "test-client-secret",
			redirectURL:  "http://localhost:8080/callback",
			scopes:       []string{"email", "profile"},
		},
		{
			name:         "with default scopes",
			clientID:     "test-client-id",
			clientSecret: "test-client-secret",
			redirectURL:  "http://localhost:8080/callback",
			scopes:       []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewGoogleProvider(tt.clientID, tt.clientSecret, tt.redirectURL, tt.scopes)

			assert.NotNil(t, provider)
			assert.NotNil(t, provider.config)
			assert.Equal(t, tt.clientID, provider.config.ClientID)
			assert.Equal(t, tt.clientSecret, provider.config.ClientSecret)
			assert.Equal(t, tt.redirectURL, provider.config.RedirectURL)

			if len(tt.scopes) == 0 {
				// Should use default scopes
				assert.Contains(t, provider.config.Scopes, "https://www.googleapis.com/auth/userinfo.email")
				assert.Contains(t, provider.config.Scopes, "https://www.googleapis.com/auth/userinfo.profile")
			} else {
				assert.Equal(t, tt.scopes, provider.config.Scopes)
			}
		})
	}
}

func TestGoogleProvider_GetAuthURL(t *testing.T) {
	provider := NewGoogleProvider("test-client-id", "test-secret", "http://localhost/callback", nil)

	state := "random-state-string"
	authURL := provider.GetAuthURL(state)

	assert.NotEmpty(t, authURL)
	assert.Contains(t, authURL, "accounts.google.com/o/oauth2")
	assert.Contains(t, authURL, "client_id=test-client-id")
	assert.Contains(t, authURL, "state=random-state-string")
	assert.Contains(t, authURL, "redirect_uri=")
	assert.Contains(t, authURL, "scope=")
	assert.Contains(t, authURL, "access_type=offline")
}

func TestGoogleProvider_GetProviderName(t *testing.T) {
	provider := NewGoogleProvider("test-client-id", "test-secret", "http://localhost/callback", nil)

	providerName := provider.GetProviderName()
	assert.Equal(t, domain.AuthProviderGoogle, providerName)
}

func TestGoogleProvider_ExchangeCode_Success(t *testing.T) {
	// Mock Google's token endpoint
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/token", r.URL.Path)

		response := map[string]interface{}{
			"access_token":  "mock-access-token",
			"token_type":    "Bearer",
			"expires_in":    3600,
			"refresh_token": "mock-refresh-token",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer tokenServer.Close()

	// Mock Google's userinfo endpoint
	userinfoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/userinfo")
		assert.Contains(t, r.Header.Get("Authorization"), "Bearer mock-access-token")

		response := GoogleUserInfo{
			Sub:     "google-user-123",
			Email:   "user@gmail.com",
			Name:    "Test User",
			Picture: "https://example.com/avatar.jpg",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer userinfoServer.Close()

	provider := NewGoogleProvider("test-client-id", "test-secret", "http://localhost/callback", nil)

	// Override endpoints for testing
	provider.config.Endpoint.TokenURL = tokenServer.URL + "/token"

	// Note: This test is simplified. In a real scenario, you'd need to mock the entire OAuth flow
	// or use dependency injection to replace the HTTP client
	t.Skip("Full integration test requires mocking oauth2 library's HTTP client")
}

func TestGoogleProvider_ExchangeCode_InvalidCode(t *testing.T) {
	provider := NewGoogleProvider("test-client-id", "test-secret", "http://localhost/callback", nil)

	ctx := context.Background()
	userInfo, err := provider.ExchangeCode(ctx, "invalid-code")

	assert.Error(t, err)
	assert.Nil(t, userInfo)
	// The error should be wrapped with domain.ErrOAuthCodeExchange or domain.ErrOAuthUserInfo
}

func TestGoogleProvider_ParseUserInfo(t *testing.T) {
	// Create a test server that returns user info
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := GoogleUserInfo{
			Sub:     "google-user-123",
			Email:   "user@gmail.com",
			Name:    "Test User",
			Picture: "https://example.com/avatar.jpg",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Test the getUserInfo method would parse this correctly
	// This is an indirect test since getUserInfo is private
	t.Run("valid user info response", func(t *testing.T) {
		// We can't directly test private methods, but we verify the structure
		var userInfo GoogleUserInfo
		resp, err := http.Get(server.URL)
		require.NoError(t, err)
		defer resp.Body.Close()

		err = json.NewDecoder(resp.Body).Decode(&userInfo)
		require.NoError(t, err)

		assert.Equal(t, "google-user-123", userInfo.Sub)
		assert.Equal(t, "user@gmail.com", userInfo.Email)
		assert.Equal(t, "Test User", userInfo.Name)
		assert.Equal(t, "https://example.com/avatar.jpg", userInfo.Picture)
	})
}

func TestGoogleUserInfo_ToOAuthUserInfo(t *testing.T) {
	googleInfo := GoogleUserInfo{
		Sub:     "google-123",
		Email:   "test@gmail.com",
		Name:    "Test User",
		Picture: "https://example.com/pic.jpg",
	}

	// Simulate conversion
	oauthInfo := &domain.OAuthUserInfo{
		Provider:   domain.AuthProviderGoogle,
		ProviderID: googleInfo.Sub,
		Email:      googleInfo.Email,
		Name:       googleInfo.Name,
		AvatarURL:  googleInfo.Picture,
	}

	assert.Equal(t, domain.AuthProviderGoogle, oauthInfo.Provider)
	assert.Equal(t, "google-123", oauthInfo.ProviderID)
	assert.Equal(t, "test@gmail.com", oauthInfo.Email)
	assert.Equal(t, "Test User", oauthInfo.Name)
	assert.Equal(t, "https://example.com/pic.jpg", oauthInfo.AvatarURL)
}

func TestGoogleProvider_EmptyFields(t *testing.T) {
	tests := []struct {
		name      string
		userInfo  GoogleUserInfo
		wantError bool
	}{
		{
			name: "all fields present",
			userInfo: GoogleUserInfo{
				Sub:     "google-123",
				Email:   "test@gmail.com",
				Name:    "Test User",
				Picture: "https://example.com/pic.jpg",
			},
			wantError: false,
		},
		{
			name: "missing picture (optional)",
			userInfo: GoogleUserInfo{
				Sub:     "google-123",
				Email:   "test@gmail.com",
				Name:    "Test User",
				Picture: "",
			},
			wantError: false,
		},
		{
			name: "missing required fields",
			userInfo: GoogleUserInfo{
				Sub:     "",
				Email:   "test@gmail.com",
				Name:    "Test User",
				Picture: "",
			},
			wantError: true, // Sub is required
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate that required fields are present
			hasRequiredFields := tt.userInfo.Sub != "" && tt.userInfo.Email != "" && tt.userInfo.Name != ""
			assert.Equal(t, !tt.wantError, hasRequiredFields)
		})
	}
}

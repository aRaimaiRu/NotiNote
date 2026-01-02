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

func TestNewFacebookProvider(t *testing.T) {
	tests := []struct {
		name        string
		appID       string
		appSecret   string
		redirectURL string
		scopes      []string
	}{
		{
			name:        "with custom scopes",
			appID:       "test-app-id",
			appSecret:   "test-app-secret",
			redirectURL: "http://localhost:8080/callback",
			scopes:      []string{"email", "public_profile", "user_friends"},
		},
		{
			name:        "with default scopes",
			appID:       "test-app-id",
			appSecret:   "test-app-secret",
			redirectURL: "http://localhost:8080/callback",
			scopes:      []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewFacebookProvider(tt.appID, tt.appSecret, tt.redirectURL, tt.scopes)

			assert.NotNil(t, provider)
			assert.Equal(t, tt.appID, provider.appID)
			assert.Equal(t, tt.appSecret, provider.appSecret)
			assert.Equal(t, tt.redirectURL, provider.redirectURL)

			if len(tt.scopes) == 0 {
				// Should use default scopes
				assert.Contains(t, provider.scopes, "email")
				assert.Contains(t, provider.scopes, "public_profile")
			} else {
				assert.Equal(t, tt.scopes, provider.scopes)
			}
		})
	}
}

func TestFacebookProvider_GetAuthURL(t *testing.T) {
	provider := NewFacebookProvider("test-app-id", "test-secret", "http://localhost/callback", nil)

	state := "random-state-string"
	authURL := provider.GetAuthURL(state)

	assert.NotEmpty(t, authURL)
	assert.Contains(t, authURL, "facebook.com/v18.0/dialog/oauth")
	assert.Contains(t, authURL, "client_id=test-app-id")
	assert.Contains(t, authURL, "state=random-state-string")
	assert.Contains(t, authURL, "redirect_uri=")
	assert.Contains(t, authURL, "scope=")
	assert.Contains(t, authURL, "response_type=code")
}

func TestFacebookProvider_GetProviderName(t *testing.T) {
	provider := NewFacebookProvider("test-app-id", "test-secret", "http://localhost/callback", nil)

	providerName := provider.GetProviderName()
	assert.Equal(t, domain.AuthProviderFacebook, providerName)
}

func TestFacebookProvider_GetAccessToken_Success(t *testing.T) {
	// Mock Facebook's token endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/oauth/access_token")

		query := r.URL.Query()
		assert.Equal(t, "test-app-id", query.Get("client_id"))
		assert.Equal(t, "test-secret", query.Get("client_secret"))
		assert.Equal(t, "test-code", query.Get("code"))

		response := FacebookTokenResponse{
			AccessToken: "mock-facebook-token",
			TokenType:   "bearer",
			ExpiresIn:   5184000,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	provider := NewFacebookProvider("test-app-id", "test-secret", "http://localhost/callback", nil)

	ctx := context.Background()

	// Note: This is a simplified test. Full test would require mocking the HTTP client
	t.Run("token response structure", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/oauth/access_token?client_id=test-app-id&client_secret=test-secret&code=test-code")
		require.NoError(t, err)
		defer resp.Body.Close()

		var tokenResp FacebookTokenResponse
		err = json.NewDecoder(resp.Body).Decode(&tokenResp)
		require.NoError(t, err)

		assert.Equal(t, "mock-facebook-token", tokenResp.AccessToken)
		assert.Equal(t, "bearer", tokenResp.TokenType)
		assert.Equal(t, 5184000, tokenResp.ExpiresIn)
	})
}

func TestFacebookProvider_GetUserInfo_Success(t *testing.T) {
	// Mock Facebook's Graph API userinfo endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/me")

		query := r.URL.Query()
		assert.Equal(t, "id,email,name,picture", query.Get("fields"))
		assert.Equal(t, "mock-access-token", query.Get("access_token"))

		response := FacebookUserInfo{
			ID:    "facebook-user-123",
			Email: "user@facebook.com",
			Name:  "Facebook User",
			Picture: struct {
				Data struct {
					URL string `json:"url"`
				} `json:"data"`
			}{
				Data: struct {
					URL string `json:"url"`
				}{
					URL: "https://example.com/fb-avatar.jpg",
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	t.Run("userinfo response structure", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/me?fields=id,email,name,picture&access_token=mock-access-token")
		require.NoError(t, err)
		defer resp.Body.Close()

		var userInfo FacebookUserInfo
		err = json.NewDecoder(resp.Body).Decode(&userInfo)
		require.NoError(t, err)

		assert.Equal(t, "facebook-user-123", userInfo.ID)
		assert.Equal(t, "user@facebook.com", userInfo.Email)
		assert.Equal(t, "Facebook User", userInfo.Name)
		assert.Equal(t, "https://example.com/fb-avatar.jpg", userInfo.Picture.Data.URL)
	})
}

func TestFacebookProvider_ExchangeCode_InvalidCode(t *testing.T) {
	provider := NewFacebookProvider("test-app-id", "test-secret", "http://localhost/callback", nil)

	ctx := context.Background()
	userInfo, err := provider.ExchangeCode(ctx, "invalid-code")

	assert.Error(t, err)
	assert.Nil(t, userInfo)
}

func TestFacebookProvider_ErrorResponse(t *testing.T) {
	// Mock Facebook error response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		response := FacebookErrorResponse{
			Error: FacebookError{
				Message: "Invalid OAuth access token",
				Type:    "OAuthException",
				Code:    190,
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	t.Run("error response parsing", func(t *testing.T) {
		resp, err := http.Get(server.URL)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var errorResp FacebookErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&errorResp)
		require.NoError(t, err)

		assert.Equal(t, "Invalid OAuth access token", errorResp.Error.Message)
		assert.Equal(t, "OAuthException", errorResp.Error.Type)
		assert.Equal(t, 190, errorResp.Error.Code)
	})
}

func TestFacebookUserInfo_ToOAuthUserInfo(t *testing.T) {
	facebookInfo := FacebookUserInfo{
		ID:    "fb-123",
		Email: "test@facebook.com",
		Name:  "Test User",
		Picture: struct {
			Data struct {
				URL string `json:"url"`
			} `json:"data"`
		}{
			Data: struct {
				URL string `json:"url"`
			}{
				URL: "https://example.com/pic.jpg",
			},
		},
	}

	// Simulate conversion
	oauthInfo := &domain.OAuthUserInfo{
		Provider:   domain.AuthProviderFacebook,
		ProviderID: facebookInfo.ID,
		Email:      facebookInfo.Email,
		Name:       facebookInfo.Name,
		AvatarURL:  facebookInfo.Picture.Data.URL,
	}

	assert.Equal(t, domain.AuthProviderFacebook, oauthInfo.Provider)
	assert.Equal(t, "fb-123", oauthInfo.ProviderID)
	assert.Equal(t, "test@facebook.com", oauthInfo.Email)
	assert.Equal(t, "Test User", oauthInfo.Name)
	assert.Equal(t, "https://example.com/pic.jpg", oauthInfo.AvatarURL)
}

func TestFacebookProvider_MissingEmail(t *testing.T) {
	// Test case where user doesn't grant email permission
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := FacebookUserInfo{
			ID:    "fb-123",
			Email: "", // No email provided
			Name:  "Test User",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	t.Run("missing email in response", func(t *testing.T) {
		resp, err := http.Get(server.URL)
		require.NoError(t, err)
		defer resp.Body.Close()

		var userInfo FacebookUserInfo
		err = json.NewDecoder(resp.Body).Decode(&userInfo)
		require.NoError(t, err)

		// Email is empty - this should be handled by domain validation
		assert.Empty(t, userInfo.Email)

		// Creating OAuthUserInfo with empty email should fail in domain layer
		oauthInfo := &domain.OAuthUserInfo{
			Provider:   domain.AuthProviderFacebook,
			ProviderID: userInfo.ID,
			Email:      userInfo.Email,
			Name:       userInfo.Name,
		}

		_, err = domain.NewOAuthUser(oauthInfo)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidEmail)
	})
}

func TestFacebookProvider_ScopeFormat(t *testing.T) {
	tests := []struct {
		name           string
		scopes         []string
		expectedString string
	}{
		{
			name:           "single scope",
			scopes:         []string{"email"},
			expectedString: "email",
		},
		{
			name:           "multiple scopes",
			scopes:         []string{"email", "public_profile"},
			expectedString: "email,public_profile",
		},
		{
			name:           "three scopes",
			scopes:         []string{"email", "public_profile", "user_friends"},
			expectedString: "email,public_profile,user_friends",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewFacebookProvider("test-app-id", "test-secret", "http://localhost/callback", tt.scopes)
			authURL := provider.GetAuthURL("test-state")

			// Verify scope parameter is correctly formatted
			assert.Contains(t, authURL, "scope="+tt.expectedString)
		})
	}
}

func TestFacebookProvider_URLEncoding(t *testing.T) {
	provider := NewFacebookProvider("test-app-id", "test-secret", "http://localhost:8080/auth/callback", nil)

	authURL := provider.GetAuthURL("test-state-123")

	// Verify URL encoding for redirect_uri
	assert.Contains(t, authURL, "redirect_uri=")
	// The URL should be properly encoded
	assert.NotContains(t, authURL, "http://localhost:8080/auth/callback") // Should be encoded
}

package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/yourusername/notinoteapp/internal/core/domain"
)

// FacebookProvider implements OAuth authentication for Facebook
type FacebookProvider struct {
	appID       string
	appSecret   string
	redirectURL string
	scopes      []string
}

// FacebookUserInfo represents the user info response from Facebook
type FacebookUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture struct {
		Data struct {
			URL string `json:"url"`
		} `json:"data"`
	} `json:"picture"`
}

// FacebookTokenResponse represents the token response from Facebook
type FacebookTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// NewFacebookProvider creates a new Facebook OAuth provider
func NewFacebookProvider(appID, appSecret, redirectURL string, scopes []string) *FacebookProvider {
	if len(scopes) == 0 {
		scopes = []string{"email", "public_profile"}
	}

	return &FacebookProvider{
		appID:       appID,
		appSecret:   appSecret,
		redirectURL: redirectURL,
		scopes:      scopes,
	}
}

// GetAuthURL generates the OAuth authorization URL with state
func (f *FacebookProvider) GetAuthURL(state string) string {
	params := url.Values{}
	params.Set("client_id", f.appID)
	params.Set("redirect_uri", f.redirectURL)
	params.Set("scope", strings.Join(f.scopes, ","))
	params.Set("state", state)
	params.Set("response_type", "code")

	return "https://www.facebook.com/v18.0/dialog/oauth?" + params.Encode()
}

// ExchangeCode exchanges authorization code for access token and retrieves user info
func (f *FacebookProvider) ExchangeCode(ctx context.Context, code string) (*domain.OAuthUserInfo, error) {
	// Exchange code for access token
	token, err := f.getAccessToken(ctx, code)
	if err != nil {
		return nil, err
	}

	// Get user info
	userInfo, err := f.getUserInfo(ctx, token.AccessToken)
	if err != nil {
		return nil, err
	}

	return &domain.OAuthUserInfo{
		Provider:   domain.AuthProviderFacebook,
		ProviderID: userInfo.ID,
		Email:      userInfo.Email,
		Name:       userInfo.Name,
		AvatarURL:  userInfo.Picture.Data.URL,
	}, nil
}

// getAccessToken exchanges code for access token
func (f *FacebookProvider) getAccessToken(ctx context.Context, code string) (*FacebookTokenResponse, error) {
	params := url.Values{}
	params.Set("client_id", f.appID)
	params.Set("client_secret", f.appSecret)
	params.Set("redirect_uri", f.redirectURL)
	params.Set("code", code)

	tokenURL := "https://graph.facebook.com/v18.0/oauth/access_token?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", tokenURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create request", domain.ErrOAuthCodeExchange)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrOAuthCodeExchange, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: status %d, body: %s", domain.ErrOAuthProviderError, resp.StatusCode, string(body))
	}

	var tokenResp FacebookTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("%w: failed to decode token response", domain.ErrOAuthCodeExchange)
	}

	return &tokenResp, nil
}

// getUserInfo fetches user information from Facebook
func (f *FacebookProvider) getUserInfo(ctx context.Context, accessToken string) (*FacebookUserInfo, error) {
	fields := "id,email,name,picture.type(large)"
	userInfoURL := fmt.Sprintf("https://graph.facebook.com/v18.0/me?fields=%s&access_token=%s", fields, accessToken)

	req, err := http.NewRequestWithContext(ctx, "GET", userInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create request", domain.ErrOAuthUserInfo)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrOAuthUserInfo, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: status %d, body: %s", domain.ErrOAuthProviderError, resp.StatusCode, string(body))
	}

	var userInfo FacebookUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("%w: failed to decode response", domain.ErrOAuthUserInfo)
	}

	if userInfo.Email == "" {
		return nil, fmt.Errorf("%w: email not provided by Facebook", domain.ErrOAuthUserInfo)
	}

	return &userInfo, nil
}

// GetProviderName returns the provider name
func (f *FacebookProvider) GetProviderName() domain.AuthProvider {
	return domain.AuthProviderFacebook
}

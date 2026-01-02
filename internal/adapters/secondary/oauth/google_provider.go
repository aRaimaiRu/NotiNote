package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/yourusername/notinoteapp/internal/core/domain"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GoogleProvider implements OAuth authentication for Google
type GoogleProvider struct {
	config *oauth2.Config
}

// GoogleUserInfo represents the user info response from Google
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

// NewGoogleProvider creates a new Google OAuth provider
func NewGoogleProvider(clientID, clientSecret, redirectURL string, scopes []string) *GoogleProvider {
	if len(scopes) == 0 {
		scopes = []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		}
	}

	return &GoogleProvider{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       scopes,
			Endpoint:     google.Endpoint,
		},
	}
}

// GetAuthURL generates the OAuth authorization URL with state
func (g *GoogleProvider) GetAuthURL(state string) string {
	return g.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// ExchangeCode exchanges authorization code for access token and retrieves user info
func (g *GoogleProvider) ExchangeCode(ctx context.Context, code string) (*domain.OAuthUserInfo, error) {
	// Exchange code for token
	token, err := g.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrOAuthCodeExchange, err)
	}

	// Get user info
	userInfo, err := g.getUserInfo(ctx, token.AccessToken)
	if err != nil {
		return nil, err
	}

	return &domain.OAuthUserInfo{
		Provider:   domain.AuthProviderGoogle,
		ProviderID: userInfo.ID,
		Email:      userInfo.Email,
		Name:       userInfo.Name,
		AvatarURL:  userInfo.Picture,
	}, nil
}

// getUserInfo fetches user information from Google
func (g *GoogleProvider) getUserInfo(ctx context.Context, accessToken string) (*GoogleUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create request", domain.ErrOAuthUserInfo)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

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

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("%w: failed to decode response", domain.ErrOAuthUserInfo)
	}

	if !userInfo.VerifiedEmail {
		return nil, fmt.Errorf("%w: email not verified", domain.ErrOAuthUserInfo)
	}

	return &userInfo, nil
}

// GetProviderName returns the provider name
func (g *GoogleProvider) GetProviderName() domain.AuthProvider {
	return domain.AuthProviderGoogle
}

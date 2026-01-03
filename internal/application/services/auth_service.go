package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/yourusername/notinoteapp/internal/application/dto"
	"github.com/yourusername/notinoteapp/internal/core/domain"
	"github.com/yourusername/notinoteapp/internal/core/ports"
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo       ports.UserRepository
	passwordHasher ports.PasswordHasher
	tokenService   ports.TokenService
	stateGenerator ports.StateGenerator
	oauthProviders map[domain.AuthProvider]ports.OAuthProvider
}

// NewAuthService creates a new authentication service
func NewAuthService(
	userRepo ports.UserRepository,
	passwordHasher ports.PasswordHasher,
	tokenService ports.TokenService,
	stateGenerator ports.StateGenerator,
) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		passwordHasher: passwordHasher,
		tokenService:   tokenService,
		stateGenerator: stateGenerator,
		oauthProviders: make(map[domain.AuthProvider]ports.OAuthProvider),
	}
}

// RegisterOAuthProvider registers an OAuth provider
func (s *AuthService) RegisterOAuthProvider(provider ports.OAuthProvider) {
	s.oauthProviders[provider.GetProviderName()] = provider
}

// Register registers a new user with email and password
func (s *AuthService) Register(ctx context.Context, email, password, name string) (*dto.AuthResponse, error) {
	// Validate email
	if err := domain.ValidateEmail(email); err != nil {
		return nil, err
	}

	// Validate password
	if err := domain.ValidatePassword(password); err != nil {
		return nil, err
	}

	// Validate name
	if err := domain.ValidateName(name); err != nil {
		return nil, err
	}

	// Check if user already exists
	existingUser, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	// Hash password
	passwordHash, err := s.passwordHasher.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user, err := domain.NewUser(email, name, passwordHash)
	if err != nil {
		return nil, err
	}

	// Save user to database
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate tokens
	return s.generateAuthResponse(user)
}

// Login authenticates a user with email and password
func (s *AuthService) Login(ctx context.Context, email, password string) (*dto.AuthResponse, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Check if user is OAuth user
	if user.IsOAuthUser() {
		return nil, fmt.Errorf("this account uses %s login. Please use %s to sign in", user.Provider, user.Provider)
	}

	// Check if user is active
	if !user.IsActive {
		return nil, domain.ErrUserInactive
	}

	// Verify password
	if !s.passwordHasher.CheckPassword(password, user.PasswordHash) {
		return nil, domain.ErrInvalidCredentials
	}

	// Generate tokens
	return s.generateAuthResponse(user)
}

// GetOAuthURL generates the OAuth authorization URL
func (s *AuthService) GetOAuthURL(ctx context.Context, provider domain.AuthProvider) (string, error) {
	oauthProvider, ok := s.oauthProviders[provider]
	if !ok {
		return "", fmt.Errorf("oauth provider %s not supported", provider)
	}

	// Generate state for CSRF protection
	state, err := s.stateGenerator.GenerateState()
	if err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}

	// Store state in Redis with 10 minute expiration
	if err := s.stateGenerator.StoreState(ctx, state, 600); err != nil {
		return "", fmt.Errorf("failed to store state: %w", err)
	}

	// Generate authorization URL
	authURL := oauthProvider.GetAuthURL(state)
	return authURL, nil
}

// HandleOAuthCallback handles the OAuth callback
func (s *AuthService) HandleOAuthCallback(ctx context.Context, provider domain.AuthProvider, code, state string) (*dto.AuthResponse, error) {
	// Validate state to prevent CSRF
	valid, err := s.stateGenerator.GetState(ctx, state)
	if err != nil {
		return nil, fmt.Errorf("failed to validate state: %w", err)
	}
	if !valid {
		return nil, domain.ErrOAuthStateMismatch
	}

	// Get OAuth provider
	oauthProvider, ok := s.oauthProviders[provider]
	if !ok {
		return nil, fmt.Errorf("oauth provider %s not supported", provider)
	}

	// Exchange code for user info
	userInfo, err := oauthProvider.ExchangeCode(ctx, code)
	if err != nil {
		return nil, err
	}

	// Check if user already exists with this provider
	user, err := s.userRepo.FindByProvider(ctx, userInfo.Provider, userInfo.ProviderID)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return nil, fmt.Errorf("failed to find user by provider: %w", err)
	}

	// If user exists, login
	if user != nil {
		// Check if user is active
		if !user.IsActive {
			return nil, domain.ErrUserInactive
		}

		// Update user info (name, avatar) if changed
		if user.Name != userInfo.Name || user.AvatarURL != userInfo.AvatarURL {
			user.Name = userInfo.Name
			user.AvatarURL = userInfo.AvatarURL
			if err := s.userRepo.Update(ctx, user); err != nil {
				// Log error but don't fail login
				fmt.Printf("failed to update user info: %v\n", err)
			}
		}

		return s.generateAuthResponse(user)
	}

	// Check if user exists with same email but different provider
	existingUser, err := s.userRepo.FindByEmail(ctx, userInfo.Email)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, fmt.Errorf("an account with this email already exists using %s. Please use %s to sign in", existingUser.Provider, existingUser.Provider)
	}

	// Create new user
	newUser, err := domain.NewOAuthUser(userInfo)
	if err != nil {
		return nil, err
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return s.generateAuthResponse(newUser)
}

// RefreshToken refreshes an access token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*dto.AuthResponse, error) {
	// Validate refresh token and get user info
	userID, email, err := s.tokenService.ValidateToken(refreshToken)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	// Get user from database
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidToken
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Verify email matches
	if user.Email != email {
		return nil, domain.ErrInvalidToken
	}

	// Check if user is active
	if !user.IsActive {
		return nil, domain.ErrUserInactive
	}

	// Generate new tokens
	return s.generateAuthResponse(user)
}

// GetUserByID retrieves a user by their ID
func (s *AuthService) GetUserByID(ctx context.Context, userID int64) (*domain.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// VerifyGoogleToken verifies a Google ID token from frontend SDK
func (s *AuthService) VerifyGoogleToken(ctx context.Context, idToken string) (*dto.AuthResponse, error) {
	// Get Google provider
	googleProvider, ok := s.oauthProviders[domain.AuthProviderGoogle]
	if !ok {
		return nil, fmt.Errorf("google OAuth provider not registered")
	}

	// Type assert to access VerifyIDToken method
	type GoogleTokenVerifier interface {
		VerifyIDToken(ctx context.Context, idToken string) (*domain.OAuthUserInfo, error)
	}

	verifier, ok := googleProvider.(GoogleTokenVerifier)
	if !ok {
		return nil, fmt.Errorf("google provider does not support token verification")
	}

	// Verify token and get user info
	userInfo, err := verifier.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, err
	}

	// Process OAuth user info (create or update user)
	return s.processOAuthUser(ctx, userInfo)
}

// VerifyFacebookToken verifies a Facebook access token from frontend SDK
func (s *AuthService) VerifyFacebookToken(ctx context.Context, accessToken string) (*dto.AuthResponse, error) {
	// Get Facebook provider
	facebookProvider, ok := s.oauthProviders[domain.AuthProviderFacebook]
	if !ok {
		return nil, fmt.Errorf("facebook OAuth provider not registered")
	}

	// Type assert to access VerifyAccessToken method
	type FacebookTokenVerifier interface {
		VerifyAccessToken(ctx context.Context, accessToken string) (*domain.OAuthUserInfo, error)
	}

	verifier, ok := facebookProvider.(FacebookTokenVerifier)
	if !ok {
		return nil, fmt.Errorf("facebook provider does not support token verification")
	}

	// Verify token and get user info
	userInfo, err := verifier.VerifyAccessToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	// Process OAuth user info (create or update user)
	return s.processOAuthUser(ctx, userInfo)
}

// processOAuthUser handles creating or updating a user from OAuth info
func (s *AuthService) processOAuthUser(ctx context.Context, userInfo *domain.OAuthUserInfo) (*dto.AuthResponse, error) {
	// Check if user already exists with this provider
	user, err := s.userRepo.FindByProvider(ctx, userInfo.Provider, userInfo.ProviderID)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return nil, fmt.Errorf("failed to find user by provider: %w", err)
	}

	// If user exists, login
	if user != nil {
		// Check if user is active
		if !user.IsActive {
			return nil, domain.ErrUserInactive
		}

		// Update user info (name, avatar) if changed
		if user.Name != userInfo.Name || user.AvatarURL != userInfo.AvatarURL {
			user.Name = userInfo.Name
			user.AvatarURL = userInfo.AvatarURL
			if err := s.userRepo.Update(ctx, user); err != nil {
				// Log error but don't fail login
				fmt.Printf("failed to update user info: %v\n", err)
			}
		}

		return s.generateAuthResponse(user)
	}

	// Check if user exists with same email but different provider
	existingUser, err := s.userRepo.FindByEmail(ctx, userInfo.Email)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, fmt.Errorf("an account with this email already exists using %s. Please use %s to sign in", existingUser.Provider, existingUser.Provider)
	}

	// Create new user
	newUser, err := domain.NewOAuthUser(userInfo)
	if err != nil {
		return nil, err
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return s.generateAuthResponse(newUser)
}

// generateAuthResponse generates access and refresh tokens
func (s *AuthService) generateAuthResponse(user *domain.User) (*dto.AuthResponse, error) {
	accessToken, err := s.tokenService.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.tokenService.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// ExpiresAt will be set by handler based on JWT expiration
	return dto.NewAuthResponse(user, accessToken, refreshToken, 0), nil
}

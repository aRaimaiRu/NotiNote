package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/notinoteapp/internal/core/domain"
	"github.com/yourusername/notinoteapp/internal/core/ports"
)

// Mock implementations

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	// Simulate ID assignment
	user.ID = 1
	return nil
}

func (m *MockUserRepository) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) FindByProvider(ctx context.Context, provider domain.AuthProvider, providerID string) (*domain.User, error) {
	args := m.Called(ctx, provider, providerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]*domain.User, int64, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*domain.User), args.Get(1).(int64), args.Error(2)
}

type MockPasswordHasher struct {
	mock.Mock
}

func (m *MockPasswordHasher) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockPasswordHasher) CheckPassword(password, hash string) bool {
	args := m.Called(password, hash)
	return args.Bool(0)
}

type MockTokenService struct {
	mock.Mock
}

func (m *MockTokenService) GenerateToken(userID int64, email string) (string, error) {
	args := m.Called(userID, email)
	return args.String(0), args.Error(1)
}

func (m *MockTokenService) GenerateRefreshToken(userID int64, email string) (string, error) {
	args := m.Called(userID, email)
	return args.String(0), args.Error(1)
}

func (m *MockTokenService) ValidateToken(token string) (userID int64, email string, err error) {
	args := m.Called(token)
	return args.Get(0).(int64), args.String(1), args.Error(2)
}

func (m *MockTokenService) RefreshToken(refreshToken string) (string, error) {
	args := m.Called(refreshToken)
	return args.String(0), args.Error(1)
}

type MockStateGenerator struct {
	mock.Mock
}

func (m *MockStateGenerator) GenerateState() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockStateGenerator) ValidateState(state, expected string) bool {
	args := m.Called(state, expected)
	return args.Bool(0)
}

func (m *MockStateGenerator) StoreState(ctx context.Context, state string, ttl int) error {
	args := m.Called(ctx, state, ttl)
	return args.Error(0)
}

func (m *MockStateGenerator) GetState(ctx context.Context, state string) (bool, error) {
	args := m.Called(ctx, state)
	return args.Bool(0), args.Error(1)
}

type MockOAuthProvider struct {
	mock.Mock
}

func (m *MockOAuthProvider) GetAuthURL(state string) string {
	args := m.Called(state)
	return args.String(0)
}

func (m *MockOAuthProvider) ExchangeCode(ctx context.Context, code string) (*domain.OAuthUserInfo, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.OAuthUserInfo), args.Error(1)
}

func (m *MockOAuthProvider) GetProviderName() domain.AuthProvider {
	args := m.Called()
	return args.Get(0).(domain.AuthProvider)
}

// Tests

func TestAuthService_Register_Success(t *testing.T) {
	// Setup mocks
	userRepo := new(MockUserRepository)
	passwordHasher := new(MockPasswordHasher)
	tokenService := new(MockTokenService)

	userRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(nil, domain.ErrUserNotFound)
	passwordHasher.On("HashPassword", "Password123!").Return("hashed-password", nil)
	userRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)
	tokenService.On("GenerateToken", int64(1), "test@example.com").Return("access-token", nil)
	tokenService.On("GenerateRefreshToken", int64(1), "test@example.com").Return("refresh-token", nil)

	// Create service
	service := NewAuthService(userRepo, passwordHasher, tokenService, nil)

	// Execute
	ctx := context.Background()
	resp, err := service.Register(ctx, "test@example.com", "Password123!", "Test User")

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test@example.com", resp.User.Email)
	assert.Equal(t, "Test User", resp.User.Name)
	assert.Equal(t, "access-token", resp.AccessToken)
	assert.Equal(t, "refresh-token", resp.RefreshToken)

	userRepo.AssertExpectations(t)
	passwordHasher.AssertExpectations(t)
	tokenService.AssertExpectations(t)
}

func TestAuthService_Register_UserAlreadyExists(t *testing.T) {
	userRepo := new(MockUserRepository)
	passwordHasher := new(MockPasswordHasher)

	existingUser := &domain.User{
		ID:    1,
		Email: "test@example.com",
		Name:  "Existing User",
	}
	userRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(existingUser, nil)

	service := NewAuthService(userRepo, passwordHasher, nil, nil)

	ctx := context.Background()
	resp, err := service.Register(ctx, "test@example.com", "Password123!", "Test User")

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrUserAlreadyExists)
	assert.Nil(t, resp)

	userRepo.AssertExpectations(t)
}

func TestAuthService_Register_InvalidEmail(t *testing.T) {
	service := NewAuthService(nil, nil, nil, nil)

	ctx := context.Background()
	resp, err := service.Register(ctx, "invalid-email", "Password123!", "Test User")

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidEmail)
	assert.Nil(t, resp)
}

func TestAuthService_Register_WeakPassword(t *testing.T) {
	service := NewAuthService(nil, nil, nil, nil)

	ctx := context.Background()
	resp, err := service.Register(ctx, "test@example.com", "weak", "Test User")

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrPasswordTooWeak)
	assert.Nil(t, resp)
}

func TestAuthService_Login_Success(t *testing.T) {
	userRepo := new(MockUserRepository)
	passwordHasher := new(MockPasswordHasher)
	tokenService := new(MockTokenService)

	user := &domain.User{
		ID:           1,
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "hashed-password",
		Provider:     domain.AuthProviderEmail,
		IsActive:     true,
	}

	userRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(user, nil)
	passwordHasher.On("CheckPassword", "Password123!", "hashed-password").Return(true)
	tokenService.On("GenerateToken", int64(1), "test@example.com").Return("access-token", nil)
	tokenService.On("GenerateRefreshToken", int64(1), "test@example.com").Return("refresh-token", nil)

	service := NewAuthService(userRepo, passwordHasher, tokenService, nil)

	ctx := context.Background()
	resp, err := service.Login(ctx, "test@example.com", "Password123!")

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test@example.com", resp.User.Email)
	assert.Equal(t, "access-token", resp.AccessToken)
	assert.Equal(t, "refresh-token", resp.RefreshToken)

	userRepo.AssertExpectations(t)
	passwordHasher.AssertExpectations(t)
	tokenService.AssertExpectations(t)
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	userRepo := new(MockUserRepository)
	passwordHasher := new(MockPasswordHasher)

	user := &domain.User{
		ID:           1,
		Email:        "test@example.com",
		PasswordHash: "hashed-password",
		Provider:     domain.AuthProviderEmail,
		IsActive:     true,
	}

	userRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(user, nil)
	passwordHasher.On("CheckPassword", "WrongPassword!", "hashed-password").Return(false)

	service := NewAuthService(userRepo, passwordHasher, nil, nil)

	ctx := context.Background()
	resp, err := service.Login(ctx, "test@example.com", "WrongPassword!")

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
	assert.Nil(t, resp)

	userRepo.AssertExpectations(t)
	passwordHasher.AssertExpectations(t)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	userRepo := new(MockUserRepository)

	userRepo.On("FindByEmail", mock.Anything, "notfound@example.com").Return(nil, domain.ErrUserNotFound)

	service := NewAuthService(userRepo, nil, nil, nil)

	ctx := context.Background()
	resp, err := service.Login(ctx, "notfound@example.com", "Password123!")

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
	assert.Nil(t, resp)

	userRepo.AssertExpectations(t)
}

func TestAuthService_Login_InactiveUser(t *testing.T) {
	userRepo := new(MockUserRepository)

	user := &domain.User{
		ID:       1,
		Email:    "test@example.com",
		IsActive: false,
	}

	userRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(user, nil)

	service := NewAuthService(userRepo, nil, nil, nil)

	ctx := context.Background()
	resp, err := service.Login(ctx, "test@example.com", "Password123!")

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrUserInactive)
	assert.Nil(t, resp)

	userRepo.AssertExpectations(t)
}

func TestAuthService_Login_OAuthUserCannotLoginWithPassword(t *testing.T) {
	userRepo := new(MockUserRepository)

	user := &domain.User{
		ID:           1,
		Email:        "test@example.com",
		Provider:     domain.AuthProviderGoogle,
		ProviderID:   "google-123",
		PasswordHash: "",
		IsActive:     true,
	}

	userRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(user, nil)

	service := NewAuthService(userRepo, nil, nil, nil)

	ctx := context.Background()
	resp, err := service.Login(ctx, "test@example.com", "Password123!")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "registered with google")
	assert.Nil(t, resp)

	userRepo.AssertExpectations(t)
}

func TestAuthService_GetOAuthURL_Success(t *testing.T) {
	stateGen := new(MockStateGenerator)
	oauthProvider := new(MockOAuthProvider)

	stateGen.On("GenerateState").Return("random-state", nil)
	stateGen.On("StoreState", mock.Anything, "random-state", 600).Return(nil)
	oauthProvider.On("GetAuthURL", "random-state").Return("https://accounts.google.com/oauth?state=random-state", nil)

	oauthProviders := map[domain.AuthProvider]ports.OAuthProvider{
		domain.AuthProviderGoogle: oauthProvider,
	}

	service := NewAuthService(nil, nil, nil, stateGen, oauthProviders)

	ctx := context.Background()
	authURL, err := service.GetOAuthURL(ctx, domain.AuthProviderGoogle)

	require.NoError(t, err)
	assert.Contains(t, authURL, "https://accounts.google.com/oauth")
	assert.Contains(t, authURL, "state=random-state")

	stateGen.AssertExpectations(t)
	oauthProvider.AssertExpectations(t)
}

func TestAuthService_GetOAuthURL_UnsupportedProvider(t *testing.T) {
	service := NewAuthService(nil, nil, nil, nil, map[domain.AuthProvider]ports.OAuthProvider{})

	ctx := context.Background()
	authURL, err := service.GetOAuthURL(ctx, domain.AuthProviderGoogle)

	assert.Error(t, err)
	assert.Empty(t, authURL)
	assert.Contains(t, err.Error(), "not supported")
}

func TestAuthService_HandleOAuthCallback_NewUser(t *testing.T) {
	userRepo := new(MockUserRepository)
	tokenService := new(MockTokenService)
	stateGen := new(MockStateGenerator)
	oauthProvider := new(MockOAuthProvider)

	oauthUserInfo := &domain.OAuthUserInfo{
		Provider:   domain.AuthProviderGoogle,
		ProviderID: "google-123",
		Email:      "newuser@gmail.com",
		Name:       "New User",
		AvatarURL:  "https://example.com/avatar.jpg",
	}

	stateGen.On("GetState", mock.Anything, "valid-state").Return(true, nil)
	oauthProvider.On("ExchangeCode", mock.Anything, "auth-code").Return(oauthUserInfo, nil)
	userRepo.On("FindByProvider", mock.Anything, domain.AuthProviderGoogle, "google-123").Return(nil, domain.ErrUserNotFound)
	userRepo.On("FindByEmail", mock.Anything, "newuser@gmail.com").Return(nil, domain.ErrUserNotFound)
	userRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)
	tokenService.On("GenerateToken", int64(1), "newuser@gmail.com").Return("access-token", nil)
	tokenService.On("GenerateRefreshToken", int64(1), "newuser@gmail.com").Return("refresh-token", nil)

	oauthProviders := map[domain.AuthProvider]ports.OAuthProvider{
		domain.AuthProviderGoogle: oauthProvider,
	}

	service := NewAuthService(userRepo, nil, tokenService, stateGen, oauthProviders)

	ctx := context.Background()
	resp, err := service.HandleOAuthCallback(ctx, domain.AuthProviderGoogle, "auth-code", "valid-state")

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "newuser@gmail.com", resp.User.Email)
	assert.Equal(t, "access-token", resp.AccessToken)

	stateGen.AssertExpectations(t)
	oauthProvider.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	tokenService.AssertExpectations(t)
}

func TestAuthService_HandleOAuthCallback_ExistingUser(t *testing.T) {
	userRepo := new(MockUserRepository)
	tokenService := new(MockTokenService)
	stateGen := new(MockStateGenerator)
	oauthProvider := new(MockOAuthProvider)

	existingUser := &domain.User{
		ID:         1,
		Email:      "existing@gmail.com",
		Name:       "Existing User",
		Provider:   domain.AuthProviderGoogle,
		ProviderID: "google-123",
		IsActive:   true,
	}

	oauthUserInfo := &domain.OAuthUserInfo{
		Provider:   domain.AuthProviderGoogle,
		ProviderID: "google-123",
		Email:      "existing@gmail.com",
		Name:       "Updated Name",
		AvatarURL:  "https://example.com/new-avatar.jpg",
	}

	stateGen.On("GetState", mock.Anything, "valid-state").Return(true, nil)
	oauthProvider.On("ExchangeCode", mock.Anything, "auth-code").Return(oauthUserInfo, nil)
	userRepo.On("FindByProvider", mock.Anything, domain.AuthProviderGoogle, "google-123").Return(existingUser, nil)
	userRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)
	tokenService.On("GenerateToken", int64(1), "existing@gmail.com").Return("access-token", nil)
	tokenService.On("GenerateRefreshToken", int64(1), "existing@gmail.com").Return("refresh-token", nil)

	oauthProviders := map[domain.AuthProvider]ports.OAuthProvider{
		domain.AuthProviderGoogle: oauthProvider,
	}

	service := NewAuthService(userRepo, nil, tokenService, stateGen, oauthProviders)

	ctx := context.Background()
	resp, err := service.HandleOAuthCallback(ctx, domain.AuthProviderGoogle, "auth-code", "valid-state")

	require.NoError(t, err)
	assert.NotNil(t, resp)

	stateGen.AssertExpectations(t)
	oauthProvider.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	tokenService.AssertExpectations(t)
}

func TestAuthService_HandleOAuthCallback_InvalidState(t *testing.T) {
	stateGen := new(MockStateGenerator)

	stateGen.On("GetState", mock.Anything, "invalid-state").Return(false, nil)

	service := NewAuthService(nil, nil, nil, stateGen)

	ctx := context.Background()
	resp, err := service.HandleOAuthCallback(ctx, domain.AuthProviderGoogle, "auth-code", "invalid-state")

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrOAuthStateMismatch)
	assert.Nil(t, resp)

	stateGen.AssertExpectations(t)
}

func TestAuthService_RefreshToken_Success(t *testing.T) {
	tokenService := new(MockTokenService)

	tokenService.On("RefreshToken", "valid-refresh-token").Return("new-access-token", nil)

	service := NewAuthService(nil, nil, tokenService, nil)

	ctx := context.Background()
	newToken, err := service.RefreshToken(ctx, "valid-refresh-token")

	require.NoError(t, err)
	assert.Equal(t, "new-access-token", newToken)

	tokenService.AssertExpectations(t)
}

func TestAuthService_RefreshToken_InvalidToken(t *testing.T) {
	tokenService := new(MockTokenService)

	tokenService.On("RefreshToken", "invalid-token").Return("", errors.New("invalid token"))

	service := NewAuthService(nil, nil, tokenService, nil)

	ctx := context.Background()
	newToken, err := service.RefreshToken(ctx, "invalid-token")

	assert.Error(t, err)
	assert.Empty(t, newToken)

	tokenService.AssertExpectations(t)
}

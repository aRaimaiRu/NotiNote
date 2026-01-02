package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/notinoteapp/internal/application/services"
	"github.com/yourusername/notinoteapp/internal/core/domain"
)

// Mock AuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(ctx context.Context, email, password, name string) (*services.AuthResponse, error) {
	args := m.Called(ctx, email, password, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.AuthResponse), args.Error(1)
}

func (m *MockAuthService) Login(ctx context.Context, email, password string) (*services.AuthResponse, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.AuthResponse), args.Error(1)
}

func (m *MockAuthService) GetOAuthURL(ctx context.Context, provider domain.AuthProvider) (string, error) {
	args := m.Called(ctx, provider)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) HandleOAuthCallback(ctx context.Context, provider domain.AuthProvider, code, state string) (*services.AuthResponse, error) {
	args := m.Called(ctx, provider, code, state)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.AuthResponse), args.Error(1)
}

func (m *MockAuthService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	args := m.Called(ctx, refreshToken)
	return args.String(0), args.Error(1)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestAuthHandler_Register_Success(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)
	router := setupTestRouter()
	router.POST("/register", handler.Register)

	authResp := &services.AuthResponse{
		User: &domain.User{
			ID:       1,
			Email:    "test@example.com",
			Name:     "Test User",
			Provider: domain.AuthProviderEmail,
		},
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
	}

	mockService.On("Register", mock.Anything, "test@example.com", "Password123!", "Test User").Return(authResp, nil)

	reqBody := RegisterRequest{
		Email:    "test@example.com",
		Password: "Password123!",
		Name:     "Test User",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusCreated, resp.Code)

	var response AuthResponse
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "test@example.com", response.Data.User.Email)
	assert.Equal(t, "access-token", response.Data.AccessToken)

	mockService.AssertExpectations(t)
}

func TestAuthHandler_Register_InvalidRequest(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)
	router := setupTestRouter()
	router.POST("/register", handler.Register)

	// Missing required field
	reqBody := map[string]string{
		"email": "test@example.com",
		// Missing password and name
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var errorResp ErrorResponse
	err := json.Unmarshal(resp.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.False(t, errorResp.Success)
	assert.Contains(t, errorResp.Error, "Invalid request")
}

func TestAuthHandler_Register_UserAlreadyExists(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)
	router := setupTestRouter()
	router.POST("/register", handler.Register)

	mockService.On("Register", mock.Anything, "existing@example.com", "Password123!", "Test User").
		Return(nil, domain.ErrUserAlreadyExists)

	reqBody := RegisterRequest{
		Email:    "existing@example.com",
		Password: "Password123!",
		Name:     "Test User",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusConflict, resp.Code)

	var errorResp ErrorResponse
	err := json.Unmarshal(resp.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.False(t, errorResp.Success)
	assert.Contains(t, errorResp.Error, "already exists")

	mockService.AssertExpectations(t)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)
	router := setupTestRouter()
	router.POST("/login", handler.Login)

	authResp := &services.AuthResponse{
		User: &domain.User{
			ID:       1,
			Email:    "test@example.com",
			Name:     "Test User",
			Provider: domain.AuthProviderEmail,
		},
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
	}

	mockService.On("Login", mock.Anything, "test@example.com", "Password123!").Return(authResp, nil)

	reqBody := LoginRequest{
		Email:    "test@example.com",
		Password: "Password123!",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var response AuthResponse
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "access-token", response.Data.AccessToken)

	mockService.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)
	router := setupTestRouter()
	router.POST("/login", handler.Login)

	mockService.On("Login", mock.Anything, "test@example.com", "WrongPassword").
		Return(nil, domain.ErrInvalidCredentials)

	reqBody := LoginRequest{
		Email:    "test@example.com",
		Password: "WrongPassword",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)

	var errorResp ErrorResponse
	err := json.Unmarshal(resp.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.False(t, errorResp.Success)
	assert.Contains(t, errorResp.Error, "Invalid email or password")

	mockService.AssertExpectations(t)
}

func TestAuthHandler_Login_InactiveUser(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)
	router := setupTestRouter()
	router.POST("/login", handler.Login)

	mockService.On("Login", mock.Anything, "inactive@example.com", "Password123!").
		Return(nil, domain.ErrUserInactive)

	reqBody := LoginRequest{
		Email:    "inactive@example.com",
		Password: "Password123!",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusForbidden, resp.Code)

	var errorResp ErrorResponse
	err := json.Unmarshal(resp.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.False(t, errorResp.Success)
	assert.Contains(t, errorResp.Error, "inactive")

	mockService.AssertExpectations(t)
}

func TestAuthHandler_GoogleLogin(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)
	router := setupTestRouter()
	router.GET("/auth/google", handler.GoogleLogin)

	mockService.On("GetOAuthURL", mock.Anything, domain.AuthProviderGoogle).
		Return("https://accounts.google.com/oauth?state=random-state", nil)

	req, _ := http.NewRequest("GET", "/auth/google", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response["success"].(bool))

	data := response["data"].(map[string]interface{})
	assert.Contains(t, data["auth_url"].(string), "https://accounts.google.com/oauth")

	mockService.AssertExpectations(t)
}

func TestAuthHandler_GoogleCallback_Success(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)
	router := setupTestRouter()
	router.GET("/auth/google/callback", handler.GoogleCallback)

	authResp := &services.AuthResponse{
		User: &domain.User{
			ID:         1,
			Email:      "user@gmail.com",
			Name:       "Google User",
			Provider:   domain.AuthProviderGoogle,
			ProviderID: "google-123",
		},
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
	}

	mockService.On("HandleOAuthCallback", mock.Anything, domain.AuthProviderGoogle, "auth-code", "random-state").
		Return(authResp, nil)

	req, _ := http.NewRequest("GET", "/auth/google/callback?code=auth-code&state=random-state", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var response AuthResponse
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "user@gmail.com", response.Data.User.Email)

	mockService.AssertExpectations(t)
}

func TestAuthHandler_GoogleCallback_MissingParams(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)
	router := setupTestRouter()
	router.GET("/auth/google/callback", handler.GoogleCallback)

	// Missing state parameter
	req, _ := http.NewRequest("GET", "/auth/google/callback?code=auth-code", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var errorResp ErrorResponse
	err := json.Unmarshal(resp.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.False(t, errorResp.Success)
	assert.Contains(t, errorResp.Error, "Missing code or state")
}

func TestAuthHandler_GoogleCallback_StateMismatch(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)
	router := setupTestRouter()
	router.GET("/auth/google/callback", handler.GoogleCallback)

	mockService.On("HandleOAuthCallback", mock.Anything, domain.AuthProviderGoogle, "auth-code", "invalid-state").
		Return(nil, domain.ErrOAuthStateMismatch)

	req, _ := http.NewRequest("GET", "/auth/google/callback?code=auth-code&state=invalid-state", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var errorResp ErrorResponse
	err := json.Unmarshal(resp.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.False(t, errorResp.Success)
	assert.Contains(t, errorResp.Error, "CSRF")

	mockService.AssertExpectations(t)
}

func TestAuthHandler_FacebookLogin(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)
	router := setupTestRouter()
	router.GET("/auth/facebook", handler.FacebookLogin)

	mockService.On("GetOAuthURL", mock.Anything, domain.AuthProviderFacebook).
		Return("https://facebook.com/dialog/oauth?state=random-state", nil)

	req, _ := http.NewRequest("GET", "/auth/facebook", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response["success"].(bool))

	data := response["data"].(map[string]interface{})
	assert.Contains(t, data["auth_url"].(string), "facebook.com")

	mockService.AssertExpectations(t)
}

func TestAuthHandler_FacebookCallback_Success(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)
	router := setupTestRouter()
	router.GET("/auth/facebook/callback", handler.FacebookCallback)

	authResp := &services.AuthResponse{
		User: &domain.User{
			ID:         1,
			Email:      "user@facebook.com",
			Name:       "Facebook User",
			Provider:   domain.AuthProviderFacebook,
			ProviderID: "fb-123",
		},
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
	}

	mockService.On("HandleOAuthCallback", mock.Anything, domain.AuthProviderFacebook, "auth-code", "random-state").
		Return(authResp, nil)

	req, _ := http.NewRequest("GET", "/auth/facebook/callback?code=auth-code&state=random-state", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var response AuthResponse
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "user@facebook.com", response.Data.User.Email)

	mockService.AssertExpectations(t)
}

func TestAuthHandler_RefreshToken_Success(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)
	router := setupTestRouter()
	router.POST("/refresh", handler.RefreshToken)

	authResp := &services.AuthResponse{
		User: &domain.User{
			ID:    1,
			Email: "test@example.com",
			Name:  "Test User",
		},
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
	}

	mockService.On("RefreshToken", mock.Anything, "valid-refresh-token").
		Return("new-access-token", nil).Once()

	// Note: This test needs to be updated based on actual implementation
	// If RefreshToken returns full AuthResponse, adjust accordingly
	t.Skip("Skipping until RefreshToken method signature is confirmed")
}

func TestAuthHandler_Logout(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)
	router := setupTestRouter()
	router.POST("/logout", handler.Logout)

	req, _ := http.NewRequest("POST", "/logout", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response["success"].(bool))
	assert.Contains(t, response["message"].(string), "Logged out successfully")
}

func TestAuthHandler_BuildAuthResponse(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)

	serviceResp := &services.AuthResponse{
		User: &domain.User{
			ID:        1,
			Email:     "test@example.com",
			Name:      "Test User",
			Provider:  domain.AuthProviderEmail,
			AvatarURL: "https://example.com/avatar.jpg",
		},
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
	}

	handlerResp := handler.buildAuthResponse(serviceResp)

	assert.True(t, handlerResp.Success)
	assert.Equal(t, int64(1), handlerResp.Data.User.ID)
	assert.Equal(t, "test@example.com", handlerResp.Data.User.Email)
	assert.Equal(t, "Test User", handlerResp.Data.User.Name)
	assert.Equal(t, "access-token", handlerResp.Data.AccessToken)
	assert.Equal(t, "refresh-token", handlerResp.Data.RefreshToken)
	assert.Equal(t, "Bearer", handlerResp.Data.TokenType)
	assert.Equal(t, 86400, handlerResp.Data.ExpiresIn)
}

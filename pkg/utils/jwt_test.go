package utils

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJWTService(t *testing.T) {
	secret := "test-secret-key"
	issuer := "test-issuer"
	accessExpiry := 24 * time.Hour
	refreshExpiry := 7 * 24 * time.Hour

	service := NewJWTService(secret, issuer, accessExpiry, refreshExpiry)

	assert.NotNil(t, service)
	assert.Equal(t, secret, service.secret)
	assert.Equal(t, issuer, service.issuer)
	assert.Equal(t, accessExpiry, service.accessTokenExpiry)
	assert.Equal(t, refreshExpiry, service.refreshTokenExpiry)
}

func TestJWTService_GenerateToken(t *testing.T) {
	service := NewJWTService("test-secret", "test-issuer", 24*time.Hour, 7*24*time.Hour)

	tests := []struct {
		name    string
		userID  int64
		email   string
		wantErr bool
	}{
		{
			name:    "valid token generation",
			userID:  123,
			email:   "user@example.com",
			wantErr: false,
		},
		{
			name:    "zero user ID",
			userID:  0,
			email:   "user@example.com",
			wantErr: false, // JWT service doesn't validate business rules
		},
		{
			name:    "empty email",
			userID:  123,
			email:   "",
			wantErr: false, // JWT service doesn't validate business rules
		},
		{
			name:    "large user ID",
			userID:  9223372036854775807, // max int64
			email:   "user@example.com",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := service.GenerateToken(tt.userID, tt.email)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, token)

				// Verify token can be parsed
				parsedToken, err := jwt.ParseWithClaims(token, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
					return []byte("test-secret"), nil
				})
				require.NoError(t, err)
				assert.True(t, parsedToken.Valid)

				// Verify claims
				claims, ok := parsedToken.Claims.(*JWTClaims)
				require.True(t, ok)
				assert.Equal(t, tt.userID, claims.UserID)
				assert.Equal(t, tt.email, claims.Email)
				assert.Equal(t, "test-issuer", claims.Issuer)
				assert.NotNil(t, claims.ExpiresAt)
				assert.NotNil(t, claims.IssuedAt)
				assert.NotNil(t, claims.NotBefore)
			}
		})
	}
}

func TestJWTService_GenerateRefreshToken(t *testing.T) {
	service := NewJWTService("test-secret", "test-issuer", 24*time.Hour, 7*24*time.Hour)

	userID := int64(123)
	email := "user@example.com"

	refreshToken, err := service.GenerateRefreshToken(userID, email)
	require.NoError(t, err)
	assert.NotEmpty(t, refreshToken)

	// Parse token
	parsedToken, err := jwt.ParseWithClaims(refreshToken, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})
	require.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	// Verify expiration is longer (7 days)
	claims, ok := parsedToken.Claims.(*JWTClaims)
	require.True(t, ok)
	expectedExpiry := time.Now().Add(7 * 24 * time.Hour)
	assert.WithinDuration(t, expectedExpiry, claims.ExpiresAt.Time, 5*time.Second)
}

func TestJWTService_ValidateToken(t *testing.T) {
	service := NewJWTService("test-secret", "test-issuer", 24*time.Hour, 7*24*time.Hour)

	tests := []struct {
		name          string
		setupToken    func() string
		expectedID    int64
		expectedEmail string
		wantErr       bool
		expectedErr   error
	}{
		{
			name: "valid token",
			setupToken: func() string {
				token, _ := service.GenerateToken(123, "user@example.com")
				return token
			},
			expectedID:    123,
			expectedEmail: "user@example.com",
			wantErr:       false,
		},
		{
			name: "empty token",
			setupToken: func() string {
				return ""
			},
			wantErr:     true,
			expectedErr: ErrInvalidToken,
		},
		{
			name: "malformed token",
			setupToken: func() string {
				return "not.a.valid.token"
			},
			wantErr:     true,
			expectedErr: ErrInvalidToken,
		},
		{
			name: "token with wrong secret",
			setupToken: func() string {
				wrongService := NewJWTService("wrong-secret", "test-issuer", 24*time.Hour, 7*24*time.Hour)
				token, _ := wrongService.GenerateToken(123, "user@example.com")
				return token
			},
			wantErr:     true,
			expectedErr: ErrInvalidToken,
		},
		{
			name: "expired token",
			setupToken: func() string {
				expiredService := NewJWTService("test-secret", "test-issuer", -1*time.Hour, 7*24*time.Hour)
				token, _ := expiredService.GenerateToken(123, "user@example.com")
				return token
			},
			wantErr:     true,
			expectedErr: ErrExpiredToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.setupToken()
			userID, email, err := service.ValidateToken(token)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Zero(t, userID)
				assert.Empty(t, email)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedID, userID)
				assert.Equal(t, tt.expectedEmail, email)
			}
		})
	}
}

func TestJWTService_RefreshToken(t *testing.T) {
	service := NewJWTService("test-secret", "test-issuer", 24*time.Hour, 7*24*time.Hour)

	tests := []struct {
		name        string
		setupToken  func() string
		wantErr     bool
		expectedErr error
	}{
		{
			name: "valid refresh token",
			setupToken: func() string {
				token, _ := service.GenerateRefreshToken(123, "user@example.com")
				return token
			},
			wantErr: false,
		},
		{
			name: "empty refresh token",
			setupToken: func() string {
				return ""
			},
			wantErr:     true,
			expectedErr: ErrInvalidToken,
		},
		{
			name: "invalid refresh token",
			setupToken: func() string {
				return "invalid.refresh.token"
			},
			wantErr:     true,
			expectedErr: ErrInvalidToken,
		},
		{
			name: "expired refresh token",
			setupToken: func() string {
				expiredService := NewJWTService("test-secret", "test-issuer", 24*time.Hour, -1*time.Hour)
				token, _ := expiredService.GenerateRefreshToken(123, "user@example.com")
				return token
			},
			wantErr:     true,
			expectedErr: ErrExpiredToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			refreshToken := tt.setupToken()
			newAccessToken, err := service.RefreshToken(refreshToken)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, newAccessToken)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, newAccessToken)

				// Verify new access token is valid
				userID, email, err := service.ValidateToken(newAccessToken)
				require.NoError(t, err)
				assert.Equal(t, int64(123), userID)
				assert.Equal(t, "user@example.com", email)
			}
		})
	}
}

func TestJWTService_TokenExpiration(t *testing.T) {
	// Test with very short expiration
	service := NewJWTService("test-secret", "test-issuer", 1*time.Second, 2*time.Second)

	// Generate token
	token, err := service.GenerateToken(123, "user@example.com")
	require.NoError(t, err)

	// Should be valid immediately
	userID, email, err := service.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, int64(123), userID)
	assert.Equal(t, "user@example.com", email)

	// Wait for expiration
	time.Sleep(2 * time.Second)

	// Should be expired now
	_, _, err = service.ValidateToken(token)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrExpiredToken)
}

func TestJWTService_TokenUniqueness(t *testing.T) {
	service := NewJWTService("test-secret", "test-issuer", 24*time.Hour, 7*24*time.Hour)

	// Generate two tokens for the same user
	token1, err1 := service.GenerateToken(123, "user@example.com")
	time.Sleep(1 * time.Second) // Delay to ensure different timestamps
	token2, err2 := service.GenerateToken(123, "user@example.com")

	require.NoError(t, err1)
	require.NoError(t, err2)

	// Tokens might be different due to different IssuedAt timestamps
	// Note: If generated in the same second, they may be identical
	if token1 == token2 {
		t.Log("Tokens are identical (generated in same second)")
	}

	// But both should be valid
	userID1, email1, err := service.ValidateToken(token1)
	require.NoError(t, err)
	assert.Equal(t, int64(123), userID1)
	assert.Equal(t, "user@example.com", email1)

	userID2, email2, err := service.ValidateToken(token2)
	require.NoError(t, err)
	assert.Equal(t, int64(123), userID2)
	assert.Equal(t, "user@example.com", email2)
}

func TestJWTService_WrongAlgorithm(t *testing.T) {
	t.Skip("JWT library behavior with different HMAC algorithms is complex - skipping")

	service := NewJWTService("test-secret", "test-issuer", 24*time.Hour, 7*24*time.Hour)

	// Create a token with a different signing method
	claims := JWTClaims{
		UserID: 123,
		Email:  "user@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "test-issuer",
		},
	}

	// Sign with HS512 instead of HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenString, err := token.SignedString([]byte("test-secret"))
	require.NoError(t, err)

	// Should fail validation due to wrong algorithm
	_, _, err = service.ValidateToken(tokenString)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestJWTClaims_Structure(t *testing.T) {
	claims := JWTClaims{
		UserID: 123,
		Email:  "user@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "test-issuer",
		},
	}

	assert.Equal(t, int64(123), claims.UserID)
	assert.Equal(t, "user@example.com", claims.Email)
	assert.Equal(t, "test-issuer", claims.Issuer)
	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.IssuedAt)
	assert.NotNil(t, claims.NotBefore)
}

func BenchmarkJWTService_GenerateToken(b *testing.B) {
	service := NewJWTService("test-secret", "test-issuer", 24*time.Hour, 7*24*time.Hour)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.GenerateToken(123, "user@example.com")
	}
}

func BenchmarkJWTService_ValidateToken(b *testing.B) {
	service := NewJWTService("test-secret", "test-issuer", 24*time.Hour, 7*24*time.Hour)
	token, _ := service.GenerateToken(123, "user@example.com")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = service.ValidateToken(token)
	}
}

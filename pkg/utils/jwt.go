package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// JWTClaims represents the JWT claims
type JWTClaims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// JWTService handles JWT token operations
type JWTService struct {
	secret              string
	issuer              string
	accessTokenExpiry   time.Duration
	refreshTokenExpiry  time.Duration
}

// NewJWTService creates a new JWT service
func NewJWTService(secret, issuer string, accessExpiry, refreshExpiry time.Duration) *JWTService {
	return &JWTService{
		secret:              secret,
		issuer:              issuer,
		accessTokenExpiry:   accessExpiry,
		refreshTokenExpiry:  refreshExpiry,
	}
}

// GenerateToken generates a JWT access token for a user
func (j *JWTService) GenerateToken(userID int64, email string) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    j.issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secret))
}

// GenerateRefreshToken generates a JWT refresh token
func (j *JWTService) GenerateRefreshToken(userID int64, email string) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(j.refreshTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    j.issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secret))
}

// ValidateToken validates a JWT token and returns claims
func (j *JWTService) ValidateToken(tokenString string) (userID int64, email string, err error) {
	claims := &JWTClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(j.secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return 0, "", ErrExpiredToken
		}
		return 0, "", ErrInvalidToken
	}

	if !token.Valid {
		return 0, "", ErrInvalidToken
	}

	return claims.UserID, claims.Email, nil
}

// RefreshToken generates a new access token from a refresh token
func (j *JWTService) RefreshToken(refreshToken string) (string, error) {
	// Validate refresh token
	userID, email, err := j.ValidateToken(refreshToken)
	if err != nil {
		return "", err
	}

	// Generate new access token
	return j.GenerateToken(userID, email)
}

package ports

import (
	"context"

	"github.com/yourusername/notinoteapp/internal/core/domain"
)

// OAuthProvider defines the interface for OAuth authentication providers
type OAuthProvider interface {
	// GetAuthURL generates the OAuth authorization URL with state
	GetAuthURL(state string) string

	// ExchangeCode exchanges authorization code for access token and retrieves user info
	ExchangeCode(ctx context.Context, code string) (*domain.OAuthUserInfo, error)

	// GetProviderName returns the provider name (google, facebook, etc.)
	GetProviderName() domain.AuthProvider
}

// PasswordHasher defines the interface for password hashing
type PasswordHasher interface {
	// HashPassword hashes a plain text password
	HashPassword(password string) (string, error)

	// CheckPassword compares a plain text password with a hash
	CheckPassword(password, hash string) bool
}

// TokenService defines the interface for JWT token operations
type TokenService interface {
	// GenerateToken generates a JWT token for a user
	GenerateToken(userID int64, email string) (string, error)

	// GenerateRefreshToken generates a refresh token
	GenerateRefreshToken(userID int64, email string) (string, error)

	// ValidateToken validates a JWT token and returns claims
	ValidateToken(token string) (userID int64, email string, err error)

	// RefreshToken generates a new access token from a refresh token
	RefreshToken(refreshToken string) (string, error)
}

// StateGenerator defines the interface for OAuth state generation and validation
type StateGenerator interface {
	// GenerateState generates a random state string for CSRF protection
	GenerateState() (string, error)

	// ValidateState validates that a state matches expected value
	ValidateState(state, expected string) bool

	// StoreState temporarily stores state (e.g., in Redis) with expiration
	StoreState(ctx context.Context, state string, ttl int) error

	// GetState retrieves and deletes stored state (one-time use)
	GetState(ctx context.Context, state string) (bool, error)
}

// EmailService defines the interface for sending emails
type EmailService interface {
	// SendWelcomeEmail sends a welcome email to new users
	SendWelcomeEmail(ctx context.Context, to, name string) error

	// SendPasswordResetEmail sends a password reset email
	SendPasswordResetEmail(ctx context.Context, to, resetToken string) error

	// SendNotificationEmail sends a notification email
	SendNotificationEmail(ctx context.Context, to, subject, body string) error
}

// NotificationSender defines the interface for sending push notifications
type NotificationSender interface {
	// SendPushNotification sends a push notification to a device
	SendPushNotification(ctx context.Context, deviceToken, title, body string, data map[string]string) error

	// SendToMultipleDevices sends a push notification to multiple devices
	SendToMultipleDevices(ctx context.Context, deviceTokens []string, title, body string, data map[string]string) error
}

// CacheService defines the interface for caching operations
type CacheService interface {
	// Set stores a value in cache with TTL
	Set(ctx context.Context, key string, value interface{}, ttl int) error

	// Get retrieves a value from cache
	Get(ctx context.Context, key string) (interface{}, error)

	// Delete removes a value from cache
	Delete(ctx context.Context, key string) error

	// Exists checks if a key exists in cache
	Exists(ctx context.Context, key string) (bool, error)
}

// QueueService defines the interface for queue operations
type QueueService interface {
	// Push adds an item to the queue
	Push(ctx context.Context, queueName string, data interface{}) error

	// Pop retrieves and removes an item from the queue (blocking)
	Pop(ctx context.Context, queueName string, timeout int) (interface{}, error)

	// PushWithDelay adds an item to a delayed queue
	PushWithDelay(ctx context.Context, queueName string, data interface{}, delaySeconds int) error

	// GetQueueDepth returns the number of items in a queue
	GetQueueDepth(ctx context.Context, queueName string) (int64, error)
}

package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration
type Config struct {
	Server       ServerConfig
	Database     DatabaseConfig
	Redis        RedisConfig
	JWT          JWTConfig
	OAuth        OAuthConfig
	CORS         CORSConfig
	RateLimit    RateLimitConfig
	Notification NotificationConfig
	FCM          FCMConfig
	Log          LogConfig
}

// FCMConfig holds Firebase Cloud Messaging configuration
type FCMConfig struct {
	CredentialsFile string
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         string
	Mode         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string
	Port            string
	Name            string
	User            string
	Password        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
	PoolSize int
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret            string
	Expiration        time.Duration
	RefreshExpiration time.Duration
}

// OAuthConfig holds OAuth configuration
type OAuthConfig struct {
	Google   OAuthProviderConfig
	Facebook OAuthProviderConfig
	State    StateConfig
}

// OAuthProviderConfig holds OAuth provider configuration
type OAuthProviderConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// StateConfig holds OAuth state configuration
type StateConfig struct {
	Secret string
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	RequestsPerSecond int
	Burst             int
}

// NotificationConfig holds notification system configuration
type NotificationConfig struct {
	SchedulerInterval time.Duration
	WorkerCount       int
	MaxRetries        int
	RetryBackoff      time.Duration
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level  string
	Format string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			Mode:         getEnv("GIN_MODE", "debug"),
			ReadTimeout:  parseDuration(getEnv("SERVER_READ_TIMEOUT", "30s"), 30*time.Second),
			WriteTimeout: parseDuration(getEnv("SERVER_WRITE_TIMEOUT", "30s"), 30*time.Second),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			Name:            getEnv("DB_NAME", "notinoteapp"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", ""),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    parseInt(getEnv("DB_MAX_OPEN_CONNS", "25"), 25),
			MaxIdleConns:    parseInt(getEnv("DB_MAX_IDLE_CONNS", "5"), 5),
			ConnMaxLifetime: parseDuration(getEnv("DB_CONN_MAX_LIFETIME", "5m"), 5*time.Minute),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       parseInt(getEnv("REDIS_DB", "0"), 0),
			PoolSize: parseInt(getEnv("REDIS_POOL_SIZE", "10"), 10),
		},
		JWT: JWTConfig{
			Secret:            getEnv("JWT_SECRET", "change_this_secret_key"),
			Expiration:        parseDuration(getEnv("JWT_EXPIRATION", "24h"), 24*time.Hour),
			RefreshExpiration: parseDuration(getEnv("JWT_REFRESH_EXPIRATION", "168h"), 168*time.Hour),
		},
		OAuth: OAuthConfig{
			Google: OAuthProviderConfig{
				ClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
				ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
				RedirectURL:  getEnv("GOOGLE_REDIRECT_URL", ""),
			},
			Facebook: OAuthProviderConfig{
				ClientID:     getEnv("FACEBOOK_APP_ID", ""),
				ClientSecret: getEnv("FACEBOOK_APP_SECRET", ""),
				RedirectURL:  getEnv("FACEBOOK_REDIRECT_URL", ""),
			},
			State: StateConfig{
				Secret: getEnv("OAUTH_STATE_SECRET", "change_this_state_secret"),
			},
		},
		CORS: CORSConfig{
			AllowedOrigins: parseStringSlice(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:8080")),
			AllowedMethods: parseStringSlice(getEnv("CORS_ALLOWED_METHODS", "GET,POST,PUT,DELETE,OPTIONS")),
			AllowedHeaders: parseStringSlice(getEnv("CORS_ALLOWED_HEADERS", "Authorization,Content-Type")),
		},
		RateLimit: RateLimitConfig{
			RequestsPerSecond: parseInt(getEnv("RATE_LIMIT_REQUESTS_PER_SECOND", "10"), 10),
			Burst:             parseInt(getEnv("RATE_LIMIT_BURST", "20"), 20),
		},
		Notification: NotificationConfig{
			SchedulerInterval: parseDuration(getEnv("NOTIFICATION_SCHEDULER_INTERVAL", "30s"), 30*time.Second),
			WorkerCount:       parseInt(getEnv("NOTIFICATION_WORKER_COUNT", "5"), 5),
			MaxRetries:        parseInt(getEnv("NOTIFICATION_MAX_RETRIES", "3"), 3),
			RetryBackoff:      parseDuration(getEnv("NOTIFICATION_RETRY_BACKOFF", "1m"), 1*time.Minute),
		},
		FCM: FCMConfig{
			CredentialsFile: getEnv("FCM_CREDENTIALS_FILE", ""),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}

	// Validate required configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.JWT.Secret == "change_this_secret_key" {
		return fmt.Errorf("JWT_SECRET must be set to a secure value")
	}
	if c.Database.Password == "" {
		return fmt.Errorf("DB_PASSWORD must be set")
	}
	return nil
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseInt(s string, defaultValue int) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return defaultValue
}

func parseDuration(s string, defaultValue time.Duration) time.Duration {
	if d, err := time.ParseDuration(s); err == nil {
		return d
	}
	return defaultValue
}

func parseStringSlice(s string) []string {
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config holds Redis configuration
type Config struct {
	Host     string
	Port     string
	Password string
	DB       int
	PoolSize int
}

// NewClient creates a new Redis client
func NewClient(config Config) (*redis.Client, error) {
	addr := fmt.Sprintf("%s:%s", config.Host, config.Port)

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: config.Password,
		DB:       config.DB,
		PoolSize: config.PoolSize,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return client, nil
}

// Close closes the Redis client
func Close(client *redis.Client) error {
	return client.Close()
}

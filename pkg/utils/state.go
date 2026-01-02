package utils

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStateGenerator implements OAuth state generation and validation using Redis
type RedisStateGenerator struct {
	redis  *redis.Client
	prefix string
}

// NewRedisStateGenerator creates a new Redis-based state generator
func NewRedisStateGenerator(redisClient *redis.Client) *RedisStateGenerator {
	return &RedisStateGenerator{
		redis:  redisClient,
		prefix: "oauth:state:",
	}
}

// GenerateState generates a random state string for CSRF protection
func (s *RedisStateGenerator) GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random state: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// ValidateState validates that a state matches expected value
func (s *RedisStateGenerator) ValidateState(state, expected string) bool {
	return state == expected && state != ""
}

// StoreState temporarily stores state in Redis with expiration (TTL in seconds)
func (s *RedisStateGenerator) StoreState(ctx context.Context, state string, ttl int) error {
	key := s.prefix + state
	duration := time.Duration(ttl) * time.Second

	err := s.redis.Set(ctx, key, "1", duration).Err()
	if err != nil {
		return fmt.Errorf("failed to store state in redis: %w", err)
	}

	return nil
}

// GetState retrieves and deletes stored state (one-time use)
// Returns true if state exists and was deleted, false otherwise
func (s *RedisStateGenerator) GetState(ctx context.Context, state string) (bool, error) {
	key := s.prefix + state

	// Get the value
	val, err := s.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil // State doesn't exist
	}
	if err != nil {
		return false, fmt.Errorf("failed to get state from redis: %w", err)
	}

	// Delete the state (one-time use)
	if err := s.redis.Del(ctx, key).Err(); err != nil {
		return false, fmt.Errorf("failed to delete state from redis: %w", err)
	}

	return val == "1", nil
}

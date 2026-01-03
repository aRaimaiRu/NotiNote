package main

import (
	"time"

	"github.com/yourusername/notinoteapp/pkg/logger"
)

func main() {
	// Test different formats
	println("=== Testing Text Format (Colorful) ===\n")
	logger.Init("debug", "text")

	// Test different log levels
	logger.Debug("This is a debug message for development")

	logger.Info("Application started successfully")

	logger.WithFields(map[string]interface{}{
		"user_id":  12345,
		"username": "johndoe",
		"email":    "john@example.com",
	}).Info("User logged in")

	logger.WithFields(map[string]interface{}{
		"method":   "POST",
		"path":     "/api/v1/notes",
		"status":   201,
		"latency":  "15.3ms",
		"size":     "1.2KB",
	}).Info("HTTP request completed")

	logger.Warn("Database connection pool is running low")

	logger.WithFields(map[string]interface{}{
		"available": 2,
		"max":       10,
		"threshold": 5,
	}).Warn("Connection pool usage warning")

	logger.Error("Failed to send notification")

	logger.WithFields(map[string]interface{}{
		"notification_id": 789,
		"user_id":         456,
		"error":           "FCM token invalid",
		"retry_count":     3,
	}).Error("Notification delivery failed")

	logger.WithFields(map[string]interface{}{
		"request_id": "req-abc-123",
		"endpoint":   "/api/v1/auth/login",
		"duration":   time.Duration(500 * time.Millisecond),
		"success":    true,
		"attempts":   1,
	}).Info("Authentication successful")

	// Test JSON format
	println("\n\n=== Testing JSON Format (Production) ===\n")
	logger.Init("info", "json")

	logger.Info("Application started in production mode")

	logger.WithFields(map[string]interface{}{
		"method":   "GET",
		"path":     "/api/v1/notes",
		"status":   200,
		"latency":  "8.5ms",
		"user_id":  12345,
	}).Info("HTTP request")

	logger.WithFields(map[string]interface{}{
		"error":    "database connection timeout",
		"database": "postgres",
		"timeout":  "30s",
	}).Error("Database error occurred")

	println("\n=== Logger Test Complete ===")
}

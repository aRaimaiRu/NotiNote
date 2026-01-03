package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yourusername/notinoteapp/pkg/logger"
)

// Logger returns a gin middleware for logging HTTP requests
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get status code
		statusCode := c.Writer.Status()

		// Build full path with query string
		fullPath := path
		if raw != "" {
			fullPath = path + "?" + raw
		}

		// Get response size
		responseSize := c.Writer.Size()

		// Prepare log fields
		fields := logrus.Fields{
			"status":   statusCode,
			"method":   c.Request.Method,
			"path":     fullPath,
			"ip":       c.ClientIP(),
			"latency":  formatLatency(latency),
			"size":     formatBytes(responseSize),
		}

		// Add user ID if authenticated
		if userID, exists := c.Get("user_id"); exists {
			fields["user_id"] = userID
		}

		// Add error if present
		if len(c.Errors) > 0 {
			fields["error"] = c.Errors.String()
		}

		// Determine log level based on status code
		msg := fmt.Sprintf("%s %s", c.Request.Method, fullPath)
		entry := logger.WithFields(fields)

		switch {
		case statusCode >= 500:
			entry.Error(msg)
		case statusCode >= 400:
			entry.Warn(msg)
		case statusCode >= 300:
			entry.Info(msg)
		default:
			entry.Info(msg)
		}
	}
}

// formatLatency formats the latency duration for better readability
func formatLatency(d time.Duration) string {
	switch {
	case d < time.Microsecond:
		return fmt.Sprintf("%dns", d.Nanoseconds())
	case d < time.Millisecond:
		return fmt.Sprintf("%.2fμs", float64(d.Nanoseconds())/1000.0)
	case d < time.Second:
		return fmt.Sprintf("%.2fms", float64(d.Nanoseconds())/1000000.0)
	default:
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
}

// formatBytes formats the byte size for better readability
func formatBytes(bytes int) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%dB", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

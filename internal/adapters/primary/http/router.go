package http

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/yourusername/notinoteapp/internal/adapters/primary/http/handlers"
	"github.com/yourusername/notinoteapp/internal/adapters/primary/http/middleware"
	"github.com/yourusername/notinoteapp/pkg/config"
)

// RouterConfig holds router configuration
type RouterConfig struct {
	AuthHandler *handlers.AuthHandler
	Config      *config.Config
}

// SetupRouter sets up the HTTP router with all routes
func SetupRouter(cfg RouterConfig) *gin.Engine {
	// Set Gin mode
	gin.SetMode(cfg.Config.Server.Mode)

	// Create router
	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())

	// CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.Config.CORS.AllowedOrigins,
		AllowMethods:     cfg.Config.CORS.AllowedMethods,
		AllowHeaders:     cfg.Config.CORS.AllowedHeaders,
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
			"time":   time.Now().UTC(),
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", cfg.AuthHandler.Register)
			auth.POST("/login", cfg.AuthHandler.Login)
			auth.POST("/refresh", cfg.AuthHandler.RefreshToken)

			// OAuth verification routes (frontend-initiated)
			auth.POST("/google/verify", cfg.AuthHandler.VerifyGoogleToken)
			auth.POST("/facebook/verify", cfg.AuthHandler.VerifyFacebookToken)
		}

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(cfg.Config.JWT.Secret))
		{
			// User routes
			protected.GET("/me", cfg.AuthHandler.GetCurrentUser)

			// Notes routes (placeholder for future implementation)
			// notes := protected.Group("/notes")
			// {
			// 	notes.GET("", noteHandler.List)
			// 	notes.POST("", noteHandler.Create)
			// 	notes.GET("/:id", noteHandler.Get)
			// 	notes.PUT("/:id", noteHandler.Update)
			// 	notes.DELETE("/:id", noteHandler.Delete)
			// }

			// Notifications routes (placeholder for future implementation)
			// notifications := protected.Group("/notifications")
			// {
			// 	notifications.GET("", notificationHandler.List)
			// 	notifications.POST("", notificationHandler.Create)
			// 	notifications.GET("/:id", notificationHandler.Get)
			// 	notifications.PUT("/:id", notificationHandler.Update)
			// 	notifications.DELETE("/:id", notificationHandler.Delete)
			// }

			// Devices routes (placeholder for future implementation)
			// devices := protected.Group("/devices")
			// {
			// 	devices.POST("", deviceHandler.Register)
			// 	devices.DELETE("/:id", deviceHandler.Unregister)
			// }
		}
	}

	return router
}

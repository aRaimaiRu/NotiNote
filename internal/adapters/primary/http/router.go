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
	NoteHandler *handlers.NoteHandler
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

			// Notes routes
			if cfg.NoteHandler != nil {
				notes := protected.Group("/notes")
				{
					// Basic CRUD operations
					notes.GET("", cfg.NoteHandler.ListNotes)
					notes.POST("", cfg.NoteHandler.CreateNote)
					notes.GET("/search", cfg.NoteHandler.SearchNotes)
					notes.GET("/:id", cfg.NoteHandler.GetNote)
					notes.PUT("/:id", cfg.NoteHandler.UpdateNote)
					notes.DELETE("/:id", cfg.NoteHandler.DeleteNote)

					// Note lifecycle operations
					notes.POST("/:id/archive", cfg.NoteHandler.ArchiveNote)
					notes.POST("/:id/unarchive", cfg.NoteHandler.UnarchiveNote)
					notes.POST("/:id/restore", cfg.NoteHandler.RestoreNote)
					notes.POST("/:id/move", cfg.NoteHandler.MoveNote)

					// Hierarchy operations
					notes.GET("/:id/children", cfg.NoteHandler.GetChildren)
					notes.GET("/:id/ancestors", cfg.NoteHandler.GetAncestors)

					// Block operations
					notes.PUT("/:id/blocks", cfg.NoteHandler.ReplaceBlocks)
					notes.POST("/:id/blocks", cfg.NoteHandler.AddBlock)
					notes.PATCH("/:id/blocks/:block_id", cfg.NoteHandler.UpdateBlock)
					notes.DELETE("/:id/blocks/:block_id", cfg.NoteHandler.DeleteBlock)
					notes.POST("/:id/blocks/reorder", cfg.NoteHandler.ReorderBlocks)

					// View and properties
					notes.PUT("/:id/view", cfg.NoteHandler.UpdateViewMetadata)
					notes.PUT("/:id/properties", cfg.NoteHandler.UpdateProperties)

					// Favorite and tags
					notes.PATCH("/:id/favorite", cfg.NoteHandler.ToggleFavorite)
					notes.POST("/:id/tags/:tag_id", cfg.NoteHandler.AddTagToNote)
					notes.DELETE("/:id/tags/:tag_id", cfg.NoteHandler.RemoveTagFromNote)
				}
			}

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

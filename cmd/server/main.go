package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	httpAdapter "github.com/yourusername/notinoteapp/internal/adapters/primary/http"
	"github.com/yourusername/notinoteapp/internal/adapters/primary/http/handlers"
	redisCache "github.com/yourusername/notinoteapp/internal/adapters/secondary/cache/redis"
	"github.com/yourusername/notinoteapp/internal/adapters/secondary/database/postgres"
	"github.com/yourusername/notinoteapp/internal/adapters/secondary/database/postgres/repositories"
	"github.com/yourusername/notinoteapp/internal/adapters/secondary/messaging/fcm"
	"github.com/yourusername/notinoteapp/internal/adapters/secondary/oauth"
	"github.com/yourusername/notinoteapp/internal/application/services"
	"github.com/yourusername/notinoteapp/internal/core/ports"
	coreServices "github.com/yourusername/notinoteapp/internal/core/services"
	"github.com/yourusername/notinoteapp/pkg/config"
	"github.com/yourusername/notinoteapp/pkg/logger"
	"github.com/yourusername/notinoteapp/pkg/utils"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger.Init(cfg.Log.Level, cfg.Log.Format)
	logger.Info("Starting NotiNoteApp server...")

	// Connect to database
	dbConfig := postgres.Config{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		DBName:          cfg.Database.Name,
		SSLMode:         cfg.Database.SSLMode,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
		LogLevel:        cfg.Log.Level,
	}

	db, err := postgres.NewConnection(dbConfig)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := postgres.Close(db); err != nil {
			logger.Errorf("Error closing database: %v", err)
		}
	}()

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db)
	noteRepo := repositories.NewNoteRepository(db)
	deviceRepo := repositories.NewDeviceRepository(db)
	reminderRepo := repositories.NewReminderRepository(db)
	notificationLogRepo := repositories.NewNotificationLogRepository(db)

	// Initialize utilities
	passwordHasher := utils.NewBcryptPasswordHasher()
	tokenService := utils.NewJWTService(cfg.JWT.Secret, "notinoteapp", cfg.JWT.Expiration, cfg.JWT.RefreshExpiration)

	// Connect to Redis for OAuth state management
	redisClient, err := redisCache.NewClient(redisCache.Config{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})
	if err != nil {
		logger.Warnf("Failed to connect to Redis: %v. OAuth may not work properly.", err)
		// Continue without Redis for now, OAuth will fail if used
	}
	defer func() {
		if redisClient != nil {
			if err := redisCache.Close(redisClient); err != nil {
				logger.Errorf("Error closing Redis: %v", err)
			}
		}
	}()

	stateGenerator := utils.NewRedisStateGenerator(redisClient)

	// Initialize services
	authService := services.NewAuthService(
		userRepo,
		passwordHasher,
		tokenService,
		stateGenerator,
	)

	// Import core services package for note service
	noteService := coreServices.NewNoteService(noteRepo)

	// Register OAuth providers
	if cfg.OAuth.Google.ClientID != "" && cfg.OAuth.Google.ClientSecret != "" {
		googleProvider := oauth.NewGoogleProvider(
			cfg.OAuth.Google.ClientID,
			cfg.OAuth.Google.ClientSecret,
			cfg.OAuth.Google.RedirectURL,
			[]string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
		)
		authService.RegisterOAuthProvider(googleProvider)
		logger.Info("Google OAuth provider registered")
	}

	if cfg.OAuth.Facebook.ClientID != "" && cfg.OAuth.Facebook.ClientSecret != "" {
		facebookProvider := oauth.NewFacebookProvider(
			cfg.OAuth.Facebook.ClientID,
			cfg.OAuth.Facebook.ClientSecret,
			cfg.OAuth.Facebook.RedirectURL,
			[]string{"email", "public_profile"},
		)
		authService.RegisterOAuthProvider(facebookProvider)
		logger.Info("Facebook OAuth provider registered")
	}

	// Initialize FCM sender (optional - only if credentials file exists)
	var fcmSender ports.NotificationSender
	var notificationScheduler *services.NotificationScheduler

	if cfg.FCM.CredentialsFile != "" {
		if _, err := os.Stat(cfg.FCM.CredentialsFile); err == nil {
			logrusLogger := logrus.New()
			logrusLogger.SetLevel(logrus.InfoLevel)

			fcmSender, err = fcm.NewFCMSender(cfg.FCM.CredentialsFile, logrusLogger)
			if err != nil {
				logger.Warnf("Failed to initialize FCM sender: %v. Push notifications will not work.", err)
			} else {
				logger.Info("FCM sender initialized successfully")
			}
		} else {
			logger.Warnf("FCM credentials file not found at %s. Push notifications will not work.", cfg.FCM.CredentialsFile)
		}
	}

	// Initialize notification services
	logrusLogger := logrus.New()
	logrusLogger.SetLevel(logrus.InfoLevel)

	deviceService := services.NewDeviceService(deviceRepo, logrusLogger)
	reminderService := services.NewReminderService(reminderRepo, noteRepo, logrusLogger)

	// Initialize notification service and scheduler (only if FCM is available)
	var notificationService *services.NotificationService
	if fcmSender != nil {
		notificationService = services.NewNotificationService(
			deviceRepo,
			notificationLogRepo,
			fcmSender,
			logrusLogger,
		)

		// Initialize and start notification scheduler
		notificationScheduler = services.NewNotificationScheduler(
			reminderRepo,
			notificationService,
			&cfg.Notification,
			logrusLogger,
		)
		notificationScheduler.Start()
		logger.Info("Notification scheduler started")
	} else {
		logger.Warn("Notification service not initialized - FCM sender unavailable")
	}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	noteHandler := handlers.NewNoteHandler(noteService)
	deviceHandler := handlers.NewDeviceHandler(deviceService, logrusLogger)
	reminderHandler := handlers.NewReminderHandler(reminderService, logrusLogger)

	// Setup router
	router := httpAdapter.SetupRouter(httpAdapter.RouterConfig{
		AuthHandler:     authHandler,
		NoteHandler:     noteHandler,
		DeviceHandler:   deviceHandler,
		ReminderHandler: reminderHandler,
		Config:          cfg,
	})

	// Create HTTP server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Infof("Server listening on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Stop notification scheduler first
	if notificationScheduler != nil {
		logger.Info("Stopping notification scheduler...")
		notificationScheduler.Stop()
		logger.Info("Notification scheduler stopped")
	}

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited successfully")
}

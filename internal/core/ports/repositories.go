package ports

import (
	"context"

	"github.com/yourusername/notinoteapp/internal/core/domain"
)

// UserRepository defines the interface for user data persistence
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *domain.User) error

	// FindByID finds a user by ID
	FindByID(ctx context.Context, id int64) (*domain.User, error)

	// FindByEmail finds a user by email
	FindByEmail(ctx context.Context, email string) (*domain.User, error)

	// FindByProvider finds a user by OAuth provider and provider ID
	FindByProvider(ctx context.Context, provider domain.AuthProvider, providerID string) (*domain.User, error)

	// Update updates user information
	Update(ctx context.Context, user *domain.User) error

	// Delete soft deletes a user
	Delete(ctx context.Context, id int64) error

	// List retrieves users with pagination
	List(ctx context.Context, limit, offset int) ([]*domain.User, int64, error)
}

// NoteRepository defines the interface for note data persistence
type NoteRepository interface {
	// Create creates a new note
	Create(ctx context.Context, note interface{}) error

	// FindByID finds a note by ID
	FindByID(ctx context.Context, id int64) (interface{}, error)

	// FindByUserID finds all notes for a user
	FindByUserID(ctx context.Context, userID int64, limit, offset int) ([]interface{}, int64, error)

	// Update updates a note
	Update(ctx context.Context, note interface{}) error

	// Delete deletes a note
	Delete(ctx context.Context, id int64) error

	// Search searches notes by title or content
	Search(ctx context.Context, userID int64, query string, tags []string, limit, offset int) ([]interface{}, int64, error)
}

// NotificationRepository defines the interface for notification data persistence
type NotificationRepository interface {
	// Create creates a new notification
	Create(ctx context.Context, notification interface{}) error

	// FindByID finds a notification by ID
	FindByID(ctx context.Context, id int64) (interface{}, error)

	// FindByNoteID finds all notifications for a note
	FindByNoteID(ctx context.Context, noteID int64) ([]interface{}, error)

	// FindPending finds all pending notifications that are due
	FindPending(ctx context.Context, limit int) ([]interface{}, error)

	// Update updates a notification
	Update(ctx context.Context, notification interface{}) error

	// Delete deletes a notification
	Delete(ctx context.Context, id int64) error

	// UpdateStatus updates notification status
	UpdateStatus(ctx context.Context, id int64, status string) error
}

// DeviceRepository defines the interface for device data persistence
type DeviceRepository interface {
	// Create creates a new device registration
	Create(ctx context.Context, device interface{}) error

	// FindByID finds a device by ID
	FindByID(ctx context.Context, id int64) (interface{}, error)

	// FindByUserID finds all devices for a user
	FindByUserID(ctx context.Context, userID int64) ([]interface{}, error)

	// FindByToken finds a device by token
	FindByToken(ctx context.Context, token string) (interface{}, error)

	// Update updates device information
	Update(ctx context.Context, device interface{}) error

	// Delete deletes a device
	Delete(ctx context.Context, id int64) error

	// DeactivateByToken deactivates a device by token
	DeactivateByToken(ctx context.Context, token string) error
}

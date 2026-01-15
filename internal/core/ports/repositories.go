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

// NoteFilters represents filtering options for notes
type NoteFilters struct {
	ParentID    *int64
	IsArchived  *bool
	ViewType    *domain.ViewType
	Properties  map[string]interface{} // Filter by custom properties
	SearchQuery string                 // Full-text search on title
	Limit       int
	Offset      int
	SortBy      string // "created_at", "updated_at", "title", "position"
	SortOrder   string // "asc", "desc"
}

// NoteRepository defines the interface for note data persistence
type NoteRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, note *domain.Note) error
	FindByID(ctx context.Context, id int64) (*domain.Note, error)
	Update(ctx context.Context, note *domain.Note) error
	Delete(ctx context.Context, id int64) error

	// User notes with filtering
	FindByUserID(ctx context.Context, userID int64, filters NoteFilters) ([]*domain.Note, int64, error)

	// Hierarchy operations
	FindChildren(ctx context.Context, parentID int64) ([]*domain.Note, error)
	FindDescendants(ctx context.Context, parentID int64) ([]*domain.Note, error)
	FindAncestors(ctx context.Context, noteID int64) ([]*domain.Note, error)
	MoveNote(ctx context.Context, noteID int64, newParentID *int64, newPosition int) error

	// Block operations
	UpdateBlocks(ctx context.Context, noteID int64, blocks []domain.Block) error

	// Search and filter
	Search(ctx context.Context, userID int64, query string, filters NoteFilters) ([]*domain.Note, int64, error)

	// Bulk operations
	BulkArchive(ctx context.Context, noteIDs []int64) error
	BulkDelete(ctx context.Context, noteIDs []int64) error

	// Permission check (for ownership)
	CheckOwnership(ctx context.Context, noteID, userID int64) (bool, error)

	// Tag operations
	AddTag(ctx context.Context, noteID int64, tagID string) error
	RemoveTag(ctx context.Context, noteID int64, tagID string) error
	GetNoteTags(ctx context.Context, noteID int64) ([]domain.Tag, error)
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

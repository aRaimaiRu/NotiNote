package ports

import (
	"context"
	"time"

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
	Create(ctx context.Context, device *domain.Device) error

	// FindByID finds a device by ID
	FindByID(ctx context.Context, id int64) (*domain.Device, error)

	// FindByUserID finds all devices for a user
	FindByUserID(ctx context.Context, userID int64) ([]*domain.Device, error)

	// FindActiveByUserID finds all active devices for a user
	FindActiveByUserID(ctx context.Context, userID int64) ([]*domain.Device, error)

	// FindByToken finds a device by token
	FindByToken(ctx context.Context, token string) (*domain.Device, error)

	// FindByUserIDAndToken finds a device by user ID and token
	FindByUserIDAndToken(ctx context.Context, userID int64, token string) (*domain.Device, error)

	// Update updates device information
	Update(ctx context.Context, device *domain.Device) error

	// Delete deletes a device
	Delete(ctx context.Context, id int64) error

	// DeleteByToken deletes a device by user ID and token
	DeleteByToken(ctx context.Context, userID int64, token string) error

	// UpdateLastUsed updates the last used timestamp
	UpdateLastUsed(ctx context.Context, id int64) error

	// DeactivateStaleDevices deactivates devices not used since the given time
	DeactivateStaleDevices(ctx context.Context, before time.Time) (int64, error)
}

// ReminderQueryParams represents filtering options for reminders
type ReminderQueryParams struct {
	IsEnabled *bool
	FromDate  *time.Time
	ToDate    *time.Time
	Limit     int
	Offset    int
}

// ReminderRepository defines the interface for reminder data persistence
type ReminderRepository interface {
	// Create creates a new reminder
	Create(ctx context.Context, reminder *domain.Reminder) error

	// FindByID finds a reminder by ID
	FindByID(ctx context.Context, id int64) (*domain.Reminder, error)

	// FindByNoteID finds all reminders for a note
	FindByNoteID(ctx context.Context, noteID int64) ([]*domain.Reminder, error)

	// FindByUserID finds all reminders for a user with filters
	FindByUserID(ctx context.Context, userID int64, params *ReminderQueryParams) ([]*domain.Reminder, error)

	// FindDueReminders finds all enabled reminders that are due (next_trigger_at <= until)
	FindDueReminders(ctx context.Context, until time.Time, limit int) ([]*domain.Reminder, error)

	// Update updates a reminder
	Update(ctx context.Context, reminder *domain.Reminder) error

	// Delete deletes a reminder
	Delete(ctx context.Context, id int64) error

	// DeleteByNoteID deletes all reminders for a note
	DeleteByNoteID(ctx context.Context, noteID int64) error

	// UpdateNextTrigger updates the next trigger time and last triggered time
	UpdateNextTrigger(ctx context.Context, id int64, nextTrigger time.Time, lastTriggered time.Time) error

	// IncrementTriggerCount increments the trigger count for a reminder
	IncrementTriggerCount(ctx context.Context, id int64) error

	// CheckOwnership checks if a reminder belongs to a user
	CheckOwnership(ctx context.Context, reminderID, userID int64) (bool, error)
}

// NotificationLogRepository defines the interface for notification log data persistence
type NotificationLogRepository interface {
	// Create creates a new notification log entry
	Create(ctx context.Context, log *domain.NotificationLog) error

	// FindByID finds a log entry by ID
	FindByID(ctx context.Context, id int64) (*domain.NotificationLog, error)

	// FindByUserID finds log entries for a user
	FindByUserID(ctx context.Context, userID int64, limit, offset int) ([]*domain.NotificationLog, int64, error)

	// FindByReminderID finds log entries for a reminder
	FindByReminderID(ctx context.Context, reminderID int64) ([]*domain.NotificationLog, error)

	// FindPendingLogs finds all pending notification logs
	FindPendingLogs(ctx context.Context, limit int) ([]*domain.NotificationLog, error)

	// UpdateStatus updates the status of a notification log
	UpdateStatus(ctx context.Context, id int64, status domain.NotificationStatus, errorMessage string) error

	// MarkAsSent marks a log as successfully sent
	MarkAsSent(ctx context.Context, id int64, fcmMessageID string) error

	// DeleteOldLogs deletes logs older than the given time
	DeleteOldLogs(ctx context.Context, before time.Time) (int64, error)
}

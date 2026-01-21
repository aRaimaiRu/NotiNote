package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/yourusername/notinoteapp/internal/adapters/secondary/database/postgres/models"
	"github.com/yourusername/notinoteapp/internal/core/domain"
	"github.com/yourusername/notinoteapp/internal/core/ports"
	"gorm.io/gorm"
)

// ReminderRepository implements the reminder repository interface using PostgreSQL
type ReminderRepository struct {
	db *gorm.DB
}

// NewReminderRepository creates a new reminder repository
func NewReminderRepository(db *gorm.DB) *ReminderRepository {
	return &ReminderRepository{db: db}
}

// Create creates a new reminder
func (r *ReminderRepository) Create(ctx context.Context, reminder *domain.Reminder) error {
	dbReminder := &models.Reminder{}
	dbReminder.FromDomain(reminder)

	if err := r.db.WithContext(ctx).Create(dbReminder).Error; err != nil {
		return err
	}

	// Update domain reminder with generated ID
	reminder.ID = dbReminder.ID
	reminder.CreatedAt = dbReminder.CreatedAt
	reminder.UpdatedAt = dbReminder.UpdatedAt

	return nil
}

// FindByID finds a reminder by ID
func (r *ReminderRepository) FindByID(ctx context.Context, id int64) (*domain.Reminder, error) {
	var dbReminder models.Reminder
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&dbReminder).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrReminderNotFound
		}
		return nil, err
	}

	return dbReminder.ToDomain(), nil
}

// FindByNoteID finds all reminders for a note
func (r *ReminderRepository) FindByNoteID(ctx context.Context, noteID int64) ([]*domain.Reminder, error) {
	var dbReminders []models.Reminder
	if err := r.db.WithContext(ctx).
		Where("note_id = ?", noteID).
		Order("next_trigger_at ASC").
		Find(&dbReminders).Error; err != nil {
		return nil, err
	}

	reminders := make([]*domain.Reminder, len(dbReminders))
	for i, dbReminder := range dbReminders {
		reminders[i] = dbReminder.ToDomain()
	}

	return reminders, nil
}

// FindByUserID finds all reminders for a user with filters
func (r *ReminderRepository) FindByUserID(ctx context.Context, userID int64, params *ports.ReminderQueryParams) ([]*domain.Reminder, error) {
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)

	if params != nil {
		if params.IsEnabled != nil {
			query = query.Where("is_enabled = ?", *params.IsEnabled)
		}
		if params.FromDate != nil {
			query = query.Where("next_trigger_at >= ?", *params.FromDate)
		}
		if params.ToDate != nil {
			query = query.Where("next_trigger_at <= ?", *params.ToDate)
		}
		if params.Limit > 0 {
			query = query.Limit(params.Limit)
		}
		if params.Offset > 0 {
			query = query.Offset(params.Offset)
		}
	}

	var dbReminders []models.Reminder
	if err := query.Order("next_trigger_at ASC").Find(&dbReminders).Error; err != nil {
		return nil, err
	}

	reminders := make([]*domain.Reminder, len(dbReminders))
	for i, dbReminder := range dbReminders {
		reminders[i] = dbReminder.ToDomain()
	}

	return reminders, nil
}

// FindDueReminders finds all enabled reminders that are due (next_trigger_at <= until)
func (r *ReminderRepository) FindDueReminders(ctx context.Context, until time.Time, limit int) ([]*domain.Reminder, error) {
	var dbReminders []models.Reminder
	query := r.db.WithContext(ctx).
		Where("is_enabled = ? AND next_trigger_at <= ?", true, until).
		Order("next_trigger_at ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&dbReminders).Error; err != nil {
		return nil, err
	}

	reminders := make([]*domain.Reminder, len(dbReminders))
	for i, dbReminder := range dbReminders {
		reminders[i] = dbReminder.ToDomain()
	}

	return reminders, nil
}

// Update updates a reminder
func (r *ReminderRepository) Update(ctx context.Context, reminder *domain.Reminder) error {
	dbReminder := &models.Reminder{}
	dbReminder.FromDomain(reminder)

	result := r.db.WithContext(ctx).
		Model(&models.Reminder{}).
		Where("id = ?", reminder.ID).
		Updates(dbReminder)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrReminderNotFound
	}

	return nil
}

// Delete deletes a reminder
func (r *ReminderRepository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&models.Reminder{}, id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrReminderNotFound
	}

	return nil
}

// DeleteByNoteID deletes all reminders for a note
func (r *ReminderRepository) DeleteByNoteID(ctx context.Context, noteID int64) error {
	result := r.db.WithContext(ctx).
		Where("note_id = ?", noteID).
		Delete(&models.Reminder{})

	return result.Error
}

// UpdateNextTrigger updates the next trigger time and last triggered time
func (r *ReminderRepository) UpdateNextTrigger(ctx context.Context, id int64, nextTrigger time.Time, lastTriggered time.Time) error {
	result := r.db.WithContext(ctx).
		Model(&models.Reminder{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"next_trigger_at":   nextTrigger,
			"last_triggered_at": lastTriggered,
			"updated_at":        time.Now(),
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrReminderNotFound
	}

	return nil
}

// IncrementTriggerCount increments the trigger count for a reminder
func (r *ReminderRepository) IncrementTriggerCount(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).
		Model(&models.Reminder{}).
		Where("id = ?", id).
		UpdateColumn("trigger_count", gorm.Expr("trigger_count + 1"))

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrReminderNotFound
	}

	return nil
}

// CheckOwnership checks if a reminder belongs to a user
func (r *ReminderRepository) CheckOwnership(ctx context.Context, reminderID, userID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Reminder{}).
		Where("id = ? AND user_id = ?", reminderID, userID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

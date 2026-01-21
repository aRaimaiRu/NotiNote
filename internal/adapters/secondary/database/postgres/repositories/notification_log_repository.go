package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/yourusername/notinoteapp/internal/adapters/secondary/database/postgres/models"
	"github.com/yourusername/notinoteapp/internal/core/domain"
	"gorm.io/gorm"
)

// NotificationLogRepository implements the notification log repository interface using PostgreSQL
type NotificationLogRepository struct {
	db *gorm.DB
}

// NewNotificationLogRepository creates a new notification log repository
func NewNotificationLogRepository(db *gorm.DB) *NotificationLogRepository {
	return &NotificationLogRepository{db: db}
}

// Create creates a new notification log entry
func (r *NotificationLogRepository) Create(ctx context.Context, log *domain.NotificationLog) error {
	dbLog := &models.NotificationLog{}
	dbLog.FromDomain(log)

	if err := r.db.WithContext(ctx).Create(dbLog).Error; err != nil {
		return err
	}

	// Update domain log with generated ID
	log.ID = dbLog.ID
	log.CreatedAt = dbLog.CreatedAt

	return nil
}

// FindByID finds a log entry by ID
func (r *NotificationLogRepository) FindByID(ctx context.Context, id int64) (*domain.NotificationLog, error) {
	var dbLog models.NotificationLog
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&dbLog).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotificationLogNotFound
		}
		return nil, err
	}

	return dbLog.ToDomain(), nil
}

// FindByUserID finds log entries for a user with pagination
func (r *NotificationLogRepository) FindByUserID(ctx context.Context, userID int64, limit, offset int) ([]*domain.NotificationLog, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).
		Model(&models.NotificationLog{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var dbLogs []models.NotificationLog
	query := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&dbLogs).Error; err != nil {
		return nil, 0, err
	}

	logs := make([]*domain.NotificationLog, len(dbLogs))
	for i, dbLog := range dbLogs {
		logs[i] = dbLog.ToDomain()
	}

	return logs, total, nil
}

// FindByReminderID finds log entries for a reminder
func (r *NotificationLogRepository) FindByReminderID(ctx context.Context, reminderID int64) ([]*domain.NotificationLog, error) {
	var dbLogs []models.NotificationLog
	if err := r.db.WithContext(ctx).
		Where("reminder_id = ?", reminderID).
		Order("created_at DESC").
		Find(&dbLogs).Error; err != nil {
		return nil, err
	}

	logs := make([]*domain.NotificationLog, len(dbLogs))
	for i, dbLog := range dbLogs {
		logs[i] = dbLog.ToDomain()
	}

	return logs, nil
}

// FindPendingLogs finds all pending notification logs for retry
func (r *NotificationLogRepository) FindPendingLogs(ctx context.Context, limit int) ([]*domain.NotificationLog, error) {
	var dbLogs []models.NotificationLog
	query := r.db.WithContext(ctx).
		Where("status = ?", domain.NotificationStatusPending).
		Order("created_at ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&dbLogs).Error; err != nil {
		return nil, err
	}

	logs := make([]*domain.NotificationLog, len(dbLogs))
	for i, dbLog := range dbLogs {
		logs[i] = dbLog.ToDomain()
	}

	return logs, nil
}

// UpdateStatus updates the status of a notification log
func (r *NotificationLogRepository) UpdateStatus(ctx context.Context, id int64, status domain.NotificationStatus, errorMessage string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if errorMessage != "" {
		updates["error_message"] = errorMessage
	}
	if status == domain.NotificationStatusSent {
		now := time.Now()
		updates["sent_at"] = now
	}

	result := r.db.WithContext(ctx).
		Model(&models.NotificationLog{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrNotificationLogNotFound
	}

	return nil
}

// MarkAsSent marks a log as successfully sent
func (r *NotificationLogRepository) MarkAsSent(ctx context.Context, id int64, fcmMessageID string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&models.NotificationLog{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":         domain.NotificationStatusSent,
			"fcm_message_id": fcmMessageID,
			"sent_at":        now,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrNotificationLogNotFound
	}

	return nil
}

// DeleteOldLogs deletes logs older than the given time
func (r *NotificationLogRepository) DeleteOldLogs(ctx context.Context, before time.Time) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("created_at < ?", before).
		Delete(&models.NotificationLog{})

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

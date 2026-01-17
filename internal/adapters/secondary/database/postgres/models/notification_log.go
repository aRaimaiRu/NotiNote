package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/yourusername/notinoteapp/internal/core/domain"
)

// StringMapJSON is a wrapper for map[string]string to handle JSON serialization with GORM
type StringMapJSON map[string]string

// Scan implements the sql.Scanner interface for StringMapJSON
func (s *StringMapJSON) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	var data map[string]string
	if err := json.Unmarshal(bytes, &data); err != nil {
		return err
	}
	*s = data
	return nil
}

// Value implements the driver.Valuer interface for StringMapJSON
func (s StringMapJSON) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

// NotificationLog represents the database model for notification logs
type NotificationLog struct {
	ID           int64                     `gorm:"primaryKey;autoIncrement"`
	ReminderID   *int64                    `gorm:"index:idx_notif_log_reminder"`
	UserID       int64                     `gorm:"not null;index:idx_notif_log_user"`
	DeviceID     *int64                    `gorm:"index:idx_notif_log_device"`
	Title        string                    `gorm:"type:varchar(255);not null"`
	Body         string                    `gorm:"type:text"`
	Data         StringMapJSON             `gorm:"type:jsonb"`
	Status       domain.NotificationStatus `gorm:"type:notification_status;not null;default:'pending';index:idx_notif_log_status,where:status = 'pending'"`
	ErrorMessage string                    `gorm:"type:text"`
	FCMMessageID string                    `gorm:"type:varchar(255)"`
	ScheduledAt  *time.Time                `gorm:"type:timestamptz"`
	SentAt       *time.Time                `gorm:"type:timestamptz"`
	CreatedAt    time.Time                 `gorm:"type:timestamptz;autoCreateTime;index:idx_notif_log_created,sort:desc"`
}

// TableName specifies the table name for GORM
func (NotificationLog) TableName() string {
	return "notification_logs"
}

// ToDomain converts database model to domain entity
func (nl *NotificationLog) ToDomain() *domain.NotificationLog {
	return &domain.NotificationLog{
		ID:           nl.ID,
		ReminderID:   nl.ReminderID,
		UserID:       nl.UserID,
		DeviceID:     nl.DeviceID,
		Title:        nl.Title,
		Body:         nl.Body,
		Data:         nl.Data,
		Status:       nl.Status,
		ErrorMessage: nl.ErrorMessage,
		FCMMessageID: nl.FCMMessageID,
		ScheduledAt:  nl.ScheduledAt,
		SentAt:       nl.SentAt,
		CreatedAt:    nl.CreatedAt,
	}
}

// FromDomain converts domain entity to database model
func (nl *NotificationLog) FromDomain(domainLog *domain.NotificationLog) {
	nl.ID = domainLog.ID
	nl.ReminderID = domainLog.ReminderID
	nl.UserID = domainLog.UserID
	nl.DeviceID = domainLog.DeviceID
	nl.Title = domainLog.Title
	nl.Body = domainLog.Body
	nl.Data = domainLog.Data
	nl.Status = domainLog.Status
	nl.ErrorMessage = domainLog.ErrorMessage
	nl.FCMMessageID = domainLog.FCMMessageID
	nl.ScheduledAt = domainLog.ScheduledAt
	nl.SentAt = domainLog.SentAt
	nl.CreatedAt = domainLog.CreatedAt
}

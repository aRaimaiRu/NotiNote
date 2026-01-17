package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/yourusername/notinoteapp/internal/core/domain"
)

// RepeatConfigJSON is a wrapper for RepeatConfig to handle JSON serialization with GORM
type RepeatConfigJSON struct {
	*domain.RepeatConfig
}

// Scan implements the sql.Scanner interface for RepeatConfigJSON
func (r *RepeatConfigJSON) Scan(value interface{}) error {
	if value == nil {
		r.RepeatConfig = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	var config domain.RepeatConfig
	if err := json.Unmarshal(bytes, &config); err != nil {
		return err
	}
	r.RepeatConfig = &config
	return nil
}

// Value implements the driver.Valuer interface for RepeatConfigJSON
func (r RepeatConfigJSON) Value() (driver.Value, error) {
	if r.RepeatConfig == nil {
		return nil, nil
	}
	return json.Marshal(r.RepeatConfig)
}

// Reminder represents the database model for note reminders
type Reminder struct {
	ID              int64              `gorm:"primaryKey;autoIncrement"`
	NoteID          int64              `gorm:"not null;index:idx_reminder_note"`
	UserID          int64              `gorm:"not null;index:idx_reminder_user"`
	Title           string             `gorm:"type:varchar(255);not null"`
	Message         string             `gorm:"type:text"`
	ScheduledAt     time.Time          `gorm:"type:timestamptz;not null"`
	RepeatType      domain.RepeatType  `gorm:"type:repeat_type;not null;default:'once'"`
	RepeatConfig    RepeatConfigJSON   `gorm:"type:jsonb"`
	RepeatEndAt     *time.Time         `gorm:"type:timestamptz"`
	IsEnabled       bool               `gorm:"not null;default:true"`
	NextTriggerAt   time.Time          `gorm:"type:timestamptz;not null;index:idx_reminder_trigger,where:is_enabled = true"`
	LastTriggeredAt *time.Time         `gorm:"type:timestamptz"`
	TriggerCount    int                `gorm:"not null;default:0"`
	CreatedAt       time.Time          `gorm:"type:timestamptz;autoCreateTime"`
	UpdatedAt       time.Time          `gorm:"type:timestamptz;autoUpdateTime"`
}

// TableName specifies the table name for GORM
func (Reminder) TableName() string {
	return "note_reminders"
}

// ToDomain converts database model to domain entity
func (r *Reminder) ToDomain() *domain.Reminder {
	return &domain.Reminder{
		ID:              r.ID,
		NoteID:          r.NoteID,
		UserID:          r.UserID,
		Title:           r.Title,
		Message:         r.Message,
		ScheduledAt:     r.ScheduledAt,
		RepeatType:      r.RepeatType,
		RepeatConfig:    r.RepeatConfig.RepeatConfig,
		RepeatEndAt:     r.RepeatEndAt,
		IsEnabled:       r.IsEnabled,
		NextTriggerAt:   r.NextTriggerAt,
		LastTriggeredAt: r.LastTriggeredAt,
		TriggerCount:    r.TriggerCount,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
}

// FromDomain converts domain entity to database model
func (r *Reminder) FromDomain(domainReminder *domain.Reminder) {
	r.ID = domainReminder.ID
	r.NoteID = domainReminder.NoteID
	r.UserID = domainReminder.UserID
	r.Title = domainReminder.Title
	r.Message = domainReminder.Message
	r.ScheduledAt = domainReminder.ScheduledAt
	r.RepeatType = domainReminder.RepeatType
	r.RepeatConfig = RepeatConfigJSON{RepeatConfig: domainReminder.RepeatConfig}
	r.RepeatEndAt = domainReminder.RepeatEndAt
	r.IsEnabled = domainReminder.IsEnabled
	r.NextTriggerAt = domainReminder.NextTriggerAt
	r.LastTriggeredAt = domainReminder.LastTriggeredAt
	r.TriggerCount = domainReminder.TriggerCount
	r.CreatedAt = domainReminder.CreatedAt
	r.UpdatedAt = domainReminder.UpdatedAt
}

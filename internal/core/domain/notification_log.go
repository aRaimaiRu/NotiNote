package domain

import (
	"time"
)

// NotificationStatus represents the delivery status of a notification
type NotificationStatus string

const (
	NotificationStatusPending   NotificationStatus = "pending"
	NotificationStatusSent      NotificationStatus = "sent"
	NotificationStatusFailed    NotificationStatus = "failed"
	NotificationStatusCancelled NotificationStatus = "cancelled"
)

// NotificationLog represents a log entry for a sent notification
type NotificationLog struct {
	ID           int64              `json:"id"`
	ReminderID   *int64             `json:"reminder_id,omitempty"` // Can be null if reminder deleted
	UserID       int64              `json:"user_id"`
	DeviceID     *int64             `json:"device_id,omitempty"` // Can be null if device deleted
	Title        string             `json:"title"`
	Body         string             `json:"body,omitempty"`
	Data         map[string]string  `json:"data,omitempty"`
	Status       NotificationStatus `json:"status"`
	ErrorMessage string             `json:"error_message,omitempty"`
	FCMMessageID string             `json:"fcm_message_id,omitempty"`
	ScheduledAt  *time.Time         `json:"scheduled_at,omitempty"`
	SentAt       *time.Time         `json:"sent_at,omitempty"`
	CreatedAt    time.Time          `json:"created_at"`
}

// NewNotificationLog creates a new notification log entry
func NewNotificationLog(userID int64, reminderID *int64, deviceID *int64, title, body string) *NotificationLog {
	now := time.Now()
	return &NotificationLog{
		UserID:      userID,
		ReminderID:  reminderID,
		DeviceID:    deviceID,
		Title:       title,
		Body:        body,
		Status:      NotificationStatusPending,
		ScheduledAt: &now,
		CreatedAt:   now,
	}
}

// MarkAsSent marks the notification as successfully sent
func (nl *NotificationLog) MarkAsSent(fcmMessageID string) {
	nl.Status = NotificationStatusSent
	nl.FCMMessageID = fcmMessageID
	now := time.Now()
	nl.SentAt = &now
}

// MarkAsFailed marks the notification as failed
func (nl *NotificationLog) MarkAsFailed(errorMessage string) {
	nl.Status = NotificationStatusFailed
	nl.ErrorMessage = errorMessage
}

// MarkAsCancelled marks the notification as cancelled
func (nl *NotificationLog) MarkAsCancelled() {
	nl.Status = NotificationStatusCancelled
}

// SetData sets additional data payload for the notification
func (nl *NotificationLog) SetData(data map[string]string) {
	nl.Data = data
}

// IsValidNotificationStatus checks if a status is valid
func IsValidNotificationStatus(status NotificationStatus) bool {
	switch status {
	case NotificationStatusPending, NotificationStatusSent, NotificationStatusFailed, NotificationStatusCancelled:
		return true
	default:
		return false
	}
}

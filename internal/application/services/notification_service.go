package services

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yourusername/notinoteapp/internal/core/domain"
	"github.com/yourusername/notinoteapp/internal/core/ports"
)

// NotificationService handles sending notifications to users
type NotificationService struct {
	deviceRepo ports.DeviceRepository
	logRepo    ports.NotificationLogRepository
	fcmSender  ports.NotificationSender
	logger     *logrus.Logger
}

// NewNotificationService creates a new notification service
func NewNotificationService(
	deviceRepo ports.DeviceRepository,
	logRepo ports.NotificationLogRepository,
	fcmSender ports.NotificationSender,
	logger *logrus.Logger,
) *NotificationService {
	return &NotificationService{
		deviceRepo: deviceRepo,
		logRepo:    logRepo,
		fcmSender:  fcmSender,
		logger:     logger,
	}
}

// NotificationPayload represents the notification content
type NotificationPayload struct {
	Title string
	Body  string
	Data  map[string]string
}

// SendToUser sends a notification to all active devices for a user
func (s *NotificationService) SendToUser(ctx context.Context, userID int64, reminderID *int64, payload *NotificationPayload) error {
	// Get all active devices for the user
	devices, err := s.deviceRepo.FindActiveByUserID(ctx, userID)
	if err != nil {
		s.logger.WithError(err).WithField("user_id", userID).Error("Failed to get user devices")
		return fmt.Errorf("failed to get user devices: %w", err)
	}

	if len(devices) == 0 {
		s.logger.WithField("user_id", userID).Warn("No active devices found for user")
		return nil
	}

	// Send to each device
	var lastErr error
	successCount := 0

	for _, device := range devices {
		// Create notification log
		log := domain.NewNotificationLog(
			userID,
			reminderID,
			&device.ID,
			payload.Title,
			payload.Body,
		)
		log.SetData(payload.Data)

		if err := s.logRepo.Create(ctx, log); err != nil {
			s.logger.WithError(err).Warn("Failed to create notification log")
		}

		// Send notification
		err := s.fcmSender.SendPushNotification(ctx, device.DeviceToken, payload.Title, payload.Body, payload.Data)
		if err != nil {
			lastErr = err
			s.logger.WithError(err).WithFields(logrus.Fields{
				"user_id":   userID,
				"device_id": device.ID,
			}).Error("Failed to send notification to device")

			// Update log with failure
			if log.ID != 0 {
				s.logRepo.UpdateStatus(ctx, log.ID, domain.NotificationStatusFailed, err.Error())
			}
		} else {
			successCount++
			// Update log with success
			if log.ID != 0 {
				s.logRepo.MarkAsSent(ctx, log.ID, "")
			}

			// Update device last used time
			s.deviceRepo.UpdateLastUsed(ctx, device.ID)
		}
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":       userID,
		"device_count":  len(devices),
		"success_count": successCount,
	}).Info("Notification send completed")

	if successCount == 0 && lastErr != nil {
		return fmt.Errorf("failed to send notification to any device: %w", lastErr)
	}

	return nil
}

// SendToDevice sends a notification to a specific device
func (s *NotificationService) SendToDevice(ctx context.Context, device *domain.Device, reminderID *int64, payload *NotificationPayload) error {
	// Create notification log
	log := domain.NewNotificationLog(
		device.UserID,
		reminderID,
		&device.ID,
		payload.Title,
		payload.Body,
	)
	log.SetData(payload.Data)

	if err := s.logRepo.Create(ctx, log); err != nil {
		s.logger.WithError(err).Warn("Failed to create notification log")
	}

	// Send notification
	err := s.fcmSender.SendPushNotification(ctx, device.DeviceToken, payload.Title, payload.Body, payload.Data)
	if err != nil {
		// Update log with failure
		if log.ID != 0 {
			s.logRepo.UpdateStatus(ctx, log.ID, domain.NotificationStatusFailed, err.Error())
		}
		return fmt.Errorf("failed to send notification: %w", err)
	}

	// Update log with success
	if log.ID != 0 {
		s.logRepo.MarkAsSent(ctx, log.ID, "")
	}

	// Update device last used time
	s.deviceRepo.UpdateLastUsed(ctx, device.ID)

	return nil
}

// SendReminderNotification sends a reminder notification
func (s *NotificationService) SendReminderNotification(ctx context.Context, reminder *domain.Reminder) error {
	payload := &NotificationPayload{
		Title: reminder.Title,
		Body:  reminder.Message,
		Data: map[string]string{
			"type":        "reminder",
			"note_id":     fmt.Sprintf("%d", reminder.NoteID),
			"reminder_id": fmt.Sprintf("%d", reminder.ID),
			"click_url":   fmt.Sprintf("/notes?id=%d", reminder.NoteID),
		},
	}

	if payload.Body == "" {
		payload.Body = "You have a reminder for this note"
	}

	return s.SendToUser(ctx, reminder.UserID, &reminder.ID, payload)
}

// GetUserNotificationLogs returns notification logs for a user
func (s *NotificationService) GetUserNotificationLogs(ctx context.Context, userID int64, limit, offset int) ([]*domain.NotificationLog, int64, error) {
	return s.logRepo.FindByUserID(ctx, userID, limit, offset)
}

// CleanupOldLogs removes logs older than the specified duration
func (s *NotificationService) CleanupOldLogs(ctx context.Context, olderThan time.Duration) (int64, error) {
	before := time.Now().Add(-olderThan)
	count, err := s.logRepo.DeleteOldLogs(ctx, before)
	if err != nil {
		s.logger.WithError(err).Error("Failed to cleanup old notification logs")
		return 0, err
	}

	if count > 0 {
		s.logger.WithField("deleted_count", count).Info("Cleaned up old notification logs")
	}

	return count, nil
}

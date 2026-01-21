package services

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yourusername/notinoteapp/internal/core/domain"
	"github.com/yourusername/notinoteapp/internal/core/ports"
)

// ReminderService handles reminder CRUD operations
type ReminderService struct {
	reminderRepo ports.ReminderRepository
	noteRepo     ports.NoteRepository
	logger       *logrus.Logger
}

// NewReminderService creates a new reminder service
func NewReminderService(
	reminderRepo ports.ReminderRepository,
	noteRepo ports.NoteRepository,
	logger *logrus.Logger,
) *ReminderService {
	return &ReminderService{
		reminderRepo: reminderRepo,
		noteRepo:     noteRepo,
		logger:       logger,
	}
}

// CreateReminderRequest represents a request to create a reminder
type CreateReminderRequest struct {
	Title        string               `json:"title" binding:"required"`
	Message      string               `json:"message"`
	ScheduledAt  time.Time            `json:"scheduled_at" binding:"required"`
	RepeatType   domain.RepeatType    `json:"repeat_type"`
	RepeatConfig *domain.RepeatConfig `json:"repeat_config"`
	RepeatEndAt  *time.Time           `json:"repeat_end_at"`
}

// UpdateReminderRequest represents a request to update a reminder
type UpdateReminderRequest struct {
	Title        *string              `json:"title"`
	Message      *string              `json:"message"`
	ScheduledAt  *time.Time           `json:"scheduled_at"`
	RepeatType   *domain.RepeatType   `json:"repeat_type"`
	RepeatConfig *domain.RepeatConfig `json:"repeat_config"`
	RepeatEndAt  *time.Time           `json:"repeat_end_at"`
	IsEnabled    *bool                `json:"is_enabled"`
}

// CreateReminder creates a new reminder for a note
func (s *ReminderService) CreateReminder(ctx context.Context, userID int64, noteID int64, req CreateReminderRequest) (*domain.Reminder, error) {
	// Verify note ownership
	isOwner, err := s.noteRepo.CheckOwnership(ctx, noteID, userID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to check note ownership")
		return nil, err
	}
	if !isOwner {
		return nil, domain.ErrUnauthorizedAccess
	}

	// Create reminder
	reminder, err := domain.NewReminder(noteID, userID, req.Title, req.ScheduledAt)
	if err != nil {
		return nil, err
	}

	if req.Message != "" {
		reminder.UpdateMessage(req.Message)
	}

	// Set repeat configuration if provided
	if req.RepeatType != "" && req.RepeatType != domain.RepeatTypeOnce {
		if err := reminder.SetRepeat(req.RepeatType, req.RepeatConfig, req.RepeatEndAt); err != nil {
			return nil, err
		}
	}

	if err := s.reminderRepo.Create(ctx, reminder); err != nil {
		s.logger.WithError(err).Error("Failed to create reminder")
		return nil, err
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":     userID,
		"note_id":     noteID,
		"reminder_id": reminder.ID,
	}).Info("Reminder created successfully")

	return reminder, nil
}

// GetReminder gets a reminder by ID
func (s *ReminderService) GetReminder(ctx context.Context, userID int64, reminderID int64) (*domain.Reminder, error) {
	reminder, err := s.reminderRepo.FindByID(ctx, reminderID)
	if err != nil {
		return nil, err
	}

	if reminder.UserID != userID {
		return nil, domain.ErrReminderAccessDenied
	}

	return reminder, nil
}

// ListUserReminders returns all reminders for a user
func (s *ReminderService) ListUserReminders(ctx context.Context, userID int64, params *ports.ReminderQueryParams) ([]*domain.Reminder, error) {
	reminders, err := s.reminderRepo.FindByUserID(ctx, userID, params)
	if err != nil {
		s.logger.WithError(err).Error("Failed to list user reminders")
		return nil, err
	}
	return reminders, nil
}

// ListNoteReminders returns all reminders for a note
func (s *ReminderService) ListNoteReminders(ctx context.Context, userID int64, noteID int64) ([]*domain.Reminder, error) {
	// Verify note ownership
	isOwner, err := s.noteRepo.CheckOwnership(ctx, noteID, userID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to check note ownership")
		return nil, err
	}
	if !isOwner {
		return nil, domain.ErrUnauthorizedAccess
	}

	reminders, err := s.reminderRepo.FindByNoteID(ctx, noteID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to list note reminders")
		return nil, err
	}
	return reminders, nil
}

// UpdateReminder updates an existing reminder
func (s *ReminderService) UpdateReminder(ctx context.Context, userID int64, reminderID int64, req UpdateReminderRequest) (*domain.Reminder, error) {
	// Get existing reminder
	reminder, err := s.reminderRepo.FindByID(ctx, reminderID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if reminder.UserID != userID {
		return nil, domain.ErrReminderAccessDenied
	}

	// Apply updates
	if req.Title != nil {
		if err := reminder.UpdateTitle(*req.Title); err != nil {
			return nil, err
		}
	}

	if req.Message != nil {
		reminder.UpdateMessage(*req.Message)
	}

	if req.ScheduledAt != nil {
		if err := reminder.UpdateScheduledAt(*req.ScheduledAt); err != nil {
			return nil, err
		}
	}

	if req.RepeatType != nil {
		if err := reminder.SetRepeat(*req.RepeatType, req.RepeatConfig, req.RepeatEndAt); err != nil {
			return nil, err
		}
	}

	if req.IsEnabled != nil {
		if *req.IsEnabled {
			reminder.Enable()
		} else {
			reminder.Disable()
		}
	}

	if err := s.reminderRepo.Update(ctx, reminder); err != nil {
		s.logger.WithError(err).Error("Failed to update reminder")
		return nil, err
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":     userID,
		"reminder_id": reminderID,
	}).Info("Reminder updated successfully")

	return reminder, nil
}

// DeleteReminder deletes a reminder
func (s *ReminderService) DeleteReminder(ctx context.Context, userID int64, reminderID int64) error {
	// Verify ownership
	isOwner, err := s.reminderRepo.CheckOwnership(ctx, reminderID, userID)
	if err != nil {
		return err
	}
	if !isOwner {
		return domain.ErrReminderAccessDenied
	}

	if err := s.reminderRepo.Delete(ctx, reminderID); err != nil {
		s.logger.WithError(err).Error("Failed to delete reminder")
		return err
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":     userID,
		"reminder_id": reminderID,
	}).Info("Reminder deleted successfully")

	return nil
}

// ToggleReminder toggles the enabled state of a reminder
func (s *ReminderService) ToggleReminder(ctx context.Context, userID int64, reminderID int64) (*domain.Reminder, error) {
	reminder, err := s.reminderRepo.FindByID(ctx, reminderID)
	if err != nil {
		return nil, err
	}

	if reminder.UserID != userID {
		return nil, domain.ErrReminderAccessDenied
	}

	reminder.Toggle()

	if err := s.reminderRepo.Update(ctx, reminder); err != nil {
		s.logger.WithError(err).Error("Failed to toggle reminder")
		return nil, err
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":     userID,
		"reminder_id": reminderID,
		"is_enabled":  reminder.IsEnabled,
	}).Info("Reminder toggled successfully")

	return reminder, nil
}

// SnoozeReminder delays the reminder by the specified duration
func (s *ReminderService) SnoozeReminder(ctx context.Context, userID int64, reminderID int64, duration time.Duration) (*domain.Reminder, error) {
	reminder, err := s.reminderRepo.FindByID(ctx, reminderID)
	if err != nil {
		return nil, err
	}

	if reminder.UserID != userID {
		return nil, domain.ErrReminderAccessDenied
	}

	reminder.Snooze(duration)

	if err := s.reminderRepo.Update(ctx, reminder); err != nil {
		s.logger.WithError(err).Error("Failed to snooze reminder")
		return nil, err
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":         userID,
		"reminder_id":     reminderID,
		"next_trigger_at": reminder.NextTriggerAt,
	}).Info("Reminder snoozed successfully")

	return reminder, nil
}

// FindDueReminders finds reminders that are due for triggering
func (s *ReminderService) FindDueReminders(ctx context.Context, limit int) ([]*domain.Reminder, error) {
	return s.reminderRepo.FindDueReminders(ctx, time.Now(), limit)
}

// MarkReminderTriggered updates a reminder after it has been triggered
func (s *ReminderService) MarkReminderTriggered(ctx context.Context, reminder *domain.Reminder) error {
	reminder.UpdateNextTrigger()

	if err := s.reminderRepo.Update(ctx, reminder); err != nil {
		s.logger.WithError(err).Error("Failed to update reminder after trigger")
		return err
	}

	if err := s.reminderRepo.IncrementTriggerCount(ctx, reminder.ID); err != nil {
		s.logger.WithError(err).Warn("Failed to increment trigger count")
	}

	return nil
}

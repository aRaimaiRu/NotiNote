package services

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yourusername/notinoteapp/internal/core/domain"
	"github.com/yourusername/notinoteapp/internal/core/ports"
	"github.com/yourusername/notinoteapp/pkg/config"
)

// NotificationScheduler handles background scheduling of notifications
type NotificationScheduler struct {
	reminderRepo    ports.ReminderRepository
	notificationSvc *NotificationService
	config          *config.NotificationConfig
	logger          *logrus.Logger
	stopCh          chan struct{}
	wg              sync.WaitGroup
	running         bool
	mu              sync.Mutex
}

// NewNotificationScheduler creates a new notification scheduler
func NewNotificationScheduler(
	reminderRepo ports.ReminderRepository,
	notificationSvc *NotificationService,
	cfg *config.NotificationConfig,
	logger *logrus.Logger,
) *NotificationScheduler {
	return &NotificationScheduler{
		reminderRepo:    reminderRepo,
		notificationSvc: notificationSvc,
		config:          cfg,
		logger:          logger,
		stopCh:          make(chan struct{}),
	}
}

// Start begins the scheduler loop
func (s *NotificationScheduler) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.stopCh = make(chan struct{})
	s.mu.Unlock()

	s.wg.Add(1)
	go s.run()

	s.logger.WithField("interval", s.config.SchedulerInterval).Info("Notification scheduler started")
}

// Stop gracefully stops the scheduler
func (s *NotificationScheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	close(s.stopCh)
	s.wg.Wait()

	s.logger.Info("Notification scheduler stopped")
}

// IsRunning returns whether the scheduler is currently running
func (s *NotificationScheduler) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

func (s *NotificationScheduler) run() {
	defer s.wg.Done()

	// Use configured interval, default to 30 seconds
	interval := s.config.SchedulerInterval
	if interval == 0 {
		interval = 30 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Process immediately on start
	s.processReminders()

	for {
		select {
		case <-s.stopCh:
			s.logger.Info("Scheduler received stop signal")
			return
		case <-ticker.C:
			s.processReminders()
		}
	}
}

func (s *NotificationScheduler) processReminders() {
	ctx := context.Background()

	// Find all reminders that are due
	dueReminders, err := s.reminderRepo.FindDueReminders(ctx, time.Now(), 100)
	if err != nil {
		s.logger.WithError(err).Error("Failed to find due reminders")
		return
	}

	if len(dueReminders) == 0 {
		return
	}

	s.logger.WithField("count", len(dueReminders)).Debug("Found due reminders to process")

	// Process each reminder with worker pool
	workerCount := s.config.WorkerCount
	if workerCount == 0 {
		workerCount = 5
	}

	reminderChan := make(chan *domain.Reminder, len(dueReminders))
	var processWg sync.WaitGroup

	// Start workers
	for i := 0; i < workerCount; i++ {
		processWg.Add(1)
		go func(workerID int) {
			defer processWg.Done()
			for reminder := range reminderChan {
				s.triggerReminder(ctx, reminder)
			}
		}(i)
	}

	// Send reminders to workers
	for _, reminder := range dueReminders {
		reminderChan <- reminder
	}
	close(reminderChan)

	// Wait for all workers to finish
	processWg.Wait()

	s.logger.WithField("processed_count", len(dueReminders)).Info("Finished processing due reminders")
}

func (s *NotificationScheduler) triggerReminder(ctx context.Context, reminder *domain.Reminder) {
	logger := s.logger.WithFields(logrus.Fields{
		"reminder_id": reminder.ID,
		"note_id":     reminder.NoteID,
		"user_id":     reminder.UserID,
	})

	// Send notification
	err := s.notificationSvc.SendReminderNotification(ctx, reminder)
	if err != nil {
		logger.WithError(err).Error("Failed to send reminder notification")
		// Continue to update the reminder state even if notification failed
	} else {
		logger.Info("Reminder notification sent successfully")
	}

	// Update reminder after trigger
	reminder.UpdateNextTrigger()

	// Increment trigger count
	if err := s.reminderRepo.IncrementTriggerCount(ctx, reminder.ID); err != nil {
		logger.WithError(err).Warn("Failed to increment trigger count")
	}

	// Check if reminder should be disabled
	if reminder.RepeatType == domain.RepeatTypeOnce {
		// One-time reminder - disable after trigger
		reminder.Disable()
	} else if reminder.RepeatEndAt != nil && reminder.NextTriggerAt.After(*reminder.RepeatEndAt) {
		// Recurring reminder past end date - disable
		reminder.Disable()
	}

	// Save updated reminder
	if err := s.reminderRepo.Update(ctx, reminder); err != nil {
		logger.WithError(err).Error("Failed to update reminder after trigger")
		return
	}

	logger.WithFields(logrus.Fields{
		"next_trigger_at": reminder.NextTriggerAt,
		"is_enabled":      reminder.IsEnabled,
	}).Debug("Reminder updated after trigger")
}

// ProcessSingleReminder allows manual triggering of a specific reminder (for testing)
func (s *NotificationScheduler) ProcessSingleReminder(ctx context.Context, reminderID int64) error {
	reminder, err := s.reminderRepo.FindByID(ctx, reminderID)
	if err != nil {
		return err
	}

	s.triggerReminder(ctx, reminder)
	return nil
}

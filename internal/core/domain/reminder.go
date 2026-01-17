package domain

import (
	"errors"
	"sort"
	"time"
)

// RepeatType represents the type of repetition for a reminder
type RepeatType string

const (
	RepeatTypeOnce    RepeatType = "once"
	RepeatTypeDaily   RepeatType = "daily"
	RepeatTypeWeekly  RepeatType = "weekly"
	RepeatTypeMonthly RepeatType = "monthly"
)

// RepeatConfig holds the configuration for recurring reminders
type RepeatConfig struct {
	// Days is used for weekly repeat: 0=Sunday, 1=Monday, ..., 6=Saturday
	Days []int `json:"days,omitempty"`
	// Day is used for monthly repeat: 1-31 for specific day, -1 for last day of month
	Day int `json:"day,omitempty"`
}

// Reminder represents a scheduled notification for a note
type Reminder struct {
	ID              int64         `json:"id"`
	NoteID          int64         `json:"note_id"`
	UserID          int64         `json:"user_id"`
	Title           string        `json:"title"`
	Message         string        `json:"message,omitempty"`
	ScheduledAt     time.Time     `json:"scheduled_at"`
	RepeatType      RepeatType    `json:"repeat_type"`
	RepeatConfig    *RepeatConfig `json:"repeat_config,omitempty"`
	RepeatEndAt     *time.Time    `json:"repeat_end_at,omitempty"`
	IsEnabled       bool          `json:"is_enabled"`
	NextTriggerAt   time.Time     `json:"next_trigger_at"`
	LastTriggeredAt *time.Time    `json:"last_triggered_at,omitempty"`
	TriggerCount    int           `json:"trigger_count"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`

	// Relations (loaded optionally)
	Note *Note `json:"note,omitempty"`
}

// Reminder-specific domain errors
var (
	ErrReminderNotFound     = errors.New("reminder not found")
	ErrInvalidRepeatConfig  = errors.New("invalid repeat configuration")
	ErrInvalidRepeatType    = errors.New("invalid repeat type")
	ErrInvalidReminderTitle = errors.New("reminder title is required")
)

// NewReminder creates a new Reminder with validation
func NewReminder(noteID, userID int64, title string, scheduledAt time.Time) (*Reminder, error) {
	if title == "" {
		return nil, ErrInvalidReminderTitle
	}
	if scheduledAt.Before(time.Now()) {
		return nil, ErrInvalidScheduleTime
	}

	now := time.Now()
	return &Reminder{
		NoteID:        noteID,
		UserID:        userID,
		Title:         title,
		ScheduledAt:   scheduledAt,
		RepeatType:    RepeatTypeOnce,
		IsEnabled:     true,
		NextTriggerAt: scheduledAt,
		TriggerCount:  0,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// IsValidRepeatType checks if a repeat type is valid
func IsValidRepeatType(repeatType RepeatType) bool {
	switch repeatType {
	case RepeatTypeOnce, RepeatTypeDaily, RepeatTypeWeekly, RepeatTypeMonthly:
		return true
	default:
		return false
	}
}

// SetRepeat configures the repeat settings for the reminder
func (r *Reminder) SetRepeat(repeatType RepeatType, config *RepeatConfig, endAt *time.Time) error {
	if !IsValidRepeatType(repeatType) {
		return ErrInvalidRepeatType
	}

	// Validate config based on repeat type
	if repeatType == RepeatTypeWeekly {
		if config == nil || len(config.Days) == 0 {
			return ErrInvalidRepeatConfig
		}
		// Validate days are in range 0-6
		for _, day := range config.Days {
			if day < 0 || day > 6 {
				return ErrInvalidRepeatConfig
			}
		}
	}

	if repeatType == RepeatTypeMonthly {
		if config == nil {
			return ErrInvalidRepeatConfig
		}
		// Validate day is in valid range: 1-31 or -1 for last day
		if config.Day != -1 && (config.Day < 1 || config.Day > 31) {
			return ErrInvalidRepeatConfig
		}
	}

	r.RepeatType = repeatType
	r.RepeatConfig = config
	r.RepeatEndAt = endAt
	r.UpdatedAt = time.Now()

	return nil
}

// CalculateNextTrigger calculates the next trigger time based on repeat configuration
// The 'from' parameter should be the last trigger time or current time
func (r *Reminder) CalculateNextTrigger(from time.Time) time.Time {
	switch r.RepeatType {
	case RepeatTypeOnce:
		// One-time reminders don't have a next trigger after firing
		return r.ScheduledAt

	case RepeatTypeDaily:
		return r.calculateNextDaily(from)

	case RepeatTypeWeekly:
		return r.calculateNextWeekly(from)

	case RepeatTypeMonthly:
		return r.calculateNextMonthly(from)

	default:
		return r.ScheduledAt
	}
}

// calculateNextDaily calculates the next daily trigger
func (r *Reminder) calculateNextDaily(from time.Time) time.Time {
	// Get the time of day from the scheduled time
	hour, min, sec := r.ScheduledAt.Clock()

	// Start from the next day
	next := from.AddDate(0, 0, 1)
	next = time.Date(next.Year(), next.Month(), next.Day(), hour, min, sec, 0, r.ScheduledAt.Location())

	// If the calculated time is still in the past, move to next day
	for !next.After(from) {
		next = next.AddDate(0, 0, 1)
	}

	return next
}

// calculateNextWeekly calculates the next weekly trigger based on configured days
func (r *Reminder) calculateNextWeekly(from time.Time) time.Time {
	if r.RepeatConfig == nil || len(r.RepeatConfig.Days) == 0 {
		return r.ScheduledAt
	}

	// Get the time of day from the scheduled time
	hour, min, sec := r.ScheduledAt.Clock()

	// Sort days for consistent iteration
	days := make([]int, len(r.RepeatConfig.Days))
	copy(days, r.RepeatConfig.Days)
	sort.Ints(days)

	// Start checking from the next day
	check := from.AddDate(0, 0, 1)

	// Check up to 8 days (covers all possible cases)
	for i := 0; i < 8; i++ {
		checkDay := int(check.Weekday())

		for _, targetDay := range days {
			if checkDay == targetDay {
				next := time.Date(check.Year(), check.Month(), check.Day(), hour, min, sec, 0, r.ScheduledAt.Location())
				if next.After(from) {
					return next
				}
			}
		}
		check = check.AddDate(0, 0, 1)
	}

	// Fallback: return one week from scheduled time
	return r.ScheduledAt.AddDate(0, 0, 7)
}

// calculateNextMonthly calculates the next monthly trigger
func (r *Reminder) calculateNextMonthly(from time.Time) time.Time {
	if r.RepeatConfig == nil {
		return r.ScheduledAt
	}

	// Get the time of day from the scheduled time
	hour, min, sec := r.ScheduledAt.Clock()

	// Start from the next month
	nextMonth := from.AddDate(0, 1, 0)
	year, month := nextMonth.Year(), nextMonth.Month()

	var targetDay int
	if r.RepeatConfig.Day == -1 {
		// Last day of month
		targetDay = lastDayOfMonth(year, month)
	} else {
		targetDay = r.RepeatConfig.Day
		// Adjust if day doesn't exist in target month
		lastDay := lastDayOfMonth(year, month)
		if targetDay > lastDay {
			targetDay = lastDay
		}
	}

	next := time.Date(year, month, targetDay, hour, min, sec, 0, r.ScheduledAt.Location())

	// If the calculated time is still in the past, move to next month
	if !next.After(from) {
		nextMonth = nextMonth.AddDate(0, 1, 0)
		year, month = nextMonth.Year(), nextMonth.Month()

		if r.RepeatConfig.Day == -1 {
			targetDay = lastDayOfMonth(year, month)
		} else {
			targetDay = r.RepeatConfig.Day
			lastDay := lastDayOfMonth(year, month)
			if targetDay > lastDay {
				targetDay = lastDay
			}
		}
		next = time.Date(year, month, targetDay, hour, min, sec, 0, r.ScheduledAt.Location())
	}

	return next
}

// lastDayOfMonth returns the last day of the given month
func lastDayOfMonth(year int, month time.Month) int {
	// Go to the first day of next month, then subtract one day
	firstOfNext := time.Date(year, month+1, 1, 0, 0, 0, 0, time.UTC)
	lastOfMonth := firstOfNext.AddDate(0, 0, -1)
	return lastOfMonth.Day()
}

// UpdateNextTrigger updates the next trigger time after a successful trigger
func (r *Reminder) UpdateNextTrigger() {
	now := time.Now()
	r.LastTriggeredAt = &now
	r.TriggerCount++
	r.UpdatedAt = now

	if r.RepeatType == RepeatTypeOnce {
		// Disable one-time reminders after trigger
		r.IsEnabled = false
	} else {
		// Calculate next trigger for recurring reminders
		r.NextTriggerAt = r.CalculateNextTrigger(now)

		// Check if we've reached the end date
		if r.RepeatEndAt != nil && r.NextTriggerAt.After(*r.RepeatEndAt) {
			r.IsEnabled = false
		}
	}
}

// Enable enables the reminder
func (r *Reminder) Enable() {
	r.IsEnabled = true
	r.UpdatedAt = time.Now()
}

// Disable disables the reminder
func (r *Reminder) Disable() {
	r.IsEnabled = false
	r.UpdatedAt = time.Now()
}

// Toggle toggles the enabled state
func (r *Reminder) Toggle() {
	r.IsEnabled = !r.IsEnabled
	r.UpdatedAt = time.Now()
}

// Snooze delays the next trigger by the specified duration
func (r *Reminder) Snooze(duration time.Duration) {
	r.NextTriggerAt = time.Now().Add(duration)
	r.UpdatedAt = time.Now()
}

// UpdateTitle updates the reminder title
func (r *Reminder) UpdateTitle(title string) error {
	if title == "" {
		return ErrInvalidReminderTitle
	}
	r.Title = title
	r.UpdatedAt = time.Now()
	return nil
}

// UpdateMessage updates the reminder message
func (r *Reminder) UpdateMessage(message string) {
	r.Message = message
	r.UpdatedAt = time.Now()
}

// UpdateScheduledAt updates the scheduled time and recalculates next trigger
func (r *Reminder) UpdateScheduledAt(scheduledAt time.Time) error {
	if scheduledAt.Before(time.Now()) {
		return ErrInvalidScheduleTime
	}
	r.ScheduledAt = scheduledAt
	r.NextTriggerAt = scheduledAt
	r.UpdatedAt = time.Now()
	return nil
}

// IsDue returns true if the reminder is due for triggering
func (r *Reminder) IsDue() bool {
	return r.IsEnabled && time.Now().After(r.NextTriggerAt)
}

// IsExpired returns true if the reminder has reached its end date
func (r *Reminder) IsExpired() bool {
	if r.RepeatEndAt == nil {
		return false
	}
	return time.Now().After(*r.RepeatEndAt)
}

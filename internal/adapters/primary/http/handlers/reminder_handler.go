package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yourusername/notinoteapp/internal/application/services"
	"github.com/yourusername/notinoteapp/internal/core/domain"
	"github.com/yourusername/notinoteapp/internal/core/ports"
)

// ReminderHandler handles reminder-related HTTP requests
type ReminderHandler struct {
	reminderService *services.ReminderService
	logger          *logrus.Logger
}

// NewReminderHandler creates a new reminder handler
func NewReminderHandler(reminderService *services.ReminderService, logger *logrus.Logger) *ReminderHandler {
	return &ReminderHandler{
		reminderService: reminderService,
		logger:          logger,
	}
}

// CreateReminderRequest represents a reminder creation request
type CreateReminderRequest struct {
	Title        string               `json:"title" binding:"required,min=1,max=255"`
	Message      string               `json:"message"`
	ScheduledAt  time.Time            `json:"scheduled_at" binding:"required"`
	RepeatType   domain.RepeatType    `json:"repeat_type"`
	RepeatConfig *domain.RepeatConfig `json:"repeat_config"`
	RepeatEndAt  *time.Time           `json:"repeat_end_at"`
}

// UpdateReminderRequest represents a reminder update request
type UpdateReminderRequest struct {
	Title        *string              `json:"title"`
	Message      *string              `json:"message"`
	ScheduledAt  *time.Time           `json:"scheduled_at"`
	RepeatType   *domain.RepeatType   `json:"repeat_type"`
	RepeatConfig *domain.RepeatConfig `json:"repeat_config"`
	RepeatEndAt  *time.Time           `json:"repeat_end_at"`
	IsEnabled    *bool                `json:"is_enabled"`
}

// SnoozeRequest represents a snooze request
type SnoozeRequest struct {
	Duration string `json:"duration" binding:"required"` // e.g., "10m", "1h", "1d"
}

// Create creates a new reminder for a note
// POST /api/v1/notes/:id/reminders
func (h *ReminderHandler) Create(c *gin.Context) {
	userID := c.GetInt64("user_id")

	noteID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid note ID",
		})
		return
	}

	var req CreateReminderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request: " + err.Error(),
		})
		return
	}

	serviceReq := services.CreateReminderRequest{
		Title:        req.Title,
		Message:      req.Message,
		ScheduledAt:  req.ScheduledAt,
		RepeatType:   req.RepeatType,
		RepeatConfig: req.RepeatConfig,
		RepeatEndAt:  req.RepeatEndAt,
	}

	reminder, err := h.reminderService.CreateReminder(c.Request.Context(), userID, noteID, serviceReq)
	if err != nil {
		if err == domain.ErrUnauthorizedAccess {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Access denied to this note",
			})
			return
		}
		if err == domain.ErrInvalidScheduleTime {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Schedule time must be in the future",
			})
			return
		}
		h.logger.WithError(err).Error("Failed to create reminder")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create reminder",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    reminder,
	})
}

// ListByNote returns all reminders for a specific note
// GET /api/v1/notes/:id/reminders
func (h *ReminderHandler) ListByNote(c *gin.Context) {
	userID := c.GetInt64("user_id")

	noteID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid note ID",
		})
		return
	}

	reminders, err := h.reminderService.ListNoteReminders(c.Request.Context(), userID, noteID)
	if err != nil {
		if err == domain.ErrUnauthorizedAccess {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Access denied to this note",
			})
			return
		}
		h.logger.WithError(err).Error("Failed to list note reminders")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to list reminders",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"reminders": reminders,
		},
	})
}

// List returns all reminders for the current user
// GET /api/v1/reminders
func (h *ReminderHandler) List(c *gin.Context) {
	userID := c.GetInt64("user_id")

	// Parse query parameters
	var params *ports.ReminderQueryParams
	if c.Query("enabled") != "" || c.Query("from") != "" || c.Query("to") != "" {
		params = &ports.ReminderQueryParams{}

		if enabledStr := c.Query("enabled"); enabledStr != "" {
			enabled := enabledStr == "true"
			params.IsEnabled = &enabled
		}

		if fromStr := c.Query("from"); fromStr != "" {
			if fromDate, err := time.Parse(time.RFC3339, fromStr); err == nil {
				params.FromDate = &fromDate
			}
		}

		if toStr := c.Query("to"); toStr != "" {
			if toDate, err := time.Parse(time.RFC3339, toStr); err == nil {
				params.ToDate = &toDate
			}
		}

		if limitStr := c.Query("limit"); limitStr != "" {
			if limit, err := strconv.Atoi(limitStr); err == nil {
				params.Limit = limit
			}
		}

		if offsetStr := c.Query("offset"); offsetStr != "" {
			if offset, err := strconv.Atoi(offsetStr); err == nil {
				params.Offset = offset
			}
		}
	}

	reminders, err := h.reminderService.ListUserReminders(c.Request.Context(), userID, params)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list user reminders")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to list reminders",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"reminders": reminders,
		},
	})
}

// Get returns a specific reminder
// GET /api/v1/reminders/:id
func (h *ReminderHandler) Get(c *gin.Context) {
	userID := c.GetInt64("user_id")

	reminderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid reminder ID",
		})
		return
	}

	reminder, err := h.reminderService.GetReminder(c.Request.Context(), userID, reminderID)
	if err != nil {
		if err == domain.ErrReminderNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Reminder not found",
			})
			return
		}
		if err == domain.ErrReminderAccessDenied {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Access denied to this reminder",
			})
			return
		}
		h.logger.WithError(err).Error("Failed to get reminder")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get reminder",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    reminder,
	})
}

// Update updates an existing reminder
// PUT /api/v1/reminders/:id
func (h *ReminderHandler) Update(c *gin.Context) {
	userID := c.GetInt64("user_id")

	reminderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid reminder ID",
		})
		return
	}

	var req UpdateReminderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request: " + err.Error(),
		})
		return
	}

	serviceReq := services.UpdateReminderRequest{
		Title:        req.Title,
		Message:      req.Message,
		ScheduledAt:  req.ScheduledAt,
		RepeatType:   req.RepeatType,
		RepeatConfig: req.RepeatConfig,
		RepeatEndAt:  req.RepeatEndAt,
		IsEnabled:    req.IsEnabled,
	}

	reminder, err := h.reminderService.UpdateReminder(c.Request.Context(), userID, reminderID, serviceReq)
	if err != nil {
		if err == domain.ErrReminderNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Reminder not found",
			})
			return
		}
		if err == domain.ErrReminderAccessDenied {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Access denied to this reminder",
			})
			return
		}
		if err == domain.ErrInvalidScheduleTime {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Schedule time must be in the future",
			})
			return
		}
		h.logger.WithError(err).Error("Failed to update reminder")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update reminder",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    reminder,
	})
}

// Delete removes a reminder
// DELETE /api/v1/reminders/:id
func (h *ReminderHandler) Delete(c *gin.Context) {
	userID := c.GetInt64("user_id")

	reminderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid reminder ID",
		})
		return
	}

	err = h.reminderService.DeleteReminder(c.Request.Context(), userID, reminderID)
	if err != nil {
		if err == domain.ErrReminderNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Reminder not found",
			})
			return
		}
		if err == domain.ErrReminderAccessDenied {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Access denied to this reminder",
			})
			return
		}
		h.logger.WithError(err).Error("Failed to delete reminder")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete reminder",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Reminder deleted successfully",
	})
}

// Toggle enables or disables a reminder
// PATCH /api/v1/reminders/:id/toggle
func (h *ReminderHandler) Toggle(c *gin.Context) {
	userID := c.GetInt64("user_id")

	reminderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid reminder ID",
		})
		return
	}

	reminder, err := h.reminderService.ToggleReminder(c.Request.Context(), userID, reminderID)
	if err != nil {
		if err == domain.ErrReminderNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Reminder not found",
			})
			return
		}
		if err == domain.ErrReminderAccessDenied {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Access denied to this reminder",
			})
			return
		}
		h.logger.WithError(err).Error("Failed to toggle reminder")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to toggle reminder",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    reminder,
	})
}

// Snooze delays a reminder by the specified duration
// POST /api/v1/reminders/:id/snooze
func (h *ReminderHandler) Snooze(c *gin.Context) {
	userID := c.GetInt64("user_id")

	reminderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid reminder ID",
		})
		return
	}

	var req SnoozeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request: " + err.Error(),
		})
		return
	}

	// Parse duration string (e.g., "10m", "1h", "1d")
	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		// Try parsing as days
		if len(req.Duration) > 1 && req.Duration[len(req.Duration)-1] == 'd' {
			days, parseErr := strconv.Atoi(req.Duration[:len(req.Duration)-1])
			if parseErr == nil {
				duration = time.Duration(days) * 24 * time.Hour
			} else {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   "Invalid duration format. Use formats like '10m', '1h', '1d'",
				})
				return
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid duration format. Use formats like '10m', '1h', '1d'",
			})
			return
		}
	}

	reminder, err := h.reminderService.SnoozeReminder(c.Request.Context(), userID, reminderID, duration)
	if err != nil {
		if err == domain.ErrReminderNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Reminder not found",
			})
			return
		}
		if err == domain.ErrReminderAccessDenied {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Access denied to this reminder",
			})
			return
		}
		h.logger.WithError(err).Error("Failed to snooze reminder")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to snooze reminder",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    reminder,
	})
}

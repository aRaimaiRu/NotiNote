package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yourusername/notinoteapp/internal/application/services"
	"github.com/yourusername/notinoteapp/internal/core/domain"
)

// DeviceHandler handles device-related HTTP requests
type DeviceHandler struct {
	deviceService *services.DeviceService
	logger        *logrus.Logger
}

// NewDeviceHandler creates a new device handler
func NewDeviceHandler(deviceService *services.DeviceService, logger *logrus.Logger) *DeviceHandler {
	return &DeviceHandler{
		deviceService: deviceService,
		logger:        logger,
	}
}

// RegisterDeviceRequest represents a device registration request
type RegisterDeviceRequest struct {
	DeviceToken string            `json:"device_token" binding:"required"`
	DeviceType  domain.DeviceType `json:"device_type" binding:"required,oneof=web android ios"`
	DeviceName  string            `json:"device_name"`
	BrowserInfo string            `json:"browser_info"`
}

// UnregisterByTokenRequest represents a request to unregister by token
type UnregisterByTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

// Register registers a new device for push notifications
// POST /api/v1/devices
func (h *DeviceHandler) Register(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req RegisterDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request: " + err.Error(),
		})
		return
	}

	serviceReq := services.RegisterDeviceRequest{
		DeviceToken: req.DeviceToken,
		DeviceType:  req.DeviceType,
		DeviceName:  req.DeviceName,
		BrowserInfo: req.BrowserInfo,
	}

	device, err := h.deviceService.RegisterDevice(c.Request.Context(), userID, serviceReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to register device")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to register device",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    device,
	})
}

// List returns all devices for the current user
// GET /api/v1/devices
func (h *DeviceHandler) List(c *gin.Context) {
	userID := c.GetInt64("user_id")

	devices, err := h.deviceService.ListUserDevices(c.Request.Context(), userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list devices")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to list devices",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"devices": devices,
		},
	})
}

// Unregister removes a device by ID
// DELETE /api/v1/devices/:id
func (h *DeviceHandler) Unregister(c *gin.Context) {
	userID := c.GetInt64("user_id")

	deviceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid device ID",
		})
		return
	}

	err = h.deviceService.UnregisterDevice(c.Request.Context(), userID, deviceID)
	if err != nil {
		if err == domain.ErrDeviceNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Device not found",
			})
			return
		}
		h.logger.WithError(err).Error("Failed to unregister device")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to unregister device",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Device unregistered successfully",
	})
}

// UnregisterByToken removes a device by token
// DELETE /api/v1/devices/token
func (h *DeviceHandler) UnregisterByToken(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req UnregisterByTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request: " + err.Error(),
		})
		return
	}

	err := h.deviceService.UnregisterByToken(c.Request.Context(), userID, req.Token)
	if err != nil {
		if err == domain.ErrDeviceNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Device not found",
			})
			return
		}
		h.logger.WithError(err).Error("Failed to unregister device by token")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to unregister device",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Device unregistered successfully",
	})
}

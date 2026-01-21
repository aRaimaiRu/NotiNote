package services

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yourusername/notinoteapp/internal/core/domain"
	"github.com/yourusername/notinoteapp/internal/core/ports"
)

// DeviceService handles device registration and management
type DeviceService struct {
	deviceRepo ports.DeviceRepository
	logger     *logrus.Logger
}

// NewDeviceService creates a new device service
func NewDeviceService(
	deviceRepo ports.DeviceRepository,
	logger *logrus.Logger,
) *DeviceService {
	return &DeviceService{
		deviceRepo: deviceRepo,
		logger:     logger,
	}
}

// RegisterDeviceRequest represents a request to register a device
type RegisterDeviceRequest struct {
	DeviceToken string            `json:"device_token" binding:"required"`
	DeviceType  domain.DeviceType `json:"device_type" binding:"required"`
	DeviceName  string            `json:"device_name"`
	BrowserInfo string            `json:"browser_info"`
}

// RegisterDevice registers a new device for push notifications
func (s *DeviceService) RegisterDevice(ctx context.Context, userID int64, req RegisterDeviceRequest) (*domain.Device, error) {
	// Check if device already exists for this user
	existingDevice, err := s.deviceRepo.FindByUserIDAndToken(ctx, userID, req.DeviceToken)
	if err == nil && existingDevice != nil {
		// Device exists, update last used time and reactivate if needed
		existingDevice.Activate()
		existingDevice.UpdateLastUsed()
		if req.DeviceName != "" {
			existingDevice.SetDeviceName(req.DeviceName)
		}
		if req.BrowserInfo != "" {
			existingDevice.SetBrowserInfo(req.BrowserInfo)
		}
		
		if err := s.deviceRepo.Update(ctx, existingDevice); err != nil {
			s.logger.WithError(err).Error("Failed to update existing device")
			return nil, err
		}
		return existingDevice, nil
	}

	// Create new device
	device, err := domain.NewDevice(userID, req.DeviceToken, req.DeviceType)
	if err != nil {
		return nil, err
	}

	if req.DeviceName != "" {
		device.SetDeviceName(req.DeviceName)
	}
	if req.BrowserInfo != "" {
		device.SetBrowserInfo(req.BrowserInfo)
	}

	if err := s.deviceRepo.Create(ctx, device); err != nil {
		s.logger.WithError(err).Error("Failed to create device")
		return nil, err
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":     userID,
		"device_type": req.DeviceType,
	}).Info("Device registered successfully")

	return device, nil
}

// ListUserDevices returns all devices for a user
func (s *DeviceService) ListUserDevices(ctx context.Context, userID int64) ([]*domain.Device, error) {
	devices, err := s.deviceRepo.FindByUserID(ctx, userID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to list user devices")
		return nil, err
	}
	return devices, nil
}

// UnregisterDevice removes a device registration
func (s *DeviceService) UnregisterDevice(ctx context.Context, userID int64, deviceID int64) error {
	// Verify ownership
	device, err := s.deviceRepo.FindByID(ctx, deviceID)
	if err != nil {
		return err
	}
	if device.UserID != userID {
		return domain.ErrUnauthorizedAccess
	}

	if err := s.deviceRepo.Delete(ctx, deviceID); err != nil {
		s.logger.WithError(err).Error("Failed to delete device")
		return err
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":   userID,
		"device_id": deviceID,
	}).Info("Device unregistered successfully")

	return nil
}

// UnregisterByToken removes a device by token
func (s *DeviceService) UnregisterByToken(ctx context.Context, userID int64, token string) error {
	if err := s.deviceRepo.DeleteByToken(ctx, userID, token); err != nil {
		s.logger.WithError(err).Error("Failed to delete device by token")
		return err
	}

	s.logger.WithField("user_id", userID).Info("Device unregistered by token successfully")
	return nil
}

// GetActiveDevices returns all active devices for a user
func (s *DeviceService) GetActiveDevices(ctx context.Context, userID int64) ([]*domain.Device, error) {
	devices, err := s.deviceRepo.FindActiveByUserID(ctx, userID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get active devices")
		return nil, err
	}
	return devices, nil
}

// DeactivateStaleDevices deactivates devices not used in the given duration
func (s *DeviceService) DeactivateStaleDevices(ctx context.Context, staleDuration time.Duration) (int64, error) {
	before := time.Now().Add(-staleDuration)
	count, err := s.deviceRepo.DeactivateStaleDevices(ctx, before)
	if err != nil {
		s.logger.WithError(err).Error("Failed to deactivate stale devices")
		return 0, err
	}

	if count > 0 {
		s.logger.WithField("count", count).Info("Deactivated stale devices")
	}

	return count, nil
}

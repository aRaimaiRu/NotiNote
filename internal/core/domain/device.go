package domain

import (
	"errors"
	"time"
)

// DeviceType represents the type/platform of a device
type DeviceType string

const (
	DeviceTypeWeb     DeviceType = "web"
	DeviceTypeAndroid DeviceType = "android"
	DeviceTypeIOS     DeviceType = "ios"
)

// Device represents a user's device registered for push notifications
type Device struct {
	ID          int64      `json:"id"`
	UserID      int64      `json:"user_id"`
	DeviceToken string     `json:"device_token"`
	DeviceType  DeviceType `json:"device_type"`
	DeviceName  string     `json:"device_name,omitempty"`
	BrowserInfo string     `json:"browser_info,omitempty"`
	IsActive    bool       `json:"is_active"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Device-specific domain errors
var (
	ErrDeviceAlreadyExists = errors.New("device already registered for this user")
	ErrInvalidDeviceType   = errors.New("invalid device type")
)

// NewDevice creates a new Device with validation
func NewDevice(userID int64, deviceToken string, deviceType DeviceType) (*Device, error) {
	if deviceToken == "" {
		return nil, ErrInvalidDeviceToken
	}
	if !IsValidDeviceType(deviceType) {
		return nil, ErrInvalidDeviceType
	}

	now := time.Now()
	return &Device{
		UserID:      userID,
		DeviceToken: deviceToken,
		DeviceType:  deviceType,
		IsActive:    true,
		LastUsedAt:  &now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// IsValidDeviceType checks if a device type is valid
func IsValidDeviceType(deviceType DeviceType) bool {
	switch deviceType {
	case DeviceTypeWeb, DeviceTypeAndroid, DeviceTypeIOS:
		return true
	default:
		return false
	}
}

// SetDeviceName sets the device name
func (d *Device) SetDeviceName(name string) {
	d.DeviceName = name
	d.UpdatedAt = time.Now()
}

// SetBrowserInfo sets the browser information (for web devices)
func (d *Device) SetBrowserInfo(info string) {
	d.BrowserInfo = info
	d.UpdatedAt = time.Now()
}

// Activate activates the device
func (d *Device) Activate() {
	d.IsActive = true
	d.UpdatedAt = time.Now()
}

// Deactivate deactivates the device
func (d *Device) Deactivate() {
	d.IsActive = false
	d.UpdatedAt = time.Now()
}

// UpdateLastUsed updates the last used timestamp
func (d *Device) UpdateLastUsed() {
	now := time.Now()
	d.LastUsedAt = &now
	d.UpdatedAt = now
}

// UpdateToken updates the device token (e.g., when FCM token is refreshed)
func (d *Device) UpdateToken(newToken string) error {
	if newToken == "" {
		return ErrInvalidDeviceToken
	}
	d.DeviceToken = newToken
	d.UpdatedAt = time.Now()
	return nil
}

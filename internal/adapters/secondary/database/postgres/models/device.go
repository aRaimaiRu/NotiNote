package models

import (
	"time"

	"github.com/yourusername/notinoteapp/internal/core/domain"
)

// Device represents the database model for user devices
type Device struct {
	ID          int64             `gorm:"primaryKey;autoIncrement"`
	UserID      int64             `gorm:"not null;index:idx_device_user_active,where:is_active = true"`
	DeviceToken string            `gorm:"type:text;not null;index:idx_device_token"`
	DeviceType  domain.DeviceType `gorm:"type:device_type;not null"`
	DeviceName  string            `gorm:"size:255"`
	BrowserInfo string            `gorm:"size:255"`
	IsActive    bool              `gorm:"not null;default:true"`
	LastUsedAt  *time.Time        `gorm:"type:timestamptz"`
	CreatedAt   time.Time         `gorm:"type:timestamptz;autoCreateTime"`
	UpdatedAt   time.Time         `gorm:"type:timestamptz;autoUpdateTime"`
}

// TableName specifies the table name for GORM
func (Device) TableName() string {
	return "user_devices"
}

// ToDomain converts database model to domain entity
func (d *Device) ToDomain() *domain.Device {
	return &domain.Device{
		ID:          d.ID,
		UserID:      d.UserID,
		DeviceToken: d.DeviceToken,
		DeviceType:  d.DeviceType,
		DeviceName:  d.DeviceName,
		BrowserInfo: d.BrowserInfo,
		IsActive:    d.IsActive,
		LastUsedAt:  d.LastUsedAt,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

// FromDomain converts domain entity to database model
func (d *Device) FromDomain(domainDevice *domain.Device) {
	d.ID = domainDevice.ID
	d.UserID = domainDevice.UserID
	d.DeviceToken = domainDevice.DeviceToken
	d.DeviceType = domainDevice.DeviceType
	d.DeviceName = domainDevice.DeviceName
	d.BrowserInfo = domainDevice.BrowserInfo
	d.IsActive = domainDevice.IsActive
	d.LastUsedAt = domainDevice.LastUsedAt
	d.CreatedAt = domainDevice.CreatedAt
	d.UpdatedAt = domainDevice.UpdatedAt
}

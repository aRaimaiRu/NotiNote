package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/yourusername/notinoteapp/internal/adapters/secondary/database/postgres/models"
	"github.com/yourusername/notinoteapp/internal/core/domain"
	"gorm.io/gorm"
)

// DeviceRepository implements the device repository interface using PostgreSQL
type DeviceRepository struct {
	db *gorm.DB
}

// NewDeviceRepository creates a new device repository
func NewDeviceRepository(db *gorm.DB) *DeviceRepository {
	return &DeviceRepository{db: db}
}

// Create creates a new device registration
func (r *DeviceRepository) Create(ctx context.Context, device *domain.Device) error {
	dbDevice := &models.Device{}
	dbDevice.FromDomain(device)

	if err := r.db.WithContext(ctx).Create(dbDevice).Error; err != nil {
		// Check for unique constraint violation
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.ErrDeviceAlreadyExists
		}
		return err
	}

	// Update domain device with generated ID
	device.ID = dbDevice.ID
	device.CreatedAt = dbDevice.CreatedAt
	device.UpdatedAt = dbDevice.UpdatedAt

	return nil
}

// FindByID finds a device by ID
func (r *DeviceRepository) FindByID(ctx context.Context, id int64) (*domain.Device, error) {
	var dbDevice models.Device
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&dbDevice).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrDeviceNotFound
		}
		return nil, err
	}

	return dbDevice.ToDomain(), nil
}

// FindByUserID finds all devices for a user
func (r *DeviceRepository) FindByUserID(ctx context.Context, userID int64) ([]*domain.Device, error) {
	var dbDevices []models.Device
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&dbDevices).Error; err != nil {
		return nil, err
	}

	devices := make([]*domain.Device, len(dbDevices))
	for i, dbDevice := range dbDevices {
		devices[i] = dbDevice.ToDomain()
	}

	return devices, nil
}

// FindActiveByUserID finds all active devices for a user
func (r *DeviceRepository) FindActiveByUserID(ctx context.Context, userID int64) ([]*domain.Device, error) {
	var dbDevices []models.Device
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_active = ?", userID, true).
		Order("last_used_at DESC NULLS LAST").
		Find(&dbDevices).Error; err != nil {
		return nil, err
	}

	devices := make([]*domain.Device, len(dbDevices))
	for i, dbDevice := range dbDevices {
		devices[i] = dbDevice.ToDomain()
	}

	return devices, nil
}

// FindByToken finds a device by token
func (r *DeviceRepository) FindByToken(ctx context.Context, token string) (*domain.Device, error) {
	var dbDevice models.Device
	if err := r.db.WithContext(ctx).
		Where("device_token = ?", token).
		First(&dbDevice).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrDeviceNotFound
		}
		return nil, err
	}

	return dbDevice.ToDomain(), nil
}

// FindByUserIDAndToken finds a device by user ID and token
func (r *DeviceRepository) FindByUserIDAndToken(ctx context.Context, userID int64, token string) (*domain.Device, error) {
	var dbDevice models.Device
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND device_token = ?", userID, token).
		First(&dbDevice).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrDeviceNotFound
		}
		return nil, err
	}

	return dbDevice.ToDomain(), nil
}

// Update updates device information
func (r *DeviceRepository) Update(ctx context.Context, device *domain.Device) error {
	dbDevice := &models.Device{}
	dbDevice.FromDomain(device)

	result := r.db.WithContext(ctx).
		Model(&models.Device{}).
		Where("id = ?", device.ID).
		Updates(dbDevice)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrDeviceNotFound
	}

	return nil
}

// Delete deletes a device
func (r *DeviceRepository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&models.Device{}, id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrDeviceNotFound
	}

	return nil
}

// DeleteByToken deletes a device by user ID and token
func (r *DeviceRepository) DeleteByToken(ctx context.Context, userID int64, token string) error {
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND device_token = ?", userID, token).
		Delete(&models.Device{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrDeviceNotFound
	}

	return nil
}

// UpdateLastUsed updates the last used timestamp
func (r *DeviceRepository) UpdateLastUsed(ctx context.Context, id int64) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&models.Device{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_used_at": now,
			"updated_at":   now,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrDeviceNotFound
	}

	return nil
}

// DeactivateStaleDevices deactivates devices not used since the given time
func (r *DeviceRepository) DeactivateStaleDevices(ctx context.Context, before time.Time) (int64, error) {
	result := r.db.WithContext(ctx).
		Model(&models.Device{}).
		Where("is_active = ? AND (last_used_at < ? OR last_used_at IS NULL)", true, before).
		Updates(map[string]interface{}{
			"is_active":  false,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

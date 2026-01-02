package models

import (
	"time"

	"github.com/yourusername/notinoteapp/internal/core/domain"
	"gorm.io/gorm"
)

// User represents the database model for users
type User struct {
	ID           int64             `gorm:"primaryKey;autoIncrement"`
	Email        string            `gorm:"uniqueIndex;not null;size:255"`
	Name         string            `gorm:"not null;size:255"`
	PasswordHash string            `gorm:"size:255"`
	Provider     domain.AuthProvider `gorm:"type:varchar(20);not null;default:'email'"`
	ProviderID   string            `gorm:"size:255;index:idx_provider_id"`
	AvatarURL    string            `gorm:"size:500"`
	IsActive     bool              `gorm:"not null;default:true"`
	CreatedAt    time.Time         `gorm:"autoCreateTime"`
	UpdatedAt    time.Time         `gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt    `gorm:"index"`
}

// TableName specifies the table name for GORM
func (User) TableName() string {
	return "users"
}

// ToDomain converts database model to domain entity
func (u *User) ToDomain() *domain.User {
	return &domain.User{
		ID:           u.ID,
		Email:        u.Email,
		Name:         u.Name,
		PasswordHash: u.PasswordHash,
		Provider:     u.Provider,
		ProviderID:   u.ProviderID,
		AvatarURL:    u.AvatarURL,
		IsActive:     u.IsActive,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

// FromDomain converts domain entity to database model
func (u *User) FromDomain(domainUser *domain.User) {
	u.ID = domainUser.ID
	u.Email = domainUser.Email
	u.Name = domainUser.Name
	u.PasswordHash = domainUser.PasswordHash
	u.Provider = domainUser.Provider
	u.ProviderID = domainUser.ProviderID
	u.AvatarURL = domainUser.AvatarURL
	u.IsActive = domainUser.IsActive
	u.CreatedAt = domainUser.CreatedAt
	u.UpdatedAt = domainUser.UpdatedAt
}

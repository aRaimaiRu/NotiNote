package repositories

import (
	"context"
	"errors"

	"github.com/yourusername/notinoteapp/internal/adapters/secondary/database/postgres/models"
	"github.com/yourusername/notinoteapp/internal/core/domain"
	"gorm.io/gorm"
)

// UserRepository implements the user repository interface using PostgreSQL
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	dbUser := &models.User{}
	dbUser.FromDomain(user)

	if err := r.db.WithContext(ctx).Create(dbUser).Error; err != nil {
		return err
	}

	// Update domain user with generated ID
	user.ID = dbUser.ID
	user.CreatedAt = dbUser.CreatedAt
	user.UpdatedAt = dbUser.UpdatedAt

	return nil
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	var dbUser models.User
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&dbUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	return dbUser.ToDomain(), nil
}

// FindByEmail finds a user by email
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var dbUser models.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&dbUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	return dbUser.ToDomain(), nil
}

// FindByProvider finds a user by OAuth provider and provider ID
func (r *UserRepository) FindByProvider(ctx context.Context, provider domain.AuthProvider, providerID string) (*domain.User, error) {
	var dbUser models.User
	if err := r.db.WithContext(ctx).
		Where("provider = ? AND provider_id = ?", provider, providerID).
		First(&dbUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	return dbUser.ToDomain(), nil
}

// Update updates user information
func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	dbUser := &models.User{}
	dbUser.FromDomain(user)

	result := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", user.ID).
		Updates(dbUser)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

// Delete soft deletes a user
func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&models.User{}, id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

// List retrieves users with pagination
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*domain.User, int64, error) {
	var dbUsers []models.User
	var total int64

	// Count total users
	if err := r.db.WithContext(ctx).Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated users
	if err := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&dbUsers).Error; err != nil {
		return nil, 0, err
	}

	// Convert to domain models
	users := make([]*domain.User, len(dbUsers))
	for i, dbUser := range dbUsers {
		users[i] = dbUser.ToDomain()
	}

	return users, total, nil
}

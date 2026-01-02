package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/notinoteapp/internal/adapters/secondary/database/postgres/models"
	"github.com/yourusername/notinoteapp/internal/core/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	// Auto migrate the models
	err = db.AutoMigrate(&models.User{})
	require.NoError(t, err)

	return db
}

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &domain.User{
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "hashed-password",
		Provider:     domain.AuthProviderEmail,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Verify ID was assigned
	assert.NotZero(t, user.ID)
	assert.NotZero(t, user.CreatedAt)
	assert.NotZero(t, user.UpdatedAt)

	// Verify user was actually created
	var dbUser models.User
	err = db.First(&dbUser, user.ID).Error
	require.NoError(t, err)
	assert.Equal(t, user.Email, dbUser.Email)
	assert.Equal(t, user.Name, dbUser.Name)
}

func TestUserRepository_Create_DuplicateEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user1 := &domain.User{
		Email:        "duplicate@example.com",
		Name:         "User 1",
		PasswordHash: "hash1",
		Provider:     domain.AuthProviderEmail,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	user2 := &domain.User{
		Email:        "duplicate@example.com",
		Name:         "User 2",
		PasswordHash: "hash2",
		Provider:     domain.AuthProviderEmail,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(ctx, user1)
	require.NoError(t, err)

	err = repo.Create(ctx, user2)
	assert.Error(t, err) // Should fail due to unique constraint
}

func TestUserRepository_FindByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create a user first
	user := &domain.User{
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "hashed-password",
		Provider:     domain.AuthProviderEmail,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Find by ID
	foundUser, err := repo.FindByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.Email, foundUser.Email)
	assert.Equal(t, user.Name, foundUser.Name)
	assert.Equal(t, user.Provider, foundUser.Provider)
}

func TestUserRepository_FindByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user, err := repo.FindByID(ctx, 99999)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrUserNotFound)
	assert.Nil(t, user)
}

func TestUserRepository_FindByEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create a user
	user := &domain.User{
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "hashed-password",
		Provider:     domain.AuthProviderEmail,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Find by email
	foundUser, err := repo.FindByEmail(ctx, "test@example.com")
	require.NoError(t, err)
	assert.Equal(t, user.ID, foundUser.ID)
	assert.Equal(t, user.Email, foundUser.Email)
	assert.Equal(t, user.Name, foundUser.Name)
}

func TestUserRepository_FindByEmail_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user, err := repo.FindByEmail(ctx, "notfound@example.com")
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrUserNotFound)
	assert.Nil(t, user)
}

func TestUserRepository_FindByProvider(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create an OAuth user
	user := &domain.User{
		Email:      "oauth@gmail.com",
		Name:       "OAuth User",
		Provider:   domain.AuthProviderGoogle,
		ProviderID: "google-123",
		AvatarURL:  "https://example.com/avatar.jpg",
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Find by provider
	foundUser, err := repo.FindByProvider(ctx, domain.AuthProviderGoogle, "google-123")
	require.NoError(t, err)
	assert.Equal(t, user.ID, foundUser.ID)
	assert.Equal(t, user.Email, foundUser.Email)
	assert.Equal(t, user.Provider, foundUser.Provider)
	assert.Equal(t, user.ProviderID, foundUser.ProviderID)
}

func TestUserRepository_FindByProvider_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user, err := repo.FindByProvider(ctx, domain.AuthProviderGoogle, "nonexistent-id")
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrUserNotFound)
	assert.Nil(t, user)
}

func TestUserRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create a user
	user := &domain.User{
		Email:        "test@example.com",
		Name:         "Original Name",
		PasswordHash: "hashed-password",
		Provider:     domain.AuthProviderEmail,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Update the user
	user.Name = "Updated Name"
	user.IsActive = false
	user.UpdatedAt = time.Now()

	err = repo.Update(ctx, user)
	require.NoError(t, err)

	// Verify update
	foundUser, err := repo.FindByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", foundUser.Name)
	assert.False(t, foundUser.IsActive)
}

func TestUserRepository_Update_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &domain.User{
		ID:           99999,
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "hashed-password",
	}

	err := repo.Update(ctx, user)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrUserNotFound)
}

func TestUserRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create a user
	user := &domain.User{
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "hashed-password",
		Provider:     domain.AuthProviderEmail,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Delete the user (soft delete)
	err = repo.Delete(ctx, user.ID)
	require.NoError(t, err)

	// Verify user is soft deleted (GORM's deleted_at is set)
	var dbUser models.User
	err = db.Unscoped().First(&dbUser, user.ID).Error
	require.NoError(t, err)
	assert.NotNil(t, dbUser.DeletedAt)

	// Verify user cannot be found with normal query
	foundUser, err := repo.FindByID(ctx, user.ID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrUserNotFound)
	assert.Nil(t, foundUser)
}

func TestUserRepository_Delete_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	err := repo.Delete(ctx, 99999)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrUserNotFound)
}

func TestUserRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create multiple users
	for i := 1; i <= 5; i++ {
		user := &domain.User{
			Email:        "user" + string(rune(i)) + "@example.com",
			Name:         "User " + string(rune(i)),
			PasswordHash: "hashed-password",
			Provider:     domain.AuthProviderEmail,
			IsActive:     true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		err := repo.Create(ctx, user)
		require.NoError(t, err)
		time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	}

	// List users with pagination
	users, total, err := repo.List(ctx, 3, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, users, 3)

	// Second page
	users, total, err = repo.List(ctx, 3, 3)
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, users, 2)
}

func TestUserRepository_List_Empty(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	users, total, err := repo.List(ctx, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Len(t, users, 0)
}

func TestUserRepository_OAuthUserCreation(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create OAuth user
	oauthInfo := &domain.OAuthUserInfo{
		Provider:   domain.AuthProviderGoogle,
		ProviderID: "google-123",
		Email:      "oauth@gmail.com",
		Name:       "OAuth User",
		AvatarURL:  "https://example.com/avatar.jpg",
	}

	user, err := domain.NewOAuthUser(oauthInfo)
	require.NoError(t, err)

	err = repo.Create(ctx, user)
	require.NoError(t, err)

	// Verify OAuth user was created correctly
	foundUser, err := repo.FindByProvider(ctx, domain.AuthProviderGoogle, "google-123")
	require.NoError(t, err)
	assert.Equal(t, domain.AuthProviderGoogle, foundUser.Provider)
	assert.Equal(t, "google-123", foundUser.ProviderID)
	assert.Empty(t, foundUser.PasswordHash) // OAuth users don't have passwords
	assert.Equal(t, "https://example.com/avatar.jpg", foundUser.AvatarURL)
}

func TestUserRepository_EmailUserVsOAuthUser(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create email user
	emailUser := &domain.User{
		Email:        "test@example.com",
		Name:         "Email User",
		PasswordHash: "hashed-password",
		Provider:     domain.AuthProviderEmail,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := repo.Create(ctx, emailUser)
	require.NoError(t, err)

	// Create OAuth user with different email
	oauthUser := &domain.User{
		Email:      "oauth@gmail.com",
		Name:       "OAuth User",
		Provider:   domain.AuthProviderGoogle,
		ProviderID: "google-123",
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	err = repo.Create(ctx, oauthUser)
	require.NoError(t, err)

	// Verify both users exist
	foundEmailUser, err := repo.FindByEmail(ctx, "test@example.com")
	require.NoError(t, err)
	assert.Equal(t, domain.AuthProviderEmail, foundEmailUser.Provider)
	assert.NotEmpty(t, foundEmailUser.PasswordHash)

	foundOAuthUser, err := repo.FindByEmail(ctx, "oauth@gmail.com")
	require.NoError(t, err)
	assert.Equal(t, domain.AuthProviderGoogle, foundOAuthUser.Provider)
	assert.Empty(t, foundOAuthUser.PasswordHash)
	assert.NotEmpty(t, foundOAuthUser.ProviderID)
}

func TestUserRepository_UpdateOAuthUserProfile(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create OAuth user
	user := &domain.User{
		Email:      "oauth@gmail.com",
		Name:       "Original Name",
		Provider:   domain.AuthProviderGoogle,
		ProviderID: "google-123",
		AvatarURL:  "https://example.com/old-avatar.jpg",
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Simulate profile update from OAuth provider
	user.Name = "Updated Name"
	user.AvatarURL = "https://example.com/new-avatar.jpg"
	user.UpdatedAt = time.Now()

	err = repo.Update(ctx, user)
	require.NoError(t, err)

	// Verify update
	foundUser, err := repo.FindByProvider(ctx, domain.AuthProviderGoogle, "google-123")
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", foundUser.Name)
	assert.Equal(t, "https://example.com/new-avatar.jpg", foundUser.AvatarURL)
}

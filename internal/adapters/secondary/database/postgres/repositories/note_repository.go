package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/yourusername/notinoteapp/internal/adapters/secondary/database/postgres/models"
	"github.com/yourusername/notinoteapp/internal/core/domain"
	"github.com/yourusername/notinoteapp/internal/core/ports"
	"gorm.io/gorm"
)

// NoteRepository implements the note repository interface using PostgreSQL
type NoteRepository struct {
	db *gorm.DB
}

// NewNoteRepository creates a new note repository
func NewNoteRepository(db *gorm.DB) *NoteRepository {
	return &NoteRepository{db: db}
}

// Create creates a new note
func (r *NoteRepository) Create(ctx context.Context, note *domain.Note) error {
	dbNote := &models.Note{}
	dbNote.FromDomain(note)

	if err := r.db.WithContext(ctx).Create(dbNote).Error; err != nil {
		return fmt.Errorf("failed to create note: %w", err)
	}

	// Update domain note with generated fields
	note.ID = dbNote.ID
	note.CreatedAt = dbNote.CreatedAt
	note.UpdatedAt = dbNote.UpdatedAt
	note.Path = dbNote.Path // Set by database trigger

	return nil
}

// FindByID finds a note by ID
func (r *NoteRepository) FindByID(ctx context.Context, id int64) (*domain.Note, error) {
	var dbNote models.Note

	err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = ?", id, false).
		First(&dbNote).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNoteNotFound
		}
		return nil, fmt.Errorf("failed to find note: %w", err)
	}

	return dbNote.ToDomain(), nil
}

// Update updates a note
func (r *NoteRepository) Update(ctx context.Context, note *domain.Note) error {
	dbNote := &models.Note{}
	dbNote.FromDomain(note)

	result := r.db.WithContext(ctx).
		Model(&models.Note{}).
		Where("id = ? AND is_deleted = ?", note.ID, false).
		Updates(dbNote)

	if result.Error != nil {
		return fmt.Errorf("failed to update note: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return domain.ErrNoteNotFound
	}

	return nil
}

// Delete soft deletes a note
func (r *NoteRepository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).
		Model(&models.Note{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_deleted": true,
			"deleted_at": gorm.Expr("CURRENT_TIMESTAMP"),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to delete note: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return domain.ErrNoteNotFound
	}

	return nil
}

// FindByUserID finds all notes for a user with filtering and pagination
func (r *NoteRepository) FindByUserID(ctx context.Context, userID int64, filters ports.NoteFilters) ([]*domain.Note, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.Note{}).
		Where("user_id = ? AND is_deleted = ?", userID, false)

	// Apply filters
	query = r.applyFilters(query, filters)

	// Count total matching records
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count notes: %w", err)
	}

	// Apply sorting
	query = r.applySorting(query, filters)

	// Apply pagination
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	var dbNotes []models.Note
	if err := query.Find(&dbNotes).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to find notes: %w", err)
	}

	notes := make([]*domain.Note, len(dbNotes))
	for i, dbNote := range dbNotes {
		notes[i] = dbNote.ToDomain()
	}

	return notes, total, nil
}

// FindChildren finds direct children of a parent note
func (r *NoteRepository) FindChildren(ctx context.Context, parentID int64) ([]*domain.Note, error) {
	var dbNotes []models.Note

	err := r.db.WithContext(ctx).
		Where("parent_id = ? AND is_deleted = ?", parentID, false).
		Order("position ASC").
		Find(&dbNotes).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find children: %w", err)
	}

	notes := make([]*domain.Note, len(dbNotes))
	for i, dbNote := range dbNotes {
		notes[i] = dbNote.ToDomain()
	}

	return notes, nil
}

// FindDescendants finds all descendants of a parent note using materialized path
func (r *NoteRepository) FindDescendants(ctx context.Context, parentID int64) ([]*domain.Note, error) {
	// First get the parent to get its path
	parent, err := r.FindByID(ctx, parentID)
	if err != nil {
		return nil, err
	}

	var dbNotes []models.Note

	// Use path pattern matching for efficient descendant query
	// If parent path is "/1/23/", this matches all notes with path like "/1/23/.../"
	err = r.db.WithContext(ctx).
		Where("path LIKE ? AND id != ? AND is_deleted = ?", parent.Path+"%", parentID, false).
		Order("path ASC, position ASC").
		Find(&dbNotes).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find descendants: %w", err)
	}

	notes := make([]*domain.Note, len(dbNotes))
	for i, dbNote := range dbNotes {
		notes[i] = dbNote.ToDomain()
	}

	return notes, nil
}

// FindAncestors finds all ancestors of a note using materialized path
func (r *NoteRepository) FindAncestors(ctx context.Context, noteID int64) ([]*domain.Note, error) {
	// Get the note to parse its path
	note, err := r.FindByID(ctx, noteID)
	if err != nil {
		return nil, err
	}

	// Parse ancestor IDs from path
	// Path format: "/1/23/456/" -> ancestor IDs: [1, 23]
	ancestorIDs := r.parseAncestorIDs(note.Path, noteID)
	if len(ancestorIDs) == 0 {
		return []*domain.Note{}, nil
	}

	var dbNotes []models.Note

	err = r.db.WithContext(ctx).
		Where("id IN ? AND is_deleted = ?", ancestorIDs, false).
		Order("depth ASC").
		Find(&dbNotes).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find ancestors: %w", err)
	}

	notes := make([]*domain.Note, len(dbNotes))
	for i, dbNote := range dbNotes {
		notes[i] = dbNote.ToDomain()
	}

	return notes, nil
}

// MoveNote moves a note to a new parent and position
func (r *NoteRepository) MoveNote(ctx context.Context, noteID int64, newParentID *int64, newPosition int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Get current note
		var note models.Note
		if err := tx.Where("id = ?", noteID).First(&note).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrNoteNotFound
			}
			return err
		}

		// Check for circular reference
		if newParentID != nil {
			if *newParentID == noteID {
				return domain.ErrCircularReference
			}

			// Get new parent to check if note is ancestor of new parent
			var newParent models.Note
			if err := tx.Where("id = ?", *newParentID).First(&newParent).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return domain.ErrInvalidParentNote
				}
				return err
			}

			// Check if note's path is prefix of new parent's path (would create cycle)
			if strings.HasPrefix(newParent.Path, note.Path) {
				return domain.ErrCircularReference
			}

			// Check max depth
			if newParent.Depth+1 > domain.MaxNestingDepth {
				return domain.ErrMaxDepthExceeded
			}
		}

		// Update parent and position
		// The trigger will automatically update path and depth
		updates := map[string]interface{}{
			"position": newPosition,
		}

		if newParentID == nil {
			updates["parent_id"] = gorm.Expr("NULL")
		} else {
			updates["parent_id"] = *newParentID
		}

		if err := tx.Model(&note).Updates(updates).Error; err != nil {
			return err
		}

		return nil
	})
}

// UpdateBlocks updates the blocks of a note
func (r *NoteRepository) UpdateBlocks(ctx context.Context, noteID int64, blocks []domain.Block) error {
	blocksJSON, err := json.Marshal(blocks)
	if err != nil {
		return fmt.Errorf("failed to marshal blocks: %w", err)
	}

	result := r.db.WithContext(ctx).
		Model(&models.Note{}).
		Where("id = ? AND is_deleted = ?", noteID, false).
		Update("blocks", blocksJSON)

	if result.Error != nil {
		return fmt.Errorf("failed to update blocks: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return domain.ErrNoteNotFound
	}

	return nil
}

// Search searches notes by title with filters
func (r *NoteRepository) Search(ctx context.Context, userID int64, query string, filters ports.NoteFilters) ([]*domain.Note, int64, error) {
	dbQuery := r.db.WithContext(ctx).Model(&models.Note{}).
		Where("user_id = ? AND is_deleted = ?", userID, false)

	// Full-text search on title
	if query != "" {
		dbQuery = dbQuery.Where("to_tsvector('english', title) @@ plainto_tsquery('english', ?)", query)
	}

	// Apply filters
	dbQuery = r.applyFilters(dbQuery, filters)

	// Count total
	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count notes: %w", err)
	}

	// Apply sorting
	dbQuery = r.applySorting(dbQuery, filters)

	// Apply pagination
	if filters.Limit > 0 {
		dbQuery = dbQuery.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		dbQuery = dbQuery.Offset(filters.Offset)
	}

	var dbNotes []models.Note
	if err := dbQuery.Find(&dbNotes).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to search notes: %w", err)
	}

	notes := make([]*domain.Note, len(dbNotes))
	for i, dbNote := range dbNotes {
		notes[i] = dbNote.ToDomain()
	}

	return notes, total, nil
}

// BulkArchive archives multiple notes
func (r *NoteRepository) BulkArchive(ctx context.Context, noteIDs []int64) error {
	if len(noteIDs) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).
		Model(&models.Note{}).
		Where("id IN ?", noteIDs).
		Update("is_archived", true)

	if result.Error != nil {
		return fmt.Errorf("failed to bulk archive notes: %w", result.Error)
	}

	return nil
}

// BulkDelete soft deletes multiple notes
func (r *NoteRepository) BulkDelete(ctx context.Context, noteIDs []int64) error {
	if len(noteIDs) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).
		Model(&models.Note{}).
		Where("id IN ?", noteIDs).
		Updates(map[string]interface{}{
			"is_deleted": true,
			"deleted_at": gorm.Expr("CURRENT_TIMESTAMP"),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to bulk delete notes: %w", result.Error)
	}

	return nil
}

// CheckOwnership checks if a user owns a note
func (r *NoteRepository) CheckOwnership(ctx context.Context, noteID, userID int64) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.Note{}).
		Where("id = ? AND user_id = ? AND is_deleted = ?", noteID, userID, false).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to check ownership: %w", err)
	}

	return count > 0, nil
}

// Helper methods

// applyFilters applies filters to a query
func (r *NoteRepository) applyFilters(query *gorm.DB, filters ports.NoteFilters) *gorm.DB {
	if filters.ParentID != nil {
		query = query.Where("parent_id = ?", *filters.ParentID)
	}

	if filters.IsArchived != nil {
		query = query.Where("is_archived = ?", *filters.IsArchived)
	}

	if filters.SearchQuery != "" {
		query = query.Where("to_tsvector('english', title) @@ plainto_tsquery('english', ?)", filters.SearchQuery)
	}

	// TODO: Add property filtering when needed
	// This would require JSONB queries like:
	// query.Where("properties->>'status' = ?", value)

	return query
}

// applySorting applies sorting to a query
func (r *NoteRepository) applySorting(query *gorm.DB, filters ports.NoteFilters) *gorm.DB {
	sortBy := filters.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}

	sortOrder := filters.SortOrder
	if sortOrder == "" {
		sortOrder = "desc"
	}

	// Validate sortBy to prevent SQL injection
	validSortFields := map[string]bool{
		"created_at": true,
		"updated_at": true,
		"title":      true,
		"position":   true,
	}

	if !validSortFields[sortBy] {
		sortBy = "created_at"
	}

	// Validate sortOrder
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	return query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))
}

// parseAncestorIDs parses ancestor IDs from a materialized path
// Path format: "/1/23/456/" -> returns [1, 23] (excluding the note itself)
func (r *NoteRepository) parseAncestorIDs(path string, excludeID int64) []int64 {
	// Remove leading and trailing slashes
	path = strings.Trim(path, "/")
	if path == "" {
		return []int64{}
	}

	// Split by slash
	parts := strings.Split(path, "/")
	ancestorIDs := make([]int64, 0, len(parts)-1)

	for _, part := range parts {
		var id int64
		if _, err := fmt.Sscanf(part, "%d", &id); err == nil {
			if id != excludeID {
				ancestorIDs = append(ancestorIDs, id)
			}
		}
	}

	return ancestorIDs
}

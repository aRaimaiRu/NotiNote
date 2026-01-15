package services

import (
	"context"
	"fmt"
	"time"

	"github.com/yourusername/notinoteapp/internal/core/domain"
	"github.com/yourusername/notinoteapp/internal/core/ports"
)

// NoteService implements business logic for note operations
type NoteService struct {
	noteRepo ports.NoteRepository
}

// NewNoteService creates a new NoteService instance
func NewNoteService(noteRepo ports.NoteRepository) *NoteService {
	return &NoteService{
		noteRepo: noteRepo,
	}
}

// CreateNote creates a new note with validation
func (s *NoteService) CreateNote(ctx context.Context, userID int64, title string, parentID *int64) (*domain.Note, error) {
	// Create new note using domain factory
	note, err := domain.NewNote(userID, title)
	if err != nil {
		return nil, fmt.Errorf("failed to create note: %w", err)
	}

	// Set parent if provided
	if parentID != nil {
		// Verify parent exists and user owns it
		parent, err := s.noteRepo.FindByID(ctx, *parentID)
		if err != nil {
			return nil, fmt.Errorf("parent note not found: %w", err)
		}

		if parent.UserID != userID {
			return nil, domain.ErrUnauthorizedAccess
		}

		// Check nesting depth
		if parent.Depth >= 10 {
			return nil, domain.ErrMaxDepthExceeded
		}

		// Set parent with depth
		if err := note.SetParent(parentID, parent.Depth); err != nil {
			return nil, fmt.Errorf("failed to set parent: %w", err)
		}
	}

	// Save to repository
	if err := s.noteRepo.Create(ctx, note); err != nil {
		return nil, fmt.Errorf("failed to save note: %w", err)
	}

	return note, nil
}

// GetNote retrieves a note by ID with ownership validation
func (s *NoteService) GetNote(ctx context.Context, noteID, userID int64) (*domain.Note, error) {
	note, err := s.noteRepo.FindByID(ctx, noteID)
	if err != nil {
		return nil, fmt.Errorf("note not found: %w", err)
	}

	// Verify ownership
	if note.UserID != userID {
		return nil, domain.ErrUnauthorizedAccess
	}

	return note, nil
}

// UpdateNote updates an existing note with validation
func (s *NoteService) UpdateNote(ctx context.Context, noteID, userID int64, title *string, icon *string, coverImage *string) (*domain.Note, error) {
	// Retrieve existing note
	note, err := s.GetNote(ctx, noteID, userID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if title != nil {
		if len(*title) == 0 || len(*title) > 500 {
			return nil, domain.ErrInvalidNoteTitle
		}
		note.Title = *title
	}

	if icon != nil {
		note.Icon = *icon
	}

	if coverImage != nil {
		note.CoverImage = *coverImage
	}

	// Save changes
	if err := s.noteRepo.Update(ctx, note); err != nil {
		return nil, fmt.Errorf("failed to update note: %w", err)
	}

	return note, nil
}

// DeleteNote soft deletes a note and all its descendants
func (s *NoteService) DeleteNote(ctx context.Context, noteID, userID int64) error {
	// Verify ownership
	note, err := s.GetNote(ctx, noteID, userID)
	if err != nil {
		return err
	}

	// Soft delete the note
	note.SoftDelete()

	// Get all descendants and soft delete them
	descendants, err := s.noteRepo.FindDescendants(ctx, noteID)
	if err != nil {
		return fmt.Errorf("failed to get descendants: %w", err)
	}

	// Collect IDs for bulk delete
	descendantIDs := make([]int64, len(descendants))
	for i, desc := range descendants {
		descendantIDs[i] = desc.ID
	}

	// Bulk soft delete descendants
	if len(descendantIDs) > 0 {
		if err := s.noteRepo.BulkDelete(ctx, descendantIDs); err != nil {
			return fmt.Errorf("failed to delete descendants: %w", err)
		}
	}

	// Update the parent note
	if err := s.noteRepo.Update(ctx, note); err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}

	return nil
}

// RestoreNote restores a soft-deleted note
func (s *NoteService) RestoreNote(ctx context.Context, noteID, userID int64) (*domain.Note, error) {
	// Get the deleted note
	note, err := s.noteRepo.FindByID(ctx, noteID)
	if err != nil {
		return nil, fmt.Errorf("note not found: %w", err)
	}

	// Verify ownership
	if note.UserID != userID {
		return nil, domain.ErrUnauthorizedAccess
	}

	// Restore the note
	note.Restore()

	if err := s.noteRepo.Update(ctx, note); err != nil {
		return nil, fmt.Errorf("failed to restore note: %w", err)
	}

	return note, nil
}

// ArchiveNote archives a note
func (s *NoteService) ArchiveNote(ctx context.Context, noteID, userID int64) (*domain.Note, error) {
	note, err := s.GetNote(ctx, noteID, userID)
	if err != nil {
		return nil, err
	}

	note.Archive()

	if err := s.noteRepo.Update(ctx, note); err != nil {
		return nil, fmt.Errorf("failed to archive note: %w", err)
	}

	return note, nil
}

// UnarchiveNote unarchives a note
func (s *NoteService) UnarchiveNote(ctx context.Context, noteID, userID int64) (*domain.Note, error) {
	note, err := s.noteRepo.FindByID(ctx, noteID)
	if err != nil {
		return nil, fmt.Errorf("note not found: %w", err)
	}

	if note.UserID != userID {
		return nil, domain.ErrUnauthorizedAccess
	}

	note.IsArchived = false

	if err := s.noteRepo.Update(ctx, note); err != nil {
		return nil, fmt.Errorf("failed to unarchive note: %w", err)
	}

	return note, nil
}

// ListNotes retrieves notes with filtering and pagination
func (s *NoteService) ListNotes(ctx context.Context, userID int64, filters ports.NoteFilters) ([]*domain.Note, int64, error) {
	return s.noteRepo.FindByUserID(ctx, userID, filters)
}

// GetChildren retrieves direct children of a note
func (s *NoteService) GetChildren(ctx context.Context, parentID, userID int64) ([]*domain.Note, error) {
	// Verify parent ownership
	if _, err := s.GetNote(ctx, parentID, userID); err != nil {
		return nil, err
	}

	return s.noteRepo.FindChildren(ctx, parentID)
}

// GetDescendants retrieves all descendants of a note
func (s *NoteService) GetDescendants(ctx context.Context, parentID, userID int64) ([]*domain.Note, error) {
	// Verify parent ownership
	if _, err := s.GetNote(ctx, parentID, userID); err != nil {
		return nil, err
	}

	return s.noteRepo.FindDescendants(ctx, parentID)
}

// GetAncestors retrieves all ancestors of a note (breadcrumb trail)
func (s *NoteService) GetAncestors(ctx context.Context, noteID, userID int64) ([]*domain.Note, error) {
	// Verify note ownership
	if _, err := s.GetNote(ctx, noteID, userID); err != nil {
		return nil, err
	}

	return s.noteRepo.FindAncestors(ctx, noteID)
}

// MoveNote moves a note to a new parent with validation
func (s *NoteService) MoveNote(ctx context.Context, noteID, userID int64, newParentID *int64, newPosition int) error {
	// Verify ownership of the note being moved
	note, err := s.GetNote(ctx, noteID, userID)
	if err != nil {
		return err
	}

	// If new parent is provided, verify ownership and nesting depth
	if newParentID != nil {
		parent, err := s.GetNote(ctx, *newParentID, userID)
		if err != nil {
			return fmt.Errorf("new parent not found: %w", err)
		}

		// Check if moving would exceed max depth
		// Get descendants count to estimate new depth
		descendants, err := s.noteRepo.FindDescendants(ctx, noteID)
		if err != nil {
			return fmt.Errorf("failed to check descendants: %w", err)
		}

		maxDescendantDepth := 0
		for _, desc := range descendants {
			relativeDepth := desc.Depth - note.Depth
			if relativeDepth > maxDescendantDepth {
				maxDescendantDepth = relativeDepth
			}
		}

		newDepth := parent.Depth + 1 + maxDescendantDepth
		if newDepth > 10 {
			return domain.ErrMaxDepthExceeded
		}
	}

	// Perform the move
	if err := s.noteRepo.MoveNote(ctx, noteID, newParentID, newPosition); err != nil {
		return fmt.Errorf("failed to move note: %w", err)
	}

	return nil
}

// AddBlock adds a new block to a note
func (s *NoteService) AddBlock(ctx context.Context, noteID, userID int64, blockType domain.BlockType, content *domain.BlockContent) (*domain.Note, error) {
	note, err := s.GetNote(ctx, noteID, userID)
	if err != nil {
		return nil, err
	}

	// Validate block type and content
	if blockType == "" {
		return nil, domain.ErrInvalidBlockType
	}
	if content == nil {
		return nil, fmt.Errorf("block content is required")
	}

	// Create block with generated ID
	block := domain.Block{
		ID:      generateBlockID(),
		Type:    blockType,
		Content: content,
	}

	// Add block using domain method
	if err := note.AddBlock(block); err != nil {
		return nil, fmt.Errorf("failed to add block: %w", err)
	}

	// Save updated blocks
	if err := s.noteRepo.UpdateBlocks(ctx, noteID, note.Blocks); err != nil {
		return nil, fmt.Errorf("failed to save blocks: %w", err)
	}

	return note, nil
}

// generateBlockID generates a unique block ID (simplified UUID)
func generateBlockID() string {
	return fmt.Sprintf("block_%d", time.Now().UnixNano())
}

// UpdateBlock updates an existing block
func (s *NoteService) UpdateBlock(ctx context.Context, noteID, userID int64, blockID string, content *domain.BlockContent) (*domain.Note, error) {
	note, err := s.GetNote(ctx, noteID, userID)
	if err != nil {
		return nil, err
	}

	// Update block using domain method
	if err := note.UpdateBlock(blockID, content); err != nil {
		return nil, fmt.Errorf("failed to update block: %w", err)
	}

	// Save updated blocks
	if err := s.noteRepo.UpdateBlocks(ctx, noteID, note.Blocks); err != nil {
		return nil, fmt.Errorf("failed to save blocks: %w", err)
	}

	return note, nil
}

// DeleteBlock removes a block from a note
func (s *NoteService) DeleteBlock(ctx context.Context, noteID, userID int64, blockID string) (*domain.Note, error) {
	note, err := s.GetNote(ctx, noteID, userID)
	if err != nil {
		return nil, err
	}

	// Delete block using domain method
	if err := note.DeleteBlock(blockID); err != nil {
		return nil, fmt.Errorf("failed to delete block: %w", err)
	}

	// Save updated blocks
	if err := s.noteRepo.UpdateBlocks(ctx, noteID, note.Blocks); err != nil {
		return nil, fmt.Errorf("failed to save blocks: %w", err)
	}

	return note, nil
}

// ReorderBlocks changes the order of blocks
func (s *NoteService) ReorderBlocks(ctx context.Context, noteID, userID int64, blockOrder []string) (*domain.Note, error) {
	note, err := s.GetNote(ctx, noteID, userID)
	if err != nil {
		return nil, err
	}

	// Convert []string to map[string]int for the domain method
	blockOrders := make(map[string]int)
	for i, blockID := range blockOrder {
		blockOrders[blockID] = i
	}

	// Reorder blocks using domain method
	if err := note.ReorderBlocks(blockOrders); err != nil {
		return nil, fmt.Errorf("failed to reorder blocks: %w", err)
	}

	// Save updated blocks
	if err := s.noteRepo.UpdateBlocks(ctx, noteID, note.Blocks); err != nil {
		return nil, fmt.Errorf("failed to save blocks: %w", err)
	}

	return note, nil
}

// ReplaceBlocks replaces all blocks in a note
func (s *NoteService) ReplaceBlocks(ctx context.Context, noteID, userID int64, blocks []domain.Block) (*domain.Note, error) {
	note, err := s.GetNote(ctx, noteID, userID)
	if err != nil {
		return nil, err
	}

	// Validate all blocks
	for i, block := range blocks {
		if block.Type == "" {
			return nil, domain.ErrInvalidBlockType
		}
		if block.Content == nil {
			return nil, domain.ErrInvalidBlockContent
		}
		// Ensure block has an ID
		if block.ID == "" {
			blocks[i].ID = generateBlockID()
		}
	}

	note.Blocks = blocks

	// Save updated blocks
	if err := s.noteRepo.UpdateBlocks(ctx, noteID, note.Blocks); err != nil {
		return nil, fmt.Errorf("failed to save blocks: %w", err)
	}

	return note, nil
}

// SearchNotes searches notes by query
func (s *NoteService) SearchNotes(ctx context.Context, userID int64, query string, filters ports.NoteFilters) ([]*domain.Note, int64, error) {
	return s.noteRepo.Search(ctx, userID, query, filters)
}

// UpdateViewMetadata updates the view metadata for a note
func (s *NoteService) UpdateViewMetadata(ctx context.Context, noteID, userID int64, viewMetadata *domain.ViewMetadata) (*domain.Note, error) {
	note, err := s.GetNote(ctx, noteID, userID)
	if err != nil {
		return nil, err
	}

	// Validate view metadata
	if viewMetadata != nil {
		if viewMetadata.ViewType != domain.ViewTypeTable &&
			viewMetadata.ViewType != domain.ViewTypeBoard &&
			viewMetadata.ViewType != domain.ViewTypeList {
			return nil, domain.ErrInvalidViewType
		}
	}

	note.ViewMetadata = viewMetadata

	if err := s.noteRepo.Update(ctx, note); err != nil {
		return nil, fmt.Errorf("failed to update view metadata: %w", err)
	}

	return note, nil
}

// UpdateProperties updates custom properties for a note
func (s *NoteService) UpdateProperties(ctx context.Context, noteID, userID int64, properties map[string]interface{}) (*domain.Note, error) {
	note, err := s.GetNote(ctx, noteID, userID)
	if err != nil {
		return nil, err
	}

	note.Properties = properties

	if err := s.noteRepo.Update(ctx, note); err != nil {
		return nil, fmt.Errorf("failed to update properties: %w", err)
	}

	return note, nil
}

// ToggleFavorite toggles the favorite status of a note
func (s *NoteService) ToggleFavorite(ctx context.Context, noteID, userID int64) (*domain.Note, error) {
	note, err := s.GetNote(ctx, noteID, userID)
	if err != nil {
		return nil, err
	}

	// Toggle favorite using domain method
	note.ToggleFavorite()

	if err := s.noteRepo.Update(ctx, note); err != nil {
		return nil, fmt.Errorf("failed to toggle favorite: %w", err)
	}

	return note, nil
}

// AddTag adds a tag to a note
func (s *NoteService) AddTag(ctx context.Context, noteID, userID int64, tagID string) (*domain.Note, error) {
	// Verify note ownership
	_, err := s.GetNote(ctx, noteID, userID)
	if err != nil {
		return nil, err
	}

	// Add tag via repository
	if err := s.noteRepo.AddTag(ctx, noteID, tagID); err != nil {
		return nil, fmt.Errorf("failed to add tag: %w", err)
	}

	// Reload note with updated tags
	updatedNote, err := s.GetNote(ctx, noteID, userID)
	if err != nil {
		return nil, err
	}

	return updatedNote, nil
}

// RemoveTag removes a tag from a note
func (s *NoteService) RemoveTag(ctx context.Context, noteID, userID int64, tagID string) (*domain.Note, error) {
	// Verify note ownership
	_, err := s.GetNote(ctx, noteID, userID)
	if err != nil {
		return nil, err
	}

	// Remove tag via repository
	if err := s.noteRepo.RemoveTag(ctx, noteID, tagID); err != nil {
		return nil, fmt.Errorf("failed to remove tag: %w", err)
	}

	// Reload note with updated tags
	updatedNote, err := s.GetNote(ctx, noteID, userID)
	if err != nil {
		return nil, err
	}

	return updatedNote, nil
}

package dtos

import (
	"time"

	"github.com/yourusername/notinoteapp/internal/core/domain"
)

// CreateNoteRequest represents the request to create a new note
type CreateNoteRequest struct {
	Title    string `json:"title" binding:"required,min=1,max=500"`
	ParentID *int64 `json:"parent_id,omitempty"`
	Icon     string `json:"icon,omitempty"`
	Cover    string `json:"cover_image,omitempty"`
}

// UpdateNoteRequest represents the request to update a note
type UpdateNoteRequest struct {
	Title      *string `json:"title,omitempty" binding:"omitempty,min=1,max=500"`
	Icon       *string `json:"icon,omitempty"`
	CoverImage *string `json:"cover_image,omitempty"`
}

// MoveNoteRequest represents the request to move a note
type MoveNoteRequest struct {
	NewParentID *int64 `json:"new_parent_id,omitempty"`
	Position    int    `json:"position" binding:"min=0"`
}

// AddBlockRequest represents the request to add a block
type AddBlockRequest struct {
	Type    domain.BlockType    `json:"type" binding:"required"`
	Content *domain.BlockContent `json:"content" binding:"required"`
}

// UpdateBlockRequest represents the request to update a block
type UpdateBlockRequest struct {
	Content *domain.BlockContent `json:"content" binding:"required"`
}

// ReplaceBlocksRequest represents the request to replace all blocks
type ReplaceBlocksRequest struct {
	Blocks []domain.Block `json:"blocks" binding:"required"`
}

// ReorderBlocksRequest represents the request to reorder blocks
type ReorderBlocksRequest struct {
	BlockIDs []string `json:"block_ids" binding:"required,min=1"`
}

// UpdateViewMetadataRequest represents the request to update view metadata
type UpdateViewMetadataRequest struct {
	ViewType   domain.ViewType              `json:"view_type" binding:"required"`
	Properties []domain.ViewProperty        `json:"properties,omitempty"`
	Filters    []domain.ViewFilter          `json:"filters,omitempty"`
	Sorts      []domain.ViewSort            `json:"sorts,omitempty"`
}

// UpdatePropertiesRequest represents the request to update custom properties
type UpdatePropertiesRequest struct {
	Properties map[string]interface{} `json:"properties" binding:"required"`
}

// NoteResponse represents the response for a single note
type NoteResponse struct {
	ID           int64                  `json:"id"`
	UserID       int64                  `json:"user_id"`
	ParentID     *int64                 `json:"parent_id,omitempty"`
	Title        string                 `json:"title"`
	Icon         string                 `json:"icon,omitempty"`
	CoverImage   string                 `json:"cover_image,omitempty"`
	Blocks       []domain.Block         `json:"blocks"`
	ViewMetadata *domain.ViewMetadata   `json:"view_metadata,omitempty"`
	Properties   map[string]interface{} `json:"properties,omitempty"`
	Path         string                 `json:"path"`
	Depth        int                    `json:"depth"`
	Position     int                    `json:"position"`
	IsArchived   bool                   `json:"is_archived"`
	IsDeleted    bool                   `json:"is_deleted"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// NoteListResponse represents the response for a list of notes
type NoteListResponse struct {
	Notes      []NoteResponse     `json:"notes"`
	Pagination PaginationResponse `json:"pagination"`
}

// PaginationResponse represents pagination metadata
type PaginationResponse struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// NoteSummaryResponse represents a minimal note summary for lists
type NoteSummaryResponse struct {
	ID         int64     `json:"id"`
	Title      string    `json:"title"`
	Icon       string    `json:"icon,omitempty"`
	ParentID   *int64    `json:"parent_id,omitempty"`
	Depth      int       `json:"depth"`
	IsArchived bool      `json:"is_archived"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// NoteTreeResponse represents a hierarchical note structure
type NoteTreeResponse struct {
	Note     NoteSummaryResponse  `json:"note"`
	Children []NoteTreeResponse   `json:"children,omitempty"`
}

// BreadcrumbResponse represents a breadcrumb trail
type BreadcrumbResponse struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Icon  string `json:"icon,omitempty"`
}

// ToNoteResponse converts a domain note to a response DTO
func ToNoteResponse(note *domain.Note) NoteResponse {
	return NoteResponse{
		ID:           note.ID,
		UserID:       note.UserID,
		ParentID:     note.ParentID,
		Title:        note.Title,
		Icon:         note.Icon,
		CoverImage:   note.CoverImage,
		Blocks:       note.Blocks,
		ViewMetadata: note.ViewMetadata,
		Properties:   note.Properties,
		Path:         note.Path,
		Depth:        note.Depth,
		Position:     note.Position,
		IsArchived:   note.IsArchived,
		IsDeleted:    note.IsDeleted,
		CreatedAt:    note.CreatedAt,
		UpdatedAt:    note.UpdatedAt,
	}
}

// ToNoteListResponse converts a list of domain notes to a list response
func ToNoteListResponse(notes []*domain.Note, page, limit int, total int64) NoteListResponse {
	noteResponses := make([]NoteResponse, len(notes))
	for i, note := range notes {
		noteResponses[i] = ToNoteResponse(note)
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return NoteListResponse{
		Notes: noteResponses,
		Pagination: PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}
}

// ToNoteSummaryResponse converts a domain note to a summary response
func ToNoteSummaryResponse(note *domain.Note) NoteSummaryResponse {
	return NoteSummaryResponse{
		ID:         note.ID,
		Title:      note.Title,
		Icon:       note.Icon,
		ParentID:   note.ParentID,
		Depth:      note.Depth,
		IsArchived: note.IsArchived,
		CreatedAt:  note.CreatedAt,
		UpdatedAt:  note.UpdatedAt,
	}
}

// ToBreadcrumbResponses converts ancestor notes to breadcrumb trail
func ToBreadcrumbResponses(ancestors []*domain.Note) []BreadcrumbResponse {
	breadcrumbs := make([]BreadcrumbResponse, len(ancestors))
	for i, ancestor := range ancestors {
		breadcrumbs[i] = BreadcrumbResponse{
			ID:    ancestor.ID,
			Title: ancestor.Title,
			Icon:  ancestor.Icon,
		}
	}
	return breadcrumbs
}

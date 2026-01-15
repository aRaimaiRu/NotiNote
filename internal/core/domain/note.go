package domain

import (
	"errors"
	"time"
)

// BlockType represents the type of content block in a note
type BlockType string

const (
	BlockTypeParagraph    BlockType = "paragraph"
	BlockTypeHeading1     BlockType = "heading_1"
	BlockTypeHeading2     BlockType = "heading_2"
	BlockTypeHeading3     BlockType = "heading_3"
	BlockTypeHeading4     BlockType = "heading_4"
	BlockTypeHeading5     BlockType = "heading_5"
	BlockTypeHeading6     BlockType = "heading_6"
	BlockTypeBulletList   BlockType = "bullet_list"
	BlockTypeNumberedList BlockType = "numbered_list"
	BlockTypeCheckbox     BlockType = "checkbox"
	BlockTypeQuote        BlockType = "quote"
	BlockTypeCode         BlockType = "code"
	BlockTypeDivider      BlockType = "divider"
)

// RichTextStyle represents inline text formatting (bold, italic, etc.)
type RichTextStyle struct {
	Bold          bool   `json:"bold,omitempty"`
	Italic        bool   `json:"italic,omitempty"`
	Underline     bool   `json:"underline,omitempty"`
	Strikethrough bool   `json:"strikethrough,omitempty"`
	Code          bool   `json:"code,omitempty"`          // Inline code
	Link          string `json:"link,omitempty"`          // URL for hyperlinks
	Color         string `json:"color,omitempty"`         // Text color
	Background    string `json:"background,omitempty"`    // Background color
}

// RichTextSegment represents a segment of text with optional formatting
type RichTextSegment struct {
	Text  string         `json:"text"`
	Style *RichTextStyle `json:"style,omitempty"`
}

// BlockContent represents the content structure of a block
type BlockContent struct {
	// For text-based blocks (paragraph, heading, quote, list items)
	RichText []RichTextSegment `json:"rich_text,omitempty"`

	// For checkbox blocks
	Checked *bool `json:"checked,omitempty"`

	// For code blocks
	Language string `json:"language,omitempty"` // Programming language for syntax highlighting
	Code     string `json:"code,omitempty"`     // Raw code content

	// For list items with nested children
	Children []Block `json:"children,omitempty"`
}

// Block represents a content block in a note (similar to Notion blocks)
type Block struct {
	ID      string        `json:"id"`      // UUID v4
	Type    BlockType     `json:"type"`
	Content *BlockContent `json:"content"`
	Order   int           `json:"order"`   // Position in the note
}

// ViewType represents different ways to display notes in a database
type ViewType string

const (
	ViewTypeTable   ViewType = "table"
	ViewTypeBoard   ViewType = "board"   // Kanban board
	ViewTypeList    ViewType = "list"
	ViewTypeGallery ViewType = "gallery"
)

// PropertyType represents the data type of custom properties in database views
type PropertyType string

const (
	PropertyTypeText        PropertyType = "text"
	PropertyTypeNumber      PropertyType = "number"
	PropertyTypeSelect      PropertyType = "select"
	PropertyTypeMultiSelect PropertyType = "multi_select"
	PropertyTypeDate        PropertyType = "date"
	PropertyTypeCheckbox    PropertyType = "checkbox"
	PropertyTypeURL         PropertyType = "url"
	PropertyTypeEmail       PropertyType = "email"
	PropertyTypePerson      PropertyType = "person"
)

// ViewProperty defines a column/property in database views
type ViewProperty struct {
	ID       string       `json:"id"`
	Name     string       `json:"name"`
	Type     PropertyType `json:"type"`
	Options  []string     `json:"options,omitempty"`  // For select/multi-select
	Visible  bool         `json:"visible"`
	Width    int          `json:"width,omitempty"`    // Column width in pixels
	Position int          `json:"position"`           // Column order
}

// ViewFilter represents a filter condition in database views
type ViewFilter struct {
	PropertyID string      `json:"property_id"`
	Operator   string      `json:"operator"` // "equals", "contains", "is_empty", "is_not_empty", "greater_than", etc.
	Value      interface{} `json:"value"`
}

// ViewSort represents a sort configuration in database views
type ViewSort struct {
	PropertyID string `json:"property_id"`
	Direction  string `json:"direction"` // "asc" or "desc"
}

// ViewMetadata contains configuration for database views
type ViewMetadata struct {
	ViewType   ViewType       `json:"view_type"`
	Properties []ViewProperty `json:"properties"`
	Filters    []ViewFilter   `json:"filters,omitempty"`
	Sorts      []ViewSort     `json:"sorts,omitempty"`
}

// Tag represents a tag entity for categorizing notes
type Tag struct {
	ID        string    `json:"id"`
	UserID    int64     `json:"user_id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Note represents a note entity in the domain (similar to Notion pages)
type Note struct {
	ID           int64                  `json:"id"`
	UserID       int64                  `json:"user_id"`
	ParentID     *int64                 `json:"parent_id,omitempty"`
	Title        string                 `json:"title"`
	Icon         string                 `json:"icon,omitempty"`
	CoverImage   string                 `json:"cover_image,omitempty"`
	Blocks       []Block                `json:"blocks"`
	ViewMetadata *ViewMetadata          `json:"view_metadata,omitempty"`
	Properties   map[string]interface{} `json:"properties,omitempty"`
	Path         string                 `json:"path"`
	Depth        int                    `json:"depth"`
	Position     int                    `json:"position"`
	IsArchived   bool                   `json:"is_archived"`
	IsDeleted    bool                   `json:"is_deleted"`
	IsFavorite   bool                   `json:"is_favorite"`
	Tags         []Tag                  `json:"tags,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// Domain errors for notes (note-specific errors only, common errors in errors.go)
var (
	// ErrNoteNotFound is defined in errors.go
	// ErrUnauthorizedAccess is defined in errors.go
	ErrInvalidNoteTitle     = errors.New("note title is required and must be between 1 and 500 characters")
	ErrInvalidParentNote    = errors.New("invalid parent note")
	ErrCircularReference    = errors.New("circular reference detected in hierarchy")
	ErrInvalidBlockType     = errors.New("invalid block type")
	ErrInvalidBlockContent  = errors.New("block content is required")
	ErrInvalidBlockOrder    = errors.New("invalid block order")
	ErrMaxDepthExceeded     = errors.New("maximum nesting depth (10 levels) exceeded")
	ErrInvalidBlockID       = errors.New("block ID is required")
	ErrBlockNotFound        = errors.New("block not found")
	ErrInvalidViewType      = errors.New("invalid view type")
)

const (
	MaxNestingDepth  = 10
	MaxTitleLength   = 500
	MinTitleLength   = 1
)

// NewNote creates a new note with validation
func NewNote(userID int64, title string) (*Note, error) {
	if err := ValidateNoteTitle(title); err != nil {
		return nil, err
	}

	now := time.Now()
	return &Note{
		UserID:     userID,
		Title:      title,
		Blocks:     []Block{},
		Properties: make(map[string]interface{}),
		Depth:      0,
		Position:   0,
		IsArchived: false,
		IsDeleted:  false,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

// ValidateNoteTitle validates the note title
func ValidateNoteTitle(title string) error {
	if len(title) < MinTitleLength || len(title) > MaxTitleLength {
		return ErrInvalidNoteTitle
	}
	return nil
}

// SetParent sets the parent note and validates hierarchy
func (n *Note) SetParent(parentID *int64, parentDepth int) error {
	if parentID == nil {
		n.ParentID = nil
		n.Depth = 0
		return nil
	}

	// Prevent self-referencing
	if *parentID == n.ID {
		return ErrCircularReference
	}

	// Check max depth
	newDepth := parentDepth + 1
	if newDepth > MaxNestingDepth {
		return ErrMaxDepthExceeded
	}

	n.ParentID = parentID
	n.Depth = newDepth
	return nil
}

// AddBlock adds a new block to the note
func (n *Note) AddBlock(block Block) error {
	if block.ID == "" {
		return ErrInvalidBlockID
	}

	if block.Type == "" {
		return ErrInvalidBlockType
	}

	// Set order to end of list if not specified
	if block.Order == 0 {
		block.Order = len(n.Blocks)
	}

	n.Blocks = append(n.Blocks, block)
	n.UpdatedAt = time.Now()
	return nil
}

// UpdateBlock updates an existing block by ID
func (n *Note) UpdateBlock(blockID string, content *BlockContent) error {
	if blockID == "" {
		return ErrInvalidBlockID
	}

	for i, block := range n.Blocks {
		if block.ID == blockID {
			n.Blocks[i].Content = content
			n.UpdatedAt = time.Now()
			return nil
		}
	}
	return ErrBlockNotFound
}

// DeleteBlock removes a block from the note by ID
func (n *Note) DeleteBlock(blockID string) error {
	if blockID == "" {
		return ErrInvalidBlockID
	}

	for i, block := range n.Blocks {
		if block.ID == blockID {
			// Remove block at index i
			n.Blocks = append(n.Blocks[:i], n.Blocks[i+1:]...)
			n.UpdatedAt = time.Now()
			return nil
		}
	}
	return ErrBlockNotFound
}

// ReorderBlocks reorders blocks based on a map of block ID to new order
func (n *Note) ReorderBlocks(blockOrders map[string]int) error {
	if len(blockOrders) == 0 {
		return nil
	}

	for i := range n.Blocks {
		if newOrder, exists := blockOrders[n.Blocks[i].ID]; exists {
			if newOrder < 0 {
				return ErrInvalidBlockOrder
			}
			n.Blocks[i].Order = newOrder
		}
	}

	n.UpdatedAt = time.Now()
	return nil
}

// SetBlocks replaces all blocks (used for full content updates)
func (n *Note) SetBlocks(blocks []Block) error {
	// Validate all blocks have IDs and types
	for _, block := range blocks {
		if block.ID == "" {
			return ErrInvalidBlockID
		}
		if block.Type == "" {
			return ErrInvalidBlockType
		}
	}

	n.Blocks = blocks
	n.UpdatedAt = time.Now()
	return nil
}

// Archive archives the note
func (n *Note) Archive() {
	n.IsArchived = true
	n.UpdatedAt = time.Now()
}

// Unarchive unarchives the note
func (n *Note) Unarchive() {
	n.IsArchived = false
	n.UpdatedAt = time.Now()
}

// SoftDelete marks the note as deleted (soft delete)
func (n *Note) SoftDelete() {
	n.IsDeleted = true
	n.UpdatedAt = time.Now()
}

// Restore restores a soft-deleted note
func (n *Note) Restore() {
	n.IsDeleted = false
	n.UpdatedAt = time.Now()
}

// ToggleFavorite toggles the favorite status of a note
func (n *Note) ToggleFavorite() {
	n.IsFavorite = !n.IsFavorite
	n.UpdatedAt = time.Now()
}

// UpdateTitle updates the note title with validation
func (n *Note) UpdateTitle(title string) error {
	if err := ValidateNoteTitle(title); err != nil {
		return err
	}

	n.Title = title
	n.UpdatedAt = time.Now()
	return nil
}

// UpdateIcon updates the note icon
func (n *Note) UpdateIcon(icon string) {
	n.Icon = icon
	n.UpdatedAt = time.Now()
}

// UpdateCoverImage updates the note cover image
func (n *Note) UpdateCoverImage(coverImage string) {
	n.CoverImage = coverImage
	n.UpdatedAt = time.Now()
}

// UpdatePosition updates the note position among siblings
func (n *Note) UpdatePosition(position int) error {
	if position < 0 {
		return errors.New("position must be non-negative")
	}

	n.Position = position
	n.UpdatedAt = time.Now()
	return nil
}

// SetProperty sets a custom property value
func (n *Note) SetProperty(key string, value interface{}) {
	if n.Properties == nil {
		n.Properties = make(map[string]interface{})
	}
	n.Properties[key] = value
	n.UpdatedAt = time.Now()
}

// DeleteProperty removes a custom property
func (n *Note) DeleteProperty(key string) {
	if n.Properties != nil {
		delete(n.Properties, key)
		n.UpdatedAt = time.Now()
	}
}

// SetViewMetadata sets the view configuration for database views
func (n *Note) SetViewMetadata(metadata *ViewMetadata) {
	n.ViewMetadata = metadata
	n.UpdatedAt = time.Now()
}

// IsValidBlockType checks if a block type is valid
func IsValidBlockType(blockType BlockType) bool {
	validTypes := map[BlockType]bool{
		BlockTypeParagraph:    true,
		BlockTypeHeading1:     true,
		BlockTypeHeading2:     true,
		BlockTypeHeading3:     true,
		BlockTypeHeading4:     true,
		BlockTypeHeading5:     true,
		BlockTypeHeading6:     true,
		BlockTypeBulletList:   true,
		BlockTypeNumberedList: true,
		BlockTypeCheckbox:     true,
		BlockTypeQuote:        true,
		BlockTypeCode:         true,
		BlockTypeDivider:      true,
	}
	return validTypes[blockType]
}

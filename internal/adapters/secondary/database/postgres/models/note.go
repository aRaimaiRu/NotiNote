package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/yourusername/notinoteapp/internal/core/domain"
	"gorm.io/gorm"
)

// Note represents the database model for notes
type Note struct {
	ID           int64          `gorm:"primaryKey;autoIncrement"`
	UserID       int64          `gorm:"not null;index:idx_notes_user_id"`
	ParentID     *int64         `gorm:"index:idx_notes_parent_id"`
	Title        string         `gorm:"not null;size:500"`
	Icon         string         `gorm:"size:100"`
	CoverImage   string         `gorm:"size:500"`
	Blocks       BlocksJSON     `gorm:"type:jsonb;not null;default:'[]'"`
	ViewMetadata ViewMetadataJSON `gorm:"type:jsonb"`
	Properties   PropertiesJSON `gorm:"type:jsonb;default:'{}'"`
	Path         string         `gorm:"size:1000;index:idx_notes_path"`
	Depth        int            `gorm:"not null;default:0"`
	Position     int            `gorm:"not null;default:0;index:idx_notes_position"`
	IsArchived   bool           `gorm:"not null;default:false"`
	IsDeleted    bool           `gorm:"not null;default:false"`
	IsFavorite   bool           `gorm:"not null;default:false"`
	CreatedAt    time.Time      `gorm:"autoCreateTime;index:idx_notes_created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

// Custom JSON types for GORM to handle JSONB columns

// BlocksJSON is a custom type for storing blocks as JSONB
type BlocksJSON []domain.Block

// Scan implements the sql.Scanner interface for reading from database
func (b *BlocksJSON) Scan(value interface{}) error {
	if value == nil {
		*b = []domain.Block{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	var blocks []domain.Block
	if err := json.Unmarshal(bytes, &blocks); err != nil {
		return err
	}

	*b = blocks
	return nil
}

// Value implements the driver.Valuer interface for writing to database
func (b BlocksJSON) Value() (driver.Value, error) {
	if len(b) == 0 {
		return "[]", nil
	}
	return json.Marshal(b)
}

// ViewMetadataJSON is a custom type for storing view metadata as JSONB
type ViewMetadataJSON struct {
	Data *domain.ViewMetadata
}

// Scan implements the sql.Scanner interface
func (v *ViewMetadataJSON) Scan(value interface{}) error {
	if value == nil {
		v.Data = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	var metadata domain.ViewMetadata
	if err := json.Unmarshal(bytes, &metadata); err != nil {
		return err
	}

	v.Data = &metadata
	return nil
}

// Value implements the driver.Valuer interface
func (v ViewMetadataJSON) Value() (driver.Value, error) {
	if v.Data == nil {
		return nil, nil
	}
	return json.Marshal(v.Data)
}

// PropertiesJSON is a custom type for storing properties as JSONB
type PropertiesJSON map[string]interface{}

// Scan implements the sql.Scanner interface
func (p *PropertiesJSON) Scan(value interface{}) error {
	if value == nil {
		*p = make(map[string]interface{})
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	var props map[string]interface{}
	if err := json.Unmarshal(bytes, &props); err != nil {
		return err
	}

	*p = props
	return nil
}

// Value implements the driver.Valuer interface
func (p PropertiesJSON) Value() (driver.Value, error) {
	if len(p) == 0 {
		return "{}", nil
	}
	return json.Marshal(p)
}

// TableName specifies the table name for GORM
func (Note) TableName() string {
	return "notes"
}

// ToDomain converts database model to domain entity
func (n *Note) ToDomain() *domain.Note {
	blocks := []domain.Block(n.Blocks)
	if blocks == nil {
		blocks = []domain.Block{}
	}

	props := map[string]interface{}(n.Properties)
	if props == nil {
		props = make(map[string]interface{})
	}

	return &domain.Note{
		ID:           n.ID,
		UserID:       n.UserID,
		ParentID:     n.ParentID,
		Title:        n.Title,
		Icon:         n.Icon,
		CoverImage:   n.CoverImage,
		Blocks:       blocks,
		ViewMetadata: n.ViewMetadata.Data,
		Properties:   props,
		Path:         n.Path,
		Depth:        n.Depth,
		Position:     n.Position,
		IsArchived:   n.IsArchived,
		IsDeleted:    n.IsDeleted,
		IsFavorite:   n.IsFavorite,
		Tags:         []domain.Tag{}, // Tags loaded separately in repository
		CreatedAt:    n.CreatedAt,
		UpdatedAt:    n.UpdatedAt,
	}
}

// FromDomain converts domain entity to database model
func (n *Note) FromDomain(domainNote *domain.Note) {
	n.ID = domainNote.ID
	n.UserID = domainNote.UserID
	n.ParentID = domainNote.ParentID
	n.Title = domainNote.Title
	n.Icon = domainNote.Icon
	n.CoverImage = domainNote.CoverImage
	n.Blocks = BlocksJSON(domainNote.Blocks)
	n.ViewMetadata = ViewMetadataJSON{Data: domainNote.ViewMetadata}
	n.Properties = PropertiesJSON(domainNote.Properties)
	n.Path = domainNote.Path
	n.Depth = domainNote.Depth
	n.Position = domainNote.Position
	n.IsArchived = domainNote.IsArchived
	n.IsDeleted = domainNote.IsDeleted
	n.IsFavorite = domainNote.IsFavorite
	n.CreatedAt = domainNote.CreatedAt
	n.UpdatedAt = domainNote.UpdatedAt
}

// BeforeCreate is a GORM hook that runs before creating a note
func (n *Note) BeforeCreate(tx *gorm.DB) error {
	// Ensure blocks is initialized
	if n.Blocks == nil {
		n.Blocks = BlocksJSON{}
	}

	// Ensure properties is initialized
	if n.Properties == nil {
		n.Properties = PropertiesJSON{}
	}

	return nil
}

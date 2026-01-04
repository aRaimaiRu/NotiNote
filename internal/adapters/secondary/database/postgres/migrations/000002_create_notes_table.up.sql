-- Create block type enum for Notion-like content blocks
CREATE TYPE block_type AS ENUM (
    'paragraph',
    'heading_1',
    'heading_2',
    'heading_3',
    'heading_4',
    'heading_5',
    'heading_6',
    'bullet_list',
    'numbered_list',
    'checkbox',
    'quote',
    'code',
    'divider'
);

-- Create notes table with JSONB support for flexible block-based content
CREATE TABLE notes (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    parent_id BIGINT,
    title VARCHAR(500) NOT NULL,
    icon VARCHAR(100),
    cover_image VARCHAR(500),

    -- Content stored as JSONB array of blocks
    -- Each block: {"id": "uuid", "type": "paragraph", "content": {...}, "order": 0}
    -- Supports rich text with inline formatting: bold, italic, underline, strikethrough, links, code
    blocks JSONB NOT NULL DEFAULT '[]'::jsonb,

    -- Metadata for database views (table, board, list, gallery)
    -- {"view_type": "table", "properties": [...], "filters": [...], "sorts": [...]}
    view_metadata JSONB,

    -- Properties for database items (when note is used in a database view)
    -- {"status": "In Progress", "priority": "High", "tags": ["work"], "date": "2024-01-15"}
    properties JSONB DEFAULT '{}'::jsonb,

    -- Hierarchy path for efficient querying (materialized path pattern)
    -- Format: "/1/23/456/" where numbers are ancestor IDs
    -- Enables efficient descendant queries: WHERE path LIKE '/1/23/%'
    path VARCHAR(1000),

    -- Depth in hierarchy (0 = root level, max 10 levels)
    depth INTEGER NOT NULL DEFAULT 0,

    -- Position among siblings for ordering
    position INTEGER NOT NULL DEFAULT 0,

    -- Soft delete support
    is_archived BOOLEAN NOT NULL DEFAULT false,
    is_deleted BOOLEAN NOT NULL DEFAULT false,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_id) REFERENCES notes(id) ON DELETE CASCADE
);

-- Create indexes for performance optimization

-- User notes lookup (most common query)
CREATE INDEX idx_notes_user_id ON notes(user_id) WHERE is_deleted = false;

-- Hierarchy queries (parent-child relationships)
CREATE INDEX idx_notes_parent_id ON notes(parent_id) WHERE is_deleted = false;

-- Efficient descendant/ancestor queries using materialized path
CREATE INDEX idx_notes_path ON notes USING btree(path) WHERE is_deleted = false;

-- Sorting by creation time
CREATE INDEX idx_notes_created_at ON notes(created_at DESC);

-- Sorting siblings by position
CREATE INDEX idx_notes_position ON notes(parent_id, position);

-- GIN indexes for JSONB searches (enables efficient property filtering)
CREATE INDEX idx_notes_blocks ON notes USING GIN(blocks);
CREATE INDEX idx_notes_properties ON notes USING GIN(properties);

-- Full-text search on title
CREATE INDEX idx_notes_title_search ON notes USING GIN(to_tsvector('english', title));

-- Soft delete lookup
CREATE INDEX idx_notes_deleted_at ON notes(deleted_at);

-- Archived notes filter
CREATE INDEX idx_notes_archived ON notes(user_id, is_archived) WHERE is_deleted = false;

-- Create trigger to automatically update updated_at timestamp
-- Note: This reuses the function created in 000001_create_users_table.up.sql
CREATE TRIGGER update_notes_updated_at
    BEFORE UPDATE ON notes
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create function to automatically maintain hierarchy path
-- This trigger updates path and depth when parent_id changes
CREATE OR REPLACE FUNCTION update_note_hierarchy()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.parent_id IS NULL THEN
        -- Root level note
        NEW.path = '/' || NEW.id || '/';
        NEW.depth = 0;
    ELSE
        -- Child note: concatenate parent path with current ID
        SELECT path || NEW.id || '/', depth + 1
        INTO NEW.path, NEW.depth
        FROM notes
        WHERE id = NEW.parent_id;

        -- Prevent exceeding max depth (10 levels)
        IF NEW.depth > 10 THEN
            RAISE EXCEPTION 'Maximum nesting depth (10 levels) exceeded';
        END IF;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER maintain_note_hierarchy
    BEFORE INSERT OR UPDATE OF parent_id ON notes
    FOR EACH ROW
    EXECUTE FUNCTION update_note_hierarchy();

-- Add documentation comments for clarity
COMMENT ON TABLE notes IS 'Stores notes with block-based content similar to Notion, supporting rich text formatting and hierarchical organization';
COMMENT ON COLUMN notes.id IS 'Unique note identifier';
COMMENT ON COLUMN notes.user_id IS 'Owner of the note';
COMMENT ON COLUMN notes.parent_id IS 'Parent note ID for hierarchical organization (NULL for root notes)';
COMMENT ON COLUMN notes.title IS 'Note title (required, max 500 characters)';
COMMENT ON COLUMN notes.icon IS 'Optional emoji or icon identifier';
COMMENT ON COLUMN notes.cover_image IS 'Optional cover image URL';
COMMENT ON COLUMN notes.blocks IS 'JSONB array of content blocks with rich text formatting';
COMMENT ON COLUMN notes.view_metadata IS 'Configuration for database views (table, board, list, gallery)';
COMMENT ON COLUMN notes.properties IS 'Custom properties when note is used as database item (status, priority, tags, dates, etc.)';
COMMENT ON COLUMN notes.path IS 'Materialized path for efficient hierarchy queries (format: /1/23/456/)';
COMMENT ON COLUMN notes.depth IS 'Depth in hierarchy (0 = root, max 10)';
COMMENT ON COLUMN notes.position IS 'Position among siblings for ordering';
COMMENT ON COLUMN notes.is_archived IS 'Whether note is archived (hidden from default views)';
COMMENT ON COLUMN notes.is_deleted IS 'Soft delete flag (allows recovery)';
COMMENT ON COLUMN notes.created_at IS 'Note creation timestamp';
COMMENT ON COLUMN notes.updated_at IS 'Last modification timestamp';
COMMENT ON COLUMN notes.deleted_at IS 'Soft delete timestamp';

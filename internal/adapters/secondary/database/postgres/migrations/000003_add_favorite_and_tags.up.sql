-- Add is_favorite column to notes table
ALTER TABLE notes ADD COLUMN is_favorite BOOLEAN NOT NULL DEFAULT false;

-- Create index for favorite filtering
CREATE INDEX idx_notes_favorite ON notes(user_id, is_favorite) WHERE is_deleted = false;

-- Create tags table
CREATE TABLE tags (
    id VARCHAR(100) PRIMARY KEY, -- UUID or slug-based ID (e.g., "tag-work", "tag-personal")
    user_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    color VARCHAR(50) NOT NULL DEFAULT 'gray', -- Color identifier (e.g., "blue", "red", "green")
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(user_id, name) -- Each user can only have one tag with a given name
);

-- Create note_tags junction table for many-to-many relationship
CREATE TABLE note_tags (
    note_id BIGINT NOT NULL,
    tag_id VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (note_id, tag_id),
    FOREIGN KEY (note_id) REFERENCES notes(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

-- Create indexes for tags
CREATE INDEX idx_tags_user_id ON tags(user_id);
CREATE INDEX idx_note_tags_note_id ON note_tags(note_id);
CREATE INDEX idx_note_tags_tag_id ON note_tags(tag_id);

-- Create trigger to automatically update tags.updated_at timestamp
CREATE TRIGGER update_tags_updated_at
    BEFORE UPDATE ON tags
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments for clarity
COMMENT ON COLUMN notes.is_favorite IS 'Whether note is marked as favorite/starred';
COMMENT ON TABLE tags IS 'User-defined tags for categorizing notes';
COMMENT ON COLUMN tags.id IS 'Unique tag identifier (UUID or slug)';
COMMENT ON COLUMN tags.name IS 'Display name of the tag';
COMMENT ON COLUMN tags.color IS 'Color identifier for UI display (e.g., blue, red, green)';
COMMENT ON TABLE note_tags IS 'Many-to-many relationship between notes and tags';

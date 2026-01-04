-- Drop triggers
DROP TRIGGER IF EXISTS maintain_note_hierarchy ON notes;
DROP TRIGGER IF EXISTS update_notes_updated_at ON notes;

-- Drop functions
DROP FUNCTION IF EXISTS update_note_hierarchy();

-- Drop indexes
DROP INDEX IF EXISTS idx_notes_user_id;
DROP INDEX IF EXISTS idx_notes_parent_id;
DROP INDEX IF EXISTS idx_notes_path;
DROP INDEX IF EXISTS idx_notes_created_at;
DROP INDEX IF EXISTS idx_notes_position;
DROP INDEX IF EXISTS idx_notes_blocks;
DROP INDEX IF EXISTS idx_notes_properties;
DROP INDEX IF EXISTS idx_notes_title_search;
DROP INDEX IF EXISTS idx_notes_deleted_at;
DROP INDEX IF EXISTS idx_notes_archived;

-- Drop notes table
DROP TABLE IF EXISTS notes;

-- Drop custom type
DROP TYPE IF EXISTS block_type;

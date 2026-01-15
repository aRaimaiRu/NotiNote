-- Drop note_tags junction table
DROP TABLE IF EXISTS note_tags;

-- Drop tags table
DROP TABLE IF EXISTS tags;

-- Drop favorite index
DROP INDEX IF EXISTS idx_notes_favorite;

-- Remove is_favorite column from notes
ALTER TABLE notes DROP COLUMN IF EXISTS is_favorite;

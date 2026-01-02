-- Drop trigger
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_provider_id;
DROP INDEX IF EXISTS idx_users_deleted_at;

-- Drop users table
DROP TABLE IF EXISTS users;

-- Drop custom type
DROP TYPE IF EXISTS auth_provider;

-- Create custom type for auth provider
CREATE TYPE auth_provider AS ENUM ('email', 'google', 'facebook');

-- Create users table
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255),
    provider auth_provider NOT NULL DEFAULT 'email',
    provider_id VARCHAR(255),
    avatar_url VARCHAR(500),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes
CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_provider_id ON users(provider, provider_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

-- Create trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE users IS 'Stores user account information';
COMMENT ON COLUMN users.id IS 'Unique user identifier';
COMMENT ON COLUMN users.email IS 'User email address (unique)';
COMMENT ON COLUMN users.name IS 'User display name';
COMMENT ON COLUMN users.password_hash IS 'Bcrypt hashed password (null for OAuth users)';
COMMENT ON COLUMN users.provider IS 'Authentication provider (email, google, facebook)';
COMMENT ON COLUMN users.provider_id IS 'OAuth provider user ID (null for email users)';
COMMENT ON COLUMN users.avatar_url IS 'User profile picture URL';
COMMENT ON COLUMN users.is_active IS 'Whether user account is active';
COMMENT ON COLUMN users.created_at IS 'Account creation timestamp';
COMMENT ON COLUMN users.updated_at IS 'Last update timestamp';
COMMENT ON COLUMN users.deleted_at IS 'Soft delete timestamp';

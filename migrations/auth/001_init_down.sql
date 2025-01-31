-- Drop triggers first
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop functions
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_tokens_user_id;

-- Drop tables (order matters due to foreign key constraints)
DROP TABLE IF EXISTS tokens;
DROP TABLE IF EXISTS users;

-- Drop trigger first
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_tokens_user_id;
DROP INDEX IF EXISTS idx_tokens_token;

-- Drop tables in correct order (child tables first)
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS users;

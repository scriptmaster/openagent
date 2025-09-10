-- 002_users.sql: Create users table
-- Users table
CREATE TABLE IF NOT EXISTS ai.users (
    id SERIAL PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    is_admin BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_logged_in TIMESTAMP WITH TIME ZONE
);

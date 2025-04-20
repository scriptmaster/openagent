-- Add password_hash column to users table
ALTER TABLE ai.users
ADD COLUMN IF NOT EXISTS password_hash TEXT NOT NULL DEFAULT ''; -- Add temporary default

-- Remove the temporary default constraint
ALTER TABLE ai.users
ALTER COLUMN password_hash DROP DEFAULT; 
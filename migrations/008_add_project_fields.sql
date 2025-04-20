-- Add missing fields to ai.projects table
ALTER TABLE ai.projects
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE;

-- Set updated_at for existing rows to created_at initially
UPDATE ai.projects SET updated_at = created_at WHERE updated_at IS NULL; 
-- 007_settings.sql: Create settings table with corrected UNIQUE constraint
-- Settings table
CREATE TABLE IF NOT EXISTS ai.settings (
    id SERIAL PRIMARY KEY,
    key TEXT NOT NULL,
    value TEXT,
    description TEXT,
    scope TEXT NOT NULL CHECK (scope IN ('system', 'project', 'user')),
    scope_id INTEGER, -- NULL for system, project_id for project, user_id for user
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(key, scope)
);

-- Initial settings
INSERT INTO ai.settings (key, value, description, scope) 
VALUES ('app_name', 'Data Manager', 'Application name', 'system')
ON CONFLICT DO NOTHING; 
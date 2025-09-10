-- 004_project_dbs.sql: Create project database connections table
-- Project database connections
CREATE TABLE IF NOT EXISTS ai.project_dbs (
    id SERIAL PRIMARY KEY,
    project_id INTEGER REFERENCES ai.projects(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    db_type TEXT NOT NULL, -- postgresql, mysql, etc.
    connection_string TEXT NOT NULL, -- base64 encoded
    schema_name TEXT NOT NULL DEFAULT 'public',
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(project_id, name)
);

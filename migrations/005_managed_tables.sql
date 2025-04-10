-- 005_managed_tables.sql: Create managed tables metadata
-- Tables metadata
CREATE TABLE IF NOT EXISTS ai.managed_tables (
    id SERIAL PRIMARY KEY,
    project_id INTEGER REFERENCES ai.projects(id) ON DELETE CASCADE,
    project_db_id INTEGER REFERENCES ai.project_dbs(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    schema_name TEXT NOT NULL DEFAULT 'public',
    description TEXT,
    initialized BOOLEAN NOT NULL DEFAULT FALSE,
    read_only BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(project_db_id, schema_name, name)
); 
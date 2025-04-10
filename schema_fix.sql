-- Create schema if not exists
CREATE SCHEMA IF NOT EXISTS ai;

-- Drop all tables in the correct order to avoid dependency issues
DROP TABLE IF EXISTS ai.managed_columns;
DROP TABLE IF EXISTS ai.managed_tables;
DROP TABLE IF EXISTS ai.project_dbs;
DROP TABLE IF EXISTS ai.projects;
DROP TABLE IF EXISTS ai.settings;
DROP TABLE IF EXISTS ai.users;

-- Users table
CREATE TABLE IF NOT EXISTS ai.users (
    id SERIAL PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    is_admin BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_logged_in TIMESTAMP WITH TIME ZONE
);

-- Projects table
CREATE TABLE IF NOT EXISTS ai.projects (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    domain_name TEXT,
    created_by INTEGER REFERENCES ai.users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

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

-- Columns metadata (helpful for UI display customization)
CREATE TABLE IF NOT EXISTS ai.managed_columns (
    id SERIAL PRIMARY KEY,
    table_id INTEGER REFERENCES ai.managed_tables(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    display_name TEXT,
    type TEXT NOT NULL,
    ordinal INTEGER NOT NULL,
    visible BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(table_id, name)
);

-- Settings - with fixed UNIQUE constraint
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
VALUES ('app_name', 'Data Manager', 'Application name', 'system'); 
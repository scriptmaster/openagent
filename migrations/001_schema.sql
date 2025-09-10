-- 001_schema.sql: Create the basic schema
-- Create schema if not exists
CREATE SCHEMA IF NOT EXISTS ai;

CREATE TABLE ai.migrations (
    id SERIAL PRIMARY KEY,
    filename VARCHAR(255) NOT NULL UNIQUE,
    applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

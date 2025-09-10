-- 003_projects.sql: Create projects table
-- Projects table
CREATE TABLE IF NOT EXISTS ai.projects (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    domain_name TEXT,
    options TEXT,
    created_by INTEGER REFERENCES ai.users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

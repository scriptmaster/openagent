-- 011_create_pages_table.sql: Create pages table for project-specific content
CREATE TABLE IF NOT EXISTS ai.pages (
    id SERIAL PRIMARY KEY,
    project_id INTEGER REFERENCES ai.projects(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    slug TEXT NOT NULL,
    html_content TEXT,
    is_landing BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    meta_title TEXT,
    meta_description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(project_id, slug)
);

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_pages_project_landing ON ai.pages(project_id, is_landing) WHERE is_landing = TRUE;
CREATE INDEX IF NOT EXISTS idx_pages_project_slug ON ai.pages(project_id, slug);

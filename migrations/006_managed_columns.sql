-- 006_managed_columns.sql: Create managed columns metadata
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
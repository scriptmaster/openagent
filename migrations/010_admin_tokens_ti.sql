-- 010_admin_tokens_ti.sql: Create admin_tokens table and index (ti = table + index)
-- Change token column from UUID to VARCHAR to support UUID-ADMIN-TOKEN-SHA1-HASHED-TIMESTAMP format

-- First, drop the existing table and recreate it with the new format
DROP TABLE IF EXISTS ai.admin_tokens;

CREATE TABLE IF NOT EXISTS ai.admin_tokens (
    token VARCHAR(255) PRIMARY KEY,
    token_date DATE NOT NULL UNIQUE, -- Ensure only one token per date
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_admin_tokens_date ON ai.admin_tokens (token_date);

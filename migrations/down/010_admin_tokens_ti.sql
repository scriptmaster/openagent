-- Revert the admin_tokens table back to UUID format (ti = table + index)
DROP TABLE IF EXISTS ai.admin_tokens;

CREATE TABLE IF NOT EXISTS ai.admin_tokens (
    token UUID PRIMARY KEY,
    token_date DATE NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_admin_tokens_date ON ai.admin_tokens (token_date);

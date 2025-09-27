-- name: settings/create
INSERT INTO ai.settings (key, value, description, scope, scope_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, key, value, description, scope, scope_id, updated_at

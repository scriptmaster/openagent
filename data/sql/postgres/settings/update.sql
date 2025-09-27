-- name: settings/update
UPDATE ai.settings 
SET value = $2, description = $3, updated_at = NOW()
WHERE id = $1
RETURNING id, key, value, description, scope, scope_id, updated_at

-- name: settings/upsert_no_scope_id
INSERT INTO ai.settings (key, value, scope, scope_id, updated_at)
VALUES ($1, $2, $3, NULL, NOW())
ON CONFLICT (key, scope, COALESCE(scope_id, 0))
DO UPDATE SET value = $2, updated_at = NOW()
RETURNING id, key, value, description, scope, scope_id, updated_at

-- name: settings/upsert_with_scope_id
INSERT INTO ai.settings (key, value, scope, scope_id, updated_at)
VALUES ($1, $2, $3, $4, NOW())
ON CONFLICT (key, scope, COALESCE(scope_id, 0))
DO UPDATE SET value = $2, updated_at = NOW()
RETURNING id, key, value, description, scope, scope_id, updated_at

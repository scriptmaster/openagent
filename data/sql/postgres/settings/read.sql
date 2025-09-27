-- name: settings/read_by_id
SELECT id, key, value, description, scope, scope_id, updated_at
FROM ai.settings
WHERE id = $1

-- name: settings/read_by_key_and_scope_no_id
SELECT id, key, value, description, scope, scope_id, updated_at
FROM ai.settings
WHERE key = $1 AND scope = $2 AND scope_id IS NULL

-- name: settings/read_by_key_and_scope_with_id
SELECT id, key, value, description, scope, scope_id, updated_at
FROM ai.settings
WHERE key = $1 AND scope = $2 AND scope_id = $3

-- name: settings/get_scoped
SELECT id, key, value, description, scope, scope_id, updated_at
FROM ai.settings
WHERE key = $1 AND scope = $2 AND scope_id = $3 
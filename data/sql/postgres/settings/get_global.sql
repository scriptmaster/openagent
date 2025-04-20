-- name: settings/get_global
SELECT id, key, value, description, scope, scope_id, updated_at
FROM ai.settings
WHERE key = $1 AND scope = 'global'
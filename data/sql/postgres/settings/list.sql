-- name: settings/list_all
SELECT id, key, value, description, scope, scope_id, updated_at
FROM ai.settings
ORDER BY key

-- name: settings/list_by_scope
SELECT id, key, value, description, scope, scope_id, updated_at
FROM ai.settings
WHERE scope = $1
ORDER BY key

-- name: settings/list_by_scope_and_id
SELECT id, key, value, description, scope, scope_id, updated_at
FROM ai.settings
WHERE scope = $1 AND scope_id = $2
ORDER BY key

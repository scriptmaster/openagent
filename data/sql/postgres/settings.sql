-- name: GetSettingScopedNoID
SELECT id, key, value, description, scope, scope_id, updated_at
FROM ai.settings
WHERE key = $1 AND scope = $2 AND scope_id IS NULL;

-- name: GetSettingScopedWithID
SELECT id, key, value, description, scope, scope_id, updated_at
FROM ai.settings
WHERE key = $1 AND scope = $2 AND scope_id = $3;

-- name: UpsertSettingNoScopeID
INSERT INTO ai.settings (key, value, scope, scope_id, updated_at)
VALUES ($1, $2, $3, NULL, NOW())
ON CONFLICT (key, scope, COALESCE(scope_id, 0))
DO UPDATE SET value = $2, updated_at = NOW();

-- name: UpsertSettingWithScopeID
INSERT INTO ai.settings (key, value, scope, scope_id, updated_at)
VALUES ($1, $2, $3, $4, NOW())
ON CONFLICT (key, scope, COALESCE(scope_id, 0))
DO UPDATE SET value = $2, updated_at = NOW();

-- name: GetAllSettings
SELECT id, key, value, description, scope, scope_id, updated_at
FROM ai.settings
ORDER BY key; 
-- name: settings/delete
DELETE FROM ai.settings WHERE id = $1

-- name: settings/delete_by_key_and_scope
DELETE FROM ai.settings 
WHERE key = $1 AND scope = $2 AND COALESCE(scope_id, 0) = COALESCE($3, 0)

-- name: projects/list
SELECT id, name, description, domain_name, options, created_at, updated_at, created_by, is_active FROM ai.projects ORDER BY created_at DESC 
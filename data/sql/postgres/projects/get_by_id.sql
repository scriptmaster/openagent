-- name: projects/get_by_id
SELECT id, name, description, domain_name, options, created_at, updated_at, created_by, is_active FROM ai.projects WHERE id = $1 
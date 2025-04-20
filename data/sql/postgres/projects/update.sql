-- name: projects/update
UPDATE ai.projects
SET name = $1, description = $2, domain_name = $3, is_active = $4, updated_at = $5, options = $6
WHERE id = $7 
-- name: projects/create
INSERT INTO ai.projects (name, description, domain_name, created_by, is_active, created_at, updated_at, options)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id 
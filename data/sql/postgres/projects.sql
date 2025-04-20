-- name: CreateProjectLegacy -- Renamed to avoid conflict if projects package has its own
INSERT INTO ai.projects (name, description, domain_name, created_by, created_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, name, description, domain_name, created_by, created_at;

-- name: ListProjectsLegacy
SELECT id, name, description, domain_name, created_by, created_at
FROM ai.projects
ORDER BY created_at DESC; 
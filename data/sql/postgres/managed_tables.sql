-- name: GetManagedTablesByProject
SELECT id, project_id, project_db_id, name, schema_name, description, initialized, read_only, created_at
FROM ai.managed_tables
WHERE project_id = $1 AND project_db_id = $2;

-- name: InsertManagedTable
INSERT INTO ai.managed_tables (project_id, project_db_id, name, schema_name, description, initialized, read_only)
VALUES ($1, $2, $3, $4, $5, true, $6)
RETURNING id, project_id, project_db_id, name, schema_name, description, initialized, read_only, created_at;

-- name: UpdateManagedTableStatus
UPDATE ai.managed_tables
SET initialized = $2, read_only = $3
WHERE id = $1;

-- name: GetManagedTablesByProjectDBID
SELECT id, project_id, project_db_id, name, schema_name, description, initialized, read_only, created_at
FROM ai.managed_tables
WHERE project_db_id = $1
ORDER BY name; 
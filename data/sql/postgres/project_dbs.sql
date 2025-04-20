-- name: GetProjectDBByID
SELECT id, project_id, name, description, db_type, connection_string, schema_name, is_default, created_at
FROM ai.project_dbs
WHERE id = $1;

-- name: UnsetProjectDBDefault
UPDATE ai.project_dbs
SET is_default = false
WHERE project_id = $1 AND id != $2;

-- name: InsertProjectDB
INSERT INTO ai.project_dbs (project_id, name, description, db_type, connection_string, schema_name, is_default)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, project_id, name, description, db_type, connection_string, schema_name, is_default, created_at;

-- name: GetProjectDBsByProjectID
SELECT id, project_id, name, description, db_type, connection_string, schema_name, is_default, created_at
FROM ai.project_dbs
WHERE project_id = $1
ORDER BY name;

-- name: UpdateProjectDB
UPDATE ai.project_dbs
SET name = $1, description = $2, db_type = $3, connection_string = $4, schema_name = $5, is_default = $6
WHERE id = $7;

-- name: DeleteProjectDB
DELETE FROM ai.project_dbs
WHERE id = $1; 
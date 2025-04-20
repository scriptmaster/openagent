-- name: GetManagedColumnsByTable
SELECT id, name, display_name, data_type, ordinal, visible
FROM ai.managed_columns
WHERE managed_table_id = $1
ORDER BY ordinal;

-- name: InsertManagedColumn
INSERT INTO ai.managed_columns (table_id, name, display_name, type, ordinal, visible)
VALUES ($1, $2, $2, $3, $4, true);

-- name: UpdateManagedColumnVisibility
UPDATE ai.managed_columns
SET display_name = $2, visible = $3
WHERE id = $1; 
-- name: CheckSchemaExists
SELECT EXISTS(SELECT 1 FROM information_schema.schemata WHERE schema_name = 'ai');

-- name: CreateDatabase
CREATE DATABASE "%s"; -- Placeholder for dynamic DB name, handled in Go code

-- name: ListDatabaseTables
SELECT table_name
FROM information_schema.tables
WHERE table_schema = $1
AND table_type = 'BASE TABLE'
ORDER BY table_name;

-- name: ListSchemas
SELECT schema_name
FROM information_schema.schemata
WHERE schema_name NOT IN ('information_schema', 'pg_catalog', 'pg_toast')
ORDER BY schema_name;

-- name: GetTableColumns
SELECT
    column_name,
    data_type,
    is_nullable = 'YES' as is_nullable,
    ordinal_position,
    column_default
FROM
    information_schema.columns
WHERE
    table_schema = $1
    AND table_name = $2
ORDER BY
    ordinal_position; 
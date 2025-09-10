-- name: CheckMigrationsTableExists
SELECT EXISTS(
    SELECT 1 FROM information_schema.tables 
    WHERE table_schema = 'ai' 
    AND table_name = 'migrations'
);

-- name: GetLastAppliedMigration
SELECT COALESCE(MAX(CAST(SUBSTRING(filename FROM '^(\d+)') AS INTEGER)), 0) as last_migration
FROM ai.migrations
WHERE filename ~ '^\d+_.*\.sql$';

-- name: GetAppliedMigrations
SELECT filename, applied_at
FROM ai.migrations
WHERE filename ~ '^\d+_.*\.sql$'
ORDER BY CAST(SUBSTRING(filename FROM '^(\d+)') AS INTEGER);

-- name: InsertMigration
INSERT INTO ai.migrations (filename, applied_at)
VALUES ($1, NOW())
RETURNING id;

-- name: CheckMigrationExists
SELECT EXISTS(
    SELECT 1 FROM ai.migrations 
    WHERE filename = $1
);

-- name: GetMigrationCount
SELECT COUNT(*) as count
FROM ai.migrations
WHERE filename ~ '^\d+_.*\.sql$';

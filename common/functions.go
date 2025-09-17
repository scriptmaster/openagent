package common

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

// SQLQueries holds all SQL queries for different database types
type SQLQueries struct {
	Postgres map[string]string `yaml:"postgres"`
	SQLite   map[string]string `yaml:"sqlite"`
	MySQL    map[string]string `yaml:"mysql"`
	MSSQL    map[string]string `yaml:"mssql"`
}

var queries SQLQueries

// LoadSQLQueries loads SQL queries from the YAML file
func LoadSQLQueries(filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read SQL queries file: %w", err)
	}

	if err := yaml.Unmarshal(data, &queries); err != nil {
		return fmt.Errorf("failed to parse SQL queries: %w", err)
	}

	return nil
}

// GetQuery returns the appropriate SQL query for the current database driver
func GetQuery(db *sql.DB, queryName string) (string, error) {
	// Get driver name
	driverName := getDriverName(db)

	// Get queries for the current driver
	var driverQueries map[string]string
	switch driverName {
	case "postgres", "pgx":
		driverQueries = queries.Postgres
	case "sqlite3", "sqlite":
		driverQueries = queries.SQLite
	case "mysql":
		driverQueries = queries.MySQL
	case "mssql":
		driverQueries = queries.MSSQL
	default:
		return "", fmt.Errorf("unsupported database driver: %s", driverName)
	}

	// Get the query
	query, ok := driverQueries[queryName]
	if !ok {
		return "", fmt.Errorf("query not found: %s", queryName)
	}

	return query, nil
}

// GetEnvOrDefault looks up an environment variable or returns a fallback.
func GetEnvOrDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	log.Printf("Using default for env var %s: %s", key, fallback)
	return fallback
}

func GetDefaultFallback(key string) string {
	switch key {
	case "APP_VERSION":
		return "1.0.0.0"
	case "SYSADMIN_EMAIL":
		return "admin@example.com"
	default:
		return ""
	}
}

var (
	// envCache stores environment variables after their first lookup for faster access.
	envCache = make(map[string]string)
	// envCacheMutex protects envCache from concurrent access.
	envCacheMutex = &sync.RWMutex{}
)

// GetEnv looks up an environment variable, first checking a local cache, then os.LookupEnv,
// and finally falling back to a default. The result is cached for subsequent calls.
func GetEnv(key string) string {
	// Try to get from cache first (read lock)
	envCacheMutex.RLock()
	if value, ok := envCache[key]; ok {
		envCacheMutex.RUnlock()
		return value
	}
	envCacheMutex.RUnlock()

	// Not in cache, acquire write lock to fetch and store
	envCacheMutex.Lock()
	defer envCacheMutex.Unlock()

	// Double-check cache in case another goroutine populated it while we were waiting for the write lock
	if value, ok := envCache[key]; ok {
		return value
	}

	// Look up from environment
	if value, ok := os.LookupEnv(key); ok {
		envCache[key] = value
		return value
	}

	// If not found in environment, use fallback
	fallback := GetDefaultFallback(key)
	log.Printf("Using default for env var %s: %s", key, fallback)
	envCache[key] = fallback
	return fallback
}

// getDriverName returns the normalized driver name from a database connection
func getDriverName(db *sql.DB) string {
	return strings.ToLower(fmt.Sprintf("%T", db.Driver()))
}

// GetTimestampFunction returns the appropriate timestamp function for the current database driver
func GetTimestampFunction(db *sql.DB) string {
	driverName := getDriverName(db)
	switch driverName {
	case "postgres", "pgx":
		return "NOW()"
	case "sqlite3", "sqlite":
		return "CURRENT_TIMESTAMP"
	case "mysql":
		return "NOW()"
	case "mssql":
		return "GETDATE()"
	default:
		log.Printf("Warning: Unknown database driver %s, using NOW() as default", driverName)
		return "NOW()"
	}
}

// GetParameterPlaceholder returns the appropriate parameter placeholder for the current database driver
func GetParameterPlaceholder(db *sql.DB, index int) string {
	driverName := getDriverName(db)
	switch driverName {
	case "postgres", "pgx":
		return fmt.Sprintf("$%d", index)
	case "sqlite3", "sqlite", "mysql":
		return "?"
	case "mssql":
		return fmt.Sprintf("@p%d", index)
	default:
		log.Printf("Warning: Unknown database driver %s, using ? as default", driverName)
		return "?"
	}
}

// GetSchemaPrefix returns the appropriate schema prefix for the current database driver
func GetSchemaPrefix(db *sql.DB) string {
	driverName := getDriverName(db)
	switch driverName {
	case "postgres", "pgx":
		return "ai."
	case "sqlite3", "sqlite", "mysql", "mssql":
		return ""
	default:
		log.Printf("Warning: Unknown database driver %s, using empty schema prefix", driverName)
		return ""
	}
}

// GetReturningClause returns the appropriate RETURNING clause for the current database driver
func GetReturningClause(db *sql.DB) string {
	driverName := getDriverName(db)
	switch driverName {
	case "postgres", "pgx":
		return "RETURNING"
	case "sqlite3", "sqlite", "mysql", "mssql":
		return ""
	default:
		log.Printf("Warning: Unknown database driver %s, using empty RETURNING clause", driverName)
		return ""
	}
}

// GetLastAppliedMigration returns the highest migration number that has been applied
func GetLastAppliedMigration(db *sql.DB) (int, error) {
	// First, check if the ai.migrations table exists.
	// This is crucial because the first migration (001_schema.sql) creates this table.
	var tableExists bool
	checkTableQuery := MustGetSQL("CheckMigrationsTableExists")
	err := db.QueryRow(checkTableQuery).Scan(&tableExists)
	if err != nil {
		// If there's an error checking table existence (e.g., database connection issue,
		// permissions problem), it's a critical error that should be propagated.
		return 0, fmt.Errorf("failed to check if migrations table exists: %w", err)
	}

	if !tableExists {
		// If the ai.migrations table does not exist, it implies that no migrations
		// have been applied yet. The system is in its initial state.
		return 0, nil
	}

	// If the table exists, proceed to query for the highest applied migration number.
	var lastApplied int
	getLastMigrationQuery := MustGetSQL("GetLastAppliedMigration")
	err = db.QueryRow(getLastMigrationQuery).Scan(&lastApplied)
	if err != nil {
		// The 'GetLastAppliedMigration' SQL query uses COALESCE(MAX(...), 0),
		// so it should return 0 if the table is empty, rather than sql.ErrNoRows.
		// Any error here would indicate a more serious issue (e.g., malformed query,
		// database connection problem after initial check).
		return 0, fmt.Errorf("failed to get last applied migration from existing table: %w", err)
	}
	return lastApplied, nil
}

// GetAppliedMigrations returns a list of all applied migrations
func GetAppliedMigrations(db *sql.DB) ([]struct {
	Filename  string
	AppliedAt time.Time
}, error) {
	getAppliedMigrationsQuery := MustGetSQL("GetAppliedMigrations")
	rows, err := db.Query(getAppliedMigrationsQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}
	defer rows.Close()

	var migrations []struct {
		Filename  string
		AppliedAt time.Time
	}

	for rows.Next() {
		var migration struct {
			Filename  string
			AppliedAt time.Time
		}
		if err := rows.Scan(&migration.Filename, &migration.AppliedAt); err != nil {
			return nil, fmt.Errorf("failed to scan migration row: %w", err)
		}
		migrations = append(migrations, migration)
	}

	return migrations, rows.Err()
}

// ResetMigrationTracking removes all migration records from ai.migrations table
// This should be used with caution as it will cause all migrations to be re-applied
func ResetMigrationTracking(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM ai.migrations WHERE filename ~ '^\\d+_.*\\.sql$'")
	if err != nil {
		return fmt.Errorf("failed to reset migration tracking: %w", err)
	}
	return nil
}

// HTML minification regex patterns
var (
	aggressiveWhitespacePattern = regexp.MustCompile(`>\s*\n\s*<`)
	whitespaceAfterTagPattern   = regexp.MustCompile(`>\s*\n\s*`)
	whitespaceBeforeTagPattern  = regexp.MustCompile(`\s*\n\s*<`)
	remainingNewlinesPattern    = regexp.MustCompile(`(\n\s*)+`)
)

// MinifyHTML performs aggressive HTML minification by removing whitespace
// Set preserveWhitespace=true to skip minification (useful for hydration)
func MinifyHTML(content string, preserveWhitespace bool) string {
	if preserveWhitespace {
		return content
	}

	// Aggressive HTML Minification - remove ALL newlines and spaces after > and before <
	// 1. Remove all whitespace (including newlines) between tags
	content = aggressiveWhitespacePattern.ReplaceAllString(content, "><")

	// 2. Remove all newlines and spaces after any > character
	content = whitespaceAfterTagPattern.ReplaceAllString(content, ">")

	// 3. Remove all newlines and spaces before any < character
	content = whitespaceBeforeTagPattern.ReplaceAllString(content, "<")

	// 4. Replace any remaining newlines and spaces with a single space
	content = remainingNewlinesPattern.ReplaceAllString(content, " ")

	return content
}

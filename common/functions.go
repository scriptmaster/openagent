package common

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

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

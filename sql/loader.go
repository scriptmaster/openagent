package sql

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	queries     map[string]map[string]string
	queriesOnce sync.Once
)

// LoadSQLQueries loads SQL queries from the data/postgres directory
func LoadSQLQueries() error {
	var err error
	queriesOnce.Do(func() {
		queries = make(map[string]map[string]string)
		queries["postgres"] = make(map[string]string)

		// Read all SQL files from data/postgres directory
		files, err := os.ReadDir("data/postgres")
		if err != nil {
			return
		}

		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
				// Read the SQL file
				data, err := os.ReadFile(filepath.Join("data/postgres", file.Name()))
				if err != nil {
					continue
				}

				// Split the content into individual queries
				content := string(data)
				queryBlocks := strings.Split(content, "--")

				for _, block := range queryBlocks {
					block = strings.TrimSpace(block)
					if block == "" {
						continue
					}

					// Extract query name from comment and query content
					lines := strings.SplitN(block, "\n", 2)
					if len(lines) != 2 {
						continue
					}

					queryName := strings.TrimSpace(lines[0])
					queryContent := strings.TrimSpace(lines[1])

					// Store the query
					queries["postgres"][queryName] = queryContent
				}
			}
		}
	})
	return err
}

// GetQuery returns a SQL query for the given database type and query name
func GetQuery(dbType, queryName string) (string, error) {
	if err := LoadSQLQueries(); err != nil {
		return "", err
	}

	dbQueries, ok := queries[dbType]
	if !ok {
		return "", fmt.Errorf("no queries found for database type: %s", dbType)
	}

	query, ok := dbQueries[queryName]
	if !ok {
		return "", fmt.Errorf("query not found: %s", queryName)
	}

	return query, nil
}

// GetQueryForDB returns a SQL query for the given database and query name
func GetQueryForDB(db *sql.DB, queryName string) (string, error) {
	// Get database driver name
	conn, err := db.Driver().Open("")
	if err != nil {
		return "", fmt.Errorf("failed to get database driver: %v", err)
	}
	defer conn.Close()

	// Convert driver name to lowercase for case-insensitive comparison
	dbType := strings.ToLower(fmt.Sprintf("%T", conn))

	// Map driver types to our supported database types
	switch {
	case strings.Contains(dbType, "mysql"):
		return GetQuery("mysql", queryName)
	case strings.Contains(dbType, "postgres"):
		return GetQuery("postgres", queryName)
	case strings.Contains(dbType, "sqlite"):
		return GetQuery("sqlite", queryName)
	default:
		return "", fmt.Errorf("unsupported database type: %s", dbType)
	}
}

package common

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var sqlQueries = make(map[string]string)
var queryNameRegex = regexp.MustCompile(`--\s*name:\s*(\S+)`) // Regex to find "-- name: QueryName"

// LoadNamedSQLFiles recursively scans a directory for .sql files,
// parses named queries (-- name: QueryName), and stores them.
func LoadNamedSQLFiles(dirPath string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %q: %w", path, err)
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".sql") {
			if parseErr := parseSQLFile(path); parseErr != nil {
				return fmt.Errorf("error parsing SQL file %q: %w", path, parseErr)
			}
		}
		return nil
	})
}

// parseSQLFile reads a single SQL file and extracts named queries.
func parseSQLFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentQueryName string
	var currentQuery strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		if matches := queryNameRegex.FindStringSubmatch(trimmedLine); len(matches) > 1 {
			// Found a new query name directive
			if currentQueryName != "" && currentQuery.Len() > 0 {
				// Store the previously accumulated query
				sqlQueries[currentQueryName] = strings.TrimSpace(currentQuery.String())
			}
			// Start accumulating the new query
			currentQueryName = matches[1]
			currentQuery.Reset()
		} else if currentQueryName != "" && !strings.HasPrefix(trimmedLine, "--") {
			// If we have a current query name and the line is not a comment, append it
			currentQuery.WriteString(line)
			currentQuery.WriteString("\n") // Keep newlines for readability
		}
	}

	// Store the last query in the file
	if currentQueryName != "" && currentQuery.Len() > 0 {
		sqlQueries[currentQueryName] = strings.TrimSpace(currentQuery.String())
	}

	return scanner.Err()
}

// GetSQL retrieves a loaded SQL query by name.
func GetSQL(name string) (string, error) {
	query, ok := sqlQueries[name]
	if !ok {
		return "", fmt.Errorf("SQL query named '%s' not found", name)
	}
	return query, nil
}

// MustGetSQL retrieves a loaded SQL query by name, panicking if not found.
// Useful during initialization phases.
func MustGetSQL(name string) string {
	query, err := GetSQL(name)
	if err != nil {
		panic(err)
	}
	return query
}

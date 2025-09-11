package common

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// QueryInfo holds information about a SQL query
type QueryInfo struct {
	Name         string
	ParamCount   int
	ParamDetails string
	Query        string
}

// ExecuteRequest represents a request to execute a query
type ExecuteRequest struct {
	QueryName string
	Params    []interface{}
}

// ExecuteResponse represents the response from executing a query
type ExecuteResponse struct {
	Success      bool            `json:"success"`
	Error        string          `json:"error,omitempty"`
	Columns      []string        `json:"columns,omitempty"`
	Rows         [][]interface{} `json:"rows,omitempty"`
	RowCount     int             `json:"rowCount"`
	Duration     string          `json:"duration"`
	QueryType    string          `json:"queryType"`
	RowsAffected int64           `json:"rowsAffected,omitempty"`
}

// getSQLDirectory returns the SQL directory path based on database driver
func getSQLDirectory() string {
	driver := os.Getenv("DB_DRIVER")
	if driver == "" {
		driver = "postgres" // default to postgres
	}
	return fmt.Sprintf("./data/sql/%s", driver)
}

// GetAvailableQueries returns all available queries grouped by parameter count
func GetAvailableQueries() (map[int][]QueryInfo, error) {
	// Get SQL directory based on database driver
	sqlDir := getSQLDirectory()

	// Get all SQL files from the dynamic SQL directory
	files, err := os.ReadDir(sqlDir)
	if err != nil {
		return nil, fmt.Errorf("error reading SQL directory '%s': %v", sqlDir, err)
	}

	// Map to store queries grouped by parameter count
	queryGroups := make(map[int][]QueryInfo)

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".sql") {
			// Read the file to find named queries
			filePath := fmt.Sprintf("%s/%s", sqlDir, file.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
				log.Printf("Error reading file %s: %v", filePath, err)
				continue
			}

			// Find all named queries in the file
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "-- name:") {
					// Extract query name
					parts := strings.Fields(line)
					if len(parts) >= 3 {
						queryName := parts[2]

						// Get the query content
						query, err := GetSQL(queryName)
						if err != nil {
							log.Printf("Error getting query %s: %v", queryName, err)
							continue
						}

						paramCount := CountQueryParameters(query)
						paramDetails := ExtractParameterDetails(query)

						queryInfo := QueryInfo{
							Name:         queryName,
							ParamCount:   paramCount,
							ParamDetails: paramDetails,
							Query:        query,
						}

						queryGroups[paramCount] = append(queryGroups[paramCount], queryInfo)
					}
				}
			}
		}
	}

	// Sort queries within each group
	for count := range queryGroups {
		sort.Slice(queryGroups[count], func(i, j int) bool {
			return queryGroups[count][i].Name < queryGroups[count][j].Name
		})
	}

	return queryGroups, nil
}

// ExecuteQuery executes a query and returns the results
func ExecuteQuery(req ExecuteRequest, db *sql.DB) ExecuteResponse {
	// Get the SQL query
	query, err := GetSQL(req.QueryName)
	if err != nil {
		return ExecuteResponse{
			Success: false,
			Error:   fmt.Sprintf("Query '%s' not found", req.QueryName),
		}
	}

	// Add default parameters for common queries
	params := AddDefaultParameters(req.QueryName, req.Params)

	// Execute the query
	start := time.Now()
	response := executeQueryInternal(query, params, req.QueryName, db)
	duration := time.Since(start)
	response.Duration = duration.String()

	return response
}

// CountQueryParameters counts the number of $1, $2, etc. parameters in a query
func CountQueryParameters(query string) int {
	maxParam := 0
	words := strings.Fields(query)

	for _, word := range words {
		if strings.HasPrefix(word, "$") {
			// Extract the number after $
			numStr := word[1:]
			// Remove any trailing punctuation
			for i, char := range numStr {
				if char < '0' || char > '9' {
					numStr = numStr[:i]
					break
				}
			}

			if num, err := strconv.Atoi(numStr); err == nil {
				if num > maxParam {
					maxParam = num
				}
			}
		}
	}

	return maxParam
}

// ExtractParameterDetails extracts parameter details from SQL comments
func ExtractParameterDetails(query string) string {
	lines := strings.Split(query, "\n")
	var details []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for parameter documentation comments
		if strings.HasPrefix(line, "-- $") {
			// Extract parameter info like "-- $1: schema_name"
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				param := strings.TrimSpace(parts[0][3:]) // Remove "-- $"
				desc := strings.TrimSpace(parts[1])
				details = append(details, fmt.Sprintf("%s=%s", param, desc))
			}
		}
	}

	if len(details) > 0 {
		return strings.Join(details, ", ")
	}

	// If no detailed comments, provide generic parameter names
	paramCount := CountQueryParameters(query)
	if paramCount > 0 {
		var genericParams []string
		for i := 1; i <= paramCount; i++ {
			genericParams = append(genericParams, fmt.Sprintf("$%d", i))
		}
		return strings.Join(genericParams, ", ")
	}

	return ""
}

// AddDefaultParameters adds default parameters for common queries
func AddDefaultParameters(queryName string, params []interface{}) []interface{} {
	switch queryName {
	case "ListDatabaseTables":
		// Default to 'public' schema if no schema provided
		if len(params) == 0 {
			params = append(params, "public")
		}
	case "GetTableColumns":
		// Default to 'public' schema if only table name provided
		if len(params) == 1 {
			params = append([]interface{}{"public"}, params...)
		}
	}
	return params
}

// Internal helper functions

func executeQueryInternal(query string, params []interface{}, queryName string, db *sql.DB) ExecuteResponse {
	// Determine query type
	queryUpper := strings.ToUpper(strings.TrimSpace(query))

	if strings.HasPrefix(queryUpper, "SELECT") {
		return executeSelectQuery(query, params, queryName, db)
	} else if strings.HasPrefix(queryUpper, "INSERT") || strings.HasPrefix(queryUpper, "UPDATE") || strings.HasPrefix(queryUpper, "DELETE") {
		return executeModifyQuery(query, params, queryName, db)
	} else {
		return executeGenericQuery(query, params, queryName, db)
	}
}

func executeSelectQuery(query string, params []interface{}, queryName string, db *sql.DB) ExecuteResponse {
	rows, err := db.Query(query, params...)
	if err != nil {
		return ExecuteResponse{
			Success: false,
			Error:   fmt.Sprintf("Error executing query '%s': %v", queryName, err),
		}
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return ExecuteResponse{
			Success: false,
			Error:   fmt.Sprintf("Error getting columns: %v", err),
		}
	}

	var allRows [][]interface{}
	rowCount := 0

	for rows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Scan the row
		if err := rows.Scan(valuePtrs...); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		// Convert values to strings for JSON serialization
		rowData := make([]interface{}, len(values))
		for i, val := range values {
			if val == nil {
				rowData[i] = nil
			} else {
				// Handle different data types properly
				switch v := val.(type) {
				case []byte:
					// Convert byte array to string
					rowData[i] = string(v)
				default:
					rowData[i] = v
				}
			}
		}

		allRows = append(allRows, rowData)
		rowCount++
	}

	if err = rows.Err(); err != nil {
		return ExecuteResponse{
			Success: false,
			Error:   fmt.Sprintf("Error iterating rows: %v", err),
		}
	}

	return ExecuteResponse{
		Success:   true,
		Columns:   columns,
		Rows:      allRows,
		RowCount:  rowCount,
		QueryType: "SELECT",
	}
}

func executeModifyQuery(query string, params []interface{}, queryName string, db *sql.DB) ExecuteResponse {
	result, err := db.Exec(query, params...)
	if err != nil {
		return ExecuteResponse{
			Success: false,
			Error:   fmt.Sprintf("Error executing query '%s': %v", queryName, err),
		}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return ExecuteResponse{
			Success: false,
			Error:   fmt.Sprintf("Error getting rows affected: %v", err),
		}
	}

	return ExecuteResponse{
		Success:      true,
		RowCount:     int(rowsAffected),
		QueryType:    "MODIFY",
		RowsAffected: rowsAffected,
	}
}

func executeGenericQuery(query string, params []interface{}, queryName string, db *sql.DB) ExecuteResponse {
	// Try as SELECT first
	rows, err := db.Query(query, params...)
	if err != nil {
		// If SELECT fails, try as EXEC
		result, execErr := db.Exec(query, params...)
		if execErr != nil {
			return ExecuteResponse{
				Success: false,
				Error:   fmt.Sprintf("Error executing query '%s': %v (also tried as EXEC: %v)", queryName, err, execErr),
			}
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return ExecuteResponse{
				Success: false,
				Error:   fmt.Sprintf("Error getting rows affected: %v", err),
			}
		}

		return ExecuteResponse{
			Success:      true,
			RowCount:     int(rowsAffected),
			QueryType:    "EXEC",
			RowsAffected: rowsAffected,
		}
	}
	defer rows.Close()

	// If SELECT worked, display results
	return executeSelectQuery(query, params, queryName, db)
}

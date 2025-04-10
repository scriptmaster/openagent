package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	// PostgreSQL driver
	_ "github.com/lib/pq"
)

var (
	db     *sql.DB
	dbOnce sync.Once
)

// Database schema initialization SQL
const initSchema = `
-- Create schema if not exists
CREATE SCHEMA IF NOT EXISTS ai;

-- Users table
CREATE TABLE IF NOT EXISTS ai.users (
    id SERIAL PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    is_admin BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_logged_in TIMESTAMP WITH TIME ZONE
);

-- Projects table
CREATE TABLE IF NOT EXISTS ai.projects (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    domain_name TEXT,
    created_by INTEGER REFERENCES ai.users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Project database connections
CREATE TABLE IF NOT EXISTS ai.project_dbs (
    id SERIAL PRIMARY KEY,
    project_id INTEGER REFERENCES ai.projects(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    db_type TEXT NOT NULL, -- postgresql, mysql, etc.
    connection_string TEXT NOT NULL, -- base64 encoded
    schema_name TEXT NOT NULL DEFAULT 'public',
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(project_id, name)
);

-- Tables metadata
CREATE TABLE IF NOT EXISTS ai.managed_tables (
    id SERIAL PRIMARY KEY,
    project_id INTEGER REFERENCES ai.projects(id) ON DELETE CASCADE,
    project_db_id INTEGER REFERENCES ai.project_dbs(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    schema_name TEXT NOT NULL DEFAULT 'public',
    description TEXT,
    initialized BOOLEAN NOT NULL DEFAULT FALSE,
    read_only BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(project_db_id, schema_name, name)
);

-- Columns metadata (helpful for UI display customization)
CREATE TABLE IF NOT EXISTS ai.managed_columns (
    id SERIAL PRIMARY KEY,
    table_id INTEGER REFERENCES ai.managed_tables(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    display_name TEXT,
    type TEXT NOT NULL,
    ordinal INTEGER NOT NULL,
    visible BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(table_id, name)
);

-- Settings
CREATE TABLE IF NOT EXISTS ai.settings (
    id SERIAL PRIMARY KEY,
    key TEXT NOT NULL,
    value TEXT,
    description TEXT,
    scope TEXT NOT NULL CHECK (scope IN ('system', 'project', 'user')),
    scope_id INTEGER, -- NULL for system, project_id for project, user_id for user
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(key, scope, COALESCE(scope_id, 0))
);

-- Initial settings
INSERT INTO ai.settings (key, value, description, scope) 
VALUES ('app_name', 'Data Manager', 'Application name', 'system') 
ON CONFLICT (key, scope, COALESCE(scope_id, 0)) DO NOTHING;
`

// Project represents a project
type Project struct {
	ID          int
	Name        string
	Description string
	DomainName  string
	CreatedBy   int
	CreatedAt   time.Time
}

// ProjectDB represents a database connection for a project
type ProjectDB struct {
	ID               int
	ProjectID        int
	Name             string
	Description      string
	DBType           string
	ConnectionString string
	SchemaName       string
	IsDefault        bool
	CreatedAt        time.Time
}

// InitDB initializes the database connection
func InitDB() (*sql.DB, error) {
	var err error
	dbOnce.Do(func() {
		// Get connection details from environment variables
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		dbname := os.Getenv("DB_NAME")

		// Check if PostgreSQL configuration is complete
		if host == "" || port == "" || user == "" || dbname == "" {
			err = fmt.Errorf("incomplete PostgreSQL configuration, server running in maintenance mode")
			return
		}

		// Create connection string
		connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbname)

		// Connect to the database with retry
		var dbErr error
		for attempts := 0; attempts < 3; attempts++ {
			log.Printf("Connecting to PostgreSQL (attempt %d)...", attempts+1)
			db, dbErr = sql.Open("postgres", connStr)
			if dbErr == nil {
				// Test the connection
				if err := db.Ping(); err == nil {
					break
				}
			}
			log.Printf("Failed to connect: %v. Retrying in 2 seconds...", dbErr)
			time.Sleep(2 * time.Second)
		}

		if dbErr != nil {
			err = fmt.Errorf("failed to connect to PostgreSQL: %v", dbErr)
			return
		}

		// Configure connection pool
		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(5 * time.Minute)

		// Initialize database schema
		if _, err := db.Exec(initSchema); err != nil {
			err = fmt.Errorf("failed to initialize PostgreSQL schema: %v", err)
			return
		}

		log.Println("Database connection established and schema initialized")
	})

	if err != nil {
		// Mark that we're in maintenance mode
		log.Printf("WARNING: %v", err)
		SetMaintenanceMode(true)
		return nil, err
	}

	// Not in maintenance mode
	SetMaintenanceMode(false)
	return db, nil
}

// Global maintenance mode flag
var (
	inMaintenanceMode bool
	maintenanceMutex  sync.RWMutex
)

// SetMaintenanceMode sets the maintenance mode flag
func SetMaintenanceMode(enabled bool) {
	maintenanceMutex.Lock()
	defer maintenanceMutex.Unlock()
	inMaintenanceMode = enabled
}

// IsMaintenanceMode checks if the server is in maintenance mode
func IsMaintenanceMode() bool {
	maintenanceMutex.RLock()
	defer maintenanceMutex.RUnlock()
	return inMaintenanceMode
}

// UpdateDatabaseConfig writes new database configuration to .env file
func UpdateDatabaseConfig(host, port, user, password, dbname string) error {
	// Read existing .env file
	content, err := os.ReadFile(".env")
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error reading .env file: %v", err)
	}

	// Parse existing content
	lines := strings.Split(string(content), "\n")
	newLines := []string{}
	dbConfigFound := map[string]bool{
		"DB_HOST":     false,
		"DB_PORT":     false,
		"DB_USER":     false,
		"DB_PASSWORD": false,
		"DB_NAME":     false,
	}

	// Update existing lines
	for _, line := range lines {
		if line == "" {
			newLines = append(newLines, line)
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			newLines = append(newLines, line)
			continue
		}

		key := strings.TrimSpace(parts[0])
		switch key {
		case "DB_HOST":
			newLines = append(newLines, fmt.Sprintf("DB_HOST=%s", host))
			dbConfigFound["DB_HOST"] = true
		case "DB_PORT":
			newLines = append(newLines, fmt.Sprintf("DB_PORT=%s", port))
			dbConfigFound["DB_PORT"] = true
		case "DB_USER":
			newLines = append(newLines, fmt.Sprintf("DB_USER=%s", user))
			dbConfigFound["DB_USER"] = true
		case "DB_PASSWORD":
			newLines = append(newLines, fmt.Sprintf("DB_PASSWORD=%s", password))
			dbConfigFound["DB_PASSWORD"] = true
		case "DB_NAME":
			newLines = append(newLines, fmt.Sprintf("DB_NAME=%s", dbname))
			dbConfigFound["DB_NAME"] = true
		default:
			newLines = append(newLines, line)
		}
	}

	// Add missing configurations
	if !dbConfigFound["DB_HOST"] {
		newLines = append(newLines, fmt.Sprintf("DB_HOST=%s", host))
	}
	if !dbConfigFound["DB_PORT"] {
		newLines = append(newLines, fmt.Sprintf("DB_PORT=%s", port))
	}
	if !dbConfigFound["DB_USER"] {
		newLines = append(newLines, fmt.Sprintf("DB_USER=%s", user))
	}
	if !dbConfigFound["DB_PASSWORD"] {
		newLines = append(newLines, fmt.Sprintf("DB_PASSWORD=%s", password))
	}
	if !dbConfigFound["DB_NAME"] {
		newLines = append(newLines, fmt.Sprintf("DB_NAME=%s", dbname))
	}

	// Write back to .env file
	err = os.WriteFile(".env", []byte(strings.Join(newLines, "\n")), 0644)
	if err != nil {
		return fmt.Errorf("error writing .env file: %v", err)
	}

	return nil
}

// UserService provides methods to work with users
type UserService struct {
	db *sql.DB
}

// NewUserService creates a new UserService
func NewUserService(db *sql.DB) *UserService {
	return &UserService{db: db}
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(email string) (User, error) {
	var user User
	query := `SELECT id, email, is_admin, created_at, last_logged_in FROM ai.users WHERE email = $1`
	err := s.db.QueryRow(query, email).Scan(
		&user.ID, &user.Email, &user.IsAdmin, &user.CreatedAt, &user.LastLoggedIn)
	if err != nil {
		if err == sql.ErrNoRows {
			return User{}, fmt.Errorf("user not found")
		}
		return User{}, err
	}
	return user, nil
}

// CreateUser creates a new user
func (s *UserService) CreateUser(email string) (User, error) {
	// Check if any users exist (first user is admin)
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM ai.users").Scan(&count)
	if err != nil {
		return User{}, err
	}

	isAdmin := count == 0 // First user is admin

	var user User
	query := `
		INSERT INTO ai.users (email, is_admin) 
		VALUES ($1, $2) 
		RETURNING id, email, is_admin, created_at
	`
	err = s.db.QueryRow(query, email, isAdmin).Scan(&user.ID, &user.Email, &user.IsAdmin, &user.CreatedAt)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

// GetOrCreateUser retrieves a user by email or creates a new one
func (s *UserService) GetOrCreateUser(email string) (User, bool, error) {
	user, err := s.GetUserByEmail(email)
	if err != nil {
		if err.Error() == "user not found" {
			newUser, err := s.CreateUser(email)
			if err != nil {
				return User{}, false, err
			}
			return newUser, true, nil // true = created new user
		}
		return User{}, false, err
	}
	return user, false, nil // false = existing user
}

// UpdateUserLastLogin updates the last_logged_in timestamp for a user
func (s *UserService) UpdateUserLastLogin(userID int) error {
	_, err := s.db.Exec(`
		UPDATE ai.users 
		SET last_logged_in = NOW() 
		WHERE id = $1
	`, userID)
	return err
}

// ProjectService provides methods to work with projects
type ProjectService struct {
	db *sql.DB
}

// NewProjectService creates a new ProjectService
func NewProjectService(db *sql.DB) *ProjectService {
	return &ProjectService{db: db}
}

// CreateProject creates a new project
func (s *ProjectService) CreateProject(name, description, domainName string, createdBy int) (Project, error) {
	var project Project
	query := `
		INSERT INTO ai.projects (name, description, domain_name, created_by) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id, name, description, domain_name, created_by, created_at
	`
	err := s.db.QueryRow(query, name, description, domainName, createdBy).Scan(
		&project.ID, &project.Name, &project.Description, &project.DomainName,
		&project.CreatedBy, &project.CreatedAt)
	if err != nil {
		return Project{}, err
	}
	return project, nil
}

// GetProjects retrieves all projects
func (s *ProjectService) GetProjects() ([]Project, error) {
	rows, err := s.db.Query(`
		SELECT id, name, description, domain_name, created_by, created_at 
		FROM ai.projects 
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.DomainName,
			&p.CreatedBy, &p.CreatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, nil
}

// TableService provides methods to work with tables
type TableService struct {
	db *sql.DB
}

// ManagedTable represents a table's metadata
type ManagedTable struct {
	ID          int
	ProjectID   int
	ProjectDBID int
	Name        string
	SchemaName  string
	Description string
	Initialized bool
	ReadOnly    bool
	CreatedAt   time.Time
}

// ManagedColumn represents a column's metadata for UI display
type ManagedColumn struct {
	ID          int
	TableID     int
	Name        string
	DisplayName string
	Type        string
	Ordinal     int
	Visible     bool
	CreatedAt   time.Time
}

// NewTableService creates a new TableService
func NewTableService(db *sql.DB) *TableService {
	return &TableService{db: db}
}

// GetManagedTables retrieves all managed tables for a project
func (s *TableService) GetManagedTables(projectID int) ([]ManagedTable, error) {
	rows, err := s.db.Query(`
		SELECT id, project_id, project_db_id, name, schema_name, description, initialized, read_only, created_at 
		FROM ai.managed_tables 
		WHERE project_id = $1
		ORDER BY name
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []ManagedTable
	for rows.Next() {
		var t ManagedTable
		if err := rows.Scan(&t.ID, &t.ProjectID, &t.ProjectDBID, &t.Name, &t.SchemaName,
			&t.Description, &t.Initialized, &t.ReadOnly, &t.CreatedAt); err != nil {
			return nil, err
		}
		tables = append(tables, t)
	}
	return tables, nil
}

// GetManagedTablesByProjectDB retrieves managed tables for a specific project database connection
func (s *TableService) GetManagedTablesByProjectDB(projectDBID int) ([]ManagedTable, error) {
	rows, err := s.db.Query(`
		SELECT id, project_id, project_db_id, name, schema_name, description, initialized, read_only, created_at 
		FROM ai.managed_tables 
		WHERE project_db_id = $1
		ORDER BY name
	`, projectDBID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []ManagedTable
	for rows.Next() {
		var t ManagedTable
		if err := rows.Scan(&t.ID, &t.ProjectID, &t.ProjectDBID, &t.Name, &t.SchemaName,
			&t.Description, &t.Initialized, &t.ReadOnly, &t.CreatedAt); err != nil {
			return nil, err
		}
		tables = append(tables, t)
	}
	return tables, nil
}

// GetManagedColumns retrieves all managed columns for a table
func (s *TableService) GetManagedColumns(tableID int) ([]ManagedColumn, error) {
	rows, err := s.db.Query(`
		SELECT id, table_id, name, display_name, type, ordinal, visible, created_at 
		FROM ai.managed_columns 
		WHERE table_id = $1
		ORDER BY ordinal
	`, tableID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []ManagedColumn
	for rows.Next() {
		var c ManagedColumn
		var displayName sql.NullString
		if err := rows.Scan(&c.ID, &c.TableID, &c.Name, &displayName, &c.Type, &c.Ordinal, &c.Visible, &c.CreatedAt); err != nil {
			return nil, err
		}
		if displayName.Valid {
			c.DisplayName = displayName.String
		} else {
			c.DisplayName = c.Name // Default to column name if display name is not set
		}
		columns = append(columns, c)
	}
	return columns, nil
}

// DirectDataService provides methods to work directly with database tables
type DirectDataService struct {
	db            *sql.DB
	dbConnections map[int]*sql.DB // Cache of database connections by project_db_id
	mu            sync.RWMutex    // Mutex to protect the connections map
}

// NewDirectDataService creates a new DirectDataService
func NewDirectDataService(db *sql.DB) *DirectDataService {
	return &DirectDataService{
		db:            db,
		dbConnections: make(map[int]*sql.DB),
	}
}

// getConnection gets or creates a database connection for a ProjectDB
func (s *DirectDataService) getConnection(ctx context.Context, projectDBID int) (*sql.DB, error) {
	// Check cache first
	s.mu.RLock()
	conn, exists := s.dbConnections[projectDBID]
	s.mu.RUnlock()

	if exists && conn != nil {
		// Test if connection is still valid
		if err := conn.PingContext(ctx); err == nil {
			return conn, nil
		}
		// Connection is stale, remove it
		s.mu.Lock()
		delete(s.dbConnections, projectDBID)
		s.mu.Unlock()
	}

	// Get project DB info
	var projectDB ProjectDB
	query := `
		SELECT id, project_id, name, description, db_type, connection_string, schema_name, is_default, created_at 
		FROM ai.project_dbs 
		WHERE id = $1
	`
	err := s.db.QueryRowContext(ctx, query, projectDBID).Scan(
		&projectDB.ID, &projectDB.ProjectID, &projectDB.Name, &projectDB.Description,
		&projectDB.DBType, &projectDB.ConnectionString, &projectDB.SchemaName,
		&projectDB.IsDefault, &projectDB.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get project DB info: %v", err)
	}

	// Decode connection string
	connStr, err := DecodeConnectionString(projectDB.ConnectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to decode connection string: %v", err)
	}

	// Create new connection
	var newConn *sql.DB
	switch projectDB.DBType {
	case "postgresql":
		newConn, err = sql.Open("postgres", connStr)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to database: %v", err)
		}

		// Test connection
		if err := newConn.PingContext(ctx); err != nil {
			newConn.Close()
			return nil, fmt.Errorf("failed to ping database: %v", err)
		}

	// Add more database types as needed
	default:
		return nil, fmt.Errorf("unsupported database type: %s", projectDB.DBType)
	}

	// Cache the connection
	s.mu.Lock()
	s.dbConnections[projectDBID] = newConn
	s.mu.Unlock()

	return newConn, nil
}

// GetTableRows retrieves rows from a managed table
func (s *DirectDataService) GetTableRows(ctx context.Context, projectDBID int, schemaName, tableName string, limit, offset int) ([]map[string]interface{}, error) {
	// Get connection
	conn, err := s.getConnection(ctx, projectDBID)
	if err != nil {
		return nil, err
	}

	// Build query with safe table name (to prevent SQL injection)
	// Note: In a real system, you'd want to validate tableName against a list of allowed tables
	query := fmt.Sprintf(`
		SELECT * FROM %s.%s
		LIMIT $1 OFFSET $2
	`, schemaName, tableName)

	rows, err := conn.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Get column names from the query
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// Initialize result slice
	var result []map[string]interface{}

	// Prepare values holders for each row
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// Iterate through rows
	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		// Convert row to map
		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]

			// Handle different types
			switch v := val.(type) {
			case []byte:
				// Convert []byte to string for text values
				row[col] = string(v)
			default:
				row[col] = v
			}
		}

		result = append(result, row)
	}

	return result, nil
}

// InsertTableRow inserts a row into a managed table
func (s *DirectDataService) InsertTableRow(ctx context.Context, projectDBID int, schemaName, tableName string, data map[string]interface{}) error {
	// Get connection
	conn, err := s.getConnection(ctx, projectDBID)
	if err != nil {
		return err
	}

	// Extract columns and values from data
	var columns []string
	var placeholders []string
	var values []interface{}

	i := 1
	for col, val := range data {
		columns = append(columns, col)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		values = append(values, val)
		i++
	}

	// Build the query
	query := fmt.Sprintf(
		"INSERT INTO %s.%s (%s) VALUES (%s)",
		schemaName,
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	// Execute the insert
	_, err = conn.ExecContext(ctx, query, values...)
	return err
}

// UpdateTableRow updates a row in a managed table
func (s *DirectDataService) UpdateTableRow(ctx context.Context, projectDBID int, schemaName, tableName string, idColumn string, idValue interface{}, data map[string]interface{}) error {
	// Get connection
	conn, err := s.getConnection(ctx, projectDBID)
	if err != nil {
		return err
	}

	// Extract columns and values from data
	var setClauses []string
	var values []interface{}

	i := 1
	for col, val := range data {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, i))
		values = append(values, val)
		i++
	}

	// Add the ID value to values
	values = append(values, idValue)

	// Build the query
	query := fmt.Sprintf(
		"UPDATE %s.%s SET %s WHERE %s = $%d",
		schemaName,
		tableName,
		strings.Join(setClauses, ", "),
		idColumn,
		i,
	)

	// Execute the update
	_, err = conn.ExecContext(ctx, query, values...)
	return err
}

// DeleteTableRow deletes a row from a managed table
func (s *DirectDataService) DeleteTableRow(ctx context.Context, projectDBID int, schemaName, tableName string, idColumn string, idValue interface{}) error {
	// Get connection
	conn, err := s.getConnection(ctx, projectDBID)
	if err != nil {
		return err
	}

	query := fmt.Sprintf("DELETE FROM %s.%s WHERE %s = $1", schemaName, tableName, idColumn)
	_, err = conn.ExecContext(ctx, query, idValue)
	return err
}

// CloseConnections closes all database connections
func (s *DirectDataService) CloseConnections() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, conn := range s.dbConnections {
		if conn != nil {
			conn.Close()
		}
		delete(s.dbConnections, id)
	}
}

// SettingsService provides methods to work with settings
type SettingsService struct {
	db *sql.DB
}

// Setting represents a system setting
type Setting struct {
	ID          int
	Key         string
	Value       string
	Description string
	Scope       string // "system", "project", or "user"
	ScopeID     int    // ID of the project or user (0 for system)
	UpdatedAt   time.Time
}

// NewSettingsService creates a new SettingsService
func NewSettingsService(db *sql.DB) *SettingsService {
	return &SettingsService{db: db}
}

// GetSetting retrieves a setting by key and scope
func (s *SettingsService) GetSetting(key string, scope string, scopeID *int) (Setting, error) {
	var setting Setting
	var query string
	var args []interface{}

	if scopeID == nil {
		query = `SELECT id, key, value, description, scope, scope_id, updated_at 
				FROM ai.settings 
				WHERE key = $1 AND scope = $2 AND scope_id IS NULL`
		args = []interface{}{key, scope}
	} else {
		query = `SELECT id, key, value, description, scope, scope_id, updated_at 
				FROM ai.settings 
				WHERE key = $1 AND scope = $2 AND scope_id = $3`
		args = []interface{}{key, scope, *scopeID}
	}

	var scopeIDNull sql.NullInt64
	err := s.db.QueryRow(query, args...).Scan(
		&setting.ID, &setting.Key, &setting.Value, &setting.Description,
		&setting.Scope, &scopeIDNull, &setting.UpdatedAt)

	if err != nil {
		return Setting{}, err
	}

	if scopeIDNull.Valid {
		setting.ScopeID = int(scopeIDNull.Int64)
	}

	return setting, nil
}

// UpdateSetting updates a setting with scope
func (s *SettingsService) UpdateSetting(key, value, scope string, scopeID *int) error {
	var query string
	var args []interface{}

	if scopeID == nil {
		query = `
			INSERT INTO ai.settings (key, value, scope, scope_id, updated_at)
			VALUES ($1, $2, $3, NULL, NOW())
			ON CONFLICT (key, scope, COALESCE(scope_id, 0)) 
			DO UPDATE SET value = $2, updated_at = NOW()
		`
		args = []interface{}{key, value, scope}
	} else {
		query = `
			INSERT INTO ai.settings (key, value, scope, scope_id, updated_at)
			VALUES ($1, $2, $3, $4, NOW())
			ON CONFLICT (key, scope, COALESCE(scope_id, 0)) 
			DO UPDATE SET value = $2, updated_at = NOW()
		`
		args = []interface{}{key, value, scope, *scopeID}
	}

	_, err := s.db.Exec(query, args...)
	return err
}

// GetAllSettings retrieves all settings
func (s *SettingsService) GetAllSettings() ([]Setting, error) {
	rows, err := s.db.Query(`
		SELECT id, key, value, description, scope, scope_id, updated_at 
		FROM ai.settings 
		ORDER BY key
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []Setting
	for rows.Next() {
		var setting Setting
		var scopeID sql.NullInt64
		if err := rows.Scan(&setting.ID, &setting.Key, &setting.Value, &setting.Description,
			&setting.Scope, &scopeID, &setting.UpdatedAt); err != nil {
			return nil, err
		}
		if scopeID.Valid {
			setting.ScopeID = int(scopeID.Int64)
		}
		settings = append(settings, setting)
	}
	return settings, nil
}

// DatabaseMetadataService provides methods to work with database metadata
type DatabaseMetadataService struct {
	db          *sql.DB
	dataService *DirectDataService
}

// NewDatabaseMetadataService creates a new DatabaseMetadataService
func NewDatabaseMetadataService(db *sql.DB, dataService *DirectDataService) *DatabaseMetadataService {
	return &DatabaseMetadataService{
		db:          db,
		dataService: dataService,
	}
}

// ListDatabaseTables returns a list of all tables in the database with management status
func (s *DatabaseMetadataService) ListDatabaseTables(ctx context.Context, projectID int, projectDBID int) ([]TableMetadata, error) {
	// First get managed tables info for this project DB
	managedTables := make(map[string]ManagedTable)

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, schema_name, description, initialized, read_only
		FROM ai.managed_tables
		WHERE project_id = $1 AND project_db_id = $2
	`, projectID, projectDBID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var t ManagedTable
		if err := rows.Scan(&t.ID, &t.Name, &t.SchemaName, &t.Description, &t.Initialized, &t.ReadOnly); err != nil {
			return nil, err
		}
		// Use format schema.tablename as the key
		key := fmt.Sprintf("%s.%s", t.SchemaName, t.Name)
		managedTables[key] = t
	}

	// Get the project DB info
	var projectDB ProjectDB
	err = s.db.QueryRowContext(ctx, `
		SELECT id, project_id, name, description, db_type, connection_string, schema_name, is_default
		FROM ai.project_dbs
		WHERE id = $1
	`, projectDBID).Scan(
		&projectDB.ID, &projectDB.ProjectID, &projectDB.Name, &projectDB.Description,
		&projectDB.DBType, &projectDB.ConnectionString, &projectDB.SchemaName, &projectDB.IsDefault)
	if err != nil {
		return nil, err
	}

	// Get a connection to the database
	conn, err := s.dataService.getConnection(ctx, projectDBID)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Get schema list to search in
	var schemas []string
	if projectDB.SchemaName == "*" {
		// Get all schemas except system ones
		schemaQuery := `
			SELECT schema_name 
			FROM information_schema.schemata 
			WHERE schema_name NOT IN ('information_schema', 'pg_catalog', 'pg_toast')
			ORDER BY schema_name
		`
		schemaRows, err := conn.QueryContext(ctx, schemaQuery)
		if err != nil {
			return nil, err
		}
		defer schemaRows.Close()

		for schemaRows.Next() {
			var schemaName string
			if err := schemaRows.Scan(&schemaName); err != nil {
				return nil, err
			}
			schemas = append(schemas, schemaName)
		}
	} else {
		// Just use the specified schema
		schemas = []string{projectDB.SchemaName}
	}

	// Now get tables for each schema
	var tables []TableMetadata
	for _, schema := range schemas {
		// Query for tables in this schema
		query := `
			SELECT table_name 
			FROM information_schema.tables 
			WHERE table_schema = $1
			AND table_type = 'BASE TABLE'
			ORDER BY table_name
		`

		tableRows, err := conn.QueryContext(ctx, query, schema)
		if err != nil {
			return nil, err
		}

		for tableRows.Next() {
			var tableName string
			if err := tableRows.Scan(&tableName); err != nil {
				tableRows.Close()
				return nil, err
			}

			// Create full name key
			key := fmt.Sprintf("%s.%s", schema, tableName)

			// Check if this table is managed
			managedTable, isManaged := managedTables[key]

			metadata := TableMetadata{
				Name:       tableName,
				SchemaName: schema,
				IsManaged:  isManaged,
			}

			if isManaged {
				metadata.Description = managedTable.Description
				metadata.ManagedID = managedTable.ID
				metadata.ReadOnly = managedTable.ReadOnly
				metadata.Initialized = managedTable.Initialized
			}

			tables = append(tables, metadata)
		}
		tableRows.Close()
	}

	return tables, nil
}

// TableMetadata contains information about a database table
type TableMetadata struct {
	Name        string
	SchemaName  string
	Description string
	IsManaged   bool
	ManagedID   int
	ReadOnly    bool
	Initialized bool
}

// ColumnMetadata contains information about a database column
type ColumnMetadata struct {
	Name        string
	Type        string
	IsNullable  bool
	OrdinalPos  int
	DisplayName string
	Visible     bool
	ManagedID   int
	IsManaged   bool
}

// GetTableColumns returns all columns for a database table with management status
func (s *DatabaseMetadataService) GetTableColumns(ctx context.Context, projectDBID int, tableID int, schema, tableName string) ([]ColumnMetadata, error) {
	// First get managed columns if this is a managed table
	managedColumns := make(map[string]ManagedColumn)

	if tableID > 0 {
		rows, err := s.db.QueryContext(ctx, `
			SELECT id, name, display_name, type, ordinal, visible
			FROM ai.managed_columns
			WHERE table_id = $1
			ORDER BY ordinal
		`, tableID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var c ManagedColumn
			var displayName sql.NullString
			if err := rows.Scan(&c.ID, &c.Name, &displayName, &c.Type, &c.Ordinal, &c.Visible); err != nil {
				return nil, err
			}
			if displayName.Valid {
				c.DisplayName = displayName.String
			} else {
				c.DisplayName = c.Name
			}
			managedColumns[c.Name] = c
		}
	}

	// Now get schema columns from external DB
	// Get a connection to the database
	conn, err := s.dataService.getConnection(ctx, projectDBID)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Now get schema columns
	query := `
		SELECT 
			column_name, 
			data_type,
			is_nullable = 'YES' as is_nullable,
			ordinal_position
		FROM 
			information_schema.columns
		WHERE 
			table_schema = $1 
			AND table_name = $2
		ORDER BY 
			ordinal_position
	`

	rows, err := conn.QueryContext(ctx, query, schema, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []ColumnMetadata
	for rows.Next() {
		var col ColumnMetadata
		if err := rows.Scan(&col.Name, &col.Type, &col.IsNullable, &col.OrdinalPos); err != nil {
			return nil, err
		}

		// Check if column is managed
		managedColumn, isManaged := managedColumns[col.Name]
		col.IsManaged = isManaged

		if isManaged {
			col.DisplayName = managedColumn.DisplayName
			col.Visible = managedColumn.Visible
			col.ManagedID = managedColumn.ID
		} else {
			col.DisplayName = col.Name
			col.Visible = true
		}

		columns = append(columns, col)
	}

	return columns, nil
}

// InitializeTable marks a table as initialized and creates managed column entries
func (s *DatabaseMetadataService) InitializeTable(ctx context.Context, projectID, projectDBID int, schema, tableName, description string, readOnly bool) (ManagedTable, error) {
	// Start a transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ManagedTable{}, err
	}
	defer tx.Rollback()

	// Insert into managed_tables
	var table ManagedTable
	err = tx.QueryRowContext(ctx, `
		INSERT INTO ai.managed_tables (project_id, project_db_id, name, schema_name, description, initialized, read_only)
		VALUES ($1, $2, $3, $4, $5, true, $6)
		RETURNING id, project_id, project_db_id, name, schema_name, description, initialized, read_only, created_at
	`, projectID, projectDBID, tableName, schema, description, readOnly).Scan(
		&table.ID, &table.ProjectID, &table.ProjectDBID, &table.Name, &table.SchemaName,
		&table.Description, &table.Initialized, &table.ReadOnly, &table.CreatedAt)

	if err != nil {
		return ManagedTable{}, err
	}

	// Get columns from information schema of the external DB
	// Get a connection to the database
	conn, err := s.dataService.getConnection(ctx, projectDBID)
	if err != nil {
		return ManagedTable{}, fmt.Errorf("failed to connect to database: %v", err)
	}

	query := `
		SELECT 
			column_name, 
			data_type,
			ordinal_position
		FROM 
			information_schema.columns
		WHERE 
			table_schema = $1 
			AND table_name = $2
		ORDER BY 
			ordinal_position
	`

	rows, err := conn.QueryContext(ctx, query, schema, tableName)
	if err != nil {
		return ManagedTable{}, err
	}
	defer rows.Close()

	// Insert managed columns
	for rows.Next() {
		var colName, colType string
		var ordinal int
		if err := rows.Scan(&colName, &colType, &ordinal); err != nil {
			return ManagedTable{}, err
		}

		// Map external DB data type to our type system
		mappedType := mapDbTypeToColumnType(colType)

		_, err = tx.ExecContext(ctx, `
			INSERT INTO ai.managed_columns (table_id, name, display_name, type, ordinal, visible)
			VALUES ($1, $2, $2, $3, $4, true)
		`, table.ID, colName, mappedType, ordinal)

		if err != nil {
			return ManagedTable{}, err
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return ManagedTable{}, err
	}

	return table, nil
}

// UpdateTableStatus updates the initialized and read_only status of a managed table
func (s *DatabaseMetadataService) UpdateTableStatus(ctx context.Context, tableID int, initialized, readOnly bool) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE ai.managed_tables
		SET initialized = $2, read_only = $3
		WHERE id = $1
	`, tableID, initialized, readOnly)
	return err
}

// UpdateColumnVisibility updates the visibility and display name of a managed column
func (s *DatabaseMetadataService) UpdateColumnVisibility(ctx context.Context, columnID int, displayName string, visible bool) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE ai.managed_columns
		SET display_name = $2, visible = $3
		WHERE id = $1
	`, columnID, displayName, visible)
	return err
}

// mapDbTypeToColumnType maps PostgreSQL data types to our system's column types
func mapDbTypeToColumnType(dbType string) string {
	switch dbType {
	case "integer", "bigint", "smallint", "decimal", "numeric", "real", "double precision":
		return "number"
	case "boolean":
		return "boolean"
	case "timestamp", "timestamp with time zone", "date", "time", "time with time zone":
		return "date"
	default:
		return "text" // Default for varchar, char, text, etc.
	}
}

// ProjectDBService provides methods to work with project database connections
type ProjectDBService struct {
	db *sql.DB
}

// NewProjectDBService creates a new ProjectDBService
func NewProjectDBService(db *sql.DB) *ProjectDBService {
	return &ProjectDBService{db: db}
}

// EncodeConnectionString encodes a connection string to base64
func EncodeConnectionString(connStr string) string {
	return base64.StdEncoding.EncodeToString([]byte(connStr))
}

// DecodeConnectionString decodes a base64 encoded connection string
func DecodeConnectionString(encoded string) (string, error) {
	bytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CreateProjectDB creates a new database connection for a project
func (s *ProjectDBService) CreateProjectDB(projectID int, name, description, dbType,
	connectionString, schemaName string, isDefault bool) (ProjectDB, error) {

	// Encode connection string
	encodedConnStr := EncodeConnectionString(connectionString)

	// Start a transaction
	tx, err := s.db.Begin()
	if err != nil {
		return ProjectDB{}, err
	}
	defer tx.Rollback()

	// If this is the default connection, unset any existing default
	if isDefault {
		_, err = tx.Exec(`
			UPDATE ai.project_dbs
			SET is_default = false
			WHERE project_id = $1
		`, projectID)
		if err != nil {
			return ProjectDB{}, err
		}
	}

	// Create the new connection
	var projectDB ProjectDB
	query := `
		INSERT INTO ai.project_dbs (project_id, name, description, db_type, connection_string, schema_name, is_default) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) 
		RETURNING id, project_id, name, description, db_type, connection_string, schema_name, is_default, created_at
	`
	err = tx.QueryRow(query, projectID, name, description, dbType, encodedConnStr, schemaName, isDefault).Scan(
		&projectDB.ID, &projectDB.ProjectID, &projectDB.Name, &projectDB.Description,
		&projectDB.DBType, &projectDB.ConnectionString, &projectDB.SchemaName, &projectDB.IsDefault, &projectDB.CreatedAt)
	if err != nil {
		return ProjectDB{}, err
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return ProjectDB{}, err
	}

	return projectDB, nil
}

// GetProjectDBs retrieves all database connections for a project
func (s *ProjectDBService) GetProjectDBs(projectID int) ([]ProjectDB, error) {
	rows, err := s.db.Query(`
		SELECT id, project_id, name, description, db_type, connection_string, schema_name, is_default, created_at 
		FROM ai.project_dbs 
		WHERE project_id = $1
		ORDER BY name
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projectDBs []ProjectDB
	for rows.Next() {
		var db ProjectDB
		if err := rows.Scan(&db.ID, &db.ProjectID, &db.Name, &db.Description,
			&db.DBType, &db.ConnectionString, &db.SchemaName, &db.IsDefault, &db.CreatedAt); err != nil {
			return nil, err
		}
		projectDBs = append(projectDBs, db)
	}
	return projectDBs, nil
}

// GetProjectDB retrieves a database connection by ID
func (s *ProjectDBService) GetProjectDB(id int) (ProjectDB, error) {
	var projectDB ProjectDB
	query := `
		SELECT id, project_id, name, description, db_type, connection_string, schema_name, is_default, created_at 
		FROM ai.project_dbs 
		WHERE id = $1
	`
	err := s.db.QueryRow(query, id).Scan(&projectDB.ID, &projectDB.ProjectID, &projectDB.Name,
		&projectDB.Description, &projectDB.DBType, &projectDB.ConnectionString,
		&projectDB.SchemaName, &projectDB.IsDefault, &projectDB.CreatedAt)
	if err != nil {
		return ProjectDB{}, err
	}
	return projectDB, nil
}

// UpdateProjectDB updates a database connection
func (s *ProjectDBService) UpdateProjectDB(id int, name, description, dbType,
	connectionString, schemaName string, isDefault bool) error {

	// Retrieve current DB connection info
	currentDB, err := s.GetProjectDB(id)
	if err != nil {
		return err
	}

	// Encode connection string if it changed
	encodedConnStr := currentDB.ConnectionString
	if connectionString != "" {
		// Connection string provided - it's a new one
		encodedConnStr = EncodeConnectionString(connectionString)
	}

	// Start a transaction
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// If this is the default connection, unset any existing default
	if isDefault {
		_, err = tx.Exec(`
			UPDATE ai.project_dbs
			SET is_default = false
			WHERE project_id = $1 AND id != $2
		`, currentDB.ProjectID, id)
		if err != nil {
			return err
		}
	}

	// Update the connection
	_, err = tx.Exec(`
		UPDATE ai.project_dbs
		SET name = $1, description = $2, db_type = $3, connection_string = $4, schema_name = $5, is_default = $6
		WHERE id = $7
	`, name, description, dbType, encodedConnStr, schemaName, isDefault, id)
	if err != nil {
		return err
	}

	// Commit transaction
	return tx.Commit()
}

// DeleteProjectDB deletes a database connection
func (s *ProjectDBService) DeleteProjectDB(id int) error {
	_, err := s.db.Exec(`
		DELETE FROM ai.project_dbs
		WHERE id = $1
	`, id)
	return err
}

// TestConnection tests a database connection
func (s *ProjectDBService) TestConnection(projectDB ProjectDB) error {
	// Decode connection string
	connStr, err := DecodeConnectionString(projectDB.ConnectionString)
	if err != nil {
		return fmt.Errorf("failed to decode connection string: %v", err)
	}

	// Test connection based on db type
	switch projectDB.DBType {
	case "postgresql":
		testDB, err := sql.Open("postgres", connStr)
		if err != nil {
			return fmt.Errorf("failed to open connection: %v", err)
		}
		defer testDB.Close()

		err = testDB.Ping()
		if err != nil {
			return fmt.Errorf("failed to ping database: %v", err)
		}

	// Add more database types as needed
	default:
		return fmt.Errorf("unsupported database type: %s", projectDB.DBType)
	}

	return nil
}

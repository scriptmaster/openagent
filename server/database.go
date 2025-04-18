package server

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/scriptmaster/openagent/auth"

	// PostgreSQL driver
	_ "github.com/lib/pq"
	// MySQL driver
	_ "github.com/go-sql-driver/mysql"
)

var (
	db     *sql.DB
	dbOnce sync.Once
)

type User = auth.User

// applySchema runs the SQL in migrations directory to initialize the database
func applySchema(db *sql.DB) error {
	// First check if the ai schema exists - if not, reset migration tracking
	var schemaExists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM information_schema.schemata WHERE schema_name = 'ai')").Scan(&schemaExists)
	if err != nil {
		return fmt.Errorf("failed to check if schema exists: %v", err)
	}

	if !schemaExists {
		log.Println("AI schema does not exist - resetting migration tracking to apply all migrations")
		if err := UpdateMigrationStart(0); err != nil {
			log.Printf("Warning: Failed to reset MIGRATION_START in .env: %v", err)
		}
	}

	// Read and execute migration files in order
	files, err := os.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %v", err)
	}

	// Sort filenames to ensure they're applied in order
	fileNames := make([]string, 0, len(files))
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		// Include numbered migration files (supports 001_*.sql format)
		if strings.HasPrefix(file.Name(), "0") || strings.HasPrefix(file.Name(), "1") {
			fileNames = append(fileNames, file.Name())
		}
	}
	sort.Strings(fileNames)

	if len(fileNames) == 0 {
		return fmt.Errorf("no migration files found in migrations directory")
	}

	// Check for last applied migration number
	lastApplied := 0
	if !schemaExists {
		// Schema doesn't exist, so start from scratch
		lastApplied = 0
	} else if migStart := os.Getenv("MIGRATION_START"); migStart != "" {
		// Parse the migration start number, handle both "007" and "7" formats
		lastApplied, err = strconv.Atoi(migStart)
		if err != nil {
			log.Printf("Warning: Invalid MIGRATION_START value: %s", migStart)
			lastApplied = 0
		}
	}

	// Filter migrations to only include those higher than lastApplied
	var pendingMigrations []string
	var highestApplied int

	for _, fileName := range fileNames {
		// Extract migration number from filename (e.g., "001" from "001_schema.sql")
		numStr := strings.SplitN(fileName, "_", 2)[0]
		num, err := strconv.Atoi(numStr)
		if err != nil {
			log.Printf("Warning: Invalid migration filename format: %s", fileName)
			continue
		}

		// Only apply migrations with higher numbers than last applied
		if num > lastApplied {
			pendingMigrations = append(pendingMigrations, fileName)
		}

		// Track highest migration number
		if num > highestApplied {
			highestApplied = num
		}
	}

	// No new migrations to apply
	if len(pendingMigrations) == 0 {
		log.Printf("No new migrations to apply (last applied: %d)", lastApplied)
		return nil
	}

	log.Printf("Found %d pending migrations to apply", len(pendingMigrations))

	// Execute each pending migration
	for _, fileName := range pendingMigrations {
		log.Printf("Applying migration: %s", fileName)

		// Read migration file
		migrationPath := filepath.Join("migrations", fileName)
		migration, err := os.ReadFile(migrationPath)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %v", fileName, err)
		}

		// Execute migration
		if _, err := db.Exec(string(migration)); err != nil {
			return fmt.Errorf("failed to apply migration %s: %v", fileName, err)
		}
	}

	// Update MIGRATION_START in .env file
	if highestApplied > lastApplied {
		if err := UpdateMigrationStart(highestApplied); err != nil {
			log.Printf("Warning: Failed to update MIGRATION_START in .env: %v", err)
		} else {
			log.Printf("Updated MIGRATION_START to %d", highestApplied)
		}
	}

	log.Println("All migrations applied successfully")
	return nil
}

// updateMigrationStart updates the MIGRATION_START value in the .env file
func updateMigrationStart(migrationNum int) error {
	// Format migration number with leading zeros for consistent display
	formattedNum := fmt.Sprintf("%03d", migrationNum)

	// Read existing .env file
	content, err := os.ReadFile(".env")
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error reading .env file: %v", err)
	}

	// Parse existing content
	lines := strings.Split(string(content), "\n")
	newLines := []string{}
	migrationStartFound := false

	// Update existing MIGRATION_START line if found
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
		if key == "MIGRATION_START" {
			newLines = append(newLines, fmt.Sprintf("MIGRATION_START=%s", formattedNum))
			migrationStartFound = true
		} else {
			newLines = append(newLines, line)
		}
	}

	// Add MIGRATION_START if not found
	if !migrationStartFound {
		newLines = append(newLines, fmt.Sprintf("MIGRATION_START=%s", formattedNum))
	}

	// Write back to .env file
	return os.WriteFile(".env", []byte(strings.Join(newLines, "\n")), 0644)
}

// UpdateMigrationStart updates the MIGRATION_START value in the .env file (exported version)
func UpdateMigrationStart(migrationNum int) error {
	// Format migration number with leading zeros
	formattedNum := fmt.Sprintf("%03d", migrationNum)

	// Update the environment variable in the current process
	os.Setenv("MIGRATION_START", formattedNum)

	// Update the .env file
	return updateMigrationStart(migrationNum)
}

// RunMigrations runs the database migrations
func RunMigrations() error {
	return applySchema(db)
}

// GetDB returns the database connection pool
func GetDB() *sql.DB {
	return db
}

// InitDB initializes the database connection
func InitDB() (*sql.DB, error) {
	// Get individual connection parameters from environment
	driver := os.Getenv("DB_DRIVER")
	if driver == "" {
		driver = "postgres" // Default to postgres if not specified
	}

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	// Validate required parameters
	if host == "" || port == "" || user == "" || password == "" || dbname == "" {
		return nil, fmt.Errorf("database connection parameters are not set (DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME required)")
	}

	var dsn string
	switch driver {
	case "postgres":
		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbname)
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
			user, password, host, port, dbname)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", driver)
	}

	// Open a direct connection to the database
	sqlDB, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test the connection
	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Successfully connected to database using %s driver", driver)
	return sqlDB, nil
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
func (s *UserService) GetUserByEmail(email string) (auth.User, error) {
	var user auth.User
	query := `SELECT id, email, is_admin, created_at, last_logged_in FROM ai.users WHERE email = $1`
	err := s.db.QueryRow(query, email).Scan(
		&user.ID, &user.Email, &user.IsAdmin, &user.CreatedAt, &user.LastLoggedIn)
	if err != nil {
		if err == sql.ErrNoRows {
			return auth.User{}, fmt.Errorf("user not found")
		}
		return auth.User{}, err
	}
	return user, nil
}

// CreateUser creates a new user
func (s *UserService) CreateUser(email string) (auth.User, error) {
	// Check if any users exist (first user is admin)
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM ai.users").Scan(&count)
	if err != nil {
		return auth.User{}, err
	}

	isAdmin := count == 0 // First user is admin

	var user auth.User
	query := `
		INSERT INTO ai.users (email, is_admin) 
		VALUES ($1, $2) 
		RETURNING id, email, is_admin, created_at
	`
	err = s.db.QueryRow(query, email, isAdmin).Scan(&user.ID, &user.Email, &user.IsAdmin, &user.CreatedAt)
	if err != nil {
		return auth.User{}, err
	}

	return user, nil
}

// GetOrCreateUser retrieves a user by email or creates a new one
func (s *UserService) GetOrCreateUser(email string) (auth.User, bool, error) {
	user, err := s.GetUserByEmail(email)
	if err != nil {
		if err.Error() == "user not found" {
			newUser, err := s.CreateUser(email)
			if err != nil {
				return auth.User{}, false, err
			}
			return newUser, true, nil // true = created new user
		}
		return auth.User{}, false, err
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

// NewProjectService creates a new ProjectService
func NewProjectService(db *sql.DB) *ProjectService {
	return &ProjectService{db: db}
}

// CreateProject creates a new project
func (s *ProjectService) CreateProject(ctx context.Context, name, description, domainName string, createdBy *auth.User) (Project, error) {
	var project Project
	query := `
		INSERT INTO ai.projects (name, description, domain_name, created_by, created_at) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id, name, description, domain_name, created_by, created_at
	`
	now := time.Now()
	err := s.db.QueryRowContext(ctx, query, name, description, domainName, createdBy.ID, now).Scan(
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

// GetProjectMembers retrieves members of a project
func (s *ProjectService) GetProjectMembers(ctx context.Context, projectID int) ([]auth.User, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT u.id, u.email, u.is_admin, u.created_at, u.last_logged_in 
		FROM ai.users u
		JOIN ai.project_members pm ON u.id = pm.user_id
		WHERE pm.project_id = $1
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []auth.User
	for rows.Next() {
		var user auth.User
		if err := rows.Scan(&user.ID, &user.Email, &user.IsAdmin, &user.CreatedAt, &user.LastLoggedIn); err != nil {
			return nil, err
		}
		members = append(members, user)
	}
	return members, nil
}

// AddProjectMember adds a user to a project
func (s *ProjectService) AddProjectMember(ctx context.Context, projectID int, user *auth.User) error {
	query := `INSERT INTO ai.project_members (project_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := s.db.ExecContext(ctx, query, projectID, user.ID)
	return err
}

// RemoveProjectMember removes a user from a project
func (s *ProjectService) RemoveProjectMember(ctx context.Context, projectID int, user *auth.User) error {
	query := `DELETE FROM ai.project_members WHERE project_id = $1 AND user_id = $2`
	_, err := s.db.ExecContext(ctx, query, projectID, user.ID)
	return err
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
		SELECT id, managed_table_id, name, display_name, data_type, ordinal, visible, created_at 
		FROM ai.managed_columns 
		WHERE managed_table_id = $1
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
		if err := rows.Scan(&c.ID, &c.ManagedTableID, &c.Name, &displayName, &c.DataType, &c.Ordinal, &c.Visible, &c.CreatedAt); err != nil {
			return nil, err
		}
		if displayName.Valid {
			c.DisplayName = displayName.String
		} else {
			c.DisplayName = c.Name
		}
		columns = append(columns, c)
	}
	return columns, nil
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
				TableName:  tableName,
				SchemaName: schema,
				IsManaged:  isManaged,
			}

			if isManaged {
				metadata.Description = managedTable.Description
				metadata.ManagedTableID = managedTable.ID
				metadata.ReadOnly = managedTable.ReadOnly
				metadata.Initialized = managedTable.Initialized
			}

			tables = append(tables, metadata)
		}
		tableRows.Close()
	}

	return tables, nil
}

// GetTableColumns returns all columns for a database table with management status
func (s *DatabaseMetadataService) GetTableColumns(ctx context.Context, projectDBID int, tableID int, schema, tableName string) ([]ColumnMetadata, error) {
	managedColumns := make(map[string]ManagedColumn)
	if tableID > 0 {
		rows, err := s.db.QueryContext(ctx, `
			SELECT id, name, display_name, data_type, ordinal, visible 
			FROM ai.managed_columns 
			WHERE managed_table_id = $1
			ORDER BY ordinal
		`, tableID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var c ManagedColumn
			var displayName sql.NullString
			if err := rows.Scan(&c.ID, &c.Name, &displayName, &c.DataType, &c.Ordinal, &c.Visible); err != nil {
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
			ordinal_position,
			column_default
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
		var defaultValue sql.NullString
		var isNullableBool bool
		if err := rows.Scan(&col.ColumnName, &col.DataType, &isNullableBool, &col.OrdinalPos, &defaultValue); err != nil {
			return nil, err
		}
		if isNullableBool {
			col.IsNullable = "YES"
		} else {
			col.IsNullable = "NO"
		}
		if defaultValue.Valid {
			col.DefaultValue = defaultValue.String
		}

		// Check if column is managed
		managedColumn, isManaged := managedColumns[col.ColumnName]
		col.IsManaged = isManaged

		if isManaged {
			col.DisplayName = managedColumn.DisplayName
			col.Visible = managedColumn.Visible
			col.ManagedColumnID = managedColumn.ID
			col.SystemType = managedColumn.ColumnType
		} else {
			col.DisplayName = col.ColumnName
			col.Visible = true
			col.SystemType = mapDbTypeToColumnType(col.DataType)
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

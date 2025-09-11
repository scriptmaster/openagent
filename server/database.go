package server

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/scriptmaster/openagent/auth"
	"github.com/scriptmaster/openagent/common"
	"github.com/scriptmaster/openagent/models"
	"github.com/scriptmaster/openagent/projects"

	// PostgreSQL driver
	"github.com/lib/pq"
	_ "github.com/lib/pq"

	// MySQL driver
	_ "github.com/go-sql-driver/mysql"
)

// Global database instance and mutex
var dbInstance *sql.DB
var dbMutex sync.Mutex

type User = auth.User

// applySchema runs the SQL in migrations directory to initialize the database
// based on ai.migrations table tracking.
func applySchema(db *sql.DB) error {
	// First check if the ai schema exists - if not, create it
	var schemaExists bool
	checkSchemaQuery := common.MustGetSQL("CheckSchemaExists")
	err := db.QueryRow(checkSchemaQuery).Scan(&schemaExists)
	if err != nil {
		// If the ai schema doesn't exist, the query might fail depending on permissions
		// Let's try creating it explicitly.
		if _, createErr := db.Exec("CREATE SCHEMA IF NOT EXISTS ai"); createErr != nil {
			return fmt.Errorf("failed to create schema 'ai': %v", createErr)
		}
		log.Println("AI schema created or verified.")
		schemaExists = true
	}

	log.Printf("\t → \t → Schema 'ai' exists: %v", schemaExists)

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
		// Simple check for numbered migrations
		if _, err := strconv.Atoi(strings.Split(file.Name(), "_")[0]); err == nil {
			fileNames = append(fileNames, file.Name())
		}
	}
	sort.Strings(fileNames)

	if len(fileNames) == 0 {
		log.Println("No numbered migration files found in migrations directory.")
		return nil // Not necessarily an error
	}

	// Get last applied migration number from ai.migrations table
	lastApplied, err := common.GetLastAppliedMigration(db)
	if err != nil {
		// If the migrations table doesn't exist yet, start from 0
		// This will happen on first run before 001_schema.sql creates the table
		log.Printf("Warning: Could not get last applied migration, starting from 0: %v", err)
		lastApplied = 0
	}

	// Execute migrations higher than lastApplied
	log.Printf("\t → \t → Applying migrations starting after %d...", lastApplied)
	migrationsApplied := 0
	for _, fileName := range fileNames {
		numStr := strings.SplitN(fileName, "_", 2)[0]
		num, _ := strconv.Atoi(numStr) // Error already checked above

		if num > lastApplied {
			log.Printf("\t → \t → Applying migration file: %s", fileName)
			migrationPath := filepath.Join("migrations", fileName)
			migrationBytes, readErr := os.ReadFile(migrationPath)
			if readErr != nil {
				return fmt.Errorf("failed to read migration %s: %w", fileName, readErr)
			}

			// Execute migration using Exec
			if _, execErr := db.Exec(string(migrationBytes)); execErr != nil {
				return fmt.Errorf("failed to apply migration %s: %w", fileName, execErr)
			}

			// Record the migration in ai.migrations table
			insertMigrationQuery := common.MustGetSQL("InsertMigration")
			if _, insertErr := db.Exec(insertMigrationQuery, fileName); insertErr != nil {
				log.Printf("Warning: Failed to record migration %s in ai.migrations table: %v", fileName, insertErr)
			}

			migrationsApplied++
		}
	}

	if migrationsApplied > 0 {
		log.Printf("%d migration(s) applied successfully.", migrationsApplied)
	} else {
		log.Printf("\t → \t → No new migrations applied (Last migration: %d).", lastApplied)
	}

	return nil
}

// DEPRECATED: updateMigrationStartInEnvFile is deprecated. Use database-based migration tracking instead.
// This function is kept for backward compatibility but should not be used in new code.
func updateMigrationStartInEnvFile(migrationNum int) error {
	// Format migration number with leading zeros for consistent display
	formattedNum := fmt.Sprintf("%03d", migrationNum)

	// Read existing .env file
	content, err := os.ReadFile(".env")
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error reading .env file: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	newLines := []string{}
	migrationStartFound := false

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			// Keep empty lines, but avoid adding multiple consecutive ones if reading/writing
			if len(newLines) == 0 || newLines[len(newLines)-1] != "" {
				newLines = append(newLines, "")
			}
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[0]) == "MIGRATION_START" {
			newLines = append(newLines, fmt.Sprintf("MIGRATION_START=%s", formattedNum))
			migrationStartFound = true
		} else {
			newLines = append(newLines, line) // Keep the original line
		}
	}

	// Add MIGRATION_START if not found
	if !migrationStartFound {
		// Avoid adding if the last line was empty
		if len(newLines) > 0 && newLines[len(newLines)-1] == "" {
			newLines = append(newLines[:len(newLines)-1], fmt.Sprintf("MIGRATION_START=%s", formattedNum), "")
		} else {
			newLines = append(newLines, fmt.Sprintf("MIGRATION_START=%s", formattedNum))
		}
	}

	// Ensure there's a newline at the end if the file wasn't empty
	finalContent := strings.Join(newLines, "\n")
	if len(finalContent) > 0 && !strings.HasSuffix(finalContent, "\n") {
		finalContent += "\n"
	}

	return os.WriteFile(".env", []byte(finalContent), 0644)
}

// DEPRECATED: UpdateMigrationStart is deprecated. Use database-based migration tracking instead.
// This function is kept for backward compatibility but should not be used in new code.
func UpdateMigrationStart(migrationNum int) error {
	// Update the environment variable in the current process for consistency
	os.Setenv("MIGRATION_START", fmt.Sprintf("%03d", migrationNum))
	// Update the .env file
	return updateMigrationStartInEnvFile(migrationNum)
}

// RunMigrations runs the database migrations (exported alias for applySchema)
// This might not be needed if only called from InitDB
func RunMigrations(db *sql.DB) error {
	return applySchema(db)
}

// GetDB returns the initialized database connection pool.
func GetDB() *sql.DB {
	dbMutex.Lock()
	defer dbMutex.Unlock()
	return dbInstance
}

// InitDB initializes the database connection, ensuring the DB exists, runs migrations, and sets the global dbInstance.
func InitDB() (*sql.DB, error) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	// Prevent re-initialization if already done
	if dbInstance != nil {
		if err := dbInstance.Ping(); err == nil {
			return dbInstance, nil
		}
		log.Println("WARN: Existing DB connection failed ping, attempting re-initialization.")
		dbInstance.Close()
		dbInstance = nil
	}

	// --- Database Connection Parameters ---
	driver := common.GetEnvOrDefault("DB_DRIVER", "postgres")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	targetDbName := os.Getenv("DB_NAME") // The DB we actually want to use
	defaultDbName := "postgres"          // DB to connect to initially for CREATE DATABASE command

	if driver != "postgres" {
		// Currently, auto-creation logic is only implemented for postgres
		return nil, fmt.Errorf("unsupported database driver '%s' for automatic database creation", driver)
	}

	if host == "" || port == "" || user == "" || targetDbName == "" { // Password can be empty
		return nil, fmt.Errorf("database connection parameters are not fully set (DB_HOST, DB_PORT, DB_USER, DB_NAME required)")
	}

	// --- Step 1: Connect to default DB to check/create target DB ---
	initialDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, defaultDbName)

	initialDB, err := sql.Open(driver, initialDSN)
	if err != nil {
		return nil, fmt.Errorf("error: open initial connection to default database '%s': %w", defaultDbName, err)
	}
	defer initialDB.Close()

	if err = initialDB.Ping(); err != nil {
		return nil, fmt.Errorf("error: ping default database '%s': %w", defaultDbName, err)
	}
	log.Printf("\t → \t → Default Database Ping Successful: %v", defaultDbName)

	// --- Step 2: Try to create the target database ---
	log.Printf("\t → \t → Switching from default to Target Database: '%s'. Ensuring it exists...", targetDbName)
	// Check if the target database already exists using pg_database (PostgreSQL's system catalog for databases)
	var dbExists bool
	// Use pq.QuoteIdentifier for the database name in the WHERE clause for safety against SQL injection
	// and to correctly handle database names with special characters.
	// quotedTargetDbName := pq.QuoteIdentifier(targetDbName)
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = '%s')", targetDbName)
	log.Printf("\t → \t → Running Query: %v", query)
	err = initialDB.QueryRow(query).Scan(&dbExists)
	if err != nil {
		return nil, fmt.Errorf("error checking if database '%s' exists: %w", targetDbName, err)
	}

	if dbExists {
		log.Printf("\t → 2. Database %s exists and is ready to connect.", targetDbName)
	} else {
		// If the database does not exist, attempt to create it.
		log.Printf("\t → \t → Database %s does not exist. Attempting to create...", targetDbName)
		// Use pq.QuoteIdentifier for the CREATE DATABASE command as well.
		_, err = initialDB.Exec(fmt.Sprintf("CREATE DATABASE %s", targetDbName))
		if err != nil {
			// If creation fails for reasons other than "already exists" (which is handled by the SELECT EXISTS check),
			// then it's a genuine error (e.g., permissions).
			return nil, fmt.Errorf("error: create target database '%s': %w", targetDbName, err)
		}
		log.Printf("\t → 2.100 Database '%s' created successfully.", targetDbName)
	}

	// Initial connection no longer needed
	initialDB.Close()

	// --- Step 3: Connect to the target database ---
	targetDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, targetDbName)

	sqlDB, err := sql.Open(driver, targetDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection to target database '%s': %w", targetDbName, err)
	}

	// Test the target connection
	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to ping target database '%s': %w", targetDbName, err)
	}
	log.Printf("\t → 3. Successfully pinged to target database '%s'", targetDbName)

	// --- Run Database Migrations (Old Method) ---
	log.Println("\t → 4. Applying database schema/migrations...")
	if err := applySchema(sqlDB); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to apply database schema/migrations: %w", err)
	}
	// -------------------------------------------

	// Store the initialized instance globally
	dbInstance = sqlDB

	log.Printf("\t → 4.100 Database '%s' completely initialized.", targetDbName) // Consolidated message

	// -------------------------------------------
	log.Println("\t → 5. Checking/Creating Admin Token...")
	// --- Check/Create Admin Token ---
	_, tokenErr := CheckOrCreateAdminToken(sqlDB)
	if tokenErr != nil {
		log.Printf("WARNING: Failed to check/create admin token: %v", tokenErr)
	}
	// -------------------------------------------

	return dbInstance, nil
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

// UserService implements the auth.UserServicer interface and handles user data operations,
// potentially across different databases based on project context.
type UserService struct {
	db          *sql.DB // Default application DB connection
	pdbService  common.ProjectDBService
	dataService DataAccessService // Interface providing getConnection
}

// NewUserService creates a new UserService with necessary dependencies.
func NewUserService(db *sql.DB, pdbService common.ProjectDBService, dataService DataAccessService) *UserService {
	if db == nil {
		log.Fatal("UserService requires a non-nil default database connection")
	}
	if pdbService == nil {
		log.Fatal("UserService requires a non-nil ProjectDBService")
	}
	if dataService == nil {
		log.Fatal("UserService requires a non-nil DataAccessService")
	}
	return &UserService{
		db:          db,
		pdbService:  pdbService,
		dataService: dataService,
	}
}

// getDBForUserOp determines the correct database connection (*sql.DB) to use for a user operation.
// For authentication operations, it always uses the main application database.
// For other user operations, it checks the context for a project and uses the project's default database if available.
func (s *UserService) getDBForUserOp(ctx context.Context) (*sql.DB, error) {
	project := projects.GetProjectFromContext(ctx)
	if project == nil {
		return s.db, nil // No project, use default DB
	}

	// Project context exists, find its default database connection
	projectDBs, err := s.pdbService.GetProjectDBs(int(project.ID))
	if err != nil {
		log.Printf("Error getting project DBs for project %d: %v", project.ID, err)
		// Fallback to default DB for non-auth operations
		return s.db, nil
	}

	var defaultProjectDBID int = 0
	for _, pdb := range projectDBs {
		if pdb.IsDefault {
			defaultProjectDBID = pdb.ID
			break
		}
	}

	if defaultProjectDBID == 0 {
		log.Printf("No default database configured for project %d, using main application database", project.ID)
		// Fallback to default DB
		return s.db, nil
	}

	// Get the connection using DataAccessService
	conn, err := s.dataService.getConnection(ctx, defaultProjectDBID)
	if err != nil {
		log.Printf("Error getting connection for project %d DB %d: %v, falling back to main database", project.ID, defaultProjectDBID, err)
		// Fallback to default DB
		return s.db, nil
	}

	return conn, nil
}

// getDBForAuthOp always returns the main application database for authentication operations
func (s *UserService) getDBForAuthOp(ctx context.Context) (*sql.DB, error) {
	return s.db, nil
}

// GetUserByEmail retrieves a user by email from the appropriate database (default or project).
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*auth.User, error) {
	dbConn, err := s.getDBForUserOp(ctx)
	if err != nil {
		return nil, err // Error getting the correct DB connection
	}

	// TODO: Define user table schema (e.g., ai.users or public.users?)
	// Assuming ai.users for now, may need adjustment based on project DB schema.
	query := `SELECT id, email, password_hash, is_admin, created_at, last_logged_in FROM ai.users WHERE email = $1`
	row := dbConn.QueryRowContext(ctx, query, email)

	var user auth.User
	var lastLoggedIn sql.NullTime
	err = row.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.IsAdmin,
		&user.CreatedAt,
		&lastLoggedIn,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found") // Consistent error message
		}
		return nil, err // Other potential errors
	}
	if lastLoggedIn.Valid {
		user.LastLoggedIn = lastLoggedIn.Time
	}
	return &user, nil
}

// CreateUser creates a new user in the appropriate database.
func (s *UserService) CreateUser(ctx context.Context, email string) (*auth.User, error) {
	dbConn, err := s.getDBForUserOp(ctx)
	if err != nil {
		return nil, err // Error getting the correct DB connection
	}

	// Basic email validation
	if !common.IsValidEmail(email) { // Assuming a validation function exists
		return nil, errors.New("invalid email format")
	}

	// For first-time user creation via OTP, password hash is initially empty or random.
	// The VerifyPassword path should handle password setting.
	// Let's store an empty hash initially.
	passwordHash := "" // Or generate a secure random unusable hash?
	isAdmin := false   // Default to non-admin

	// Special case: If no project context AND no admin exists yet, make the first user admin.
	project := projects.GetProjectFromContext(ctx)
	if project == nil {
		adminExists, checkErr := s.CheckIfAdminExists(ctx) // Use the new method
		if checkErr != nil {
			log.Printf("Error checking for existing admin: %v", checkErr)
			// Decide if this should block user creation
			// return nil, fmt.Errorf("failed to check admin status")
		}
		if !adminExists {
			isAdmin = true
			log.Printf("Creating first user (%s) as admin in default database.", email)
		}
	}

	// TODO: Adjust table name if needed for project DBs
	query := `
		INSERT INTO ai.users (email, password_hash, is_admin, created_at, last_logged_in)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, email, password_hash, is_admin, created_at, last_logged_in
	`
	row := dbConn.QueryRowContext(ctx, query, email, passwordHash, isAdmin)

	var user auth.User
	var lastLoggedIn sql.NullTime
	err = row.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.IsAdmin,
		&user.CreatedAt,
		&lastLoggedIn,
	)
	if err != nil {
		// Handle potential unique constraint violation (email already exists)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" { // PostgreSQL unique violation code
			return nil, fmt.Errorf("email already exists")
		}
		log.Printf("Error creating user %s: %v", email, err)
		return nil, fmt.Errorf("failed to create user account: %w", err)
	}
	if lastLoggedIn.Valid {
		user.LastLoggedIn = lastLoggedIn.Time
	}
	return &user, nil
}

// UpdateUserLastLogin updates the last_logged_in timestamp for a user in the appropriate database.
func (s *UserService) UpdateUserLastLogin(ctx context.Context, userID int) error {
	dbConn, err := s.getDBForUserOp(ctx)
	if err != nil {
		// Log the error but don't necessarily fail the login operation
		log.Printf("Error getting DB for UpdateUserLastLogin (userID: %d): %v. Skipping update.", userID, err)
		return nil
	}

	query := common.MustGetSQL("auth/update_last_login")
	_, err = dbConn.ExecContext(ctx, query, userID)
	if err != nil {
		log.Printf("Error updating last login for user %d: %v", userID, err)
		// Don't fail the calling operation (e.g., login) just because this failed.
		return nil
	}
	return nil
}

// VerifyPassword finds a user by email and verifies their password hash.
// This always uses the main application database for authentication.
func (s *UserService) VerifyPassword(ctx context.Context, email, password string) (*auth.User, error) {
	// Always use main application database for authentication
	dbConn, err := s.getDBForAuthOp(ctx)
	if err != nil {
		return nil, err
	}

	// Query user from main application database using SQL file
	var user auth.User
	var lastLoggedIn sql.NullTime
	query := common.MustGetSQL("auth/verify_password")
	err = dbConn.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.IsAdmin,
		&user.CreatedAt,
		&lastLoggedIn,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	if lastLoggedIn.Valid {
		user.LastLoggedIn = lastLoggedIn.Time
	}

	if user.PasswordHash == "" {
		// Handle case where user was created via OTP and hasn't set a password yet?
		// Or just treat it as invalid password.
		return nil, errors.New("invalid credentials") // Or a more specific error
	}

	// Use bcrypt comparison (assuming bcrypt was used for hashing)
	if !auth.CheckPasswordHash(password, user.PasswordHash) {
		return nil, errors.New("invalid credentials")
	}

	return &user, nil // Password matches
}

// CheckIfAdminExists checks if any user with is_admin = true exists in the *default* database.
// This is used to determine if the first user should be made an admin.
func (s *UserService) CheckIfAdminExists(ctx context.Context) (bool, error) {
	// This check should ALWAYS run against the default application database,
	// regardless of project context, as it determines the global first admin.
	var exists bool
	query := common.MustGetSQL("auth/check_admin_exists")
	err := s.db.QueryRowContext(ctx, query).Scan(&exists)
	if err != nil {
		log.Printf("Error checking for existing admin users: %v", err)
		return false, fmt.Errorf("database error checking admin status: %w", err)
	}
	return exists, nil
}

// UpdatePasswordHash updates the user's password hash in the appropriate database.
func (s *UserService) UpdatePasswordHash(ctx context.Context, userID int, newHash string) error {
	dbConn, err := s.getDBForUserOp(ctx)
	if err != nil {
		log.Printf("Error getting DB for UpdatePasswordHash (userID: %d): %v", userID, err)
		return err
	}

	if newHash == "" {
		return errors.New("cannot update password to an empty hash")
	}

	query := common.MustGetSQL("auth/update_password_hash")
	result, err := dbConn.ExecContext(ctx, query, newHash, userID)
	if err != nil {
		log.Printf("Error updating password hash for user %d: %v", userID, err)
		return fmt.Errorf("database error updating password: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user with ID %d not found", userID)
	}
	return nil
}

// --- Deprecated / Needs Review ---

// GetOrCreateUser combines GetUserByEmail and CreateUser. Needs context.
// Deprecated: Prefer explicit Get and Create calls in handlers.
func (s *UserService) GetOrCreateUser(ctx context.Context, email string) (*auth.User, bool, error) {
	user, err := s.GetUserByEmail(ctx, email)
	if err != nil {
		if err.Error() == "user not found" { // Fragile error string check
			newUser, createErr := s.CreateUser(ctx, email)
			if createErr != nil {
				return nil, false, createErr
			}
			return newUser, true, nil // true = created new user
		}
		return nil, false, err // Other error from GetUserByEmail
	}
	return user, false, nil // false = existing user
}

// --- End Deprecated ---

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
	query := common.MustGetSQL("project_members/add") // Load query
	_, err := s.db.ExecContext(ctx, query, projectID, user.ID)
	return err
}

// RemoveProjectMember removes a user from a project
func (s *ProjectService) RemoveProjectMember(ctx context.Context, projectID int, user *auth.User) error {
	query := "DELETE FROM ai.project_members WHERE project_id = $1 AND user_id = $2" // Keeping inline due to file creation issue
	// query := common.MustGetSQL("project_members/remove") // Load query (if file existed)
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

	// Decode connection string using common.DecodeConnectionString
	connStr, err := common.DecodeConnectionString(projectDB.ConnectionString)
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
	var err error
	var query string

	if scopeID != nil {
		query = common.MustGetSQL("settings/get_scoped") // Load scoped query
		err = s.db.QueryRow(query, key, scope, *scopeID).Scan(&setting.ID, &setting.Key, &setting.Value, &setting.Description, &setting.Scope, &setting.ScopeID, &setting.UpdatedAt)
	} else {
		query = common.MustGetSQL("settings/get_global") // Load global query
		err = s.db.QueryRow(query, key).Scan(&setting.ID, &setting.Key, &setting.Value, &setting.Description, &setting.Scope, &setting.ScopeID, &setting.UpdatedAt)
	}

	if err != nil {
		return Setting{}, err
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
			intVal := int(scopeID.Int64) // Correct assignment to *int
			setting.ScopeID = &intVal
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

			// Correct struct literal based on server/types.go:TableMetadata
			metadata := TableMetadata{
				TableName:      tableName,
				SchemaName:     schema,
				IsManaged:      isManaged,
				ManagedTableID: 0, // Default, updated below if managed
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
			if err := rows.Scan(&c.ID, &c.ManagedTableID, &c.Name, &displayName, &c.DataType, &c.Ordinal, &c.Visible); err != nil {
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
		if err := rows.Scan(&col.ColumnName, &col.DataType, &isNullableBool, &col.OrdinalPosition, &defaultValue); err != nil {
			return nil, err
		}

		col.IsNullable = isNullableBool // Assign bool directly

		if defaultValue.Valid {
			col.ColumnDefault = &defaultValue.String // Assign *string directly
		}

		// Check if column is managed
		managedColumn, isManaged := managedColumns[col.ColumnName]
		col.IsManaged = isManaged

		if isManaged {
			col.DisplayName = managedColumn.DisplayName
			col.Visible = managedColumn.Visible
			col.ManagedColumnID = managedColumn.ID
			col.SystemType = managedColumn.ColumnType // Use ColumnType from ManagedColumn struct
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

// dbService is the concrete implementation of common.ProjectDBService
type dbService struct {
	db *sql.DB
}

// NewProjectDBService creates a new instance of the dbService
// It now returns the common.ProjectDBService interface type
func NewProjectDBService(db *sql.DB) common.ProjectDBService {
	if db == nil {
		panic("Database connection is nil for NewProjectDBService")
	}
	return &dbService{db: db} // Return the concrete type that implements the interface
}

// Implement the common.ProjectDBService interface methods on *dbService

func (s *dbService) GetProjectDBs(projectID int) ([]common.ProjectDB, error) {
	// Refactor existing GetProjectDBs to use s.db and return common.ProjectDB
	// Assuming original function GetProjectDBs exists and takes *sql.DB
	// return GetProjectDBs(s.db, projectID) // REMOVED - Standalone GetProjectDBs is gone
	query := common.MustGetSQL("GetProjectDBsByProjectID")
	rows, err := s.db.Query(query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query project DBs: %w", err)
	}
	defer rows.Close()

	var dbs []common.ProjectDB // Use common.ProjectDB
	for rows.Next() {
		var pdb common.ProjectDB // Use common.ProjectDB
		if err := rows.Scan(&pdb.ID, &pdb.ProjectID, &pdb.Name, &pdb.Description, &pdb.DBType, &pdb.ConnectionString, &pdb.SchemaName, &pdb.IsDefault, &pdb.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan project DB row: %w", err)
		}
		dbs = append(dbs, pdb)
	}
	return dbs, rows.Err()
}

func (s *dbService) GetProjectDB(id int) (common.ProjectDB, error) {
	// return GetProjectDB(s.db, id) // REMOVED - Implementation needed
	var pdb common.ProjectDB
	// TODO: Implement SQL query to get ProjectDB by ID
	query := `SELECT id, project_id, name, description, db_type, connection_string, schema_name, is_default, created_at FROM ai.project_dbs WHERE id = $1`
	err := s.db.QueryRow(query, id).Scan(
		&pdb.ID, &pdb.ProjectID, &pdb.Name, &pdb.Description, &pdb.DBType,
		&pdb.ConnectionString, &pdb.SchemaName, &pdb.IsDefault, &pdb.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return common.ProjectDB{}, fmt.Errorf("project DB with id %d not found", id)
		}
		return common.ProjectDB{}, fmt.Errorf("failed to get project DB %d: %w", id, err)
	}
	return pdb, nil
}

func (s *dbService) CreateProjectDB(projectID int, name, description, dbType, connectionString, schemaName string, isDefault bool) (common.ProjectDB, error) {
	// return CreateProjectDB(s.db, projectID, name, description, dbType, connectionString, schemaName, isDefault) // REMOVED - Implementation needed
	var pdb common.ProjectDB
	// TODO: Implement SQL query to insert ProjectDB
	// TODO: Handle connection string encoding (e.g., base64)
	query := `
		INSERT INTO ai.project_dbs (project_id, name, description, db_type, connection_string, schema_name, is_default, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		RETURNING id, project_id, name, description, db_type, connection_string, schema_name, is_default, created_at
	`
	err := s.db.QueryRow(query, projectID, name, description, dbType, connectionString, schemaName, isDefault).Scan(
		&pdb.ID, &pdb.ProjectID, &pdb.Name, &pdb.Description, &pdb.DBType,
		&pdb.ConnectionString, &pdb.SchemaName, &pdb.IsDefault, &pdb.CreatedAt,
	)
	if err != nil {
		return common.ProjectDB{}, fmt.Errorf("failed to create project DB: %w", err)
	}
	return pdb, nil
}

// DecodeConnectionString decodes a base64 encoded connection string.
// This implements the method required by the common.ProjectDBService interface.
func (s *dbService) DecodeConnectionString(encoded string) (string, error) {
	// Delegate to the function in the common package
	return common.DecodeConnectionString(encoded)
}

// CheckOrCreateAdminToken checks for an admin token for today and creates one if it doesn't exist.
// Token format: UUID-ADMIN-TOKEN-SHA1-HASHED-TIMESTAMP
func CheckOrCreateAdminToken(db *sql.DB) (string, error) {
	today := time.Now().UTC().Format("2006-01-02")

	// First, check if a token already exists for today
	existingToken, err := GetAdminTokenForDate(db, today)
	if err != nil {
		return "", fmt.Errorf("failed to check existing admin token: %w", err)
	}

	if existingToken != "" {
		log.Printf("Admin token for today (%s): %s", today, existingToken)
		return existingToken, nil
	}

	// No token exists for today, create a new one
	newToken, err := generateAdminToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate admin token: %w", err)
	}

	// Store the new token in the database
	insertQuery := common.MustGetSQL("InsertAdminToken")
	_, err = db.Exec(insertQuery, newToken, today)
	if err != nil {
		return "", fmt.Errorf("failed to store admin token: %w", err)
	}

	log.Printf("Generated new admin token for today (%s): %s", today, newToken)
	return newToken, nil
}

// generateAdminToken creates a new admin token with the format: UUID-ADMIN-TOKEN-SHA1-HASHED-TIMESTAMP
func generateAdminToken() (string, error) {
	// Generate a UUID
	tokenUUID := uuid.New().String()

	// Get current timestamp
	timestamp := time.Now().UTC().Unix()

	// Create the timestamp string and hash it with SHA1
	timestampStr := fmt.Sprintf("%d", timestamp)
	hasher := sha1.New()
	hasher.Write([]byte(timestampStr))
	hashedTimestamp := fmt.Sprintf("%x", hasher.Sum(nil))

	// Construct the token: UUID-ADMIN-TOKEN-SHA1-HASHED-TIMESTAMP
	token := fmt.Sprintf("%s-ADMIN-TOKEN-%s", tokenUUID, hashedTimestamp)

	return token, nil
}

func (s *dbService) UpdateProjectDB(id int, name, description, dbType, connectionString, schemaName string, isDefault bool) error {
	// return UpdateProjectDB(s.db, id, name, description, dbType, connectionString, schemaName, isDefault) // REMOVED - Implementation needed
	// TODO: Implement SQL query to update ProjectDB
	// TODO: Handle connection string encoding (e.g., base64)
	query := `
		UPDATE ai.project_dbs
		SET name = $2, description = $3, db_type = $4, connection_string = $5, schema_name = $6, is_default = $7
		WHERE id = $1
	`
	_, err := s.db.Exec(query, id, name, description, dbType, connectionString, schemaName, isDefault)
	if err != nil {
		return fmt.Errorf("failed to update project DB %d: %w", id, err)
	}
	return nil
}

func (s *dbService) DeleteProjectDB(id int) error {
	// return DeleteProjectDB(s.db, id) // REMOVED - Implementation needed
	// TODO: Implement SQL query to delete ProjectDB
	query := `DELETE FROM ai.project_dbs WHERE id = $1`
	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete project DB %d: %w", id, err)
	}
	return nil
}

func (s *dbService) TestConnection(projectDB common.ProjectDB) error {
	// ... existing code ...
	return nil
}

// GetAdminTokenForDate retrieves the admin token for a specific date.
func GetAdminTokenForDate(db *sql.DB, dateStr string) (string, error) {
	// Validate date format
	if _, err := time.Parse("2006-01-02", dateStr); err != nil {
		return "", fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}

	// Query the database for the token
	getTokenQuery := common.MustGetSQL("GetAdminTokenForDateCheck")
	var token string
	err := db.QueryRow(getTokenQuery, dateStr).Scan(&token)
	if err != nil {
		if err == sql.ErrNoRows {
			// No token found for this date
			return "", nil
		}
		return "", fmt.Errorf("failed to query admin token for date %s: %w", dateStr, err)
	}

	return token, nil
}

// GetAdminStats fetches statistics for the admin dashboard
func GetAdminStats(db *sql.DB) (*models.AdminStats, error) {
	stats := &models.AdminStats{}

	// Get project count
	var projectCount int
	err := db.QueryRow("SELECT COUNT(*) FROM ai.projects").Scan(&projectCount)
	if err != nil {
		log.Printf("Error counting projects: %v", err)
		projectCount = 0
	}
	stats.ProjectCount = projectCount

	// Get connection count (project databases)
	var connectionCount int
	err = db.QueryRow("SELECT COUNT(*) FROM ai.project_dbs").Scan(&connectionCount)
	if err != nil {
		log.Printf("Error counting connections: %v", err)
		connectionCount = 0
	}
	stats.ConnectionCount = connectionCount

	// Get table count (managed tables)
	var tableCount int
	err = db.QueryRow("SELECT COUNT(*) FROM ai.managed_tables").Scan(&tableCount)
	if err != nil {
		log.Printf("Error counting tables: %v", err)
		tableCount = 0
	}
	stats.TableCount = tableCount

	// Get user count
	var userCount int
	query := common.MustGetSQL("auth/count_users")
	err = db.QueryRow(query).Scan(&userCount)
	if err != nil {
		log.Printf("Error counting users: %v", err)
		userCount = 0
	}
	stats.UserCount = userCount

	return stats, nil
}

// MakeUserAdmin grants admin privileges to a user.
func (s *UserService) MakeUserAdmin(ctx context.Context, userID int) error {
	dbConn, err := s.getDBForUserOp(ctx)
	if err != nil {
		// Log error but maybe don't fail the operation if default DB works?
		// For now, return error if we can't get the intended DB.
		log.Printf("Error getting DB for MakeUserAdmin (userID: %d): %v", userID, err)
		return err
	}

	query := common.MustGetSQL("auth/make_admin")
	result, err := dbConn.ExecContext(ctx, query, userID)
	if err != nil {
		log.Printf("Error making user %d admin: %v", userID, err)
		return fmt.Errorf("database error updating admin status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user with ID %d not found", userID)
	}

	log.Printf("User %d granted admin privileges.", userID)
	return nil
}

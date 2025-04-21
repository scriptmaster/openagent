package server

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/joho/godotenv" // Keep godotenv for loading .env

	// Keep auth for UserService
	"github.com/scriptmaster/openagent/common" // Updated import path
)

var (
	templates   *template.Template
	appVersion  string
	sessionSalt string
)

// StartServer initializes and starts the web server.
// Moved from main.go and Exported.
func StartServer() error {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Load .env file from project root
	err := godotenv.Load()
	if err != nil {
		// Log warning if .env is not found in root
		log.Printf("Warning: Error loading .env file from project root: %v", err)
	}

	// Determine SQL directory based on DB_DRIVER
	dbDriver := common.GetEnvOrDefault("DB_DRIVER", "postgres")
	// Default base path is now ./data/sql
	sqlBaseDir := common.GetEnvOrDefault("SQL_DIR", "./data/sql")
	sqlDir := filepath.Join(sqlBaseDir, dbDriver) // Construct path like ./data/sql/postgres

	// Load SQL queries from the driver-specific directory
	if err := common.LoadNamedSQLFiles(sqlDir); err != nil {
		log.Fatalf("CRITICAL: Failed to load SQL files from %s: %v", sqlDir, err)
	}
	log.Printf("Successfully loaded SQL queries from %s", sqlDir)

	// Initialize database (uses InitDB from the server package)
	db, err := InitDB()
	if err != nil {
		// If DB init fails, log fatal is too strong, return error instead
		// to allow main to handle shutdown gracefully or enter maintenance.
		log.Printf("Failed to initialize database: %v", err)
		// Check if we are in maintenance mode (set by InitDB)
		if !IsMaintenanceMode() {
			// If not already in maintenance, set it and attempt minimal start
			SetMaintenanceMode(true)
			log.Println("Entering maintenance mode due to DB init failure.")
			// Potentially start a minimal maintenance server here if desired
			// For now, just return the error to stop normal startup
			return fmt.Errorf("failed to initialize database: %v", err)
		}
		// If already in maintenance mode from InitDB, allow startup to continue
		// to potentially serve maintenance pages. db will be nil.
		log.Println("Continuing server start in maintenance mode...")
	}

	// Initialize services
	pdbService := NewProjectDBService(db)
	dataService := NewDirectDataService(db)
	userService := NewUserService(db, pdbService, dataService) // Use server.NewUserService
	// projectService := NewProjectService(db) // ProjectService initialization might depend on repository
	/* // Commented out projectService and settingsService initialization as they are done in routes.go
	sqlxDB := sqlx.NewDb(db, "postgres") // Assuming postgres, adjust if needed
	projectRepo := projects.NewProjectRepository(sqlxDB)
	projectService := projects.NewProjectService(db, projectRepo) // Use projects.NewProjectService
	settingsService := NewSettingsService(db)
	*/

	// Check if setup is needed (no admin user) - Logic moved to RegisterRoutes or handlers
	/*
	   adminExists, err := userService.CheckIfAdminExists(context.Background()) // Pass context
	   if err != nil {
	       log.Printf("CRITICAL: Failed to check for admin user: %v. Server cannot start securely.", err)
	       // os.Exit(1) // Consider exiting if this check fails
	       // For now, log and continue, but configuration page might be needed
	   }
	*/

	// Load session salt (used by auth middleware/routes)
	salt := common.GetEnvOrDefault("SESSION_SALT", "default-insecure-salt-change-me")
	if salt == "default-insecure-salt-change-me" {
		log.Println("WARNING: Using default insecure session salt. Set SESSION_SALT environment variable.")
	}

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Register all routes (uses RegisterRoutes from the server package)
	// Pass the initialized userService and salt.
	// Other services (db, templates, projectService, etc.) will be initialized *within* RegisterRoutes.
	RegisterRoutes(mux, userService, salt)

	log.Println("Server starting on port " + common.GetEnvOrDefault("PORT", "8800"))
	if err := http.ListenAndServe(common.GetEnvOrDefault("PORT", "8800"), mux); err != nil {
		return err
	}
	return nil
}

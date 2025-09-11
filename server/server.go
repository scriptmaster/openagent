package server

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/joho/godotenv" // Keep godotenv for loading .env

	// Keep auth for UserService
	"github.com/scriptmaster/openagent/common" // Updated import path
)

var (
	globalTemplates *TemplateEngine
	appVersion      string
	sessionSalt     string
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
		log.Printf("\t → 1. CRITICAL: Failed to load SQL files from %s: %v", sqlDir, err)
		return err
	} else {
		log.Printf("\t → 1. Successfully loaded SQL queries from %s", sqlDir)
	}

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

	// Initialize the 3 services
	pdbService := NewProjectDBService(db)
	dataService := NewDirectDataService(db)
	userService := NewUserService(db, pdbService, dataService) // Use server.NewUserService

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
	salt := common.GetEnvOrDefault("SESSION_SALT", "DEFAULT-SALT-72815ECE-99A4-45FD-98C0-38D9EE04813F")
	if salt == "DEFAULT-SALT-72815ECE-99A4-45FD-98C0-38D9EE04813F" {
		log.Println("WARNING: Using default insecure session salt. Set SESSION_SALT environment variable.")
	}

	// Create a new ServeMux
	router := http.NewServeMux()

	log.Printf("\t → 6. Registering Routes")
	// Register all routes (uses RegisterRoutes from the server package)
	// Pass the initialized userService and salt.
	// Other services (db, templates, projectService, etc.) will be initialized *within* RegisterRoutes.
	RegisterRoutes(router, userService, salt)

	startAddress := ":" + common.GetEnvOrDefault("PORT", "8800")
	log.Println("Server starting on " + startAddress)

	if err := http.ListenAndServe(startAddress, router); err != nil {
		return err
	}

	log.Println("!! Server STARTED !! " + startAddress)
	return nil
}

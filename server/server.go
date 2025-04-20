package server

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/joho/godotenv"                 // Keep godotenv for loading .env
	"github.com/scriptmaster/openagent/auth"   // Keep auth for UserService
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

	// Create user service (Requires *sql.DB, handle nil case if in maintenance)
	var userService *auth.UserService
	if db != nil {
		userService = auth.NewUserService(db)
		if userService == nil {
			return fmt.Errorf("failed to create user service")
		}
	} else {
		log.Println("Warning: Database connection is nil, running without user service features.")
		// userService remains nil, routes needing it should handle this
	}

	// Create main router
	mux := http.NewServeMux()

	salt := GetSessionSalt()

	// Register all routes (uses RegisterRoutes from the server package)
	// Pass potentially nil userService if in maintenance mode
	RegisterRoutes(mux, userService, salt)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:    ":" + common.GetEnvOrDefault("PORT", "8800"), // Use common.GetEnvOrDefault
		Handler: mux,
	}

	// Channel to listen for server errors
	serverErrors := make(chan error, 1)

	// Start the server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", httpServer.Addr)
		serverErrors <- httpServer.ListenAndServe()
	}()

	// Return the error from the channel (blocks until server stops or errors)
	return <-serverErrors
}

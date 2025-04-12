package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/scriptmaster/openagent/auth"
)

// StartServer initializes and starts the web server
func StartServer() error {
	log.Println("--- Server Starting ---")

	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: .env file not found, using environment variables.")
	}

	// Initialize database
	db, err := InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create user service
	userService := auth.NewUserService(db)

	// Create main router
	mux := http.NewServeMux()

	// Register all routes
	RegisterRoutes(mux, userService)

	// Create HTTP server
	server := &http.Server{
		Addr:    ":" + getEnvOrDefault("PORT", "8800"),
		Handler: mux,
	}

	// Channel to listen for server errors
	serverErrors := make(chan error, 1)

	// Start the server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", server.Addr)
		serverErrors <- server.ListenAndServe()
	}()

	// Return from this function, which will be picked up as a server error
	return <-serverErrors
}

func main() {
	// Set up signal catching
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- StartServer()
	}()

	// Wait for termination signal or server error
	select {
	case err := <-errCh:
		log.Fatalf("Server failed to start: %v", err)
	case sig := <-signals:
		// Get app version for goodbye message
		appVersion := os.Getenv("APP_VERSION")
		if appVersion == "" {
			appVersion = "1.0.0.0" // Default if not set
		}

		// Print goodbye message
		log.Printf("\n📡 Received signal %v. Bye bye! OpenAgent version %s shutting down...\n", sig, appVersion)

		// Exit gracefully
		os.Exit(0)
	}
}

package server

import (
	"crypto/sha256"
	"encoding/hex"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/scriptmaster/openagent/admin"
	"github.com/scriptmaster/openagent/auth"
	"github.com/scriptmaster/openagent/projects"
)

var (
	templates   *template.Template
	appVersion  string
	sessionSalt string
)

// RegisterRoutes registers all application routes for the server package
func RegisterRoutes(mux *http.ServeMux, userService *auth.UserService, salt string) {
	// Initialize the global templates variable
	templates = template.Must(template.New("").Funcs(GetTemplateFuncs()).ParseGlob("tpl/*.html"))

	// Load app version into the global appVersion variable
	appVersion = os.Getenv("APP_VERSION")
	if appVersion == "" {
		appVersion = "1.0.0.0" // Default if not set
	}
	// Assign the passed-in salt to the global sessionSalt variable
	sessionSalt = salt

	// Setup static file server
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Register routes from sub-packages
	auth.RegisterAuthRoutes(mux, templates, userService)
	admin.RegisterAdminRoutes(mux, templates, auth.IsMaintenanceAuthenticated,
		UpdateDatabaseConfig, UpdateMigrationStart, InitDB, sessionSalt)
	projects.RegisterProjectRoutes(mux, templates, userService)

	// Register agent routes (internal to server package)
	registerAgentRoutes(mux)

	// Register main package routes (now handled by auth package)
	// mux.HandleFunc("/login", HandleLogin)   // Removed - Handled by auth.RegisterAuthRoutes
	// mux.HandleFunc("/logout", HandleLogout) // Removed - Handled by auth.RegisterAuthRoutes
}

// registerAgentRoutes registers all agent-related routes
// This is now internal to the server package
func registerAgentRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/agent", HandleAgent)   // Expects exported HandleAgent
	mux.HandleFunc("/start", HandleStart)   // Expects exported HandleStart
	mux.HandleFunc("/next", HandleNextStep) // Expects exported HandleNextStep
	mux.HandleFunc("/status", HandleStatus) // Expects exported HandleStatus
}

// handleLogin moved to handlers.go and renamed HandleLogin

// handleLogout moved to handlers.go and renamed HandleLogout

// HandleLogin and HandleLogout should be imported from handlers.go (implicitly done by package)
// Placeholder definitions if not in handlers.go
// func HandleLogin(w http.ResponseWriter, r *http.Request) {}
// func HandleLogout(w http.ResponseWriter, r *http.Request) {}

// The agent handlers HandleAgent, HandleStart, HandleNextStep, HandleStatus
// must be defined and exported in agent.go

// GetSessionSalt returns the session salt generated during route registration.
// It must be called after RegisterRoutes has been executed.
func GetSessionSalt() string {
	// Return the globally stored salt
	// Ensure sessionSalt is initialized before this is called (e.g., in RegisterRoutes)
	if sessionSalt == "" {
		// This shouldn't happen in the normal flow where RegisterRoutes is called first.
		// Maybe generate a temporary one or log a warning?
		log.Println("Warning: GetSessionSalt called before sessionSalt was initialized in RegisterRoutes.")
		// Fallback or panic might be appropriate depending on strictness needed.
		// For now, let's recalculate based on current env var as a fallback.
		version := os.Getenv("APP_VERSION")
		if version == "" {
			version = "1.0.0.0"
		}
		return generateSessionSalt(version) // Use the helper if available
	}
	return sessionSalt
}

// generateSessionSalt generates a salt based on the app version.
// This should ideally be defined once, maybe in functions.go or kept private here.
func generateSessionSalt(version string) string {
	// Simple salt generation (replace with more robust method if needed)
	h := sha256.New()
	h.Write([]byte(version))
	h.Write([]byte("-openagent-secret-salt-value")) // Add a static secret
	return hex.EncodeToString(h.Sum(nil))[:16]      // Use first 16 chars
}

// Helper imports needed for generateSessionSalt
// "crypto/sha256" is already imported
// "encoding/hex" is already imported
// "log" is already imported
// "os" is already imported

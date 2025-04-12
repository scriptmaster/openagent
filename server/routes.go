package server

import (
	"html/template"
	"log"
	"net/http"

	"github.com/scriptmaster/openagent/admin"
	"github.com/scriptmaster/openagent/auth"
	"github.com/scriptmaster/openagent/common"
	"github.com/scriptmaster/openagent/projects"
)

// RegisterRoutes registers all application routes for the server package
func RegisterRoutes(mux *http.ServeMux, userService *auth.UserService, salt string) {
	// Load app version into the global appVersion variable
	appVersion = common.GetEnvOrDefault("APP_VERSION", "1.0.0.0")
	// Assign the passed-in salt to the global sessionSalt variable
	sessionSalt = salt

	// Initialize the global templates variable
	templates = template.Must(template.New("").Funcs(GetTemplateFuncs()).ParseGlob("tpl/*.html"))
	// Initialize agent templates
	InitAgentTemplates(templates)

	// Initialize database connection
	db, err := InitDB()
	if err != nil {
		log.Printf("Failed to initialize database: %v", err)
		SetMaintenanceMode(true)
	}

	// Setup static file server
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Register routes from sub-packages
	auth.RegisterAuthRoutes(mux, templates, userService)
	admin.RegisterAdminRoutes(mux, templates, auth.IsMaintenanceAuthenticated,
		UpdateDatabaseConfig, UpdateMigrationStart, InitDB, sessionSalt)

	// Only register project routes if database is available
	if db != nil {
		projects.RegisterProjectRoutes(mux, templates, userService, db)
	}

	// Register agent routes (internal to server package)
	registerAgentRoutes(mux)
}

// registerAgentRoutes registers all agent-related routes
// This is now internal to the server package
func registerAgentRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/agent", HandleAgent)   // Expects exported HandleAgent
	mux.HandleFunc("/start", HandleStart)   // Expects exported HandleStart
	mux.HandleFunc("/next", HandleNextStep) // Expects exported HandleNextStep
	mux.HandleFunc("/status", HandleStatus) // Expects exported HandleStatus
}

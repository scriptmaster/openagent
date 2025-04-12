package main

import (
	"html/template"
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

// RegisterRoutes registers all application routes
func RegisterRoutes(mux *http.ServeMux, userService *auth.UserService) {
	// Initialize template engine with helper functions
	templates = template.Must(template.New("").Funcs(GetTemplateFuncs()).ParseGlob("tpl/*.html"))

	// Load app version from environment and generate session salt
	appVersion = os.Getenv("APP_VERSION")
	if appVersion == "" {
		appVersion = "1.0.0.0" // Default if not set
	}
	sessionSalt = generateSessionSalt(appVersion)

	// Setup static file server
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Register routes from packages
	auth.RegisterAuthRoutes(mux, templates, userService)
	admin.RegisterAdminRoutes(mux, templates, auth.IsMaintenanceAuthenticated,
		UpdateDatabaseConfig, UpdateMigrationStart, InitDB, sessionSalt)
	projects.RegisterProjectRoutes(mux, templates, userService)

	// Register agent routes
	registerAgentRoutes(mux)

	// Register main package routes
	mux.HandleFunc("/login", handleLogin)   // Use handleLogin from handlers.go
	mux.HandleFunc("/logout", handleLogout) // Use handleLogout from handlers.go
}

// registerAgentRoutes registers all agent-related routes
func registerAgentRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/agent", handleAgent)
	mux.HandleFunc("/start", handleStart)
	mux.HandleFunc("/next", handleNextStep)
	mux.HandleFunc("/status", handleStatus)
}

// handleLogin moved to handlers.go

// handleLogout moved to handlers.go

package main

import (
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/scriptmaster/openagent/admin"
	"github.com/scriptmaster/openagent/auth"
	"github.com/scriptmaster/openagent/projects"
)

var (
	templates   *template.Template
	appVersion  string
	sessionSalt string
)

// PageData represents the data passed to templates
type PageData struct {
	AppName    string
	PageTitle  string
	User       auth.User
	Error      string
	Projects   []interface{}
	Project    interface{}
	AdminEmail string
	AppVersion string
}

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
}

// registerAgentRoutes registers all agent-related routes
func registerAgentRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/agent", handleAgent)
	mux.HandleFunc("/start", handleStart)
	mux.HandleFunc("/next", handleNextStep)
	mux.HandleFunc("/status", handleStatus)
}

// handleLogin displays the login page
func handleLogin(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		AppName:    "OpenAgent",
		PageTitle:  "Login - OpenAgent",
		AdminEmail: os.Getenv("SYSADMIN_EMAIL"),
		AppVersion: appVersion,
	}

	if err := templates.ExecuteTemplate(w, "login.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleLogout clears the session cookie and redirects to login
func handleLogout(w http.ResponseWriter, r *http.Request) {
	// Clear the session cookie by setting an expired cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,                              // Delete cookie immediately
		Expires:  time.Now().Add(-24 * time.Hour), // Expired
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect to login page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

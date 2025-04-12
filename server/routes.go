package server

import (
	"database/sql"
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

	// Initialize the global templates variable
	templates = template.Must(template.New("").Funcs(GetTemplateFuncs()).ParseGlob("tpl/*.html"))
	templates = template.Must(templates.ParseGlob("tpl/_partials/*.html"))
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
		RegisterRootRoutes(mux, templates, userService, db)
		projects.RegisterProjectRoutes(mux, templates, userService, db)
	}

	// Register agent routes (internal to server package)
	registerAgentRoutes(mux)

	// Register catch-all handler for 404 errors - must be last
	mux.HandleFunc("/{path:.*}", func(w http.ResponseWriter, r *http.Request) {
		common.Handle404(w, r, templates)
	})
}

func RegisterRootRoutes(mux *http.ServeMux, templates *template.Template, userService *auth.UserService, db *sql.DB) {
	projectService, err := projects.NewProjectService(db)
	if err != nil {
		log.Printf("Failed to create project service: %v", err)
		SetMaintenanceMode(true)
		return
	}

	// Root and project page routes handler
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			// Handle root path
			projects.HandleIndexRoute(w, r, templates, projectService, userService)
			return
		}

		// Check if current domain has a project and the path matches a project page
		projects.HandleProjectPageRoute(w, r, templates, projectService, userService)
	})
}

// registerAgentRoutes registers all agent-related routes
// This is now internal to the server package
func registerAgentRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/agent", HandleAgent)   // Expects exported HandleAgent
	mux.HandleFunc("/start", HandleStart)   // Expects exported HandleStart
	mux.HandleFunc("/next", HandleNextStep) // Expects exported HandleNextStep
	mux.HandleFunc("/status", HandleStatus) // Expects exported HandleStatus
}

package server

import (
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/scriptmaster/openagent/admin"
	"github.com/scriptmaster/openagent/auth"
	"github.com/scriptmaster/openagent/common"
	"github.com/scriptmaster/openagent/projects"
)

// RegisterRoutes sets up all the application routes
func RegisterRoutes(mux *http.ServeMux, userService auth.UserServicer, salt string) {
	// Load HTML templates
	templates := LoadTemplates()

	// Initialize database and other services (handle potential nil db)
	db := GetDB() // Get the initialized DB instance
	var projectService projects.ProjectService
	var pdbService common.ProjectDBService
	// var settingsService *SettingsService // Declared and not used
	// var dataService DataAccessService // Declared and not used

	if db != nil {
		log.Println("Database connection available, initializing DB-dependent services.")
		pdbService = NewProjectDBService(db)
		// settingsService = NewSettingsService(db) // Commented out: Not used yet
		// dataService = NewDirectDataService(db) // Commented out: Not used yet

		// Initialize project service (needs sqlx wrapper)
		sqlxDB := sqlx.NewDb(db, common.GetEnvOrDefault("DB_DRIVER", "postgres")) // Use configured driver
		projectRepo := projects.NewProjectRepository(sqlxDB)
		projectService = projects.NewProjectService(db, projectRepo)

	} else {
		log.Println("Warning: Database connection is nil. DB-dependent services will not be available.")
		// Services remain nil
	}

	// --- Static Files ---
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// --- Exempt Paths Middleware (Applied Later) ---
	exemptPaths := map[string]bool{
		"/static/":          true,
		"/login":            true,
		"/logout":           true,
		"/auth/request-otp": true,
		"/auth/verify-otp":  true,
		"/password-login":   true,
		"/config":           true,
		"/config/save":      true,
		"/maintenance":      true,
		"/admin/login":      true,
	}

	// Create a base mux for routes that might be wrapped by middleware
	baseMux := http.NewServeMux()

	// --- Register Non-Project/Public Routes Directly on baseMux ---

	// Maintenance Page
	baseMux.HandleFunc("/maintenance", func(w http.ResponseWriter, r *http.Request) {
		admin.HandleMaintenance(w, r, templates, auth.IsMaintenanceAuthenticated)
	})
	// Admin Login for Maintenance
	/* // Commented out: admin.HandleAdminLogin undefined
	baseMux.HandleFunc("/admin/login", func(w http.ResponseWriter, r *http.Request) {
		admin.HandleAdminLogin(w, r, templates)
	})
	*/

	// Configuration Page & Save Endpoint
	baseMux.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		HandleConfigPage(w, r)
	})
	baseMux.HandleFunc("/config/save", func(w http.ResponseWriter, r *http.Request) {
		if userService == nil || projectService == nil {
			common.JSONError(w, "System not fully configured", http.StatusServiceUnavailable)
			return
		}
		HandleConfigSubmit(w, r, userService, projectService)
	})

	// --- Register Auth Routes --- (Use baseMux)
	if userService != nil {
		auth.RegisterAuthRoutes(baseMux, templates, userService)
	} else {
		// Handle auth routes gracefully if userService is nil
		log.Println("Auth routes disabled: userService is nil (DB connection likely failed)")
		// Optionally redirect /login, etc., to /config
		handleNilService(baseMux, "/login", "/auth/")
	}

	// --- Register Project Routes --- (Use baseMux)
	if db != nil && userService != nil && pdbService != nil { // Ensure all needed services are available
		projects.RegisterProjectRoutes(baseMux, templates, userService, db, pdbService)
	} else {
		// Handle project routes gracefully if services are nil
		log.Println("Project routes disabled: required services (db, userService, pdbService) not available")
		handleNilService(baseMux, "/projects", "/api/projects/")
	}

	// --- Register Other Protected Routes --- (Use baseMux)

	// Dashboard (requires auth)
	baseMux.Handle("/dashboard", auth.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if projectService == nil { // Check if projectService is initialized
			http.Error(w, "Project service not available", http.StatusInternalServerError)
			return
		}
		HandleDashboard(w, r, projectService)
	})))

	// Voice Page (assuming it needs auth)
	baseMux.Handle("/voice", auth.AuthMiddleware(http.HandlerFunc(HandleVoicePage)))

	// --- Apply Middleware ---
	finalHandler := HostProjectMiddleware(baseMux, projectService, userService, exemptPaths)
	mux.Handle("/", finalHandler) // Route ALL requests through the middleware first
}

// handleNilService registers handlers for paths when required services are unavailable.
func handleNilService(mux *http.ServeMux, paths ...string) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			common.JSONError(w, "Service unavailable: Database not configured or connection failed.", http.StatusServiceUnavailable)
		} else {
			http.Redirect(w, r, "/config?error=db_unavailable", http.StatusSeeOther)
		}
	}
	for _, path := range paths {
		mux.HandleFunc(path, handler)
	}
}

// LoadTemplates loads HTML templates from the tpl directory
func LoadTemplates() *template.Template {
	// Start with base name and functions
	baseTemplate := template.New("base").Funcs(GetTemplateFuncs())
	// Parse all files from tpl (including layout.html)
	templates, err := baseTemplate.ParseGlob("tpl/*.html")
	if err != nil {
		log.Fatalf("FATAL: Failed to parse tpl/*.html: %v", err)
	}
	// Parse partials into the same template set
	templates = template.Must(templates.ParseGlob("tpl/_partials/*.html"))
	return templates
}

// Commented out duplicate GetTemplateFuncs
/*
// GetTemplateFuncs returns a map of template functions
func GetTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		// Add any custom template functions here
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
	}
}
*/

// Commented out potentially obsolete RegisterRootRoutes
/*
func RegisterRootRoutes(mux *http.ServeMux, templates *template.Template, userService auth.UserServicer, db *sql.DB, projectService projects.ProjectService) {
	// ... implementation ...
}
*/

// Commented out unused registerAgentRoutes
/*
func registerAgentRoutes(mux *http.ServeMux) {
	// ... implementation ...
}
*/

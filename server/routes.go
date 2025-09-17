package server

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/scriptmaster/openagent/admin"
	"github.com/scriptmaster/openagent/auth"
	"github.com/scriptmaster/openagent/common"
	"github.com/scriptmaster/openagent/projects"
)

// RegisterRoutes sets up all the application routes
func RegisterRoutes(router *http.ServeMux, userService auth.UserServicer, salt string) {
	// Initialize database and other services (handle potential nil db)
	db := GetDB() // Get the initialized DB instance
	var projectService projects.ProjectService
	var pdbService common.ProjectDBService
	// var settingsService *SettingsService // Declared and not used
	// var dataService DataAccessService // Declared and not used

	if db != nil {
		log.Println("\t → \t → 6.1 Database connection available, initializing DB-dependent services and routes.")
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

	staticFilesRoot := "./static"
	log.Printf("\t → \t → 6.2 Static files served from: %v", staticFilesRoot)

	// --- Favicon Route (direct to main router, no middleware, BEFORE static files) ---
	router.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/favicon.ico")
	})

	// --- Static Files with Cache Headers ---
	fs := http.FileServer(http.Dir(staticFilesRoot))
	router.Handle("/static/", http.StripPrefix("/static/", addCacheHeaders(fs)))

	// --- TSX Generated CSS and JS Files with Cache Headers ---
	router.HandleFunc("/tsx/css/", func(w http.ResponseWriter, r *http.Request) {
		// Add cache headers for generated CSS files (shorter cache than static files)
		w.Header().Set("Cache-Control", "public, max-age=3600") // 1 hour
		w.Header().Set("ETag", `"`+fmt.Sprintf("%d", time.Now().Unix())+`"`)
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Extract filename from path
		filename := strings.TrimPrefix(r.URL.Path, "/tsx/css/")
		if filename == "" {
			http.NotFound(w, r)
			return
		}

		// Serve from generated/css directory
		filePath := fmt.Sprintf("./tpl/generated/css/%s", filename)
		http.ServeFile(w, r, filePath)
	})

	router.HandleFunc("/tsx/js/", func(w http.ResponseWriter, r *http.Request) {
		// Add cache headers for generated JS files (shorter cache than static files)
		w.Header().Set("Cache-Control", "public, max-age=3600") // 1 hour
		w.Header().Set("ETag", `"`+fmt.Sprintf("%d", time.Now().Unix())+`"`)
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Extract filename from path
		filename := strings.TrimPrefix(r.URL.Path, "/tsx/js/")
		if filename == "" {
			http.NotFound(w, r)
			return
		}

		// Serve from generated/js directory (all files consolidated here)
		filePath := fmt.Sprintf("./tpl/generated/js/%s", filename)
		http.ServeFile(w, r, filePath)
	})

	// --- Register Non-Project/Public Routes Directly on main router ---

	maintenanceRoutePath := "/maintenance"
	log.Printf("\t → \t → 6.3 Creating maintenance route: %v", maintenanceRoutePath)
	// Maintenance Page
	router.HandleFunc(maintenanceRoutePath, func(w http.ResponseWriter, r *http.Request) {
		admin.HandleMaintenance(w, r, globalTemplates, auth.IsMaintenanceAuthenticated)
	})

	configRoutePath := "/config"
	log.Printf("\t → \t → 6.3 Creating config route paths: %v and ./save", configRoutePath)
	// Configuration Page & Save Endpoint
	router.HandleFunc(configRoutePath, func(w http.ResponseWriter, r *http.Request) {
		HandleConfigPage(w, r)
	})
	router.HandleFunc("/config/save", func(w http.ResponseWriter, r *http.Request) {
		if userService == nil || projectService == nil {
			common.JSONError(w, "System not fully configured", http.StatusServiceUnavailable)
			return
		}
		HandleConfigSubmit(w, r, userService, projectService)
	})

	// --- Register Auth Routes --- (Use main router)
	if userService != nil {
		auth.RegisterAuthRoutes(router, globalTemplates, userService)
	} else {
		// Handle auth routes gracefully if userService is nil
		log.Println("Auth routes disabled: userService is nil (DB connection likely failed)")
		// Optionally redirect /login, etc., to /config
		handleNilService(router, "/login", "/auth/")
	}

	// --- Register Project Routes --- (Use main router)
	if db != nil && userService != nil && pdbService != nil { // Ensure all needed services are available
		projects.RegisterProjectRoutes(router, globalTemplates, userService, db, pdbService)
	} else {
		// Handle project routes gracefully if services are nil
		log.Println("Project routes disabled: required services (db, userService, pdbService) not available")
		handleNilService(router, "/projects", "/api/projects/")
	}

	// --- Register Admin Routes --- (Use main router)
	log.Printf("\t → \t → 6.4 Registering admin routes")
	if db != nil {
		// Create a function to update database config (placeholder for now)
		updateDatabaseConfig := func(host, port, user, password, dbname string) error {
			// TODO: Implement database configuration update
			log.Printf("Database config update requested: %s:%s/%s user=%s", host, port, dbname, user)
			return nil
		}

		// Register admin routes with all required dependencies
		admin.RegisterAdminRoutes(router, globalTemplates, auth.IsMaintenanceAuthenticated, updateDatabaseConfig, InitDB, salt, GetDB, GetAdminStats)
	} else {
		// Handle admin routes gracefully if database is not available
		log.Println("Admin routes disabled: database connection not available")
		handleNilService(router, "/admin", "/maintenance/")
	}

	// --- Register Other Protected Routes --- (Use main router)

	log.Printf("\t → \t → 6.X Setting /dashboard handler with Auth")
	// Dashboard (requires auth) - redirects admin users to /admin
	router.Handle("/dashboard", auth.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := auth.GetUserFromContext(r.Context())
		if user != nil && user.IsAdmin {
			// Admin users get redirected to admin dashboard
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
			return
		}

		if projectService == nil { // Check if projectService is initialized
			http.Error(w, "Project service not available", http.StatusInternalServerError)
			return
		}
		HandleDashboard(w, r, projectService)
	})))

	log.Printf("\t → \t → 6.X Setting /voice handler with Auth")
	// Voice Page (assuming it needs auth)
	router.Handle("/voice", auth.AuthMiddleware(http.HandlerFunc(HandleVoicePage)))

	log.Printf("\t → \t → 6.X Setting /agent handler with Auth")
	// Agent Page (assuming it needs auth)
	router.Handle("/agent", auth.AuthMiddleware(http.HandlerFunc(HandleAgentPage)))

	// Test route for template system
	router.HandleFunc("/test", HandleTestPage)

	// Root route handler - serve default index page
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Only handle root path "/" - serve the default index page
		if r.URL.Path == "/" {
			// Always serve the default index page with "Welcome to OpenAgent" message
			HandleIndexPage(w, r)
		} else {
			// Handle 404 for non-root requests
			admin.Handle404(w, r, globalTemplates)
		}
	})

	// No middleware needed - all routes are directly on main router
}

// handleNilService registers handlers for paths when required services are unavailable.
func handleNilService(router *http.ServeMux, paths ...string) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			common.JSONError(w, "Service unavailable: Database not configured or connection failed.", http.StatusServiceUnavailable)
		} else {
			http.Redirect(w, r, "/config?error=db_unavailable", http.StatusSeeOther)
		}
	}
	for _, path := range paths {
		router.HandleFunc(path, handler)
	}
}

// LoadTemplates loads templates using the unified template engine
func LoadTemplates() *TemplateEngine {
	return NewTemplateEngine()
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

// addCacheHeaders wraps an http.Handler to add cache headers for static files
func addCacheHeaders(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add cache headers for static files
		w.Header().Set("Cache-Control", "public, max-age=86400") // 1 day
		w.Header().Set("ETag", `"`+fmt.Sprintf("%d", time.Now().Unix())+`"`)

		// Add security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Call the original handler
		h.ServeHTTP(w, r)
	})
}

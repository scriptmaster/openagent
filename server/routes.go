package server

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/scriptmaster/openagent/admin"
	"github.com/scriptmaster/openagent/auth"
	"github.com/scriptmaster/openagent/common"
	"github.com/scriptmaster/openagent/projects"
)

// RegisterRoutes registers all application routes for the server package
func RegisterRoutes(mux *http.ServeMux, userService *auth.UserService, salt string) {
	// Load app version
	appVersion = common.GetEnvOrDefault("APP_VERSION", "1.0.0.0")

	// Initialize templates
	templates = template.Must(template.New("").Funcs(GetTemplateFuncs()).ParseGlob("tpl/*.html"))
	templates = template.Must(templates.ParseGlob("tpl/_partials/*.html"))
	InitAgentTemplates(templates) // Assuming this is still needed

	// Initialize database connection (includes admin token check now)
	db, err := InitDB()
	if err != nil {
		log.Printf("CRITICAL: Failed to initialize database: %v. Entering maintenance.", err)
		SetMaintenanceMode(true)
		// Don't exit, allow maintenance routes
	}

	// Initialize Services Needed by Middleware/Handlers
	// Make sure db is not nil if needed by services, handle maintenance mode appropriately
	var projectService projects.ProjectService
	if db != nil {
		sqlxDB := sqlx.NewDb(db, "postgres") // Assuming postgres driver
		projectRepo := projects.NewProjectRepository(sqlxDB)
		projectService = projects.NewProjectService(db, projectRepo) // db might be nil here if InitDB failed!
	} else {
		log.Println("WARN: Database not available, project functionality will be limited.")
		// Handle the case where projectService is nil if necessary in middleware/handlers
		// Or provide a dummy service that always returns "not found"?
	}

	// --- Middleware Definition ---
	exemptPaths := []string{
		"/static/",
		"/login",
		"/logout",
		"/request-otp",
		"/verify-otp",
		"/password-login",
		"/config",        // Allow access to the config page itself
		"/config/submit", // Allow access to the submit endpoint
		"/maintenance",   // Allow maintenance routes
		// Add any other public API endpoints or essential paths here
	}
	// IMPORTANT: Apply middleware *after* static files and exempt routes
	// We create a new mux here to wrap the main one with middleware easily
	baseMux := http.NewServeMux()

	// --- Register Exempt Routes Directly on baseMux ---
	fs := http.FileServer(http.Dir("static"))
	baseMux.Handle("/static/", http.StripPrefix("/static/", fs))

	auth.RegisterAuthRoutes(baseMux, templates, userService) // Handles login, logout, otp, etc.
	admin.RegisterAdminRoutes(baseMux, templates, auth.IsMaintenanceAuthenticated,
		UpdateDatabaseConfig, UpdateMigrationStart, InitDB, sessionSalt) // Handles /maintenance/*

	// Config page submit handler (exempted, but needs services)
	baseMux.HandleFunc("/config/submit", func(w http.ResponseWriter, r *http.Request) {
		// Check if services are available due to potential DB init failure
		if userService == nil || projectService == nil {
			common.JSONError(w, "Service not available due to configuration issue", http.StatusServiceUnavailable)
			return
		}
		HandleConfigSubmit(w, r, userService, projectService)
	})

	// --- Register Protected/Project-Specific Routes (will be wrapped) ---

	// Dashboard (requires auth, doesn't necessarily need project context from middleware)
	baseMux.Handle("/dashboard", auth.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HandleDashboard(w, r, projectService) // Pass projectService
	})))

	// Voice Page (assuming it needs auth and possibly project context?)
	baseMux.Handle("/voice", auth.AuthMiddleware(http.HandlerFunc(HandleVoicePage)))

	// Project List Page (needs auth, project context determined by host via middleware)
	// The HandleProjectsRoute might need refactoring if projectService isn't passed directly anymore
	// or if it should fetch projects based on user permissions instead of host context?
	// For now, let's assume it needs auth. The project context is handled by middleware.
	if projectService != nil { // Only register if service is available
		baseMux.Handle("/projects", auth.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Refactor: HandleProjectsRoute expects projectService passed in.
			// We might need to modify HandleProjectsRoute or HandleProjects
			// to get the service differently or rely purely on context/user.
			// For now, passing the globally available one (if not nil).
			// This needs careful review based on HandleProjectsRoute/HandleProjects implementation.
			projects.HandleProjectsRoute(w, r, templates, projectService, userService)
		})))
	}

	// Root route "/" (and domain-specific project pages)
	// RegisterRootRoutes sets up its own HandleFunc for "/"
	// This needs careful integration with the middleware.
	// The middleware *should* add project context *before* this handler is called for a specific domain.
	// The handler inside RegisterRootRoutes should perhaps use GetProjectFromContext now?
	if projectService != nil {
		RegisterRootRoutes(baseMux, templates, userService, db, projectService) // Pass projectService
	} else {
		// Handle "/" when DB/ProjectService is down (maybe redirect to maintenance/config?)
		baseMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if IsMaintenanceMode() {
				admin.HandleMaintenance(w, r, templates, auth.IsMaintenanceAuthenticated)
				return
			}
			// Redirect to config page if project service isn't available?
			HandleConfigPage(w, r) // Show config page
		})
	}

	// --- Apply Middleware ---
	// The HostProjectMiddleware wraps all routes registered on baseMux *except* the exempt ones handled above.
	// Correct approach: Wrap the baseMux itself.

	finalHandler := HostProjectMiddleware(baseMux, projectService, userService, exemptPaths) // Pass userService here
	mux.Handle("/", finalHandler)                                                            // Route ALL requests through the middleware first

	// Remove the old catch-all from the main mux if it existed.
}

// Modify RegisterRootRoutes to accept ProjectService and potentially use context
func RegisterRootRoutes(mux *http.ServeMux, templates *template.Template, userService *auth.UserService, db *sql.DB, projectService projects.ProjectService) {
	// Don't need to init projectService here anymore, it's passed in.

	// Root and project page routes handler
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Middleware should have already determined if we are here because:
		// 1. It's the actual root path "/" for an unknown host (served by HandleIndexRoute -> HandleIndex)
		// 2. It's a path on a known project host (served by HandleProjectPageRoute)

		// Let the existing project handlers decide based on path and context
		projectFromCtx := GetProjectFromContext(r.Context()) // Get project potentially added by middleware

		if r.URL.Path == "/" && projectFromCtx == nil {
			// Handle root path for unconfigured domain (middleware sends to config, but maybe HandleIndex is fallback?)
			// This case might be handled by middleware redirecting to config already.
			// If we reach here, perhaps it's an explicit visit to "/" on an unknown host.
			projects.HandleIndexRoute(w, r, templates, projectService, userService)
		} else if projectFromCtx != nil {
			// Host matched a project in middleware, let the page route handler render it.
			// HandleProjectPageRoute might need adjustment to use projectFromCtx if available.
			projects.HandleProjectPageRoute(w, r, templates, projectService, userService)
		} else {
			// Fallback: No project context, not root path? Should be 404.
			// This case *should* be handled by the HostProjectMiddleware sending to config.
			// If it reaches here, something is wrong in the logic flow.
			log.Printf("WARN: Reached root handler unexpectedly for path %s on host %s", r.URL.Path, r.Host)
			common.Handle404(w, r, templates)
		}
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

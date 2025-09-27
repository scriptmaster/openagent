package admin

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/scriptmaster/openagent/auth"
	"github.com/scriptmaster/openagent/models"
	"github.com/scriptmaster/openagent/types"
)

// RegisterAdminRoutes registers all admin-related routes
func RegisterAdminRoutes(router *http.ServeMux, templates types.TemplateEngineInterface, isMaintenanceAuthenticated func(r *http.Request) bool, updateDatabaseConfig func(host, port, user, password, dbname string) error, initDB func() (*sql.DB, error), sessionSalt string, getDB func() *sql.DB, getAdminStats func(*sql.DB) (*models.AdminStats, error)) {
	// Admin dashboard - requires admin authentication
	router.Handle("/admin", auth.AuthMiddleware(auth.IsAdminMiddleware(http.HandlerFunc(CreateAdminHandler(templates, getDB, getAdminStats)))))

	router.HandleFunc("/maintenance/auth", CreateMaintenanceAuthHandler(sessionSalt))
	router.HandleFunc("/maintenance/config", CreateMaintenanceConfigHandler(templates, isMaintenanceAuthenticated))
	router.HandleFunc("/maintenance/configure", CreateMaintenanceConfigureHandler(templates, isMaintenanceAuthenticated, updateDatabaseConfig))
	router.HandleFunc("/maintenance/initialize-schema", CreateInitializeSchemaHandler(isMaintenanceAuthenticated, initDB))

	// Admin CLI - requires admin authentication
	router.Handle("/admin/cli", auth.AuthMiddleware(auth.IsAdminMiddleware(http.HandlerFunc(CreateAdminCLIHandler(templates, getDB)))))

	// Admin CLI API endpoints
	router.Handle("/admin/cli/api/queries", auth.AuthMiddleware(auth.IsAdminMiddleware(http.HandlerFunc(CreateCLIQueriesAPIHandler()))))
	router.Handle("/admin/cli/api/execute", auth.AuthMiddleware(auth.IsAdminMiddleware(http.HandlerFunc(CreateCLIExecuteAPIHandler(getDB)))))

	// Additional admin routes
	router.Handle("/admin/connections", auth.AuthMiddleware(auth.IsAdminMiddleware(http.HandlerFunc(CreateAdminConnectionsHandler(templates, getDB)))))
	router.Handle("/admin/tables", auth.AuthMiddleware(auth.IsAdminMiddleware(http.HandlerFunc(CreateAdminTablesHandler(templates, getDB)))))
	router.Handle("/admin/settings", auth.AuthMiddleware(auth.IsAdminMiddleware(http.HandlerFunc(CreateAdminSettingsHandler(templates, getDB)))))

	// General routes (accessible to all authenticated users)
	router.Handle("/connections", auth.AuthMiddleware(http.HandlerFunc(CreateConnectionsHandler(templates, getDB))))
	router.Handle("/tables", auth.AuthMiddleware(http.HandlerFunc(CreateTablesHandler(templates, getDB))))
	router.Handle("/profile", auth.AuthMiddleware(http.HandlerFunc(CreateProfileHandler(templates))))
}

// TODO: Check Unused???
// MaintenanceHandler handles requests when in maintenance mode
func MaintenanceHandler(next http.Handler, isMaintenanceMode func() bool, isMaintenanceAuthenticated func(r *http.Request) bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip maintenance mode check for static files
		if strings.HasPrefix(r.URL.Path, "/static/") {
			next.ServeHTTP(w, r)
			return
		}

		// Always allow access to maintenance endpoints
		if r.URL.Path == "/maintenance" || r.URL.Path == "/maintenance/auth" {
			next.ServeHTTP(w, r)
			return
		}

		// Special handling for login page - in maintenance mode redirect to /maintenance
		if r.URL.Path == "/login" && isMaintenanceMode() {
			http.Redirect(w, r, "/maintenance", http.StatusSeeOther)
			return
		}

		// Check for maintenance configuration access
		if strings.HasPrefix(r.URL.Path, "/maintenance/") {
			if !isMaintenanceAuthenticated(r) {
				// Not authenticated, redirect to maintenance login
				http.Redirect(w, r, "/maintenance", http.StatusSeeOther)
				return
			}
			// Authenticated, allow access
			next.ServeHTTP(w, r)
			return
		}

		// If in maintenance mode, check authentication
		if isMaintenanceMode() {
			if !isMaintenanceAuthenticated(r) {
				// Not authenticated, redirect to maintenance login
				http.Redirect(w, r, "/maintenance", http.StatusSeeOther)
				return
			}

			// Already authenticated for maintenance, redirect to configuration
			if r.URL.Path == "/maintenance" {
				http.Redirect(w, r, "/maintenance/config", http.StatusSeeOther)
				return
			}
		}

		// Continue normal processing
		next.ServeHTTP(w, r)
	})
}

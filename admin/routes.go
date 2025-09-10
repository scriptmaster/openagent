package admin

import (
	"database/sql"
	"html/template"
	"net/http"
	"strings"
)

// RegisterAdminRoutes registers all admin-related routes
func RegisterAdminRoutes(mux *http.ServeMux, templates *template.Template, isMaintenanceAuthenticated func(r *http.Request) bool, updateDatabaseConfig func(host, port, user, password, dbname string) error, initDB func() (*sql.DB, error), sessionSalt string) {
	// Admin dashboard
	mux.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		HandleAdmin(w, r, templates)
	})

	// Maintenance routes
	mux.HandleFunc("/maintenance", func(w http.ResponseWriter, r *http.Request) {
		HandleMaintenance(w, r, templates, isMaintenanceAuthenticated)
	})

	mux.HandleFunc("/maintenance/auth", func(w http.ResponseWriter, r *http.Request) {
		HandleMaintenanceAuth(w, r, sessionSalt)
	})

	mux.HandleFunc("/maintenance/config", func(w http.ResponseWriter, r *http.Request) {
		HandleMaintenanceConfig(w, r, templates, isMaintenanceAuthenticated)
	})

	mux.HandleFunc("/maintenance/configure", func(w http.ResponseWriter, r *http.Request) {
		HandleMaintenanceConfigure(w, r, templates, isMaintenanceAuthenticated, updateDatabaseConfig)
	})

	mux.HandleFunc("/maintenance/initialize-schema", func(w http.ResponseWriter, r *http.Request) {
		HandleInitializeSchema(w, r, isMaintenanceAuthenticated, initDB)
	})
}

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

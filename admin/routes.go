package admin

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	"github.com/scriptmaster/openagent/auth"
	"github.com/scriptmaster/openagent/common"
	"github.com/scriptmaster/openagent/models"
	"github.com/scriptmaster/openagent/types"
)

// RegisterAdminRoutes registers all admin-related routes
func RegisterAdminRoutes(router *http.ServeMux, templates types.TemplateEngineInterface, isMaintenanceAuthenticated func(r *http.Request) bool, updateDatabaseConfig func(host, port, user, password, dbname string) error, initDB func() (*sql.DB, error), sessionSalt string, getDB func() *sql.DB, getAdminStats func(*sql.DB) (*models.AdminStats, error)) {
	// Admin dashboard - requires admin authentication
	router.Handle("/admin", auth.AuthMiddleware(auth.IsAdminMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user from context (set by auth middleware)
		user := auth.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "User not found in context", http.StatusInternalServerError)
			return
		}

		// Get database connection and fetch stats
		db := getDB()
		var stats *models.AdminStats
		if db != nil {
			var err error
			stats, err = getAdminStats(db)
			if err != nil {
				log.Printf("Error fetching admin stats: %v", err)
				stats = &models.AdminStats{}
			}
		} else {
			stats = &models.AdminStats{}
		}

		data := models.PageData{
			AppName:        "OpenAgent",
			PageTitle:      "Admin Dashboard - OpenAgent",
			User:           user,
			AdminEmail:     common.GetEnv("SYSADMIN_EMAIL"),
			AppVersion:     common.GetEnv("APP_VERSION"),
			Stats:          stats,
			RecentActivity: []interface{}{},                             // Empty for now, can be populated later
			SystemHealth:   map[string]interface{}{"status": "healthy"}, // Basic system health info
		}

		if err := templates.ExecuteTemplate(w, "admin.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}))))

	// Maintenance routes
	// mux.HandleFunc("/maintenance", func(w http.ResponseWriter, r *http.Request) {
	// 	HandleMaintenance(w, r, templates, isMaintenanceAuthenticated)
	// })

	router.HandleFunc("/maintenance/auth", func(w http.ResponseWriter, r *http.Request) {
		HandleMaintenanceAuth(w, r, sessionSalt)
	})

	router.HandleFunc("/maintenance/config", func(w http.ResponseWriter, r *http.Request) {
		HandleMaintenanceConfig(w, r, templates, isMaintenanceAuthenticated)
	})

	router.HandleFunc("/maintenance/configure", func(w http.ResponseWriter, r *http.Request) {
		HandleMaintenanceConfigure(w, r, templates, isMaintenanceAuthenticated, updateDatabaseConfig)
	})

	router.HandleFunc("/maintenance/initialize-schema", func(w http.ResponseWriter, r *http.Request) {
		HandleInitializeSchema(w, r, isMaintenanceAuthenticated, initDB)
	})

	// Admin CLI - requires admin authentication
	router.Handle("/admin/cli", auth.AuthMiddleware(auth.IsAdminMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HandleAdminCLI(w, r, templates, getDB)
	}))))

	// Admin CLI API endpoints
	router.Handle("/admin/cli/api/queries", auth.AuthMiddleware(auth.IsAdminMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HandleCLIQueriesAPI(w, r)
	}))))

	router.Handle("/admin/cli/api/execute", auth.AuthMiddleware(auth.IsAdminMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HandleCLIExecuteAPI(w, r, getDB)
	}))))

	// Additional admin routes
	router.Handle("/admin/connections", auth.AuthMiddleware(auth.IsAdminMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HandleAdminConnections(w, r, templates, getDB)
	}))))

	router.Handle("/admin/tables", auth.AuthMiddleware(auth.IsAdminMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HandleAdminTables(w, r, templates, getDB)
	}))))

	router.Handle("/admin/settings", auth.AuthMiddleware(auth.IsAdminMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HandleAdminSettings(w, r, templates, getDB)
	}))))

	// // User management routes
	// router.Handle("/users", auth.AuthMiddleware(auth.IsAdminMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	HandleUsers(w, r, templates, getDB)
	// }))))

	// General routes (accessible to all authenticated users)
	router.Handle("/connections", auth.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HandleConnections(w, r, templates, getDB)
	})))

	router.Handle("/tables", auth.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HandleTables(w, r, templates, getDB)
	})))

	router.Handle("/profile", auth.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HandleProfile(w, r, templates)
	})))
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

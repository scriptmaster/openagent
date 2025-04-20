package projects

import (
	"database/sql"
	"html/template"
	"net/http"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/scriptmaster/openagent/auth"
	"github.com/scriptmaster/openagent/common"
)

// RegisterProjectRoutes registers HTML and API routes for projects
func RegisterProjectRoutes(mux *http.ServeMux, templates *template.Template, userService *auth.UserService, db *sql.DB) {
	// Initialize project repository and service
	sqlxDB := sqlx.NewDb(db, "postgres") // Or the appropriate driver
	projectRepo := NewProjectRepository(sqlxDB)
	projectService := NewProjectService(db, projectRepo)

	// --- HTML Page Routes ---
	// Handle the main /projects page (renders HTML)
	mux.Handle("/projects", auth.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HandleProjectsRoute(w, r, templates, projectService, userService)
	})))

	// Handle root index page (might list projects)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		HandleIndexRoute(w, r, templates, projectService, userService)
	})
	// Note: Project specific page routes (e.g., /project/{id}) might be handled within HandleProjectPageRoute
	// called from server/routes.go based on domain context.

	// --- API Routes (/api/projects) ---
	apiProjectHandler := func(w http.ResponseWriter, r *http.Request) {
		user := auth.GetUserFromContext(r.Context())
		if user == nil {
			common.JSONError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Basic routing based on method and path structure
		path := r.URL.Path
		if path == "/api/projects" || path == "/api/projects/" {
			switch r.Method {
			case http.MethodGet:
				HandleListProjectsAPI(w, r, projectService)
			case http.MethodPost:
				HandleCreateProjectAPI(w, r, projectService, user)
			default:
				common.JSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if strings.HasPrefix(path, "/api/projects/") {
			switch r.Method {
			case http.MethodPut:
				HandleUpdateProjectAPI(w, r, projectService)
			case http.MethodDelete:
				HandleDeleteProjectAPI(w, r, projectService)
			default:
				common.JSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			common.Handle404(w, r, templates) // Or JSON 404?
		}
	}

	// Register the API handler under /api/projects/, ensuring auth
	mux.Handle("/api/projects/", auth.AuthMiddleware(http.HandlerFunc(apiProjectHandler)))
}

// Removed placeholder for HandleProjectsRoute

// Removed placeholder for HandleIndexRoute

// Removed placeholder for HandleProjectPageRoute

package projects

import (
	"database/sql"
	"html/template"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/scriptmaster/openagent/auth"
)

// RegisterProjectRoutes registers all project-related routes
func RegisterProjectRoutes(mux *http.ServeMux, templates *template.Template, userService *auth.UserService, db *sql.DB) {
	// Initialize project repository and service
	sqlxDB := sqlx.NewDb(db, "postgres") // Or the appropriate driver
	projectRepo := NewProjectRepository(sqlxDB)
	projectService := NewProjectService(db, projectRepo)

	// Project routes
	mux.HandleFunc("/projects", func(w http.ResponseWriter, r *http.Request) {
		HandleProjectsRoute(w, r, templates, projectService, userService)
	})
}

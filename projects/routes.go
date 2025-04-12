package projects

import (
	"database/sql"
	"html/template"
	"net/http"

	"github.com/scriptmaster/openagent/auth"
)

// RegisterProjectRoutes registers all project-related routes
func RegisterProjectRoutes(mux *http.ServeMux, templates *template.Template, userService *auth.UserService, db *sql.DB) {
	// Initialize project service
	projectService, err := NewProjectService(db)
	if err != nil {
		panic("failed to initialize project service: " + err.Error())
	}

	// Project routes
	mux.HandleFunc("/projects", func(w http.ResponseWriter, r *http.Request) {
		HandleProjectsRoute(w, r, templates, projectService, userService)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		HandleIndexRoute(w, r, templates, projectService, userService)
	})
}

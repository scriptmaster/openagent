package projects

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/scriptmaster/openagent/auth"
	"github.com/scriptmaster/openagent/common"
	"github.com/scriptmaster/openagent/types"
)

// RegisterProjectRoutes registers HTML and API routes for projects
func RegisterProjectRoutes(router *http.ServeMux, templates types.TemplateEngineInterface, userService auth.UserServicer, db *sql.DB, projectDBService common.ProjectDBService) {
	if db == nil || userService == nil || projectDBService == nil {
		// Handle project routes gracefully if services are nil
		log.Println("Project routes disabled: required services (db, userService, pdbService) not available")
		router.HandleFunc("/projects", HandleNilService)
		router.HandleFunc("/api/projects/", HandleNilServiceAPI)
		return
	}

	// Initialize project repository and service
	sqlxDB := sqlx.NewDb(db, "postgres") // Or the appropriate driver
	projectRepo := NewProjectRepository(sqlxDB)
	projectService := NewProjectService(db, projectRepo)

	// --- HTML Page Routes ---
	// Handle the main /projects page (renders HTML)
	router.Handle("/projects", auth.AuthMiddleware(http.HandlerFunc(CreateProjectsHandler(templates, projectService, userService))))
	router.Handle("/api/projects/", auth.AuthMiddleware(http.HandlerFunc(CreateProjectsAPIHandler(templates, projectService, projectDBService))))
}

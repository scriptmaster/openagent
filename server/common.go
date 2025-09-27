package server

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/scriptmaster/openagent/common"
	"github.com/scriptmaster/openagent/projects"
)

// LoadTemplates loads templates using the unified template engine
func LoadTemplates() *TemplateEngine {
	return NewTemplateEngine()
}

// HandleNilService registers handlers for paths when required services are unavailable
func HandleNilService(router *http.ServeMux, paths ...string) {
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

// addCacheHeaders wraps an http.Handler to add cache headers for static files
// This is the old cache implementation - kept for backward compatibility
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

// ServiceContainer holds all the services
type ServiceContainer struct {
	ProjectService projects.ProjectService
	PDBService     common.ProjectDBService
	DB             *sql.DB
}

// GetServices initializes and returns all services
func GetServices(db *sql.DB) ServiceContainer {
	container := ServiceContainer{DB: db}

	if db != nil {
		log.Println("\t → \t → 6.1 Database connection available, initializing DB-dependent services and routes.")
		container.PDBService = NewProjectDBService(db)

		// Initialize project service (needs sqlx wrapper)
		sqlxDB := sqlx.NewDb(db, common.GetEnvOrDefault("DB_DRIVER", "postgres"))
		projectRepo := projects.NewProjectRepository(sqlxDB)
		container.ProjectService = projects.NewProjectService(db, projectRepo)
	} else {
		log.Println("Warning: Database connection is nil. DB-dependent services will not be available.")
	}

	return container
}

package projects

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/scriptmaster/openagent/auth"
	"github.com/scriptmaster/openagent/common"
	"github.com/scriptmaster/openagent/models"
)

// HandleProjectsAPI handles the /api/projects endpoint
func HandleProjectsAPI(w http.ResponseWriter, r *http.Request, service ProjectService) {
	if r.Method == http.MethodGet {
		projects, err := service.List()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(projects)
		return
	}

	if r.Method == http.MethodPost {
		var project Project
		if err := json.NewDecoder(r.Body).Decode(&project); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Generate ID via repository since it's an int64
		project.CreatedAt = time.Now()
		project.UpdatedAt = time.Now()

		id, err := service.Create(&project)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		project.ID = id
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(project)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// HandleProjectAPI handles the /api/projects/{id} endpoint
func HandleProjectAPI(w http.ResponseWriter, r *http.Request, service ProjectService) {
	idStr := r.URL.Path[len("/api/projects/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		project, err := service.GetByID(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(project)

	case http.MethodPut:
		var project Project
		if err := json.NewDecoder(r.Body).Decode(&project); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		project.ID = id
		project.UpdatedAt = time.Now()

		if err := service.Update(&project); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(project)

	case http.MethodDelete:
		if err := service.Delete(id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleProjects handles the projects page
func HandleProjects(w http.ResponseWriter, r *http.Request, templates *template.Template, service ProjectService) {
	// Get user from context
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get all projects
	projects, err := service.List()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Prepare template data using ProjectsPageData struct
	data := ProjectsPageData{
		AppName:    "OpenAgent",
		PageTitle:  "Projects",
		User:       *user,
		Projects:   projects,
		AppVersion: common.GetEnvOrDefault("APP_VERSION", "1.0.0.0"),
	}

	// Execute the template
	if err := templates.ExecuteTemplate(w, "projects.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// HandleIndex handles the root route
func HandleIndex(w http.ResponseWriter, r *http.Request, templates *template.Template, service ProjectService) {
	// Get user from context
	user := auth.GetUserFromContext(r.Context())

	// If user is logged in, redirect to dashboard
	if user != nil {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	// Prepare page data with defaults
	data := models.PageData{
		AppName:    common.GetEnvOrDefault("APP_NAME", "OpenAgent"),
		PageTitle:  "Welcome to OpenAgent",
		AppVersion: common.GetEnvOrDefault("APP_VERSION", "1.0.0.0"),
	}

	// Try to get project based on host
	host := r.Host
	if host != "" {
		project, err := service.GetByDomain(host)
		if err == nil && project != nil {
			data.Project = project
			data.PageTitle = project.Name
		}
	}

	// Execute the template
	if err := templates.ExecuteTemplate(w, "index.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// HandleProjectsRoute handles the /projects route with authentication
func HandleProjectsRoute(w http.ResponseWriter, r *http.Request, templates *template.Template, projectService ProjectService, userService *auth.UserService) {
	// Get user from session
	user, err := userService.GetUserFromSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Add user to request context
	ctx := auth.SetUserContext(r.Context(), user)
	HandleProjects(w, r.WithContext(ctx), templates, projectService)
}

// HandleIndexRoute handles the root route with optional authentication
func HandleIndexRoute(w http.ResponseWriter, r *http.Request, templates *template.Template, projectService ProjectService, userService *auth.UserService) {
	// Get user from session
	user, err := userService.GetUserFromSession(r)
	if err != nil {
		// For index page, we don't redirect to login
		HandleIndex(w, r, templates, projectService)
		return
	}

	// Add user to request context
	ctx := auth.SetUserContext(r.Context(), user)
	HandleIndex(w, r.WithContext(ctx), templates, projectService)
}

// HandleProjectPageRoute handles project-specific pages based on domain and path
func HandleProjectPageRoute(w http.ResponseWriter, r *http.Request, templates *template.Template, projectService ProjectService, userService *auth.UserService) {
	// Get project by current domain
	project, err := projectService.GetByDomain(r.Host)
	if err != nil || project == nil {
		// No project found for this domain, serve 404 page
		common.Handle404(w, r, templates)
		return
	}

	// Get user from session if available
	user, _ := userService.GetUserFromSession(r)
	if user != nil {
		// Add user to request context
		ctx := auth.SetUserContext(r.Context(), user)
		r = r.WithContext(ctx)
	}

	// Prepare page data
	data := models.PageData{
		AppName:    common.GetEnvOrDefault("APP_NAME", "OpenAgent"),
		PageTitle:  project.Name,
		AppVersion: common.GetEnvOrDefault("APP_VERSION", "1.0.0.0"),
		Project:    project,
		User:       user,
	}

	// Execute the template
	if err := templates.ExecuteTemplate(w, "index.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

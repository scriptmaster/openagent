package projects

import (
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/scriptmaster/openagent/auth"
	"github.com/scriptmaster/openagent/common"
	"github.com/scriptmaster/openagent/models"
)

// --- Interfaces (to avoid import cycles) ---

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
		log.Printf("Error fetching projects list: %v", err)
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
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	HandleProjects(w, r, templates, projectService)
}

// HandleIndexRoute handles the root route with optional authentication
func HandleIndexRoute(w http.ResponseWriter, r *http.Request, templates *template.Template, projectService ProjectService, userService *auth.UserService) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		log.Println("No user context found for HandleIndexRoute")
	}
	HandleIndex(w, r, templates, projectService)
}

// HandleProjectPageRoute handles project-specific pages based on domain and path
func HandleProjectPageRoute(w http.ResponseWriter, r *http.Request, templates *template.Template, projectService ProjectService, userService *auth.UserService) {
	user := auth.GetUserFromContext(r.Context())
	project := GetProjectFromContext(r.Context())

	if project == nil {
		log.Printf("WARN: HandleProjectPageRoute called without project context for host %s, path %s", r.Host, r.URL.Path)
		common.Handle404(w, r, templates)
		return
	}

	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
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

// --- API Handlers ---

// HandleListProjectsAPI handles GET requests to list projects as JSON
func HandleListProjectsAPI(w http.ResponseWriter, r *http.Request, projectService ProjectService) {
	projects, err := projectService.List()
	if err != nil {
		log.Printf("API Error fetching projects: %v", err)
		common.JSONError(w, "Failed to fetch projects", http.StatusInternalServerError)
		return
	}
	common.JSONResponse(w, projects)
}

// HandleCreateProjectAPI handles POST requests to create a new project
func HandleCreateProjectAPI(w http.ResponseWriter, r *http.Request, projectService ProjectService, user *auth.User) {
	var project Project
	if err := json.NewDecoder(r.Body).Decode(&project); err != nil {
		common.JSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validation
	if project.Name == "" || project.Domain == "" {
		common.JSONError(w, "Project Name and Domain are required", http.StatusBadRequest)
		return
	}

	// Set creator ID (assuming middleware provides user)
	if user == nil {
		common.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	project.CreatedBy = int64(user.ID)
	project.IsActive = true // Default to active on creation?

	newID, err := projectService.Create(&project)
	if err != nil {
		log.Printf("API Error creating project: %v", err)
		common.JSONError(w, "Failed to create project: "+err.Error(), http.StatusInternalServerError)
		return
	}

	project.ID = newID
	w.WriteHeader(http.StatusCreated)
	common.JSONResponse(w, project)
}

// HandleUpdateProjectAPI handles PUT requests to update a project
func HandleUpdateProjectAPI(w http.ResponseWriter, r *http.Request, projectService ProjectService) {
	// Extract project ID from URL path, e.g., /api/projects/123
	path := r.URL.Path
	prefix := "/api/projects/"
	if !strings.HasPrefix(path, prefix) {
		common.JSONError(w, "Invalid URL path prefix for project update", http.StatusBadRequest)
		return
	}
	idStr := strings.TrimPrefix(path, prefix)

	projectID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		common.JSONError(w, "Invalid project ID in URL", http.StatusBadRequest)
		return
	}

	var projectUpdates Project
	if err := json.NewDecoder(r.Body).Decode(&projectUpdates); err != nil {
		common.JSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Ensure the ID from the body matches the URL (or set it)
	projectUpdates.ID = projectID

	// Basic validation
	if projectUpdates.Name == "" || projectUpdates.Domain == "" {
		common.JSONError(w, "Project Name and Domain are required", http.StatusBadRequest)
		return
	}

	err = projectService.Update(&projectUpdates)
	if err != nil {
		log.Printf("API Error updating project %d: %v", projectID, err)
		if errors.Is(err, ErrProjectNotFound) {
			common.JSONError(w, "Project not found", http.StatusNotFound)
		} else {
			common.JSONError(w, "Failed to update project: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	common.JSONResponse(w, projectUpdates) // Return updated project
}

// HandleDeleteProjectAPI handles DELETE requests to delete a project
func HandleDeleteProjectAPI(w http.ResponseWriter, r *http.Request, projectService ProjectService) {
	// Extract project ID from URL path, e.g., /api/projects/123
	path := r.URL.Path
	prefix := "/api/projects/"
	if !strings.HasPrefix(path, prefix) {
		common.JSONError(w, "Invalid URL path prefix for project deletion", http.StatusBadRequest)
		return
	}
	idStr := strings.TrimPrefix(path, prefix)

	projectID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		common.JSONError(w, "Invalid project ID in URL", http.StatusBadRequest)
		return
	}

	err = projectService.Delete(projectID)
	if err != nil {
		log.Printf("API Error deleting project %d: %v", projectID, err)
		if errors.Is(err, ErrProjectNotFound) {
			// Arguably, deleting a non-existent resource is not an error (idempotent)
			w.WriteHeader(http.StatusNoContent)
		} else {
			common.JSONError(w, "Failed to delete project: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent) // Success, no content to return
}

// UpdateProjectDBConfigRequest represents the expected JSON body
type UpdateProjectDBConfigRequest struct {
	DBType           string `json:"db_type"`           // e.g., postgres, mysql
	SchemaName       string `json:"schema_name"`       // New schema name
	ConnectionString string `json:"connection_string"` // Base64 encoded new connection string
	// Fields for options A, B, C could be combined or use a type field
	// We'll use presence/absence of fields for now
}

// HandleUpdateProjectDBConfigAPI handles PUT /api/projects/{id}/dbconfig
// Accepts common.ProjectDBService injected from routes
func HandleUpdateProjectDBConfigAPI(w http.ResponseWriter, r *http.Request, projectService ProjectService, projectDBService common.ProjectDBService) {
	// Extract project ID from URL
	path := r.URL.Path
	prefix := "/api/projects/"
	suffix := "/dbconfig"
	if !strings.HasPrefix(path, prefix) || !strings.HasSuffix(path, suffix) {
		common.JSONError(w, "Invalid URL path", http.StatusBadRequest)
		return
	}
	idStr := strings.TrimSuffix(strings.TrimPrefix(path, prefix), suffix)
	projectID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		common.JSONError(w, "Invalid project ID in URL", http.StatusBadRequest)
		return
	}

	// Decode request body
	var req UpdateProjectDBConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.JSONError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Use the injected projectDBService directly
	if projectDBService == nil {
		log.Println("ERROR: HandleUpdateProjectDBConfigAPI called with nil projectDBService")
		common.JSONError(w, "Database service not available", http.StatusInternalServerError)
		return
	}

	// --- Feature Logic ---
	// 1. Find the *default* ProjectDB for the given projectID
	//    (Need a method in ProjectDBService for this, e.g., GetProjectDBs(projectID) for now and find the default.
	dbs, err := projectDBService.GetProjectDBs(int(projectID))
	if err != nil {
		log.Printf("Error fetching project DBs for project %d: %v", projectID, err)
		common.JSONError(w, "Failed to retrieve project database configurations", http.StatusInternalServerError)
		return
	}

	var defaultDB *common.ProjectDB // Use common.ProjectDB type
	for i := range dbs {
		if dbs[i].IsDefault {
			tempDB := dbs[i]    // Create a temporary variable
			defaultDB = &tempDB // Assign the address of the temp variable
			break
		}
	}

	if defaultDB == nil {
		log.Printf("No default database configuration found for project %d", projectID)
		common.JSONError(w, "No default database configuration found for this project", http.StatusNotFound)
		return
	}

	// 2. Determine which option (A, B, C) based on request fields
	// Option C: Change entire connection (db_type and connection_string provided)
	if req.DBType != "" && req.ConnectionString != "" {
		log.Printf("Handling DB Config Option C for project %d, DB ID %d", projectID, defaultDB.ID)
		// Decode connection string from request (assuming base64)
		// TODO: Test the new connection string before saving?
		err = projectDBService.UpdateProjectDB(
			defaultDB.ID,
			defaultDB.Name,        // Keep original name?
			defaultDB.Description, // Keep original desc?
			req.DBType,
			req.ConnectionString, // Save encoded string
			req.SchemaName,       // Use new schema name if provided, else keep old?
			defaultDB.IsDefault,  // Keep default status
		)
		// TODO: Add more logic for SchemaName update if only DBType/ConnStr provided
		// For now, assume SchemaName is always provided with ConnStr in Option C

		// Option B: Change only schema (schema_name provided, others empty)
	} else if req.SchemaName != "" && req.DBType == "" && req.ConnectionString == "" {
		log.Printf("Handling DB Config Option B for project %d, DB ID %d", projectID, defaultDB.ID)
		err = projectDBService.UpdateProjectDB(
			defaultDB.ID,
			defaultDB.Name,
			defaultDB.Description,
			defaultDB.DBType,           // Keep old type
			defaultDB.ConnectionString, // Keep old connection string
			req.SchemaName,             // Use new schema name
			defaultDB.IsDefault,
		)
		// Option A: Change DB name/schema (using *current* conn string - not implemented fully based on spec E)
		// Spec says "don't expose connection string on UI".
		// Backend *could* re-use existing string, but how to get DB name change?
		// For now, Option A is implicitly handled by Option B if only schema changes,
		// or Option C if type/conn changes. A specific DB name change isn't handled.

	} else {
		common.JSONError(w, "Invalid combination of parameters for DB config update", http.StatusBadRequest)
		return
	}

	// 3. Call UpdateProjectDB with appropriate parameters
	if err != nil {
		log.Printf("Error updating project DB config for project %d, DB ID %d: %v", projectID, defaultDB.ID, err)
		common.JSONError(w, "Failed to update database configuration: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Fetch the updated DB to return it
	updatedDB, err := projectDBService.GetProjectDB(defaultDB.ID)
	if err != nil {
		log.Printf("Error fetching updated project DB %d: %v", defaultDB.ID, err)
		common.JSONResponse(w, map[string]string{"message": "Database configuration updated successfully"})
		return
	}

	common.JSONResponse(w, updatedDB)
}

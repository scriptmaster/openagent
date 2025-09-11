package projects

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/scriptmaster/openagent/auth"
	"github.com/scriptmaster/openagent/common"
	"github.com/scriptmaster/openagent/models"
	"github.com/scriptmaster/openagent/types"
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
func HandleProjects(w http.ResponseWriter, r *http.Request, templates types.TemplateEngineInterface, service ProjectService) {
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
func HandleIndex(w http.ResponseWriter, r *http.Request, templates types.TemplateEngineInterface, service ProjectService) {
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
func HandleProjectsRoute(w http.ResponseWriter, r *http.Request, templates types.TemplateEngineInterface, projectService ProjectService, userService auth.UserServicer) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	HandleProjects(w, r, templates, projectService)
}

// HandleIndexRoute handles the root route with optional authentication
func HandleIndexRoute(w http.ResponseWriter, r *http.Request, templates types.TemplateEngineInterface, projectService ProjectService, userService auth.UserServicer) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		log.Println("No user context found for HandleIndexRoute")
	}
	HandleIndex(w, r, templates, projectService)
}

// HandleProjectPageRoute handles project-specific pages based on domain and path
func HandleProjectPageRoute(w http.ResponseWriter, r *http.Request, templates types.TemplateEngineInterface, projectService ProjectService, userService auth.UserServicer) {
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

// UpdateProjectDBConfigRequest defines the structure for the request body
// when updating a project's database configuration.
type UpdateProjectDBConfigRequest struct {
	UpdateType       string `json:"update_type"`                 // "schema_only", "db_name_schema", "full_connection"
	DBType           string `json:"db_type,omitempty"`           // Required for "full_connection", e.g., postgres, mysql
	SchemaName       string `json:"schema_name"`                 // New schema name, required for all types
	DBName           string `json:"db_name,omitempty"`           // New DB name, required for "db_name_schema"
	ConnectionString string `json:"connection_string,omitempty"` // Base64 encoded new full connection string, required for "full_connection"
	ProjectDBID      int64  `json:"project_db_id"`               // ID of the ai.project_dbs record to update
	Name             string `json:"name,omitempty"`              // Optional: New name for the ProjectDB entry itself
	Description      string `json:"description,omitempty"`       // Optional: New description for the ProjectDB entry
	IsDefault        *bool  `json:"is_default,omitempty"`        // Optional: Make this connection the default for the project
}

// HandleUpdateProjectDBConfigAPI handles requests to update a project's database configuration.
func HandleUpdateProjectDBConfigAPI(w http.ResponseWriter, r *http.Request, projectService ProjectService, projectDBService common.ProjectDBService) {
	// Extract project ID from URL (This is the parent project's ID)
	path := r.URL.Path
	prefix := "/api/projects/"
	suffix := "/dbconfig"
	if !strings.HasPrefix(path, prefix) || !strings.HasSuffix(path, suffix) {
		common.JSONError(w, "Invalid URL path format", http.StatusBadRequest)
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

	// --- Service Check ---
	if projectDBService == nil {
		log.Println("ERROR: HandleUpdateProjectDBConfigAPI called with nil projectDBService")
		common.JSONError(w, "Database service not available", http.StatusInternalServerError)
		return
	}
	if projectService == nil {
		log.Println("ERROR: HandleUpdateProjectDBConfigAPI called with nil projectService")
		common.JSONError(w, "Project service not available", http.StatusInternalServerError)
		return
	}

	// --- Input Validation ---
	if req.ProjectDBID == 0 {
		common.JSONError(w, "Missing required field: project_db_id", http.StatusBadRequest)
		return
	}
	if req.SchemaName == "" {
		common.JSONError(w, "Missing required field: schema_name", http.StatusBadRequest)
		return
	}

	var newEncodedConnStr string
	var newDbType string

	// Fetch the existing ProjectDB record
	// Note: projectDBService methods use int, but projectID and req.ProjectDBID are int64
	existingProjectDB, err := projectDBService.GetProjectDB(int(req.ProjectDBID))
	if err != nil {
		log.Printf("Error fetching project DB %d: %v", req.ProjectDBID, err)
		common.JSONError(w, fmt.Sprintf("Failed to retrieve project database configuration with ID %d", req.ProjectDBID), http.StatusNotFound)
		return
	}

	// Ensure the ProjectDB belongs to the project in the URL
	if existingProjectDB.ProjectID != int(projectID) {
		common.JSONError(w, "ProjectDB record does not belong to the specified project", http.StatusForbidden)
		return
	}

	// --- Feature Logic based on UpdateType ---
	switch req.UpdateType {
	case "schema_only": // Option B
		if existingProjectDB.ConnectionString == "" {
			common.JSONError(w, "Cannot update schema_only for a ProjectDB with no existing connection string", http.StatusBadRequest)
			return
		}
		// Keep existing connection string and DB type
		newEncodedConnStr = existingProjectDB.ConnectionString
		newDbType = existingProjectDB.DBType

	case "db_name_schema": // Option A
		if req.DBName == "" {
			common.JSONError(w, "Missing required field for db_name_schema update: db_name", http.StatusBadRequest)
			return
		}
		if existingProjectDB.ConnectionString == "" {
			common.JSONError(w, "Cannot update db_name_schema for a ProjectDB with no existing connection string", http.StatusBadRequest)
			return
		}

		// Decode existing connection string
		decodedConnStr, err := projectDBService.DecodeConnectionString(existingProjectDB.ConnectionString)
		if err != nil {
			log.Printf("Error decoding existing connection string for ProjectDB %d: %v", existingProjectDB.ID, err)
			common.JSONError(w, "Internal error processing existing configuration", http.StatusInternalServerError)
			return
		}

		// Update the DB name part (this requires parsing the conn string, which is driver-specific)
		// TODO: Implement robust connection string parsing and modification based on DBType
		// For now, assume a simple key=value format for demonstration (e.g., for postgres)
		updatedConnStr := updateConnectionStringDBName(decodedConnStr, req.DBName, existingProjectDB.DBType)
		if updatedConnStr == "" { // Indicate failure
			common.JSONError(w, fmt.Sprintf("Failed to update database name in connection string for type %s", existingProjectDB.DBType), http.StatusInternalServerError)
			return
		}

		// Re-encode the modified connection string
		newEncodedConnStr = common.EncodeConnectionString(updatedConnStr)
		newDbType = existingProjectDB.DBType

	case "full_connection": // Option C
		if req.DBType == "" {
			common.JSONError(w, "Missing required field for full_connection update: db_type", http.StatusBadRequest)
			return
		}
		if req.ConnectionString == "" {
			common.JSONError(w, "Missing required field for full_connection update: connection_string", http.StatusBadRequest)
			return
		}
		// Validate if the provided string is valid base64 before using it
		_, err := common.DecodeConnectionString(req.ConnectionString) // Use the service decoder for consistency
		if err != nil {
			common.JSONError(w, "Invalid base64 encoding for connection_string", http.StatusBadRequest)
			return
		}

		newEncodedConnStr = req.ConnectionString
		newDbType = req.DBType

	default:
		common.JSONError(w, fmt.Sprintf("Invalid update_type: %s. Must be one of 'schema_only', 'db_name_schema', 'full_connection'", req.UpdateType), http.StatusBadRequest)
		return
	}

	// --- Prepare fields for UpdateProjectDB ---
	// Use existing values unless overridden by the request
	name := existingProjectDB.Name
	if req.Name != "" {
		name = req.Name
	}
	description := existingProjectDB.Description
	if req.Description != "" {
		description = req.Description
	}
	isDefault := existingProjectDB.IsDefault
	if req.IsDefault != nil {
		isDefault = *req.IsDefault
	}

	// --- Test the new connection (optional but recommended) ---
	// Create a temporary ProjectDB struct with the new details to test
	testDB := common.ProjectDB{
		DBType:           newDbType,
		ConnectionString: newEncodedConnStr,
		SchemaName:       req.SchemaName, // Use the new schema name for testing
	}
	if err := projectDBService.TestConnection(testDB); err != nil {
		log.Printf("New DB connection test failed for ProjectDB %d update: %v", req.ProjectDBID, err)
		// Provide a more specific error message if possible
		common.JSONError(w, fmt.Sprintf("Failed to connect using the new database configuration: %v", err), http.StatusBadRequest)
		return
	}

	// --- Update the ProjectDB record in the database ---
	err = projectDBService.UpdateProjectDB(
		int(req.ProjectDBID),
		name,
		description,
		newDbType,
		newEncodedConnStr,
		req.SchemaName, // Update with the new schema name from the request
		isDefault,
	)
	if err != nil {
		log.Printf("Error updating ProjectDB %d: %v", req.ProjectDBID, err)
		common.JSONError(w, "Failed to save database configuration changes", http.StatusInternalServerError)
		return
	}

	// --- Update Project Options (Requirement D) ---
	// Fetch the parent project
	project, err := projectService.GetByID(projectID)
	if err != nil {
		// Log the error, but proceed with success response as DB config was saved
		log.Printf("WARN: Failed to fetch project %d after updating DB config %d: %v", projectID, req.ProjectDBID, err)
	} else {
		if project.Options == nil {
			project.Options = make(ProjectOptions)
		}
		// Store relevant *new* config details in the project's Options JSONB field
		// We store under a key related to the ProjectDB ID for clarity
		dbConfigKey := fmt.Sprintf("db_config_%d", req.ProjectDBID)
		project.Options[dbConfigKey] = map[string]interface{}{
			"name":              name,
			"db_type":           newDbType,
			"schema_name":       req.SchemaName,
			"connection_string": newEncodedConnStr, // Store the encoded string
			"is_default":        isDefault,
			"updated_at":        time.Now().UTC().Format(time.RFC3339),
		}

		if err := projectService.Update(project); err != nil {
			log.Printf("WARN: Failed to update project options for project %d after updating DB config %d: %v", projectID, req.ProjectDBID, err)
			// Don't fail the request, but log the warning
		}
	}

	// --- Success Response ---
	// Optionally return the updated ProjectDB object
	updatedProjectDB, _ := projectDBService.GetProjectDB(int(req.ProjectDBID)) // Ignore error, best effort
	common.JSONResponse(w, map[string]interface{}{
		"message": "Database configuration updated successfully",
		"data":    updatedProjectDB, // Send back the (potentially updated) data
	})
}

// updateConnectionStringDBName is a helper function to modify the database name
// in a connection string. This is a simplified example and needs robust implementation.
// It returns an empty string on failure.
func updateConnectionStringDBName(connStr, newDBName, dbType string) string {
	// VERY basic example for PostgreSQL key=value format
	if dbType == "postgres" || dbType == "postgresql" {
		parts := strings.Fields(connStr) // Split by space
		found := false
		for i, part := range parts {
			if strings.HasPrefix(part, "dbname=") {
				parts[i] = "dbname=" + newDBName
				found = true
				break
			}
		}
		if !found {
			// dbname might not be present, append it
			parts = append(parts, "dbname="+newDBName)
		}
		return strings.Join(parts, " ")
	}

	// Add logic for other database types (MySQL, MSSQL etc.) here
	log.Printf("WARN: updateConnectionStringDBName not implemented for dbType: %s", dbType)
	return "" // Indicate failure for unsupported types
}

/*
// HandleProjectSettingsAPI handles settings specific to a project
func HandleProjectSettingsAPI(w http.ResponseWriter, r *http.Request, settingsService common.SettingsService) {
	// ... existing code ...
}
*/

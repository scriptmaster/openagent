package server

import (
	"log"
	"net/http"

	"encoding/json"
	"strings"
	"time"

	"github.com/scriptmaster/openagent/auth"
	"github.com/scriptmaster/openagent/common"
	"github.com/scriptmaster/openagent/models"
	"github.com/scriptmaster/openagent/projects"
)

// DashboardPageData holds data specific to the dashboard template
type DashboardPageData struct {
	models.PageData // Embed common page data
	ProjectCount    int
}

// HandleDashboard serves the main dashboard page
func HandleDashboard(w http.ResponseWriter, r *http.Request, projectService projects.ProjectService) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		// Should be caught by middleware, but double-check
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if projectService == nil {
		log.Printf("WARN: ProjectService is nil in HandleDashboard")
		// Render dashboard with 0 projects or show an error?
		// For now, render with 0
		data := DashboardPageData{
			PageData: models.PageData{
				AppName:    common.GetEnvOrDefault("APP_NAME", "OpenAgent"),
				PageTitle:  "Dashboard",
				User:       user,
				AppVersion: appVersion,
			},
			ProjectCount: 0,
		}
		if err := templates.ExecuteTemplate(w, "layout.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Fetch projects to get the count
	projectList, err := projectService.List()
	if err != nil {
		log.Printf("Error fetching project list for dashboard: %v", err)
		// Render dashboard but maybe show an error getting count?
		// For now, show 0
		projectList = []*projects.Project{}
	}

	data := DashboardPageData{
		PageData: models.PageData{
			AppName:    common.GetEnvOrDefault("APP_NAME", "OpenAgent"),
			PageTitle:  "Dashboard",
			User:       user,
			AppVersion: appVersion, // Use global appVersion loaded in routes
		},
		ProjectCount: len(projectList),
	}

	// Execute the template
	if err := templates.ExecuteTemplate(w, "layout.html", data); err != nil { // Assuming layout includes dashboard
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleVoicePage serves the voice agent page
func HandleVoicePage(w http.ResponseWriter, r *http.Request) {
	// Ensure templates are initialized (assuming 'templates' is the global var)
	if templates == nil {
		http.Error(w, "Templates not initialized", http.StatusInternalServerError)
		log.Println("Error: HandleVoicePage called before templates were initialized")
		return
	}

	// Execute the voice template
	// You might want to pass data similar to other pages if needed (e.g., AppName, User)
	err := templates.ExecuteTemplate(w, "voice.html", nil) // Passing nil data for now
	if err != nil {
		log.Printf("Error executing voice template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// HandleConfigPage serves the initial configuration page
func HandleConfigPage(w http.ResponseWriter, r *http.Request) {
	// Get user from context if available (might be nil)
	user := auth.GetUserFromContext(r.Context())

	// Ensure templates are initialized
	if templates == nil {
		http.Error(w, "Templates not initialized", http.StatusInternalServerError)
		log.Println("Error: HandleConfigPage called before templates were initialized")
		return
	}

	// Prepare data for the template
	data := models.PageData{ // Using generic PageData, adapt if needed
		AppName:    common.GetEnvOrDefault("APP_NAME", "OpenAgent"),
		PageTitle:  "System Configuration",
		User:       user, // Pass user info if available
		AppVersion: common.GetEnvOrDefault("APP_VERSION", "1.0.0.0"),
		// Add any specific flags or data needed for config page
		// e.g., pass the current host?
		CurrentHost: strings.Split(r.Host, ":")[0],
	}

	// Execute the config template
	err := templates.ExecuteTemplate(w, "config.html", data)
	if err != nil {
		log.Printf("Error executing config template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// ConfigSubmitRequest defines the structure for the config form submission
type ConfigSubmitRequest struct {
	AdminToken    string   `json:"admin_token"`
	ProjectName   string   `json:"project_name"`
	ProjectDesc   string   `json:"project_desc"`
	PrimaryHost   string   `json:"primary_host"`
	AdminEmail    string   `json:"admin_email"`
	AdminPassword string   `json:"admin_password"`
	RedirectHosts []string `json:"redirect_hosts"` // Assuming frontend sends as array
	AliasHosts    []string `json:"alias_hosts"`    // Assuming frontend sends as array
}

// HandleConfigSubmit processes the configuration form submission
func HandleConfigSubmit(w http.ResponseWriter, r *http.Request, userService *auth.UserService, projectService projects.ProjectService) {
	if r.Method != http.MethodPost {
		common.JSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ConfigSubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.JSONError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// --- Validation ---
	if req.AdminToken == "" || req.ProjectName == "" || req.PrimaryHost == "" || req.AdminEmail == "" || req.AdminPassword == "" {
		common.JSONError(w, "Missing required fields (Token, Project Name, Primary Host, Admin Email, Admin Password)", http.StatusBadRequest)
		return
	}

	// --- Verify Admin Token ---
	db := GetDB() // Assuming GetDB() retrieves the initialized *sql.DB
	if db == nil {
		common.JSONError(w, "Database connection not available", http.StatusInternalServerError)
		return
	}
	today := time.Now().UTC().Format("2006-01-02")
	validToken, err := GetAdminTokenForDate(db, today)
	if err != nil {
		log.Printf("Error retrieving admin token for validation: %v", err)
		common.JSONError(w, "Error validating setup token", http.StatusInternalServerError)
		return
	}
	if validToken == "" || req.AdminToken != validToken {
		common.JSONError(w, "Invalid or expired setup token", http.StatusForbidden)
		return
	}

	// --- Create Admin User ---
	// Use the CreateUser function that handles hashing
	adminUser, err := userService.CreateUser(req.AdminEmail, req.AdminPassword)
	if err != nil {
		// Handle potential "user already exists" errors more gracefully if needed
		log.Printf("Error creating admin user '%s': %v", req.AdminEmail, err)
		common.JSONError(w, "Failed to create admin user: "+err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Admin user %s created successfully.", adminUser.Email)
	// Ensure the user is explicitly marked as admin (CreateUser should handle first user, but maybe force here?)
	if !adminUser.IsAdmin {
		err = userService.MakeUserAdmin(adminUser.Email) // Make sure MakeUserAdmin exists and works
		if err != nil {
			log.Printf("WARN: Failed to ensure user %s is admin: %v", adminUser.Email, err)
			// Decide if this is fatal. Maybe proceed?
		} else {
			log.Printf("User %s explicitly set as admin.", adminUser.Email)
		}
	}

	// --- Create Project ---
	// Create the options map
	projectOptions := projects.ProjectOptions{
		"redirect_hosts": req.RedirectHosts,
		"alias_hosts":    req.AliasHosts,
	}

	newProject := &projects.Project{
		Name:        req.ProjectName,
		Description: req.ProjectDesc,
		Domain:      req.PrimaryHost,
		CreatedBy:   int64(adminUser.ID),
		IsActive:    true,
		Options:     projectOptions, // Assign the options map
	}

	projectID, err := projectService.Create(newProject)
	if err != nil {
		log.Printf("Error creating project '%s' for domain '%s': %v", newProject.Name, newProject.Domain, err)
		common.JSONError(w, "Failed to create project: "+err.Error(), http.StatusInternalServerError)
		// Consider rolling back user creation? Or inform user the project failed?
		return
	}
	log.Printf("Project %d created successfully: %s (%s)", projectID, newProject.Name, newProject.Domain)

	common.JSONResponse(w, map[string]string{"message": "Configuration successful! Please log in."})
}

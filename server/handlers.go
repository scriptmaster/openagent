package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/scriptmaster/openagent/admin"
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
		if err := globalTemplates.ExecuteTemplate(w, "dashboard.html", data); err != nil {
			log.Printf("Error executing layout template for dashboard: %v", err)
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

	// Execute the dashboard template (which will use the logged-in layout)
	if err := globalTemplates.ExecuteTemplate(w, "dashboard.html", data); err != nil {
		log.Printf("Error executing dashboard template: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleVoicePage serves the voice agent page
func HandleVoicePage(w http.ResponseWriter, r *http.Request) {
	// Ensure templates are initialized (assuming 'templates' is the global var)
	if globalTemplates == nil {
		http.Error(w, "Templates not initialized", http.StatusInternalServerError)
		log.Println("Error: HandleVoicePage called before templates were initialized")
		return
	}

	// Execute the voice template
	// You might want to pass data similar to other pages if needed (e.g., AppName, User)
	err := globalTemplates.ExecuteTemplate(w, "voice.html", nil) // Passing nil data for now
	if err != nil {
		log.Printf("Error executing voice template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// HandleAgentPage serves the agent page
func HandleAgentPage(w http.ResponseWriter, r *http.Request) {
	// Ensure templates are initialized (assuming 'templates' is the global var)
	if globalTemplates == nil {
		http.Error(w, "Templates not initialized", http.StatusInternalServerError)
		log.Println("Error: HandleAgentPage called before templates were initialized")
		return
	}

	// Execute the agent template
	// You might want to pass data similar to other pages if needed (e.g., AppName, User)
	err := globalTemplates.ExecuteTemplate(w, "agent.html", nil) // Passing nil data for now
	if err != nil {
		log.Printf("Error executing agent template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// HandleTestPage serves a test page to verify template system
func HandleTestPage(w http.ResponseWriter, r *http.Request) {
	// log.Printf("\t → \t → 6.20 Route: /test handler")
	if globalTemplates == nil {
		http.Error(w, "Templates not initialized", http.StatusInternalServerError)
		return
	}

	// Set cache headers for the page
	SetCacheHeaders(w, r, r.URL.Path)

	data := models.PageData{
		AppName:    common.GetEnv("APP_NAME"),
		PageTitle:  "Test Page",
		AppVersion: AppVersion,
	}

	err := globalTemplates.ExecuteTemplate(w, "test.html", data)
	if err != nil {
		log.Printf("Error executing test template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// HandleConfigPage serves the initial configuration page
func HandleConfigPage(w http.ResponseWriter, r *http.Request) {
	// Get user from context if available (might be nil)
	user := auth.GetUserFromContext(r.Context())

	// Ensure templates are initialized
	if globalTemplates == nil {
		http.Error(w, "Templates not initialized", http.StatusInternalServerError)
		log.Println("Error: HandleConfigPage called before templates were initialized")
		return
	}

	// Prepare data for the template
	data := models.PageData{ // Using generic PageData, adapt if needed
		AppName:    common.GetEnv("APP_NAME"),
		PageTitle:  "System Configuration",
		User:       user, // Pass user info if available
		AppVersion: AppVersion,
		// Add any specific flags or data needed for config page
		// e.g., pass the current host?
		CurrentHost: strings.Split(r.Host, ":")[0],
	}

	// Execute the config page template
	err := globalTemplates.ExecuteTemplate(w, "config.html", data)
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

// HandleConfigSubmit handles the configuration form submission
func HandleConfigSubmit(w http.ResponseWriter, r *http.Request, userService auth.UserServicer, projectService projects.ProjectService) {
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

	// Create or find the admin user
	// Use GetUserByEmail first, then CreateUser if not found
	adminUser, err := userService.GetUserByEmail(r.Context(), req.AdminEmail)
	if err != nil {
		if strings.Contains(err.Error(), "user not found") {
			log.Printf("Admin user %s not found, creating...", req.AdminEmail)
			adminUser, err = userService.CreateUser(r.Context(), req.AdminEmail) // Pass context
			if err != nil {
				common.JSONError(w, "Failed to create admin user: "+err.Error(), http.StatusInternalServerError)
				return
			}
			// Ensure the newly created user is admin
			if !adminUser.IsAdmin {
				err = userService.MakeUserAdmin(r.Context(), adminUser.ID) // Use new method
				if err != nil {
					log.Printf("Failed to make user %d (%s) admin: %v", adminUser.ID, adminUser.Email, err)
					// Continue but log error
				}
			}
		} else {
			common.JSONError(w, "Error checking admin user: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// User exists, ensure they are admin
		if !adminUser.IsAdmin {
			err = userService.MakeUserAdmin(r.Context(), adminUser.ID) // Use new method
			if err != nil {
				log.Printf("Failed to make existing user %d (%s) admin: %v", adminUser.ID, adminUser.Email, err)
				// Continue but log error
			}
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

// HandleVersion serves the application version information
func HandleVersion(w http.ResponseWriter, r *http.Request) {
	// log.Printf("\t → \t → 6.10 Route: /version handler")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"version": "%s", "timestamp": "%s"}`, AppVersion, time.Now().UTC().Format(time.RFC3339))
}

// Handle404 serves a custom 404 page
func Handle404(w http.ResponseWriter, r *http.Request) {
	if globalTemplates == nil {
		http.Error(w, "Templates not initialized", http.StatusInternalServerError)
		return
	}

	// Set cache headers for the page
	SetCacheHeaders(w, r, r.URL.Path)

	// Create 404 page data
	pageData := models.PageData{
		AppName:    "OpenAgent",
		AppVersion: AppVersion,
		PageTitle:  "Page Not Found",
		User:       nil,
	}

	// Render the 404 template
	if err := globalTemplates.ExecuteTemplate(w, "error_404.html", pageData); err != nil {
		log.Printf("Error rendering 404 page: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// HandleFavicon serves the favicon
func HandleFavicon(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/favicon.ico")
}

// HandleTSXCSS serves generated CSS files with cache headers
func HandleTSXCSS(w http.ResponseWriter, r *http.Request) {
	// Set cache headers using our comprehensive cache system
	SetCacheHeaders(w, r, r.URL.Path)
	w.Header().Set("X-Content-Type-Options", "nosniff")

	// Extract filename from path
	filename := strings.TrimPrefix(r.URL.Path, "/tsx/css/")
	if filename == "" {
		http.NotFound(w, r)
		return
	}

	// Serve from generated/css directory
	filePath := fmt.Sprintf("./tpl/generated/css/%s", filename)

	// Generate ETag for the file
	etag, err := GenerateETag(filePath)
	if err == nil {
		// Handle conditional request
		if HandleConditionalRequest(w, r, etag) {
			return // 304 Not Modified
		}
	}

	http.ServeFile(w, r, filePath)
}

// HandleTSXJS serves generated JS files with cache headers
func HandleTSXJS(w http.ResponseWriter, r *http.Request) {
	// Set cache headers using our comprehensive cache system
	SetCacheHeaders(w, r, r.URL.Path)
	w.Header().Set("X-Content-Type-Options", "nosniff")

	// Extract filename from path
	filename := strings.TrimPrefix(r.URL.Path, "/tsx/js/")
	if filename == "" {
		http.NotFound(w, r)
		return
	}

	// Serve from generated/js directory
	filePath := fmt.Sprintf("./tpl/generated/js/%s", filename)

	// Generate ETag for the file
	etag, err := GenerateETag(filePath)
	if err == nil {
		// Handle conditional request
		if HandleConditionalRequest(w, r, etag) {
			return // 304 Not Modified
		}
	}

	http.ServeFile(w, r, filePath)
}

// HandleMaintenance serves the maintenance page
func HandleMaintenance(w http.ResponseWriter, r *http.Request) {
	// log.Printf("\t → \t → 6.3 Creating maintenance route: /maintenance")
	admin.HandleMaintenance(w, r, globalTemplates, auth.IsMaintenanceAuthenticated)
}

// HandleConfig serves the configuration page
func HandleConfig(w http.ResponseWriter, r *http.Request) {
	// log.Printf("\t → \t → 6.3 Creating config route paths: /config and /config/save")
	HandleConfigPage(w, r)
}

// HandleConfigSave handles configuration save requests
func HandleConfigSave(w http.ResponseWriter, r *http.Request) {
	// This will be handled by the inline function in routes.go
	// since it needs access to userService and projectService
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// CreateConfigSaveHandler creates a config save handler with the required services
func CreateConfigSaveHandler(userService auth.UserServicer, projectService projects.ProjectService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if userService == nil || projectService == nil {
			common.JSONError(w, "System not fully configured", http.StatusServiceUnavailable)
			return
		}
		HandleConfigSubmit(w, r, userService, projectService)
	}
}

// CreateFaviconHandler creates a favicon handler
func CreateFaviconHandler() http.HandlerFunc {
	log.Printf("\t → \t → 6.2.1 Route: /favicon.ico")
	return HandleFavicon
}

// CreateTSXCSSHandler creates a TSX CSS handler
func CreateTSXCSSHandler() http.HandlerFunc {
	log.Printf("\t → \t → 6.2.2 Route: /tsx/css/")
	return HandleTSXCSS
}

// CreateTSXJSHandler creates a TSX JS handler
func CreateTSXJSHandler() http.HandlerFunc {
	log.Printf("\t → \t → 6.2.2 Route: /tsx/js/")
	return HandleTSXJS
}

// CreateVoiceHandler creates a voice handler
func CreateVoiceHandler() http.HandlerFunc {
	log.Printf("\t → \t → 6.8 Setting /voice handler with Auth")
	return HandleVoicePage
}

// CreateAgentHandler creates an agent handler
func CreateAgentHandler() http.HandlerFunc {
	log.Printf("\t → \t → 6.9 Setting /agent handler with Auth")
	return HandleAgentPage
}

// CreateVersionHandler creates a version handler
func CreateVersionHandler() http.HandlerFunc {
	log.Printf("\t → \t → 6.10 Route: /version handler")
	return HandleVersion
}

// CreateTestHandler creates a test handler
func CreateTestHandler() http.HandlerFunc {
	log.Printf("\t → \t → 6.20 Route: /test handler")
	return HandleTestPage
}

// CreateDashboardHandler creates a dashboard handler with the required services
func CreateDashboardHandler(projectService projects.ProjectService) http.HandlerFunc {
	log.Printf("\t → \t → 6.7 Setting /dashboard handler with Auth")
	return func(w http.ResponseWriter, r *http.Request) {
		user := auth.GetUserFromContext(r.Context())
		if user != nil && user.IsAdmin {
			// Admin users get redirected to admin dashboard
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
			return
		}

		if projectService == nil {
			http.Error(w, "Project service not available", http.StatusInternalServerError)
			return
		}
		HandleDashboard(w, r, projectService)
	}
}

// HandleDashboardWithAuth handles dashboard with admin redirect logic
func HandleDashboardWithAuth(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user != nil && user.IsAdmin {
		// Admin users get redirected to admin dashboard
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	// Get projectService from context or global variable
	// For now, we'll handle this in the route definition
	http.Error(w, "Project service not available", http.StatusInternalServerError)
}

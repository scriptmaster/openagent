package admin

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/scriptmaster/openagent/auth"
	"github.com/scriptmaster/openagent/common"
	"github.com/scriptmaster/openagent/models"
	"github.com/scriptmaster/openagent/projects"
	"github.com/scriptmaster/openagent/types"
)

// HandleAdmin displays the admin dashboard
func HandleAdmin(w http.ResponseWriter, r *http.Request, templates types.TemplateEngineInterface) {
	// Get user from context (set by auth middleware)
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	data := models.PageData{
		AppName:    "OpenAgent",
		PageTitle:  "Admin Dashboard - OpenAgent",
		User:       user,
		AdminEmail: common.GetEnv("SYSADMIN_EMAIL"),
		AppVersion: common.GetEnv("APP_VERSION"),
		Stats:      &models.AdminStats{}, // Will be populated by the route handler
	}

	if err := templates.ExecuteTemplate(w, "admin.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleMaintenance displays the maintenance login page
func HandleMaintenance(w http.ResponseWriter, r *http.Request, templates types.TemplateEngineInterface, isMaintenanceAuthenticated func(r *http.Request) bool) {
	// If already authenticated, redirect to config page
	if isMaintenanceAuthenticated(r) {
		http.Redirect(w, r, "/maintenance/config", http.StatusSeeOther)
		return
	}

	// Show login page using the new struct from types.go
	data := MaintenanceLoginData{
		Error:      r.URL.Query().Get("error"),
		AdminEmail: common.GetEnv("SYSADMIN_EMAIL"),
		AppVersion: common.GetEnv("APP_VERSION"),
	}

	if err := templates.ExecuteTemplate(w, "maintenance-login.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleMaintenanceAuth processes the maintenance token submission
func HandleMaintenanceAuth(w http.ResponseWriter, r *http.Request, sessionSalt string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/maintenance?error="+url.QueryEscape("Failed to parse form"), http.StatusSeeOther)
		return
	}

	// Validate token
	submittedToken := r.FormValue("token")
	envToken := common.GetEnv("MAINTENANCE_TOKEN")

	if envToken == "" || submittedToken != envToken {
		http.Redirect(w, r, "/maintenance?error="+url.QueryEscape("Invalid maintenance token"), http.StatusSeeOther)
		return
	}

	// Set authentication cookie
	auth.SetMaintenanceCookie(w, sessionSalt)

	// Redirect to configuration page
	http.Redirect(w, r, "/maintenance/config", http.StatusSeeOther)
}

// HandleMaintenanceConfig displays the maintenance configuration page
func HandleMaintenanceConfig(w http.ResponseWriter, r *http.Request, templates types.TemplateEngineInterface, isMaintenanceAuthenticated func(r *http.Request) bool) {
	// Verify authentication (MaintenanceHandler should handle this, but double-check)
	if !isMaintenanceAuthenticated(r) {
		http.Redirect(w, r, "/maintenance", http.StatusSeeOther)
		return
	}

	// Get error/success from query parameters if present
	errorMsg := r.URL.Query().Get("error")
	successMsg := r.URL.Query().Get("success")

	// Parse current version
	versionParts := []string{"1", "0", "0", "0"} // Default
	appVersion := common.GetEnv("APP_VERSION")
	if appVersion != "" {
		parts := strings.Split(appVersion, ".")
		if len(parts) == 4 {
			versionParts = parts
		}
	}

	major, _ := strconv.Atoi(versionParts[0])
	minor, _ := strconv.Atoi(versionParts[1])
	patch, _ := strconv.Atoi(versionParts[2])
	build, _ := strconv.Atoi(versionParts[3])

	// Get current migration start number - always retrieve the latest from environment
	migrationStart := common.GetEnv("MIGRATION_START")
	if migrationStart == "" {
		migrationStart = "000" // Ensure consistent formatting with leading zeros
	}

	// Use the new struct from types.go
	data := MaintenanceConfigData{
		DBHost:         common.GetEnv("DB_HOST"),
		DBPort:         common.GetEnv("DB_PORT"),
		DBUser:         common.GetEnv("DB_USER"),
		DBPassword:     common.GetEnv("DB_PASSWORD"), // Populate even if not always shown
		DBName:         common.GetEnv("DB_NAME"),
		AdminEmail:     common.GetEnv("SYSADMIN_EMAIL"),
		Error:          errorMsg,
		Success:        successMsg,
		VersionMajor:   major,
		VersionMinor:   minor,
		VersionPatch:   patch,
		VersionBuild:   build,
		MigrationStart: migrationStart,
		SMTPHost:       common.GetEnv("SMTP_HOST"),
		SMTPPort:       common.GetEnv("SMTP_PORT"),
		SMTPUser:       common.GetEnv("SMTP_USER"),
		SMTPPassword:   common.GetEnv("SMTP_PASSWORD"), // Populate even if not always shown
		SMTPFrom:       common.GetEnv("SMTP_FROM"),
	}

	if err := templates.ExecuteTemplate(w, "maintenance.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleMaintenanceConfigure updates database configuration and attempts to reconnect
func HandleMaintenanceConfigure(w http.ResponseWriter, r *http.Request, templates types.TemplateEngineInterface, isMaintenanceAuthenticated func(r *http.Request) bool, updateDatabaseConfig func(host, port, user, password, dbname string) error) {
	// Verify authentication (should be caught by middleware, but double-check)
	if !isMaintenanceAuthenticated(r) {
		http.Redirect(w, r, "/maintenance", http.StatusSeeOther)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		handleMaintenanceError(w, r, templates, "Failed to parse form data: "+err.Error())
		return
	}

	// Get database form values
	host := r.FormValue("db_host")
	port := r.FormValue("db_port")
	user := r.FormValue("db_user")
	password := r.FormValue("db_password")
	dbname := r.FormValue("db_name")
	resetMigrations := r.FormValue("reset_migrations") == "1"
	migrationStart := r.FormValue("migration_start")

	// Get SMTP values
	smtpHost := r.FormValue("smtp_host")
	smtpPort := r.FormValue("smtp_port")
	smtpUser := r.FormValue("smtp_user")
	smtpPassword := r.FormValue("smtp_password")
	smtpFrom := r.FormValue("smtp_from")

	// Get version components
	majorStr := r.FormValue("version_major")
	minorStr := r.FormValue("version_minor")
	patchStr := r.FormValue("version_patch")

	// Validate database fields
	if host == "" || port == "" || user == "" || dbname == "" {
		handleMaintenanceError(w, r, templates, "All database fields except password are required")
		return
	}

	// Validate version inputs
	major, err := strconv.Atoi(majorStr)
	if err != nil || major < 0 {
		handleMaintenanceError(w, r, templates, "Invalid major version number")
		return
	}

	minor, err := strconv.Atoi(minorStr)
	if err != nil || minor < 0 {
		handleMaintenanceError(w, r, templates, "Invalid minor version number")
		return
	}

	patch, err := strconv.Atoi(patchStr)
	if err != nil || patch < 0 {
		handleMaintenanceError(w, r, templates, "Invalid patch version number")
		return
	}

	// Handle migration tracking - now using database-based tracking
	if resetMigrations {
		// Reset migration tracking to apply all migrations
		log.Println("Migration tracking reset requested")
		// Get database connection - we need to get this from the initDB function
		// For now, we'll skip the reset if we can't get the DB connection
		log.Println("Migration reset requested but database connection not available in this context")
		// TODO: Pass database connection to this handler or get it from a global
		// For now, we'll just log the request
	} else if migrationStart != "" {
		// Migration start number is no longer used with database-based tracking
		// This is kept for backward compatibility but doesn't do anything
		log.Printf("Migration start number '%s' provided but ignored - using database-based tracking", migrationStart)
	}

	// Construct the updated configuration
	configUpdates := map[string]string{
		"DB_HOST":       host,
		"DB_PORT":       port,
		"DB_USER":       user,
		"DB_PASSWORD":   password,
		"DB_NAME":       dbname,
		"SMTP_HOST":     smtpHost,
		"SMTP_PORT":     smtpPort,
		"SMTP_USER":     smtpUser,
		"SMTP_PASSWORD": smtpPassword,
		"SMTP_FROM":     smtpFrom,
		// Update version in environment
		"APP_VERSION": fmt.Sprintf("%d.%d.%d.%d", major, minor, patch, GetBuildNumber()+1), // Increment build number on update
	}

	// Update environment variables and .env file
	if err := UpdateEnvFile(configUpdates); err != nil {
		handleMaintenanceError(w, r, templates, "Failed to update configuration file: "+err.Error())
		return
	}

	// Log the update attempt
	log.Printf("Configuration updated: DB=%s:%s/%s User=%s SMTP=%s:%s", host, port, dbname, user, smtpHost, smtpPort)

	// Trigger restart logic here (e.g., send signal, use supervisor, etc.)
	// For simplicity, we'll just log and expect manual/external restart for now
	log.Println("Configuration saved. Server restart required to apply changes.")

	// Redirect back to config page with success message
	http.Redirect(w, r, "/maintenance/config?success="+url.QueryEscape("Configuration saved. Restart server to apply changes."), http.StatusSeeOther)
}

// handleMaintenanceError redirects back to the config page with an error message
func handleMaintenanceError(w http.ResponseWriter, r *http.Request, templates types.TemplateEngineInterface, errorMsg string) {
	log.Printf("Maintenance configuration error: %s", errorMsg)
	// Parse current version
	versionParts := []string{"1", "0", "0", "0"} // Default
	appVersion := common.GetEnv("APP_VERSION")
	if appVersion != "" {
		parts := strings.Split(appVersion, ".")
		if len(parts) == 4 {
			versionParts = parts
		}
	}

	major, _ := strconv.Atoi(versionParts[0])
	minor, _ := strconv.Atoi(versionParts[1])
	patch, _ := strconv.Atoi(versionParts[2])
	build, _ := strconv.Atoi(versionParts[3])

	// Get current migration start number
	migrationStart := common.GetEnv("MIGRATION_START")
	if migrationStart == "" {
		migrationStart = "000"
	}

	// Populate data with current/submitted values to redisplay the form
	data := MaintenanceConfigData{
		DBHost:         r.FormValue("db_host"), // Use submitted value
		DBPort:         r.FormValue("db_port"),
		DBUser:         r.FormValue("db_user"),
		DBPassword:     r.FormValue("db_password"),
		DBName:         r.FormValue("db_name"),
		AdminEmail:     common.GetEnv("SYSADMIN_EMAIL"),
		Error:          errorMsg,
		VersionMajor:   major,
		VersionMinor:   minor,
		VersionPatch:   patch,
		VersionBuild:   build,
		MigrationStart: migrationStart,
		SMTPHost:       r.FormValue("smtp_host"), // Use submitted value
		SMTPPort:       r.FormValue("smtp_port"),
		SMTPUser:       r.FormValue("smtp_user"),
		SMTPPassword:   r.FormValue("smtp_password"),
		SMTPFrom:       r.FormValue("smtp_from"),
	}

	if err := templates.ExecuteTemplate(w, "maintenance.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleInitializeSchema attempts to initialize the database schema.
func HandleInitializeSchema(w http.ResponseWriter, r *http.Request, isMaintenanceAuthenticated func(r *http.Request) bool, initDB func() (*sql.DB, error)) {
	if !isMaintenanceAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Println("Attempting to initialize database schema...")
	db, err := initDB()
	if err != nil {
		msg := fmt.Sprintf("Failed to initialize database schema: %v", err)
		log.Println(msg)
		http.Redirect(w, r, "/maintenance/config?error="+url.QueryEscape(msg), http.StatusSeeOther)
		return
	}
	db.Close()

	// After successful initialization, reset migration tracking in database
	if err := common.ResetMigrationTracking(db); err != nil {
		log.Printf("Warning: Failed to reset migration tracking after schema init: %v", err)
	}

	log.Println("Database schema initialized successfully.")
	http.Redirect(w, r, "/maintenance/config?success="+url.QueryEscape("Database schema initialized successfully."), http.StatusSeeOther)
}

// Use shared types from common package

// HandleAdminCLI displays the admin CLI interface
func HandleAdminCLI(w http.ResponseWriter, r *http.Request, templates types.TemplateEngineInterface, getDB func() *sql.DB) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	data := models.PageData{
		AppName:    "OpenAgent",
		PageTitle:  "Admin CLI - OpenAgent",
		User:       user,
		AdminEmail: common.GetEnv("SYSADMIN_EMAIL"),
		AppVersion: common.GetEnv("APP_VERSION"),
	}

	if err := templates.ExecuteTemplate(w, "admin-cli.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleCLIQueriesAPI returns the list of available queries grouped by parameter count
func HandleCLIQueriesAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Use shared function to get available queries
	queryGroups, err := common.GetAvailableQueries()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(queryGroups)
}

// HandleCLIExecuteAPI executes a query and returns the results
func HandleCLIExecuteAPI(w http.ResponseWriter, r *http.Request, getDB func() *sql.DB) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req common.ExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Use shared function to execute query
	response := common.ExecuteQuery(req, getDB())

	json.NewEncoder(w).Encode(response)
}

// HandleAdminConnections handles the admin connections page
func HandleAdminConnections(w http.ResponseWriter, r *http.Request, templates types.TemplateEngineInterface, getDB func() *sql.DB) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	data := models.PageData{
		AppName:    "OpenAgent",
		PageTitle:  "Database Connections - Admin",
		User:       user,
		AdminEmail: common.GetEnv("SYSADMIN_EMAIL"),
		AppVersion: common.GetEnv("APP_VERSION"),
	}

	if err := templates.ExecuteTemplate(w, "admin-connections.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleAdminTables handles the admin tables page
func HandleAdminTables(w http.ResponseWriter, r *http.Request, templates types.TemplateEngineInterface, getDB func() *sql.DB) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	data := models.PageData{
		AppName:    "OpenAgent",
		PageTitle:  "Database Tables - Admin",
		User:       user,
		AdminEmail: common.GetEnv("SYSADMIN_EMAIL"),
		AppVersion: common.GetEnv("APP_VERSION"),
	}

	if err := templates.ExecuteTemplate(w, "admin-tables.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleAdminSettings handles the admin settings page
func HandleAdminSettings(w http.ResponseWriter, r *http.Request, templates types.TemplateEngineInterface, getDB func() *sql.DB) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	data := models.PageData{
		AppName:    "OpenAgent",
		PageTitle:  "Admin Settings",
		User:       user,
		AdminEmail: common.GetEnv("SYSADMIN_EMAIL"),
		AppVersion: common.GetEnv("APP_VERSION"),
	}

	if err := templates.ExecuteTemplate(w, "admin-settings.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// // HandleUsers handles the users management page
// func HandleUsers(w http.ResponseWriter, r *http.Request, templates *template.Template, getDB func() *sql.DB) {
// 	user := auth.GetUserFromContext(r.Context())
// 	if user == nil {
// 		http.Error(w, "User not found in context", http.StatusInternalServerError)
// 		return
// 	}

// 	data := models.PageData{
// 		AppName:    "OpenAgent",
// 		PageTitle:  "User Management",
// 		User:       user,
// 		AdminEmail: common.GetEnv("SYSADMIN_EMAIL"),
// 		AppVersion: common.GetEnv("APP_VERSION"),
// 	}

// 	if err := templates.ExecuteTemplate(w, "users.html", data); err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 	}
// }

// HandleConnections handles the connections page for regular users
func HandleConnections(w http.ResponseWriter, r *http.Request, templates types.TemplateEngineInterface, getDB func() *sql.DB) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	data := models.PageData{
		AppName:    "OpenAgent",
		PageTitle:  "Database Connections",
		User:       user,
		AdminEmail: common.GetEnv("SYSADMIN_EMAIL"),
		AppVersion: common.GetEnv("APP_VERSION"),
	}

	if err := templates.ExecuteTemplate(w, "connections.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleTables handles the tables page for regular users
func HandleTables(w http.ResponseWriter, r *http.Request, templates types.TemplateEngineInterface, getDB func() *sql.DB) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	data := models.PageData{
		AppName:    "OpenAgent",
		PageTitle:  "Database Tables",
		User:       user,
		AdminEmail: common.GetEnv("SYSADMIN_EMAIL"),
		AppVersion: common.GetEnv("APP_VERSION"),
	}

	if err := templates.ExecuteTemplate(w, "tables.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleProfile handles the user profile page
func HandleProfile(w http.ResponseWriter, r *http.Request, templates types.TemplateEngineInterface) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	data := models.PageData{
		AppName:    "OpenAgent",
		PageTitle:  "User Profile",
		User:       user,
		AdminEmail: common.GetEnv("SYSADMIN_EMAIL"),
		AppVersion: common.GetEnv("APP_VERSION"),
	}

	if err := templates.ExecuteTemplate(w, "profile.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Handle404 serves a custom 404 page with context information
func Handle404(w http.ResponseWriter, r *http.Request, templates types.TemplateEngineInterface) {
	user := auth.GetUserFromContext(r.Context())
	project := projects.GetProjectFromContext(r.Context())

	// Get host information
	host := r.Host
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		if forwardedHost := r.Header.Get("X-Forwarded-Host"); forwardedHost != "" {
			host = forwardedHost
		}
	}
	host = strings.Split(host, ":")[0]

	// Prepare 404 data
	data := struct {
		AppName    string
		PageTitle  string
		User       *auth.User
		Project    *projects.Project
		Host       string
		AdminEmail string
		AppVersion string
	}{
		AppName:    "OpenAgent",
		PageTitle:  "404 - Page Not Found",
		User:       user,
		Project:    project,
		Host:       host,
		AdminEmail: common.GetEnv("SYSADMIN_EMAIL"),
		AppVersion: common.GetEnv("APP_VERSION"),
	}

	w.WriteHeader(http.StatusNotFound)
	if err := templates.ExecuteTemplate(w, "error_404.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

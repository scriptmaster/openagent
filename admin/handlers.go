package admin

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/scriptmaster/openagent/models"
)

// HandleAdmin displays the admin dashboard
func HandleAdmin(w http.ResponseWriter, r *http.Request, templates *template.Template) {
	data := models.PageData{
		AppName:    "OpenAgent",
		PageTitle:  "Admin Dashboard - OpenAgent",
		AdminEmail: os.Getenv("SYSADMIN_EMAIL"),
		AppVersion: os.Getenv("APP_VERSION"),
	}

	if err := templates.ExecuteTemplate(w, "admin.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleMaintenance displays the maintenance login page
func HandleMaintenance(w http.ResponseWriter, r *http.Request, templates *template.Template, isMaintenanceAuthenticated func(r *http.Request) bool) {
	// If already authenticated, redirect to config page
	if isMaintenanceAuthenticated(r) {
		http.Redirect(w, r, "/maintenance/config", http.StatusSeeOther)
		return
	}

	// Show login page
	data := struct {
		Error      string
		AdminEmail string
		AppVersion string
	}{
		Error:      r.URL.Query().Get("error"),
		AdminEmail: os.Getenv("SYSADMIN_EMAIL"),
		AppVersion: os.Getenv("APP_VERSION"),
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
	envToken := os.Getenv("MAINTENANCE_TOKEN")

	if envToken == "" || submittedToken != envToken {
		http.Redirect(w, r, "/maintenance?error="+url.QueryEscape("Invalid maintenance token"), http.StatusSeeOther)
		return
	}

	// Set authentication cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "maintenance_auth",
		Value:    "authenticated_" + sessionSalt[:8], // Add partial version salt to invalidate on restart
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		MaxAge:   int(1 * time.Hour.Seconds()),
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect to configuration page
	http.Redirect(w, r, "/maintenance/config", http.StatusSeeOther)
}

// HandleMaintenanceConfig displays the maintenance configuration page
func HandleMaintenanceConfig(w http.ResponseWriter, r *http.Request, templates *template.Template, isMaintenanceAuthenticated func(r *http.Request) bool) {
	// Verify authentication (MaintenanceHandler should handle this, but double-check)
	if !isMaintenanceAuthenticated(r) {
		http.Redirect(w, r, "/maintenance", http.StatusSeeOther)
		return
	}

	// Get error from query parameters if present
	errorMsg := r.URL.Query().Get("error")

	// Parse current version
	versionParts := []string{"1", "0", "0", "0"} // Default
	appVersion := os.Getenv("APP_VERSION")
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
	migrationStart := os.Getenv("MIGRATION_START")
	if migrationStart == "" {
		migrationStart = "000" // Ensure consistent formatting with leading zeros
	}

	data := struct {
		DBHost         string
		DBPort         string
		DBUser         string
		DBPassword     string
		DBName         string
		Error          string
		Success        string
		AdminEmail     string
		VersionMajor   int
		VersionMinor   int
		VersionPatch   int
		VersionBuild   int
		MigrationStart string
		SMTPHost       string
		SMTPPort       string
		SMTPUser       string
		SMTPPassword   string
		SMTPFrom       string
	}{
		DBHost:         os.Getenv("DB_HOST"),
		DBPort:         os.Getenv("DB_PORT"),
		DBUser:         os.Getenv("DB_USER"),
		DBPassword:     os.Getenv("DB_PASSWORD"),
		DBName:         os.Getenv("DB_NAME"),
		AdminEmail:     os.Getenv("SYSADMIN_EMAIL"),
		Error:          errorMsg,
		VersionMajor:   major,
		VersionMinor:   minor,
		VersionPatch:   patch,
		VersionBuild:   build,
		MigrationStart: migrationStart,
		SMTPHost:       os.Getenv("SMTP_HOST"),
		SMTPPort:       os.Getenv("SMTP_PORT"),
		SMTPUser:       os.Getenv("SMTP_USER"),
		SMTPPassword:   os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:       os.Getenv("SMTP_FROM"),
	}

	if err := templates.ExecuteTemplate(w, "maintenance.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleMaintenanceConfigure updates database configuration and attempts to reconnect
func HandleMaintenanceConfigure(w http.ResponseWriter, r *http.Request, templates *template.Template, isMaintenanceAuthenticated func(r *http.Request) bool, updateDatabaseConfig func(host, port, user, password, dbname string) error, updateMigrationStart func(migrationNum int) error) {
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
		handleMaintenanceError(w, r, templates, isMaintenanceAuthenticated, "Failed to parse form data: "+err.Error())
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
		handleMaintenanceError(w, r, templates, isMaintenanceAuthenticated, "All database fields except password are required")
		return
	}

	// Validate version inputs
	major, err := strconv.Atoi(majorStr)
	if err != nil || major < 0 {
		handleMaintenanceError(w, r, templates, isMaintenanceAuthenticated, "Invalid major version number")
		return
	}

	minor, err := strconv.Atoi(minorStr)
	if err != nil || minor < 0 {
		handleMaintenanceError(w, r, templates, isMaintenanceAuthenticated, "Invalid minor version number")
		return
	}

	patch, err := strconv.Atoi(patchStr)
	if err != nil || patch < 0 {
		handleMaintenanceError(w, r, templates, isMaintenanceAuthenticated, "Invalid patch version number")
		return
	}

	// Handle migration tracking
	if resetMigrations {
		// Reset migration tracking to apply all migrations
		log.Println("Migration tracking reset requested")
		if err := updateMigrationStart(0); err != nil {
			log.Printf("Warning: Failed to reset migration tracking: %v", err)
		} else {
			log.Println("Migration tracking reset to 0")
		}
	} else if migrationStart != "" {
		// Parse migration start number, handling different formats (4, 04, 004)
		migNum, err := strconv.Atoi(migrationStart)
		if err != nil {
			handleMaintenanceError(w, r, templates, isMaintenanceAuthenticated, "Invalid migration number: "+err.Error())
			return
		}

		// Update migration tracking to the specified number
		log.Printf("Setting migration tracking to %d", migNum)
		if err := updateMigrationStart(migNum); err != nil {
			log.Printf("Warning: Failed to update migration tracking: %v", err)
		}
	}

	// Update database configuration
	if err := updateDatabaseConfig(host, port, user, password, dbname); err != nil {
		handleMaintenanceError(w, r, templates, isMaintenanceAuthenticated, "Failed to update database configuration: "+err.Error())
		return
	}

	// Update environment variables for database
	os.Setenv("DB_HOST", host)
	os.Setenv("DB_PORT", port)
	os.Setenv("DB_USER", user)
	os.Setenv("DB_PASSWORD", password)
	os.Setenv("DB_NAME", dbname)

	// Update SMTP environment variables if provided
	if smtpHost != "" {
		os.Setenv("SMTP_HOST", smtpHost)
	}
	if smtpPort != "" {
		os.Setenv("SMTP_PORT", smtpPort)
	}
	if smtpUser != "" {
		os.Setenv("SMTP_USER", smtpUser)
	}
	if smtpPassword != "" {
		os.Setenv("SMTP_PASSWORD", smtpPassword)
	}
	if smtpFrom != "" {
		os.Setenv("SMTP_FROM", smtpFrom)
	}

	// Update .env file with SMTP settings
	updateEnvFile := func(lines []string) []string {
		smtpVars := map[string]string{
			"SMTP_HOST":     smtpHost,
			"SMTP_PORT":     smtpPort,
			"SMTP_USER":     smtpUser,
			"SMTP_PASSWORD": smtpPassword,
			"SMTP_FROM":     smtpFrom,
		}

		// Track which SMTP variables we've updated
		updatedVars := map[string]bool{}

		// Update existing lines
		for i, line := range lines {
			if line == "" {
				continue
			}

			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			if val, exists := smtpVars[key]; exists && val != "" {
				lines[i] = key + "=" + val
				updatedVars[key] = true
			}
		}

		// Add missing SMTP vars that have values
		for key, val := range smtpVars {
			if !updatedVars[key] && val != "" {
				lines = append(lines, key+"="+val)
			}
		}

		return lines
	}

	// Handle version update
	// Parse current version to get build number
	buildNumber := 0
	appVersion := os.Getenv("APP_VERSION")
	if appVersion != "" {
		parts := strings.Split(appVersion, ".")
		if len(parts) == 4 {
			buildNumber, _ = strconv.Atoi(parts[3])
		}
	}

	// Construct new version
	newVersion := fmt.Sprintf("%d.%d.%d.%d", major, minor, patch, buildNumber)

	// Update .env file with version
	envPath := ".env"
	content, err := os.ReadFile(envPath)
	if err != nil && !os.IsNotExist(err) {
		handleMaintenanceError(w, r, templates, isMaintenanceAuthenticated, "Failed to read .env file: "+err.Error())
		return
	}

	// Parse .env content line by line
	lines := strings.Split(string(content), "\n")
	versionLineFound := false
	for i, line := range lines {
		if strings.HasPrefix(line, "APP_VERSION=") {
			lines[i] = "APP_VERSION=" + newVersion
			versionLineFound = true
			break
		}
	}

	// If APP_VERSION line not found, add it
	if !versionLineFound {
		lines = append(lines, "APP_VERSION="+newVersion)
	}

	// Update SMTP settings
	lines = updateEnvFile(lines)

	// Write back to .env
	err = os.WriteFile(envPath, []byte(strings.Join(lines, "\n")), 0644)
	if err != nil {
		handleMaintenanceError(w, r, templates, isMaintenanceAuthenticated, "Failed to update .env file: "+err.Error())
		return
	}

	// Update global variables
	os.Setenv("APP_VERSION", newVersion)

	// Build success message
	successMsg := "Configuration updated successfully!"
	if resetMigrations {
		successMsg += " Migration tracking reset to apply all migrations."
	} else if migrationStart != "" {
		successMsg += fmt.Sprintf(" Migration tracking set to %s.", migrationStart)
	}
	successMsg += " Server will restart in 5 seconds..."

	log.Printf("Application configuration updated. Version: %s, Database: %s@%s:%s/%s",
		newVersion, user, host, port, dbname)

	// Parse current version for template
	versionParts := strings.Split(newVersion, ".")
	vMajor, _ := strconv.Atoi(versionParts[0])
	vMinor, _ := strconv.Atoi(versionParts[1])
	vPatch, _ := strconv.Atoi(versionParts[2])
	vBuild, _ := strconv.Atoi(versionParts[3])

	// Get updated migration start - always retrieve the latest from environment
	updatedMigrationStart := os.Getenv("MIGRATION_START")
	if updatedMigrationStart == "" {
		updatedMigrationStart = "000" // Ensure consistent formatting with leading zeros
	}

	// Show success message and redirect to home after a delay
	data := struct {
		DBHost         string
		DBPort         string
		DBUser         string
		DBPassword     string
		DBName         string
		Error          string
		Success        string
		AdminEmail     string
		VersionMajor   int
		VersionMinor   int
		VersionPatch   int
		VersionBuild   int
		MigrationStart string
		SMTPHost       string
		SMTPPort       string
		SMTPUser       string
		SMTPPassword   string
		SMTPFrom       string
	}{
		DBHost:         host,
		DBPort:         port,
		DBUser:         user,
		DBPassword:     password,
		DBName:         dbname,
		Success:        successMsg,
		AdminEmail:     os.Getenv("SYSADMIN_EMAIL"),
		VersionMajor:   vMajor,
		VersionMinor:   vMinor,
		VersionPatch:   vPatch,
		VersionBuild:   vBuild,
		MigrationStart: updatedMigrationStart,
		SMTPHost:       smtpHost,
		SMTPPort:       smtpPort,
		SMTPUser:       smtpUser,
		SMTPPassword:   smtpPassword,
		SMTPFrom:       smtpFrom,
	}

	w.Header().Set("Refresh", "5;url=/")
	if err := templates.ExecuteTemplate(w, "maintenance.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleMaintenanceError displays error message on maintenance page
func handleMaintenanceError(w http.ResponseWriter, r *http.Request, templates *template.Template, isMaintenanceAuthenticated func(r *http.Request) bool, errorMsg string) {
	// Verify authentication (should be caught by middleware, but double-check)
	if !isMaintenanceAuthenticated(r) {
		http.Redirect(w, r, "/maintenance", http.StatusSeeOther)
		return
	}

	// Parse current version
	versionParts := []string{"1", "0", "0", "0"} // Default
	appVersion := os.Getenv("APP_VERSION")
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

	// Get current migration start or use form value if present
	migrationStart := r.FormValue("migration_start")
	if migrationStart == "" {
		migrationStart = os.Getenv("MIGRATION_START")
		if migrationStart == "" {
			migrationStart = "000" // Ensure consistent formatting with leading zeros
		}
	}

	data := struct {
		DBHost         string
		DBPort         string
		DBUser         string
		DBPassword     string
		DBName         string
		Error          string
		Success        string
		AdminEmail     string
		VersionMajor   int
		VersionMinor   int
		VersionPatch   int
		VersionBuild   int
		MigrationStart string
		SMTPHost       string
		SMTPPort       string
		SMTPUser       string
		SMTPPassword   string
		SMTPFrom       string
	}{
		DBHost:         r.FormValue("db_host"),
		DBPort:         r.FormValue("db_port"),
		DBUser:         r.FormValue("db_user"),
		DBPassword:     r.FormValue("db_password"),
		DBName:         r.FormValue("db_name"),
		Error:          errorMsg,
		AdminEmail:     os.Getenv("SYSADMIN_EMAIL"),
		VersionMajor:   major,
		VersionMinor:   minor,
		VersionPatch:   patch,
		VersionBuild:   build,
		MigrationStart: migrationStart,
		SMTPHost:       r.FormValue("smtp_host"),
		SMTPPort:       r.FormValue("smtp_port"),
		SMTPUser:       r.FormValue("smtp_user"),
		SMTPPassword:   r.FormValue("smtp_password"),
		SMTPFrom:       r.FormValue("smtp_from"),
	}

	if err := templates.ExecuteTemplate(w, "maintenance.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleInitializeSchema manually executes the database schema initialization
func HandleInitializeSchema(w http.ResponseWriter, r *http.Request, isMaintenanceAuthenticated func(r *http.Request) bool, updateMigrationStart func(migrationNum int) error, initDB func() (*sql.DB, error)) {
	// Verify authentication
	if !isMaintenanceAuthenticated(r) {
		http.Redirect(w, r, "/maintenance", http.StatusSeeOther)
		return
	}

	// Reset migration tracking to apply all migrations
	if err := updateMigrationStart(0); err != nil {
		http.Redirect(w, r, "/maintenance/config?error="+url.QueryEscape("Failed to reset migration tracking: "+err.Error()), http.StatusSeeOther)
		return
	}

	// Get database connection
	db, err := initDB()
	if err != nil {
		// Return to maintenance config with error
		http.Redirect(w, r, "/maintenance/config?error="+url.QueryEscape("Failed to connect to database: "+err.Error()), http.StatusSeeOther)
		return
	}
	defer db.Close()

	// Return to maintenance config with success message
	http.Redirect(w, r, "/maintenance/config?success=Database+schema+initialized+successfully", http.StatusSeeOther)
}

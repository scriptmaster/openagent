package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Global variables
var (
	templates *template.Template
	// Session key for maintenance mode authentication
	maintenanceSessionKey = "maintenance_auth"
	// Current application version
	appVersion string
	// Session salt derived from version
	sessionSalt string
)

// Template helper functions
var templateFuncs = template.FuncMap{
	"formatTime": func(t interface{}) string {
		if t == nil {
			return "Never"
		}

		switch v := t.(type) {
		case time.Time:
			return v.Format("Jan 02, 2006 15:04:05")
		case *time.Time:
			if v == nil {
				return "Never"
			}
			return v.Format("Jan 02, 2006 15:04:05")
		case sql.NullTime:
			if !v.Valid {
				return "Never"
			}
			return v.Time.Format("Jan 02, 2006 15:04:05")
		default:
			return "Invalid time format"
		}
	},
	"formatDate": func(t interface{}) string {
		if t == nil {
			return "Never"
		}

		switch v := t.(type) {
		case time.Time:
			return v.Format("Jan 02, 2006")
		case *time.Time:
			if v == nil {
				return "Never"
			}
			return v.Format("Jan 02, 2006")
		case sql.NullTime:
			if !v.Valid {
				return "Never"
			}
			return v.Time.Format("Jan 02, 2006")
		default:
			return "Invalid date format"
		}
	},
	"safeHTML": func(s string) template.HTML {
		return template.HTML(s)
	},
}

// Response object for JSON APIs
type JSONResponse struct {
	Success  bool        `json:"success"`
	Message  string      `json:"message,omitempty"`
	Data     interface{} `json:"data,omitempty"`
	Redirect string      `json:"redirect,omitempty"`
}

// Page data for templates
type PageData struct {
	AppName    string
	PageTitle  string
	User       User
	Error      string
	Projects   []interface{}
	Project    interface{}
	AdminEmail string
	AppVersion string
}

// Maintenance session for secure access
func isMaintenanceAuthenticated(r *http.Request) bool {
	// Check for maintenance cookie
	cookie, err := r.Cookie(maintenanceSessionKey)
	if err != nil {
		return false
	}

	// Expected value with current version salt
	expected := "authenticated_" + sessionSalt[:8]

	// Validate cookie value matches current version
	return cookie.Value == expected
}

// MaintenanceHandler handles requests when in maintenance mode
func MaintenanceHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip maintenance mode check for static files
		if strings.HasPrefix(r.URL.Path, "/static/") {
			next.ServeHTTP(w, r)
			return
		}

		// Always allow access to maintenance endpoints
		if r.URL.Path == "/maintenance" || r.URL.Path == "/maintenance/auth" {
			next.ServeHTTP(w, r)
			return
		}

		// Special handling for login page - in maintenance mode redirect to /maintenance
		if r.URL.Path == "/login" && IsMaintenanceMode() {
			http.Redirect(w, r, "/maintenance", http.StatusSeeOther)
			return
		}

		// Check for maintenance configuration access
		if strings.HasPrefix(r.URL.Path, "/maintenance/") {
			if !isMaintenanceAuthenticated(r) {
				// Not authenticated, redirect to maintenance login
				http.Redirect(w, r, "/maintenance", http.StatusSeeOther)
				return
			}
			// Authenticated, allow access
			next.ServeHTTP(w, r)
			return
		}

		// If in maintenance mode, check authentication
		if IsMaintenanceMode() {
			if !isMaintenanceAuthenticated(r) {
				// Not authenticated, redirect to maintenance login
				http.Redirect(w, r, "/maintenance", http.StatusSeeOther)
				return
			}

			// Already authenticated for maintenance, redirect to configuration
			if r.URL.Path == "/maintenance" {
				http.Redirect(w, r, "/maintenance/config", http.StatusSeeOther)
				return
			}
		}

		// Continue normal processing
		next.ServeHTTP(w, r)
	})
}

// StartServer initializes and starts the web server
func StartServer() {
	log.Println("--- Server Starting ---")

	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: .env file not found, using environment variables.")
	}

	// Increment build version on startup - this will also invalidate all previous sessions
	if err := incrementBuildVersion(); err != nil {
		log.Printf("Warning: failed to increment build version: %v", err)
	}

	// Initialize database
	db, err := InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create user service
	userService := NewUserService(db)

	// Initialize template engine
	templates = template.Must(template.New("").Funcs(templateFuncs).ParseGlob("tpl/*.html"))

	// Initialize agent templates
	InitAgentTemplates(templates)

	// Setup static file server
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Auth handlers
	http.HandleFunc("/auth/request-otp", func(w http.ResponseWriter, r *http.Request) {
		handleAuthRequestOTP(w, r, userService)
	})
	http.HandleFunc("/auth/verify-otp", func(w http.ResponseWriter, r *http.Request) {
		handleAuthVerifyOTP(w, r, userService)
	})

	// Page handlers
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			AppName:    "OpenAgent",
			PageTitle:  "Login - OpenAgent",
			AdminEmail: os.Getenv("SYSADMIN_EMAIL"),
			AppVersion: appVersion,
		}
		if err := templates.ExecuteTemplate(w, "login.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// Authenticated routes (use middleware)
	authMux := http.NewServeMux()
	authMux.HandleFunc("/", handleIndex)
	authMux.HandleFunc("/projects", handleProjects)
	authMux.HandleFunc("/admin", handleAdmin)

	// Apply auth middleware
	http.Handle("/", MaintenanceHandler(AuthMiddleware(authMux)))

	// Agent API endpoints from agent.go
	http.HandleFunc("/agent.html", handleAgent)
	http.HandleFunc("/start", handleStart)
	http.HandleFunc("/next", handleNextStep)
	http.HandleFunc("/status", handleStatus)

	// Maintenance handlers
	http.HandleFunc("/maintenance", handleMaintenance)
	http.HandleFunc("/maintenance/auth", handleMaintenanceAuth)
	http.HandleFunc("/maintenance/config", handleMaintenanceConfig)
	http.HandleFunc("/maintenance/configure", handleMaintenanceConfigure)
	http.HandleFunc("/maintenance/initialize-schema", handleInitializeSchema)

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8800"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// Handle OTP request
func handleAuthRequestOTP(w http.ResponseWriter, r *http.Request, userService *UserService) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON request
	var req struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		sendJSONResponse(w, false, "Invalid request", nil, "")
		return
	}

	// Validate email
	if req.Email == "" {
		sendJSONResponse(w, false, "Email is required", nil, "")
		return
	}

	// Check if user exists, create if not
	_, err := userService.GetUserByEmail(req.Email)
	if err != nil {
		// User doesn't exist, create a new one
		_, err = userService.CreateUser(req.Email)
		if err != nil {
			log.Printf("Failed to create user: %v", err)
			sendJSONResponse(w, false, "Failed to create user account", nil, "")
			return
		}
	}

	// Send OTP
	err = SendOTP(req.Email)
	if err != nil {
		log.Printf("Failed to send OTP: %v", err)
		sendJSONResponse(w, false, "Failed to send OTP", nil, "")
		return
	}

	sendJSONResponse(w, true, "OTP sent successfully", nil, "")
}

// Handle OTP verification
func handleAuthVerifyOTP(w http.ResponseWriter, r *http.Request, userService *UserService) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON request
	var req struct {
		Email string `json:"email"`
		OTP   string `json:"otp"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		sendJSONResponse(w, false, "Invalid request", nil, "")
		return
	}

	// Validate inputs
	if req.Email == "" || req.OTP == "" {
		sendJSONResponse(w, false, "Email and OTP are required", nil, "")
		return
	}

	// Verify OTP
	valid, err := VerifyOTP(req.Email, req.OTP)
	if err != nil || !valid {
		errorMsg := "Invalid OTP"
		if err != nil {
			errorMsg = err.Error()
		}
		sendJSONResponse(w, false, errorMsg, nil, "")
		return
	}

	// Get or create user
	user, err := userService.GetUserByEmail(req.Email)
	if err != nil {
		sendJSONResponse(w, false, "User not found", nil, "")
		return
	}

	// Create session
	token, err := CreateSession(user)
	if err != nil {
		sendJSONResponse(w, false, "Failed to create session", nil, "")
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		MaxAge:   int(24 * time.Hour.Seconds()),
		SameSite: http.SameSiteLaxMode,
	})

	// Return successful response with redirect
	redirect := "/"
	if user.IsAdmin {
		redirect = "/admin"
	}

	sendJSONResponse(w, true, "Authentication successful", nil, redirect)
}

// Helper function to send JSON responses
func sendJSONResponse(w http.ResponseWriter, success bool, message string, data interface{}, redirect string) {
	response := JSONResponse{
		Success:  success,
		Message:  message,
		Data:     data,
		Redirect: redirect,
	}

	w.Header().Set("Content-Type", "application/json")
	if !success {
		w.WriteHeader(http.StatusBadRequest)
	}

	json.NewEncoder(w).Encode(response)
}

// Placeholder handlers for authenticated routes
func handleIndex(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		AppName:    "OpenAgent",
		PageTitle:  "Dashboard - OpenAgent",
		AdminEmail: os.Getenv("SYSADMIN_EMAIL"),
		AppVersion: appVersion,
		Project:    nil, // Include Project field with nil value
	}

	if err := templates.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleProjects(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		AppName:    "OpenAgent",
		PageTitle:  "Projects - OpenAgent",
		AdminEmail: os.Getenv("SYSADMIN_EMAIL"),
		AppVersion: appVersion,
	}

	if err := templates.ExecuteTemplate(w, "projects.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleAdmin(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		AppName:    "OpenAgent",
		PageTitle:  "Admin Dashboard - OpenAgent",
		AdminEmail: os.Getenv("SYSADMIN_EMAIL"),
		AppVersion: appVersion,
	}

	if err := templates.ExecuteTemplate(w, "admin.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleMaintenance displays the maintenance login page
func handleMaintenance(w http.ResponseWriter, r *http.Request) {
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
		AppVersion: appVersion,
	}

	if err := templates.ExecuteTemplate(w, "maintenance-login.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleMaintenanceAuth processes the maintenance token submission
func handleMaintenanceAuth(w http.ResponseWriter, r *http.Request) {
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
		Name:     maintenanceSessionKey,
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

// handleMaintenanceConfig displays the maintenance configuration page
func handleMaintenanceConfig(w http.ResponseWriter, r *http.Request) {
	// Verify authentication (MaintenanceHandler should handle this, but double-check)
	if !isMaintenanceAuthenticated(r) {
		http.Redirect(w, r, "/maintenance", http.StatusSeeOther)
		return
	}

	// Get error from query parameters if present
	errorMsg := r.URL.Query().Get("error")

	// If in maintenance mode and no specific error provided, check the database connection
	if IsMaintenanceMode() && errorMsg == "" {
		// Try connecting to the database to get a more specific error
		_, dbErr := InitDB()
		if dbErr != nil {
			errorMsg = fmt.Sprintf("Database Error: %v", dbErr)
		}
	}

	// Parse current version
	versionParts := []string{"1", "0", "0", "0"} // Default
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
	}

	if err := templates.ExecuteTemplate(w, "maintenance.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleMaintenanceConfigure updates database configuration and attempts to reconnect
func handleMaintenanceConfigure(w http.ResponseWriter, r *http.Request) {
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
		handleMaintenanceError(w, r, "Failed to parse form data: "+err.Error())
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

	// Get version components
	majorStr := r.FormValue("version_major")
	minorStr := r.FormValue("version_minor")
	patchStr := r.FormValue("version_patch")

	// Validate database fields
	if host == "" || port == "" || user == "" || dbname == "" {
		handleMaintenanceError(w, r, "All database fields except password are required")
		return
	}

	// Validate version inputs
	major, err := strconv.Atoi(majorStr)
	if err != nil || major < 0 {
		handleMaintenanceError(w, r, "Invalid major version number")
		return
	}

	minor, err := strconv.Atoi(minorStr)
	if err != nil || minor < 0 {
		handleMaintenanceError(w, r, "Invalid minor version number")
		return
	}

	patch, err := strconv.Atoi(patchStr)
	if err != nil || patch < 0 {
		handleMaintenanceError(w, r, "Invalid patch version number")
		return
	}

	// Handle migration tracking
	if resetMigrations {
		// Reset migration tracking to apply all migrations
		log.Println("Migration tracking reset requested")
		if err := UpdateMigrationStart(0); err != nil {
			log.Printf("Warning: Failed to reset migration tracking: %v", err)
		} else {
			log.Println("Migration tracking reset to 0")
		}
	} else if migrationStart != "" {
		// Parse migration start number, handling different formats (4, 04, 004)
		migNum, err := strconv.Atoi(migrationStart)
		if err != nil {
			handleMaintenanceError(w, r, "Invalid migration number: "+err.Error())
			return
		}

		// Update migration tracking to the specified number
		log.Printf("Setting migration tracking to %d", migNum)
		if err := UpdateMigrationStart(migNum); err != nil {
			log.Printf("Warning: Failed to update migration tracking: %v", err)
		}
	}

	// Update database configuration
	if err := UpdateDatabaseConfig(host, port, user, password, dbname); err != nil {
		handleMaintenanceError(w, r, "Failed to update database configuration: "+err.Error())
		return
	}

	// Update environment variables for database
	os.Setenv("DB_HOST", host)
	os.Setenv("DB_PORT", port)
	os.Setenv("DB_USER", user)
	os.Setenv("DB_PASSWORD", password)
	os.Setenv("DB_NAME", dbname)

	// Try to connect to database
	db, err := InitDB()
	if err != nil {
		handleMaintenanceError(w, r, "Failed to connect to database: "+err.Error())
		return
	}
	// Close the test connection
	if db != nil {
		db.Close()
	}

	// Handle version update
	// Parse current version to get build number
	buildNumber := 0
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
		handleMaintenanceError(w, r, "Failed to read .env file: "+err.Error())
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

	// Write back to .env
	err = os.WriteFile(envPath, []byte(strings.Join(lines, "\n")), 0644)
	if err != nil {
		handleMaintenanceError(w, r, "Failed to update .env file: "+err.Error())
		return
	}

	// Update global variables
	appVersion = newVersion
	sessionSalt = generateSessionSalt(appVersion)
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
	}

	w.Header().Set("Refresh", "5;url=/")
	if err := templates.ExecuteTemplate(w, "maintenance.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleMaintenanceError displays error message on maintenance page
func handleMaintenanceError(w http.ResponseWriter, r *http.Request, errorMsg string) {
	// Verify authentication (should be caught by middleware, but double-check)
	if !isMaintenanceAuthenticated(r) {
		http.Redirect(w, r, "/maintenance", http.StatusSeeOther)
		return
	}

	// Parse current version
	versionParts := []string{"1", "0", "0", "0"} // Default
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
	}

	if err := templates.ExecuteTemplate(w, "maintenance.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// incrementBuildVersion increases the build number in APP_VERSION and updates the .env file
func incrementBuildVersion() error {
	// Read current version from environment
	currentVersion := os.Getenv("APP_VERSION")
	if currentVersion == "" {
		currentVersion = "1.0.0.0" // Default if not set
	}

	// Parse version
	parts := strings.Split(currentVersion, ".")
	if len(parts) != 4 {
		// Invalid format, initialize to default
		parts = []string{"1", "0", "0", "0"}
	}

	// Increment build number (last part)
	buildNumber, err := strconv.Atoi(parts[3])
	if err != nil {
		buildNumber = 0 // Reset if parsing failed
	}
	buildNumber++
	parts[3] = strconv.Itoa(buildNumber)

	// Reassemble version string
	newVersion := strings.Join(parts, ".")
	appVersion = newVersion

	// Update session salt
	sessionSalt = generateSessionSalt(appVersion)

	// Load current .env content
	envPath := ".env"
	content, err := os.ReadFile(envPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to read .env file: %v", err)
		}
		// File doesn't exist, create it with just the version
		content = []byte("")
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

	// Write back to .env
	err = os.WriteFile(envPath, []byte(strings.Join(lines, "\n")), 0644)
	if err != nil {
		return fmt.Errorf("failed to update .env file: %v", err)
	}

	// Update environment variable
	os.Setenv("APP_VERSION", newVersion)
	log.Printf("Application version updated to %s", newVersion)
	return nil
}

// generateSessionSalt creates a unique salt based on the app version
func generateSessionSalt(version string) string {
	h := sha256.New()
	h.Write([]byte(version))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// handleInitializeSchema manually executes the database schema initialization
func handleInitializeSchema(w http.ResponseWriter, r *http.Request) {
	// Verify authentication
	if !isMaintenanceAuthenticated(r) {
		http.Redirect(w, r, "/maintenance", http.StatusSeeOther)
		return
	}

	// Reset migration tracking to apply all migrations
	if err := UpdateMigrationStart(0); err != nil {
		http.Redirect(w, r, "/maintenance/config?error="+url.QueryEscape("Failed to reset migration tracking: "+err.Error()), http.StatusSeeOther)
		return
	}

	// Get database connection
	db, err := InitDB()
	if err != nil {
		// Return to maintenance config with error
		http.Redirect(w, r, "/maintenance/config?error="+url.QueryEscape("Failed to connect to database: "+err.Error()), http.StatusSeeOther)
		return
	}
	defer db.Close()

	// Return to maintenance config with success message
	http.Redirect(w, r, "/maintenance/config?success=Database+schema+initialized+successfully", http.StatusSeeOther)
}

// Main function - wrapper that calls StartServer
func main() {
	// Uncomment the following line to start the agent instead of the main server
	// StartAgent()

	// Start the main server
	StartServer()
}

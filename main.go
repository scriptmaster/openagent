package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Global variables
var (
	templates *template.Template
	// Session key for maintenance mode authentication
	maintenanceSessionKey = "maintenance_auth"
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
	AppName   string
	PageTitle string
	User      User
	Error     string
	Projects  []interface{}
}

// Maintenance session for secure access
func isMaintenanceAuthenticated(r *http.Request) bool {
	// Check for maintenance cookie
	cookie, err := r.Cookie(maintenanceSessionKey)
	if err != nil {
		return false
	}

	// Validate cookie value
	return cookie.Value == "authenticated"
}

// MaintenanceHandler handles requests when in maintenance mode
func MaintenanceHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip maintenance mode check for static files
		if strings.HasPrefix(r.URL.Path, "/static/") {
			next.ServeHTTP(w, r)
			return
		}

		// Allow access to maintenance authentication endpoints
		if r.URL.Path == "/maintenance" || r.URL.Path == "/maintenance/auth" {
			next.ServeHTTP(w, r)
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

	// Load environment variables
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
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
			AppName:   "OpenAgent",
			PageTitle: "Login - OpenAgent",
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
	http.HandleFunc("/maintenance/configure", handleMaintenanceConfigure)

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
	fmt.Fprintf(w, "Welcome to the dashboard!")
}

func handleProjects(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Projects page")
}

func handleAdmin(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Admin dashboard")
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
		Error string
	}{
		Error: r.URL.Query().Get("error"),
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
		Value:    "authenticated",
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

	data := struct {
		DBHost     string
		DBPort     string
		DBUser     string
		DBPassword string
		DBName     string
		Error      string
		Success    string
	}{
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
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

	// Get form values
	host := r.FormValue("db_host")
	port := r.FormValue("db_port")
	user := r.FormValue("db_user")
	password := r.FormValue("db_password")
	dbname := r.FormValue("db_name")

	// Validate required fields
	if host == "" || port == "" || user == "" || dbname == "" {
		handleMaintenanceError(w, r, "All fields except password are required")
		return
	}

	// Update .env file
	if err := UpdateDatabaseConfig(host, port, user, password, dbname); err != nil {
		handleMaintenanceError(w, r, "Failed to update configuration: "+err.Error())
		return
	}

	// Update environment variables
	os.Setenv("DB_HOST", host)
	os.Setenv("DB_PORT", port)
	os.Setenv("DB_USER", user)
	os.Setenv("DB_PASSWORD", password)
	os.Setenv("DB_NAME", dbname)

	// Try to reconnect to database
	db, err := InitDB()
	if err != nil {
		handleMaintenanceError(w, r, "Failed to connect to database: "+err.Error())
		return
	}

	// Close the test connection
	if db != nil {
		db.Close()
	}

	// Show success message and redirect to home after a delay
	data := struct {
		DBHost     string
		DBPort     string
		DBUser     string
		DBPassword string
		DBName     string
		Error      string
		Success    string
	}{
		DBHost:     host,
		DBPort:     port,
		DBUser:     user,
		DBPassword: password,
		DBName:     dbname,
		Success:    "Configuration updated successfully! Server will restart in 5 seconds...",
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

	data := struct {
		DBHost     string
		DBPort     string
		DBUser     string
		DBPassword string
		DBName     string
		Error      string
		Success    string
	}{
		DBHost:     r.FormValue("db_host"),
		DBPort:     r.FormValue("db_port"),
		DBUser:     r.FormValue("db_user"),
		DBPassword: r.FormValue("db_password"),
		DBName:     r.FormValue("db_name"),
		Error:      errorMsg,
	}

	if err := templates.ExecuteTemplate(w, "maintenance.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Main function - wrapper that calls StartServer
func main() {
	// Uncomment the following line to start the agent instead of the main server
	// StartAgent()

	// Start the main server
	StartServer()
}

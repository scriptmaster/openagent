package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

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
	templates := template.Must(template.ParseGlob("go/*.html"))

	// Setup static file server
	fs := http.FileServer(http.Dir("go/static"))
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
	http.Handle("/", AuthMiddleware(authMux))

	// Agent API endpoints from agent.go
	http.HandleFunc("/agent.html", handleRoot)
	http.HandleFunc("/start", handleStart)
	http.HandleFunc("/next", handleNextStep)
	http.HandleFunc("/status", handleStatus)

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
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

// Main function - wrapper that calls StartServer
func main() {
	// Uncomment the following line to start the agent instead of the main server
	// StartAgent()

	// Start the main server
	StartServer()
}

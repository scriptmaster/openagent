package auth

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/scriptmaster/openagent/common" // Updated import path
)

// Global template variable for auth handlers
var authTemplates *template.Template

// InitAuthTemplates initializes the template variable for this package
func InitAuthTemplates(t *template.Template) {
	authTemplates = t
}

// HandleLogin displays the login page
func HandleLogin(w http.ResponseWriter, r *http.Request) {
	// Ensure templates are initialized
	if authTemplates == nil {
		http.Error(w, "Auth templates not initialized", http.StatusInternalServerError)
		log.Println("Error: HandleLogin called before InitAuthTemplates")
		return
	}
	data := struct {
		AppName    string
		PageTitle  string
		AdminEmail string
		AppVersion string
		Error      string
	}{
		AppName:    common.GetEnvOrDefault("APP_NAME", "OpenAgent"),
		PageTitle:  "Login - " + common.GetEnvOrDefault("APP_NAME", "OpenAgent"),
		AdminEmail: os.Getenv("SYSADMIN_EMAIL"),
		AppVersion: os.Getenv("APP_VERSION"),
		Error:      r.URL.Query().Get("error"),
	}
	if err := authTemplates.ExecuteTemplate(w, "login.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleLogout clears session cookies and redirects to login
func HandleLogout(w http.ResponseWriter, r *http.Request) {
	// Get versioned cookie name
	cookieName := GetSessionCookieName()

	// Clear the versioned session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		Expires:  time.Now().Add(-24 * time.Hour),
		SameSite: http.SameSiteLaxMode,
	})

	// Clear the regular session cookie for backward compatibility
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		Expires:  time.Now().Add(-24 * time.Hour),
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect to login page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// HandleRequestOTP handles OTP requests
func HandleRequestOTP(w http.ResponseWriter, r *http.Request, userService *UserService) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if userService is nil
	if userService == nil {
		log.Printf("Error: UserService is nil in HandleRequestOTP")
		SendJSONResponse(w, false, "System error: User service not available", nil, "")
		return
	}

	// Parse JSON request using OTPRequest struct
	var req OTPRequest // Uses OTPRequest from types.go
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		SendJSONResponse(w, false, "Invalid request", nil, "")
		return
	}

	// Validate email
	if req.Email == "" {
		SendJSONResponse(w, false, "Email is required", nil, "")
		return
	}

	// Check if user exists
	user, err := userService.GetUserByEmail(req.Email)
	if err != nil {
		// User doesn't exist - check if we should allow registration

		// Check if any admin exists in the system
		adminExists, err := userService.CheckIfAdminExists()
		if err != nil {
			log.Printf("Failed to check if admin exists: %v", err)
			SendJSONResponse(w, false, "System error, please try again later", nil, "")
			return
		}

		if adminExists {
			// Admin exists but this user doesn't - reject login
			log.Printf("Login attempt for non-existent user: %s", req.Email)
			SendJSONResponse(w, false, "No account found with this email. Please contact administrator.", nil, "")
			return
		}

		// No admin exists - this is first-user scenario
		// Create the user (who will be an admin)
		user, err = userService.CreateUser(req.Email)
		if err != nil {
			log.Printf("Failed to create first admin user: %v", err)
			SendJSONResponse(w, false, "Failed to create user account", nil, "")
			return
		}

		log.Printf("Created first admin user: %s", req.Email)
	}

	// Send OTP
	if err := SendOTP(req.Email); err != nil {
		log.Printf("Failed to send OTP for user %s: %v", user.Email, err)
		SendJSONResponse(w, false, "Failed to send OTP", nil, "")
		return
	}

	log.Printf("OTP sent successfully to user: %s", user.Email)
	SendJSONResponse(w, true, "OTP sent successfully", nil, "")
}

// HandleVerifyOTP handles OTP verification
func HandleVerifyOTP(w http.ResponseWriter, r *http.Request, userService *UserService) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON request
	var req OTPVerifyRequest // Uses OTPVerifyRequest from types.go
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		SendJSONResponse(w, false, "Invalid request", nil, "")
		return
	}

	// Validate request
	if req.Email == "" || req.OTP == "" {
		SendJSONResponse(w, false, "Email and OTP are required", nil, "")
		return
	}

	// Verify OTP
	valid, err := VerifyOTP(req.Email, req.OTP)
	if err != nil {
		log.Printf("OTP verification error: %v", err)
		SendJSONResponse(w, false, "Failed to verify OTP", nil, "")
		return
	}

	if !valid {
		SendJSONResponse(w, false, "Invalid OTP", nil, "")
		return
	}

	// Get user
	user, err := userService.GetUserByEmail(req.Email)
	if err != nil {
		log.Printf("Failed to get user: %v", err)
		SendJSONResponse(w, false, "User not found", nil, "")
		return
	}

	// Update last login
	if err := userService.UpdateUserLastLogin(user.ID); err != nil {
		log.Printf("Failed to update last login: %v", err)
		// Continue anyway, this is not critical
	}

	// Create session
	session, err := CreateSession(user)
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		SendJSONResponse(w, false, "Failed to create session", nil, "")
		return
	}

	// Get versioned cookie name
	cookieName := GetSessionCookieName()

	// Set versioned session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    session.Token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   int((session.ExpiresAt.Sub(time.Now())).Seconds()),
		SameSite: http.SameSiteLaxMode,
	})

	// Also set regular session cookie for backward compatibility
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    session.Token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   int((session.ExpiresAt.Sub(time.Now())).Seconds()),
		SameSite: http.SameSiteLaxMode,
	})

	SendJSONResponse(w, true, "Login successful", nil, "/")
}

// SendJSONResponse sends a JSON response
func SendJSONResponse(w http.ResponseWriter, success bool, message string, data interface{}, redirect string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(JSONResponse{ // Uses JSONResponse from types.go
		Success:  success,
		Message:  message,
		Data:     data,
		Redirect: redirect,
	})
}

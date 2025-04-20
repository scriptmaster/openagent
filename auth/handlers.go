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
		User       *User
	}{
		AppName:    common.GetEnvOrDefault("APP_NAME", "OpenAgent"),
		PageTitle:  "Login - " + common.GetEnvOrDefault("APP_NAME", "OpenAgent"),
		AdminEmail: os.Getenv("SYSADMIN_EMAIL"),
		AppVersion: common.GetEnvOrDefault("APP_VERSION", "1.0.0.0"),
		Error:      r.URL.Query().Get("error"),
		User:       nil,
	}
	if err := authTemplates.ExecuteTemplate(w, "login.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleLogout clears session cookies and redirects to login
func HandleLogout(w http.ResponseWriter, r *http.Request) {
	cookieName := GetSessionCookieName()

	// Clear the versioned JWT session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,              // Expire immediately
		Expires:  time.Unix(0, 0), // Set explicit expiry in the past
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil,
	})

	// No need to clear the old "session" cookie anymore if unused

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
		// For the first user, we need a temporary password to store.
		// Ideally, this should prompt for setup, but for now, use a default or random one.
		// IMPORTANT: This is insecure for production. First user setup needs a proper flow.
		// Using the email temporarily as a placeholder password for hashing.
		tempPassword := req.Email // Or generate a random string
		log.Printf("WARN: Creating first user (%s) with temporary password derived from email. Needs proper setup flow.", req.Email)
		user, err = userService.CreateUser(req.Email, tempPassword)
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
		// Send a specific response indicating OTP failure
		SendJSONResponse(w, false, "Failed to send OTP. You can try logging in with your password.", map[string]bool{"otp_error": true}, "")
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
		// Don't fail the login for this, just log it
	}

	// Create session JWT
	jwtString, err := CreateSession(user) // Returns JWT string
	if err != nil {
		log.Printf("Failed to create session JWT: %v", err)
		SendJSONResponse(w, false, "Failed to create session", nil, "")
		return
	}

	// Set session cookie with JWT
	cookieName := GetSessionCookieName()
	expiryDuration := 168 * time.Hour // Match JWT expiry (7 days)
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    jwtString,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   int(expiryDuration.Seconds()),
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil, // Add Secure flag if using HTTPS
	})
	// No need for backward compatible "session" cookie with JWT

	// Determine redirect based on admin status
	redirectURL := "/" // Default for non-admins
	if user.IsAdmin {
		redirectURL = "/dashboard"
	}

	SendJSONResponse(w, true, "OTP verified successfully", nil, redirectURL)
}

// PasswordLoginRequest represents the request body for password login
type PasswordLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// HandlePasswordLogin handles login attempts using email and password
func HandlePasswordLogin(w http.ResponseWriter, r *http.Request, userService *UserService) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON request
	var req PasswordLoginRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		SendJSONResponse(w, false, "Invalid request format", nil, "")
		return
	}

	// Validate request
	if req.Email == "" || req.Password == "" {
		SendJSONResponse(w, false, "Email and password are required", nil, "")
		return
	}

	// Verify password
	user, err := userService.VerifyPassword(req.Email, req.Password)
	if err != nil {
		// Log the specific error for debugging, but send a generic message to the client
		log.Printf("Password verification failed for %s: %v", req.Email, err)
		SendJSONResponse(w, false, "Invalid email or password", nil, "")
		return
	}

	// Update last login
	if err := userService.UpdateUserLastLogin(user.ID); err != nil {
		log.Printf("Failed to update last login: %v", err)
		// Don't fail the login for this, just log it
	}

	// Create session JWT
	jwtString, err := CreateSession(user)
	if err != nil {
		log.Printf("Failed to create session JWT: %v", err)
		SendJSONResponse(w, false, "Failed to create session", nil, "")
		return
	}

	// Set session cookie with JWT
	cookieName := GetSessionCookieName()
	expiryDuration := 168 * time.Hour // Match JWT expiry (7 days)
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    jwtString,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   int(expiryDuration.Seconds()),
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil, // Add Secure flag if using HTTPS
	})
	// No need for backward compatible "session" cookie with JWT

	// Determine redirect based on admin status
	redirectURL := "/" // Default for non-admins
	if user.IsAdmin {
		redirectURL = "/dashboard"
	}

	SendJSONResponse(w, true, "Login successful", nil, redirectURL)
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

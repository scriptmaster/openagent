package auth

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// JSONResponse object for API responses
type JSONResponse struct {
	Success  bool        `json:"success"`
	Message  string      `json:"message,omitempty"`
	Data     interface{} `json:"data,omitempty"`
	Redirect string      `json:"redirect,omitempty"`
}

// OTPRequest represents the request body for requesting an OTP
type OTPRequest struct {
	Email string `json:"email"`
}

// OTPVerifyRequest represents the request body for verifying an OTP
type OTPVerifyRequest struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

// HandleRequestOTP handles OTP request
func HandleRequestOTP(w http.ResponseWriter, r *http.Request, userService *UserService) {
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
	var req OTPVerifyRequest
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
	json.NewEncoder(w).Encode(JSONResponse{
		Success:  success,
		Message:  message,
		Data:     data,
		Redirect: redirect,
	})
}

package auth

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/scriptmaster/openagent/common"
)

// ProfilePageData holds data for the profile page template
type ProfilePageData struct {
	AppName    string
	PageTitle  string
	User       *User // Use pointer to easily check if logged in
	AppVersion string
	Error      string
	Success    string
}

// HandleProfilePage renders the user profile page.
func HandleProfilePage(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login?error=unauthorized", http.StatusSeeOther)
		return
	}

	if authTemplates == nil {
		http.Error(w, "Auth templates not initialized", http.StatusInternalServerError)
		log.Println("Error: HandleProfilePage called before InitAuthTemplates")
		return
	}

	data := ProfilePageData{
		AppName:    common.GetEnvOrDefault("APP_NAME", "OpenAgent"),
		PageTitle:  "User Profile",
		User:       user,
		AppVersion: common.GetEnvOrDefault("APP_VERSION", "1.0.0.0"),
		Error:      r.URL.Query().Get("error"),
		Success:    r.URL.Query().Get("success"),
	}

	if err := authTemplates.ExecuteTemplate(w, "profile.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// --- API Handlers ---

// UpdateProfileRequest defines the structure for profile update API calls
type UpdateProfileRequest struct {
	Name        *string `json:"name,omitempty"`         // Pointer to distinguish empty vs not provided
	ProfileIcon *string `json:"profile_icon,omitempty"` // Pointer for icon URL
}

// HandleUpdateProfileAPI handles updates to user name and profile icon.
func HandleUpdateProfileAPI(w http.ResponseWriter, r *http.Request, userService UserServicer) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		common.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.JSONError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	updated := false
	// Update Name if provided
	if req.Name != nil {
		if err := userService.UpdateUserName(r.Context(), user.ID, *req.Name); err != nil {
			log.Printf("Error updating name for user %d: %v", user.ID, err)
			common.JSONError(w, "Failed to update name", http.StatusInternalServerError)
			return
		}
		user.Name = *req.Name // Update user in context for immediate reflection (if needed)
		updated = true
	}

	// Update Profile Icon if provided
	if req.ProfileIcon != nil {
		// TODO: Add validation for URL format?
		if err := userService.UpdateUserProfileIcon(r.Context(), user.ID, *req.ProfileIcon); err != nil {
			log.Printf("Error updating profile icon for user %d: %v", user.ID, err)
			common.JSONError(w, "Failed to update profile icon", http.StatusInternalServerError)
			return
		}
		user.ProfileIcon = *req.ProfileIcon // Update user in context
		updated = true
	}

	if !updated {
		common.JSONResponse(w, map[string]string{"message": "No changes detected"})
		return
	}

	// Return success with the updated user data (excluding password hash)
	common.JSONResponse(w, map[string]interface{}{
		"message": "Profile updated successfully",
		"user":    user, // User object from context should reflect changes made
	})
}

// ChangePasswordRequest defines the structure for the password change API call
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"` // Optional if using OTP
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
	OTP             string `json:"otp"` // OTP for verification
}

// HandleChangePasswordAPI handles password change requests.
// Requires OTP verification.
func HandleChangePasswordAPI(w http.ResponseWriter, r *http.Request, userService UserServicer) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		common.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.JSONError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Basic Validation
	if req.OTP == "" {
		common.JSONError(w, "OTP is required", http.StatusBadRequest)
		return
	}
	if req.NewPassword == "" || req.ConfirmPassword == "" {
		common.JSONError(w, "New password and confirmation are required", http.StatusBadRequest)
		return
	}
	if req.NewPassword != req.ConfirmPassword {
		common.JSONError(w, "New password and confirmation do not match", http.StatusBadRequest)
		return
	}
	// TODO: Add password strength validation?

	// Verify OTP
	valid, err := VerifyOTP(user.Email, req.OTP)
	if err != nil || !valid {
		log.Printf("OTP verification failed for user %d during password change: %v (valid: %t)", user.ID, err, valid)
		errMsg := "Invalid OTP"
		if err != nil {
			errMsg = "OTP verification failed: " + err.Error()
		}
		common.JSONError(w, errMsg, http.StatusBadRequest)
		return
	}

	// Generate hash for the new password
	newHash, err := GeneratePasswordHash(req.NewPassword)
	if err != nil {
		log.Printf("Error hashing new password for user %d: %v", user.ID, err)
		common.JSONError(w, "Failed to process new password", http.StatusInternalServerError)
		return
	}

	// Update the password hash in the database
	if err := userService.UpdatePasswordHash(r.Context(), user.ID, newHash); err != nil {
		log.Printf("Error updating password hash for user %d: %v", user.ID, err)
		common.JSONError(w, "Failed to update password", http.StatusInternalServerError)
		return
	}

	common.JSONResponse(w, map[string]string{"message": "Password changed successfully"})
}

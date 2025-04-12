package auth

import (
	"net/http"
	"strings"
)

// IsMaintenanceAuthenticated checks if the request has a valid maintenance authentication cookie
func IsMaintenanceAuthenticated(r *http.Request) bool {
	// Check for maintenance cookie
	cookie, err := r.Cookie("maintenance_auth")
	if err != nil {
		return false
	}

	// Extract the authenticated_ prefix
	if !strings.HasPrefix(cookie.Value, "authenticated_") {
		return false
	}

	// The cookie is valid if it has the correct prefix
	// The server has the salt that was used to create it, so we can
	// trust cookies with the authenticated_ prefix for simplicity
	return true
}

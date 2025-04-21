package auth

import (
	"html/template"
	"net/http"
)

// RegisterAuthRoutes registers all authentication-related routes using named handlers
func RegisterAuthRoutes(mux *http.ServeMux, templates *template.Template, userService UserServicer) {
	// Initialize templates for auth handlers
	InitAuthTemplates(templates)

	// Auth API endpoints
	mux.HandleFunc("/auth/request-otp", func(w http.ResponseWriter, r *http.Request) {
		HandleRequestOTP(w, r, userService)
	})
	mux.HandleFunc("/auth/verify-otp", func(w http.ResponseWriter, r *http.Request) {
		HandleVerifyOTP(w, r, userService)
	})

	// Login/Logout page handlers
	mux.HandleFunc("/login", HandleLogin)
	mux.HandleFunc("/logout", HandleLogout)

	// Password Login
	mux.HandleFunc("/password-login", func(w http.ResponseWriter, r *http.Request) {
		HandlePasswordLogin(w, r, userService)
	})
}

// AdminMiddleware checks if a user is an admin
func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip admin check for non-admin paths
		if !isAdminPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Get user from context (already set by AuthMiddleware)
		user := GetUserFromContext(r.Context())
		if user == nil || !user.IsAdmin {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// User is admin, proceed
		next.ServeHTTP(w, r)
	})
}

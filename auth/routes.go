package auth

import (
	"log"
	"net/http"

	"github.com/scriptmaster/openagent/types"
)

// RegisterAuthRoutes registers all authentication-related routes using named handlers
func RegisterAuthRoutes(router *http.ServeMux, templates types.TemplateEngineInterface, userService UserServicer) {
	// Initialize templates for auth handlers
	InitAuthTemplates(templates)

	log.Printf("\t → \t → 6.X Registering Auth Routes /auth/*, /login, /logout")

	// Login/Logout page handlers
	router.HandleFunc("/login", HandleLogin)
	router.HandleFunc("/logout", HandleLogout)

	// Auth API endpoints
	router.HandleFunc("/auth/request-otp", func(w http.ResponseWriter, r *http.Request) {
		HandleRequestOTP(w, r, userService)
	})
	router.HandleFunc("/auth/verify-otp", func(w http.ResponseWriter, r *http.Request) {
		HandleVerifyOTP(w, r, userService)
	})

	// Password Login
	router.HandleFunc("/auth/password-login", func(w http.ResponseWriter, r *http.Request) {
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

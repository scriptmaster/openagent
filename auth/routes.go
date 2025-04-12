package auth

import (
	"html/template"
	"net/http"
	"time"
)

// RegisterAuthRoutes registers all authentication-related routes using named handlers
func RegisterAuthRoutes(mux *http.ServeMux, templates *template.Template, userService *UserService) {
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
}

// AuthMiddleware checks if a user is authenticated and sets user context
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth check for public paths
		if isPublicPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Get session cookie
		cookie, err := r.Cookie("session")
		if err != nil {
			// No session cookie, redirect to login
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Validate session
		session, valid := GetSession(cookie.Value)
		if !valid {
			// Invalid or expired session, clear cookie and redirect to login
			expiredCookie := &http.Cookie{
				Name:     "session",
				Value:    "",
				Path:     "/",
				Expires:  time.Unix(0, 0),
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
			}
			http.SetCookie(w, expiredCookie)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Add user to context
		ctx := SetUserContext(r.Context(), session.User)
		// Serve with the updated context
		next.ServeHTTP(w, r.WithContext(ctx))
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

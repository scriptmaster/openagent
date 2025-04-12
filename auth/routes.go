package auth

import (
	"html/template"
	"net/http"
	"os"
	"strings"
	"time"
)

// RegisterAuthRoutes registers all authentication-related routes
func RegisterAuthRoutes(mux *http.ServeMux, templates *template.Template, userService *UserService) {
	// Auth endpoints
	mux.HandleFunc("/auth/request-otp", func(w http.ResponseWriter, r *http.Request) {
		HandleRequestOTP(w, r, userService)
	})

	mux.HandleFunc("/auth/verify-otp", func(w http.ResponseWriter, r *http.Request) {
		HandleVerifyOTP(w, r, userService)
	})

	// Login/Logout pages
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			AppName    string
			PageTitle  string
			AdminEmail string
			AppVersion string
			Error      string
		}{
			AppName:    "OpenAgent",
			PageTitle:  "Login - OpenAgent",
			AdminEmail: os.Getenv("SYSADMIN_EMAIL"),
			AppVersion: os.Getenv("APP_VERSION"),
			Error:      r.URL.Query().Get("error"),
		}
		if err := templates.ExecuteTemplate(w, "login.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
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
	})
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

// isPublicPath returns true for paths that don't require authentication
func isPublicPath(path string) bool {
	publicPaths := []string{
		"/login",
		"/api/request-otp",
		"/api/verify-otp",
		"/static/",
		"/favicon.ico",
	}

	for _, pp := range publicPaths {
		if strings.HasPrefix(path, pp) {
			return true
		}
	}

	return false
}

// isAdminPath returns true for paths that require admin access
func isAdminPath(path string) bool {
	adminPaths := []string{
		"/admin",
		"/api/admin/",
	}

	for _, ap := range adminPaths {
		if strings.HasPrefix(path, ap) {
			return true
		}
	}

	return false
}

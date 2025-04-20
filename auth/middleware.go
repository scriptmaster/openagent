package auth

import (
	"log"
	"net/http"
)

// userCtxKey removed, defined in types.go

// AuthMiddleware checks for a valid session cookie and adds user info to context
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Initialize JWT secret if not already done (important for middleware)
		if jwtSecret == nil {
			if err := InitializeJWT(); err != nil {
				log.Printf("CRITICAL: JWT not initialized in middleware: %v", err)
				// Handle this potentially by redirecting to an error page or login
				http.Error(w, "Server configuration error", http.StatusInternalServerError)
				return
			}
		}

		cookieName := GetSessionCookieName()
		cookie, err := r.Cookie(cookieName)
		if err != nil {
			// No cookie found, treat as not logged in
			log.Printf("No session cookie '%s' found, redirecting to login.", cookieName)
			http.Redirect(w, r, "/login?error=unauthorized", http.StatusSeeOther)
			return
		}

		// Validate the JWT
		claims, err := ValidateJWT(cookie.Value)
		if err != nil {
			// Invalid or expired token
			log.Printf("Invalid JWT found: %v, redirecting to login.", err)
			// Clear the invalid cookie
			http.SetCookie(w, &http.Cookie{
				Name:   cookieName,
				Value:  "",
				Path:   "/",
				MaxAge: -1,
			})
			http.Redirect(w, r, "/login?error=session_expired", http.StatusSeeOther)
			return
		}

		// Token is valid, create User struct from claims
		user := &User{
			ID:      claims.UserID,
			Email:   claims.Email,
			IsAdmin: claims.IsAdmin,
			// CreatedAt/LastLoggedIn aren't typically stored in JWT,
			// Fetch from DB if needed in handlers, or omit from context User
		}

		// Add user to context
		ctx := SetUserContext(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// IsAdminMiddleware checks if the user in the context is an admin
// Assumes AuthMiddleware has already run
func IsAdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context())
		if user == nil || !user.IsAdmin {
			// User not found in context or is not an admin
			log.Printf("WARN: Admin access denied for user '%v' to path %s", user, r.URL.Path)
			// Consider redirecting to login or showing a specific "access denied" page
			http.Error(w, "Forbidden: Administrator access required", http.StatusForbidden)
			return
		}
		// User is an admin, proceed to the next handler
		next.ServeHTTP(w, r)
	})
}

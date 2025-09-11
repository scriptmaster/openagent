package auth

import (
	"net/http"
	"time"
)

// SetSessionCookie sets a session cookie with JWT token
func SetSessionCookie(w http.ResponseWriter, jwtString string, expiryDuration time.Duration) {
	cookieName := GetSessionCookieName()
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    jwtString,
		Path:     "/",
		Expires:  time.Now().Add(expiryDuration),
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearSessionCookie clears the session cookie
func ClearSessionCookie(w http.ResponseWriter) {
	cookieName := GetSessionCookieName()
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})
}

// SetMaintenanceCookie sets a maintenance authentication cookie
func SetMaintenanceCookie(w http.ResponseWriter, sessionSalt string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "maintenance_auth",
		Value:    "authenticated_" + sessionSalt[:8], // Add partial version salt to invalidate on restart
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour), // 24 hour expiry for maintenance
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearMaintenanceCookie clears the maintenance authentication cookie
func ClearMaintenanceCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "maintenance_auth",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})
}

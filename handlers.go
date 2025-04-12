package main

import (
	"net/http"
	"os"
	"time"

	"github.com/scriptmaster/openagent/models"
)

// AgentRequest represents a request to the agent
type AgentRequest struct {
	Prompt string `json:"prompt"`
}

// AgentResponse represents a response from the agent
type AgentResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// Note: Agent handlers (handleAgent, handleStart, handleNextStep, handleStatus)
// are implemented in agent.go to avoid duplication.

// NOTE: handleLogin and handleLogout were presumably already defined here or elsewhere.
// The previous edit incorrectly added duplicates.
// If they are defined elsewhere (e.g., in auth package), they should be called appropriately
// from routes.go instead of being defined here.

// --- Handlers moved from routes.go ---

// handleLogin displays the login page
func handleLogin(w http.ResponseWriter, r *http.Request) {
	// Use models.PageData
	data := models.PageData{
		AppName:    "OpenAgent",
		PageTitle:  "Login - OpenAgent",
		AdminEmail: os.Getenv("SYSADMIN_EMAIL"),
		AppVersion: appVersion, // appVersion is a global in main package
	}

	if err := templates.ExecuteTemplate(w, "login.html", data); err != nil { // templates is a global in main package
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleLogout clears the session cookie and redirects to login
func handleLogout(w http.ResponseWriter, r *http.Request) {
	// Clear the session cookie by setting an expired cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,                              // Delete cookie immediately
		Expires:  time.Now().Add(-24 * time.Hour), // Expired
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect to login page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

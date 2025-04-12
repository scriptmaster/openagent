package models

import (
	"database/sql"
	"time"

	"github.com/scriptmaster/openagent/auth"
)

// User represents a user in the system (consider moving to auth/types.go if only used there)
type User struct {
	ID           int
	Email        string
	IsAdmin      bool
	CreatedAt    time.Time
	LastLoggedIn sql.NullTime
}

// OTPData stores one-time password information (consider moving to auth/types.go)
type OTPData struct {
	Email     string
	OTP       string
	ExpiresAt time.Time
	Attempts  int
}

// Session represents a user session (consider moving to auth/types.go)
type Session struct {
	UserID    int // Consider using auth.User directly if appropriate
	Email     string
	IsAdmin   bool
	ExpiresAt time.Time
}

// JSONResponse is a standard response format for JSON APIs
type JSONResponse struct {
	Success  bool        `json:"success"`
	Message  string      `json:"message,omitempty"`
	Data     interface{} `json:"data,omitempty"`
	Redirect string      `json:"redirect,omitempty"`
}

// PageData represents common data passed to various page templates.
type PageData struct {
	AppName    string
	PageTitle  string
	User       auth.User // Using auth.User for broader context
	Error      string
	Success    string        // Added Success field for consistency
	Projects   []interface{} // Consider a more specific type like []*projects.Project if possible
	Project    interface{}   // Consider a more specific type
	AdminEmail string
	AppVersion string
	Stats      interface{} // Placeholder for stats data if needed
}

// Add other shared model structs here

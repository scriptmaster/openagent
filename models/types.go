package models

import (
	// "database/sql" // No longer needed here
	// "time" // No longer needed here

	"github.com/scriptmaster/openagent/auth"
)

// User moved to auth/types.go
// OTPData moved to auth/types.go
// Session moved to auth/types.go

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
	Success    string
	Projects   []interface{} // Consider a more specific type like []*projects.Project
	Project    interface{}   // Consider a more specific type like *projects.Project
	AdminEmail string
	AppVersion string
	Stats      interface{} // Placeholder for stats data
}

// Add other shared model structs here

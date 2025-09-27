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

// AdminStats represents statistics for the admin dashboard
type AdminStats struct {
	ProjectCount    int `json:"project_count"`
	ConnectionCount int `json:"connection_count"`
	TableCount      int `json:"table_count"`
	UserCount       int `json:"user_count"`
}

// PageData represents common data passed to various page templates.
type PageData struct {
	AppName        string
	PageTitle      string
	User           *auth.User // Changed to pointer to handle nil cases
	Error          string
	Success        string
	Projects       []interface{} // Consider a more specific type like []*projects.Project
	Project        interface{}   // Consider a more specific type like *projects.Project
	AdminEmail     string
	AppVersion     string
	Stats          *AdminStats   // Admin dashboard statistics
	CurrentHost    string        // Add CurrentHost field
	RecentActivity []interface{} // Recent database activity for admin dashboard
	SystemHealth   interface{}   // System health information for admin dashboard
}

// Page represents a project-specific page
type Page struct {
	ID              int    `json:"id" db:"id"`
	ProjectID       int    `json:"project_id" db:"project_id"`
	Title           string `json:"title" db:"title"`
	Slug            string `json:"slug" db:"slug"`
	HTMLContent     string `json:"html_content" db:"html_content"`
	IsLanding       bool   `json:"is_landing" db:"is_landing"`
	IsActive        bool   `json:"is_active" db:"is_active"`
	MetaTitle       string `json:"meta_title" db:"meta_title"`
	MetaDescription string `json:"meta_description" db:"meta_description"`
	CreatedAt       string `json:"created_at" db:"created_at"`
	UpdatedAt       string `json:"updated_at" db:"updated_at"`
}

// Add other shared model structs here

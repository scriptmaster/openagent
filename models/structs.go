package models

import (
	"database/sql"
	"time"
)

// User represents a user in the system
type User struct {
	ID           int
	Email        string
	IsAdmin      bool
	CreatedAt    time.Time
	LastLoggedIn sql.NullTime
}

// OTPData stores one-time password information
type OTPData struct {
	Email     string
	OTP       string
	ExpiresAt time.Time
	Attempts  int
}

// Session represents a user session
type Session struct {
	UserID    int
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

// PageData holds data passed to HTML templates
type PageData struct {
	AppName    string
	PageTitle  string
	User       User
	Error      string
	Projects   []interface{}
	Project    interface{}
	AdminEmail string
	AppVersion string
}

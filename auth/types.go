package auth

import (
	"context"
	"time"
)

// --- Structs from functions.go ---

// UserServicer defines the interface for user service operations
type UserServicer interface {
	// GetUserByEmail retrieves a user by email, checking project context first.
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	// CreateUser creates a user, potentially in a project-specific database.
	CreateUser(ctx context.Context, email string) (*User, error)
	// UpdateUserLastLogin updates the last login time, checking project context.
	UpdateUserLastLogin(ctx context.Context, userID int) error
	// VerifyPassword verifies a user's password, checking project context.
	VerifyPassword(ctx context.Context, email, password string) (*User, error)
	// CheckIfAdminExists checks if any admin user exists in the default database.
	CheckIfAdminExists(ctx context.Context) (bool, error)

	// MakeUserAdmin grants admin privileges to a user.
	MakeUserAdmin(ctx context.Context, userID int) error

	// --- Profile Methods ---
	// UpdatePasswordHash updates the user's password hash after verification.
	UpdatePasswordHash(ctx context.Context, userID int, newHash string) error

	// TODO: Add methods for profile updates (Name, Password, Icon) when database schema supports them
}

// User represents a user in the system
type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Store the hashed password, exclude from JSON
	IsAdmin      bool      `json:"is_admin" db:"is_admin"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	LastLoggedIn time.Time `json:"last_logged_in" db:"last_logged_in"`
}

// OTPData stores information about an OTP
type OTPData struct {
	Email     string
	OTP       string
	ExpiresAt time.Time
	Attempts  int
}

// Session represents a user session
type Session struct {
	Token     string
	User      *User
	CreatedAt time.Time
	ExpiresAt time.Time
}

// --- Structs from handlers.go ---

// JSONResponse object for API responses
type JSONResponse struct {
	Success  bool        `json:"success"`
	Message  string      `json:"message,omitempty"`
	Data     interface{} `json:"data,omitempty"`
	Redirect string      `json:"redirect,omitempty"`
}

// OTPRequest represents the request body for requesting an OTP
type OTPRequest struct {
	Email string `json:"email"`
}

// OTPVerifyRequest represents the request body for verifying an OTP
type OTPVerifyRequest struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

// --- Context Key ---

// userCtxKey is the key used for storing user in request context
// Making it private ensures it's only used within this package
type userCtxKey struct{}

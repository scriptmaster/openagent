package auth

import (
	"time"
)

// --- Structs from functions.go ---

// UserServicer defines the interface for user service operations
type UserServicer interface {
	GetUserByEmail(email string) (*User, error)
	CreateUser(email string) (*User, error)
	UpdateUserLastLogin(userID int) error
}

// User represents a user in the system
type User struct {
	ID           int
	Email        string
	IsAdmin      bool
	CreatedAt    time.Time
	LastLoggedIn time.Time // Consider using sql.NullTime if nullable
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

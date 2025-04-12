package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

// UserServicer moved to types.go
// User moved to types.go
// OTPData moved to types.go
// Session moved to types.go
// userCtxKey moved to types.go

var (
	otpStore     = make(map[string]OTPData) // Uses OTPData from types.go
	otpMutex     = &sync.Mutex{}
	sessions     = make(map[string]Session) // Uses Session from types.go
	sessionMutex = &sync.RWMutex{}
)

// SendOTP generates and sends an OTP to the specified email
func SendOTP(email string) error {
	// Generate a random 6-digit OTP
	otp, err := generateOTP(6)
	if err != nil {
		return err
	}

	// Store OTP with expiration time
	otpMutex.Lock()
	otpStore[email] = OTPData{ // Uses OTPData from types.go
		Email:     email,
		OTP:       otp,
		ExpiresAt: time.Now().Add(5 * time.Minute),
		Attempts:  0,
	}
	otpMutex.Unlock()

	// Send email with OTP
	err = sendOTPEmail(email, otp)
	if err != nil {
		log.Printf("Error sending OTP email: %v", err)
		return err
	}

	return nil
}

// VerifyOTP checks if the provided OTP is valid
func VerifyOTP(email, otp string) (bool, error) {
	otpMutex.Lock()
	defer otpMutex.Unlock()

	data, exists := otpStore[email]
	if !exists {
		return false, errors.New("no OTP request found")
	}

	// Check if OTP has expired
	if time.Now().After(data.ExpiresAt) {
		delete(otpStore, email)
		return false, errors.New("OTP has expired")
	}

	// Increment attempt counter
	data.Attempts++
	otpStore[email] = data

	// Check for too many attempts
	if data.Attempts > 3 {
		delete(otpStore, email)
		return false, errors.New("too many incorrect attempts")
	}

	// Validate OTP
	if data.OTP != otp {
		return false, errors.New("invalid OTP")
	}

	// OTP is valid, remove it from store
	delete(otpStore, email)
	return true, nil
}

// GenerateSessionToken generates a random session token with version salt
func GenerateSessionToken() (string, error) {
	// Generate random bytes
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Get app version for salting
	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "1.0.0.0" // Default version
	}

	// Add version salt to token
	h := sha256.New()
	h.Write(bytes)
	h.Write([]byte(version))

	return hex.EncodeToString(h.Sum(nil)), nil
}

// CreateSession creates a new session for the given user
func CreateSession(user *User) (Session, error) { // Uses User and Session from types.go
	token, err := GenerateSessionToken()
	if err != nil {
		return Session{}, err
	}

	session := Session{ // Uses Session from types.go
		Token:     token,
		User:      user,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hour session
	}

	// Store session
	sessionMutex.Lock()
	sessions[token] = session
	sessionMutex.Unlock()

	return session, nil
}

// GetSession retrieves a session by token
func GetSession(token string) (Session, bool) { // Uses Session from types.go
	sessionMutex.RLock()
	defer sessionMutex.RUnlock()

	session, found := sessions[token]
	if !found {
		return Session{}, false
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		delete(sessions, token)
		return Session{}, false
	}

	return session, true
}

// ClearSession removes a session by token
func ClearSession(token string) {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()
	delete(sessions, token)
}

// SetUserContext adds the user to the request context
func SetUserContext(ctx context.Context, user *User) context.Context { // Uses User from types.go
	return context.WithValue(ctx, userCtxKey{}, user)
}

// GetUserFromContext retrieves the user from the request context
func GetUserFromContext(ctx context.Context) *User { // Uses User from types.go
	userVal := ctx.Value(userCtxKey{})
	user, ok := userVal.(*User) // Type assertion to *User
	if !ok {
		// Handle case where value is not *User or nil
		log.Printf("Warning: User context value is not of type *User or is nil: %T", userVal)
		return nil
	}
	return user
}

// CleanupSessions removes expired sessions
func CleanupSessions() {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()

	now := time.Now()
	for token, session := range sessions {
		if now.After(session.ExpiresAt) {
			delete(sessions, token)
		}
	}
}

// generateOTP creates a random numeric OTP of the specified length
func generateOTP(length int) (string, error) {
	const digits = "0123456789"
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	for i := 0; i < length; i++ {
		b[i] = digits[int(b[i])%len(digits)]
	}
	return string(b), nil
}

// sendOTPEmail sends an email with the OTP
func sendOTPEmail(to, otp string) error {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	// Get SMTP settings from environment
	host := getEnvOrDefault("SMTP_HOST", "localhost")
	portStr := getEnvOrDefault("SMTP_PORT", "25")
	password := getEnvOrDefault("SMTP_PASSWORD", "")
	from := getEnvOrDefault("SMTP_FROM", "noreply@example.com")
	appName := getEnvOrDefault("APP_NAME", "OpenAgent")

	port, _ := strconv.Atoi(portStr)

	// Create message
	subject := fmt.Sprintf("Your %s login code", appName)
	body := fmt.Sprintf("Your verification code is: %s\nThis code will expire in 5 minutes.", otp)
	message := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", from, to, subject, body))

	// For Resend, the username is 'resend' and the password is the API key
	auth := smtp.PlainAuth("", "resend", password, host)

	// Send email
	err := smtp.SendMail(fmt.Sprintf("%s:%d", host, port), auth, from, []string{to}, message)
	if err != nil {
		return err
	}

	return nil
}

// getEnvOrDefault returns the value of an environment variable or a default value
func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetSessionCookieName returns the versioned session cookie name
func GetSessionCookieName() string {
	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "1.0.0.0" // Default version
	}
	// Use only major.minor for cookie name to reduce frequency of cookie invalidation on patch/build
	parts := strings.Split(version, ".")
	if len(parts) >= 2 {
		version = parts[0] + "_" + parts[1]
	}
	return "session_v" + version
}

// isPublicPath returns true for paths that don't require authentication
func isPublicPath(path string) bool {
	publicPaths := []string{
		"/login",
		"/auth/request-otp", // Changed from /api/request-otp
		"/auth/verify-otp",  // Changed from /api/verify-otp
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
		"/api/admin/", // Keep this if admin API exists
	}

	for _, ap := range adminPaths {
		if strings.HasPrefix(path, ap) {
			return true
		}
	}

	return false
}

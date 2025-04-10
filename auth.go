package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"sync"
	"time"
)

// OTP storage with expiration
type OTPStore struct {
	sync.RWMutex
	store map[string]*OTPData
}

type OTPData struct {
	Email     string
	OTP       string
	ExpiresAt time.Time
	Attempts  int
}

// User represents a user in the system
type User struct {
	ID           int
	Email        string
	IsAdmin      bool
	CreatedAt    time.Time
	LastLoggedIn sql.NullTime
}

// Global OTP store
var (
	otpStore    = &OTPStore{store: make(map[string]*OTPData)}
	otpValidity = 10 * time.Minute
	maxAttempts = 3
	otpLength   = 6
)

// GenerateOTP creates a random OTP
func GenerateOTP(length int) (string, error) {
	const otpChars = "0123456789"
	buffer := make([]byte, length)
	_, err := rand.Read(buffer)
	if err != nil {
		return "", err
	}

	otpCharsLength := len(otpChars)
	for i := 0; i < length; i++ {
		buffer[i] = otpChars[int(buffer[i])%otpCharsLength]
	}

	return string(buffer), nil
}

// SendOTP generates and sends an OTP to the user's email
func SendOTP(email string) error {
	// Generate OTP
	otp, err := GenerateOTP(otpLength)
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %v", err)
	}

	// Store OTP
	otpStore.Lock()
	otpStore.store[email] = &OTPData{
		Email:     email,
		OTP:       otp,
		ExpiresAt: time.Now().Add(otpValidity),
		Attempts:  0,
	}
	otpStore.Unlock()

	// Send email
	err = sendEmail(email, "Your Login OTP", fmt.Sprintf("Your OTP for login is: %s\nValid for %d minutes.", otp, otpValidity/time.Minute))
	if err != nil {
		return fmt.Errorf("failed to send OTP email: %v", err)
	}

	return nil
}

// VerifyOTP checks if the provided OTP is valid
func VerifyOTP(email, otp string) (bool, error) {
	otpStore.Lock()
	defer otpStore.Unlock()

	data, exists := otpStore.store[email]
	if !exists {
		return false, fmt.Errorf("no OTP found for this email")
	}

	// Check expiration
	if time.Now().After(data.ExpiresAt) {
		delete(otpStore.store, email)
		return false, fmt.Errorf("OTP has expired")
	}

	// Check attempts
	if data.Attempts >= maxAttempts {
		delete(otpStore.store, email)
		return false, fmt.Errorf("too many invalid attempts")
	}

	// Increment attempts
	data.Attempts++

	// Check OTP
	if data.OTP != otp {
		return false, fmt.Errorf("invalid OTP")
	}

	// Success - remove OTP
	delete(otpStore.store, email)
	return true, nil
}

// sendEmail sends an email using SMTP
func sendEmail(to, subject, body string) error {
	from := os.Getenv("SMTP_FROM")
	password := os.Getenv("SMTP_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")

	if from == "" || password == "" || smtpHost == "" || smtpPort == "" {
		return fmt.Errorf("email configuration is incomplete")
	}

	// Message composition
	message := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", from, to, subject, body))

	// Authentication
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Send email
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, message)
	if err != nil {
		return err
	}

	return nil
}

// generateSessionToken creates a random session token
func generateSessionToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// Sessions
var (
	sessions     = make(map[string]Session)
	sessionMutex sync.RWMutex
)

// Session represents a user session
type Session struct {
	UserID    int
	Email     string
	IsAdmin   bool
	ExpiresAt time.Time
}

// CreateSession creates a new session for a user
func CreateSession(user User) (string, error) {
	token, err := generateSessionToken()
	if err != nil {
		return "", err
	}

	sessionMutex.Lock()
	defer sessionMutex.Unlock()

	sessions[token] = Session{
		UserID:    user.ID,
		Email:     user.Email,
		IsAdmin:   user.IsAdmin,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	return token, nil
}

// GetSession retrieves a session by token
func GetSession(token string) (Session, bool) {
	sessionMutex.RLock()
	defer sessionMutex.RUnlock()

	session, exists := sessions[token]
	if !exists {
		return Session{}, false
	}

	if time.Now().After(session.ExpiresAt) {
		delete(sessions, token)
		return Session{}, false
	}

	return session, true
}

// AuthMiddleware is middleware to check authentication
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for certain paths
		path := r.URL.Path
		if path == "/" ||
			path == "/login" ||
			path == "/request-otp" ||
			path == "/verify-otp" ||
			strings.HasPrefix(path, "/static/") {
			next.ServeHTTP(w, r)
			return
		}

		// Check for session cookie
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Validate session
		session, valid := GetSession(cookie.Value)
		if !valid {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Admin-only routes
		if strings.HasPrefix(path, "/admin/") && !session.IsAdmin {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}

		// Continue to next handler
		next.ServeHTTP(w, r)
	})
}

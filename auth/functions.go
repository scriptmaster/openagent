package auth

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"net/smtp"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/scriptmaster/openagent/common"
	"golang.org/x/crypto/bcrypt"
)

// UserServicer moved to types.go
// User moved to types.go
// OTPData moved to types.go
// Session moved to types.go
// userCtxKey moved to types.go

var (
	otpStore = make(map[string]OTPData) // Uses OTPData from types.go
	otpMutex = &sync.Mutex{}
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

// GetSession and ClearSession are no longer needed with JWT
/*
func GetSession(token string) (Session, bool) { ... }
func ClearSession(token string) { ... }
*/

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
	host := getEnv("SMTP_HOST")
	portStr := getEnv("SMTP_PORT")
	password := getEnv("SMTP_PASSWORD")
	from := getEnv("SMTP_FROM")
	appName := getEnv("APP_NAME")

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

func getEnv(key string) string {
	return common.GetEnv(key)
}

// GetSessionCookieName returns the versioned session cookie name
func GetSessionCookieName() string {
	appName := getEnv("APP_NAME")
	version := getEnv("APP_VERSION")

	// Sanitize appName and version (replace spaces/invalid chars with underscore)
	re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	sanitizedAppName := re.ReplaceAllString(appName, "_")
	sanitizedVersion := re.ReplaceAllString(version, "_")

	// Construct the name
	cookieName := fmt.Sprintf("%s_%s", sanitizedAppName, sanitizedVersion)

	// Cookie names have restrictions, ensure it's valid (e.g., length, characters)
	// Basic check for length, could add more validation
	if len(cookieName) > 64 {
		cookieName = cookieName[:64] // Truncate if too long
	}

	return cookieName
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

// GeneratePasswordHash generates a bcrypt hash of the password
func GeneratePasswordHash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14) // 14 is the cost factor
	return string(bytes), err
}

// CheckPasswordHash compares a plain text password with a stored bcrypt hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

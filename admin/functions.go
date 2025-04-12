package admin

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// IsMaintenanceAuthenticated checks if the request has a valid maintenance authentication cookie
func IsMaintenanceAuthenticated(r *http.Request) bool {
	// Check for maintenance cookie
	cookie, err := r.Cookie("maintenance_auth")
	if err != nil {
		return false
	}

	// Expected value with current version salt
	expected := "authenticated_" + os.Getenv("APP_VERSION")[:8]

	// Validate cookie value matches current version
	return cookie.Value == expected
}

// IncrementBuildVersion increases the build number in APP_VERSION and updates the .env file
func IncrementBuildVersion() error {
	// Read current version from environment
	currentVersion := os.Getenv("APP_VERSION")
	if currentVersion == "" {
		currentVersion = "1.0.0.0" // Default if not set
	}

	// Parse version
	parts := strings.Split(currentVersion, ".")
	if len(parts) != 4 {
		// Invalid format, initialize to default
		parts = []string{"1", "0", "0", "0"}
	}

	// Increment build number (last part)
	buildNumber, err := strconv.Atoi(parts[3])
	if err != nil {
		buildNumber = 0 // Reset if parsing failed
	}
	buildNumber++
	parts[3] = strconv.Itoa(buildNumber)

	// Reassemble version string
	newVersion := strings.Join(parts, ".")

	// Load current .env content
	envPath := ".env"
	content, err := os.ReadFile(envPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to read .env file: %v", err)
		}
		// File doesn't exist, create it with just the version
		content = []byte("")
	}

	// Parse .env content line by line
	lines := strings.Split(string(content), "\n")
	versionLineFound := false
	for i, line := range lines {
		if strings.HasPrefix(line, "APP_VERSION=") {
			lines[i] = "APP_VERSION=" + newVersion
			versionLineFound = true
			break
		}
	}

	// If APP_VERSION line not found, add it
	if !versionLineFound {
		lines = append(lines, "APP_VERSION="+newVersion)
	}

	// Write back to .env
	err = os.WriteFile(envPath, []byte(strings.Join(lines, "\n")), 0644)
	if err != nil {
		return fmt.Errorf("failed to update .env file: %v", err)
	}

	// Update environment variable
	os.Setenv("APP_VERSION", newVersion)
	log.Printf("Application version updated to %s", newVersion)
	return nil
}

// GenerateSessionSalt creates a unique salt based on the app version
func GenerateSessionSalt(version string) string {
	h := sha256.New()
	h.Write([]byte(version))
	return fmt.Sprintf("%x", h.Sum(nil))
}

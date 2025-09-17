package admin

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/scriptmaster/openagent/common"
)

// IsMaintenanceAuthenticated checks if the request has a valid maintenance authentication cookie
func IsMaintenanceAuthenticated(r *http.Request) bool {
	// Check for maintenance cookie
	cookie, err := r.Cookie("maintenance_auth")
	if err != nil {
		return false
	}

	// Expected value with current version salt
	// Compare against first 8 chars of version for basic check
	currentVersion := common.GetEnv("APP_VERSION")
	if len(currentVersion) < 8 {
		currentVersion = "1.0.0.0_" // Default to avoid panic
	}
	expected := "authenticated_" + currentVersion[:8]

	// Validate cookie value matches current version prefix
	return cookie.Value == expected
}

// GetBuildNumber extracts the build number from the APP_VERSION env var.
func GetBuildNumber() int {
	currentVersion := common.GetEnv("APP_VERSION")
	if currentVersion == "" {
		return 0
	}
	parts := strings.Split(currentVersion, ".")
	if len(parts) != 4 {
		return 0
	}
	buildNumber, err := strconv.Atoi(parts[3])
	if err != nil {
		return 0
	}
	return buildNumber
}

func ReadFileAsString(file string) string {
	// Read the input file
	content, err := os.ReadFile(file)
	if err != nil {
		return fmt.Sprintf("failed to read input file: %v", err)
	}

	return string(content)
}

// UpdateEnvFile reads the .env file, updates or adds the provided key-value pairs,
// and writes the changes back to the file.
func UpdateEnvFile(updates map[string]string) error {
	envPath := ".env"
	content, err := os.ReadFile(envPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read .env file: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	newLines := make([]string, 0, len(lines))
	updatedKeys := make(map[string]bool)

	// Process existing lines and update values
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
			newLines = append(newLines, line) // Keep comments and empty lines
			continue
		}

		parts := strings.SplitN(trimmedLine, "=", 2)
		if len(parts) != 2 {
			newLines = append(newLines, line) // Keep malformed lines
			continue
		}

		key := strings.TrimSpace(parts[0])
		if newValue, shouldUpdate := updates[key]; shouldUpdate {
			newLines = append(newLines, key+"="+newValue)
			updatedKeys[key] = true
		} else {
			newLines = append(newLines, line) // Keep existing line
		}
	}

	// Add any keys from updates that were not found in the original file
	for key, value := range updates {
		if !updatedKeys[key] {
			newLines = append(newLines, key+"="+value)
		}
	}

	// Filter out potential trailing empty line if the original file didn't end with one
	outputContent := strings.Join(newLines, "\n")
	if len(content) > 0 && !strings.HasSuffix(string(content), "\n") && strings.HasSuffix(outputContent, "\n") {
		// Remove trailing newline if original didn't have one
		// (More robust handling might be needed depending on desired behavior)
		// outputContent = strings.TrimSuffix(outputContent, "\n")
	}

	err = os.WriteFile(envPath, []byte(outputContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to update .env file: %v", err)
	}

	// Update environment variables in the current process
	for key, value := range updates {
		os.Setenv(key, value)
	}

	log.Printf("Updated %d keys in .env file", len(updates))
	return nil
}

// GenerateSessionSalt creates a unique salt based on the app version
func GenerateSessionSalt(version string) string {
	h := sha256.New()
	h.Write([]byte(version))
	return fmt.Sprintf("%x", h.Sum(nil))
}

package common

import (
	"encoding/base64"
	"fmt"
	"regexp"
)

// Basic email validation regex (adjust as needed for stricter validation)
var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

// IsValidEmail checks if the provided string is a valid email format.
func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// DecodeConnectionString decodes a base64 encoded connection string.
// TODO: Implement proper error handling and potentially decryption if needed.
func DecodeConnectionString(encoded string) (string, error) {
	if encoded == "" {
		return "", fmt.Errorf("encoded connection string is empty")
	}
	dcodedBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 connection string: %w", err)
	}
	return string(dcodedBytes), nil
}

// EncodeConnectionString encodes a connection string using base64.
func EncodeConnectionString(connectionString string) string {
	return base64.StdEncoding.EncodeToString([]byte(connectionString))
}

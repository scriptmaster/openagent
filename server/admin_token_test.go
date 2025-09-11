package server

import (
	"regexp"
	"strings"
	"testing"
)

// TestGenerateAdminToken tests the admin token generation format
func TestGenerateAdminToken(t *testing.T) {
	token, err := generateAdminToken()
	if err != nil {
		t.Fatalf("Failed to generate admin token: %v", err)
	}

	// Check token format: UUID-ADMIN-TOKEN-SHA1-HASHED-TIMESTAMP
	// UUID format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx (36 chars)
	// SHA1 hash: 40 hex characters
	expectedPattern := `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}-ADMIN-TOKEN-[0-9a-f]{40}$`

	matched, err := regexp.MatchString(expectedPattern, token)
	if err != nil {
		t.Fatalf("Error matching token pattern: %v", err)
	}

	if !matched {
		t.Errorf("Token format is incorrect. Got: %s, Expected pattern: %s", token, expectedPattern)
	}

	// Check that token contains the expected parts
	parts := strings.Split(token, "-")
	if len(parts) < 3 {
		t.Errorf("Token should have at least 3 parts separated by '-', got %d parts", len(parts))
	}

	// Check that it contains "ADMIN-TOKEN"
	if !strings.Contains(token, "ADMIN-TOKEN") {
		t.Errorf("Token should contain 'ADMIN-TOKEN', got: %s", token)
	}

	// Check that the last part is a 40-character hex string (SHA1)
	lastPart := parts[len(parts)-1]
	if len(lastPart) != 40 {
		t.Errorf("SHA1 hash part should be 40 characters, got %d: %s", len(lastPart), lastPart)
	}

	// Test that multiple tokens are different
	token2, err := generateAdminToken()
	if err != nil {
		t.Fatalf("Failed to generate second admin token: %v", err)
	}

	if token == token2 {
		t.Errorf("Generated tokens should be different, got same token: %s", token)
	}
}

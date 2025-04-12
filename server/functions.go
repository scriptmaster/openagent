package server

import (
	"crypto/sha256"
	"encoding/hex"
	"html/template"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// GetTemplateFuncs returns the template functions map (Exported)
func GetTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"CurrentYear": func() int {
			return time.Now().Year()
		},
		"AssetPath": func(path string) string {
			// Basic implementation, may need adjustment based on actual asset handling
			return "/static/" + strings.TrimPrefix(path, "/")
		},
		"formatTime": func(t interface{}) string {
			if t == nil {
				return "Never"
			}
			switch v := t.(type) {
			case time.Time:
				return v.Format("Jan 02, 2006 15:04:05")
			case *time.Time:
				if v == nil {
					return "Never"
				}
				return v.Format("Jan 02, 2006 15:04:05")
			default:
				return "Invalid time format"
			}
		},
		"formatDate": func(t interface{}) string {
			if t == nil {
				return "Never"
			}
			switch v := t.(type) {
			case time.Time:
				return v.Format("Jan 02, 2006")
			case *time.Time:
				if v == nil {
					return "Never"
				}
				return v.Format("Jan 02, 2006")
			default:
				return "Invalid date format"
			}
		},
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"noescape": func(s string) template.HTML {
			return template.HTML(s)
		},
	}
}

// generateSessionSalt needs to be exported if called directly from outside the server package
// func generateSessionSalt() (string, error) {
// 	// ... implementation ...
// }

// parseVersion splits a version string into components with defaults
func parseVersion(version string) (major, minor, patch, build int) {
	versionParts := strings.Split(version, ".")
	if len(versionParts) != 4 {
		return 1, 0, 0, 0 // Default if not set properly
	}

	major, _ = strconv.Atoi(versionParts[0])
	minor, _ = strconv.Atoi(versionParts[1])
	patch, _ = strconv.Atoi(versionParts[2])
	build, _ = strconv.Atoi(versionParts[3])

	return major, minor, patch, build
}

// GetSessionSalt returns the session salt generated during route registration.
// It must be called after RegisterRoutes has been executed.
func GetSessionSalt() string {
	// Return the globally stored salt
	// Ensure sessionSalt is initialized before this is called (e.g., in RegisterRoutes)
	if sessionSalt == "" {
		// This shouldn't happen in the normal flow where RegisterRoutes is called first.
		// Maybe generate a temporary one or log a warning?
		log.Println("Warning: GetSessionSalt called before sessionSalt was initialized in RegisterRoutes.")
		// Fallback or panic might be appropriate depending on strictness needed.
		// For now, let's recalculate based on current env var as a fallback.
		version := os.Getenv("APP_VERSION")
		if version == "" {
			version = "1.0.0.0"
		}
		return generateSessionSalt(version) // Use the helper if available
	}
	return sessionSalt
}

// generateSessionSalt generates a salt based on the app version.
// This should ideally be defined once, maybe in functions.go or kept private here.
func generateSessionSalt(version string) string {
	// Simple salt generation (replace with more robust method if needed)
	h := sha256.New()
	h.Write([]byte(version))
	h.Write([]byte("-openagent-secret-salt-value")) // Add a static secret
	return hex.EncodeToString(h.Sum(nil))[:16]      // Use first 16 chars
}

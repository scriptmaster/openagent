package main

import (
	"crypto/sha256"
	"fmt"
	"html/template"
	"os"
	"strconv"
	"strings"
	"time"
)

// GetTemplateFuncs returns the template function map
func GetTemplateFuncs() template.FuncMap {
	return template.FuncMap{
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
	}
}

// getEnvOrDefault returns the value of environment variable or default value if not set
func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// generateSessionSalt creates a unique salt based on the app version
func generateSessionSalt(version string) string {
	h := sha256.New()
	h.Write([]byte(version))
	return fmt.Sprintf("%x", h.Sum(nil))
}

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

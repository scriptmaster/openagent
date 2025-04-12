package server

import (
	"html/template"
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

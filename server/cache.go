package server

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SetCacheHeaders sets appropriate cache headers based on the file path
func SetCacheHeaders(w http.ResponseWriter, r *http.Request, filePath string) {
	// Determine cache strategy based on file type
	if isStaticAsset(filePath) {
		setStaticAssetCacheHeaders(w, filePath)
	} else if isAPIPath(filePath) {
		setAPICacheHeaders(w)
	} else {
		setPageCacheHeaders(w)
	}
}

// isStaticAsset checks if the path is a static asset
func isStaticAsset(path string) bool {
	staticExtensions := []string{".css", ".js", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico", ".woff", ".woff2", ".ttf", ".eot"}
	path = strings.ToLower(path)

	for _, ext := range staticExtensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}

	// Check for static directory paths
	staticDirs := []string{"/static/", "/tsx/css/", "/tsx/js/"}
	for _, dir := range staticDirs {
		if strings.HasPrefix(path, dir) {
			return true
		}
	}

	return false
}

// isAPIPath checks if the path is an API endpoint
func isAPIPath(path string) bool {
	return strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/auth/")
}

// setStaticAssetCacheHeaders sets long-term cache headers for static assets
func setStaticAssetCacheHeaders(w http.ResponseWriter, filePath string) {
	// Long-term caching for static assets (1 year)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")

	// Set Expires header (1 year from now)
	expires := time.Now().AddDate(1, 0, 0)
	w.Header().Set("Expires", expires.UTC().Format(http.TimeFormat))

	// Generate and set ETag
	if etag, err := GenerateETag(filePath); err == nil {
		w.Header().Set("ETag", etag)
	}
}

// setAPICacheHeaders sets short-term cache headers for API responses
func setAPICacheHeaders(w http.ResponseWriter) {
	// Short-term caching for API responses (1 minute)
	w.Header().Set("Cache-Control", "private, max-age=60")

	// Generate ETag based on current time (for dynamic content)
	etag := fmt.Sprintf(`"%x"`, md5.Sum([]byte(time.Now().Format("2006-01-02 15:04:05"))))
	w.Header().Set("ETag", etag)
}

// setPageCacheHeaders sets medium-term cache headers for HTML pages
func setPageCacheHeaders(w http.ResponseWriter) {
	// Medium-term caching for HTML pages (5 minutes)
	w.Header().Set("Cache-Control", "public, max-age=300")

	// Generate ETag based on current time
	etag := fmt.Sprintf(`"%x"`, md5.Sum([]byte(time.Now().Format("2006-01-02 15:04"))))
	w.Header().Set("ETag", etag)
}

// GenerateETag generates an ETag for a file based on its content and modification time
func GenerateETag(filePath string) (string, error) {
	// For static files, try to get file info
	if strings.HasPrefix(filePath, "/static/") {
		// Convert URL path to file system path
		actualPath := strings.TrimPrefix(filePath, "/")
		actualPath = filepath.Join(".", actualPath)

		fileInfo, err := os.Stat(actualPath)
		if err != nil {
			// If file doesn't exist, generate ETag based on path
			return fmt.Sprintf(`"%x"`, md5.Sum([]byte(filePath))), nil
		}

		// Generate ETag based on file size, modification time, and path
		etagData := fmt.Sprintf("%s-%d-%d", filePath, fileInfo.Size(), fileInfo.ModTime().Unix())
		return fmt.Sprintf(`"%x"`, md5.Sum([]byte(etagData))), nil
	}

	// For non-static files, generate ETag based on path and current time
	etagData := fmt.Sprintf("%s-%d", filePath, time.Now().Unix()/60) // Change every minute
	return fmt.Sprintf(`"%x"`, md5.Sum([]byte(etagData))), nil
}

// HandleConditionalRequest handles If-None-Match conditional requests
func HandleConditionalRequest(w http.ResponseWriter, r *http.Request, etag string) bool {
	ifNoneMatch := r.Header.Get("If-None-Match")

	// If client has the same ETag, return 304 Not Modified
	if ifNoneMatch != "" && ifNoneMatch == etag {
		w.WriteHeader(http.StatusNotModified)
		return true
	}

	// Set ETag header for future requests
	w.Header().Set("ETag", etag)
	return false
}

// CacheMiddleware is a middleware that adds cache headers to responses
func CacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set cache headers based on the request path
		SetCacheHeaders(w, r, r.URL.Path)

		// Continue to the next handler
		next.ServeHTTP(w, r)
	})
}

// StaticFileHandler handles static files with proper cache headers
func StaticFileHandler(filePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Generate ETag for the file
		etag, err := GenerateETag(filePath)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Handle conditional request
		if HandleConditionalRequest(w, r, etag) {
			return // 304 Not Modified
		}

		// Set cache headers
		SetCacheHeaders(w, r, filePath)

		// Serve the file
		http.ServeFile(w, r, filePath)
	}
}

package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestCacheHeaders(t *testing.T) {
	tests := []struct {
		name           string
		filePath       string
		expectedCache  string
		expectedExpires bool
		expectedETag   bool
	}{
		{
			name:           "CSS file should have long-term cache",
			filePath:       "/static/css/tabler.min.css",
			expectedCache:  "public, max-age=31536000, immutable",
			expectedExpires: true,
			expectedETag:   true,
		},
		{
			name:           "JS file should have long-term cache",
			filePath:       "/static/js/tabler.min.js",
			expectedCache:  "public, max-age=31536000, immutable",
			expectedExpires: true,
			expectedETag:   true,
		},
		{
			name:           "Image file should have long-term cache",
			filePath:       "/static/img/logo.svg",
			expectedCache:  "public, max-age=31536000, immutable",
			expectedExpires: true,
			expectedETag:   true,
		},
		{
			name:           "HTML page should have short-term cache",
			filePath:       "/test",
			expectedCache:  "public, max-age=300",
			expectedExpires: false,
			expectedETag:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.filePath, nil)
			w := httptest.NewRecorder()

			// Create a test handler that sets cache headers
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				SetCacheHeaders(w, r, tt.filePath)
				w.WriteHeader(http.StatusOK)
			})

			handler.ServeHTTP(w, req)

			// Check Cache-Control header
			cacheControl := w.Header().Get("Cache-Control")
			if cacheControl != tt.expectedCache {
				t.Errorf("Expected Cache-Control %q, got %q", tt.expectedCache, cacheControl)
			}

			// Check Expires header
			expires := w.Header().Get("Expires")
			if tt.expectedExpires && expires == "" {
				t.Error("Expected Expires header to be set")
			}
			if !tt.expectedExpires && expires != "" {
				t.Error("Expected Expires header to not be set")
			}

			// Check ETag header
			etag := w.Header().Get("ETag")
			if tt.expectedETag && etag == "" {
				t.Error("Expected ETag header to be set")
			}
			if !tt.expectedETag && etag != "" {
				t.Error("Expected ETag header to not be set")
			}
		})
	}
}

func TestETagGeneration(t *testing.T) {
	// Create a temporary file for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	content := "test content for etag generation"
	
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test ETag generation
	etag, err := GenerateETag(testFile)
	if err != nil {
		t.Fatalf("Failed to generate ETag: %v", err)
	}

	if etag == "" {
		t.Error("Expected ETag to be generated")
	}

	// ETag should be consistent for the same content
	etag2, err := GenerateETag(testFile)
	if err != nil {
		t.Fatalf("Failed to generate second ETag: %v", err)
	}

	if etag != etag2 {
		t.Error("ETag should be consistent for the same file")
	}
}

func TestConditionalRequest(t *testing.T) {
	// Create a temporary file for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	content := "test content for conditional request"
	
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Generate ETag for the file
	etag, err := GenerateETag(testFile)
	if err != nil {
		t.Fatalf("Failed to generate ETag: %v", err)
	}

	tests := []struct {
		name           string
		ifNoneMatch    string
		expectedStatus int
		expectedBody   bool
	}{
		{
			name:           "No If-None-Match header should return 200",
			ifNoneMatch:    "",
			expectedStatus: http.StatusOK,
			expectedBody:   true,
		},
		{
			name:           "Matching If-None-Match should return 304",
			ifNoneMatch:    etag,
			expectedStatus: http.StatusNotModified,
			expectedBody:   false,
		},
		{
			name:           "Non-matching If-None-Match should return 200",
			ifNoneMatch:    `"different-etag"`,
			expectedStatus: http.StatusOK,
			expectedBody:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.ifNoneMatch != "" {
				req.Header.Set("If-None-Match", tt.ifNoneMatch)
			}

			w := httptest.NewRecorder()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Simulate serving a file with conditional request handling
				if HandleConditionalRequest(w, r, etag) {
					return // 304 Not Modified
				}
				
				SetCacheHeaders(w, r, "/test")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(content))
			})

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			body := w.Body.String()
			if tt.expectedBody && body == "" {
				t.Error("Expected response body to be present")
			}
			if !tt.expectedBody && body != "" {
				t.Error("Expected response body to be empty")
			}
		})
	}
}

func TestCacheHeadersForDifferentFileTypes(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/static/css/style.css", "public, max-age=31536000, immutable"},
		{"/static/js/script.js", "public, max-age=31536000, immutable"},
		{"/static/img/image.png", "public, max-age=31536000, immutable"},
		{"/static/fonts/font.woff2", "public, max-age=31536000, immutable"},
		{"/api/data", "private, max-age=60"},
		{"/dashboard", "public, max-age=300"},
		{"/test", "public, max-age=300"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				SetCacheHeaders(w, r, tt.path)
				w.WriteHeader(http.StatusOK)
			})

			handler.ServeHTTP(w, req)

			cacheControl := w.Header().Get("Cache-Control")
			if cacheControl != tt.expected {
				t.Errorf("For path %s, expected Cache-Control %q, got %q", tt.path, tt.expected, cacheControl)
			}
		})
	}
}

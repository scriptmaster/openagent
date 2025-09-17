package transpile

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// TestTranspileIntegration tests the complete transpile process end-to-end
func TestTranspileIntegration(t *testing.T) {
	// Set debug mode for detailed logging
	os.Setenv("DEBUG_TRANSPILE", "1")
	defer os.Unsetenv("DEBUG_TRANSPILE")

	// Clean up any existing generated files
	cleanupGeneratedFiles(t)

	// Step 1: Run the transpile process
	t.Log("Step 1: Running TranspileAllTemplates...")

	// Set debug mode for detailed logging
	os.Setenv("DEBUG_TRANSPILE", "1")

	// Change to project root directory for transpile process
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Change to project root (go up 2 levels from server/transpile)
	if err := os.Chdir("../../"); err != nil {
		t.Fatalf("Failed to change to project root: %v", err)
	}
	defer os.Chdir(originalDir) // Restore original directory

	if err := TranspileAllTemplates(); err != nil {
		t.Fatalf("TranspileAllTemplates failed: %v", err)
	}

	// Check if files were actually generated
	t.Log("Checking if files were generated...")
	if _, err := os.Stat("tpl/generated/css/pages_test.css"); err != nil {
		t.Logf("Warning: pages_test.css not found: %v", err)
	}
	if _, err := os.Stat("tpl/generated/js/pages_test.js"); err != nil {
		t.Logf("Warning: pages_test.js not found: %v", err)
	}

	// Step 2: Verify generated files exist
	t.Log("Step 2: Verifying generated files...")
	verifyGeneratedFiles(t)

	// Step 3: Verify CSS files have content
	t.Log("Step 3: Verifying CSS files have content...")
	verifyCSSFiles(t)

	// Step 4: Verify JS files have correct content
	t.Log("Step 4: Verifying JS files have correct content...")
	verifyJSFiles(t)

	// Step 5: Start server and test HTTP endpoints
	t.Log("Step 5: Testing HTTP endpoints...")
	testHTTPEndpoints(t)

	t.Log("✅ All integration tests passed!")
}

func cleanupGeneratedFiles(t *testing.T) {
	generatedDir := "./tpl/generated"
	if err := os.RemoveAll(generatedDir); err != nil {
		t.Logf("Warning: Could not clean generated directory: %v", err)
	}
	if err := os.MkdirAll(generatedDir, 0755); err != nil {
		t.Fatalf("Failed to create generated directory: %v", err)
	}
}

func verifyGeneratedFiles(t *testing.T) {
	// Check that consolidated CSS files exist
	expectedCSSFiles := []string{
		"tpl/generated/css/pages_test.css",
		"tpl/generated/css/app_test.css",
		"tpl/generated/css/admin_test.css",
	}

	for _, cssFile := range expectedCSSFiles {
		if _, err := os.Stat(cssFile); os.IsNotExist(err) {
			t.Errorf("Expected CSS file does not exist: %s", cssFile)
		}
	}

	// Check that JS files exist
	expectedJSFiles := []string{
		"tpl/generated/js/pages_test.js",
		"tpl/generated/js/_common.js",
	}

	for _, jsFile := range expectedJSFiles {
		if _, err := os.Stat(jsFile); os.IsNotExist(err) {
			t.Errorf("Expected JS file does not exist: %s", jsFile)
		}
	}

	// Check that TSX files exist
	expectedTSXFiles := []string{
		"tpl/generated/pages/test.tsx",
		"tpl/generated/pages/test.component.tsx",
	}

	for _, tsxFile := range expectedTSXFiles {
		if _, err := os.Stat(tsxFile); os.IsNotExist(err) {
			t.Errorf("Expected TSX file does not exist: %s", tsxFile)
		}
	}
}

func verifyCSSFiles(t *testing.T) {
	// Test that pages_test.css has content from test.html
	pagesCSSPath := "tpl/generated/css/pages_test.css"
	content, err := os.ReadFile(pagesCSSPath)
	if err != nil {
		t.Fatalf("Could not read pages_test.css: %v", err)
	}

	cssContent := string(content)
	if len(cssContent) == 0 {
		t.Error("pages_test.css is empty - CSS extraction failed")
	}

	// Check for specific CSS content from test.html
	expectedCSS := ":root"
	if !strings.Contains(cssContent, expectedCSS) {
		t.Errorf("pages_test.css missing expected CSS content. Expected to find '%s', got: %s", expectedCSS, cssContent)
	}

	// Check for CSS variable
	expectedVar := "--light-color"
	if !strings.Contains(cssContent, expectedVar) {
		t.Errorf("pages_test.css missing expected CSS variable. Expected to find '%s', got: %s", expectedVar, cssContent)
	}

	// Check for CSS value
	expectedValue := "#eee"
	if !strings.Contains(cssContent, expectedValue) {
		t.Errorf("pages_test.css missing expected CSS value. Expected to find '%s', got: %s", expectedValue, cssContent)
	}

	// Verify CSS file is not empty and has reasonable size
	if len(cssContent) < 10 {
		t.Errorf("pages_test.css too small (%d bytes), expected at least 10 bytes", len(cssContent))
	}

	t.Logf("✅ CSS extraction verified: pages_test.css contains %d bytes with expected content", len(cssContent))
}

func verifyJSFiles(t *testing.T) {
	// Test pages_test.js has correct content
	pagesJSPath := "tpl/generated/js/pages_test.js"
	content, err := os.ReadFile(pagesJSPath)
	if err != nil {
		t.Fatalf("Could not read pages_test.js: %v", err)
	}

	jsContent := string(content)

	// Check for main Test component
	if !strings.Contains(jsContent, "function Test({page})") {
		t.Error("pages_test.js missing Test component function")
	}

	// Check for Simple component embedding
	if !strings.Contains(jsContent, "function Simple({page})") {
		t.Error("pages_test.js missing Simple component embedding")
	}

	// Check for Simple component section header
	if !strings.Contains(jsContent, "🔧 SIMPLE COMPONENT JS 🔧") {
		t.Error("pages_test.js missing Simple component section header")
	}

	// Check for hydration code
	if !strings.Contains(jsContent, "window.hydrateReactApp") {
		t.Error("pages_test.js missing hydration code")
	}

	// Check for global component assignment
	if !strings.Contains(jsContent, "window.Test = Test") {
		t.Error("pages_test.js missing global component assignment")
	}

	// Test _common.js has correct content
	commonJSPath := "tpl/generated/js/_common.js"
	commonContent, err := os.ReadFile(commonJSPath)
	if err != nil {
		t.Fatalf("Could not read _common.js: %v", err)
	}

	commonJSContent := string(commonContent)

	// Check for hydrate function definition
	if !strings.Contains(commonJSContent, "window.hydrateReactApp") {
		t.Error("_common.js missing window.hydrateReactApp function definition")
	}

	// Check for React production includes (not development)
	if strings.Contains(commonJSContent, "react.development.js") {
		t.Error("_common.js should not include react.development.js")
	}
	if strings.Contains(commonJSContent, "react-dom.development.js") {
		t.Error("_common.js should not include react-dom.development.js")
	}

	// Check for React production includes
	if !strings.Contains(commonJSContent, "react.production.min.js") {
		t.Error("_common.js missing react.production.min.js")
	}
	if !strings.Contains(commonJSContent, "react-dom.production.min.js") {
		t.Error("_common.js missing react-dom.production.min.js")
	}
}

func testHTTPEndpoints(t *testing.T) {
	// Start server in background
	serverCmd := startTestServer(t)
	defer func() {
		if serverCmd != nil {
			serverCmd.Kill()
		}
	}()

	// Wait for server to start
	time.Sleep(3 * time.Second)

	// Test /test endpoint returns 200
	testURL := "http://localhost:8800/test"
	resp, err := http.Get(testURL)
	if err != nil {
		t.Fatalf("Failed to GET /test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected /test to return 200, got %d", resp.StatusCode)
	}

	// Read response body to check for correct paths
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	htmlContent := string(body)

	// Check for correct CSS link paths
	expectedCSSPath := "/tsx/css/pages_test.css"
	if !strings.Contains(htmlContent, expectedCSSPath) {
		t.Errorf("HTML missing expected CSS path. Expected to find '%s' in: %s", expectedCSSPath, htmlContent)
	}

	// Check for correct JS script paths
	expectedJSPath := "/tsx/js/pages_test.js"
	if !strings.Contains(htmlContent, expectedJSPath) {
		t.Errorf("HTML missing expected JS path. Expected to find '%s' in: %s", expectedJSPath, htmlContent)
	}

	// Test CSS file is accessible
	cssURL := "http://localhost:8800/tsx/css/pages_test.css"
	cssResp, err := http.Get(cssURL)
	if err != nil {
		t.Fatalf("Failed to GET CSS file: %v", err)
	}
	defer cssResp.Body.Close()

	if cssResp.StatusCode != http.StatusOK {
		t.Errorf("Expected CSS file to return 200, got %d", cssResp.StatusCode)
	}

	// Test JS file is accessible
	jsURL := "http://localhost:8800/tsx/js/pages_test.js"
	jsResp, err := http.Get(jsURL)
	if err != nil {
		t.Fatalf("Failed to GET JS file: %v", err)
	}
	defer jsResp.Body.Close()

	if jsResp.StatusCode != http.StatusOK {
		t.Errorf("Expected JS file to return 200, got %d", jsResp.StatusCode)
	}

	// Test _common.js is accessible
	commonURL := "http://localhost:8800/tsx/js/_common.js"
	commonResp, err := http.Get(commonURL)
	if err != nil {
		t.Fatalf("Failed to GET _common.js: %v", err)
	}
	defer commonResp.Body.Close()

	if commonResp.StatusCode != http.StatusOK {
		t.Errorf("Expected _common.js to return 200, got %d", commonResp.StatusCode)
	}
}

func startTestServer(t *testing.T) *os.Process {
	// This is a simplified version - in a real implementation,
	// you might want to use a more sophisticated server startup
	// For now, we'll assume the server is already running or can be started
	// by the test environment

	// Check if server is already running
	resp, err := http.Get("http://localhost:8800/")
	if err == nil && resp.StatusCode == http.StatusOK {
		resp.Body.Close()
		t.Log("Server already running on port 8800")
		return nil
	}

	t.Log("Note: Server startup test requires manual server start or test environment setup")
	return nil
}

// TestTranspileRegression tests for specific regressions
func TestTranspileRegression(t *testing.T) {
	// Test that Simple component embedding works
	os.Setenv("DEBUG_TRANSPILE", "1")
	defer os.Unsetenv("DEBUG_TRANSPILE")

	// Clean and run transpile
	cleanupGeneratedFiles(t)
	if err := TranspileAllTemplates(); err != nil {
		t.Fatalf("TranspileAllTemplates failed: %v", err)
	}

	// Check that Simple component is embedded in pages_test.js
	pagesJSPath := "tpl/generated/js/pages_test.js"
	content, err := os.ReadFile(pagesJSPath)
	if err != nil {
		t.Fatalf("Could not read pages_test.js: %v", err)
	}

	jsContent := string(content)

	// Regression test: Simple component should be embedded
	if !strings.Contains(jsContent, "🔧 SIMPLE COMPONENT JS 🔧") {
		t.Error("REGRESSION: Simple component embedding is broken")
	}

	// Regression test: CSS should be extracted
	pagesCSSPath := "tpl/generated/css/pages_test.css"
	cssContent, err := os.ReadFile(pagesCSSPath)
	if err != nil {
		t.Fatalf("Could not read pages_test.css: %v", err)
	}

	if len(cssContent) == 0 {
		t.Error("REGRESSION: CSS extraction is broken - pages_test.css is empty")
	}

	// Regression test: Link paths should use consolidated names
	tsxPath := "tpl/generated/pages/test.tsx"
	tsxContent, err := os.ReadFile(tsxPath)
	if err != nil {
		t.Fatalf("Could not read test.tsx: %v", err)
	}

	tsxString := string(tsxContent)
	if !strings.Contains(tsxString, "/tsx/css/pages_test.css") {
		t.Error("REGRESSION: Link paths not using consolidated CSS names")
	}
}

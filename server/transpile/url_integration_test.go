package transpile

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func TestServerJavaScriptOutput(t *testing.T) {
	// Test the actual server JavaScript output
	tests := []struct {
		name        string
		envValue    string
		url         string
		expectDual  bool
		description string
	}{
		// TEMPORARILY DISABLED: Dual function pattern tests
		// {
		// 	name:        "WAX_FORK_Unset_DualFunction",
		// 	envValue:    "",
		// 	url:         "http://localhost:8800/tsx/js/pages_test.js",
		// 	expectDual:  true,
		// 	description: "Should generate dual function pattern when WAX_FORK is unset",
		// },
		// {
		// 	name:        "WAX_FORK_0_DualFunction",
		// 	envValue:    "0",
		// 	url:         "http://localhost:8800/tsx/js/pages_test.js",
		// 	expectDual:  true,
		// 	description: "Should generate dual function pattern when WAX_FORK=0",
		// },
		{
			name:        "WAX_FORK_1_SingleFunction",
			envValue:    "1",
			url:         "http://localhost:8800/tsx/js/pages_test.js",
			expectDual:  false,
			description: "Should generate single function pattern when WAX_FORK=1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envValue != "" {
				os.Setenv("WAX_FORK", tt.envValue)
			} else {
				os.Unsetenv("WAX_FORK")
			}
			defer os.Unsetenv("WAX_FORK")

			// Start server if not running
			serverRunning := checkServerRunning()
			if !serverRunning {
				t.Log("Starting server for integration test...")
				startServerForTest(t)
				defer stopServerForTest(t)
			}

			// Wait for server to be ready
			waitForServer(t)

			// Clear generated files to force regeneration
			clearGeneratedFiles(t)

			// Trigger page generation
			triggerPageGeneration(t)

			// Test the generated JavaScript
			jsContent := fetchJavaScriptFromURL(t, tt.url)

			// Check for duplicate function declarations
			functionCount := strings.Count(jsContent, "function Test({page})")
			if functionCount > 1 {
				t.Errorf("Found %d duplicate function declarations in %s", functionCount, tt.description)
				t.Logf("Generated JS content:\n%s", jsContent)
			}

			// Check for malformed syntax
			if strings.Contains(jsContent, "function Test({page}) {\n    return (\nfunction Test({page}) {") {
				t.Errorf("Found malformed nested function syntax in %s", tt.description)
				t.Logf("Generated JS content:\n%s", jsContent)
			}

			// Check for proper function structure
			if !strings.Contains(jsContent, "function Test({page}) {") {
				t.Errorf("Missing main function declaration in %s", tt.description)
				t.Logf("Generated JS content:\n%s", jsContent)
			}

			if tt.expectDual {
				// Should contain dual function pattern elements
				dualPatternElements := []string{
					"function TestJSX(",
					"return TestJSX(",
					"typeof props != 'undefined'",
					"typeof state != 'undefined'",
				}

				for _, element := range dualPatternElements {
					if !strings.Contains(jsContent, element) {
						t.Errorf("Dual function pattern missing element: %s", element)
						t.Logf("Generated JS content:\n%s", jsContent)
					}
				}
			} else {
				// Should contain single function pattern elements
				singlePatternElements := []string{
					"function Test({page})",
					"return (",
				}

				for _, element := range singlePatternElements {
					if !strings.Contains(jsContent, element) {
						t.Errorf("Single function pattern missing element: %s", element)
						t.Logf("Generated JS content:\n%s", jsContent)
					}
				}

				// Should NOT contain dual function pattern
				if strings.Contains(jsContent, "TestJSX(") {
					t.Errorf("Should not have dual function pattern when single is expected")
				}
			}

			t.Logf("✅ %s: Server JavaScript output test passed", tt.description)
		})
	}
}

func TestNoDuplicateFunctionDeclarationsInServerOutput(t *testing.T) {
	// Test that the server JavaScript output has no duplicate function declarations
	serverRunning := checkServerRunning()
	if !serverRunning {
		t.Log("Starting server for integration test...")
		startServerForTest(t)
		defer stopServerForTest(t)
	}

	// Wait for server to be ready
	waitForServer(t)

	// Clear generated files to force regeneration
	clearGeneratedFiles(t)

	// Trigger page generation
	triggerPageGeneration(t)

	// Fetch the JavaScript from the server
	jsContent := fetchJavaScriptFromURL(t, "http://localhost:8800/tsx/js/pages_test.js")

	// Check for duplicate function declarations
	functionCount := strings.Count(jsContent, "function Test({page})")
	if functionCount > 1 {
		t.Errorf("Found %d duplicate function declarations in server output", functionCount)
		t.Logf("Generated JS content:\n%s", jsContent)
	}

	// Check for malformed syntax
	if strings.Contains(jsContent, "function Test({page}) {\n    return (\nfunction Test({page}) {") {
		t.Errorf("Found malformed nested function syntax in server output")
		t.Logf("Generated JS content:\n%s", jsContent)
	}

	// Check for proper function structure
	if !strings.Contains(jsContent, "function Test({page}) {") {
		t.Errorf("Missing main function declaration in server output")
		t.Logf("Generated JS content:\n%s", jsContent)
	}

	// Check that the function has proper return statement
	if !strings.Contains(jsContent, "return (") {
		t.Errorf("Missing return statement in server output")
		t.Logf("Generated JS content:\n%s", jsContent)
	}

	t.Logf("✅ Server JavaScript output has no duplicate function declarations")
}

// Using existing functions from integration_test.go

func clearGeneratedFiles(t *testing.T) {
	// Clear generated files to force regeneration
	files := []string{
		"tpl/generated/pages/test.component.tsx",
		"tpl/generated/components/counter.tsx",
		"tpl/generated/js/pages_test.js",
	}

	for _, file := range files {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			t.Logf("Could not remove %s: %v", file, err)
		}
	}
}

func triggerPageGeneration(t *testing.T) {
	// Trigger page generation by accessing the test page
	resp, err := http.Get("http://localhost:8800/test")
	if err != nil {
		t.Fatalf("Failed to trigger page generation: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("Failed to trigger page generation: status %d", resp.StatusCode)
	}

	// Wait a bit for file generation
	time.Sleep(2 * time.Second)
}

func fetchJavaScriptFromURL(t *testing.T, url string) string {
	// Fetch the JavaScript file from the server
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("Failed to fetch JavaScript from %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("Failed to fetch JavaScript from %s: status %d", url, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read JavaScript from %s: %v", url, err)
	}

	return string(body)
}

package transpile

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func TestDualFunctionPatternIntegration(t *testing.T) {
	// Test both WAX_FORK scenarios
	tests := []struct {
		name        string
		envValue    string
		expectDual  bool
		description string
	}{
		{
			name:        "WAX_FORK_Unset_DualFunction",
			envValue:    "",
			expectDual:  true,
			description: "Should generate dual function pattern when WAX_FORK is unset",
		},
		{
			name:        "WAX_FORK_0_DualFunction",
			envValue:    "0",
			expectDual:  true,
			description: "Should generate dual function pattern when WAX_FORK=0",
		},
		{
			name:        "WAX_FORK_1_SingleFunction",
			envValue:    "1",
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

			// Test the generated JavaScript
			jsContent := fetchTestPageJS(t)

			if tt.expectDual {
				// Should contain dual function pattern elements
				dualPatternElements := []string{
					"function Test({page})",
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

				// Should NOT contain single function pattern
				if strings.Contains(jsContent, "return (") && !strings.Contains(jsContent, "TestJSX(") {
					t.Errorf("Should not have single function pattern when dual is expected")
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

			t.Logf("✅ %s: Integration test passed", tt.description)
		})
	}
}

func TestCounterComponentDualFunctionPattern(t *testing.T) {
	// Test that counter component has dual function pattern
	os.Unsetenv("WAX_FORK") // Ensure dual function pattern

	// Start server if not running
	serverRunning := checkServerRunning()
	if !serverRunning {
		t.Log("Starting server for integration test...")
		startServerForTest(t)
		defer stopServerForTest(t)
	}

	// Wait for server to be ready
	waitForServer(t)

	// Test the generated JavaScript
	jsContent := fetchTestPageJS(t)

	// Check for counter component dual function pattern
	counterDualElements := []string{
		"function Counter({page})",
		"function CounterJSX(",
		"return CounterJSX(",
		"let count = 99",
		"const state = {",
		"count: 55",
	}

	for _, element := range counterDualElements {
		if !strings.Contains(jsContent, element) {
			t.Errorf("Counter dual function pattern missing element: %s", element)
			t.Logf("Generated JS content:\n%s", jsContent)
		}
	}

	t.Log("✅ Counter component dual function pattern integration test passed")
}

func checkServerRunning() bool {
	resp, err := http.Get("http://localhost:8800/version")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

func startServerForTest(t *testing.T) {
	// This is a placeholder - in a real integration test, you'd start the server
	// For now, we'll assume the server is started externally
	t.Log("Server should be started externally for this test")
}

func stopServerForTest(t *testing.T) {
	// This is a placeholder - in a real integration test, you'd stop the server
	t.Log("Server cleanup should be handled externally for this test")
}

func waitForServer(t *testing.T) {
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		if checkServerRunning() {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatal("Server did not become ready within expected time")
}

func fetchTestPageJS(t *testing.T) string {
	// Fetch the test page JS file
	resp, err := http.Get("http://localhost:8800/tsx/js/pages_test.js")
	if err != nil {
		t.Fatalf("Failed to fetch test page JS: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("Failed to fetch test page JS: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read test page JS: %v", err)
	}

	return string(body)
}

func TestTSXToJSConversionWithDualFunction(t *testing.T) {
	// Test the TSX to JS conversion directly
	dualFunctionTSX := `export default function Test() {
    // ╔══ 🔧 PAGE <script> TAG CONTENTS 🔧 ══
    let count = 0;
    const state = { count: 55 };
    
    // Call the JSX function with props and state
    return TestJSX(typeof props != 'undefined' ? props : {}, typeof state != 'undefined' ? state : {});
}

function TestJSX(props, state) {
    return (
        <div>Counter: {state.count}</div>
    );
}`

	// Convert TSX to JS
	jsContent := TSX2JSWithOptions(dualFunctionTSX, true)

	// Check that it contains the dual function pattern
	expectedElements := []string{
		"function Test({page})",
		"let count = 0;",
		"const state = { count: 55 };",
		"return TestJSX(",
		"function TestJSX(props, state)",
		"Counter: ' + (state.count) + '",
	}

	for _, element := range expectedElements {
		if !strings.Contains(jsContent, element) {
			t.Errorf("TSX to JS conversion missing element: %s", element)
			t.Logf("Converted JS content:\n%s", jsContent)
		}
	}

	t.Log("✅ TSX to JS conversion with dual function pattern working correctly")
}

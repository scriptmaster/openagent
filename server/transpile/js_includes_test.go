package transpile

import (
	"os"
	"strings"
	"testing"
)

func TestProcessJSIncludes(t *testing.T) {
	// Create a temporary test file with includes
	testContent := `//#include "/static/js/react.production.min.js"
//#include "/static/js/react-dom.production.min.js"

// Test content
console.log('test');`

	// Create a temporary React file for testing
	reactContent := `// React library content
window.React = { createElement: function() {} };`

	// Create tmp directory if it doesn't exist
	if err := os.MkdirAll("./tmp", 0755); err != nil {
		t.Fatalf("Failed to create tmp directory: %v", err)
	}

	// Write temporary files
	if err := os.WriteFile("./tmp/test_react.js", []byte(reactContent), 0644); err != nil {
		t.Fatalf("Failed to create test React file: %v", err)
	}
	defer os.Remove("./tmp/test_react.js")

	// Test with relative path
	testContentWithPath := strings.Replace(testContent, "/static/js/react.production.min.js", "./tmp/test_react.js", 1)

	result := processJSIncludes(testContentWithPath)

	// Check if the include was processed
	if !strings.Contains(result, "React library content") {
		t.Errorf("Include was not processed. Result: %s", result)
	}

	// Check if the include directive was removed
	if strings.Contains(result, "//#include") {
		t.Errorf("Include directive was not removed. Result: %s", result)
	}

	t.Log("✅ JS includes processing test passed")
}

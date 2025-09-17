package transpile

import (
	"os"
	"strings"
	"testing"
)

// TestInnerComponentEmbedding tests that inner components are processed and embedded
func TestInnerComponentEmbedding(t *testing.T) {
	// Set up test environment
	defer cleanupTestFiles(t)

	// Create test component files
	createTestComponentFiles(t)

	tests := []struct {
		name        string
		htmlContent string
		baseName    string
		contains    []string
		notContains []string
		description string
	}{
		{
			name: "Page with Simple component",
			htmlContent: `<main>
				<div className="container">
					<h1>Test Page</h1>
					<div id="component-simple"></div>
				</div>
				<script>
					function initPage() {
						console.log('Page initialized');
					}
				</script>
			</main>`,
			baseName: "test",
			contains: []string{
				// Main component JS
				"React.createElement('main'",
				"React.createElement('div', {className: \"container\"}",
				"React.createElement('h1', null, 'Test Page')",
				"React.createElement(Simple, {suppressHydrationWarning: true}",

				// Inner component JS (Simple component)
				"// ╔══════════════════════════════════════════════════════════════════════════════",
				"// ║                    🔧 SIMPLE COMPONENT JS 🔧",
				"// ╚══════════════════════════════════════════════════════════════════════════════",
				"React.createElement('div', {className: \"simple-component\"}",
				"React.createElement('button', {onClick: handleClick}",

				// Simple component prototype methods
				"Simple.prototype.handleClick = function()",
				"Simple.prototype.init = function()",

				// Original JS content
				"function initPage()",

				// Hydration
				"window.test = test",
				"hydrateReactApp",
			},
			notContains: []string{
				"<script>",
			},
			description: "Page with Simple component should embed Simple component JS",
		},
		{
			name: "Page with multiple components",
			htmlContent: `<main>
				<div className="container">
					<h1>Multi Component Page</h1>
					<div id="component-simple"></div>
					<div id="component-counter"></div>
				</div>
				<script>
					function initPage() {
						console.log('Page initialized');
					}
				</script>
			</main>`,
			baseName: "test",
			contains: []string{
				// Main component JS
				"React.createElement('main'",
				"React.createElement(Simple, {suppressHydrationWarning: true}",
				"React.createElement(Counter, {suppressHydrationWarning: true}",

				// Simple component JS
				"// ╔══════════════════════════════════════════════════════════════════════════════",
				"// ║                    🔧 SIMPLE COMPONENT JS 🔧",
				"// ╚══════════════════════════════════════════════════════════════════════════════",
				"Simple.prototype.handleClick = function()",

				// Counter component JS
				"// ╔══════════════════════════════════════════════════════════════════════════════",
				"// ║                    🔧 COUNTER COMPONENT JS 🔧",
				"// ╚══════════════════════════════════════════════════════════════════════════════",
				"Counter.prototype.increment = function()",
				"Counter.prototype.decrement = function()",

				// Original JS content
				"function initPage()",

				// Hydration
				"window.test = test",
				"hydrateReactApp",
			},
			notContains: []string{
				"<script>",
			},
			description: "Page with multiple components should embed all component JS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test HTML file
			inputPath := "tpl/pages/test.html"
			outputPath := "tpl/generated/pages/test.tsx"

			if err := os.WriteFile(inputPath, []byte(tt.htmlContent), 0644); err != nil {
				t.Fatalf("Failed to create test HTML file: %v", err)
			}

			// Run transpilation
			err := TranspileHtmlToTsx(inputPath, outputPath)
			if err != nil {
				t.Fatalf("TranspileHtmlToTsx failed: %v", err)
			}

			// Read generated JS file
			jsPath := "tpl/generated/js/pages_test.js"
			jsContent, err := os.ReadFile(jsPath)
			if err != nil {
				t.Fatalf("Failed to read JS file: %v", err)
			}

			// Test JS content
			jsStr := string(jsContent)
			for _, expected := range tt.contains {
				if !strings.Contains(jsStr, expected) {
					t.Errorf("JS content does not contain expected string: %v", expected)
					t.Logf("JS content: %s", jsStr)
				}
			}

			for _, unwanted := range tt.notContains {
				if strings.Contains(jsStr, unwanted) {
					t.Errorf("JS content should not contain: %v", unwanted)
					t.Logf("JS content: %s", jsStr)
				}
			}

			t.Logf("✅ %s: %s", tt.name, tt.description)
		})
	}
}

// createTestComponentFiles creates test component files
func createTestComponentFiles(t *testing.T) {
	// Create components directory
	if err := os.MkdirAll("tpl/components", 0755); err != nil {
		t.Fatalf("Failed to create components directory: %v", err)
	}

	// Simple component
	simpleHTML := `<div className="simple-component">
		<h2>Simple Component</h2>
		<button onClick={handleClick}>Click me</button>
	</div>
	<script>
		Simple.prototype.handleClick = function() {
			console.log('Simple button clicked');
		};
		
		Simple.prototype.init = function() {
			this.active = true;
		};
	</script>`

	if err := os.WriteFile("tpl/components/simple.html", []byte(simpleHTML), 0644); err != nil {
		t.Fatalf("Failed to create simple.html: %v", err)
	}

	// Counter component
	counterHTML := `<div className="counter-component">
		<h2>Counter: {this.value}</h2>
		<button onClick={increment}>+</button>
		<button onClick={decrement}>-</button>
	</div>
	<script>
		Counter.prototype.increment = function() {
			this.value++;
		};
		
		Counter.prototype.decrement = function() {
			this.value--;
		};
		
		Counter.prototype.init = function() {
			this.value = 0;
		};
	</script>`

	if err := os.WriteFile("tpl/components/counter.html", []byte(counterHTML), 0644); err != nil {
		t.Fatalf("Failed to create counter.html: %v", err)
	}
}

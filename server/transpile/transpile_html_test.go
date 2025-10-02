package transpile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestTranspileHtmlToTsx tests the main transpilation function with progressive complexity
func TestTranspileHtmlToTsx(t *testing.T) {
	// Set up test environment
	os.Setenv("DEBUG_TRANSPILE", "1")
	defer os.Unsetenv("DEBUG_TRANSPILE")

	// Clean up any existing test files
	cleanupTestFiles(t)

	tests := []struct {
		name        string
		htmlContent string
		baseName    string
		contains    []string
		notContains []string
		description string
	}{
		{
			name: "Simple div with no script",
			htmlContent: `<main>
				<div className="container">
					<h1>Hello World</h1>
				</div>
			</main>`,
			baseName: "test",
			contains: []string{
				"React.createElement('div', {className: \"container\"}",
				"React.createElement('h1', null, 'Hello World')",
			},
			notContains: []string{
				"<script>",
				"prototype",
				"window.Test",
			},
			description: "Basic HTML structure without any JavaScript",
		},
		{
			name: "Div with script",
			htmlContent: `<main>
				<div className="container">
					<h1>Hello World</h1>
				</div>
				<script>
					function handleClick() {
						console.log('Button clicked!');
					}
				</script>
			</main>`,
			baseName: "test",
			contains: []string{
				"function handleClick()",
				"console.log('Button clicked!')",
			},
			notContains: []string{
				"<script>",
				"prototype",
			},
			description: "HTML with basic JavaScript function",
		},
		{
			name: "Div with 3-4 children and script",
			htmlContent: `<main>
				<div className="container">
					<header className="header">
						<h1>Page Title</h1>
					</header>
					<nav className="navigation">
						<ul>
							<li><a href="/home">Home</a></li>
							<li><a href="/about">About</a></li>
						</ul>
					</nav>
					<main className="content">
						<p>Main content here</p>
					</main>
				</div>
				<script>
					function initPage() {
						console.log('Page initialized');
					}
					
					function handleNavigation() {
						console.log('Navigation clicked');
					}
				</script>
			</main>`,
			baseName: "test",
			contains: []string{
				"React.createElement('div', {className: \"container\"}",
				"React.createElement('header', {className: \"header\"}",
				"React.createElement('nav', {className: \"navigation\"}",
				"React.createElement('main', {className: \"content\"}",
				"function initPage()",
				"function handleNavigation()",
				"window.Test",
				"hydrateReactApp",
			},
			notContains: []string{
				"<script>",
				"prototype",
			},
			description: "Complex HTML structure with multiple children and JavaScript functions",
		},
		{
			name: "Div with custom component and script",
			htmlContent: `<main>
				<div className="container">
					<h1>Page with Component</h1>
					<div id="component-simple"></div>
					<p>Content after component</p>
				</div>
				<script>
					function handleComponent() {
						console.log('Component loaded');
					}
					
					// Simple component prototype
					Simple.prototype.init = function() {
						this.value = 0;
					};
				</script>
			</main>`,
			baseName: "test",
			contains: []string{
				"React.createElement('div', {className: \"container\"}",
				"React.createElement('h1', null, 'Page with Component')",
				"React.createElement('p', null, 'Content after component')",
				"function handleComponent()",
				"Simple.prototype.init",
				"window.Test",
				"hydrateReactApp",
			},
			notContains: []string{
				"<script>",
			},
			description: "HTML with custom component reference and prototype method",
		},
		{
			name: "Div with custom component and siblings after",
			htmlContent: `<main>
				<div className="container">
					<h1>Page Title</h1>
					<div id="component-simple"></div>
					<div className="sibling-after">
						<p>This comes after the component</p>
						<button onClick={handleClick}>Click me</button>
					</div>
				</div>
				<script>
					function handleClick() {
						console.log('Button clicked');
					}
					
					// Simple component prototype
					Simple.prototype.render = function() {
						return '<div>Simple component</div>';
					};
					
					// Another function after component
					function afterComponent() {
						console.log('After component function');
					}
				</script>
			</main>`,
			baseName: "test",
			contains: []string{
				"React.createElement('div', {className: \"container\"}",
				"React.createElement('h1', null, 'Page Title')",
				"React.createElement('div', {className: \"sibling-after\"}",
				"React.createElement('p', null, 'This comes after the component')",
				"React.createElement('button', {onClick: handleClick}, 'Click me')",
				"function handleClick()",
				"Simple.prototype.render",
				"function afterComponent()",
				"window.Test",
				"hydrateReactApp",
			},
			notContains: []string{
				"<script>",
			},
			description: "HTML with custom component and content after it",
		},
		{
			name: "Div with 2 custom components and scripts",
			htmlContent: `<main>
				<div className="container">
					<h1>Multi-Component Page</h1>
					<div id="component-simple"></div>
					<div id="component-counter"></div>
					<p>Content between components</p>
				</div>
				<script>
					function initPage() {
						console.log('Page with multiple components');
					}
					
					// Simple component prototype
					Simple.prototype.init = function() {
						this.active = true;
					};
					
					// Counter component prototype
					Counter.prototype.increment = function() {
						this.count++;
					};
					
					Counter.prototype.decrement = function() {
						this.count--;
					};
				</script>
			</main>`,
			baseName: "test",
			contains: []string{
				"React.createElement('div', {className: \"container\"}",
				"React.createElement('h1', null, 'Multi-Component Page')",
				"React.createElement('p', null, 'Content between components')",
				"function initPage()",
				"Simple.prototype.init",
				"Counter.prototype.increment",
				"Counter.prototype.decrement",
				"window.Test",
				"hydrateReactApp",
			},
			notContains: []string{
				"<script>",
			},
			description: "HTML with multiple custom components and their prototype methods",
		},
		{
			name: "Complex structure with multiple components and scripts",
			htmlContent: `<main>
				<div className="app">
					<header className="app-header">
						<h1>Complex App</h1>
						<template id="component-simple"></template>
					</header>
					<main className="app-main">
						<div className="sidebar">
							<template id="component-counter"></template>
							<nav>
								<ul>
									<li><a href="/dashboard">Dashboard</a></li>
									<li><a href="/settings">Settings</a></li>
								</ul>
							</nav>
						</div>
						<div className="content">
							<template id="component-chart"></template>
							<section>
								<h2>Main Content</h2>
								<p>This is the main content area with multiple components.</p>
							</section>
						</div>
					</main>
					<footer className="app-footer">
						<template id="component-footer"></template>
						<p>&copy; 2024 My App</p>
					</footer>
				</div>
				<script>
					// Main app initialization
					function initApp() {
						console.log('Complex app initialized');
						setupEventListeners();
					}
					
					function setupEventListeners() {
						console.log('Event listeners setup');
					}
					
					// Simple component prototype
					Simple.prototype.init = function() {
						this.visible = true;
					};
					
					Simple.prototype.toggle = function() {
						this.visible = !this.visible;
					};
					
					// Counter component prototype
					Counter.prototype.init = function() {
						this.value = 0;
						this.max = 100;
					};
					
					Counter.prototype.increment = function() {
						if (this.value < this.max) {
							this.value++;
						}
					};
					
					Counter.prototype.decrement = function() {
						if (this.value > 0) {
							this.value--;
						}
					};
					
					// Chart component prototype
					Chart.prototype.init = function() {
						this.data = [];
					};
					
					Chart.prototype.addData = function(point) {
						this.data.push(point);
					};
					
					// Footer component prototype
					Footer.prototype.init = function() {
						this.year = new Date().getFullYear();
					};
				</script>
			</main>`,
			baseName: "test",
			contains: []string{
				"React.createElement('div', {className: \"app\"}",
				"React.createElement('header', {className: \"app-header\"}",
				"React.createElement('main', {className: \"app-main\"}",
				"React.createElement(Footer, {suppressHydrationWarning: true}",
				"React.createElement('div', {className: \"sidebar\"}",
				"React.createElement('div', {className: \"content\"}",
				"React.createElement('nav', null",
				"React.createElement('section', null",
				"function initApp()",
				"function setupEventListeners()",
				"Simple.prototype.init",
				"Simple.prototype.toggle",
				"Counter.prototype.init",
				"Counter.prototype.increment",
				"Counter.prototype.decrement",
				"Chart.prototype.init",
				"Chart.prototype.addData",
				"Footer.prototype.init",
				"window.Test",
				"hydrateReactApp",
			},
			notContains: []string{
				"<script>",
			},
			description: "Complex HTML structure with multiple components and comprehensive prototype methods",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary input file
			inputPath := "tpl/pages/test.html"
			if err := os.MkdirAll("tpl/pages", 0755); err != nil {
				t.Fatalf("Failed to create test directory: %v", err)
			}

			if err := os.WriteFile(inputPath, []byte(tt.htmlContent), 0644); err != nil {
				t.Fatalf("Failed to write test input file: %v", err)
			}
			defer os.Remove(inputPath)

			// Create output path
			outputPath := "tpl/generated/pages/test.tsx"

			// Call the function under test
			err := TranspileHtmlToTsx(inputPath, outputPath)
			if err != nil {
				t.Fatalf("TranspileHtmlToTsx() error = %v", err)
			}

			// Check that the component TSX file was created
			componentPath := "tpl/generated/pages/test.component.tsx"
			componentContent, err := os.ReadFile(componentPath)
			if err != nil {
				t.Fatalf("Failed to read component file: %v", err)
			}
			defer os.Remove(componentPath)

			// Check that the JS file was created (if there was script content)
			jsPath := "tpl/generated/js/pages_test.js"
			jsContent, err := os.ReadFile(jsPath)
			if err != nil && strings.Contains(tt.htmlContent, "<script>") {
				t.Fatalf("Failed to read JS file: %v", err)
			}
			// Don't remove JS file so we can inspect it
			// defer os.Remove(jsPath)

			// Test component TSX content (should contain JSX syntax, not React.createElement calls)
			componentStr := string(componentContent)
			tsxContains := []string{
				"<div className=",
			}
			tsxNotContains := []string{
				"<script>",
				"function handleClick()",
				"React.createElement",
			}

			for _, expected := range tsxContains {
				if !strings.Contains(componentStr, expected) {
					t.Errorf("Component TSX does not contain expected string: %v", expected)
					t.Logf("Component content: %s", componentStr)
				}
			}

			for _, unwanted := range tsxNotContains {
				if strings.Contains(componentStr, unwanted) {
					t.Errorf("Component TSX should not contain: %v", unwanted)
					t.Logf("Component content: %s", componentStr)
				}
			}

			// Test JS content if it exists (should contain JS functions and React code)
			if len(jsContent) > 0 {
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
			}

			t.Logf("✅ %s: %s", tt.name, tt.description)
		})
	}
}

// TestTranspileHtmlToTsx_EdgeCases tests edge cases and error conditions
func TestTranspileHtmlToTsx_EdgeCases(t *testing.T) {
	// Set up test environment
	os.Setenv("DEBUG_TRANSPILE", "1")
	defer os.Unsetenv("DEBUG_TRANSPILE")

	// Clean up any existing test files
	cleanupTestFiles(t)

	tests := []struct {
		name        string
		htmlContent string
		baseName    string
		expectError bool
		description string
	}{
		{
			name:        "Empty HTML content",
			htmlContent: "",
			baseName:    "test",
			expectError: false,
			description: "Should handle empty HTML gracefully",
		},
		{
			name: "Only script tags",
			htmlContent: `<script>
				function test() {
					console.log('Only script');
				}
			</script>`,
			baseName:    "test",
			expectError: false,
			description: "Should handle HTML with only script tags",
		},
		{
			name: "Only style tags",
			htmlContent: `<style>
				.container { padding: 20px; }
			</style>`,
			baseName:    "test",
			expectError: false,
			description: "Should handle HTML with only style tags",
		},
		{
			name: "Malformed HTML",
			htmlContent: `<main>
				<div className="container">
					<h1>Unclosed tag
					<p>Another unclosed tag
				</div>
			</main>`,
			baseName:    "test",
			expectError: false,
			description: "Should handle malformed HTML gracefully",
		},
		{
			name: "Complex nested scripts",
			htmlContent: `<main>
				<div className="container">
					<h1>Complex Scripts</h1>
				</div>
				<script>
					// Multiple functions
					function func1() { return 'test1'; }
					function func2() { return 'test2'; }
					
					// Object with methods
					const obj = {
						method1: function() { return 'method1'; },
						method2: function() { return 'method2'; }
					};
					
					// Prototype methods
					Test.prototype.init = function() { this.value = 0; };
					Test.prototype.increment = function() { this.value++; };
				</script>
			</main>`,
			baseName:    "test",
			expectError: false,
			description: "Should handle complex JavaScript with multiple patterns",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary input file
			inputPath := "tpl/pages/test.html"
			if err := os.MkdirAll("tpl/pages", 0755); err != nil {
				t.Fatalf("Failed to create test directory: %v", err)
			}

			if err := os.WriteFile(inputPath, []byte(tt.htmlContent), 0644); err != nil {
				t.Fatalf("Failed to write test input file: %v", err)
			}
			defer os.Remove(inputPath)

			// Create output path
			outputPath := "tpl/generated/pages/test.tsx"

			// Call the function under test
			err := TranspileHtmlToTsx(inputPath, outputPath)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Clean up generated files
			os.Remove("tpl/generated/pages/test.component.tsx")
			os.Remove("tpl/generated/js/pages_test.js")

			t.Logf("✅ %s: %s", tt.name, tt.description)
		})
	}
}

// TestTranspileHtmlToTsx_FourStepProcess tests the 4-step inclusion process specifically
func TestTranspileHtmlToTsx_FourStepProcess(t *testing.T) {
	// Set up test environment
	os.Setenv("DEBUG_TRANSPILE", "1")
	defer os.Unsetenv("DEBUG_TRANSPILE")

	// Clean up any existing test files
	cleanupTestFiles(t)

	htmlContent := `<main>
		<div className="container">
			<h1>Four Step Process Test</h1>
			<div id="component-simple"></div>
			<div id="component-counter"></div>
		</div>
		<script>
			// Step 1: Basic JavaScript functions
			function initPage() {
				console.log('Page initialized');
			}
			
			// Step 2: Component prototype methods
			Simple.prototype.init = function() {
				this.active = true;
			};
			
			Counter.prototype.increment = function() {
				this.value++;
			};
			
			// Step 3: Complex JavaScript logic
			const app = {
				config: { debug: true },
				start: function() {
					console.log('App started');
				}
			};
		</script>
	</main>`

	// Create temporary input file
	inputPath := "tpl/pages/test.html"
	if err := os.MkdirAll("tpl/pages", 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	if err := os.WriteFile(inputPath, []byte(htmlContent), 0644); err != nil {
		t.Fatalf("Failed to write test input file: %v", err)
	}
	defer os.Remove(inputPath)

	// Create output path
	outputPath := "tpl/generated/pages/test.tsx"

	// Call the function under test
	err := TranspileHtmlToTsx(inputPath, outputPath)
	if err != nil {
		t.Fatalf("TranspileHtmlToTsx() error = %v", err)
	}

	// Read the generated JS file
	jsPath := "tpl/generated/js/pages_test.js"
	jsContent, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("Failed to read JS file: %v", err)
	}
	defer os.Remove(jsPath)

	jsStr := string(jsContent)

	// Test Step 1: Extract CSS and JS from HTML content
	t.Run("Step1_ExtractCSSAndJS", func(t *testing.T) {
		// Should contain the original JavaScript functions
		contains := []string{
			"function initPage()",
			"Simple.prototype.init",
			"Counter.prototype.increment",
			"const app = {",
		}

		for _, expected := range contains {
			if !strings.Contains(jsStr, expected) {
				t.Errorf("Step 1: JS content does not contain expected string: %v", expected)
			}
		}
	})

	// Test Step 2: Convert TSX to JS (React.createElement calls)
	t.Run("Step2_ConvertTSXToJS", func(t *testing.T) {
		// Should contain React.createElement calls
		contains := []string{
			"React.createElement('div', {className: \"container\"}",
			"React.createElement('h1', null, 'Four Step Process Test')",
		}

		for _, expected := range contains {
			if !strings.Contains(jsStr, expected) {
				t.Errorf("Step 2: JS content does not contain React.createElement: %v", expected)
			}
		}
	})

	// Test Step 3: Embed component JS (prototype methods)
	t.Run("Step3_EmbedComponentJS", func(t *testing.T) {
		// Should contain component prototype methods
		contains := []string{
			"Simple.prototype.init",
			"Counter.prototype.increment",
		}

		for _, expected := range contains {
			if !strings.Contains(jsStr, expected) {
				t.Errorf("Step 3: JS content does not contain component prototype: %v", expected)
			}
		}
	})

	// Test Step 4: Add hydration code
	t.Run("Step4_AddHydrationCode", func(t *testing.T) {
		// Should contain hydration code
		contains := []string{
			"window.Test",
			"hydrateReactApp",
			"React.createElement",
		}

		for _, expected := range contains {
			if !strings.Contains(jsStr, expected) {
				t.Errorf("Step 4: JS content does not contain hydration code: %v", expected)
			}
		}
	})

	// Test that script tags are removed from final output
	t.Run("ScriptTagsRemoved", func(t *testing.T) {
		notContains := []string{
			"<script>",
			"</script>",
		}

		for _, unwanted := range notContains {
			if strings.Contains(jsStr, unwanted) {
				t.Errorf("JS content should not contain script tags: %v", unwanted)
			}
		}
	})

	t.Logf("✅ Four Step Process Test: All steps working correctly")
}

// TestCSSFileGeneration tests that CSS files are always generated, even when empty
func TestCSSFileGeneration(t *testing.T) {
	// Set up test environment
	os.Setenv("DEBUG_TRANSPILE", "1")
	defer os.Unsetenv("DEBUG_TRANSPILE")

	// Clean up any existing test files
	cleanupTestFiles(t)

	tests := []struct {
		name        string
		htmlContent string
		description string
	}{
		{
			name: "HTML_with_no_styles",
			htmlContent: `<main>
				<div className="container">
					<h1>No Styles Test</h1>
				</div>
			</main>`,
			description: "HTML with no <style> tags should still generate empty CSS file",
		},
		{
			name: "HTML_with_empty_style_tag",
			htmlContent: `<main>
				<div className="container">
					<h1>Empty Style Test</h1>
				</div>
				<style></style>
			</main>`,
			description: "HTML with empty <style> tag should generate empty CSS file",
		},
		{
			name: "HTML_with_actual_styles",
			htmlContent: `<main>
				<div className="container">
					<h1>With Styles Test</h1>
				</div>
				<style>
					.container { padding: 20px; }
					h1 { color: blue; }
				</style>
			</main>`,
			description: "HTML with actual styles should generate CSS file with content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test environment
			inputPath, outputPath := setupTestEnvironment(t, tt.htmlContent)
			defer os.Remove(inputPath)

			// Call the function under test
			err := TranspileHtmlToTsx(inputPath, outputPath)
			if err != nil {
				t.Fatalf("TranspileHtmlToTsx() error = %v", err)
			}

			// Check that CSS file was created in the correct location (./tpl/generated/css/)
			// The function creates pages_filename.css for pages directory
			filename := filepath.Base(outputPath)
			filename = strings.TrimSuffix(filename, ".tsx")
			cssPath := "tpl/generated/css/pages_" + filename + ".css"
			if _, err := os.Stat(cssPath); os.IsNotExist(err) {
				t.Errorf("CSS file was not created: %s", cssPath)
			} else {
				t.Logf("✅ CSS file created: %s", cssPath)
			}

			// Read and verify CSS content
			cssContent, err := os.ReadFile(cssPath)
			if err != nil {
				t.Errorf("Failed to read CSS file: %v", err)
			}

			cssStr := string(cssContent)
			t.Logf("CSS content (%d bytes): %q", len(cssStr), cssStr)

			// Clean up
			os.Remove(cssPath)
			os.Remove(outputPath)
			os.Remove(strings.Replace(outputPath, ".tsx", ".component.tsx", 1))
		})
	}
}

// setupTestEnvironment creates necessary directories and files for testing
func setupTestEnvironment(t *testing.T, htmlContent string) (string, string) {
	// Create directories
	if err := os.MkdirAll("tpl/pages", 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	if err := os.MkdirAll("tpl/generated/pages", 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}
	if err := os.MkdirAll("tpl/generated/js", 0755); err != nil {
		t.Fatalf("Failed to create JS output directory: %v", err)
	}

	// Create input file
	inputPath := "tpl/pages/test.html"
	if err := os.WriteFile(inputPath, []byte(htmlContent), 0644); err != nil {
		t.Fatalf("Failed to write test input file: %v", err)
	}

	// Create output path
	outputPath := "tpl/generated/pages/test.tsx"

	return inputPath, outputPath
}

// cleanupTestFiles removes any test files that might exist
func cleanupTestFiles(t *testing.T) {
	files := []string{
		"tpl/pages/test.html",
		"tpl/generated/pages/test.tsx",
		"tpl/generated/pages/test.component.tsx",
		"tpl/generated/js/pages_test.js",
		// Clean up component files that might interfere with tests
		//"tpl/components/simple.html",
		//"tpl/components/counter.html",
		"tpl/generated/components/simple.tsx",
		"tpl/generated/components/counter.tsx",
		"tpl/generated/js/component_simple.js",
		"tpl/generated/js/component_counter.js",
	}

	for _, file := range files {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: Could not remove test file %s: %v", file, err)
		}
	}
}

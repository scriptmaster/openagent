package transpile

import (
	"os"
	"strings"
	"testing"
)

// TestTranspileLayoutToTsx tests the layout transpilation functionality
func TestTranspileLayoutToTsx(t *testing.T) {
	// Setup test environment
	setupLayoutTestEnvironment(t)
	defer cleanupLayoutTestFiles(t)

	tests := []struct {
		name        string
		layoutHTML  string
		expectedTSX []string
		notExpected []string
		description string
	}{
		{
			name: "Layout with includes",
			layoutHTML: `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8" />
    <title>{page.PageTitle}</title>
    <!--#include "../_partials/head.html" -->
</head>
<body>
    <div className="page">
        <!--#include "../_partials/header.html" -->
        <main>
            {children}
        </main>
        <!--#include "../_partials/footer.html" -->
    </div>
</body>
</html>`,
			expectedTSX: []string{
				"export default function Test_layout({page, children, linkPaths, scriptPaths}: {page: any, children?: any, linkPaths: any, scriptPaths: any})",
				"<html lang=\"en\">",
				"<head>",
				"<meta charset=\"UTF-8\" />",
				"<title>{page.PageTitle}</title>",
				"<link rel=\"shortcut icon\"",
				"<link rel=\"stylesheet\" href=\"/static/css/tabler.min.css\"",
				"<body>",
				"<div className=\"page\">",
				"<header class=\"navbar",
				"<main>",
				"{children}",
				"</main>",
				"<footer class=\"footer",
			},
			notExpected: []string{
				"<!--#include",
				"<!DOCTYPE html>",
				"<!--",
			},
			description: "Layout with include directives should process includes and remove comments",
		},
		{
			name: "Layout with Go template syntax",
			layoutHTML: `<!DOCTYPE html>
<html>
<head>
    <title>{{.PageTitle}} - {{.AppName}}</title>
</head>
<body>
    <div className="container">
        <h1>{{.PageTitle}}</h1>
        <p>Welcome to {{.AppName}}</p>
        {children}
    </div>
</body>
</html>`,
			expectedTSX: []string{
				"export default function Test_layout({page, children, linkPaths, scriptPaths}: {page: any, children?: any, linkPaths: any, scriptPaths: any})",
				"<html>",
				"<head>",
				"<title>{page.PageTitle} - {page.AppName}</title>",
				"<body>",
				"<div className=\"container\">",
				"<h1>{page.PageTitle}</h1>",
				"<p>Welcome to {page.AppName}</p>",
				"{children}",
			},
			notExpected: []string{
				"{{.PageTitle}}",
				"{{.AppName}}",
				"<!DOCTYPE html>",
			},
			description: "Layout should convert Go template syntax to JSX",
		},
		{
			name: "Layout with nested includes",
			layoutHTML: `<!DOCTYPE html>
<html>
<head>
    <title>Test Layout</title>
    <!--#include "../_partials/head.html" -->
</head>
<body>
    <div className="page">
        <!--#include "../_partials/header.html" -->
        <main>
            {children}
        </main>
    </div>
</body>
</html>`,
			expectedTSX: []string{
				"export default function Test_layout({page, children, linkPaths, scriptPaths}: {page: any, children?: any, linkPaths: any, scriptPaths: any})",
				"<html>",
				"<head>",
				"<title>Test Layout</title>",
				"<link rel=\"shortcut icon\"",
				"<link rel=\"stylesheet\"",
				"<body>",
				"<div className=\"page\">",
				"<header class=\"navbar",
				"<main>",
				"{children}",
			},
			notExpected: []string{
				"<!--#include",
				"<!DOCTYPE html>",
				"<!--",
			},
			description: "Layout should process nested includes correctly",
		},
		{
			name: "Layout with complex structure",
			layoutHTML: `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>{{.PageTitle}} - {{.AppName}}</title>
    <!--#include "../_partials/head.html" -->
</head>
<body class="theme-pista">
    <div className="page">
        <div className="page-wrapper">
            <!--#include "../_partials/header.html" -->
            <div className="page-body">
                <div className="container-xl">
                    <div className="row row-cards">
                        <div className="col-12">
                            {children}
                        </div>
                    </div>
                </div>
            </div>
            <!--#include "../_partials/footer.html" -->
        </div>
    </div>
</body>
</html>`,
			expectedTSX: []string{
				"export default function Test_layout({page, children, linkPaths, scriptPaths}: {page: any, children?: any, linkPaths: any, scriptPaths: any})",
				"<html lang=\"en\">",
				"<head>",
				"<meta charset=\"UTF-8\" />",
				"<meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\" />",
				"<title>{page.PageTitle} - {page.AppName}</title>",
				"<link rel=\"shortcut icon\"",
				"<link rel=\"stylesheet\"",
				"<body class=\"theme-pista\">",
				"<div className=\"page\">",
				"<div className=\"page-wrapper\">",
				"<header class=\"navbar",
				"<div className=\"page-body\">",
				"<div className=\"container-xl\">",
				"<div className=\"row row-cards\">",
				"<div className=\"col-12\">",
				"{children}",
				"<footer class=\"footer",
			},
			notExpected: []string{
				"<!--#include",
				"<!DOCTYPE html>",
				"{{.PageTitle}}",
				"{{.AppName}}",
			},
			description: "Complex layout should process all includes and template syntax",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test layout file
			layoutPath := "tpl/layouts/test_layout.html"
			if err := os.WriteFile(layoutPath, []byte(tt.layoutHTML), 0644); err != nil {
				t.Fatalf("Failed to create test layout file: %v", err)
			}

			// Transpile layout
			outputPath := "tpl/generated/layouts/test_layout.tsx"
			err := TranspileLayoutToTsx(layoutPath, outputPath)
			if err != nil {
				t.Fatalf("TranspileLayoutToTsx failed: %v", err)
			}

			// Read generated TSX
			tsxContent, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatalf("Failed to read generated TSX file: %v", err)
			}

			tsxString := string(tsxContent)

			// Debug: Print the actual TSX content
			t.Logf("Generated TSX content:\n%s", tsxString)

			// Check expected content
			for _, expected := range tt.expectedTSX {
				if !strings.Contains(tsxString, expected) {
					t.Errorf("TSX content missing expected string: %s", expected)
				}
			}

			// Check not expected content
			for _, notExpected := range tt.notExpected {
				if strings.Contains(tsxString, notExpected) {
					t.Errorf("TSX content should not contain: %s", notExpected)
				}
			}

			t.Logf("✅ %s: %s", tt.name, tt.description)
		})
	}
}

// TestProcessIncludes tests the processIncludes function directly
func TestProcessIncludes(t *testing.T) {
	// Setup test environment
	setupLayoutTestEnvironment(t)
	defer cleanupLayoutTestFiles(t)

	tests := []struct {
		name        string
		htmlContent string
		basePath    string
		expected    []string
		notExpected []string
		description string
	}{
		{
			name: "Simple include",
			htmlContent: `<head>
    <title>Test</title>
    <!--#include "../_partials/head.html" -->
</head>`,
			basePath: "tpl/layouts/test.html",
			expected: []string{
				"<link rel=\"shortcut icon\"",
				"<link rel=\"stylesheet\"",
			},
			notExpected: []string{
				"<!--#include",
			},
			description: "Should process simple include directive",
		},
		{
			name: "Multiple includes",
			htmlContent: `<body>
    <!--#include "../_partials/header.html" -->
    <main>Content</main>
    <!--#include "../_partials/footer.html" -->
</body>`,
			basePath: "tpl/layouts/test.html",
			expected: []string{
				"<header class=\"navbar",
				"<main>Content</main>",
				"<footer class=\"footer",
			},
			notExpected: []string{
				"<!--#include",
			},
			description: "Should process multiple include directives",
		},
		{
			name: "Include with comments",
			htmlContent: `<head>
    <title>Test</title>
    <!--#include "../_partials/head.html" -->
    <!-- This is a regular comment -->
</head>`,
			basePath: "tpl/layouts/test.html",
			expected: []string{
				"<link rel=\"shortcut icon\"",
				"<link rel=\"stylesheet\"",
			},
			notExpected: []string{
				"<!--#include",
			},
			description: "Should process includes (comments removed separately)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processIncludes(tt.htmlContent, tt.basePath)

			// Check expected content
			for _, expected := range tt.expected {
				if !strings.Contains(result, expected) {
					t.Errorf("Processed content missing expected string: %s", expected)
				}
			}

			// Check not expected content
			for _, notExpected := range tt.notExpected {
				if strings.Contains(result, notExpected) {
					t.Errorf("Processed content should not contain: %s", notExpected)
				}
			}

			t.Logf("✅ %s: %s", tt.name, tt.description)
		})
	}
}

// TestRemoveHTMLComments tests the removeHTMLComments function
func TestRemoveHTMLComments(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		description string
	}{
		{
			name:        "Simple comment",
			input:       `<div>Hello <!-- This is a comment --> World</div>`,
			expected:    `<div>Hello  World</div>`,
			description: "Should remove simple HTML comment",
		},
		{
			name:        "Multiple comments",
			input:       `<div>Hello <!-- Comment 1 --> World <!-- Comment 2 --> Test</div>`,
			expected:    `<div>Hello  World  Test</div>`,
			description: "Should remove multiple HTML comments",
		},
		{
			name:        "Multiline comment",
			input:       `<div>Hello <!--\nThis is a\nmultiline comment\n--> World</div>`,
			expected:    `<div>Hello  World</div>`,
			description: "Should remove multiline HTML comment",
		},
		{
			name:        "No comments",
			input:       `<div>Hello World</div>`,
			expected:    `<div>Hello World</div>`,
			description: "Should leave content unchanged when no comments",
		},
		{
			name:        "Include directive",
			input:       `<div>Hello <!--#include "file.html" --> World</div>`,
			expected:    `<div>Hello  World</div>`,
			description: "Should remove include directive comments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeHTMLComments(tt.input)
			if result != tt.expected {
				t.Errorf("Expected: %s\nGot: %s", tt.expected, result)
			}
			t.Logf("✅ %s: %s", tt.name, tt.description)
		})
	}
}

// setupLayoutTestEnvironment creates necessary directories and files for layout tests
func setupLayoutTestEnvironment(t *testing.T) {
	// Create necessary directories
	dirs := []string{
		"tpl/layouts",
		"tpl/_partials",
		"tpl/generated/layouts",
		"tpl/generated/js",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create test partial files
	headContent := `<!-- Favicon -->
<link rel="shortcut icon" type="image/x-icon" href="/static/favicon.ico" />
<link rel="icon" type="image/x-icon" href="/static/favicon.ico" />
<link rel="icon" type="image/svg+xml" href="/static/img/icon.svg" />
<link rel="icon" href="/static/favicon.ico" />

<!-- Tabler CSS -->
<link rel="stylesheet" href="/static/css/tabler.min.css" />
<!-- Tabler Icons -->
<link rel="stylesheet" href="/static/css/tabler-icons.min.css" />
<!-- Custom Styles -->
<link rel="stylesheet" href="/static/css/custom.css" />`

	headerContent := `<header class="navbar navbar-expand-md navbar-light d-print-none">
    <div class="container-xl">
        <button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbar-menu">
            <span class="navbar-toggler-icon"></span>
        </button>
        <h1 class="navbar-brand navbar-brand-autodark d-none-navbar-horizontal pe-0 pe-md-3">
            <a href="/">
                <img src="/static/img/logo.svg" height="36" alt="{page.AppName}">
            </a>
        </h1>
    </div>
</header>`

	footerContent := `<footer class="footer footer-transparent d-print-none">
    <div class="container-xl">
        <div class="row text-center align-items-center flex-row-reverse">
            <div class="col-lg-auto ms-lg-auto">
                <ul class="list-inline list-inline-dots mb-0">
                    <li class="list-inline-item"><a href="/about" class="link-secondary">About</a></li>
                    <li class="list-inline-item"><a href="/contact" class="link-secondary">Contact</a></li>
                    <li class="list-inline-item"><a href="/privacy" class="link-secondary">Privacy</a></li>
                </ul>
            </div>
        </div>
    </div>
</footer>`

	partialFiles := map[string]string{
		"tpl/_partials/head.html":   headContent,
		"tpl/_partials/header.html": headerContent,
		"tpl/_partials/footer.html": footerContent,
	}

	for filePath, content := range partialFiles {
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create partial file %s: %v", filePath, err)
		}
	}
}

// cleanupLayoutTestFiles removes test files created during layout tests
func cleanupLayoutTestFiles(t *testing.T) {
	files := []string{
		"tpl/layouts/test_layout.html",
		"tpl/generated/layouts/test_layout.tsx",
		"tpl/generated/layouts/test_layout.js",
		"tpl/generated/js/test_layout.js",
	}

	for _, file := range files {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: Could not remove test file %s: %v", file, err)
		}
	}
}

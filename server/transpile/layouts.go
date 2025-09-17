package transpile

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// ============================================================================
// LAYOUT TRANSPILATION FUNCTIONS
// ============================================================================
// This file contains functions for transpiling layout HTML files to TSX/JS
//
// Function List:
// - TranspileLayoutToTsx(inputPath, outputPath string) error
//   Main function to transpile layout HTML files to TSX
// - processLayoutContent(htmlContent string) string
//   Processes layout-specific content and structure
// - extractLayoutComponents(htmlContent string) []string
//   Extracts component references from layout files
// - buildLayoutWrapper(componentName, content string) string
//   Builds the layout wrapper structure
// ============================================================================

// TranspileLayoutToTsx converts a layout HTML file to a TSX layout component
func TranspileLayoutToTsx(inputPath, outputPath string) error {
	if isDebugTranspile() {
		fmt.Printf("DEBUG: Transpiling layout %s to %s\n", inputPath, outputPath)
	}

	// Read the input file
	content, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read layout file: %v", err)
	}

	// Convert Go template syntax to JSX
	htmlContent := string(content)

	// Process includes first (before other processing)
	htmlContent = processIncludes(htmlContent, inputPath)

	// Process layout-specific content
	htmlContent = processLayoutContent(htmlContent)

	// Integrate linkPaths and scriptPaths into the HTML content
	htmlContent = integrateLinkPathsAndScriptPaths(htmlContent)

	// Fix self-closing tags for JSX compatibility
	htmlContent = fixSelfClosingTags(htmlContent)

	// Extract base name for component naming
	baseName := filepath.Base(inputPath)
	baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName))
	componentName := convertToCamelCase(baseName)

	// Extract component references
	components := extractLayoutComponents(htmlContent)

	// Generate imports for components
	var imports string
	if len(components) > 0 {
		var importsBuilder strings.Builder
		importsBuilder.WriteString("// Component imports\n")
		for _, componentName := range components {
			importsBuilder.WriteString(fmt.Sprintf("import %s from '../components/%s';\n", componentName, strings.ToLower(componentName)))
		}
		importsBuilder.WriteString("\n")
		imports = importsBuilder.String()
	}

	// Create the layout component
	tsxContent := imports + `export default function ` + componentName + `({page, children, linkPaths, scriptPaths}: {page: any, children?: any, linkPaths: any, scriptPaths: any}) {
    return (
` + htmlContent + `
    );
}`

	// Write TSX file
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create layout output directory: %v", err)
	}

	if err := os.WriteFile(outputPath, []byte(tsxContent), 0644); err != nil {
		return fmt.Errorf("failed to write layout TSX file: %v", err)
	}

	if isDebugTranspile() {
		fmt.Printf("DEBUG: Wrote layout TSX file: %s\n", outputPath)
	}

	// Convert TSX to JS and write JS file
	finalJSContent := TSX2JS(tsxContent)

	// Determine JS file path
	jsFileName := baseName + ".js"
	jsFilePath := filepath.Join("tpl/generated/js", jsFileName)
	if err := os.MkdirAll("tpl/generated/js", 0755); err != nil {
		return fmt.Errorf("failed to create JS directory: %v", err)
	}

	if err := os.WriteFile(jsFilePath, []byte(finalJSContent), 0644); err != nil {
		return fmt.Errorf("failed to write layout JS file: %v", err)
	}

	if isDebugTranspile() {
		fmt.Printf("DEBUG: Wrote layout JS file: %s\n", jsFilePath)
	}

	return nil
}

// processLayoutContent processes layout-specific content and structure
func processLayoutContent(htmlContent string) string {
	// Handle layout-specific processing
	htmlContent = replaceUnusedInHtml(htmlContent)

	// Remove HTML comments after processing includes
	htmlContent = removeHTMLComments(htmlContent)

	return htmlContent
}

// integrateLinkPathsAndScriptPaths integrates linkPaths and scriptPaths into the HTML content
func integrateLinkPathsAndScriptPaths(htmlContent string) string {
	// Find the </head> tag and insert linkPaths before it
	headClosePattern := regexp.MustCompile(`(?i)</head>`)
	if headClosePattern.MatchString(htmlContent) {
		// Insert linkPaths before </head>
		linkPathsJSX := `{linkPaths && linkPaths.split(',').map((link: string, index: any) => (` +
			`<link rel="stylesheet" src={link} />))}\n`
		htmlContent = headClosePattern.ReplaceAllString(htmlContent, linkPathsJSX+"</head>")
	}

	// Find the </body> tag and insert scriptPaths before it
	bodyClosePattern := regexp.MustCompile(`(?i)</body>`)
	if bodyClosePattern.MatchString(htmlContent) {
		// Insert scriptPaths before </body>
		scriptPathsJSX := `{scriptPaths && scriptPaths.split(',').map((script: string, index: any) => (` +
			`<script type="text/javascript" src={script} />))}\n`
		htmlContent = bodyClosePattern.ReplaceAllString(htmlContent, scriptPathsJSX+"</body>")
	}

	return htmlContent
}

// extractLayoutComponents extracts component references from layout files
func extractLayoutComponents(htmlContent string) []string {
	var components []string

	// Parse HTML to find component references
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return components
	}

	// Walk through the HTML to find component divs
	var walker func(*html.Node)
	walker = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "div" {
			// Check for component ID attributes
			for _, attr := range n.Attr {
				if attr.Key == "id" && strings.HasPrefix(attr.Val, "component-") {
					componentName := strings.TrimPrefix(attr.Val, "component-")
					componentName = convertToCamelCase(componentName)
					components = append(components, componentName)
				}
			}
		}

		// Process children
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walker(c)
		}
	}

	walker(doc)

	return components
}

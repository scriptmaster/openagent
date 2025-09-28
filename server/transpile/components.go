package transpile

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/scriptmaster/openagent/common"
	"golang.org/x/net/html"
)

// ============================================================================
// COMPONENT TRANSPILATION FUNCTIONS
// ============================================================================
// This file contains functions for transpiling and managing React components
//
// Function List:
// - processComponentImports(htmlContent, inputPath string) (string, []string, error)
//   Processes component divs and imports component files
// - importAndTranspileComponent(componentName, inputPath string) (string, error)
//   Imports a component file and transpiles it
// - convertComponentTSXToJS(tsxContent string) string
//   Converts component TSX to JS with React.createElement calls
// - convertCounterJSXToReact(tsxContent string, componentName string) string
//   Converts component JSX to React.createElement calls
// - replaceComponentDivsWithCalls(jsContent string) string
//   Replaces component divs with component function calls in JS
// - findComponentJSFiles(htmlContent string) []string
//   Finds component JS files referenced in HTML content
// - embedComponentJS(mainJSContent string) string
//   Embeds only the component JS content that is referenced in the main JS file
// - isBuiltInComponent(componentName string) bool
//   Checks if a component name is a built-in React component or HTML element
// ============================================================================

// processComponentImports processes component divs and imports component files
func processComponentImports(htmlContent, inputPath string) (string, []string, error) {
	if isDebugTranspile() {
		fmt.Printf("DEBUG: Processing component imports for: %s\n", inputPath)
	}

	// Use regex for initial component template replacement (more reliable for this specific case)
	componentPattern := regexp.MustCompile(`<template\s+id="component[-_](\w+)"[^>]*>\s*</template>`)

	var importedComponents []string

	// Find all component matches first
	matches := componentPattern.FindAllStringSubmatch(htmlContent, -1)
	if isDebugTranspile() {
		fmt.Printf("DEBUG: Found %d component matches: %v\n", len(matches), matches)
	}

	// Process each component
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		componentName := match[1]                                     // e.g., "simple"
		componentNameCapitalized := convertToCamelCase(componentName) // e.g., "Simple"

		if isDebugTranspile() {
			fmt.Printf("DEBUG: Processing component: %s\n", componentName)
		}

		// Import and transpile the component
		_, err := importAndTranspileComponent(componentName, inputPath)
		if err != nil {
			if isDebugTranspile() {
				fmt.Printf("DEBUG: Failed to import component %s: %v\n", componentName, err)
			}
			continue
		}

		if isDebugTranspile() {
			fmt.Printf("DEBUG: Successfully imported component %s\n", componentName)
		}

		// Track this component for import statements
		importedComponents = append(importedComponents, componentNameCapitalized)
	}

	// Replace component divs with JSX components
	result := componentPattern.ReplaceAllStringFunc(htmlContent, func(match string) string {
		// Extract component name from the match
		submatches := componentPattern.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match // Return original if no match
		}

		componentName := submatches[1]                                // e.g., "simple"
		componentNameCapitalized := convertToCamelCase(componentName) // e.g., "Simple"

		// Replace the div with the component directly, with hydration warning suppression
		result := fmt.Sprintf(`<%s suppressHydrationWarning={true} />`, componentNameCapitalized)
		// result := fmt.Sprintf(`<%s suppressHydrationWarning="true"></%s>`, componentNameCapitalized, componentNameCapitalized)
		if isDebugTranspile() {
			fmt.Printf("DEBUG: Replacing component div with: %s\n", result)
		}
		return result
	})

	return result, importedComponents, nil
}

// importAndTranspileComponent imports a component file and transpiles it
func importAndTranspileComponent(componentName, inputPath string) (string, error) {
	if isDebugTranspile() {
		fmt.Printf("DEBUG: Importing component: %s\n", componentName)
	}

	// Construct component file path (try relative first, then absolute)
	componentPath := fmt.Sprintf("tpl/components/%s.html", componentName)

	// If relative path doesn't exist, try from project root
	if _, err := os.Stat(componentPath); os.IsNotExist(err) {
		// Try from project root (go up two directories from server/transpile)
		componentPath = fmt.Sprintf("../../tpl/components/%s.html", componentName)
	}

	// Check if component file exists
	if _, err := os.Stat(componentPath); os.IsNotExist(err) {
		return "", fmt.Errorf("component file not found: %s", componentPath)
	}

	// Read component file
	componentContent, err := os.ReadFile(componentPath)
	if err != nil {
		return "", fmt.Errorf("failed to read component file: %v", err)
	}

	if isDebugTranspile() {
		fmt.Printf("DEBUG: Read component file: %s (%d bytes)\n", componentPath, len(componentContent))
	}

	// Extract CSS and JS from component
	componentJSPath := fmt.Sprintf("tpl/generated/js/component_%s.js", componentName)
	_, jsContent, err := extractCSSAndJS(string(componentContent), componentPath, componentJSPath, componentName)
	if err != nil {
		return "", fmt.Errorf("failed to extract CSS/JS from component: %v", err)
	}

	if isDebugTranspile() {
		fmt.Printf("DEBUG: Extracted JS content: %d bytes\n", len(jsContent))
	}

	// Process component HTML using HTML parser
	componentHTML := string(componentContent)
	componentHTML = processComponentHTML(componentHTML)

	// Generate component TSX file
	componentTSXPath := fmt.Sprintf("tpl/generated/components/%s.tsx", componentName)
	if err := os.MkdirAll("tpl/generated/components", 0755); err != nil {
		return "", fmt.Errorf("failed to create components directory: %v", err)
	}

	// Apply HTML minification to the content before creating TSX
	// Check if HTML_WHITESPACE_NOHYDRATE=1 is set to preserve whitespace for hydration
	preserveWhitespace := os.Getenv("HTML_WHITESPACE_NOHYDRATE") == "1"
	componentHTML = common.MinifyHTML(componentHTML, preserveWhitespace)

	// Write component TSX
	componentTSX := fmt.Sprintf("export default function %s() {\n    return (\n        %s\n    );\n}",
		convertToCamelCase(componentName), componentHTML)

	if err := os.WriteFile(componentTSXPath, []byte(componentTSX), 0644); err != nil {
		return "", fmt.Errorf("failed to write component TSX: %v", err)
	}

	if isDebugTranspile() {
		fmt.Printf("DEBUG: Generated component TSX: %s\n", componentTSXPath)
	}

	// Store component JS content for later embedding
	componentJS := convertComponentTSXToJS(componentTSX)

	if isDebugTranspile() {
		fmt.Printf("DEBUG: Converted TSX to JS: %d bytes\n", len(componentJS))
		fmt.Printf("DEBUG: Component JS content: %s\n", componentJS[:min(200, len(componentJS))])
	}

	// Add the script content (prototype methods) with ASCII art
	componentJS += "\n\n// ╔══════════════════════════════════════════════════════════════════════════════\n"
	componentJS += "// ║                        🔧 COMPONENT PROTOTYPE METHODS 🔧                        \n"
	componentJS += "// ╚══════════════════════════════════════════════════════════════════════════════\n\n"
	componentJS += jsContent

	// Write component JS to file for later embedding
	componentJSFile := fmt.Sprintf("tpl/generated/js/component_%s.js", componentName)
	if err := os.WriteFile(componentJSFile, []byte(componentJS), 0644); err != nil {
		if isDebugTranspile() {
			fmt.Printf("DEBUG: Failed to write component JS file %s: %v\n", componentJSFile, err)
		}
	} else {
		if isDebugTranspile() {
			fmt.Printf("DEBUG: Wrote component JS file: %s (%d bytes)\n", componentJSFile, len(componentJS))
		}
	}

	return componentHTML, nil
}

// processComponentHTML processes component HTML using simple string processing
func processComponentHTML(htmlContent string) string {
	// Remove script tags (they'll be processed separately)
	scriptPattern := regexp.MustCompile(`(?s)<script[^>]*>.*?</script>`)
	htmlContent = scriptPattern.ReplaceAllString(htmlContent, "")

	// Remove HTML comments
	commentPattern := regexp.MustCompile(`(?s)<!--.*?-->`)
	htmlContent = commentPattern.ReplaceAllString(htmlContent, "")

	// Convert class to className
	htmlContent = strings.ReplaceAll(htmlContent, "class=", "className=")

	// Fix self-closing tags
	htmlContent = fixSelfClosingTags(htmlContent)

	return htmlContent
}

// recProcessComponentNodes processes HTML nodes for component content
func recProcessComponentNodes(n *html.Node) {
	// Remove script tags (they'll be processed separately)
	if n.Type == html.ElementNode && n.Data == "script" {
		// Remove this node by setting it to nil
		n.Type = html.ErrorNode
		return
	}

	// Process children
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		recProcessComponentNodes(c)
	}
}

// convertComponentTSXToJS converts component TSX to JS with React.createElement calls
func convertComponentTSXToJS(tsxContent string) string {
	if isDebugTranspile() {
		fmt.Printf("DEBUG: convertComponentTSXToJS called with: %s\n", tsxContent[:min(200, len(tsxContent))])
	}

	// Extract component name from the TSX content
	componentNamePattern := regexp.MustCompile(`function\s+(\w+)\s*\(`)
	matches := componentNamePattern.FindStringSubmatch(tsxContent)
	componentName := "Component" // fallback
	if len(matches) > 1 {
		componentName = matches[1]
	}

	result := convertCounterJSXToReact(tsxContent, componentName)

	if isDebugTranspile() {
		fmt.Printf("DEBUG: convertCounterJSXToReact result: %s\n", result[:min(200, len(result))])
	}

	return result
}

// convertCounterJSXToReact converts component JSX to React.createElement calls
func convertCounterJSXToReact(tsxContent string, componentName string) string {
	// Extract the JSX content from the return statement
	jsxPattern := regexp.MustCompile(`(?s)return\s*\(\s*(.*?)\s*\)\s*;`)
	matches := jsxPattern.FindStringSubmatch(tsxContent)
	if len(matches) < 2 {
		return tsxContent // Return original if no match
	}

	jsxContent := matches[1]

	// Convert the JSX to React.createElement calls using HTML parser
	result := parseJSXWithHTMLParser(jsxContent)

	// Create the final function with proper React.createElement syntax
	finalResult := fmt.Sprintf("function %s() {\n    return (\n        %s\n    );\n}", convertToCamelCase(componentName), result)

	return finalResult
}

// replaceComponentDivsWithCalls replaces component divs with component function calls in JS
func replaceComponentDivsWithCalls(jsContent string) string {
	// Parse the JS content to find React.createElement calls
	doc, err := html.Parse(strings.NewReader(jsContent))
	if err != nil {
		return jsContent
	}

	// Walk through the parsed content and replace component divs
	var result strings.Builder
	recReplaceComponentDivsInNode(doc, &result)

	return result.String()
}

// recReplaceComponentDivsInNode recursively replaces component divs in HTML nodes
func recReplaceComponentDivsInNode(n *html.Node, result *strings.Builder) {
	// This function would need to be implemented based on the specific
	// structure of the JS content and how component divs are represented
	// For now, just render the node normally
	html.Render(result, n)
}

// findComponentJSFiles finds component JS files referenced in HTML content
func findComponentJSFiles(htmlContent string) []string {
	var componentFiles []string

	// Parse HTML to find component divs
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return componentFiles
	}

	// Walk through the HTML to find component templates
	var walker func(*html.Node)
	walker = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "template" {
			// Check for component ID attributes
			for _, attr := range n.Attr {
				if attr.Key == "id" && (strings.HasPrefix(attr.Val, "component-") || strings.HasPrefix(attr.Val, "component_")) {
					componentName := strings.TrimPrefix(attr.Val, "component-")
					componentName = strings.TrimPrefix(componentName, "component_")
					componentJSFile := fmt.Sprintf("/tsx/js/component_%s.js", componentName)
					componentFiles = append(componentFiles, componentJSFile)
				}
			}
		}

		// Process children
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walker(c)
		}
	}

	walker(doc)

	return componentFiles
}

// isBuiltInComponent checks if a component name is a built-in React component or HTML element
func isBuiltInComponent(componentName string) bool {
	builtInComponents := map[string]bool{
		"div": true, "span": true, "p": true, "h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
		"button": true, "input": true, "form": true, "label": true, "select": true, "textarea": true,
		"ul": true, "ol": true, "li": true, "a": true, "img": true, "br": true, "hr": true,
		"table": true, "tr": true, "td": true, "th": true, "thead": true, "tbody": true,
		"header": true, "footer": true, "nav": true, "main": true, "section": true, "article": true,
		"React": true, "Fragment": true,
	}
	return builtInComponents[componentName]
}

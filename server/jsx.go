package server

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// Global regex patterns for JSX extraction
var (
	jsxTagPattern *regexp.Regexp
)

// init initializes global regex patterns
func init() {
	// Regex to match JSX content enclosed by <main> tags with optional attributes
	jsxTagPattern = regexp.MustCompile(`(?s)<main(?:\s+[^>]*)?>(.*?)</main>`)
}

// convertJSXToReactCreateElement converts JSX syntax to React.createElement calls using a node walker
func convertJSXToReactCreateElement(jsContent string) string {
	if isDebugTranspile() {
		fmt.Printf("DEBUG: convertJSXToReactCreateElement called with: %s\n", jsContent[:min(200, len(jsContent))])
	}
	// Use the JSX parser with node walking
	result := parseJSXWithWalker(jsContent)
	if isDebugTranspile() {
		fmt.Printf("DEBUG: convertJSXToReactCreateElement result: %s\n", result[:min(200, len(result))])
	}
	return result
}

// parseJSXWithWalker parses JSX using regex to extract <main> content and html.Parse to walk through nodes
func parseJSXWithWalker(jsx string) string {
	if isDebugTranspile() {
		fmt.Printf("DEBUG: parseJSXWithWalker called with: %s\n", jsx[:min(200, len(jsx))])
	}

	// Use regex to find JSX content enclosed by <main> tags
	matches := jsxTagPattern.FindAllStringSubmatch(jsx, -1)
	if isDebugTranspile() {
		fmt.Printf("DEBUG: Found %d matches\n", len(matches))
		for i, match := range matches {
			fmt.Printf("DEBUG: Match %d: %s\n", i, match[0][:min(50, len(match[0]))])
		}
	}

	if len(matches) == 0 {
		if isDebugTranspile() {
			fmt.Printf("DEBUG: No <main> tags found\n")
		}
		return jsx
	}

	// Take the first match (main content)
	match := matches[0]
	jsxContent := match[1] // The content inside <main>...</main>

	if isDebugTranspile() {
		fmt.Printf("DEBUG: Found <main> content: %s\n", jsxContent[:min(100, len(jsxContent))])
	}

	// Parse JSX content using html.Parse
	reactCode := parseJSXWithHTMLParser(jsxContent)

	// Replace the JSX content with React code
	// Find the original JSX tag in the original string
	originalTag := match[0] // Full match including opening and closing tags
	tagStart := strings.Index(jsx, originalTag)
	if tagStart == -1 {
		return jsx
	}

	// Replace the JSX content with React code
	result := jsx[:tagStart] + reactCode + jsx[tagStart+len(originalTag):]

	return result
}

// parseJSXWithHTMLParser uses html.Parse to parse JSX and convert to React.createElement
func parseJSXWithHTMLParser(jsxContent string) string {
	if isDebugTranspile() {
		fmt.Printf("DEBUG: parseJSXWithHTMLParser called with: %s\n", jsxContent[:min(100, len(jsxContent))])
	}
	// fmt.Printf("DEBUG: parseJSXWithHTMLParser called with: %s\n", jsxContent)

	// Wrap JSX content in html/body tags for proper parsing
	// wrappedHTML := "<html><body>" + jsxContent + "</body></html>"
	wrappedHTML := jsxContent

	// Parse the HTML
	doc, err := html.Parse(strings.NewReader(wrappedHTML))
	if err != nil {
		if isDebugTranspile() {
			fmt.Printf("DEBUG: html.Parse error: %v\n", err)
		}
		// If parsing fails, return the original content
		return jsxContent
	}

	// Walk through the parsed HTML and convert to React.createElement
	var result strings.Builder
	walkHTMLNode(doc, &result)

	if isDebugTranspile() {
		fmt.Printf("DEBUG: parseJSXWithHTMLParser result: %s\n", result.String()[:min(100, len(result.String()))])
	}

	// Fix className case after HTML parsing (HTML parser converts to lowercase)
	originalResult := result.String()
	finalResult := strings.ReplaceAll(originalResult, "{classname:", "{className:")

	// Also try replacing without the curly brace
	finalResult = strings.ReplaceAll(finalResult, "classname:", "className:")

	if isDebugTranspile() {
		fmt.Printf("DEBUG: Original result: %s\n", originalResult[:min(100, len(originalResult))])
		fmt.Printf("DEBUG: Final result: %s\n", finalResult[:min(100, len(finalResult))])
	}

	return finalResult
}

// walkHTMLNode walks through HTML nodes and converts them to React.createElement calls
func walkHTMLNode(n *html.Node, result *strings.Builder) {
	switch n.Type {
	case html.ElementNode:
		// Convert element to React.createElement
		convertElementToReact(n, result)
	case html.TextNode:
		// Convert text node
		text := strings.TrimSpace(n.Data)
		if text != "" {
			result.WriteString(fmt.Sprintf("'%s'", text))
		}
	case html.DocumentNode:
		// Skip document node, process children
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walkHTMLNode(c, result)
		}
	case html.DoctypeNode:
		// Skip doctype
		return
	default:
		// Process children for other node types
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walkHTMLNode(c, result)
		}
	}
}

// convertElementToReact converts an HTML element to React.createElement
func convertElementToReact(n *html.Node, result *strings.Builder) {
	// Skip html, body, and head tags (they're just wrappers)
	if n.Data == "html" || n.Data == "body" || n.Data == "head" {
		// Process children
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walkHTMLNode(c, result)
		}
		return
	}

	// Build props object
	props := buildPropsObject(n)

	// Process children
	var children strings.Builder
	hasChildren := false
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		// Skip empty text nodes
		if c.Type == html.TextNode && strings.TrimSpace(c.Data) == "" {
			continue
		}
		if hasChildren {
			children.WriteString(", ")
		}
		walkHTMLNode(c, &children)
		hasChildren = true
	}

	// Build React.createElement call
	result.WriteString(fmt.Sprintf("React.createElement('%s', %s", n.Data, props))

	if hasChildren {
		result.WriteString(", ")
		result.WriteString(children.String())
	}

	result.WriteString(")")
}

// buildPropsObject builds the props object for React.createElement
func buildPropsObject(n *html.Node) string {
	if len(n.Attr) == 0 {
		return "null"
	}

	var props []string
	for _, attr := range n.Attr {
		// Convert class/classname to className (JSX uses className instead of class)
		key := attr.Key
		if isDebugTranspile() {
			fmt.Printf("DEBUG: Processing attribute: %s = %s\n", key, attr.Val)
		}
		if key == "class" || key == "classname" {
			key = "className"
		}
		props = append(props, fmt.Sprintf("%s: '%s'", key, attr.Val))
	}

	return "{" + strings.Join(props, ", ") + "}"
}

// parseJSXAttributes parses JSX attributes into a JavaScript object string
func parseJSXAttributes(attributes string) string {
	if strings.TrimSpace(attributes) == "" {
		return "null"
	}

	// Simple attribute parser - converts className="value" to {className: "value"}
	attrPattern := regexp.MustCompile(`(\w+)=["']([^"']*)["']`)
	props := make([]string, 0)

	matches := attrPattern.FindAllStringSubmatch(attributes, -1)
	for _, match := range matches {
		key := match[1]
		value := match[2]
		props = append(props, fmt.Sprintf("%s: '%s'", key, value))
	}

	if len(props) == 0 {
		return "null"
	}

	return "{" + strings.Join(props, ", ") + "}"
}

// removeTypeScriptTypes removes TypeScript type annotations
func removeTypeScriptTypes(content string) string {
	// Remove object type annotations first: {param: Type} -> {param}
	// Use string replacement for common patterns
	// content = strings.ReplaceAll(content, "{page: Page}", "{page}")
	// content = strings.ReplaceAll(content, "{page: any}", "{page}")
	// content = strings.ReplaceAll(content, "{page: object}", "{page}")

	// Also handle generic object type pattern
	objectTypePattern := regexp.MustCompile(`\{(\w+):\s*[A-Za-z]+\}`)
	content = objectTypePattern.ReplaceAllString(content, "{$1}")

	// Remove function parameter types: (param: Type) -> (param)
	// Use string replacement for common patterns to avoid regex issues
	// content = strings.ReplaceAll(content, "page: Page", "page")
	// content = strings.ReplaceAll(content, "page: any", "page")
	content = strings.ReplaceAll(content, "{page}: {page: Page}", "{page}")
	content = strings.ReplaceAll(content, "{page}: {page: any}", "{page}")
	content = strings.ReplaceAll(content, "{page}: {page: object}", "{page}")
	content = strings.ReplaceAll(content, "{page}: {page}", "{page}")
	content = strings.ReplaceAll(content, "{page}: any", "{page}")

	// Remove return type annotations: ): Type { -> ) {
	returnTypePattern := regexp.MustCompile(`\):\s*[A-Za-z0-9_\[\]|&<>{}]+\s*\{`)
	content = returnTypePattern.ReplaceAllString(content, ") {")

	return content
}

// GetActualComponentName extracts the actual component name from the function signature
// If not found, uses the provided componentName as fallback (capitalized)
func GetActualComponentName(componentJS string, componentName string) string {
	// Look for function declaration pattern: function ComponentName(
	funcPattern := regexp.MustCompile(`function\s+(\w+)\s*\(`)
	matches := funcPattern.FindStringSubmatch(componentJS)

	if isDebugTranspile() {
		fmt.Printf("DEBUG: GetActualComponentName input: %s\n", componentJS[:min(100, len(componentJS))])
		fmt.Printf("DEBUG: GetActualComponentName matches: %v\n", matches)
		fmt.Printf("DEBUG: GetActualComponentName fallback: %s\n", componentName)
	}

	if len(matches) > 1 {
		if isDebugTranspile() {
			fmt.Printf("DEBUG: GetActualComponentName returning: %s\n", matches[1])
		}
		return matches[1] // Return the component name
	}

	// Fallback: capitalize the provided componentName
	if componentName != "" {
		capitalized := strings.Title(strings.ToLower(componentName))
		if isDebugTranspile() {
			fmt.Printf("DEBUG: GetActualComponentName fallback capitalized: %s\n", capitalized)
		}
		return capitalized
	}

	// Final fallback: return "Component" if no match found and no componentName provided
	if isDebugTranspile() {
		fmt.Printf("DEBUG: GetActualComponentName final fallback: Component\n")
	}
	return "Component"
}

// min helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

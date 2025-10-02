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
// JSX TRANSPILATION FUNCTIONS
// ============================================================================
// This file contains functions for converting JSX to React.createElement calls
//
// Function List:
// - TSX2JS(tsxStr string) string
//   Main function to convert TSX to JavaScript
// - convertJSXToReactCreateElement(jsContent string) string
//   Converts JSX syntax to React.createElement calls using HTML parser
// - parseJSXWithWalker(jsx string) string
//   Parses JSX using regex to extract <main> content and HTML parser to walk through nodes
// - parseJSXWithHTMLParser(jsxContent string) string
//   Uses HTML parser to parse JSX and convert to React.createElement
// - recWalkHTMLNodeWithCustomComponents(n *html.Node, result *strings.Builder, customComponents map[string]string)
//   Walks through HTML nodes and converts them to React.createElement calls
// - recConvertElementToReactWithCustomComponents(n *html.Node, result *strings.Builder, customComponents map[string]string)
//   Converts HTML elements to React.createElement calls
// - buildPropsObject(n *html.Node) string
//   Builds the props object for React.createElement
// - extractCustomComponentNames(jsxContent string) map[string]string
//   Extracts custom component names to preserve their case
// - isCustomReactComponent(componentName string) bool
//   Checks if a component name is a custom React component
// - isOnlyIndentation(text string) bool
//   Checks if text contains only indentation characters
// - createReactJSContent(originalJS, componentName string) string
//   Creates React-enhanced JS content with TSX content directly embedded
// - getActualComponentName(componentJS, componentName string) string
//   Extracts the actual component name from JS content
// - embedComponentJS(jsContent string) string
//   Embeds component JS content into the main JS file
// - removeTypeScriptTypes(tsxContent string) string
//   Removes TypeScript type annotations from TSX content
// ============================================================================

var (
	// Regex to match JSX content enclosed by <main> tags with optional attributes
	jsxTagPattern = regexp.MustCompile(`(?s)<main(?:\s+[^>]*)?>(.*?)</main>`)
)

// TSX2JS converts TSX to JavaScript
func TSX2JS(tsxStr string) string {
	if isDebugTranspile() {
		fmt.Printf("DEBUG: TSX2JS called with: %s\n", tsxStr[:min(200, len(tsxStr))])
	}

	// Check if this is a component TSX (containing: export default function)
	if strings.Contains(strings.TrimSpace(tsxStr), "export default function") {
		// This is a component TSX file, extract JSX from the return statement
		returnStart := strings.Index(tsxStr, "return (")
		returnEnd := strings.LastIndex(tsxStr, ");")
		if returnStart != -1 && returnEnd != -1 {
			returnStart += 8 // Length of "return ("
			mainContent := tsxStr[returnStart:returnEnd]

			// Fix self-closing tags before parsing
			mainContent = fixCustomJSXSelfClosingTags(mainContent)

			// Convert JSX to React.createElement calls
			jsxStr := convertJSXToReactCreateElement(mainContent)

			// Fix attribute case issues
			jsxStr = fixAttributeCases(jsxStr)

			// Clean up extra whitespace
			jsxStr = regexp.MustCompile(`\n\s*\n`).ReplaceAllString(jsxStr, "\n")
			jsxStr = strings.TrimSpace(jsxStr)

			if isDebugTranspile() {
				fmt.Printf("DEBUG: TSX2JS result: %s\n", jsxStr[:min(200, len(jsxStr))])
			}

			return jsxStr
		}
	}

	// Extract main content from JSX
	var mainContent string
	if strings.Contains(tsxStr, "<main>") {
		// Extract content between <main> tags
		mainStart := strings.Index(tsxStr, "<main>")
		mainEnd := strings.LastIndex(tsxStr, "</main>")
		if mainStart != -1 && mainEnd != -1 {
			mainStart += 6 // Length of "<main>"
			mainContent = tsxStr[mainStart:mainEnd]
		} else {
			mainContent = tsxStr
		}
	} else {
		mainContent = tsxStr
	}

	// Fix self-closing tags before parsing
	mainContent = fixCustomJSXSelfClosingTags(mainContent)

	// Convert JSX to React.createElement calls
	jsxStr := convertJSXToReactCreateElement(mainContent)

	// Fix attribute case issues (HTML parser converts to camelCase)
	jsxStr = fixAttributeCases(jsxStr)

	// Clean up extra whitespace
	jsxStr = regexp.MustCompile(`\n\s*\n`).ReplaceAllString(jsxStr, "\n")
	jsxStr = strings.TrimSpace(jsxStr)

	if isDebugTranspile() {
		fmt.Printf("DEBUG: TSX2JS result: %s\n", jsxStr[:min(200, len(jsxStr))])
	}

	return jsxStr
}

// convertJSXToReactCreateElement converts JSX syntax to React.createElement calls using HTML parser
func convertJSXToReactCreateElement(jsContent string) string {
	if isDebugTranspile() {
		fmt.Printf("DEBUG: convertJSXToReactCreateElement called with: %s\n", jsContent[:min(200, len(jsContent))])
	}

	// Use the JSX parser with HTML parser
	result := parseJSXWithHTMLParser(jsContent)
	if isDebugTranspile() {
		fmt.Printf("DEBUG: convertJSXToReactCreateElement result: %s\n", result[:min(200, len(result))])
	}
	return result
}

// parseJSXWithHTMLParser uses HTML parser to parse JSX and convert to React.createElement
func parseJSXWithHTMLParser(jsxContent string) string {
	if isDebugTranspile() {
		fmt.Printf("DEBUG: parseJSXWithHTMLParser called with: %s\n", jsxContent[:min(100, len(jsxContent))])
	}

	// Remove main tags from input if present
	if strings.Contains(jsxContent, "<main>") {
		mainStart := strings.Index(jsxContent, "<main>")
		mainEnd := strings.LastIndex(jsxContent, "</main>")
		if mainStart != -1 && mainEnd != -1 {
			mainStart += 6 // Length of "<main>"
			jsxContent = jsxContent[mainStart:mainEnd]
		}
	}

	// Handle React Fragments before HTML parsing
	jsxContent = handleReactFragments(jsxContent)

	// Fix custom JSX self-closing tags before parsing
	jsxContent = fixCustomJSXSelfClosingTags(jsxContent)

	// Convert JSX self-closing custom components to React.createElement calls directly
	// This bypasses the HTML parser issues with custom components
	jsxContent = convertJSXComponentsToReactCreateElement(jsxContent)

	// Now convert the remaining HTML elements to React.createElement calls
	// This will handle the div and span elements properly

	// Extract custom component names to preserve their case
	customComponents := extractCustomComponentNames(jsxContent)
	if isDebugTranspile() {
		fmt.Printf("DEBUG: Found custom components: %v\n", customComponents)
		fmt.Printf("DEBUG: JSX content: %s\n", jsxContent[:min(200, len(jsxContent))])
	}

	// Parse the HTML
	doc, err := html.Parse(strings.NewReader(jsxContent))
	if err != nil {
		if isDebugTranspile() {
			fmt.Printf("DEBUG: html.Parse error: %v\n", err)
		}
		// If parsing fails, return the original content
		return jsxContent
	}

	// Find the body element (HTML parser adds html/head/body structure)
	var bodyNode *html.Node
	var findBody func(*html.Node)
	findBody = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "body" {
			bodyNode = n
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findBody(c)
		}
	}
	findBody(doc)

	// Walk through the parsed HTML and convert to React.createElement
	var result strings.Builder

	// If we found a body node, process only its children
	if bodyNode != nil {
		for c := bodyNode.FirstChild; c != nil; c = c.NextSibling {
			recWalkHTMLNodeWithCustomComponents(c, &result, customComponents)
		}
	} else {
		// Check if the document has html/head/body structure
		var htmlNode *html.Node
		for c := doc.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && c.Data == "html" {
				htmlNode = c
				break
			}
		}

		if htmlNode != nil {
			// Process children of html node, skipping head
			for c := htmlNode.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode && c.Data == "body" {
					// Process body children
					for bodyChild := c.FirstChild; bodyChild != nil; bodyChild = bodyChild.NextSibling {
						recWalkHTMLNodeWithCustomComponents(bodyChild, &result, customComponents)
					}
				} else if c.Type == html.ElementNode && c.Data != "head" {
					// Process non-head elements directly
					recWalkHTMLNodeWithCustomComponents(c, &result, customComponents)
				}
			}
		} else {
			// Fallback: process the entire document
			recWalkHTMLNodeWithCustomComponents(doc, &result, customComponents)
		}
	}

	if isDebugTranspile() {
		fmt.Printf("DEBUG: parseJSXWithHTMLParser result: %s\n", result.String()[:min(100, len(result.String()))])
	}

	// Fix className case after HTML parsing (HTML parser converts to lowercase)
	originalResult := result.String()
	finalResult := strings.ReplaceAll(originalResult, "{classname:", "{className:")

	// Also try replacing without the curly brace
	finalResult = strings.ReplaceAll(finalResult, "classname:", "className:")

	if isDebugTranspile() {
		fmt.Printf("DEBUG: Final result after className fix: %s\n", finalResult[:min(100, len(finalResult))])
	}

	return finalResult
}

// recWalkHTMLNodeWithCustomComponents walks through HTML nodes and converts them to React.createElement calls
func recWalkHTMLNodeWithCustomComponents(n *html.Node, result *strings.Builder, customComponents map[string]string) {
	switch n.Type {
	case html.ElementNode:
		// Convert element to React.createElement
		recConvertElementToReactWithCustomComponents(n, result, customComponents)
	case html.TextNode:
		// Convert text node - preserve meaningful whitespace to avoid hydration mismatch
		text := n.Data // strings.TrimSpace(n.Data)
		// Only preserve whitespace if it's not just indentation (tabs/spaces at start of line)
		if text != "" && !isOnlyIndentation(n.Data) {
			// Check if this text is a React.createElement call (from our conversion)
			trimmedText := strings.TrimSpace(text)
			if strings.HasPrefix(trimmedText, "React.createElement(") {
				// This is a React.createElement call, trim whitespace and don't wrap it in quotes
				result.WriteString(trimmedText)
			} else {
				// This is regular text, escape it for JavaScript and wrap it in quotes
				escapedText := escapeJSString(text)
				result.WriteString(fmt.Sprintf("'%s'", escapedText))
			}
		}
	case html.DocumentNode:
		// Process children for document node with comma separation
		hasChildren := false
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			// Skip completely empty text nodes and indentation-only nodes
			if c.Type == html.TextNode && (c.Data == "" || isOnlyIndentation(c.Data)) {
				continue
			}
			if hasChildren {
				result.WriteString(", ")
			}
			recWalkHTMLNodeWithCustomComponents(c, result, customComponents)
			hasChildren = true
		}
	case html.DoctypeNode:
		// Skip doctype
		return
	default:
		// Process children for other node types with comma separation
		hasChildren := false
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			// Skip completely empty text nodes and indentation-only nodes
			if c.Type == html.TextNode && (c.Data == "" || isOnlyIndentation(c.Data)) {
				continue
			}
			if hasChildren {
				result.WriteString(", ")
			}
			recWalkHTMLNodeWithCustomComponents(c, result, customComponents)
			hasChildren = true
		}
	}
}

// recConvertElementToReactWithCustomComponents converts HTML elements to React.createElement calls
func recConvertElementToReactWithCustomComponents(n *html.Node, result *strings.Builder, customComponents map[string]string) {
	// Skip html, body, and head tags (they're just wrappers)
	if n.Data == "html" || n.Data == "body" || n.Data == "head" {
		// Process children
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			recWalkHTMLNodeWithCustomComponents(c, result, customComponents)
		}
		return
	}

	// Check if this is a React Fragment (HTML parser converts to lowercase)
	if n.Data == "react.fragment" {
		// Handle React Fragment - just process children without wrapper
		var children strings.Builder
		hasChildren := false
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			// Skip completely empty text nodes and indentation-only nodes
			if c.Type == html.TextNode && (c.Data == "" || isOnlyIndentation(c.Data)) {
				continue
			}
			if hasChildren {
				children.WriteString(", ")
			}
			recWalkHTMLNodeWithCustomComponents(c, &children, customComponents)
			hasChildren = true
		}
		if hasChildren {
			result.WriteString(children.String())
		}
		return
	}

	// Check if this is a custom React component
	var componentName string
	var isCustomComponent bool

	// Check if this is a temporary JSX tag (jsx_componentname)
	if strings.HasPrefix(n.Data, "jsx_") {
		// Extract original component name from data-original-component attribute
		for _, attr := range n.Attr {
			if attr.Key == "data-original-component" {
				componentName = attr.Val
				isCustomComponent = true
				if isDebugTranspile() {
					fmt.Printf("DEBUG: Found temporary JSX component: %s -> %s\n", n.Data, componentName)
				}
				break
			}
		}
		if !isCustomComponent {
			// Fallback: extract from tag name
			componentName = strings.TrimPrefix(n.Data, "jsx_")
			componentName = strings.Title(componentName)
			isCustomComponent = true
		}
	} else if customComponents != nil {
		if isDebugTranspile() {
			// fmt.Printf("DEBUG: Looking for component '%s' in customComponents map: %v\n", n.Data, customComponents)
		}
		if originalName, exists := customComponents[n.Data]; exists {
			componentName = originalName
			isCustomComponent = true
			if isDebugTranspile() {
				fmt.Printf("DEBUG: Found custom component: %s -> %s\n", n.Data, originalName)
			}
		} else {
			componentName = n.Data
			isCustomComponent = false
			if isDebugTranspile() {
				// fmt.Printf("DEBUG: Not a custom component: %s (not found in map)\n", n.Data)
			}
		}
	} else {
		componentName = n.Data
		isCustomComponent = isCustomReactComponent(n.Data)
		if isDebugTranspile() {
			fmt.Printf("DEBUG: No customComponents map, using isCustomReactComponent: %s -> %v\n", n.Data, isCustomComponent)
		}
	}

	// Build props object
	props := buildPropsObject(n)

	// Process children
	var children strings.Builder
	hasChildren := false
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		// Skip completely empty text nodes and indentation-only nodes
		if c.Type == html.TextNode && (c.Data == "" || isOnlyIndentation(c.Data)) {
			continue
		}
		if hasChildren {
			children.WriteString(", ")
		}
		recWalkHTMLNodeWithCustomComponents(c, &children, customComponents)
		hasChildren = true
	}

	// Build React.createElement call
	if isCustomComponent {
		// Custom React component - use component name directly (no quotes)
		result.WriteString(fmt.Sprintf("React.createElement(%s, %s", componentName, props))
	} else {
		// Standard HTML element - use string name (with quotes)
		result.WriteString(fmt.Sprintf("React.createElement('%s', %s", componentName, props))
	}

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

	var props strings.Builder
	props.WriteString("{")

	first := true
	for _, attr := range n.Attr {
		// Skip the data-original-component attribute as it's only for internal processing
		if attr.Key == "data-original-component" {
			continue
		}

		if !first {
			props.WriteString(", ")
		}
		first = false

		// Keep original attribute name (don't convert to camelCase)
		propName := attr.Key

		// Handle special cases
		if propName == "class" {
			props.WriteString(fmt.Sprintf("className: \"%s\"", attr.Val))
		} else if propName == "for" {
			props.WriteString(fmt.Sprintf("htmlFor: \"%s\"", attr.Val))
		} else if strings.HasPrefix(attr.Val, "{") && strings.HasSuffix(attr.Val, "}") {
			// Already a JS expression - remove the braces since we're in a JS object
			innerValue := strings.TrimPrefix(strings.TrimSuffix(attr.Val, "}"), "{")
			props.WriteString(fmt.Sprintf("%s: %s", propName, innerValue))
		} else {
			// Regular string value
			props.WriteString(fmt.Sprintf("%s: \"%s\"", propName, attr.Val))
		}
	}

	props.WriteString("}")
	return props.String()
}

// extractCustomComponentNames extracts custom component names to preserve their case
func extractCustomComponentNames(jsxContent string) map[string]string {
	customComponents := make(map[string]string)

	// Use regex to find custom components in the original JSX content
	// Look for <ComponentName> patterns where ComponentName starts with capital letter
	componentPattern := regexp.MustCompile(`<([A-Z][a-zA-Z0-9]*)[^>]*>`)
	matches := componentPattern.FindAllStringSubmatch(jsxContent, -1)

	for _, match := range matches {
		if len(match) > 1 {
			componentName := match[1]
			// Store mapping from lowercase to original case
			customComponents[strings.ToLower(componentName)] = componentName
		}
	}

	if isDebugTranspile() {
		fmt.Printf("DEBUG: Found custom components: %v\n", customComponents)
	}

	return customComponents
}

// isCustomReactComponent checks if a component name is a custom React component
func isCustomReactComponent(componentName string) bool {
	// Custom components start with a capital letter
	if len(componentName) == 0 {
		return false
	}
	firstChar := componentName[0]
	return firstChar >= 'A' && firstChar <= 'Z'
}

// isOnlyIndentation checks if text contains only indentation characters
func isOnlyIndentation(text string) bool {
	// Check if the text contains only whitespace characters
	for _, char := range text {
		if char != ' ' && char != '\t' && char != '\n' && char != '\r' {
			return false
		}
	}
	return true
}

// escapeJSString escapes a string for use in JavaScript string literals
func escapeJSString(s string) string {
	var result strings.Builder
	for _, r := range s {
		switch r {
		case '\n':
			result.WriteString("\\n")
		case '\r':
			result.WriteString("\\r")
		case '\t':
			result.WriteString("\\t")
		case '\\':
			result.WriteString("\\\\")
		case '\'':
			result.WriteString("\\'")
		case '"':
			result.WriteString("\\\"")
		default:
			result.WriteRune(r)
		}
	}
	return result.String()
}

// fixCustomJSXSelfClosingTags fixes custom JSX components to be self-closing
func fixCustomJSXSelfClosingTags(htmlContent string) string {
	if isDebugTranspile() {
		fmt.Printf("DEBUG: fixCustomJSXSelfClosingTags input: %s\n", htmlContent[:min(200, len(htmlContent))])
	}

	// Pattern to match custom JSX components that are NOT self-closing but should be
	nonSelfClosingPattern := regexp.MustCompile(`<([A-Z][a-zA-Z0-9]*)\s+([^>]*?)\s*>`)

	htmlContent = nonSelfClosingPattern.ReplaceAllStringFunc(htmlContent, func(match string) string {
		// Skip if this component already ends with />
		if strings.HasSuffix(strings.TrimSpace(match), "/>") {
			return match
		}

		// Extract component name and attributes
		submatches := nonSelfClosingPattern.FindStringSubmatch(match)
		if len(submatches) >= 3 {
			componentName := submatches[1]
			attributes := strings.TrimSpace(submatches[2])

			// Check if this component has a closing tag
			// Simple check: if the next occurrence of </componentName> is not immediately after this tag
			remainingContent := htmlContent[strings.Index(htmlContent, match)+len(match):]
			closingTag := fmt.Sprintf("</%s>", componentName)

			// If there's no closing tag or it's very far away, make it self-closing
			if !strings.Contains(remainingContent, closingTag) ||
				strings.Index(remainingContent, closingTag) > 1000 {
				// Ensure proper spacing
				if attributes != "" && !strings.HasPrefix(attributes, " ") {
					attributes = " " + attributes
				}
				return fmt.Sprintf("<%s%s />", componentName, attributes)
			}
		}
		return match
	})

	if isDebugTranspile() {
		fmt.Printf("DEBUG: fixCustomJSXSelfClosingTags output: %s\n", htmlContent[:min(200, len(htmlContent))])
	}

	return htmlContent
}

// convertJSXSelfClosingToHTML converts JSX self-closing custom components to proper HTML format
func convertJSXSelfClosingToHTML(jsxContent string) string {
	// Pattern to match JSX self-closing custom components: <ComponentName ... />
	jsxSelfClosingPattern := regexp.MustCompile(`<([A-Z][a-zA-Z0-9]*)\s+([^>]*?)\s*/>`)

	jsxContent = jsxSelfClosingPattern.ReplaceAllStringFunc(jsxContent, func(match string) string {
		// Extract component name and attributes
		submatches := jsxSelfClosingPattern.FindStringSubmatch(match)
		if len(submatches) >= 3 {
			componentName := submatches[1]
			attributes := strings.TrimSpace(submatches[2])

			// Convert to proper HTML self-closing format
			// Use a temporary tag name that won't conflict
			tempTagName := "jsx_" + strings.ToLower(componentName)

			// Ensure proper spacing
			if attributes != "" && !strings.HasPrefix(attributes, " ") {
				attributes = " " + attributes
			}

			// Add a data attribute to track the original component name
			originalNameAttr := fmt.Sprintf(" data-original-component=\"%s\"", componentName)

			return fmt.Sprintf("<%s%s%s />", tempTagName, attributes, originalNameAttr)
		}
		return match
	})

	return jsxContent
}

// convertJSXComponentsToReactCreateElement converts JSX custom components to React.createElement calls
func convertJSXComponentsToReactCreateElement(jsxContent string) string {
	// Pattern to match JSX self-closing custom components: <ComponentName ... />
	jsxComponentPattern := regexp.MustCompile(`<([A-Z][a-zA-Z0-9]*)\s*([^>]*?)\s*/>`)

	jsxContent = jsxComponentPattern.ReplaceAllStringFunc(jsxContent, func(match string) string {
		// Extract component name and attributes
		submatches := jsxComponentPattern.FindStringSubmatch(match)
		if len(submatches) >= 3 {
			componentName := submatches[1]
			attributes := strings.TrimSpace(submatches[2])

			// Convert attributes to props object
			props := "null"
			if attributes != "" {
				// Convert JSX attribute syntax to JavaScript object syntax
				// Replace = with : for proper object syntax
				jsAttributes := strings.ReplaceAll(attributes, "=", ": ")
				// Remove curly braces from JSX values like {true} -> true
				jsAttributes = strings.ReplaceAll(jsAttributes, "{", "")
				jsAttributes = strings.ReplaceAll(jsAttributes, "}", "")
				props = "{" + jsAttributes + "}"
			}

			return fmt.Sprintf("React.createElement(%s, %s)", componentName, props)
		}
		return match
	})

	// After converting components, handle multiple React.createElement calls on separate lines
	// by putting them on the same line with commas
	lines := strings.Split(jsxContent, "\n")
	var resultLines []string
	skipNext := false

	for i, line := range lines {
		if skipNext {
			skipNext = false
			continue
		}

		trimmedLine := strings.TrimSpace(line)

		// If this line contains a React.createElement call
		if strings.HasPrefix(trimmedLine, "React.createElement(") {
			// Check if the next non-empty line also contains a React.createElement call
			nextLineIndex := i + 1
			for nextLineIndex < len(lines) && strings.TrimSpace(lines[nextLineIndex]) == "" {
				nextLineIndex++
			}

			if nextLineIndex < len(lines) {
				nextLine := strings.TrimSpace(lines[nextLineIndex])
				if strings.HasPrefix(nextLine, "React.createElement(") {
					// Combine the two React.createElement calls with a comma
					resultLines = append(resultLines, trimmedLine+", "+nextLine)
					// Skip the next line since we've already processed it
					skipNext = true
					continue
				}
			}
		}

		resultLines = append(resultLines, line)
	}

	return strings.Join(resultLines, "\n")
}

// getTitleCase converts a string to title case by trimming, splitting by non-word characters,
// capitalizing the first letter of each resulting word, and joining them with spaces.
func getTitleCase(s string) string {
	s = strings.TrimSpace(s) // Trim s before processing
	if s == "" {
		return ""
	}

	// Split by one or more non-word characters (e.g., spaces, hyphens, underscores)
	re := regexp.MustCompile(`\W+`)
	parts := re.Split(s, -1)

	var capitalizedParts []string
	for _, part := range parts {
		if part == "" {
			continue // Skip empty strings that might result from multiple delimiters
		}
		// Capitalize the first letter of each word
		capitalizedParts = append(capitalizedParts, strings.ToUpper(string(part[0]))+part[1:])
	}

	// Join the capitalized words with a single space
	return strings.Join(capitalizedParts, " ")
}

// createReactJSContent creates React-enhanced JS content with TSX content directly embedded
func createReactJSContent(originalJS, componentName string) string {
	// Read the component TSX file first
	componentTsxPath := filepath.Join("tpl/generated/pages", componentName+".component.tsx")
	componentTsxContent, err := os.ReadFile(componentTsxPath)
	if err != nil {
		// If component file doesn't exist, try the regular TSX file
		tsxPath := filepath.Join("tpl/generated/pages", componentName+".tsx")
		tsxContent, err := os.ReadFile(tsxPath)
		if err != nil {
			// If neither file exists, just return the original JS
			return originalJS
		}
		componentTsxContent = tsxContent
	}

	// Convert component TSX to JS
	componentJS := TSX2JS(string(componentTsxContent))

	// Component replacement is now handled at TSX level with imports
	// actualComponentName := getActualComponentName(componentJS, componentName)

	titledComponentName := getTitleCase(componentName)

	if isDebugTranspile() {
		fmt.Printf("DEBUG: actualComponentName = '%s'\n", titledComponentName)
	}

	// Create the main component JS content first, wrapped in a function
	mainComponentJS := fmt.Sprintf(`
// ╔══ ⚛️  MAIN COMPONENT JS (TSX → JS) ⚛️ ══
function %s() {
%s
}

// ╔══ 📜 ORIGINAL JS CONTENT 📜 ══
%s
`, componentJS, titledComponentName, originalJS)

	// Embed component JS content into the main JS file
	embeddedJS := embedComponentJS(mainComponentJS)

	// Create the React-enhanced JS content with embedded components
	reactJS := fmt.Sprintf(`
%s

// ╔═ 💧 HYDRATION 💧 ══

// Make component available globally for hydration
window.%s = %s;

// React hydration using common utilities
try {
    // Use the global hydration function from _common.js
    window.hydrateReactApp('%s', { 
        page: window.pageData || {},
        container: 'main',
		layout: React.createElement('div', {}, 'Layout placeholder')
    });
} catch(e) {
    console.error('React hydration error:', e);
}`, embeddedJS, titledComponentName, titledComponentName, titledComponentName)

	if isDebugTranspile() {
		fmt.Printf("DEBUG: actualComponentName = '%s'\n", titledComponentName)
		fmt.Printf("DEBUG: Final reactJS content (first 500 chars): %s\n", reactJS[:min(500, len(reactJS))])
	}

	return reactJS
}

// getActualComponentName extracts the actual component name from JS content
func getActualComponentName(componentJS, componentName string) string {
	// Look for function declarations in the JS content
	// Pattern: function ComponentName( or export default function ComponentName(
	funcPattern := regexp.MustCompile(`(?:export\s+default\s+)?function\s+([A-Z][a-zA-Z0-9]*)`)
	matches := funcPattern.FindStringSubmatch(componentJS)
	if len(matches) > 1 {
		return matches[1]
	}

	// Fallback to the provided component name, capitalized
	return strings.Title(componentName)
}

// embedComponentJS embeds component JS content into the main JS file
func embedComponentJS(jsContent string) string {
	// This function processes the JS content to embed any component references
	// For now, we'll return the content as-is, but this could be enhanced
	// to handle component imports and references

	// Remove any existing component embedding patterns
	jsContent = regexp.MustCompile(`// Component JS.*?// End Component JS`).ReplaceAllString(jsContent, "")

	return jsContent
}

// removeTypeScriptTypes removes TypeScript type annotations from TSX content
func removeTypeScriptTypes(tsxContent string) string {
	// Remove type annotations from function parameters
	// Pattern: (param: Type) -> (param)
	tsxContent = regexp.MustCompile(`(\w+):\s*[A-Za-z][A-Za-z0-9<>\[\]|&]*`).ReplaceAllString(tsxContent, "$1")

	// Remove return type annotations
	// Pattern: ): ReturnType -> )
	tsxContent = regexp.MustCompile(`\):\s*[A-Za-z][A-Za-z0-9<>\[\]|&]*`).ReplaceAllString(tsxContent, ")")

	// Remove interface declarations
	tsxContent = regexp.MustCompile(`interface\s+\w+\s*\{[^}]*\}`).ReplaceAllString(tsxContent, "")

	// Remove type declarations
	tsxContent = regexp.MustCompile(`type\s+\w+\s*=\s*[^;]+;`).ReplaceAllString(tsxContent, "")

	return tsxContent
}

// handleReactFragments converts React Fragment syntax to React.Fragment
func handleReactFragments(jsxContent string) string {
	// Replace React Fragment syntax <>...</> with <React.Fragment>...</React.Fragment>
	// This makes it easier for the HTML parser to handle

	// First, replace opening fragments
	jsxContent = regexp.MustCompile(`<>`).ReplaceAllString(jsxContent, `<React.Fragment>`)

	// Then, replace closing fragments
	jsxContent = regexp.MustCompile(`</>`).ReplaceAllString(jsxContent, `</React.Fragment>`)

	return jsxContent
}

// fixAttributeCases fixes attribute case issues caused by HTML parser
func fixAttributeCases(jsContent string) string {
	if isDebugTranspile() {
		fmt.Printf("DEBUG: fixAttributeCases input: %s\n", jsContent[:min(200, len(jsContent))])
	}

	// Fix common attribute case issues
	jsContent = strings.ReplaceAll(jsContent, "{classname:", "{className:")
	jsContent = strings.ReplaceAll(jsContent, "classname:", "className:")
	jsContent = strings.ReplaceAll(jsContent, "{htmlfor:", "{htmlFor:")
	jsContent = strings.ReplaceAll(jsContent, "htmlfor:", "htmlFor:")
	jsContent = strings.ReplaceAll(jsContent, "{onclick:", "{onClick:")
	jsContent = strings.ReplaceAll(jsContent, "onclick:", "onClick:")
	jsContent = strings.ReplaceAll(jsContent, "{suppresshydrationwarning:", "{suppressHydrationWarning:")
	jsContent = strings.ReplaceAll(jsContent, "suppresshydrationwarning:", "suppressHydrationWarning:")
	jsContent = strings.ReplaceAll(jsContent, "{initialvalue:", "{initialValue:")
	jsContent = strings.ReplaceAll(jsContent, "initialvalue:", "initialValue:")

	// Fix data attributes (convert camelCase back to kebab-case)
	jsContent = strings.ReplaceAll(jsContent, "{datatest:", "{data-test:")
	jsContent = strings.ReplaceAll(jsContent, "datatest:", "data-test:")
	jsContent = strings.ReplaceAll(jsContent, "{dataid:", "{data-id:")
	jsContent = strings.ReplaceAll(jsContent, "dataid:", "data-id:")

	// Fix other common attributes
	jsContent = strings.ReplaceAll(jsContent, "{type:", "{type:")
	jsContent = strings.ReplaceAll(jsContent, "type:", "type:")
	jsContent = strings.ReplaceAll(jsContent, "{id:", "{id:")
	jsContent = strings.ReplaceAll(jsContent, "id:", "id:")
	jsContent = strings.ReplaceAll(jsContent, "{name:", "{name:")
	jsContent = strings.ReplaceAll(jsContent, "name:", "name:")
	jsContent = strings.ReplaceAll(jsContent, "{src:", "{src:")
	jsContent = strings.ReplaceAll(jsContent, "src:", "src:")
	jsContent = strings.ReplaceAll(jsContent, "{alt:", "{alt:")
	jsContent = strings.ReplaceAll(jsContent, "alt:", "alt:")
	jsContent = strings.ReplaceAll(jsContent, "{href:", "{href:")
	jsContent = strings.ReplaceAll(jsContent, "href:", "href:")
	jsContent = strings.ReplaceAll(jsContent, "{checked:", "{checked:")
	jsContent = strings.ReplaceAll(jsContent, "checked:", "checked:")
	jsContent = strings.ReplaceAll(jsContent, "{disabled:", "{disabled:")
	jsContent = strings.ReplaceAll(jsContent, "disabled:", "disabled:")

	if isDebugTranspile() {
		fmt.Printf("DEBUG: fixAttributeCases output: %s\n", jsContent[:min(200, len(jsContent))])
	}

	return jsContent
}

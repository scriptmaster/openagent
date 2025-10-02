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

// TextPart represents a part of text content - either a string or an expression
type TextPart struct {
	Type  string // "string" or "expression"
	Value string
}

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
	return TSX2JSWithOptions(tsxStr, false)
}

// TSX2JSWithOptions converts TSX to JavaScript with options
func TSX2JSWithOptions(tsxStr string, isInnerComponent bool) string {
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

			// Convert JSX to React.createElement calls using the full pipeline
			jsxStr := parseJSXWithHTMLParser(mainContent)

			// Fix attribute case issues
			jsxStr = fixAttributeCases(jsxStr)

			// Clean up extra whitespace
			jsxStr = regexp.MustCompile(`\n\s*\n`).ReplaceAllString(jsxStr, "\n")
			jsxStr = strings.TrimSpace(jsxStr)

			// Extract script content from the function body (before return statement) - only for inner components
			var scriptContent string
			if isInnerComponent {
				scriptContent = extractScriptContentFromTSX(tsxStr)

				if isDebugTranspile() {
					fmt.Printf("DEBUG: extractScriptContentFromTSX result: '%s'\n", scriptContent)
				}

				// Combine script content with JSX in a proper function structure (only for inner components)
				if scriptContent != "" {
					// Extract component name from TSX
					componentName := extractComponentNameFromTSX(tsxStr)
					jsxStr = fmt.Sprintf(`function %s({page}) {
    %s
    return (
        %s
    );
}`, componentName, scriptContent, jsxStr)
				}
			}

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

	// Convert JSX to React.createElement calls using the full pipeline
	jsxStr := parseJSXWithHTMLParser(mainContent)

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

	// Preprocess JSX expressions to make them HTML-parser friendly
	if isDebugTranspile() {
		fmt.Printf("DEBUG: Before preprocessJSXExpressions: %s\n", jsxContent[:min(100, len(jsxContent))])
	}
	jsxContent = preprocessJSXExpressions(jsxContent)
	if isDebugTranspile() {
		fmt.Printf("DEBUG: After preprocessJSXExpressions: %s\n", jsxContent[:min(100, len(jsxContent))])
	}

	// Fix custom JSX self-closing tags before parsing
	jsxContent = fixCustomJSXSelfClosingTags(jsxContent)

	// Convert ALL JSX elements (HTML + custom components) to React.createElement calls
	// This processes the complete structure using HTML parser and node walker
	jsxContent = convertJSXToReactCreateElement(jsxContent)

	// All processing is now done in convertJSXToReactCreateElement
	return jsxContent
}

// recWalkHTMLNodeWithCustomComponents walks through HTML nodes and converts them to React.createElement calls
func recWalkHTMLNodeWithCustomComponents(n *html.Node, result *strings.Builder, customComponents map[string]string) {
	if n == nil {
		return
	}

	switch n.Type {
	case html.ElementNode:
		// Convert element to React.createElement
		recConvertElementToReactWithCustomComponents(n, result, customComponents)
	case html.TextNode:
		// Convert text node - preserve meaningful whitespace to avoid hydration mismatch
		text := n.Data // strings.TrimSpace(n.Data)
		if isDebugTranspile() {
			fmt.Printf("DEBUG: TextNode content: '%s'\n", text)
		}
		// Only preserve whitespace if it's not just indentation (tabs/spaces at start of line)
		if text != "" && !isOnlyIndentation(n.Data) && strings.TrimSpace(text) != "" {
			// Check if this text is a React.createElement call (from our conversion)
			trimmedText := strings.TrimSpace(text)
			if strings.HasPrefix(trimmedText, "React.createElement(") {
				// This is a React.createElement call, trim whitespace and don't wrap it in quotes
				result.WriteString(trimmedText)
			} else {
				// This is regular text, process JSX interpolations
				processedText := processJSXInterpolations(text)
				if isDebugTranspile() {
					fmt.Printf("DEBUG: After processJSXInterpolations: '%s'\n", processedText)
				}

				// If the text contains interpolations, it's already properly formatted
				if strings.Contains(processedText, " + ") {
					// Text has interpolations, it's already properly formatted
					result.WriteString(processedText)
				} else {
					// No interpolations, escape and wrap in quotes
					escapedText := escapeJSString(processedText)
					if isDebugTranspile() {
						fmt.Printf("DEBUG: After escapeJSString: '%s'\n", escapedText)
					}
					result.WriteString(fmt.Sprintf("'%s'", escapedText))
				}
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
	if strings.ToLower(n.Data) == "react.fragment" {
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
		isCustomComponent = isCustomReactComponentCaseInsensitive(n.Data)
		if isDebugTranspile() {
			fmt.Printf("DEBUG: No customComponents map, using isCustomReactComponentCaseInsensitive: %s -> %v\n", n.Data, isCustomComponent)
		}
		// Capitalize component name for custom components
		if isCustomComponent {
			componentName = strings.Title(n.Data)
		}
	}

	// Build props object
	props := buildPropsObject(n)

	// Ensure props object has braces if it's not "null"
	if props != "null" && !strings.HasPrefix(props, "{") {
		props = "{" + props + "}"
		if isDebugTranspile() {
			fmt.Printf("DEBUG: Added braces to props: %s\n", props)
		}
	}

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

	// For custom components, check if they should be self-closing
	// If a custom component has children but they are actually siblings that got parsed incorrectly,
	// we should treat it as self-closing
	if isCustomComponent && hasChildren {
		// Check if the children are actually siblings that got parsed incorrectly
		// This happens when the HTML parser doesn't recognize self-closing custom components
		childrenStr := children.String()
		if strings.Contains(childrenStr, "React.createElement(") && !strings.Contains(childrenStr, ",") {
			// This looks like a single child that should actually be a sibling
			// Treat the component as self-closing
			hasChildren = false
		}
	}

	// Build React.createElement call
	if isDebugTranspile() {
		fmt.Printf("DEBUG: Building React.createElement for %s with props: %s\n", componentName, props)
	}

	// Ensure props are properly wrapped in braces
	if props != "null" && !strings.HasPrefix(props, "{") {
		props = "{" + props + "}"
		if isDebugTranspile() {
			fmt.Printf("DEBUG: Fixed props braces: %s\n", props)
		}
	}

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
	if isDebugTranspile() {
		fmt.Printf("DEBUG: buildPropsObject called for %s with %d attributes\n", n.Data, len(n.Attr))
	}
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

		// Handle special cases and fix attribute names
		if strings.HasPrefix(attr.Val, "__JSX_EXPR__") {
			// This is a preprocessed JSX expression - extract the value
			innerValue := strings.TrimPrefix(attr.Val, "__JSX_EXPR__")
			if isDebugTranspile() {
				fmt.Printf("DEBUG: Found JSX expression: %s -> %s\n", attr.Val, innerValue)
			}
			// Fix attribute name for JS expressions
			fixedPropName := propName
			if propName == "suppresshydrationwarning" {
				fixedPropName = "suppressHydrationWarning"
			} else if propName == "onclick" {
				fixedPropName = "onClick"
			} else if propName == "classname" {
				fixedPropName = "className"
			}
			props.WriteString(fmt.Sprintf("%s: %s", fixedPropName, innerValue))
		} else if propName == "class" {
			props.WriteString(fmt.Sprintf("className: \"%s\"", attr.Val))
		} else if propName == "for" {
			props.WriteString(fmt.Sprintf("htmlFor: \"%s\"", attr.Val))
		} else if propName == "suppresshydrationwarning" {
			props.WriteString(fmt.Sprintf("suppressHydrationWarning: \"%s\"", attr.Val))
		} else if propName == "onclick" {
			props.WriteString(fmt.Sprintf("onClick: \"%s\"", attr.Val))
		} else if propName == "classname" {
			props.WriteString(fmt.Sprintf("className: \"%s\"", attr.Val))
		} else if strings.HasPrefix(attr.Val, "{") && strings.HasSuffix(attr.Val, "}") {
			// Already a JS expression - remove the braces since we're in a JS object
			innerValue := strings.TrimPrefix(strings.TrimSuffix(attr.Val, "}"), "{")
			// Fix attribute name for JS expressions too
			fixedPropName := propName
			if propName == "suppresshydrationwarning" {
				fixedPropName = "suppressHydrationWarning"
			} else if propName == "onclick" {
				fixedPropName = "onClick"
			} else if propName == "classname" {
				fixedPropName = "className"
			}
			props.WriteString(fmt.Sprintf("%s: %s", fixedPropName, innerValue))
		} else if (attr.Val == "true" || attr.Val == "false") && !strings.HasPrefix(propName, "aria-") {
			// Boolean values (but not for ARIA attributes)
			fixedPropName := propName
			if propName == "suppresshydrationwarning" {
				fixedPropName = "suppressHydrationWarning"
			} else if propName == "onclick" {
				fixedPropName = "onClick"
			} else if propName == "classname" {
				fixedPropName = "className"
			}
			props.WriteString(fmt.Sprintf("%s: %s", fixedPropName, attr.Val))
		} else if propName == "checked" || propName == "disabled" {
			// Boolean attributes - these should be treated as boolean values
			props.WriteString(fmt.Sprintf("%s: true", propName))
		} else {
			// Regular string value - fix attribute name
			fixedPropName := propName
			if propName == "suppresshydrationwarning" {
				fixedPropName = "suppressHydrationWarning"
			} else if propName == "onclick" {
				fixedPropName = "onClick"
			} else if propName == "classname" {
				fixedPropName = "className"
			}

			// Check if this is a kebab-case attribute (data-*, aria-*)
			if strings.HasPrefix(propName, "data-") || strings.HasPrefix(propName, "aria-") {
				props.WriteString(fmt.Sprintf("\"%s\": \"%s\"", fixedPropName, attr.Val))
			} else {
				props.WriteString(fmt.Sprintf("%s: \"%s\"", fixedPropName, attr.Val))
			}
		}
	}

	props.WriteString("}")
	result := props.String()

	// Ensure the result always has braces (double-check)
	if result != "null" && !strings.HasPrefix(result, "{") {
		result = "{" + result + "}"
		if isDebugTranspile() {
			fmt.Printf("DEBUG: buildPropsObject added missing braces: %s\n", result)
		}
	}

	if isDebugTranspile() {
		fmt.Printf("DEBUG: buildPropsObject returning: %s\n", result)
	}
	return result
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

// isCustomReactComponentCaseInsensitive checks if a component name is a custom React component (case insensitive)
func isCustomReactComponentCaseInsensitive(componentName string) bool {
	// Known custom components (case insensitive) - only actual custom components, not HTML elements
	customComponents := map[string]bool{
		"simple":  true,
		"counter": true,
		"modal":   true,
		"card":    true,
		"list":    true,
		"item":    true,
	}

	// Check if it's a known custom component
	return customComponents[strings.ToLower(componentName)]
}

// isOnlyIndentation checks if text contains only indentation characters
func isOnlyIndentation(text string) bool {
	// Check if the text contains only whitespace characters
	if text == "" {
		return true
	}
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

// preprocessJSXExpressions converts JSX expressions to HTML-parser friendly format
func preprocessJSXExpressions(jsxContent string) string {
	// Convert JSX expressions like {true} to a special format that HTML parser can handle
	// We'll use a special attribute format that we can detect later
	jsxExpressionPattern := regexp.MustCompile(`(\w+)=\{([^}]+)\}`)

	result := jsxExpressionPattern.ReplaceAllStringFunc(jsxContent, func(match string) string {
		// Extract attribute name and value
		parts := jsxExpressionPattern.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match
		}

		attrName := parts[1]
		attrValue := parts[2]

		// Convert to a special format that we can detect in buildPropsObject
		return fmt.Sprintf(`%s="__JSX_EXPR__%s"`, attrName, attrValue)
	})

	if isDebugTranspile() {
		fmt.Printf("DEBUG: preprocessJSXExpressions result: %s\n", result[:min(100, len(result))])
	}

	return result
}

// fixCustomJSXSelfClosingTags fixes custom JSX components to be self-closing
func fixCustomJSXSelfClosingTags(htmlContent string) string {
	if isDebugTranspile() {
		fmt.Printf("DEBUG: fixCustomJSXSelfClosingTags input: %s\n", htmlContent[:min(200, len(htmlContent))])
	}

	// First, fix already self-closing custom components to be HTML-parser friendly
	// Convert <Component /> to <Component></Component> so HTML parser treats it as self-contained
	selfClosingPattern := regexp.MustCompile(`<([A-Z][a-zA-Z0-9]*)\s+([^>]*?)\s*/>`)
	htmlContent = selfClosingPattern.ReplaceAllStringFunc(htmlContent, func(match string) string {
		submatches := selfClosingPattern.FindStringSubmatch(match)
		if len(submatches) >= 3 {
			componentName := submatches[1]
			attributes := strings.TrimSpace(submatches[2])
			// Convert to opening and closing tags so HTML parser treats it as self-contained
			return fmt.Sprintf("<%s %s></%s>", componentName, attributes, componentName)
		}
		return match
	})

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
				return fmt.Sprintf("<%s%s></%s>", componentName, attributes, componentName)
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

// convertJSXToReactCreateElement converts ALL JSX elements (HTML + custom components) to React.createElement calls
// This function uses HTML parser and node walker for reliable parsing
func convertJSXToReactCreateElement(jsxContent string) string {
	if isDebugTranspile() {
		fmt.Printf("DEBUG: convertJSXToReactCreateElement called with: %s\n", jsxContent[:min(100, len(jsxContent))])
	}

	// Parse the JSX content as HTML
	doc, err := html.Parse(strings.NewReader(jsxContent))
	if err != nil {
		if isDebugTranspile() {
			fmt.Printf("DEBUG: HTML parse error in convertJSXToReactCreateElement: %v\n", err)
		}
		// If parsing fails, return the original content
		return jsxContent
	}

	// Process ALL elements (HTML + custom components) using the existing walker
	var result strings.Builder
	recWalkHTMLNodeWithCustomComponents(doc, &result, nil)

	if isDebugTranspile() {
		fmt.Printf("DEBUG: convertJSXToReactCreateElement result: %s\n", result.String()[:min(100, len(result.String()))])
	}

	return result.String()
}

// convertJSXComponentsWalkerWithArray walks through HTML nodes and collects custom components in an array
func convertJSXComponentsWalkerWithArray(n *html.Node, components *[]string) {
	if n == nil {
		return
	}

	// Check if this is a custom component (starts with uppercase letter)
	if n.Type == html.ElementNode && isCustomReactComponentCaseInsensitive(n.Data) {
		if isDebugTranspile() {
			fmt.Printf("DEBUG: Found custom component: %s\n", n.Data)
		}

		// Build props object using the same logic as buildPropsObject
		props := buildPropsObject(n)

		// Create React.createElement call - capitalize component name for custom components
		componentName := strings.Title(n.Data)
		reactCall := fmt.Sprintf("React.createElement(%s, %s)", componentName, props)
		*components = append(*components, reactCall)
	}

	// Process children recursively for all nodes
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		convertJSXComponentsWalkerWithArray(c, components)
	}
}

// convertJSXComponentsWalker walks through HTML nodes and converts custom components to React.createElement
func convertJSXComponentsWalker(n *html.Node, result *strings.Builder) {
	if n == nil {
		return
	}

	// Check if this is a custom component (starts with uppercase letter)
	if n.Type == html.ElementNode && isCustomReactComponentCaseInsensitive(n.Data) {
		if isDebugTranspile() {
			fmt.Printf("DEBUG: Found custom component: %s\n", n.Data)
		}

		// Build props object using the same logic as buildPropsObject
		props := buildPropsObject(n)

		// Create React.createElement call - capitalize component name for custom components
		componentName := strings.Title(n.Data)
		reactCall := fmt.Sprintf("React.createElement(%s, %s)", componentName, props)
		result.WriteString(reactCall)
		// Don't return here - continue processing siblings
	}

	// Process children recursively for all nodes
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		convertJSXComponentsWalker(c, result)
	}
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
	originalContent := jsContent
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
		if originalContent != jsContent {
			// Find what changed
			for i := 0; i < min(len(originalContent), len(jsContent)); i++ {
				if i >= len(originalContent) || i >= len(jsContent) || originalContent[i] != jsContent[i] {
					fmt.Printf("DEBUG: First difference at position %d\n", i)
					fmt.Printf("DEBUG: Original (20 chars): %s\n", originalContent[max(0, i-10):min(len(originalContent), i+10)])
					fmt.Printf("DEBUG: Modified (20 chars): %s\n", jsContent[max(0, i-10):min(len(jsContent), i+10)])
					break
				}
			}
		}
	}

	return jsContent
}

// extractScriptContentFromTSX extracts script content from TSX function body
func extractScriptContentFromTSX(tsxStr string) string {
	// Find the function body content between the opening brace and return statement
	funcStart := strings.Index(tsxStr, "function")
	if funcStart == -1 {
		return ""
	}

	// Find the opening brace of the function
	braceStart := strings.Index(tsxStr[funcStart:], "{")
	if braceStart == -1 {
		return ""
	}
	braceStart += funcStart + 1

	// Find the return statement
	returnStart := strings.Index(tsxStr, "return (")
	if returnStart == -1 {
		return ""
	}

	// Extract content between opening brace and return statement
	scriptContent := tsxStr[braceStart:returnStart]

	// Clean up the script content
	scriptContent = strings.TrimSpace(scriptContent)

	// Remove the comment header if present
	if strings.Contains(scriptContent, "COMPONENT <script> TAG CONTENTS") {
		lines := strings.Split(scriptContent, "\n")
		var cleanLines []string
		inScript := false
		for _, line := range lines {
			if strings.Contains(line, "COMPONENT <script> TAG CONTENTS") {
				inScript = true
				continue
			}
			if inScript && strings.TrimSpace(line) != "" {
				cleanLines = append(cleanLines, line)
			}
		}
		scriptContent = strings.Join(cleanLines, "\n")
	}

	return strings.TrimSpace(scriptContent)
}

// parseTextInterpolation parses text content and returns an array of TextPart structs
func parseTextInterpolation(text string) []TextPart {
	var parts []TextPart

	// Find all JSX interpolations like {variable}
	interpolationPattern := regexp.MustCompile(`\{([^}]+)\}`)
	lastIndex := 0

	matches := interpolationPattern.FindAllStringSubmatchIndex(text, -1)

	for _, match := range matches {
		// Add string part before the interpolation
		if match[0] > lastIndex {
			stringPart := text[lastIndex:match[0]]
			if isDebugTranspile() {
				fmt.Printf("DEBUG: parseTextInterpolation - stringPart: '%s'\n", stringPart)
			}
			if stringPart != "" {
				parts = append(parts, TextPart{Type: "string", Value: stringPart})
			}
		}

		// Add expression part
		expression := text[match[2]:match[3]] // Extract content between { and }
		parts = append(parts, TextPart{Type: "expression", Value: expression})

		lastIndex = match[1] // End of the interpolation
	}

	// Add remaining string part
	if lastIndex < len(text) {
		stringPart := text[lastIndex:]
		if isDebugTranspile() {
			fmt.Printf("DEBUG: parseTextInterpolation - remaining stringPart: '%s'\n", stringPart)
		}
		// Always add the remaining part, even if it's empty (for proper concatenation)
		parts = append(parts, TextPart{Type: "string", Value: stringPart})
	} else if len(matches) > 0 {
		// If we had interpolations but no remaining text, add an empty string part
		parts = append(parts, TextPart{Type: "string", Value: ""})
	}

	// If no interpolations found, treat the whole text as a string
	if len(parts) == 0 && strings.TrimSpace(text) != "" {
		parts = append(parts, TextPart{Type: "string", Value: text})
	}

	return parts
}

// buildTextConcatenation builds JavaScript string concatenation from TextPart array
func buildTextConcatenation(parts []TextPart) string {
	if len(parts) == 0 {
		return ""
	}

	if len(parts) == 1 {
		if parts[0].Type == "string" {
			return fmt.Sprintf("'%s'", escapeJSString(parts[0].Value))
		} else {
			// Single expression - don't wrap in quotes, just return the expression
			return parts[0].Value
		}
	}

	var result strings.Builder
	for i, part := range parts {
		if i > 0 {
			result.WriteString(" + ")
		}

		if part.Type == "string" {
			result.WriteString(fmt.Sprintf("'%s'", escapeJSString(part.Value)))
		} else {
			result.WriteString(fmt.Sprintf("(%s)", part.Value))
		}
	}

	return result.String()
}

// processJSXInterpolations processes JSX interpolations using the struct-based approach
func processJSXInterpolations(jsxContent string) string {
	// First normalize whitespace on the entire content
	normalizedContent := normalizeWhitespace(jsxContent)

	// Check if the text contains interpolations
	if !strings.Contains(normalizedContent, "{") {
		// No interpolations, return as-is (just escape for JavaScript)
		return escapeJSString(normalizedContent)
	}

	// Parse the text into parts
	parts := parseTextInterpolation(normalizedContent)

	// Build the concatenation
	result := buildTextConcatenation(parts)

	if isDebugTranspile() {
		fmt.Printf("DEBUG: processJSXInterpolations: '%s' -> '%s'\n", jsxContent, result)
	}

	return result
}

// trimWhiteSpace removes all whitespace characters from a string
func trimWhiteSpace(s string) string {
	// Remove all types of whitespace: spaces, tabs, newlines, carriage returns, non-breaking spaces
	result := strings.ReplaceAll(s, " ", "")
	result = strings.ReplaceAll(result, "\t", "")
	result = strings.ReplaceAll(result, "\n", "")
	result = strings.ReplaceAll(result, "\r", "")
	result = strings.ReplaceAll(result, "\u00A0", "") // non-breaking space
	result = strings.ReplaceAll(result, "\u2000", "") // en quad
	result = strings.ReplaceAll(result, "\u2001", "") // em quad
	result = strings.ReplaceAll(result, "\u2002", "") // en space
	result = strings.ReplaceAll(result, "\u2003", "") // em space
	result = strings.ReplaceAll(result, "\u2004", "") // three-per-em space
	result = strings.ReplaceAll(result, "\u2005", "") // four-per-em space
	result = strings.ReplaceAll(result, "\u2006", "") // six-per-em space
	result = strings.ReplaceAll(result, "\u2007", "") // figure space
	result = strings.ReplaceAll(result, "\u2008", "") // punctuation space
	result = strings.ReplaceAll(result, "\u2009", "") // thin space
	result = strings.ReplaceAll(result, "\u200A", "") // hair space
	result = strings.ReplaceAll(result, "\u202F", "") // narrow no-break space
	result = strings.ReplaceAll(result, "\u205F", "") // medium mathematical space
	result = strings.ReplaceAll(result, "\u3000", "") // ideographic space
	return result
}

// normalizeWhitespace removes leading/trailing whitespace and normalizes internal whitespace
func normalizeWhitespace(s string) string {
	// First handle literal \n characters (not actual newlines)
	result := strings.ReplaceAll(s, "\\n", " ")
	// Replace multiple consecutive whitespace characters with single spaces
	result = regexp.MustCompile(`\s+`).ReplaceAllString(result, " ")
	// Trim leading and trailing whitespace
	return strings.TrimSpace(result)
}

// extractComponentNameFromTSX extracts the component name from TSX content
func extractComponentNameFromTSX(tsxStr string) string {
	// Look for "export default function ComponentName(" pattern
	pattern := regexp.MustCompile(`export default function (\w+)\s*\(`)
	matches := pattern.FindStringSubmatch(tsxStr)
	if len(matches) > 1 {
		return matches[1]
	}

	// Fallback: look for "function ComponentName(" pattern
	pattern = regexp.MustCompile(`function (\w+)\s*\(`)
	matches = pattern.FindStringSubmatch(tsxStr)
	if len(matches) > 1 {
		return matches[1]
	}

	return "Component" // fallback
}

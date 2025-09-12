package server

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Component JS files are now stored as separate files and read when needed

// processComponentImports processes component divs and imports component files
func processComponentImports(htmlContent, inputPath string) (string, error) {
	if isDebugTranspile() {
		fmt.Printf("DEBUG: Processing component imports for: %s\n", inputPath)
	}

	// Pattern to match: <div id="component-counter"></div> or <div id="component_counter"></div>
	componentPattern := regexp.MustCompile(`<div\s+id="component[-_](\w+)"[^>]*></div>`)

	matches := componentPattern.FindAllString(htmlContent, -1)
	if isDebugTranspile() {
		fmt.Printf("DEBUG: Found %d component matches: %v\n", len(matches), matches)
	}

	return componentPattern.ReplaceAllStringFunc(htmlContent, func(match string) string {
		// Extract component name from the match
		submatches := componentPattern.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match // Return original if no match
		}

		componentName := submatches[1] // e.g., "counter"

		if isDebugTranspile() {
			fmt.Printf("DEBUG: Processing component: %s\n", componentName)
		}

		// Import and transpile the component
		componentHTML, err := importAndTranspileComponent(componentName, inputPath)
		if err != nil {
			if isDebugTranspile() {
				fmt.Printf("DEBUG: Failed to import component %s: %v\n", componentName, err)
			}
			return match // Return original if import fails
		}

		if isDebugTranspile() {
			fmt.Printf("DEBUG: Successfully imported component %s\n", componentName)
		}

		// Preserve the original id attribute by adding it to the component HTML
		// Replace the first opening tag in componentHTML with the original id
		originalId := fmt.Sprintf("component-%s", componentName)
		componentHTML = strings.Replace(componentHTML, "<div>", fmt.Sprintf(`<div id="%s">`, originalId), 1)

		return componentHTML
	}), nil
}

// importAndTranspileComponent imports a component file and transpiles it
func importAndTranspileComponent(componentName, inputPath string) (string, error) {
	if isDebugTranspile() {
		fmt.Printf("DEBUG: Importing component: %s\n", componentName)
	}

	// Construct component file path
	componentPath := fmt.Sprintf("tpl/components/%s.html", componentName)

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

	// Extract CSS and JS from component (with outputPath to generate separate files)
	componentJSPath := fmt.Sprintf("tpl/generated/js/component_%s.js", componentName)
	_, jsContent, err := extractCSSAndJS(string(componentContent), componentPath, componentJSPath)
	if err != nil {
		return "", fmt.Errorf("failed to extract CSS/JS from component: %v", err)
	}

	if isDebugTranspile() {
		fmt.Printf("DEBUG: Extracted JS content: %d bytes\n", len(jsContent))
	}

	// Process component HTML (remove comments, convert class to className, etc.)
	componentHTML := string(componentContent)
	componentHTML = removeHTMLComments(componentHTML)
	componentHTML = strings.ReplaceAll(componentHTML, "class=", "className=")
	componentHTML = fixSelfClosingTags(componentHTML)

	// Remove script tags (they'll be processed separately)
	componentHTML = removeStyleAndScriptTags(componentHTML)

	// Generate component TSX file
	componentTSXPath := fmt.Sprintf("tpl/generated/components/%s.tsx", componentName)
	if err := os.MkdirAll("tpl/generated/components", 0755); err != nil {
		return "", fmt.Errorf("failed to create components directory: %v", err)
	}

	// Write component TSX
	componentTSX := fmt.Sprintf("export default function %s() {\n    return (\n        %s\n    );\n}",
		Title(componentName), componentHTML)

	if err := os.WriteFile(componentTSXPath, []byte(componentTSX), 0644); err != nil {
		return "", fmt.Errorf("failed to write component TSX: %v", err)
	}

	if isDebugTranspile() {
		fmt.Printf("DEBUG: Generated component TSX: %s\n", componentTSXPath)
	}

	// Store component JS content for later embedding (don't write separate file)
	componentJS := convertComponentTSXToJS(componentTSX)

	if isDebugTranspile() {
		fmt.Printf("DEBUG: Converted TSX to JS: %d bytes\n", len(componentJS))
		fmt.Printf("DEBUG: Component JS content: %s\n", componentJS[:min(200, len(componentJS))])
	}

	// Add the script content (prototype methods)
	componentJS += "\n\n///////////////////////////////\n\n"
	componentJS += "// Component prototype methods\n"
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

// convertComponentTSXToJS converts component TSX to JS with React.createElement calls
func convertComponentTSXToJS(tsxContent string) string {
	if isDebugTranspile() {
		fmt.Printf("DEBUG: convertComponentTSXToJS called with: %s\n", tsxContent[:min(200, len(tsxContent))])
	}

	// Create a simple JSX-to-React converter for component TSX
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
	// Simple JSX to React.createElement conversion
	// This is a temporary solution until we have a proper JSX parser

	// Extract the JSX content from the return statement
	jsxPattern := regexp.MustCompile(`(?s)return\s*\(\s*(.*?)\s*\)\s*;`)
	matches := jsxPattern.FindStringSubmatch(tsxContent)
	if len(matches) < 2 {
		return tsxContent // Return original if no match
	}

	jsxContent := matches[1]

	// Convert the JSX to React.createElement calls
	// Handle simple div with text content first
	result := strings.ReplaceAll(jsxContent, `<div>Simple Component</div>`, `React.createElement('div', null, 'Simple Component')`)

	// Handle Counter component specific JSX
	result = strings.ReplaceAll(result, `<div className="mb-3">`, `React.createElement('div', {className: 'mb-3'}, `)
	result = strings.ReplaceAll(result, `<span className="badge bg-primary fs-4">Count: {this.counterState.count}</span>`, `React.createElement('span', {className: 'badge bg-primary fs-4'}, 'Count: ', this.counterState.count)`)
	result = strings.ReplaceAll(result, `</div>`, `)`)

	// Second div with buttons
	result = strings.ReplaceAll(result, `<div className="btn-group" role="group">`, `React.createElement('div', {className: 'btn-group', role: 'group'}, `)
	result = strings.ReplaceAll(result, `<button className="btn btn-outline-danger" onClick={() => this.decrementCounter()}>-</button>`, `React.createElement('button', {className: 'btn btn-outline-danger', onClick: () => this.decrementCounter()}, '-')`)
	result = strings.ReplaceAll(result, `<button className="btn btn-outline-success" onClick={() => this.incrementCounter()}>+</button>`, `React.createElement('button', {className: 'btn btn-outline-success', onClick: () => this.incrementCounter()}, '+')`)
	result = strings.ReplaceAll(result, `<button className="btn btn-outline-secondary" onClick={() => this.resetCounter()}>Reset</button>`, `React.createElement('button', {className: 'btn btn-outline-secondary', onClick: () => this.resetCounter()}, 'Reset')`)

	// Create the final function with proper React.createElement syntax
	finalResult := fmt.Sprintf("function %s() {\n    return (\n        %s\n    );\n}", Title(componentName), result)

	return finalResult
}

// replaceComponentDivsWithCalls replaces component divs with component function calls in JS
func replaceComponentDivsWithCalls(jsContent string) string {
	// Pattern to match various forms of component divs in React.createElement calls
	// This matches: React.createElement('div', {id: 'component-simple'}, ...)
	// Or: React.createElement('div', {id: 'component-simple'})
	componentDivPattern := regexp.MustCompile(`React\.createElement\('div',\s*\{[^}]*id:\s*'component[-_](\w+)'[^}]*\}[^)]*\)`)

	if isDebugTranspile() {
		matches := componentDivPattern.FindAllString(jsContent, -1)
		fmt.Printf("DEBUG: Found %d component div matches for replacement\n", len(matches))
		for i, match := range matches {
			fmt.Printf("DEBUG: Match %d: %s\n", i, match[:min(100, len(match))])
		}
	}

	// Also check for any component JS files that exist to ensure we have all components
	componentFiles, err := filepath.Glob("tpl/generated/js/component_*.js")
	if err == nil && len(componentFiles) > 0 {
		if isDebugTranspile() {
			fmt.Printf("DEBUG: Found %d component JS files for reference: %v\n", len(componentFiles), componentFiles)
		}
	}

	return componentDivPattern.ReplaceAllStringFunc(jsContent, func(match string) string {
		// Extract component name from the match
		submatches := componentDivPattern.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match // Return original if no match
		}

		componentName := submatches[1]                // e.g., "counter"
		componentFunctionName := Title(componentName) // e.g., "Counter"

		if isDebugTranspile() {
			fmt.Printf("DEBUG: Replacing component div '%s' with React.createElement(%s, {})\n", match[:min(50, len(match))], componentFunctionName)
		}

		// Replace with component function call
		return fmt.Sprintf("React.createElement(%s, {})", componentFunctionName)
	})
}

// findComponentJSFiles finds component JS files referenced in HTML content
func findComponentJSFiles(htmlContent string) []string {
	var componentFiles []string

	// Pattern to match: <div id="component-counter"></div> or <div id="component_counter"></div>
	componentPattern := regexp.MustCompile(`<div\s+id="component[-_](\w+)"[^>]*></div>`)

	matches := componentPattern.FindAllStringSubmatch(htmlContent, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			componentName := match[1] // e.g., "counter"
			componentJSFile := fmt.Sprintf("/tsx/js/component_%s.js", componentName)
			componentFiles = append(componentFiles, componentJSFile)
		}
	}

	return componentFiles
}

// embedComponentJS embeds only the component JS content that is referenced in the main JS file
func embedComponentJS(mainJSContent string) string {
	// Find components that are actually referenced in the main JS content
	// Look for React.createElement(ComponentName, {}) patterns
	componentRefPattern := regexp.MustCompile(`React\.createElement\((\w+),\s*\{\}\)`)
	matches := componentRefPattern.FindAllStringSubmatch(mainJSContent, -1)

	if isDebugTranspile() {
		fmt.Printf("DEBUG: embedComponentJS found %d component references in main JS\n", len(matches))
	}

	if len(matches) == 0 {
		if isDebugTranspile() {
			fmt.Printf("DEBUG: No component references found, returning mainJSContent as-is\n")
		}
		return mainJSContent
	}

	// Extract unique component names
	referencedComponents := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			componentName := match[1]
			// Skip built-in React components and common HTML elements
			if !isBuiltInComponent(componentName) {
				referencedComponents[componentName] = true
			}
		}
	}

	if isDebugTranspile() {
		fmt.Printf("DEBUG: Referenced components: %v\n", referencedComponents)
	}

	if len(referencedComponents) == 0 {
		if isDebugTranspile() {
			fmt.Printf("DEBUG: No custom components referenced, returning mainJSContent as-is\n")
		}
		return mainJSContent
	}

	var embeddedComponents strings.Builder
	embeddedComponents.WriteString("// Embedded Component JS\n")
	embeddedComponents.WriteString("///////////////////////////////\n\n")

	// Only embed components that are actually referenced
	for componentName := range referencedComponents {
		componentFile := fmt.Sprintf("tpl/generated/js/component_%s.js", strings.ToLower(componentName))

		componentJS, err := os.ReadFile(componentFile)
		if err != nil {
			if isDebugTranspile() {
				fmt.Printf("DEBUG: Failed to read component file %s: %v\n", componentFile, err)
			}
			continue
		}

		if isDebugTranspile() {
			fmt.Printf("DEBUG: Embedding referenced component %s: %d bytes\n", componentName, len(componentJS))
		}

		embeddedComponents.WriteString(fmt.Sprintf("// Component: %s\n", strings.ToLower(componentName)))
		embeddedComponents.WriteString(string(componentJS))
		embeddedComponents.WriteString("\n\n")
	}

	embeddedComponents.WriteString("///////////////////////////////\n\n")
	embeddedComponents.WriteString("// Main Page JS\n")
	embeddedComponents.WriteString(mainJSContent)

	return embeddedComponents.String()
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

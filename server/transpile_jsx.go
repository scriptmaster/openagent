package server

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Global regex patterns compiled once for performance
var (
	stylePattern            *regexp.Regexp
	scriptPattern           *regexp.Regexp
	removeStylePattern      *regexp.Regexp
	removeScriptPattern     *regexp.Regexp
	commentPattern          *regexp.Regexp
	hyphenUnderscorePattern *regexp.Regexp

	// Self-closing tags map for efficient lookup
	selfClosingTags map[string]*regexp.Regexp

	generateComponentName = true
)

// init initializes all regex patterns
func init() {
	stylePattern = regexp.MustCompile(`(?s)<style[^>]*>(.*?)</style>`)
	scriptPattern = regexp.MustCompile(`(?s)<script[^>]*>(.*?)</script>`)
	removeStylePattern = regexp.MustCompile(`(?s)<style[^>]*>.*?</style>`)
	removeScriptPattern = regexp.MustCompile(`(?s)<script[^>]*>.*?</script>`)
	commentPattern = regexp.MustCompile(`<!--(.*?)-->`)
	hyphenUnderscorePattern = regexp.MustCompile(`[-_]`)

	// Initialize self-closing tags map with their patterns
	selfClosingTags = map[string]*regexp.Regexp{
		"meta":   regexp.MustCompile(`<meta([^>]*?)(?:\s*/)?>`),
		"link":   regexp.MustCompile(`<link([^>]*?)(?:\s*/)?>`),
		"img":    regexp.MustCompile(`<img([^>]*?)(?:\s*/)?>`),
		"input":  regexp.MustCompile(`<input([^>]*?)(?:\s*/)?>`),
		"br":     regexp.MustCompile(`<br(?:\s*/)?>`),
		"hr":     regexp.MustCompile(`<hr([^>]*?)(?:\s*/)?>`),
		"area":   regexp.MustCompile(`<area([^>]*?)(?:\s*/)?>`),
		"base":   regexp.MustCompile(`<base([^>]*?)(?:\s*/)?>`),
		"col":    regexp.MustCompile(`<col([^>]*?)(?:\s*/)?>`),
		"embed":  regexp.MustCompile(`<embed([^>]*?)(?:\s*/)?>`),
		"source": regexp.MustCompile(`<source([^>]*?)(?:\s*/)?>`),
		"track":  regexp.MustCompile(`<track([^>]*?)(?:\s*/)?>`),
		"wbr":    regexp.MustCompile(`<wbr(?:\s*/)?>`),
	}
}

// TranspileHtmlToTsx reads an HTML file with custom comments and converts it to TSX.
func TranspileHtmlToTsx(inputPath, outputPath string) error {
	// Read the input HTML file
	content, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("could not read input file: %w", err)
	}

	// Convert Go template syntax to JSX
	htmlContent := string(content)

	// Remove DOCTYPE declaration - it will be added dynamically
	htmlContent = strings.ReplaceAll(htmlContent, "<!DOCTYPE html>", "")
	htmlContent = strings.ReplaceAll(htmlContent, "<!doctype html>", "")
	htmlContent = strings.ReplaceAll(htmlContent, "<!Doctype html>", "")

	// Extract CSS and JS content to separate files
	cssContent, jsContent, err := extractCSSAndJS(htmlContent, inputPath, outputPath)
	if err != nil {
		return fmt.Errorf("failed to extract CSS/JS: %v", err)
	}

	// Fix self-closing tags for JSX compatibility
	htmlContent = fixSelfClosingTags(htmlContent)

	// Validate HTML before further processing (temporarily disabled for debugging)
	// if err := validateHTML(htmlContent); err != nil {
	//	return fmt.Errorf("HTML validation failed for %s: %v", inputPath, err)
	// }

	// Clean up any leftover title content (but keep the variables for use in body)
	// Remove standalone title content that's not in tags
	htmlContent = strings.ReplaceAll(htmlContent, "{{.PageTitle}} - {{.AppName}}", "")

	// Remove Go template define blocks and template calls
	htmlContent = strings.ReplaceAll(htmlContent, "{{define \"", "")
	htmlContent = strings.ReplaceAll(htmlContent, "\"}}", "")
	htmlContent = strings.ReplaceAll(htmlContent, "{{template \"", "")
	htmlContent = strings.ReplaceAll(htmlContent, "\" .}}", "")
	htmlContent = strings.ReplaceAll(htmlContent, "{{end}}", "")

	// Handle template includes - for now, just remove them
	htmlContent = strings.ReplaceAll(htmlContent, "{{template \"header.html\" .}}", "")
	htmlContent = strings.ReplaceAll(htmlContent, "{{template \"footer.html\" .}}", "")
	htmlContent = strings.ReplaceAll(htmlContent, "header.html", "")
	htmlContent = strings.ReplaceAll(htmlContent, "footer.html", "")

	// Convert Go template variables to JSX
	htmlContent = strings.ReplaceAll(htmlContent, "{{.", "{page.")
	htmlContent = strings.ReplaceAll(htmlContent, "}}", "}")

	// Convert class to className
	htmlContent = strings.ReplaceAll(htmlContent, "class=", "className=")

	// Remove <style> and <script> tags (they've been extracted to separate files)
	htmlContent = removeStyleAndScriptTags(htmlContent)

	// Add links to extracted CSS and JS files
	baseName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	if cssContent != "" {
		// Add CSS link before </head> tag
		cssLink := fmt.Sprintf(`<link rel="stylesheet" href="/tsx/css/%s.css" />`, baseName)
		htmlContent = strings.Replace(htmlContent, "</head>", "\n    "+cssLink+"\n</head>", 1)
	}
	if jsContent != "" {
		// Add JS script at the end
		jsScript := fmt.Sprintf(`<script src="/tsx/js/%s.js"></script>`, baseName)
		htmlContent = htmlContent + "\n" + jsScript
	}

	// Convert HTML comments to JSX comments
	htmlContent = convertHtmlCommentsToJsx(htmlContent)

	// Clean up extra whitespace and empty lines
	htmlContent = strings.ReplaceAll(htmlContent, "\n\n\n", "\n")
	htmlContent = strings.ReplaceAll(htmlContent, "\n    \n", "\n")
	htmlContent = strings.ReplaceAll(htmlContent, "\n\n", "\n")
	htmlContent = strings.TrimSpace(htmlContent)

	// Generate the function header based on filename conventions
	fileName := filepath.Base(inputPath)

	componentName := ""
	if generateComponentName {
		componentName = strings.TrimSuffix(fileName, filepath.Ext(fileName))

		// Convert hyphens to CamelCase for valid function names
		componentName = convertToCamelCase(componentName)

		// Capitalize the first letter for the component name
		if len(componentName) > 0 {
			componentName = strings.ToUpper(string(componentName[0])) + componentName[1:]
		}

		if strings.Contains(fileName, ".landing_model.") {
			// Adjust component name if using a specific model
			componentName = strings.TrimSuffix(componentName, "landing_model")
			if len(componentName) > 0 {
				componentName = strings.ToUpper(string(componentName[0])) + componentName[1:]
			}
		}
	}

	// Write the final TSX file
	tsxContent := `export default function ` + componentName + `({page}: {page: Page}) {
    return (
` + htmlContent + `
    );
}`

	return os.WriteFile(outputPath, []byte(tsxContent), 0644)
}

// transpileAllTemplates finds and transpiles all HTML files.
func transpileAllTemplates() error {
	log.Println("Transpiling all HTML templates...")
	inputDir := "./tpl/pages"
	outputDir := "./tpl/generated"

	// Create the output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	files, err := filepath.Glob(filepath.Join(inputDir, "*.html"))
	if err != nil {
		return err
	}

	for _, file := range files {
		baseName := filepath.Base(file)
		tsxFileName := strings.TrimSuffix(baseName, ".html") + ".tsx"
		outputPath := filepath.Join(outputDir, tsxFileName)
		if err := TranspileHtmlToTsx(file, outputPath); err != nil {
			return fmt.Errorf("failed to transpile %s: %w", file, err)
		}
		log.Printf("Transpiled %s to %s", file, outputPath)
	}
	return nil
}

// fixSelfClosingTags converts HTML self-closing tags to JSX format
func fixSelfClosingTags(htmlContent string) string {
	// Use pre-compiled patterns for each self-closing tag
	for tag, pattern := range selfClosingTags {
		// Replace with self-closing version using the pre-compiled pattern
		htmlContent = pattern.ReplaceAllStringFunc(htmlContent, func(match string) string {
			// Check if it's already self-closing
			if strings.HasSuffix(match, "/>") {
				return match
			}
			// Extract attributes and make self-closing
			submatches := pattern.FindStringSubmatch(match)
			if len(submatches) > 1 {
				attrs := submatches[1]
				return fmt.Sprintf("<%s%s/>", tag, attrs)
			}
			// No attributes case
			return fmt.Sprintf("<%s/>", tag)
		})
	}

	return htmlContent
}

// extractCSSAndJS extracts <style> and <script> tags from HTML content
func extractCSSAndJS(htmlContent, inputPath, outputPath string) (string, string, error) {
	var cssContent strings.Builder
	var jsContent strings.Builder

	// Extract CSS from <style> tags (multiline)
	cssMatches := stylePattern.FindAllStringSubmatch(htmlContent, -1)
	for _, match := range cssMatches {
		if len(match) > 1 {
			cssContent.WriteString(match[1])
			cssContent.WriteString("\n")
		}
	}

	// Extract JS from <script> tags (multiline, excluding external scripts)
	jsMatches := scriptPattern.FindAllStringSubmatch(htmlContent, -1)
	for _, match := range jsMatches {
		if len(match) > 1 {
			jsContent.WriteString(match[1])
			jsContent.WriteString("\n")
		}
	}

	// Get the base filename without extension
	baseName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))

	// Write CSS file if there's content
	if cssContent.Len() > 0 {
		cssPath := filepath.Join(filepath.Dir(outputPath), "css", baseName+".css")
		if err := os.WriteFile(cssPath, []byte(cssContent.String()), 0644); err != nil {
			return "", "", fmt.Errorf("failed to write CSS file: %v", err)
		}
	}

	// Write JS file if there's content
	if jsContent.Len() > 0 {
		jsPath := filepath.Join(filepath.Dir(outputPath), "js", baseName+".js")
		if err := os.WriteFile(jsPath, []byte(jsContent.String()), 0644); err != nil {
			return "", "", fmt.Errorf("failed to write JS file: %v", err)
		}
	}

	return cssContent.String(), jsContent.String(), nil
}

// removeStyleAndScriptTags removes <style> and <script> tags from HTML content
func removeStyleAndScriptTags(htmlContent string) string {
	// Remove <style> tags and their content (multiline)
	htmlContent = removeStylePattern.ReplaceAllString(htmlContent, "")

	// Remove all <script> tags and their content (multiline, both inline and external)
	htmlContent = removeScriptPattern.ReplaceAllString(htmlContent, "")

	return htmlContent
}

// convertToCamelCase converts hyphenated strings to CamelCase
func convertToCamelCase(str string) string {
	parts := hyphenUnderscorePattern.Split(str, -1)
	if len(parts) == 1 {
		return parts[0]
	}

	result := parts[0]
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			result += strings.ToUpper(string(parts[i][0])) + parts[i][1:]
		}
	}

	return result
}

// convertHtmlCommentsToJsx converts HTML comments to JSX comments
func convertHtmlCommentsToJsx(htmlContent string) string {
	// Replace HTML comments with JSX comments
	htmlContent = commentPattern.ReplaceAllStringFunc(htmlContent, func(match string) string {
		// Extract the comment content (without <!-- and -->)
		content := commentPattern.FindStringSubmatch(match)[1]
		// Convert to JSX comment format: {/* ... */}
		return fmt.Sprintf("{/*%s*/}", content)
	})

	return htmlContent
}

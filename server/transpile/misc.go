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
// MISCELLANEOUS TRANSPILATION FUNCTIONS
// ============================================================================
// This file contains utility functions and Alpine.js related transpilation
//
// Function List:
// - isDebugTranspile() bool
//   Checks if DEBUG_TRANSPILE environment variable is set to "1"
// - convertAlpineAttributesToData(htmlContent string) string
//   Converts Alpine.js attributes to data attributes
// - convertToCamelCase(str string) string
//   Converts kebab-case to camelCase
// - min(a, b int) int
//   Returns the minimum of two integers
// ============================================================================

var (
	debugTranspile         = false
	shouldTranspileLayouts = true
)

// isDebugTranspile checks if DEBUG_TRANSPILE environment variable is set to "1"
func isDebugTranspile() bool {
	return debugTranspile
}

func init() {
	debugTranspile = (os.Getenv("DEBUG_TRANSPILE") == "1")
	shouldTranspileLayouts = !(os.Getenv("DEBUG_NO_LAYOUTS") == "1")
}

func ReadFileAsString(file string) string {
	// Read the input file
	content, err := os.ReadFile(file)
	if err != nil {
		return fmt.Sprintf("failed to read input file: %v", err)
	}

	return string(content)
}

// convertAlpineAttributesToData converts Alpine.js attributes to data attributes
func convertAlpineAttributesToData(htmlContent string) string {
	// Convert x-data to data-alpine-data
	htmlContent = regexp.MustCompile(`\bx-data\b`).ReplaceAllString(htmlContent, "data-alpine-data")

	// Convert x-show to data-alpine-show
	htmlContent = regexp.MustCompile(`\bx-show\b`).ReplaceAllString(htmlContent, "data-alpine-show")

	// Convert x-if to data-alpine-if
	htmlContent = regexp.MustCompile(`\bx-if\b`).ReplaceAllString(htmlContent, "data-alpine-if")

	// Convert x-for to data-alpine-for
	htmlContent = regexp.MustCompile(`\bx-for\b`).ReplaceAllString(htmlContent, "data-alpine-for")

	// Convert x-model to data-alpine-model
	htmlContent = regexp.MustCompile(`\bx-model\b`).ReplaceAllString(htmlContent, "data-alpine-model")

	// Convert x-on:click to data-alpine-on-click
	htmlContent = regexp.MustCompile(`\bx-on:click\b`).ReplaceAllString(htmlContent, "data-alpine-on-click")

	// Convert @click to data-alpine-on-click
	htmlContent = regexp.MustCompile(`\b@click\b`).ReplaceAllString(htmlContent, "data-alpine-on-click")

	return htmlContent
}

// convertToCamelCase converts kebab-case, snake_case, etc. to camelCase
func convertToCamelCase(str string) string {
	// Handle empty string
	if str == "" {
		return str
	}

	// Split by underscores, hyphens, and other separators
	re := regexp.MustCompile(`[_\-\s]+`)
	parts := re.Split(str, -1)

	if len(parts) == 1 {
		// No separators, capitalize first letter
		return strings.ToUpper(string(str[0])) + str[1:]
	}

	// Capitalize first part and subsequent parts (camelCase)
	result := ""
	for i, part := range parts {
		if part != "" {
			if i == 0 {
				// First part: capitalize first letter
				result += strings.ToUpper(string(part[0])) + part[1:]
			} else {
				// Subsequent parts: capitalize first letter (camelCase)
				result += strings.ToUpper(string(part[0])) + part[1:]
			}
		}
	}

	return result
}

func replaceClassToClassName(htmlContent string) string {
	return strings.ReplaceAll(htmlContent, "class=", "className=")
}

func replaceUnusedInHtml(htmlContent string) string {
	// Remove DOCTYPE declaration - it will be added dynamically
	htmlContent = strings.ReplaceAll(htmlContent, "<!DOCTYPE html>", "")
	htmlContent = strings.ReplaceAll(htmlContent, "<!doctype html>", "")
	htmlContent = strings.ReplaceAll(htmlContent, "<!Doctype html>", "")

	// This line was incorrectly removing the template syntax instead of converting it
	// htmlContent = strings.ReplaceAll(htmlContent, "{{.PageTitle}} - {{.AppName}}", "")

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

	// Convert Go template variables to JSX (more specific pattern)
	htmlContent = regexp.MustCompile(`\{\{\.(\w+)\}\}`).ReplaceAllString(htmlContent, "{page.$1}")
	htmlContent = strings.ReplaceAll(htmlContent, "}}", "}")

	return htmlContent
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// validateHTML validates HTML content using html.Parse
func validateHTML(htmlContent string) error {
	_, err := html.Parse(strings.NewReader(htmlContent))
	return err
}

// removeStyleAndScriptTags removes <style> and <script> tags from HTML content
func removeStyleAndScriptTags(htmlContent string) string {
	// Remove <style> tags
	htmlContent = regexp.MustCompile(`(?s)<style[^>]*>.*?</style>`).ReplaceAllString(htmlContent, "")
	// Remove <script> tags
	htmlContent = regexp.MustCompile(`(?s)<script[^>]*>.*?</script>`).ReplaceAllString(htmlContent, "")
	return htmlContent
}

// sanitizeHtml sanitizes HTML attributes by replacing dots with dashes
func sanitizeHtml(htmlContent string) string {
	// Replace dots in attribute names with dashes
	htmlContent = regexp.MustCompile(`(\w+)\.(\w+)=`).ReplaceAllString(htmlContent, "$1-$2=")
	return htmlContent
}

// removeHTMLComments removes HTML comments from content
func removeHTMLComments(htmlContent string) string {
	// Remove HTML comments
	htmlContent = regexp.MustCompile(`(?s)<!--.*?-->`).ReplaceAllString(htmlContent, "")
	return htmlContent
}

// processIncludes processes <!--#include directives in HTML content
func processIncludes(htmlContent, basePath string) string {
	// Find all include directives
	includePattern := regexp.MustCompile(`<!--\s*#include\s+"([^"]+)"\s*-->`)
	matches := includePattern.FindAllStringSubmatch(htmlContent, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		includePath := match[0] // Full match including comment
		filePath := match[1]    // Just the file path

		// Resolve relative path
		baseDir := filepath.Dir(basePath)
		fullPath := filepath.Join(baseDir, filePath)

		// Read the include file
		includeContent, err := os.ReadFile(fullPath)
		if err != nil {
			if isDebugTranspile() {
				fmt.Printf("DEBUG: Could not read include file %s: %v\n", fullPath, err)
			}
			continue
		}

		// Process nested includes recursively
		processedContent := processIncludes(string(includeContent), fullPath)

		// Replace the include directive with the file content
		htmlContent = strings.ReplaceAll(htmlContent, includePath, processedContent)
	}

	return htmlContent
}

// extractMetaTags extracts meta tags from HTML content and returns them as a string
func extractMetaTags(htmlContent string) string {
	// Find all meta tags in the HTML content
	metaPattern := regexp.MustCompile(`(?i)<meta[^>]*>`)
	matches := metaPattern.FindAllString(htmlContent, -1)

	if len(matches) == 0 {
		return ""
	}

	// Join all meta tags with newlines
	var metaTags strings.Builder
	for _, match := range matches {
		metaTags.WriteString("    ")
		metaTags.WriteString(match)
		metaTags.WriteString("\n")
	}

	return metaTags.String()
}

// MetaTag represents a structured meta tag
type MetaTag struct {
	Name     string
	Content  string
	Property string
}

// extractMetaTagsStructured extracts meta tags from HTML content and returns them as structured data
func extractMetaTagsStructured(htmlContent string) ([]MetaTag, error) {
	// Parse the HTML content
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	var metaTags []MetaTag

	// Function to traverse the HTML tree
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "meta" {
			var metaTag MetaTag

			// Extract attributes
			for _, attr := range n.Attr {
				switch attr.Key {
				case "name":
					metaTag.Name = attr.Val
				case "content":
					metaTag.Content = attr.Val
				case "property":
					metaTag.Property = attr.Val
				}
			}

			// Only add if it has content and either name or property
			if metaTag.Content != "" && (metaTag.Name != "" || metaTag.Property != "") {
				metaTags = append(metaTags, metaTag)
			}
		}

		// Traverse children
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)
	return metaTags, nil
}

// convertMetaTagsToJS converts structured meta tags to JavaScript array format
func convertMetaTagsToJS(metaTags []MetaTag, tagType string) string {
	var result strings.Builder
	result.WriteString("[")

	first := true
	for _, meta := range metaTags {
		// Filter by tag type (name or property)
		if tagType == "name" && meta.Name != "" {
			if !first {
				result.WriteString(", ")
			}
			result.WriteString(fmt.Sprintf(`{name: "%s", content: "%s"}`, meta.Name, meta.Content))
			first = false
		} else if tagType == "property" && meta.Property != "" {
			if !first {
				result.WriteString(", ")
			}
			result.WriteString(fmt.Sprintf(`{property: "%s", content: "%s"}`, meta.Property, meta.Content))
			first = false
		}
	}

	result.WriteString("]")
	return result.String()
}

// removeMetaTags removes meta tags from HTML content
func removeMetaTags(htmlContent string) string {
	// Remove meta tags from the HTML content
	metaPattern := regexp.MustCompile(`(?i)<meta[^>]*>`)
	return metaPattern.ReplaceAllString(htmlContent, "")
}

// integrateLinkPathsAndScriptPathsWithValues integrates actual linkPaths and scriptPaths into the HTML content
func integrateLinkPathsAndScriptPathsWithValues(htmlContent, linkPaths, scriptPaths string) string {
	// Find the </head> tag and insert linkPaths before it
	headClosePattern := regexp.MustCompile(`(?i)</head>`)
	if headClosePattern.MatchString(htmlContent) && linkPaths != "" {
		// Split linkPaths and create actual link tags
		links := strings.Split(linkPaths, ",")
		var linkTags strings.Builder
		for _, link := range links {
			link = strings.TrimSpace(link)
			if link != "" {
				linkTags.WriteString(fmt.Sprintf(`<link rel="stylesheet" href="%s" />`, link))
			}
		}
		htmlContent = headClosePattern.ReplaceAllString(htmlContent, linkTags.String()+"</head>")
	}

	// Find the </body> tag and insert scriptPaths before it
	bodyClosePattern := regexp.MustCompile(`(?i)</body>`)
	if bodyClosePattern.MatchString(htmlContent) && scriptPaths != "" {
		// Split scriptPaths and create actual script tags
		scripts := strings.Split(scriptPaths, ",")
		var scriptTags strings.Builder
		for _, script := range scripts {
			script = strings.TrimSpace(script)
			if script != "" {
				scriptTags.WriteString(fmt.Sprintf(`<script type="text/javascript" src="%s"></script>`, script))
			}
		}
		htmlContent = bodyClosePattern.ReplaceAllString(htmlContent, scriptTags.String()+"</body>")
	}

	return htmlContent
}

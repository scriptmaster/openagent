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

	// Also convert Go template variables in the extracted JS content
	if jsContent != "" {
		originalJSContent := jsContent
		jsContent = strings.ReplaceAll(jsContent, "{{.", "{page.")
		jsContent = strings.ReplaceAll(jsContent, "}}", "}")

		// Convert JSX-style variables in JavaScript to proper JavaScript syntax
		// Replace {page.Variable} with page.Variable (remove curly braces for JS)
		jsContent = strings.ReplaceAll(jsContent, "{page.", "page.")
		// Remove closing braces that are part of JSX variables (but not all closing braces)
		// Use regex to be more specific about which braces to remove
		jsContent = strings.ReplaceAll(jsContent, "page.CurrentHost}", "page.CurrentHost")

		// Debug: check if conversion happened
		if originalJSContent != jsContent {
			fmt.Printf("DEBUG: JS content converted for %s\n", inputPath)
		}

		// Write the updated JS content back to the file
		baseName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
		jsPath := filepath.Join(filepath.Dir(outputPath), "js", baseName+".js")
		if err := os.WriteFile(jsPath, []byte(jsContent), 0644); err != nil {
			return fmt.Errorf("failed to update JS file with converted variables: %v", err)
		}
	}

	// Convert class to className
	htmlContent = strings.ReplaceAll(htmlContent, "class=", "className=")

	// Remove <style> and <script> tags (they've been extracted to separate files)
	htmlContent = removeStyleAndScriptTags(htmlContent)

	// Add links to extracted CSS and JS files
	baseName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))

	// Prepare link and script tags for layout
	var linkTags, scriptTags string

	// Add _common.js first
	commonScript := `<script src="/tsx/js/_common.js"></script>`

	// Prepare CSS link tag
	if cssContent != "" {
		linkTags = fmt.Sprintf(`<link rel="stylesheet" href="/tsx/css/%s.css" />`, baseName)
	}

	// Prepare JS script tags
	if jsContent != "" {
		scriptTags = commonScript + "\n" + fmt.Sprintf(`<script src="/tsx/js/%s.js"></script>`, baseName)
	} else {
		scriptTags = commonScript
	}

	// Convert HTML comments to JSX comments
	htmlContent = convertHtmlCommentsToJsx(htmlContent)

	// Convert Alpine.js attributes to data attributes for JSX compatibility
	htmlContent = convertAlpineAttributesToData(htmlContent)

	// Sanitize HTML attributes by replacing dots with dashes
	htmlContent = sanitizeHtml(htmlContent)

	// Clean up extra whitespace and empty lines
	htmlContent = strings.ReplaceAll(htmlContent, "\n\n\n", "\n")
	htmlContent = strings.ReplaceAll(htmlContent, "\n    \n", "\n")
	htmlContent = strings.ReplaceAll(htmlContent, "\n\n", "\n")
	htmlContent = strings.TrimSpace(htmlContent)

	// Generate the function header based on filename conventions
	fileName := filepath.Base(inputPath)

	componentName := ""
	pageType := "Page" // Default type
	layoutName := ""   // Default layout (will be set if needed)

	if generateComponentName {
		baseName := strings.TrimSuffix(fileName, filepath.Ext(fileName))

		// Check for dot notation pattern (e.g., test.landing_page.html or index.page.landing.html)
		parts := strings.Split(baseName, ".")
		if len(parts) >= 2 {
			// Use the first part as the component name
			componentName = convertToCamelCase(parts[0])

			if len(parts) >= 3 {
				// Three-dot notation: component.pageType.layoutName.html
				// Use the second part as the page type
				pageType = convertToCamelCase(parts[1])
				// Capitalize the first letter for the page type
				if len(pageType) > 0 {
					pageType = strings.ToUpper(string(pageType[0])) + pageType[1:]
				}
				// The third part is the layout name (will be used later)
				layoutName = "layout_" + parts[2]
			} else {
				// Two-dot notation: component.pageType.html
				// Use the second part as the page type
				pageType = convertToCamelCase(parts[1])
				// Capitalize the first letter for the page type
				if len(pageType) > 0 {
					pageType = strings.ToUpper(string(pageType[0])) + pageType[1:]
				}
			}
		} else {
			// No dot notation, use the whole filename
			componentName = convertToCamelCase(baseName)
		}

		// Capitalize the first letter for the component name
		if len(componentName) > 0 {
			componentName = strings.ToUpper(string(componentName[0])) + componentName[1:]
		}

	}

	// Determine if this page needs a layout wrapper
	needsLayout := !strings.Contains(htmlContent, "<html") && !strings.Contains(htmlContent, "<head")

	var tsxContent string
	if needsLayout {
		// Determine which layout to use
		layoutImport := "layout_pages"
		if layoutName != "" {
			layoutImport = layoutName
		}

		// Use layout wrapper for pages without html/head tags
		tsxContent = `import Layout from '../layouts/` + layoutImport + `';

export default function ` + componentName + `({page}: {page: ` + pageType + `}) {
    return (
        <Layout page={page} linkTags={` + "`" + linkTags + "`" + `} scriptTags={` + "`" + scriptTags + "`" + `}>
` + htmlContent + `
        </Layout>
    );
}`
	} else {
		// Use fragment for pages with html/head tags - add CSS/JS links directly to HTML
		// Add CSS link if there's CSS content
		if linkTags != "" {
			// If </head> exists, add before it, otherwise prepend as style tag
			if strings.Contains(htmlContent, "</head>") {
				htmlContent = strings.Replace(htmlContent, "</head>", "\n    "+linkTags+"\n</head>", 1)
			} else {
				htmlContent = linkTags + "\n" + htmlContent
			}
		}

		// Add JS scripts if there's JS content
		if scriptTags != "" {
			// If </body> exists, add before it, otherwise add at end
			if strings.Contains(htmlContent, "</body>") {
				htmlContent = strings.Replace(htmlContent, "</body>", "\n"+scriptTags+"\n</body>", 1)
			} else {
				htmlContent = htmlContent + "\n" + scriptTags
			}
		}

		// Use fragment for pages with html/head tags
		tsxContent = `export default function ` + componentName + `({page}: {page: ` + pageType + `}) {
    return (
<>
` + htmlContent + `
</>
    );
}`
	}

	return os.WriteFile(outputPath, []byte(tsxContent), 0644)
}

// TranspileLayoutToTsx converts a layout HTML file to a TSX layout component
func TranspileLayoutToTsx(inputPath, outputPath string) error {
	// Read the input HTML file
	content, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("could not read input file: %w", err)
	}

	htmlContent := string(content)

	// Check if layout has required html, head, and body tags
	if !strings.Contains(htmlContent, "<html") {
		log.Printf("Warning: Layout file %s does not contain <html> tag", inputPath)
	}
	if !strings.Contains(htmlContent, "<head") {
		log.Printf("Warning: Layout file %s does not contain <head> tag", inputPath)
	}
	if !strings.Contains(htmlContent, "<body") {
		log.Printf("Warning: Layout file %s does not contain <body> tag", inputPath)
	}

	// Remove DOCTYPE declaration - it will be added dynamically
	htmlContent = strings.ReplaceAll(htmlContent, "<!DOCTYPE html>", "")
	htmlContent = strings.ReplaceAll(htmlContent, "<!doctype html>", "")
	htmlContent = strings.ReplaceAll(htmlContent, "<!Doctype html>", "")

	// Convert Go template variables to JSX
	htmlContent = strings.ReplaceAll(htmlContent, "{{.", "{page.")
	htmlContent = strings.ReplaceAll(htmlContent, "}}", "}")

	// Convert class to className
	htmlContent = strings.ReplaceAll(htmlContent, "class=", "className=")

	// Replace {children} placeholder with actual children prop
	htmlContent = strings.ReplaceAll(htmlContent, "{children}", "{children}")

	// Convert HTML comments to JSX comments
	htmlContent = convertHtmlCommentsToJsx(htmlContent)

	// Convert Alpine.js attributes to data attributes for JSX compatibility
	htmlContent = convertAlpineAttributesToData(htmlContent)

	// Sanitize HTML attributes by replacing dots with dashes
	htmlContent = sanitizeHtml(htmlContent)

	// Clean up extra whitespace and empty lines
	htmlContent = strings.ReplaceAll(htmlContent, "\n\n\n", "\n")
	htmlContent = strings.ReplaceAll(htmlContent, "\n    \n", "\n")
	htmlContent = strings.ReplaceAll(htmlContent, "\n\n", "\n")
	htmlContent = strings.TrimSpace(htmlContent)

	// Generate the layout component name
	fileName := filepath.Base(inputPath)
	componentName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	componentName = convertToCamelCase(componentName)

	// Capitalize the first letter for the component name
	if len(componentName) > 0 {
		componentName = strings.ToUpper(string(componentName[0])) + componentName[1:]
	}

	// Write the layout TSX file
	tsxContent := `export default function ` + componentName + `({page, children, linkTags, scriptTags}: {page: any, children: any, linkTags?: string, scriptTags?: string}) {
    return (
<>
` + htmlContent + `
</>
    );
}`

	return os.WriteFile(outputPath, []byte(tsxContent), 0644)
}

// transpileAllTemplates finds and transpiles all HTML files from different directories.
func transpileAllTemplates() error {
	log.Println("Transpiling all HTML templates...")

	// Define input directories and their corresponding output directories
	directories := map[string]string{
		"./tpl/pages":   "./tpl/generated/pages",
		"./tpl/admin":   "./tpl/generated/admin",
		"./tpl/app":     "./tpl/generated/app",
		"./tpl/layouts": "./tpl/generated/layouts",
	}

	for inputDir, outputDir := range directories {
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
			baseNameWithoutExt := strings.TrimSuffix(baseName, ".html")

			// Extract component name from filename (first part before dot, or whole name if no dots)
			componentName := baseNameWithoutExt
			if strings.Contains(baseNameWithoutExt, ".") {
				parts := strings.Split(baseNameWithoutExt, ".")
				componentName = parts[0] // Use first part as component name
			}

			tsxFileName := componentName + ".tsx"
			outputPath := filepath.Join(outputDir, tsxFileName)

			// Check if this is a layout file
			if strings.Contains(inputDir, "layouts") {
				if err := TranspileLayoutToTsx(file, outputPath); err != nil {
					return fmt.Errorf("failed to transpile layout %s: %w", file, err)
				}
			} else {
				if err := TranspileHtmlToTsx(file, outputPath); err != nil {
					return fmt.Errorf("failed to transpile %s: %w", file, err)
				}
			}
			log.Printf("Transpiled %s to %s", file, outputPath)
		}
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
		cssDir := filepath.Join(filepath.Dir(outputPath), "css")
		if err := os.MkdirAll(cssDir, 0755); err != nil {
			return "", "", fmt.Errorf("failed to create CSS directory: %v", err)
		}
		cssPath := filepath.Join(cssDir, baseName+".css")
		if err := os.WriteFile(cssPath, []byte(cssContent.String()), 0644); err != nil {
			return "", "", fmt.Errorf("failed to write CSS file: %v", err)
		}
	}

	// Write JS file if there's content
	if jsContent.Len() > 0 {
		jsDir := filepath.Join(filepath.Dir(outputPath), "js")
		if err := os.MkdirAll(jsDir, 0755); err != nil {
			return "", "", fmt.Errorf("failed to create JS directory: %v", err)
		}
		jsPath := filepath.Join(jsDir, baseName+".js")
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

// convertAlpineAttributesToData converts Alpine.js attributes to data attributes for JSX compatibility
func convertAlpineAttributesToData(htmlContent string) string {
	// Convert @click to data-click
	htmlContent = strings.ReplaceAll(htmlContent, `@click="`, `data-click="`)
	htmlContent = strings.ReplaceAll(htmlContent, `@submit="`, `data-submit="`)
	htmlContent = strings.ReplaceAll(htmlContent, `@submit.prevent="`, `data-submit.prevent="`)
	htmlContent = strings.ReplaceAll(htmlContent, `@change="`, `data-change="`)
	htmlContent = strings.ReplaceAll(htmlContent, `@input="`, `data-input="`)
	htmlContent = strings.ReplaceAll(htmlContent, `@focus="`, `data-focus="`)
	htmlContent = strings.ReplaceAll(htmlContent, `@blur="`, `data-blur="`)
	htmlContent = strings.ReplaceAll(htmlContent, `@keydown="`, `data-keydown="`)
	htmlContent = strings.ReplaceAll(htmlContent, `@keyup="`, `data-keyup="`)

	// Convert :attribute to data-attribute (for dynamic attributes)
	htmlContent = strings.ReplaceAll(htmlContent, `:className="`, `data-className="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:class="`, `data-class="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:disabled="`, `data-disabled="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:required="`, `data-required="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:checked="`, `data-checked="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:selected="`, `data-selected="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:value="`, `data-value="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:src="`, `data-src="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:href="`, `data-href="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:alt="`, `data-alt="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:title="`, `data-title="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:placeholder="`, `data-placeholder="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:style="`, `data-style="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:id="`, `data-id="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:name="`, `data-name="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:type="`, `data-type="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:role="`, `data-role="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:aria-label="`, `data-aria-label="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:aria-describedby="`, `data-aria-describedby="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:aria-expanded="`, `data-aria-expanded="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:aria-hidden="`, `data-aria-hidden="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:tabindex="`, `data-tabindex="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:readonly="`, `data-readonly="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:autocomplete="`, `data-autocomplete="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:autofocus="`, `data-autofocus="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:form="`, `data-form="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:formaction="`, `data-formaction="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:formenctype="`, `data-formenctype="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:formmethod="`, `data-formmethod="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:formnovalidate="`, `data-formnovalidate="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:formtarget="`, `data-formtarget="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:min="`, `data-min="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:max="`, `data-max="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:step="`, `data-step="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:pattern="`, `data-pattern="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:size="`, `data-size="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:maxlength="`, `data-maxlength="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:minlength="`, `data-minlength="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:spellcheck="`, `data-spellcheck="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:wrap="`, `data-wrap="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:rows="`, `data-rows="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:cols="`, `data-cols="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:multiple="`, `data-multiple="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:accept="`, `data-accept="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:capture="`, `data-capture="`)
	htmlContent = strings.ReplaceAll(htmlContent, `:key="`, `data-key="`)

	// Convert x-* attributes to data-x-* (Alpine.js directives)
	htmlContent = strings.ReplaceAll(htmlContent, `x-data="`, `data-x-data="`)
	htmlContent = strings.ReplaceAll(htmlContent, `x-show="`, `data-x-show="`)
	htmlContent = strings.ReplaceAll(htmlContent, `x-hide="`, `data-x-hide="`)
	htmlContent = strings.ReplaceAll(htmlContent, `x-if="`, `data-x-if="`)
	htmlContent = strings.ReplaceAll(htmlContent, `x-for="`, `data-x-for="`)
	htmlContent = strings.ReplaceAll(htmlContent, `x-model="`, `data-x-model="`)
	htmlContent = strings.ReplaceAll(htmlContent, `x-text="`, `data-x-text="`)
	htmlContent = strings.ReplaceAll(htmlContent, `x-html="`, `data-x-html="`)
	htmlContent = strings.ReplaceAll(htmlContent, `x-bind="`, `data-x-bind="`)
	htmlContent = strings.ReplaceAll(htmlContent, `x-on="`, `data-x-on="`)
	htmlContent = strings.ReplaceAll(htmlContent, `x-transition="`, `data-x-transition="`)
	htmlContent = strings.ReplaceAll(htmlContent, `x-cloak="`, `data-x-cloak="`)
	htmlContent = strings.ReplaceAll(htmlContent, `x-teleport="`, `data-x-teleport="`)
	htmlContent = strings.ReplaceAll(htmlContent, `x-effect="`, `data-x-effect="`)
	htmlContent = strings.ReplaceAll(htmlContent, `x-ignore="`, `data-x-ignore="`)
	htmlContent = strings.ReplaceAll(htmlContent, `x-ref="`, `data-x-ref="`)
	htmlContent = strings.ReplaceAll(htmlContent, `x-id="`, `data-x-id="`)

	return htmlContent
}

// sanitizeHtml replaces dots with dashes in HTML attributes for JSX compatibility
func sanitizeHtml(htmlContent string) string {
	// Use regex to find attributes with dots and replace them with dashes
	// Pattern matches: attribute="value" or attribute='value' where attribute contains dots
	attributePattern := regexp.MustCompile(`(\w+(?:\.\w+)+)=`)
	return attributePattern.ReplaceAllStringFunc(htmlContent, func(match string) string {
		// Replace dots with dashes in attribute names
		return strings.ReplaceAll(match, ".", "-")
	})
}

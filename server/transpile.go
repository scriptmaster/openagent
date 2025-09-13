package server

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// isDebugTranspile checks if DEBUG_TRANSPILE environment variable is set to "1"
func isDebugTranspile() bool {
	return debugTranspile
}

// Global regex patterns compiled once for performance
var (
	stylePattern            *regexp.Regexp
	scriptPattern           *regexp.Regexp
	removeStylePattern      *regexp.Regexp
	removeScriptPattern     *regexp.Regexp
	commentPattern          *regexp.Regexp
	hyphenUnderscorePattern *regexp.Regexp
	includePattern          *regexp.Regexp
	htmlIncludePattern      *regexp.Regexp

	// Self-closing tags map for efficient lookup
	selfClosingTags map[string]*regexp.Regexp

	generateComponentName = true
	debugTranspile        = false
)

// init initializes all regex patterns
func init() {
	debugTranspile = (os.Getenv("DEBUG_TRANSPILE") == "1")

	stylePattern = regexp.MustCompile(`(?s)<style[^>]*>(.*?)</style>`)
	scriptPattern = regexp.MustCompile(`(?s)<script[^>]*>(.*?)</script>`)
	removeStylePattern = regexp.MustCompile(`(?s)<style[^>]*>.*?</style>`)
	removeScriptPattern = regexp.MustCompile(`(?s)<script[^>]*>.*?</script>`)
	commentPattern = regexp.MustCompile(`(?s)<!--(.*?)-->`)
	hyphenUnderscorePattern = regexp.MustCompile(`[-_]`)
	includePattern = regexp.MustCompile(`(?m)^\s*//#include\s+(?:"([^"]+)"|([^\s]+))\s*$`)
	htmlIncludePattern = regexp.MustCompile(`(?s)<!--\s*#include\s+"([^"]+)"\s*-->`)

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

	// Extract CSS and JS content to separate files BEFORE any Go template processing
	cssContent, jsContent, err := extractCSSAndJS(string(content), inputPath, outputPath)
	if err != nil {
		return fmt.Errorf("failed to extract CSS/JS: %v", err)
	}

	// Convert Go template syntax to JSX
	htmlContent := string(content)

	// Clear global imported components for this file
	// Simple approach: no global variables needed

	// Process component imports (e.g., <div id="component-counter"></div>)
	var importedComponents []string
	htmlContent, importedComponents, err = processComponentImports(htmlContent, inputPath)
	if err != nil {
		return fmt.Errorf("failed to process component imports: %v", err)
	}

	// Remove DOCTYPE declaration - it will be added dynamically
	htmlContent = strings.ReplaceAll(htmlContent, "<!DOCTYPE html>", "")
	htmlContent = strings.ReplaceAll(htmlContent, "<!doctype html>", "")
	htmlContent = strings.ReplaceAll(htmlContent, "<!Doctype html>", "")

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

	// Convert Go template variables to JSX (more specific pattern)
	htmlContent = regexp.MustCompile(`\{\{\.(\w+)\}\}`).ReplaceAllString(htmlContent, "{page.$1}")
	htmlContent = strings.ReplaceAll(htmlContent, "}}", "}")

	// CSS and JS extraction/writing is handled by extractCSSAndJS function

	// Convert class to className
	htmlContent = strings.ReplaceAll(htmlContent, "class=", "className=")

	// Remove <style> and <script> tags (they've been extracted to separate files)
	htmlContent = removeStyleAndScriptTags(htmlContent)

	// Add links to extracted CSS and JS files
	baseName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))

	// Prepare link and script paths for layout
	var linkPaths, scriptPaths string

	// Prepare CSS link path
	if cssContent != "" {
		linkPaths = fmt.Sprintf("/tsx/css/%s.css", baseName)
	}

	var jsFileName string
	// Prepare JS script paths
	if jsContent != "" {
		// Determine the correct JS filename based on the output directory
		outputDir := filepath.Dir(outputPath)
		lastDirComponent := filepath.Base(outputDir)
		jsFileName = lastDirComponent + "_" + baseName + ".js"
		scriptPaths = fmt.Sprintf("/tsx/js/_common.js,/tsx/js/%s", jsFileName)
	} else {
		scriptPaths = "/tsx/js/_common.js"
	}

	// Component JS files will be embedded directly into the main JS file
	// No need to add them to script paths

	// Convert HTML comments to JSX comments
	// Remove all HTML comments to prevent React hydration errors
	htmlContent = removeHTMLComments(htmlContent)
	// Convert HTML comments to JSX comments (DISABLED - causes React hydration errors)
	// htmlContent = convertHtmlCommentsToJsx(htmlContent)

	// Convert Alpine.js attributes to data attributes for JSX compatibility
	htmlContent = convertAlpineAttributesToData(htmlContent)

	// Sanitize HTML attributes by replacing dots with dashes A/3
	htmlContent = sanitizeHtml(htmlContent)

	// // Fix self-closing tags A/3
	// htmlContent = fixSelfClosingTags(htmlContent)

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
	baseName = strings.TrimSuffix(fileName, filepath.Ext(fileName))

	if generateComponentName {
		// Check for dot notation pattern (e.g., test.landing_page.html or index.page.landing.html)
		parts := strings.Split(baseName, ".")
		if len(parts) >= 2 {
			if len(parts) >= 3 {
				// Three-dot notation: component.pageType.layoutName.html
				// Use the full filename as component name to avoid conflicts
				componentName = convertToCamelCase(baseName)
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
				// Use the first part as the component name
				componentName = convertToCamelCase(parts[0])
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

	// Use fragment for pages with html/head tags - add CSS/JS links directly to HTML
	// Add CSS link if there's CSS content
	if linkPaths != "" {
		cssLink := fmt.Sprintf(`<link rel="stylesheet" href="%s" />`, linkPaths)
		// If </head> exists, add before it, otherwise prepend as style tag
		if strings.Contains(htmlContent, "</head>") {
			htmlContent = strings.Replace(htmlContent, "</head>", "\n    "+cssLink+"\n</head>", 1)
		} else if !needsLayout {
			htmlContent = cssLink + "\n" + htmlContent
		}
	}

	// Add JS scripts if there's JS content
	if scriptPaths != "" {
		// Split scriptPaths and create script tags
		paths := strings.Split(scriptPaths, ",")
		var scriptTags string
		for _, path := range paths {
			scriptTags += fmt.Sprintf(`<script src="%s"></script>`, strings.TrimSpace(path)) + "\n"
		}
		// If </body> exists, add before it, otherwise add at end
		if strings.Contains(htmlContent, "</body>") {
			htmlContent = strings.Replace(htmlContent, "</body>", "\n"+scriptTags+"</body>", 1)
		} else if !needsLayout {
			htmlContent = htmlContent + "\n" + scriptTags
		}
	}

	// Generate imports for components
	var imports string
	if len(importedComponents) > 0 {
		if isDebugTranspile() {
			fmt.Printf("DEBUG: Found %d imported components: %v\n", len(importedComponents), importedComponents)
		}

		var importsBuilder strings.Builder
		importsBuilder.WriteString("// Component imports\n")
		for _, componentName := range importedComponents {
			importsBuilder.WriteString(fmt.Sprintf("import %s from '../components/%s';\n", componentName, strings.ToLower(componentName)))
		}
		importsBuilder.WriteString("\n")
		imports = importsBuilder.String()

		if isDebugTranspile() {
			fmt.Printf("DEBUG: Generated imports: %s\n", imports)
		}
	}

	if isDebugTranspile() {
		fmt.Printf("DEBUG: htmlContent after component processing: %s\n", htmlContent[:min(200, len(htmlContent))])
	}

	// Create the main component file (test.tsx) - contains the actual component
	tsxContent = imports + `export default function ` + componentName + `({page}) {
    return (
<main>
` + htmlContent + `
</main>
    );
}`

	// Write the component file
	componentPath := strings.Replace(outputPath, ".tsx", ".component.tsx", 1)
	if err := os.WriteFile(componentPath, []byte(tsxContent), 0644); err != nil {
		return fmt.Errorf("failed to write component file: %v", err)
	}

	if needsLayout {
		// Determine which layout to use
		layoutImport := "layout_pages"
		if layoutName != "" {
			layoutImport = layoutName
		}

		// Create the layout wrapper file (test.layout.tsx) - imports Layout and the component
		layoutContent := `import Layout from '../layouts/` + layoutImport + `';
import App from './` + baseName + `.component';

export default function ` + componentName + `Layout({page}: {page: ` + pageType + `}) {
    return (
        <Layout page={page} linkPaths={` + "`" + linkPaths + "`" + `} scriptPaths={` + "`" + scriptPaths + "`" + `}>
            <App page={page} />
        </Layout>
    );
}`
		// Write the layout wrapper file
		layoutPath := strings.Replace(outputPath, ".tsx", ".tsx", 1)
		if err := os.WriteFile(layoutPath, []byte(layoutContent), 0644); err != nil {
			return fmt.Errorf("failed to write layout file: %v", err)
		}
	} else {
		pagePath := strings.Replace(outputPath, ".tsx", ".tsx", 1)
		if err := os.WriteFile(pagePath, []byte(tsxContent), 0644); err != nil {
			return fmt.Errorf("failed to write component file: %v", err)
		}
	}

	return nil

	// return os.WriteFile(outputPath, []byte(tsxContent), 0644)
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

	// Process #include directives for partials
	htmlContent, err = processIncludes(htmlContent, inputPath)
	if err != nil {
		return fmt.Errorf("failed to process includes: %w", err)
	}

	// Remove DOCTYPE declaration - it will be added dynamically
	htmlContent = strings.ReplaceAll(htmlContent, "<!DOCTYPE html>", "")
	htmlContent = strings.ReplaceAll(htmlContent, "<!doctype html>", "")
	htmlContent = strings.ReplaceAll(htmlContent, "<!Doctype html>", "")

	// Convert Go template variables to JSX (more specific pattern)
	htmlContent = regexp.MustCompile(`\{\{\.(\w+)\}\}`).ReplaceAllString(htmlContent, "{page.$1}")
	htmlContent = strings.ReplaceAll(htmlContent, "}}", "}")

	// Convert class to className
	htmlContent = strings.ReplaceAll(htmlContent, "class=", "className=")

	// Replace {children} placeholder with actual children prop
	htmlContent = strings.ReplaceAll(htmlContent, "{children}", "{children}")

	// Dynamically inject linkTags and scriptTags based on linkPaths and scriptPaths
	htmlContent = injectDynamicTags(htmlContent)

	// Convert HTML comments to JSX comments
	// Remove all HTML comments to prevent React hydration errors
	htmlContent = removeHTMLComments(htmlContent)
	// Convert HTML comments to JSX comments (DISABLED - causes React hydration errors)
	// htmlContent = convertHtmlCommentsToJsx(htmlContent)

	// Convert Alpine.js attributes to data attributes for JSX compatibility
	htmlContent = convertAlpineAttributesToData(htmlContent)

	// Sanitize HTML attributes by replacing dots with dashes
	htmlContent = sanitizeHtml(htmlContent) // B/3

	// B/3
	htmlContent = fixSelfClosingTags(htmlContent)

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
	tsxContent := `export default function ` + componentName + `({page, children, linkPaths, scriptPaths}: {page: any, children?: any, linkPaths?: string, scriptPaths?: string}) {
    return (
<>
` + htmlContent + `
</>
    );
}`

	return os.WriteFile(outputPath, []byte(tsxContent), 0644)
}

// injectDynamicTags injects dynamic link and script tags based on paths
func injectDynamicTags(htmlContent string) string {
	// Inject linkTags before </head>
	linkTagsCode := `{linkPaths && linkPaths.split(',').map((path, index) => (
    <link key={'gen-link-'+index} rel="stylesheet" href={path.trim()} />
))}`
	htmlContent = strings.Replace(htmlContent, "</head>", "\n    "+linkTagsCode+"\n</head>", 1)

	// Inject scriptTags before </body>
	scriptTagsCode := `{scriptPaths && scriptPaths.split(',').map((path, index) => (
    <script key={'gen-script-'+index} src={path.trim()}></script>
))}`
	htmlContent = strings.Replace(htmlContent, "</body>", "\n    "+scriptTagsCode+"\n</body>", 1)

	return htmlContent
}

// TranspileAllTemplates finds and transpiles all HTML files from different directories.
func TranspileAllTemplates() error {
	log.Println("\t → 5. Transpiling all HTML templates...")

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
		// if err := os.MkdirAll(filepath.Join(inputDir, "js"), 0755); err != nil {
		// 	return err
		// }

		files, err := filepath.Glob(filepath.Join(inputDir, "*.html"))
		if err != nil {
			return err
		}

		if isDebugTranspile() {
			fmt.Printf("DEBUG: Processing %d files in %s\n", len(files), inputDir)
		}
		for _, file := range files {
			if isDebugTranspile() {
				fmt.Printf("DEBUG: Processing file: %s\n", file)
			}
			baseName := filepath.Base(file)
			baseNameWithoutExt := strings.TrimSuffix(baseName, ".html")

			// Extract component name from filename (first part before dot, or whole name if no dots)
			componentName := baseNameWithoutExt
			if strings.Contains(baseNameWithoutExt, ".") {
				parts := strings.Split(baseNameWithoutExt, ".")
				componentName = parts[0] // Use first part as component name
			}

			// For three-dot notation files, use the full filename to avoid conflicts
			// e.g., index.page.landing.html -> index.page.landing.tsx
			if strings.Count(baseNameWithoutExt, ".") >= 2 {
				componentName = baseNameWithoutExt
			}

			tsxFileName := componentName + ".tsx"
			outputPath := filepath.Join(outputDir, tsxFileName)

			// Check for view override first
			viewPath := filepath.Join("./tpl/views", strings.TrimPrefix(inputDir, "./tpl/"), tsxFileName)
			if _, err := os.Stat(viewPath); err == nil {
				// View override exists, copy it instead of transpiling
				if err := copyFile(viewPath, outputPath); err != nil {
					return fmt.Errorf("failed to copy view override %s: %w", viewPath, err)
				}
				log.Printf("Copied view override %s to %s", viewPath, outputPath)
				continue
			}

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
			// log.Printf("Transpiled %s to %s", file, outputPath)
		}
	}

	// Copy _common.js from views/js to generated/js
	commonSourcePath := "./tpl/views/js/_common.js"
	commonDestPath := "./tpl/generated/js/_common.js"

	// Create the generated/js directory if it doesn't exist
	if err := os.MkdirAll("./tpl/generated/js", 0755); err != nil {
		return fmt.Errorf("failed to create generated/js directory: %w", err)
	}

	// Check if _common.js exists in views/js and copy it with include processing
	if _, err := os.Stat(commonSourcePath); err == nil {
		// Read the source file
		content, err := os.ReadFile(commonSourcePath)
		if err != nil {
			return fmt.Errorf("failed to read _common.js: %w", err)
		}

		// Process includes in the JavaScript content
		processedContent, err := processIncludes(string(content), commonSourcePath)
		if err != nil {
			return fmt.Errorf("failed to process includes in _common.js: %w", err)
		}

		// Write the processed content to the destination
		if err := os.WriteFile(commonDestPath, []byte(processedContent), 0644); err != nil {
			return fmt.Errorf("failed to write processed _common.js: %w", err)
		}
		if isDebugTranspile() {
			log.Printf("Copied _common.js from %s to %s", commonSourcePath, commonDestPath)
		}
	} else {
		log.Printf("Warning: _common.js not found at %s", commonSourcePath)
	}

	// Copy _partials directory to generated directory for layout processing
	partialsSourcePath := "./tpl/_partials"
	partialsDestPath := "./tpl/generated/_partials"

	// Check if _partials directory exists and copy it
	if _, err := os.Stat(partialsSourcePath); err == nil {
		if err := copyDirectory(partialsSourcePath, partialsDestPath); err != nil {
			return fmt.Errorf("failed to copy _partials directory: %w", err)
		}
		if isDebugTranspile() {
			log.Printf("Copied _partials directory from %s to %s", partialsSourcePath, partialsDestPath)
		}

		// Retranspile layout files after partials are processed
		// This ensures layout files include the updated partial content
		layoutDir := "./tpl/layouts"
		layoutOutputDir := "./tpl/generated/layouts"

		if files, err := filepath.Glob(filepath.Join(layoutDir, "*.html")); err == nil {
			for _, file := range files {
				baseName := filepath.Base(file)
				baseNameWithoutExt := strings.TrimSuffix(baseName, ".html")
				tsxFileName := baseNameWithoutExt + ".tsx"
				outputPath := filepath.Join(layoutOutputDir, tsxFileName)

				// Retranspile the layout file to pick up updated partial content
				if err := TranspileLayoutToTsx(file, outputPath); err != nil {
					return fmt.Errorf("failed to retranspile layout %s: %w", file, err)
				}
				if isDebugTranspile() {
					log.Printf("Retranspiled layout %s to pick up updated partials", file)
				}
			}
		}
	} else {
		log.Printf("Warning: _partials directory not found at %s", partialsSourcePath)
	}
	log.Println("\t → \t → 5.1 Success: Transpiling all HTML templates")
	return nil
}

// copyFile copies a file from src to dst, fixing import paths for generated directory structure
// and processing #include directives
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Read the entire file content
	content, err := io.ReadAll(sourceFile)
	if err != nil {
		return err
	}

	// Process the content
	contentStr := string(content)

	// Process #include directives
	contentStr, err = processIncludes(contentStr, src)
	if err != nil {
		return fmt.Errorf("failed to process includes: %w", err)
	}

	// Apply the same transformations as HTML files for consistency
	// Convert Go template variables to JSX (more specific pattern)
	contentStr = regexp.MustCompile(`\{\{\.(\w+)\}\}`).ReplaceAllString(contentStr, "{page.$1}")
	contentStr = strings.ReplaceAll(contentStr, "}}", "}")

	// Convert class to className
	contentStr = strings.ReplaceAll(contentStr, "class=", "className=")

	// Remove all HTML comments to prevent React hydration errors
	contentStr = removeHTMLComments(contentStr)
	// Convert HTML comments to JSX comments (DISABLED - causes React hydration errors)
	// contentStr = convertHtmlCommentsToJsx(contentStr)

	// Convert Alpine.js attributes to data attributes for JSX compatibility
	contentStr = convertAlpineAttributesToData(contentStr)

	// Sanitize HTML attributes by replacing dots with dashes
	contentStr = sanitizeHtml(contentStr) // C/3

	// Fix self-closing tags
	contentStr = fixSelfClosingTags(contentStr) // C/3

	// Fix import paths for generated directory structure
	// Replace relative imports that go up too many levels
	contentStr = strings.ReplaceAll(contentStr, "../../layouts/", "../layouts/")
	contentStr = strings.ReplaceAll(contentStr, "../../generated/", "../")

	// Clean up extra whitespace and empty lines
	contentStr = strings.ReplaceAll(contentStr, "\n\n\n", "\n")
	contentStr = strings.ReplaceAll(contentStr, "\n    \n", "\n")
	contentStr = strings.ReplaceAll(contentStr, "\n\n", "\n")
	contentStr = strings.TrimSpace(contentStr)

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = destFile.WriteString(contentStr)
	return err
}

// processIncludes processes //#include "relative_path" and <!--#include "relative_path" --> directives in the content
func processIncludes(content, sourceFile string) (string, error) {
	// Find all JavaScript include directives using the global compiled pattern
	jsMatches := includePattern.FindAllStringSubmatch(content, -1)

	for _, match := range jsMatches {
		if len(match) != 3 {
			continue
		}

		// Handle both quoted and unquoted paths
		var relativePath string
		if match[1] != "" {
			relativePath = match[1] // Quoted path
		} else {
			relativePath = match[2] // Unquoted path
		}
		fullIncludeLine := match[0]

		// Handle absolute paths (starting with /) vs relative paths
		var includePath string
		if strings.HasPrefix(relativePath, "/") {
			// Absolute path - treat as relative to project root
			includePath = filepath.Join(".", relativePath)
		} else {
			// Relative path - resolve from the source file's directory
			sourceDir := filepath.Dir(sourceFile)
			includePath = filepath.Join(sourceDir, relativePath)
		}

		// Normalize the path
		includePath = filepath.Clean(includePath)

		// Read the included file
		includedContent, err := os.ReadFile(includePath)
		if err == nil {
			// Replace the include directive with the file content
			content = strings.Replace(content, fullIncludeLine, string(includedContent), 1)
		}
	}

	// Find all HTML comment include directives using the global compiled pattern
	htmlMatches := htmlIncludePattern.FindAllStringSubmatch(content, -1)

	for _, match := range htmlMatches {
		if len(match) != 2 {
			continue
		}

		relativePath := match[1]
		fullIncludeLine := match[0]

		// Handle absolute paths (starting with /) vs relative paths
		var includePath string
		if strings.HasPrefix(relativePath, "/") {
			// Absolute path - treat as relative to project root
			includePath = filepath.Join(".", relativePath)
		} else {
			// Relative path - resolve from the source file's directory
			sourceDir := filepath.Dir(sourceFile)
			includePath = filepath.Join(sourceDir, relativePath)
		}

		// Normalize the path
		includePath = filepath.Clean(includePath)

		// Read the included file
		includedContent, err := os.ReadFile(includePath)
		if err == nil {
			// Replace the include directive with the file content
			content = strings.Replace(content, fullIncludeLine, string(includedContent), 1)
		}
	}

	return content, nil
}

// copyDirectory recursively copies a directory and all its contents
func copyDirectory(src, dst string) error {
	// Create the destination directory
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	// Read the source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectories
			if err := copyDirectory(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy files
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
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
				// Ensure attributes are properly formatted
				if attrs != "" && !strings.HasPrefix(attrs, " ") {
					attrs = " " + attrs
				}
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
	if isDebugTranspile() {
		fmt.Printf("DEBUG: Found %d script matches in %s\n", len(jsMatches), inputPath)
	}
	for i, match := range jsMatches {
		if len(match) > 1 {
			if isDebugTranspile() {
				fmt.Printf("DEBUG: Script match %d length: %d\n", i, len(match[1]))
			}
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
	if isDebugTranspile() {
		fmt.Printf("DEBUG: jsContent.Len() = %d\n", jsContent.Len())
	}
	if jsContent.Len() > 0 && outputPath != "" {
		// Always write to generated/js directory
		jsDir := "tpl/generated/js"
		if err := os.MkdirAll(jsDir, 0755); err != nil {
			return "", "", fmt.Errorf("failed to create JS directory: %v", err)
		}

		// Determine filename - use pages_ prefix for page files
		var jsFileName string
		outputDir := filepath.Dir(outputPath)
		if isDebugTranspile() {
			fmt.Printf("DEBUG: outputPath=%s, outputDir=%s, baseName=%s\n", outputPath, outputDir, baseName)
		}
		if strings.Contains(outputDir, "pages") {
			jsFileName = "pages_" + baseName + ".js"
			if isDebugTranspile() {
				fmt.Printf("DEBUG: Detected pages directory, using pages_ prefix\n")
			}
		} else {
			jsFileName = baseName + ".js"
			if isDebugTranspile() {
				fmt.Printf("DEBUG: Not a pages directory, using regular filename\n")
			}
		}
		if isDebugTranspile() {
			fmt.Printf("DEBUG: jsFileName=%s\n", jsFileName)
		}
		jsPath := filepath.Join(jsDir, jsFileName)

		// Create React-enhanced JS content
		reactJSContent := createReactJSContent(jsContent.String(), baseName)

		if isDebugTranspile() {
			fmt.Printf("DEBUG: Writing JS file %s with %d characters\n", jsPath, len(reactJSContent))
			if len(reactJSContent) > 200 {
				fmt.Printf("DEBUG: First 200 chars of JS content: %s\n", reactJSContent[:200])
			} else {
				fmt.Printf("DEBUG: Full JS content: %s\n", reactJSContent)
			}
		}
		if err := os.WriteFile(jsPath, []byte(reactJSContent), 0644); err != nil {
			return "", "", fmt.Errorf("failed to write JS file[1]: %v", err)
		}
		if isDebugTranspile() {
			fmt.Printf("DEBUG: Successfully wrote JS file %s\n", jsPath)
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

// // convertHtmlCommentsToJsx converts HTML comments to JSX comments
// func convertHtmlCommentsToJsx(htmlContent string) string {
// 	// Replace HTML comments with JSX comments
// 	htmlContent = commentPattern.ReplaceAllStringFunc(htmlContent, func(match string) string {
// 		// Extract the comment content (without <!-- and -->)
// 		content := commentPattern.FindStringSubmatch(match)[1]
// 		// Convert to JSX comment format: {/* ... */}
// 		return fmt.Sprintf("{/*%s*/}", content)
// 	})
// 	return htmlContent
// }

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

	actualComponentName := GetActualComponentName(componentJS, componentName)

	if isDebugTranspile() {
		fmt.Printf("DEBUG: actualComponentName = '%s'\n", actualComponentName)
	}

	// Create the main component JS content first
	mainComponentJS := fmt.Sprintf(`
// ╔══════════════════════════════════════════════════════════════════════════════
// ║                    ⚛️  MAIN COMPONENT JS (TSX → JS) ⚛️                      
// ╚══════════════════════════════════════════════════════════════════════════════
%s

// ╔══════════════════════════════════════════════════════════════════════════════
// ║                        📜 ORIGINAL JS CONTENT 📜                            
// ╚══════════════════════════════════════════════════════════════════════════════
%s`, componentJS, originalJS)

	// Embed component JS content into the main JS file
	embeddedJS := embedComponentJS(mainComponentJS)

	// Create the React-enhanced JS content with embedded components
	reactJS := fmt.Sprintf(`
%s

// ╔══════════════════════════════════════════════════════════════════════════════
// ║                        💧 HYDRATION 💧                            
// ╚══════════════════════════════════════════════════════════════════════════════

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
}`, embeddedJS, actualComponentName, actualComponentName, actualComponentName)

	if isDebugTranspile() {
		fmt.Printf("DEBUG: actualComponentName = '%s'\n", actualComponentName)
		fmt.Printf("DEBUG: Final reactJS content (first 500 chars): %s\n", reactJS[:min(500, len(reactJS))])
	}

	return reactJS
}

func TSX2JS(tsxStr string) string {
	// Remove TypeScript types
	log.Println("TSX2JS: ", tsxStr)

	tsxStr = removeTypeScriptTypes(tsxStr)
	// log.Println("removeTypeScriptTypes: ", tsxStr)

	// Extract imports and main content
	var imports string
	var mainContent string

	if strings.HasPrefix(tsxStr, "// Component imports") {
		// Find the end of imports (look for the first "export default function")
		lines := strings.Split(tsxStr, "\n")
		var importLines []string
		var mainLines []string
		inImports := true

		for _, line := range lines {
			if inImports && strings.HasPrefix(strings.TrimSpace(line), "export default function") {
				inImports = false
				mainLines = append(mainLines, line)
			} else if inImports {
				importLines = append(importLines, line)
			} else {
				mainLines = append(mainLines, line)
			}
		}

		if len(importLines) > 0 {
			imports = strings.Join(importLines, "\n") + "\n\n"
			mainContent = strings.Join(mainLines, "\n")
			log.Printf("DEBUG: Extracted imports: %s", imports)
			log.Printf("DEBUG: Extracted mainContent: %s", mainContent[:min(200, len(mainContent))])
		} else {
			mainContent = tsxStr
		}
	} else {
		mainContent = tsxStr
	}

	// Convert JSX to React.createElement calls (only on main content)
	jsxStr := convertJSXToReactCreateElement(mainContent)
	// log.Println("convertJSXToReactCreateElement: ", jsxStr)

	// Combine imports with converted JSX
	jsxStr = imports + jsxStr

	// Fix className case (HTML parser converts to lowercase)
	jsxStr = strings.ReplaceAll(jsxStr, "{classname:", "{className:")
	jsxStr = strings.ReplaceAll(jsxStr, "classname:", "className:")
	// log.Println("className: ", jsxStr)

	// Remove import/export statements
	importPattern := regexp.MustCompile(`(?m)^import\s+.*?from\s+.*?;?\s*$`)
	jsxStr = importPattern.ReplaceAllString(jsxStr, "")

	exportPattern := regexp.MustCompile(`export\s+default\s+`)
	jsxStr = exportPattern.ReplaceAllString(jsxStr, "")
	// log.Println("FINAL JSX: ", jsxStr)

	return jsxStr
}

// removeHTMLComments removes all HTML comments from the content
func removeHTMLComments(content string) string {
	// Remove HTML comments: <!-- ... -->
	return commentPattern.ReplaceAllString(content, "")
}

# OpenAgent

## Features:

## Stories:

### Story 001: Landing page routing.
[ ] Task: If the landing page
[ ] Task: Implement host-based project routing - check for project by host, serve landing page from database (is_landing=true) or fallback to index.html 

## OpenAgent Template System

## 🚀 Functional Features

### Template Transpilation
- **HTML to JSX Conversion**: Automatically converts HTML files to TypeScript JSX components
- **Go Template Cleanup**: Removes Go template syntax (`{{.}}`, `{{define}}`, `{{template}}`) and converts to JSX
- **Self-Closing Tag Fix**: Automatically adds `/>` to HTML self-closing tags for JSX compatibility
- **Alpine.js Attribute Conversion**: Converts Alpine.js attributes (`@click`, `:disabled`, `x-data`) to `data-` prefixed attributes
- **HTML Comment Conversion**: Converts HTML comments (`<!-- -->`) to JSX comments (`{/* */}`)
- **CamelCase Component Names**: Converts hyphenated filenames to PascalCase component names

### Layout System
- **Automatic Layout Detection**: Pages without `<html>`/`<head>` tags automatically use layout wrappers
- **Fragment Fallback**: Pages with full HTML structure use React fragments (`<>...</>`)
- **Layout Function Signature**: `{page, children, linkPaths, scriptPaths}` with proper TypeScript typing
- **Dynamic Tag Injection**: Automatically injects CSS/JS tags based on comma-separated paths
- **Layout Validation**: Warns if layout files don't contain required `<html>`, `<head>`, `<body>` tags

### View Override System
- **Custom View Overrides**: Place custom TSX files in `tpl/views/` to override generated templates
- **Automatic Copying**: View overrides are automatically copied to generated directory during transpilation
- **Import Path Fixing**: Automatically adjusts import paths when copying view files
- **Thread-Safe Caching**: Caches view file existence checks to reduce disk I/O
- **Directory-Based Resolution**: Searches directories in priority order: `views/` → `pages/` → `app/` → `admin/`

### Asset Management
- **CSS/JS Extraction**: Automatically extracts inline `<style>` and `<script>` content to separate files
- **Dynamic Asset Injection**: Creates link and script tags dynamically from comma-separated paths
- **HTTP Handlers**: Serves extracted CSS/JS files via `/tsx/css/` and `/tsx/js/` routes
- **Smart Placement**: CSS links placed in `<head>`, JS scripts placed before `</body>`

### TypeScript Integration
- **Type Inference**: Infers TypeScript types from filename patterns (e.g., `test.landing_page.html` → `LandingPage`)
- **Two-Dot Notation**: `component.pageType.html` generates `{page: PageType}` typing
- **Three-Dot Notation**: `component.pageType.layoutName.html` generates custom layout with `{page: PageType}` typing
- **Default Types**: Uses `Page` type when no specific type is inferred

## 🔧 Technical Features

### Template Engine Architecture
- **TemplateEngine Wrapper**: Unified interface for different template engines (currently Wax)
- **Future-Proof Design**: Easy to switch template engines by changing the wrapper implementation
- **Compatibility Layer**: Maintains compatibility with existing Go template interface

### Directory Structure
```
tpl/
├── pages/          # Public pages (marketing, login, etc.)
├── admin/          # Admin-only pages
├── app/            # Logged-in user pages
├── layouts/        # Layout templates
├── views/          # Custom view overrides
└── generated/      # Auto-generated TSX files
    ├── pages/
    ├── admin/
    ├── app/
    ├── layouts/
    ├── css/
    └── js/
```

### Performance Optimizations
- **Compiled Regex Patterns**: Pre-compiled regex patterns for better performance
- **File Existence Caching**: Thread-safe caching of file existence checks
- **Efficient Asset Processing**: Batch processing of CSS/JS extraction
- **Smart Template Resolution**: Priority-based directory searching with caching

### Error Handling
- **Graceful Fallbacks**: Falls back to fragments when layouts fail
- **Validation Warnings**: Warns about missing required tags in layouts
- **Import Path Recovery**: Automatically fixes import paths during view copying
- **Build Constraint Handling**: Handles Alpine.js attribute conversion for JSX compatibility

## 📚 In-Depth Documentation

### _common.js - Alpine.js Integration

The `_common.js` file is automatically included in all pages and provides essential Alpine.js functionality:

#### Purpose
- **Alpine.js Plugin**: Sets up Alpine.js with `data-` prefix for JSX compatibility
- **Custom Directives**: Provides custom Alpine.js directives for common interactions
- **Event Handling**: Handles form submissions and click events with prevent default options

#### Key Features
```javascript
// Sets Alpine prefix to 'data' for JSX compatibility
Alpine.prefix('data');

// Custom directives for form and click handling
Alpine.directive('submit-prevent', ...)  // Prevents default form submission
Alpine.directive('click-prevent', ...)   // Prevents default click behavior
Alpine.directive('click', ...)           // Regular click handling
Alpine.directive('submit', ...)          // Regular form submission
```

#### Usage in Templates
```html
<!-- These Alpine.js attributes are automatically converted -->
<form data-submit-prevent="handleSubmit">
<button data-click-prevent="handleClick">
<div data-x-data="formData">
```

### Dynamic Tag Injection System

#### How It Works
1. **Path Processing**: CSS/JS paths are provided as comma-separated strings
2. **Dynamic Generation**: JSX map functions split paths and create tags
3. **Smart Placement**: Tags are injected at appropriate locations in layout

#### Generated Code Example
```jsx
// Link tags in <head>
{linkPaths && linkPaths.split(',').map((path, index) => (
    <link key={'gen-link-'+index} rel="stylesheet" href={path.trim()} />
))}

// Script tags before </body>
{scriptPaths && scriptPaths.split(',').map((path, index) => (
    <script key={'gen-script-'+index} src={path.trim()}></script>
))}
```

### View Override System

#### How to Use
1. **Create Override**: Place custom TSX file in `tpl/views/pages/component.tsx`
2. **Automatic Detection**: System automatically detects and uses override
3. **Import Path Fixing**: Import paths are automatically adjusted for generated directory

#### Example Override
```tsx
// tpl/views/pages/test.tsx
import Layout from '../layouts/layout_pages';

export default function Test({page}: {page: Page}) {
    return (
        <Layout page={page} linkPaths={``} scriptPaths={`/tsx/js/_common.js`}>
            <div className="container">
                <h1>Custom View Override: {page.PageTitle}</h1>
                <p>This overrides the generated template!</p>
            </div>
        </Layout>
    );
}
```

### TypeScript Type Inference

#### Filename Patterns
- `test.html` → `{page: Page}` (default)
- `test.landing_page.html` → `{page: LandingPage}` (two-dot notation)
- `index.page.landing.html` → `{page: Page}` with `layout_landing` (three-dot notation)

#### Generated Function Signatures
```tsx
// Default type
export default function Test({page}: {page: Page}) { ... }

// Inferred type
export default function Test({page}: {page: LandingPage}) { ... }

// Custom layout
export default function Index({page}: {page: Page}) {
    return <LayoutLanding page={page} ...>
}
```

### Asset Extraction Process

#### CSS Extraction
1. **Inline Style Detection**: Finds `<style>` tags in HTML
2. **Content Extraction**: Extracts CSS content to separate file
3. **File Creation**: Creates `tpl/generated/*/css/component.css`
4. **Link Generation**: Creates `<link>` tag with path `/tsx/css/component.css`

#### JavaScript Extraction
1. **Inline Script Detection**: Finds `<script>` tags in HTML
2. **Content Processing**: Converts Go template variables (`{{.Var}}` → `page.Var`)
3. **File Creation**: Creates `tpl/generated/*/js/component.js`
4. **Script Generation**: Creates `<script>` tag with path `/tsx/js/component.js`

### Layout System Architecture

#### Layout Detection Logic
```go
needsLayout := !strings.Contains(htmlContent, "<html") && !strings.Contains(htmlContent, "<head")
```

#### Layout Wrapper vs Fragment
- **Layout Wrapper**: Used for pages without full HTML structure
- **Fragment**: Used for pages with complete HTML documents
- **Dynamic Injection**: CSS/JS tags are injected appropriately for each case

#### Layout Function Signature
```tsx
export default function LayoutPages({
    page, 
    children, 
    linkPaths, 
    scriptPaths
}: {
    page: any, 
    children: any, 
    linkPaths?: string, 
    scriptPaths?: string
}) {
    // Dynamic tag injection happens here
}
```

## 🛠️ Development Workflow

### Adding New Pages
1. Create HTML file in appropriate directory (`pages/`, `admin/`, `app/`)
2. System automatically transpiles to TSX during server startup
3. Use view overrides in `tpl/views/` for custom implementations

### Creating Custom Layouts
1. Create layout HTML file in `tpl/layouts/`
2. Ensure it contains `<html>`, `<head>`, and `<body>` tags
3. System automatically generates TSX layout component

### Asset Management
1. Place CSS/JS in inline `<style>`/`<script>` tags
2. System automatically extracts to separate files
3. Assets are served via HTTP handlers at `/tsx/css/` and `/tsx/js/`

### TypeScript Integration
1. Use two-dot notation for custom types: `component.pageType.html`
2. Use three-dot notation for custom layouts: `component.pageType.layoutName.html`
3. System automatically infers types and generates proper function signatures

This template system provides a powerful, flexible foundation for building modern web applications with automatic transpilation, layout management, and asset optimization.
# OpenAgent System Architecture & Development Guidelines

## 🏗️ Core Architecture Principles

### 1. Component Structure (CRITICAL - DO NOT BREAK)
The application follows a strict architectural pattern that MUST be maintained:

#### Top-Level Page Components (e.g., `test.tsx`)
- **MUST** be named `{PageName}Page` (e.g., `TestPage`, `LoginPage`, `DashboardPage`)
- **MUST** import and use a `Layout` component from `../layouts/layout_pages` or `layour_admin` etc.
- **MUST** import an `App` component from `./{pagename}.component`
- **MUST** pass props to the Layout component, not render layout HTML directly, via CustomLayout if required.

```tsx
// CORRECT structure for test.tsx
import Layout from '../layouts/layout_pages';
import App from './test.component';

export default function TestPage({page}: {page: Page}) {
    return (
        <Layout page={page} linkPaths={linkPaths} scriptPaths={scriptPaths}>
            <App page={page} />
        </Layout>
    );
}
```

#### Component Files (e.g., `test.component.tsx`)
- Contains the actual page content and logic
- Receives `page` props from the parent `TestPage`

#### Layout Components (e.g., `layout_pages.tsx`)
- Handles the overall page structure (header, footer, navigation)
- Receives `page`, `linkPaths`, `scriptPaths` as props
- Renders meta tags, CSS links, and JavaScript includes

### 2. Meta Tag Architecture
Meta tags are handled through a structured approach:

1. **Extraction**: Meta tags are extracted from page HTML using `extractMetaTagsStructured()`
2. **Conversion**: Converted to JavaScript arrays (`contentMeta` and `propMeta`)
3. **Injection**: Passed to Layout component via props
4. **Rendering**: Layout component maps over arrays to render actual `<meta>` tags

```tsx
// In layout_pages.html
{page.contentMeta && page.contentMeta.map((meta: any) => (
    <meta name={meta.name} content={meta.content} />
))}
{page.propMeta && page.propMeta.map((meta: any) => (
    <meta property={meta.property} content={meta.content} />
))}
```

### 3. CSS/JS Asset Architecture (CRITICAL)
The application uses a **4-step generation process** for page-specific assets:

#### CSS Extraction & Naming Convention
- **NO CONSOLIDATION** - Each page gets its own CSS file
- **Naming Convention**: `pages_{pagename}.css` (e.g., `pages_test.css`, `pages_login.css`)
- **Path Structure**: `/tsx/css/pages_{pagename}.css`
- **Extraction**: CSS is extracted from page HTML and saved as separate files
- **Injection**: CSS paths are passed to Layout via `linkPaths` prop

#### JavaScript Generation (4-Step Process)
1. **Main JS**: `pages_{pagename}.js` - Page-specific JavaScript
2. **Component JS**: `{pagename}.component.js` - Component-specific JavaScript  
3. **Common JS**: `_common.js` - Shared JavaScript across all pages
4. **Hydration Code**: React hydration code for component initialization

#### Asset Path Structure
```
/tsx/css/pages_test.css          # Page-specific CSS
/tsx/js/pages_test.js            # Page-specific JS
/tsx/js/test.component.js        # Component-specific JS
/tsx/js/_common.js               # Common/shared JS
```

#### Integration with Layout
```tsx
// CSS injection in layout
{linkPaths && linkPaths.split(',').map((link: string) => (
    <link rel="stylesheet" href={link} />
))}

// JS injection in layout
{scriptPaths && scriptPaths.split(',').map((script: string) => (
    <script type="text/javascript" src={script} />
))}
```

## 🚫 Common Mistakes to AVOID

### Architecture Violations
- ❌ **NEVER** name top-level components `{PageName}Layout`
- ❌ **NEVER** render layout HTML directly in page components
- ❌ **NEVER** break the Layout → App component hierarchy
- ❌ **NEVER** use `dangerouslySetInnerHTML` for meta tags (causes HTML encoding issues), instead of loops over structured extractions.

### Meta Tag Issues
- ❌ **NEVER** pass meta tags as raw HTML strings
- ❌ **NEVER** use `key` attributes for one-time rendered meta tags
- ❌ **NEVER** place meta tags in the body section

### Deployment Issues
- ❌ **NEVER** deploy without explicit user permission
- ❌ **NEVER** assume deployment is needed after changes
- ❌ **ALWAYS** test locally first

## 🔄 Development Workflow

### 1. Local Development
```bash
# Start development server
go run . server

# Run tests
make test

# Build for production
go build -ldflags="-s -w" .
```

### 2. File Watching
- The server automatically watches `tpl/` directory for changes
- Templates are transpiled on file changes
- Generated files go to `tpl/generated/`

### 3. Testing Process
1. Make changes to templates or Go code
2. Test locally using `go run . server`
3. Verify changes work as expected
4. Run `make test` for comprehensive testing
5. **WAIT** for user approval before deployment

### 4. Deployment Process
- **ONLY** deploy when explicitly requested by user
- Uses Docker Compose for containerized deployment
- SCP-based deployment copies files and runs `docker compose up -d --build`
- Backup existing deployments before overwriting
- Test on remote server after deployment

## 📁 File Structure

```
openagent/
├── tpl/
│   ├── pages/           # Source HTML pages
│   ├── layouts/         # Layout templates
│   ├── components/      # Reusable components
│   └── generated/       # Transpiled TSX files
├── server/
│   └── transpile/       # Transpilation logic
├── static/              # Static assets
├── data/sql/           # Database queries
└── migrations/         # Database migrations
```

## 🛠️ Key Functions & Files

### Transpilation (`server/transpile/`)
- `templates.go`: Main transpilation logic
- `misc.go`: Utility functions for meta tag extraction
- `jsx.go`: JSX-specific transpilation

### Meta Tag Functions (`misc.go`)
- `extractMetaTagsStructured()`: Extracts meta tags into structured data
- `convertMetaTagsToJS()`: Converts to JavaScript arrays
- `removeMetaTags()`: Removes meta tags from HTML content

### Layout Templates
- `layout_pages.html`: Main application layout
- `layout_admin.html`: Admin panel layout
- `layout_marketing.html`: Marketing pages layout

## 🧪 Testing Guidelines

### Before Making Changes
1. Understand the current architecture
2. Read existing code patterns
3. Test changes locally first
4. Verify no regressions

### After Making Changes
1. Run `make test` to ensure all tests pass
2. Test the specific functionality manually
3. Check generated files for correctness
4. Verify meta tags render properly

### Common Test Commands
```bash
# Run all tests
make test

# Test specific functionality
curl -s http://localhost:8800/test | grep -A 3 -B 3 "meta name"

# Check transpiled output
cat tpl/generated/pages/test.tsx
```

## 🚀 Deployment Guidelines

### Pre-Deployment Checklist
- [ ] All tests pass locally
- [ ] Changes work as expected
- [ ] No regressions introduced
- [ ] User has explicitly approved deployment

### Deployment Commands
```bash
# Deploy to remote server
make deploy

# Check deployment status (Docker Compose)
ssh root@in.msheriff.com "cd /var/www/openagent && docker compose ps"

# View container logs
ssh root@in.msheriff.com "cd /var/www/openagent && docker compose logs -f app"

# Restart services
ssh root@in.msheriff.com "cd /var/www/openagent && docker compose restart"
```

## 📝 Code Style & Best Practices

### Go Code
- Use descriptive function names
- Add comments for complex logic
- Handle errors properly
- Follow Go naming conventions

### HTML/JSX Templates
- Use semantic HTML
- Maintain consistent indentation
- Use React fragments (`<>`) instead of divs when appropriate
- Avoid unnecessary attributes (like `key` for static content)

### Meta Tags
- Use structured data approach
- Separate `name` and `property` meta tags
- Keep meta tags in the `<head>` section
- Use proper content values

## 🔧 Troubleshooting

### Common Issues
1. **Port already in use**: Kill existing processes with `lsof -ti:8800 | xargs kill -9`
2. **Meta tags in body**: Check transpilation logic and layout structure
3. **HTML encoding**: Use structured approach instead of raw HTML strings
4. **Architecture broken**: Revert to proper Layout → App component structure

### Debug Commands
```bash
# Check running processes
lsof -i :8800

# View server logs
tail -f /var/log/openagent.log

# Test specific routes
curl -v http://localhost:8800/test
```

## 🎯 Success Criteria

A successful implementation should:
- ✅ Maintain the established architecture
- ✅ Render meta tags correctly in the head section
- ✅ Pass all tests
- ✅ Work locally before deployment
- ✅ Not break existing functionality
- ✅ Follow the user's explicit requirements

## 📞 Communication Guidelines

- **ALWAYS** ask for clarification if requirements are unclear
- **NEVER** assume deployment is needed
- **ALWAYS** test changes locally first
- **RESPECT** the user's architectural decisions
- **DOCUMENT** any changes that affect the architecture

---

**Remember**: The user has invested significant time in designing this architecture. Respect it, maintain it, and don't break it without explicit permission.

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dop251/goja"
	"github.com/evanw/esbuild/pkg/api"
)

// ESBuildGojaEngine represents the new esbuild + goja engine
type ESBuildGojaEngine struct {
	vm                    *goja.Runtime
	reactContent          string
	reactDOMServerContent string
}

// NewESBuildGojaEngine creates a new esbuild + goja engine
func NewESBuildGojaEngine() *ESBuildGojaEngine {
	vm := goja.New()
	engine := &ESBuildGojaEngine{vm: vm}

	// Load React development file
	reactContent, err := engine.loadReactDevelopment()
	if err != nil {
		log.Printf("Warning: Failed to load React development file: %v", err)
		reactContent = ""
	}
	engine.reactContent = reactContent

	// Load React DOM Server development file
	reactDOMServerContent, err := engine.loadReactDOMServerDevelopment()
	if err != nil {
		log.Printf("Warning: Failed to load React DOM Server development file: %v", err)
		reactDOMServerContent = ""
	}
	engine.reactDOMServerContent = reactDOMServerContent

	return engine
}

// loadReactDevelopment loads the React development file
func (e *ESBuildGojaEngine) loadReactDevelopment() (string, error) {
	// Try to find the react.development.js file in the current directory or parent directories
	reactPath := "react.development.js"

	// Check if file exists in current directory
	if _, err := os.Stat(reactPath); os.IsNotExist(err) {
		// Try parent directory
		reactPath = filepath.Join("..", "react.development.js")
		if _, err := os.Stat(reactPath); os.IsNotExist(err) {
			// Try project root
			reactPath = filepath.Join("..", "..", "react.development.js")
			if _, err := os.Stat(reactPath); os.IsNotExist(err) {
				return "", fmt.Errorf("react.development.js not found in current directory or parent directories")
			}
		}
	}

	content, err := os.ReadFile(reactPath)
	if err != nil {
		return "", fmt.Errorf("failed to read React development file: %v", err)
	}

	return string(content), nil
}

// loadReactDOMServerDevelopment loads the React DOM Server development file
func (e *ESBuildGojaEngine) loadReactDOMServerDevelopment() (string, error) {
	// Try to find the react-dom-server.development.js file in the current directory or parent directories
	reactDOMServerPath := "react-dom-server.development.js"

	// Check if file exists in current directory
	if _, err := os.Stat(reactDOMServerPath); os.IsNotExist(err) {
		// Try parent directory
		reactDOMServerPath = filepath.Join("..", "react-dom-server.development.js")
		if _, err := os.Stat(reactDOMServerPath); os.IsNotExist(err) {
			// Try project root
			reactDOMServerPath = filepath.Join("..", "..", "react-dom-server.development.js")
			if _, err := os.Stat(reactDOMServerPath); os.IsNotExist(err) {
				return "", fmt.Errorf("react-dom-server.development.js not found in current directory or parent directories")
			}
		}
	}

	content, err := os.ReadFile(reactDOMServerPath)
	if err != nil {
		return "", fmt.Errorf("failed to read React DOM Server development file: %v", err)
	}

	return string(content), nil
}

// TransformTSXToJS transforms TSX content to JS using esbuild (unified for SSR and client)
func (e *ESBuildGojaEngine) TransformTSXToJS(tsxContent string) (string, error) {
	result := api.Build(api.BuildOptions{
		Stdin: &api.StdinOptions{
			Contents:   tsxContent,
			ResolveDir: ".",
			Loader:     api.LoaderTSX,
		},
		JSXFactory:  "React.createElement",
		JSXFragment: "React.Fragment",
		Bundle:      true,
		Format:      api.FormatIIFE,
		GlobalName:  "Component",
		External:    []string{"react"},
		Write:       false,
		Loader: map[string]api.Loader{
			".js": api.LoaderJSX,
		},
	})

	if len(result.Errors) > 0 {
		return "", fmt.Errorf("esbuild errors: %v", result.Errors)
	}

	jsOutput := string(result.OutputFiles[0].Contents)
	return jsOutput, nil
}

// TransformTSXToBrowserIIFE builds a browser-ready IIFE exposing default as CounterJS.default
func (e *ESBuildGojaEngine) TransformTSXToBrowserIIFE(tsxContent string) (string, error) {
	result := api.Build(api.BuildOptions{
		Stdin: &api.StdinOptions{
			Contents:   tsxContent,
			ResolveDir: ".",
			Loader:     api.LoaderTSX,
		},
		JSXFactory:  "React.createElement",
		JSXFragment: "React.Fragment",
		Bundle:      true,
		Format:      api.FormatIIFE,
		GlobalName:  "CounterJS",
		Write:       false,
	})
	if len(result.Errors) > 0 {
		return "", fmt.Errorf("esbuild browser errors: %v", result.Errors)
	}
	return string(result.OutputFiles[0].Contents), nil
}

// TransformTSXToESM builds a browser-ready ES module without bundling or helpers
func (e *ESBuildGojaEngine) TransformTSXToESM(tsxContent string) (string, error) {
	result := api.Build(api.BuildOptions{
		Stdin: &api.StdinOptions{
			Contents:   tsxContent,
			ResolveDir: ".",
			Loader:     api.LoaderTSX,
		},
		JSXFactory:  "React.createElement",
		JSXFragment: "React.Fragment",
		Bundle:      false,
		Format:      api.FormatESModule,
		Platform:    api.PlatformBrowser,
		Write:       false,
	})
	if len(result.Errors) > 0 {
		return "", fmt.Errorf("esbuild esm errors: %v", result.Errors)
	}
	return string(result.OutputFiles[0].Contents), nil
}

// TransformTSXToIIFEUnbundled builds a minimal IIFE without bundling/helpers (client-side)
func (e *ESBuildGojaEngine) TransformTSXToIIFEUnbundled(tsxContent string) (string, error) {
	result := api.Build(api.BuildOptions{
		Stdin: &api.StdinOptions{
			Contents:   tsxContent,
			ResolveDir: ".",
			Loader:     api.LoaderTSX,
		},
		JSXFactory:  "React.createElement",
		JSXFragment: "React.Fragment",
		Bundle:      false,
		Format:      api.FormatIIFE,
		GlobalName:  "Component",
		Platform:    api.PlatformBrowser,
		Write:       false,
	})
	if len(result.Errors) > 0 {
		return "", fmt.Errorf("esbuild iife errors: %v", result.Errors)
	}
	return string(result.OutputFiles[0].Contents), nil
}

// ExecuteJS executes JavaScript code in the goja runtime
func (e *ESBuildGojaEngine) ExecuteJS(jsCode string) (goja.Value, error) {
	return e.vm.RunString(jsCode)
}

// RenderComponent renders a component using the goja runtime
func (e *ESBuildGojaEngine) RenderComponent(componentName string, props map[string]interface{}) (string, error) {
	// Set up mock Node.js environment first
	_, err := e.vm.RunString(`
		if (typeof process === 'undefined') {
			process = {
				env: { NODE_ENV: 'development' },
				version: 'v18.0.0',
				platform: 'darwin',
				arch: 'x64'
			};
		}
		if (typeof require === 'undefined') {
			require = function(id) {
				if (id === 'react') {
					return React;
				}
				if (id === 'react-dom') {
					return ReactDOM;
				}
				throw new Error('Module not found: ' + id);
			};
		}
		if (typeof module === 'undefined') {
			module = { exports: {} };
		}
		if (typeof exports === 'undefined') {
			exports = module.exports;
		}
		if (typeof MessageChannel === 'undefined') {
			MessageChannel = function() {
				return {
					port1: { onmessage: null, postMessage: function() {} },
					port2: { onmessage: null, postMessage: function() {} }
				};
			};
		}
		if (typeof setTimeout === 'undefined') {
			setTimeout = function(fn, delay) { return 1; };
		}
		if (typeof clearTimeout === 'undefined') {
			clearTimeout = function(id) {};
		}
		if (typeof TextEncoder === 'undefined') {
			TextEncoder = function() {
				this.encode = function(str) {
					var bytes = [];
					for (var i = 0; i < str.length; i++) {
						var char = str.charCodeAt(i);
						if (char < 0x80) {
							bytes.push(char);
						} else if (char < 0x800) {
							bytes.push(0xc0 | (char >> 6));
							bytes.push(0x80 | (char & 0x3f));
						} else {
							bytes.push(0xe0 | (char >> 12));
							bytes.push(0x80 | ((char >> 6) & 0x3f));
							bytes.push(0x80 | (char & 0x3f));
						}
					}
					return new Uint8Array(bytes);
				};
			};
		}
		if (typeof TextDecoder === 'undefined') {
			TextDecoder = function() {
				this.decode = function(bytes) {
					var str = '';
					for (var i = 0; i < bytes.length; i++) {
						str += String.fromCharCode(bytes[i]);
					}
					return str;
				};
			};
		}
		if (typeof console === 'undefined') {
			console = {
				log: function() {},
				error: function() {},
				warn: function() {},
				info: function() {},
				debug: function() {}
			};
		}
	`)
	if err != nil {
		return "", fmt.Errorf("failed to setup Node.js mocks: %v", err)
	}

	// Set up React in the runtime using the actual React development file
	if e.reactContent == "" {
		return "", fmt.Errorf("react development file not loaded")
	}

	_, err = e.vm.RunString(e.reactContent)
	if err != nil {
		return "", fmt.Errorf("failed to setup React: %v", err)
	}

	// Make React available globally from module.exports
	_, err = e.vm.RunString(`
		if (typeof module !== 'undefined' && module.exports) {
			React = module.exports;
		}
	`)
	if err != nil {
		return "", fmt.Errorf("failed to expose React globally: %v", err)
	}

	// Update the require function to use the loaded React
	_, err = e.vm.RunString(`
		require = function(id) {
			if (id === 'react') {
				return React;
			}
			if (id === 'react-dom') {
				// For server-side rendering, we don't need react-dom
				// Return a minimal mock that satisfies the server requirements
				return {
					render: function() {},
					hydrate: function() {},
					createRoot: function() { return { render: function() {} }; },
					createPortal: function() {},
					flushSync: function() {},
					findDOMNode: function() {},
					unmountComponentAtNode: function() {},
					version: '19.2.0'
				};
			}
			throw new Error('Module not found: ' + id);
		};
	`)
	if err != nil {
		return "", fmt.Errorf("failed to update require function: %v", err)
	}

	// Skip React DOM Server loading and use custom HTML renderer

	// Prepare props for SSR
	var propsJSON string
	if props == nil {
		propsJSON = "null"
	} else {
		b, _ := json.Marshal(props)
		propsJSON = string(b)
	}

	// Build SSR JS that renders with props using unified Component.default
	ssrJS := fmt.Sprintf(`
		var __SSR_PROPS = %s;
        function renderToString(element) {
			if (element === null || element === undefined || element === false || element === true) { return ''; }
			if (typeof element === 'string' || typeof element === 'number') { return String(element); }
			if (Array.isArray(element)) { return element.map(renderToString).join(''); }
			if (typeof element !== 'object') { return ''; }
			if (element.type) {
                var tag = element.type; var props = element.props || {}; var children = props.children || [];
                if (typeof tag === 'function') {
                    var propsWithChildren = props;
                    if (children !== undefined) { propsWithChildren = Object.assign({}, props, { children: children }); }
                    try { var rendered = tag(propsWithChildren); return renderToString(rendered); } catch (e) { return ''; }
                }
				var selfClosingTags = ['img','br','hr','input','meta','link'];
                if (selfClosingTags.includes(tag)) {
					var attrs = ''; for (var key in props) { if (key !== 'children' && props[key] != null) { var attrName = key === 'className' ? 'class' : key; attrs += ' ' + attrName + '="' + String(props[key]).replace(/"/g,'&quot;') + '"'; } }
                    return '<' + tag + attrs + ' />';
				}
				var attrs = ''; for (var key in props) { if (key !== 'children' && props[key] != null) { var attrName = key === 'className' ? 'class' : key; attrs += ' ' + attrName + '="' + String(props[key]).replace(/"/g,'&quot;') + '"'; } }
				var childrenHtml = ''; if (Array.isArray(children)) { childrenHtml = children.map(renderToString).join(''); } else if (children) { childrenHtml = renderToString(children); }
                return '<' + tag + attrs + '>' + childrenHtml + '</' + tag + '>';
			}
			return '';
		}
		renderToString(Component.default(__SSR_PROPS))
	`, propsJSON)

	result, err := e.vm.RunString(ssrJS)
	if err != nil {
		return "", fmt.Errorf("failed to render component to string: %v", err)
	}
	return result.String(), nil
}

// TestServer represents our test server
type TestServer struct {
	engine *ESBuildGojaEngine
}

// NewTestServer creates a new test server
func NewTestServer() *TestServer {
	return &TestServer{
		engine: NewESBuildGojaEngine(),
	}
}

// ensureClientCommonBundle downloads vendor React CJS files if missing and builds static/common_client.js
func ensureClientCommonBundle() error {
	_ = os.MkdirAll("static/vendor", 0755)
	reactPath := filepath.Join("static", "vendor", "react.production.js")
	domClientPath := filepath.Join("static", "vendor", "react-dom-client.production.js")

	if _, err := os.Stat(reactPath); os.IsNotExist(err) {
		resp, err := http.Get("https://unpkg.com/react@19.2.0/cjs/react.production.js")
		if err != nil {
			return fmt.Errorf("download react.production.js: %v", err)
		}
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read react.production.js: %v", err)
		}
		if err := os.WriteFile(reactPath, b, 0644); err != nil {
			return fmt.Errorf("write react.production.js: %v", err)
		}
	}
	if _, err := os.Stat(domClientPath); os.IsNotExist(err) {
		resp, err := http.Get("https://unpkg.com/react-dom@19.2.0/cjs/react-dom-client.production.js")
		if err != nil {
			return fmt.Errorf("download react-dom-client.production.js: %v", err)
		}
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read react-dom-client.production.js: %v", err)
		}
		if err := os.WriteFile(domClientPath, b, 0644); err != nil {
			return fmt.Errorf("write react-dom-client.production.js: %v", err)
		}
	}

	entry := strings.Join([]string{
		"var React = require(\"./static/vendor/react.production.js\");",
		"var ReactDOMClient = require(\"./static/vendor/react-dom-client.production.js\");",
		"window.React = React;",
		"window.ReactDOMClient = ReactDOMClient;",
		"window.mountIntoMain = function(Component, props){",
		"  var el = document.querySelector('main#preview') || document.querySelector('main');",
		"  if (!el) return;",
		"  var root = ReactDOMClient.createRoot(el);",
		"  root.render(React.createElement(Component, props || {}));",
		"};",
	}, "\n")

	build := api.Build(api.BuildOptions{
		Stdin: &api.StdinOptions{
			Contents:   entry,
			ResolveDir: ".",
			Sourcefile: "common_client_entry.js",
			Loader:     api.LoaderJS,
		},
		Bundle:   true,
		Platform: api.PlatformBrowser,
		Format:   api.FormatIIFE,
		Write:    false,
	})
	if len(build.Errors) > 0 {
		return fmt.Errorf("common client bundle errors: %v", build.Errors)
	}
	out := build.OutputFiles[0].Contents
	if err := os.WriteFile(filepath.Join("static", "common_client.js"), out, 0644); err != nil {
		return fmt.Errorf("write common_client.js: %v", err)
	}
	return nil
}

// buildPagesIndexJS builds static/pages_index.js that registers page components and mounts templates
func buildPagesIndexJS(engine *ESBuildGojaEngine) error {
	compPath := filepath.Join("components", "counter.html")
	b, err := os.ReadFile(compPath)
	if err != nil {
		return err
	}
	compSrc := string(b)
	var scriptCode string
	htmlOnly := compSrc
	if si := strings.Index(strings.ToLower(compSrc), "<script>"); si >= 0 {
		ei := strings.Index(strings.ToLower(compSrc), "</script>")
		if ei > si {
			scriptCode = compSrc[si+len("<script>") : ei]
			htmlOnly = compSrc[:si] + compSrc[ei+len("</script>"):]
		}
	}
	htmlOnly = strings.ReplaceAll(htmlOnly, " class=\"", " className=\"")
	tsx := "export default function Counter(props:any,state:any){\n" + strings.TrimSpace(scriptCode) + "\n  return (\n    " + strings.TrimSpace(htmlOnly) + "\n  );\n}\n"
	clientIIFE, err := engine.TransformTSXToIIFEUnbundled(tsx)
	if err != nil {
		return err
	}
	_ = os.MkdirAll("static", 0755)
	bootstrap := `(function(){var tpls=document.querySelectorAll('template[id^="component-"]');tpls.forEach(function(t){var name=t.id.replace('component-','');var C=window.Components[name];if(C){var mount=document.createElement('div');t.replaceWith(mount);var root=window.ReactDOMClient.createRoot(mount);root.render(window.React.createElement(C, {}));}});})();\n`
	pagesIndex := clientIIFE + "\n;window.Components=window.Components||{};window.Components['counter']=(Component && (Component.default||Component));\n" + bootstrap
	return os.WriteFile(filepath.Join("static", "pages_index.js"), []byte(pagesIndex), 0644)
}

// writeCommonSSR writes a file with SSR React for inspection (server continues to load as before)
func writeCommonSSR(reactContent string) error {
	wrapper := strings.Join([]string{
		"var module = {exports:{}};",
		reactContent,
		"if (typeof React === 'undefined') { React = module.exports; }",
	}, "\n")
	_ = os.MkdirAll("static", 0755)
	return os.WriteFile(filepath.Join("static", "common_ssr.js"), []byte(wrapper), 0644)
}

// handleTest handles the test endpoint
func (s *TestServer) handleTest(w http.ResponseWriter, r *http.Request) {
	// Prepare initial state
	initialState := map[string]interface{}{
		"count": 3,
	}

	// Ensure client common bundle
	if err := ensureClientCommonBundle(); err != nil {
		log.Printf("warn: ensure client bundle: %v", err)
	}

	// Also write SSR common for inspection (not used by engine)
	if s.engine.reactContent != "" {
		_ = writeCommonSSR(s.engine.reactContent)
	}

	// Read source component HTML (Plan A)
	sourcePath := filepath.Join("components", "counter.html")
	srcBytes, err := os.ReadFile(sourcePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Read component error: %v", err), http.StatusInternalServerError)
		return
	}
	src := string(srcBytes)

	// Extract optional <script>...</script> and outer HTML
	var scriptCode string
	var htmlOnly string = src
	if start := strings.Index(strings.ToLower(src), "<script>"); start >= 0 {
		end := strings.Index(strings.ToLower(src), "</script>")
		if end > start {
			scriptCode = src[start+len("<script>") : end]
			htmlOnly = src[:start] + src[end+len("</script>"):]
		}
	}

	// Normalize HTML: convert class -> className attribute
	htmlOnly = strings.ReplaceAll(htmlOnly, " class=\"", " className=\"")

	// Build Counter.tsx (Plan A): single function that contains script then returns HTML
	tsxContent := "export default function Counter(props: any, state: any) {\n" +
		strings.TrimSpace(scriptCode) + "\n" +
		"  return (\n" +
		"    " + strings.TrimSpace(htmlOnly) + "\n" +
		"  );\n" +
		"}\n"

	// Ensure generated directory exists and write files
	_ = os.MkdirAll("generated", 0755)
	_ = os.WriteFile(filepath.Join("generated", "Counter.tsx"), []byte(tsxContent), 0644)

	// Transform TSX to unified JS (used for both SSR and client)
	jsCode, err := s.engine.TransformTSXToJS(tsxContent)
	if err != nil {
		http.Error(w, fmt.Sprintf("Transform error: %v", err), http.StatusInternalServerError)
		return
	}
	_ = os.WriteFile(filepath.Join("generated", "Counter.js"), []byte(jsCode), 0644)

	// Create browser pages bundle as pure ESM (no helpers), imported by test.html
	esmJS, err := s.engine.TransformTSXToESM(tsxContent)
	if err != nil {
		http.Error(w, fmt.Sprintf("ESM build error: %v", err), http.StatusInternalServerError)
		return
	}
	_ = os.MkdirAll("static", 0755)
	_ = os.WriteFile(filepath.Join("static", "pages_test.js"), []byte(esmJS), 0644)

	// Execute the JS (SSR path)
	_, err = s.engine.ExecuteJS(jsCode)
	if err != nil {
		http.Error(w, fmt.Sprintf("Execute error: %v", err), http.StatusInternalServerError)
		return
	}

	// Render the component with state as props
	result, err := s.engine.RenderComponent("Counter", initialState)
	if err != nil {
		http.Error(w, fmt.Sprintf("Render error: %v", err), http.StatusInternalServerError)
		return
	}

	// Return the result
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"result": %q, "state": %s}`, result, func() string { b, _ := json.Marshal(initialState); return string(b) }())
}

// handleHealth handles the health check
func (s *TestServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status": "ok", "engine": "esbuild-goja"}`)
}

// handleTestHTML serves the test HTML page
func (s *TestServer) handleTestHTML(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "test.html")
}

func (s *TestServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	initial := map[string]interface{}{
		"title":     "Welcome",
		"subtitle":  "Rendered via ESBuild + Goja",
		"siteTitle": "My Site",
	}
	// Ensure client common bundle exists
	if err := ensureClientCommonBundle(); err != nil {
		log.Printf("warn: ensure client bundle: %v", err)
	}
	// Eagerly build pages_index.js bundle
	_ = buildPagesIndexJS(s.engine)

	// Read pages/index.html and appropriate layout
	pagePath := filepath.Join("pages", "index.html")
	layoutPath := filepath.Join("layouts", "layout_index.html")
	pageBytes, err := os.ReadFile(pagePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Read page error: %v", err), http.StatusInternalServerError)
		return
	}
	layoutBytes, err := os.ReadFile(layoutPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Read layout error: %v", err), http.StatusInternalServerError)
		return
	}
    pageSrc := string(pageBytes)
    layoutSrc := string(layoutBytes)

	// Extract scripts and html for page
	var pageScript string
	pageHTML := pageSrc
	if sIdx := strings.Index(strings.ToLower(pageSrc), "<script>"); sIdx >= 0 {
		eIdx := strings.Index(strings.ToLower(pageSrc), "</script>")
		if eIdx > sIdx {
			pageScript = pageSrc[sIdx+len("<script>") : eIdx]
			pageHTML = pageSrc[:sIdx] + pageSrc[eIdx+len("</script>"):]
		}
	}
    pageHTML = strings.ReplaceAll(pageHTML, " class=\"", " className=\"")

    // Replace <template id="component-xxx"></template> with <Xxx /> and collect component locals
    var componentLocals strings.Builder
    tmplRe := regexp.MustCompile(`<template\s+id=["']component-([a-zA-Z0-9_\-]+)["']\s*></template>`)
    matches := tmplRe.FindAllStringSubmatch(pageHTML, -1)
    seen := map[string]bool{}
    for _, m := range matches {
        if len(m) < 2 { continue }
        raw := m[1]
        name := raw
        if len(name) > 0 {
            name = strings.ToUpper(name[:1]) + name[1:]
        }
        // Load component HTML and emit a local function
        compPath := filepath.Join("components", strings.ToLower(raw)+".html")
        if !seen[name] {
            if b, err := os.ReadFile(compPath); err == nil {
                src := string(b)
                var scriptCode string
                htmlOnly := src
                if si := strings.Index(strings.ToLower(src), "<script>"); si >= 0 {
                    ei := strings.Index(strings.ToLower(src), "</script>")
                    if ei > si {
                        scriptCode = src[si+len("<script>") : ei]
                        htmlOnly = src[:si] + src[ei+len("</script>"):]
                    }
                }
                htmlOnly = strings.ReplaceAll(htmlOnly, " class=\"", " className=\"")
                componentLocals.WriteString("function ")
                componentLocals.WriteString(name)
                componentLocals.WriteString("(props:any,state:any){\n")
                componentLocals.WriteString(strings.TrimSpace(scriptCode))
                componentLocals.WriteString("\n  return (\n    ")
                componentLocals.WriteString(strings.TrimSpace(htmlOnly))
                componentLocals.WriteString("\n  );\n}\n\n")
                // Write generated inspection file
                _ = os.WriteFile(filepath.Join("generated", name+".tsx"), []byte("export default "+name+"\n"), 0644)
            }
            seen[name] = true
        }
        // Replace template with component jsx tag
        pageHTML = strings.Replace(pageHTML, m[0], "<"+name+" />", 1)
    }

	// Extract scripts and html for layout
	var layoutScript string
	layoutHTML := layoutSrc
	if sIdx := strings.Index(strings.ToLower(layoutSrc), "<script>"); sIdx >= 0 {
		eIdx := strings.Index(strings.ToLower(layoutSrc), "</script>")
		if eIdx > sIdx {
			layoutScript = layoutSrc[sIdx+len("<script>") : eIdx]
			layoutHTML = layoutSrc[:sIdx] + layoutSrc[eIdx+len("</script>"):]
		}
	}
	layoutHTML = strings.ReplaceAll(layoutHTML, " class=\"", " className=\"")

    // Build Layout.tsx for inspection (exported)
	layoutTSX := "export default function Layout(props:any,state:any){\n" + strings.TrimSpace(layoutScript) + "\n  return (\n    " + strings.TrimSpace(layoutHTML) + "\n  );\n}\n"
	// Local (non-exported) version for composed bundle to avoid multiple default exports
	layoutLocalTSX := "function Layout(props:any,state:any){\n" + strings.TrimSpace(layoutScript) + "\n  return (\n    " + strings.TrimSpace(layoutHTML) + "\n  );\n}\n"
    // Build PageIndex.tsx that composes Layout as children
    pageTSX := "export default function PageIndex(props:any,state:any){\n" + strings.TrimSpace(pageScript) + "\n  return React.createElement(Layout, props, (\n    " + strings.TrimSpace(pageHTML) + "\n  ));\n}\n"
    // Compose with component locals and non-exported Layout to keep a single default export in module
    composedTSX := componentLocals.String() + layoutLocalTSX + "\n" + pageTSX

	_ = os.MkdirAll("generated", 0755)
	_ = os.WriteFile(filepath.Join("generated", "Layout.tsx"), []byte(layoutTSX), 0644)
	_ = os.WriteFile(filepath.Join("generated", "PageIndex.tsx"), []byte(pageTSX), 0644)

	// Build composed JS
	js, err := s.engine.TransformTSXToJS(composedTSX)
	if err != nil {
		http.Error(w, fmt.Sprintf("Transform error: %v", err), http.StatusInternalServerError)
		return
	}
	_ = os.WriteFile(filepath.Join("generated", "PageIndex.js"), []byte(js), 0644)
	// Execute for SSR
	_, err = s.engine.ExecuteJS(js)
	if err != nil {
		http.Error(w, fmt.Sprintf("Execute error: %v", err), http.StatusInternalServerError)
		return
	}
	// Render SSR
	html, err := s.engine.RenderComponent("PageIndex", initial)
	if err != nil {
		http.Error(w, fmt.Sprintf("Render error: %v", err), http.StatusInternalServerError)
		return
	}
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    _, _ = w.Write([]byte("<!DOCTYPE html>\n"))
    fmt.Fprint(w, html)
}

func main() {
	server := NewTestServer()

	// Set up routes
	http.HandleFunc("/", server.handleIndex)
	http.HandleFunc("/test", server.handleTest)
	http.HandleFunc("/health", server.handleHealth)
	http.HandleFunc("/test.html", server.handleTestHTML)

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/generated/", http.StripPrefix("/generated/", http.FileServer(http.Dir("generated"))))
	http.Handle("/components/", http.StripPrefix("/components/", http.FileServer(http.Dir("components"))))
	http.Handle("/pages/", http.StripPrefix("/pages/", http.FileServer(http.Dir("pages"))))

	// Start server
	port := "8801"
	fmt.Printf("🚀 ESBuild + Goja Test Server starting on port %s\n", port)
	fmt.Printf("📝 Test endpoint: http://localhost:%s/test\n", port)
	fmt.Printf("❤️  Health check: http://localhost:%s/health\n", port)
	fmt.Printf("🏠 Index page: http://localhost:%s/\n", port)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

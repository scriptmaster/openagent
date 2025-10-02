package transpile

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestRecWalkHTMLNodeWithCustomComponents(t *testing.T) {
	tests := []struct {
		name           string
		htmlContent    string
		expectedOutput string
		description    string
	}{
		{
			name:           "SimpleDiv",
			htmlContent:    `<div class="container">Hello World</div>`,
			expectedOutput: `React.createElement('div', {className: "container"}, 'Hello World')`,
			description:    "Simple div with class and text content",
		},
		{
			name:           "NestedElements",
			htmlContent:    `<div class="card"><h1>Title</h1><p>Content</p></div>`,
			expectedOutput: `React.createElement('div', {className: "card"}, React.createElement('h1', null, 'Title'), React.createElement('p', null, 'Content'))`,
			description:    "Nested elements with multiple children",
		},
		{
			name:           "CustomComponent",
			htmlContent:    `<Simple suppressHydrationWarning="true" />`,
			expectedOutput: `React.createElement(Simple, {suppressHydrationWarning: "true"})`,
			description:    "Custom component with props",
		},
		{
			name:           "TextWithInterpolation",
			htmlContent:    `Counter: {count}`,
			expectedOutput: `'Counter: ' + (count) + ''`,
			description:    "Text content with JSX interpolation",
		},
		{
			name:           "MultipleInterpolations",
			htmlContent:    `Hello {name}, you have {count} items and {times} times`,
			expectedOutput: `'Hello ' + (name) + ', you have ' + (count) + ' items and ' + (times) + ' times'`,
			description:    "Text with multiple JSX interpolations",
		},
		{
			name:           "InterpolationAtStart",
			htmlContent:    `{count} items remaining`,
			expectedOutput: `(count) + ' items remaining'`,
			description:    "JSX interpolation at the start of text",
		},
		{
			name:           "InterpolationAtEnd",
			htmlContent:    `Total: {count}`,
			expectedOutput: `'Total: ' + (count) + ''`,
			description:    "JSX interpolation at the end of text",
		},
		{
			name:           "OnlyInterpolation",
			htmlContent:    `{count}`,
			expectedOutput: `(count) + ''`,
			description:    "Text content that is only a JSX interpolation",
		},
		{
			name:           "MixedContent",
			htmlContent:    `<div>Hello {name}, you have {count} items</div>`,
			expectedOutput: `React.createElement('div', null, 'Hello ' + (name) + ', you have ' + (count) + ' items')`,
			description:    "Mixed HTML and interpolated text",
		},
		{
			name:           "MultipleAttributes",
			htmlContent:    `<input type="text" id="username" class="form-control" />`,
			expectedOutput: `React.createElement('input', {type: "text", id: "username", className: "form-control"})`,
			description:    "Element with multiple attributes",
		},
		{
			name:           "SelfClosingTag",
			htmlContent:    `<img src="/logo.png" alt="Logo" />`,
			expectedOutput: `React.createElement('img', {src: "/logo.png", alt: "Logo"})`,
			description:    "Self-closing HTML tag",
		},
		{
			name:           "EmptyDiv",
			htmlContent:    `<div></div>`,
			expectedOutput: `React.createElement('div', null)`,
			description:    "Empty div with no content",
		},
		{
			name:           "TextOnly",
			htmlContent:    `Just some text`,
			expectedOutput: `'Just some text'`,
			description:    "Plain text content only",
		},
		{
			name:           "WhitespaceText",
			htmlContent:    "   \n\t   ",
			expectedOutput: ``,
			description:    "Whitespace-only text should be ignored",
		},
		{
			name:           "Fragment",
			htmlContent:    `<><div>First</div><div>Second</div></>`,
			expectedOutput: `React.createElement('div', null, 'First'), React.createElement('div', null, 'Second')`,
			description:    "React Fragment should be removed and children processed directly",
		},
		{
			name:           "ComplexNested",
			htmlContent:    `<div class="app"><header class="header"><h1>App Title</h1></header><main class="content"><Simple suppressHydrationWarning="true"></Simple><p>Welcome {user}!</p></main></div>`,
			expectedOutput: `React.createElement('div', {className: "app"}, React.createElement('header', {className: "header"}, React.createElement('h1', null, 'App Title')), React.createElement('main', {className: "content"}, React.createElement(Simple, {suppressHydrationWarning: "true"}), React.createElement('p', null, 'Welcome ' + (user) + '!')))`,
			description:    "Complex nested structure with custom components and interpolations",
		},
		{
			name:           "AdjacentComponents",
			htmlContent:    `<div><Simple suppressHydrationWarning={true} /><Counter suppressHydrationWarning={true} /></div>`,
			expectedOutput: `React.createElement(Simple, {suppressHydrationWarning: true}), React.createElement(Counter, {suppressHydrationWarning: true})`,
			description:    "Adjacent custom components should have commas between them",
		},
		{
			name:           "JSXExpressions",
			htmlContent:    `<Simple suppressHydrationWarning={true} onClick={handleClick} className={dynamicClass} />`,
			expectedOutput: `React.createElement(Simple, {suppressHydrationWarning: true, onClick: handleClick, className: dynamicClass})`,
			description:    "JSX expressions should be converted to JavaScript values",
		},
		{
			name:           "TextInterpolation",
			htmlContent:    `<div>Counter: {count}</div>`,
			expectedOutput: `React.createElement('div', null, 'Counter: ' + (count) + '')`,
			description:    "Text interpolation should properly quote strings and unquote expressions",
		},
		{
			name:           "MultipleTextInterpolation",
			htmlContent:    `<div>Hello {name}, you have {count} items</div>`,
			expectedOutput: `React.createElement('div', null, 'Hello ' + (name) + ', you have ' + (count) + ' items')`,
			description:    "Multiple text interpolations should be properly formatted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use the full pipeline for Fragment and JSX expression tests to ensure proper handling
			if tt.name == "Fragment" || tt.name == "AdjacentComponents" || tt.name == "JSXExpressions" {
				// Use parseJSXWithHTMLParser for these tests to ensure proper preprocessing
				actualOutput := parseJSXWithHTMLParser(tt.htmlContent)
				if actualOutput != tt.expectedOutput {
					t.Errorf("Test %s failed:\nExpected: %s\nActual:   %s\nDescription: %s",
						tt.name, tt.expectedOutput, actualOutput, tt.description)
				}
				return
			}

			// Parse the HTML content
			doc, err := html.Parse(strings.NewReader(tt.htmlContent))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
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

			// Process the HTML
			var result strings.Builder
			if bodyNode != nil {
				for c := bodyNode.FirstChild; c != nil; c = c.NextSibling {
					recWalkHTMLNodeWithCustomComponents(c, &result, nil)
				}
			} else {
				// Fallback: process the entire document
				recWalkHTMLNodeWithCustomComponents(doc, &result, nil)
			}

			actualOutput := result.String()

			if actualOutput != tt.expectedOutput {
				t.Errorf("Test %s failed:\nExpected: %s\nActual:   %s\nDescription: %s",
					tt.name, tt.expectedOutput, actualOutput, tt.description)
			}
		})
	}
}

func TestRecWalkHTMLNodeWithCustomComponents_EdgeCases(t *testing.T) {
	t.Run("NilNode", func(t *testing.T) {
		var result strings.Builder
		recWalkHTMLNodeWithCustomComponents(nil, &result, nil)
		if result.String() != "" {
			t.Errorf("Expected empty result for nil node, got: %s", result.String())
		}
	})

	t.Run("EmptyDocument", func(t *testing.T) {
		doc, err := html.Parse(strings.NewReader(""))
		if err != nil {
			t.Fatalf("Failed to parse empty HTML: %v", err)
		}

		var result strings.Builder
		recWalkHTMLNodeWithCustomComponents(doc, &result, nil)
		// Should handle empty document gracefully
	})

	t.Run("OnlyWhitespace", func(t *testing.T) {
		doc, err := html.Parse(strings.NewReader("   \n\t   "))
		if err != nil {
			t.Fatalf("Failed to parse whitespace HTML: %v", err)
		}

		var result strings.Builder
		recWalkHTMLNodeWithCustomComponents(doc, &result, nil)
		// Should handle whitespace-only content gracefully
	})

	t.Run("MalformedHTML", func(t *testing.T) {
		doc, err := html.Parse(strings.NewReader("<div><p>Unclosed paragraph"))
		if err != nil {
			t.Fatalf("Failed to parse malformed HTML: %v", err)
		}

		var result strings.Builder
		recWalkHTMLNodeWithCustomComponents(doc, &result, nil)
		// Should handle malformed HTML gracefully
	})
}

func TestRecWalkHTMLNodeWithCustomComponents_AttributeHandling(t *testing.T) {
	tests := []struct {
		name           string
		htmlContent    string
		expectedOutput string
		description    string
	}{
		{
			name:           "ClassAttribute",
			htmlContent:    `<div class="container">Content</div>`,
			expectedOutput: `React.createElement('div', {className: "container"}, 'Content')`,
			description:    "class attribute should become className",
		},
		{
			name:           "ForAttribute",
			htmlContent:    `<label for="input">Label</label>`,
			expectedOutput: `React.createElement('label', {htmlFor: "input"}, 'Label')`,
			description:    "for attribute should become htmlFor",
		},
		{
			name:           "JSExpression",
			htmlContent:    `<button onclick="handleClick">Click me</button>`,
			expectedOutput: `React.createElement('button', {onClick: "handleClick"}, 'Click me')`,
			description:    "onclick attribute should become onClick",
		},
		{
			name:           "DataAttributes",
			htmlContent:    `<div data-test="value" data-id="123">Content</div>`,
			expectedOutput: `React.createElement('div', {"data-test": "value", "data-id": "123"}, 'Content')`,
			description:    "data attributes should be preserved with kebab-case",
		},
		{
			name:           "BooleanAttributes",
			htmlContent:    `<input type="checkbox" checked disabled />`,
			expectedOutput: `React.createElement('input', {type: "checkbox", checked: true, disabled: true})`,
			description:    "Boolean attributes should be handled correctly",
		},
		{
			name:           "SuppressHydrationWarningAttribute",
			htmlContent:    `<div suppressHydrationWarning="true">Content</div>`,
			expectedOutput: `React.createElement('div', {suppressHydrationWarning: "true"}, 'Content')`,
			description:    "suppressHydrationWarning attribute should be properly cased",
		},
		{
			name:           "MultipleReactAttributes",
			htmlContent:    `<div className="container" suppressHydrationWarning="true" onClick="handleClick">Content</div>`,
			expectedOutput: `React.createElement('div', {className: "container", suppressHydrationWarning: "true", onClick: "handleClick"}, 'Content')`,
			description:    "Multiple React-specific attributes with proper casing",
		},
		{
			name:           "DataAttributes",
			htmlContent:    `<div data-test="value" data-id="123" data-custom-attr="test">Content</div>`,
			expectedOutput: `React.createElement('div', {"data-test": "value", "data-id": "123", "data-custom-attr": "test"}, 'Content')`,
			description:    "Data attributes should preserve kebab-case",
		},
		{
			name:           "AriaAttributes",
			htmlContent:    `<button aria-label="Close" aria-expanded="false">×</button>`,
			expectedOutput: `React.createElement('button', {"aria-label": "Close", "aria-expanded": "false"}, '×')`,
			description:    "ARIA attributes should preserve kebab-case",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := html.Parse(strings.NewReader(tt.htmlContent))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			var result strings.Builder
			recWalkHTMLNodeWithCustomComponents(doc, &result, nil)

			actualOutput := result.String()

			if actualOutput != tt.expectedOutput {
				t.Errorf("Test %s failed:\nExpected: %s\nActual:   %s\nDescription: %s",
					tt.name, tt.expectedOutput, actualOutput, tt.description)
			}
		})
	}
}

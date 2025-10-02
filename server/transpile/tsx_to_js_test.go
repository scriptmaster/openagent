package transpile

import (
	"strings"
	"testing"
)

// TestTSXToJSConversion tests the complete TSX to JS conversion pipeline
func TestTSXToJSConversion(t *testing.T) {
	tests := []struct {
		name           string
		tsxContent     string
		expectedOutput string
		description    string
	}{
		{
			name:           "SimpleDiv",
			tsxContent:     `<div className="container">Hello World</div>`,
			expectedOutput: `React.createElement('div', {className: "container"}, 'Hello World')`,
			description:    "Simple div with className and text content",
		},
		{
			name:           "NestedElements",
			tsxContent:     `<div className="card"><h1>Title</h1><p>Content</p></div>`,
			expectedOutput: `React.createElement('div', {className: "card"}, React.createElement('h1', null, 'Title'), React.createElement('p', null, 'Content'))`,
			description:    "Nested HTML elements with proper structure",
		},
		{
			name:           "CustomComponent",
			tsxContent:     `<Simple suppressHydrationWarning={true} />`,
			expectedOutput: `React.createElement(Simple, {suppressHydrationWarning: true})`,
			description:    "Custom component with JSX expression",
		},
		{
			name:           "MixedHTMLAndComponents",
			tsxContent:     `<div className="container"><Simple suppressHydrationWarning={true} /><span>Text</span></div>`,
			expectedOutput: `React.createElement('div', {className: "container"}, React.createElement(Simple, {suppressHydrationWarning: true}), React.createElement('span', null, 'Text'))`,
			description:    "Mixed HTML elements and custom components",
		},
		{
			name:           "TextInterpolation",
			tsxContent:     `<div>Counter: {count}</div>`,
			expectedOutput: `React.createElement('div', null, 'Counter: ' + (count) + '')`,
			description:    "Text content with JSX interpolation",
		},
		{
			name:           "MultipleAttributes",
			tsxContent:     `<input type="text" id="username" className="form-control" />`,
			expectedOutput: `React.createElement('input', {type: "text", id: "username", className: "form-control"})`,
			description:    "Element with multiple attributes",
		},
		{
			name:           "BooleanAttributes",
			tsxContent:     `<input type="checkbox" checked disabled />`,
			expectedOutput: `React.createElement('input', {type: "checkbox", checked: true, disabled: true})`,
			description:    "Boolean attributes should be converted to true",
		},
		{
			name:           "DataAttributes",
			tsxContent:     `<div data-test="value" data-id="123">Content</div>`,
			expectedOutput: `React.createElement('div', {"data-test": "value", "data-id": "123"}, 'Content')`,
			description:    "Data attributes should preserve kebab-case",
		},
		{
			name:           "AriaAttributes",
			tsxContent:     `<button aria-label="Close" aria-expanded="false">×</button>`,
			expectedOutput: `React.createElement('button', {"aria-label": "Close", "aria-expanded": "false"}, '×')`,
			description:    "ARIA attributes should preserve kebab-case",
		},
		{
			name:           "ReactAttributes",
			tsxContent:     `<div className="container" suppressHydrationWarning={true} onClick={handleClick}>Content</div>`,
			expectedOutput: `React.createElement('div', {className: "container", suppressHydrationWarning: true, onClick: handleClick}, 'Content')`,
			description:    "React-specific attributes with proper casing",
		},
		{
			name:           "ComplexNested",
			tsxContent:     `<div className="app"><header className="header"><h1>App Title</h1></header><main className="content"><Simple suppressHydrationWarning={true} /><p>Welcome {user}!</p></main></div>`,
			expectedOutput: `React.createElement('div', {className: "app"}, React.createElement('header', {className: "header"}, React.createElement('h1', null, 'App Title')), React.createElement('main', {className: "content"}, React.createElement(Simple, {suppressHydrationWarning: true}), React.createElement('p', null, 'Welcome ' + (user) + '!')))`,
			description:    "Complex nested structure with custom components and interpolations",
		},
		{
			name:           "AdjacentComponents",
			tsxContent:     `<div><Simple suppressHydrationWarning={true} /><Counter suppressHydrationWarning={true} /></div>`,
			expectedOutput: `React.createElement('div', null, React.createElement(Simple, {suppressHydrationWarning: true}), React.createElement(Counter, {suppressHydrationWarning: true}))`,
			description:    "Adjacent custom components should have commas between them",
		},
		{
			name:           "Fragment",
			tsxContent:     `<><div>First</div><div>Second</div></>`,
			expectedOutput: `React.createElement('div', null, 'First'), React.createElement('div', null, 'Second')`,
			description:    "React Fragment should be removed and children processed directly",
		},
		{
			name:           "SelfClosingTag",
			tsxContent:     `<img src="/logo.png" alt="Logo" />`,
			expectedOutput: `React.createElement('img', {src: "/logo.png", alt: "Logo"})`,
			description:    "Self-closing HTML tags",
		},
		{
			name:           "EmptyDiv",
			tsxContent:     `<div></div>`,
			expectedOutput: `React.createElement('div', null)`,
			description:    "Empty div element",
		},
		{
			name:           "TextOnly",
			tsxContent:     `Hello World`,
			expectedOutput: `'Hello World'`,
			description:    "Plain text content",
		},
		{
			name:           "WhitespaceText",
			tsxContent:     `   \n\t   `,
			expectedOutput: ``,
			description:    "Whitespace-only text should be ignored",
		},
		{
			name:           "MultipleTextInterpolation",
			tsxContent:     `<div>Hello {name}, you have {count} items</div>`,
			expectedOutput: `React.createElement('div', null, 'Hello ' + (name) + ', you have ' + (count) + ' items')`,
			description:    "Multiple text interpolations should be properly formatted",
		},
		{
			name:           "InterpolationAtStart",
			tsxContent:     `{count} items remaining`,
			expectedOutput: `(count) + ' items remaining'`,
			description:    "JSX interpolation at the start of text",
		},
		{
			name:           "InterpolationAtEnd",
			tsxContent:     `Total: {count}`,
			expectedOutput: `'Total: ' + (count) + ''`,
			description:    "JSX interpolation at the end of text",
		},
		{
			name:           "OnlyInterpolation",
			tsxContent:     `{count}`,
			expectedOutput: `(count) + ''`,
			description:    "Text content that is only a JSX interpolation",
		},
		{
			name:           "JSXExpressions",
			tsxContent:     `<Simple suppressHydrationWarning={true} onClick={handleClick} className={dynamicClass} />`,
			expectedOutput: `React.createElement(Simple, {suppressHydrationWarning: true, onClick: handleClick, className: dynamicClass})`,
			description:    "JSX expressions should be converted to JavaScript values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use the full TSX to JS conversion pipeline
			result := parseJSXWithHTMLParser(tt.tsxContent)

			// Clean up the result for comparison (remove extra whitespace)
			result = strings.TrimSpace(result)
			expectedOutput := strings.TrimSpace(tt.expectedOutput)

			if result != expectedOutput {
				t.Errorf("Test %s failed:\nExpected: %s\nActual:   %s\nDescription: %s",
					tt.name, expectedOutput, result, tt.description)
			}
		})
	}
}

// TestTSXToJSConversionIntegration tests the complete integration from TSX to JS
func TestTSXToJSConversionIntegration(t *testing.T) {
	tests := []struct {
		name                string
		tsxContent          string
		expectedContains    []string
		expectedNotContains []string
		description         string
	}{
		{
			name:       "CompletePageStructure",
			tsxContent: `<div className="container-xl"><div className="card"><div className="card-status-start bg-green"></div><div className="card-body"><Simple suppressHydrationWarning={true} /><span>Element beside a component. &nbsp;</span></div></div><div className="row mt-4"><div className="col-12"><Counter suppressHydrationWarning={true} /></div></div></div>`,
			expectedContains: []string{
				"React.createElement('div', {className: \"container-xl\"}",
				"React.createElement('div', {className: \"card\"}",
				"React.createElement('div', {className: \"card-status-start bg-green\"}",
				"React.createElement('div', {className: \"card-body\"}",
				"React.createElement(Simple, {suppressHydrationWarning: true}",
				"React.createElement('span', null, 'Element beside a component. &nbsp;')",
				"React.createElement('div', {className: \"row mt-4\"}",
				"React.createElement('div', {className: \"col-12\"}",
				"React.createElement(Counter, {suppressHydrationWarning: true}",
			},
			expectedNotContains: []string{
				"<>",
				"</>",
				"{true}",
				"{false}",
			},
			description: "Complete page structure should include all HTML elements and custom components",
		},
		{
			name:       "CustomComponentWithScript",
			tsxContent: `<Counter suppressHydrationWarning={true} />`,
			expectedContains: []string{
				"React.createElement(Counter, {suppressHydrationWarning: true}",
			},
			expectedNotContains: []string{
				"<Counter",
				"</Counter>",
				"{true}",
			},
			description: "Custom component should be converted to React.createElement call",
		},
		{
			name:       "TextInterpolationInComponent",
			tsxContent: `<div>Counter: {count}</div>`,
			expectedContains: []string{
				"React.createElement('div', null, 'Counter: ' + (count) + '')",
			},
			expectedNotContains: []string{
				"{count}",
				"Counter: count",
			},
			description: "Text interpolation should be converted to string concatenation",
		},
		{
			name:       "MultipleCustomComponents",
			tsxContent: `<div><Simple suppressHydrationWarning={true} /><Counter suppressHydrationWarning={true} /></div>`,
			expectedContains: []string{
				"React.createElement('div', null, React.createElement(Simple, {suppressHydrationWarning: true}), React.createElement(Counter, {suppressHydrationWarning: true}))",
			},
			expectedNotContains: []string{
				"React.createElement(Simple, {suppressHydrationWarning: true})React.createElement(Counter, {suppressHydrationWarning: true})",
			},
			description: "Multiple custom components should be properly separated with commas",
		},
		{
			name:       "ComplexNestedStructure",
			tsxContent: `<div className="app"><header className="header"><h1>App Title</h1></header><main className="content"><Simple suppressHydrationWarning={true} /><p>Welcome {user}!</p></main></div>`,
			expectedContains: []string{
				"React.createElement('div', {className: \"app\"}",
				"React.createElement('header', {className: \"header\"}",
				"React.createElement('h1', null, 'App Title')",
				"React.createElement('main', {className: \"content\"}",
				"React.createElement(Simple, {suppressHydrationWarning: true}",
				"React.createElement('p', null, 'Welcome ' + (user) + '!')",
			},
			expectedNotContains: []string{
				"<div",
				"</div>",
				"<header",
				"</header>",
				"<h1",
				"</h1>",
				"<main",
				"</main>",
				"<p",
				"</p>",
				"{user}",
			},
			description: "Complex nested structure should be fully converted to React.createElement calls",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use the full TSX to JS conversion pipeline
			result := parseJSXWithHTMLParser(tt.tsxContent)

			// Check that all expected content is present
			for _, expected := range tt.expectedContains {
				if !strings.Contains(result, expected) {
					t.Errorf("Test %s failed: Expected to contain '%s' but got: %s",
						tt.name, expected, result)
				}
			}

			// Check that unwanted content is not present
			for _, notExpected := range tt.expectedNotContains {
				if strings.Contains(result, notExpected) {
					t.Errorf("Test %s failed: Expected NOT to contain '%s' but got: %s",
						tt.name, notExpected, result)
				}
			}
		})
	}
}

// TestTSXToJSConversionEdgeCases tests edge cases and error conditions
func TestTSXToJSConversionEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		tsxContent     string
		expectedOutput string
		description    string
	}{
		{
			name:           "EmptyString",
			tsxContent:     ``,
			expectedOutput: ``,
			description:    "Empty string should return empty string",
		},
		{
			name:           "OnlyWhitespace",
			tsxContent:     `   \n\t   `,
			expectedOutput: ``,
			description:    "Whitespace-only content should be ignored",
		},
		{
			name:           "MalformedHTML",
			tsxContent:     `<div><p>Unclosed paragraph</div>`,
			expectedOutput: `React.createElement('div', null, React.createElement('p', null, 'Unclosed paragraph'))`,
			description:    "Malformed HTML should still be processed",
		},
		{
			name:           "NestedFragments",
			tsxContent:     `<><div>First</div><><span>Second</span></></>`,
			expectedOutput: `React.createElement('div', null, 'First'), React.createElement('span', null, 'Second')`,
			description:    "Nested fragments should be flattened",
		},
		{
			name:           "DeeplyNested",
			tsxContent:     `<div><div><div><div><div>Deep</div></div></div></div></div>`,
			expectedOutput: `React.createElement('div', null, React.createElement('div', null, React.createElement('div', null, React.createElement('div', null, React.createElement('div', null, 'Deep'))))`,
			description:    "Deeply nested elements should be properly converted",
		},
		{
			name:           "SpecialCharacters",
			tsxContent:     `<div>Special chars: &lt; &gt; &amp; &quot; &#39;</div>`,
			expectedOutput: `React.createElement('div', null, 'Special chars: < > & " \'')`,
			description:    "Special HTML characters should be properly escaped",
		},
		{
			name:           "UnicodeCharacters",
			tsxContent:     `<div>Unicode: 🚀 émojis 中文</div>`,
			expectedOutput: `React.createElement('div', null, 'Unicode: 🚀 émojis 中文')`,
			description:    "Unicode characters should be preserved",
		},
		{
			name:           "LongAttributeValues",
			tsxContent:     `<div className="very-long-class-name-with-many-words-and-hyphens">Content</div>`,
			expectedOutput: `React.createElement('div', {className: "very-long-class-name-with-many-words-and-hyphens"}, 'Content')`,
			description:    "Long attribute values should be preserved",
		},
		{
			name:           "MultipleBooleanAttributes",
			tsxContent:     `<input type="checkbox" checked disabled readonly />`,
			expectedOutput: `React.createElement('input', {type: "checkbox", checked: true, disabled: true, readonly: true})`,
			description:    "Multiple boolean attributes should all be converted to true",
		},
		{
			name:           "EmptyAttributes",
			tsxContent:     `<div className="" id="">Empty</div>`,
			expectedOutput: `React.createElement('div', {className: "", id: ""}, 'Empty')`,
			description:    "Empty attribute values should be preserved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use the full TSX to JS conversion pipeline
			result := parseJSXWithHTMLParser(tt.tsxContent)

			// Clean up the result for comparison
			result = strings.TrimSpace(result)
			expectedOutput := strings.TrimSpace(tt.expectedOutput)

			if result != expectedOutput {
				t.Errorf("Test %s failed:\nExpected: %s\nActual:   %s\nDescription: %s",
					tt.name, expectedOutput, result, tt.description)
			}
		})
	}
}

// TestTSXToJSConversionPerformance tests performance with large structures
func TestTSXToJSConversionPerformance(t *testing.T) {
	// Create a large nested structure
	largeTSX := `<div className="container">`
	for i := 0; i < 100; i++ {
		largeTSX += `<div className="item" data-id="` + string(rune(i)) + `">Item ` + string(rune(i)) + `</div>`
	}
	largeTSX += `</div>`

	// Test that it can handle large structures
	result := parseJSXWithHTMLParser(largeTSX)

	// Should contain the container div
	if !strings.Contains(result, "React.createElement('div', {className: \"container\"}") {
		t.Errorf("Large structure test failed: Expected to contain container div")
	}

	// Should contain many item divs
	itemCount := strings.Count(result, "React.createElement('div', {className: \"item\"}")
	if itemCount < 50 { // At least half should be processed
		t.Errorf("Large structure test failed: Expected at least 50 items, got %d", itemCount)
	}
}

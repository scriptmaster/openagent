package transpile

import (
	"strings"
	"testing"
)

func TestSimple(t *testing.T) {
	t.Log("Simple test is running!")
}

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
			name:           "TextInterpolation",
			tsxContent:     `<div>Counter: {count}</div>`,
			expectedOutput: `React.createElement('div', null, 'Counter: ' + (count) + '')`,
			description:    "Text content with JSX interpolation",
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
			name:           "ElementBesideComponent",
			tsxContent:     `<div className="container-xl"><div className="card-body"><Simple suppressHydrationWarning={true} /><span>Element beside a component.</span></div></div>`,
			expectedOutput: `React.createElement('div', {className: "container-xl"}, React.createElement('div', {className: "card-body"}, React.createElement(Simple, {suppressHydrationWarning: true}), React.createElement('span', null, 'Element beside a component.')))`,
			description:    "HTML element should be sibling to custom component, not child",
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

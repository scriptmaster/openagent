package transpile

import (
	"strings"
	"testing"
)

func TestTSX2JS_SiblingElements(t *testing.T) {
	tests := []struct {
		name           string
		inputTSX       string
		expectedOutput string
		description    string
	}{
		{
			name: "Component with sibling element",
			inputTSX: `export default function Test({page}: {page: any}) {
    return (
<div className="container-xl">
    <div className="card-body">
        <Simple suppressHydrationWarning={true} />
        <span>Element beside a component.</span>
    </div>
</div>
    );
}`,
			expectedOutput: `React.createElement('div', {className: "container-xl"}, React.createElement('div', {className: "card-body"}, React.createElement(Simple, {suppressHydrationWarning: true}), React.createElement('span', null, 'Element beside a component.')))`,
			description:    "Simple component and span should be siblings, not parent-child",
		},
		{
			name: "Multiple siblings",
			inputTSX: `export default function Test({page}: {page: any}) {
    return (
<div className="container">
    <Simple suppressHydrationWarning={true} />
    <p>First sibling</p>
    <Counter suppressHydrationWarning={true} />
    <span>Second sibling</span>
</div>
    );
}`,
			expectedOutput: `React.createElement('div', {className: "container"}, React.createElement(Simple, {suppressHydrationWarning: true}), React.createElement('p', null, 'First sibling'), React.createElement(Counter, {suppressHydrationWarning: true}), React.createElement('span', null, 'Second sibling'))`,
			description:    "Multiple components and elements should all be siblings",
		},
		{
			name: "Nested siblings",
			inputTSX: `export default function Test({page}: {page: any}) {
    return (
<div className="wrapper">
    <div className="inner">
        <Simple suppressHydrationWarning={true} />
        <div className="sibling">
            <span>Nested sibling content</span>
        </div>
    </div>
</div>
    );
}`,
			expectedOutput: `React.createElement('div', {className: "wrapper"}, React.createElement('div', {className: "inner"}, React.createElement(Simple, {suppressHydrationWarning: true}), React.createElement('div', {className: "sibling"}, React.createElement('span', null, 'Nested sibling content'))))`,
			description:    "Nested structure should preserve sibling relationships",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TSX2JS(tt.inputTSX)

			// Clean up whitespace for comparison
			result = strings.ReplaceAll(result, "\n", "")
			result = strings.ReplaceAll(result, " ", "")
			expected := strings.ReplaceAll(tt.expectedOutput, "\n", "")
			expected = strings.ReplaceAll(expected, " ", "")

			if result != expected {
				t.Errorf("TSX2JS() output mismatch")
				t.Errorf("Expected: %s", tt.expectedOutput)
				t.Errorf("Got:      %s", result)

				// Additional check: ensure siblings are not nested
				if strings.Contains(tt.description, "siblings") {
					// Check that Simple component is not wrapping the span
					if strings.Contains(result, "React.createElement(Simple") && strings.Contains(result, "Element beside a component") {
						// Find the position of Simple and span in the result
						simplePos := strings.Index(result, "React.createElement(Simple")
						spanPos := strings.Index(result, "Element beside a component")

						if simplePos != -1 && spanPos != -1 {
							// Check if span is inside Simple's children
							simpleEnd := strings.Index(result[simplePos:], "))")
							if simpleEnd != -1 {
								simpleEndPos := simplePos + simpleEnd
								if spanPos < simpleEndPos {
									t.Errorf("Sibling element is incorrectly nested inside Simple component")
									t.Errorf("Simple component ends at position %d, but span is at position %d", simpleEndPos, spanPos)
								}
							}
						}
					}
				}
			}
		})
	}
}

func TestTSX2JS_SiblingElements_Contains(t *testing.T) {
	inputTSX := `export default function Test({page}: {page: any}) {
    return (
<div className="container-xl">
    <div className="card-body">
        <Simple suppressHydrationWarning={true} />
        <span>Element beside a component.</span>
    </div>
</div>
    );
}`

	result := TSX2JS(inputTSX)

	// Test that the result contains the correct structure
	contains := []string{
		"React.createElement('div', {className: \"container-xl\"}",
		"React.createElement('div', {className: \"card-body\"}",
		"React.createElement(Simple, {suppressHydrationWarning: true}",
		"React.createElement('span', null, 'Element beside a component.')",
	}

	notContains := []string{
		"React.createElement(Simple, {suppressHydrationWarning: true}, React.createElement('span'",
	}

	for _, expected := range contains {
		if !strings.Contains(result, expected) {
			t.Errorf("TSX2JS() result should contain: %s", expected)
		}
	}

	for _, unwanted := range notContains {
		if strings.Contains(result, unwanted) {
			t.Errorf("TSX2JS() result should NOT contain: %s", unwanted)
		}
	}
}

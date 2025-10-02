package transpile

import (
	"strings"
	"testing"
)

func TestProcessComponentImports(t *testing.T) {
	tests := []struct {
		name               string
		inputHTML          string
		expectedOutput     string
		expectedComponents []string
		description        string
	}{
		{
			name: "Simple component with sibling element",
			inputHTML: `<div class="container-xl">
    <div class="card-body">
        <template id="component-simple"></template>
        <span>Element beside a component.</span>
    </div>
</div>`,
			expectedOutput: `<div class="container-xl">
    <div class="card-body">
        <Simple suppressHydrationWarning={true} />
        <span>Element beside a component.</span>
    </div>
</div>`,
			expectedComponents: []string{"Simple"},
			description:        "Component template should be replaced with Simple component, sibling span should remain as sibling",
		},
		{
			name: "Multiple components with siblings",
			inputHTML: `<div class="container">
    <template id="component-simple"></template>
    <p>Text after Simple</p>
    <template id="component-counter"></template>
    <span>Text after Counter</span>
</div>`,
			expectedOutput: `<div class="container">
    <Simple suppressHydrationWarning={true} />
    <p>Text after Simple</p>
    <Counter suppressHydrationWarning={true} />
    <span>Text after Counter</span>
</div>`,
			expectedComponents: []string{"Simple", "Counter"},
			description:        "Multiple components should be replaced, all siblings should remain as siblings",
		},
		{
			name: "Component with nested siblings",
			inputHTML: `<div class="wrapper">
    <div class="inner">
        <template id="component-simple"></template>
        <div class="sibling">
            <span>Nested sibling content</span>
        </div>
    </div>
</div>`,
			expectedOutput: `<div class="wrapper">
    <div class="inner">
        <Simple suppressHydrationWarning={true} />
        <div class="sibling">
            <span>Nested sibling content</span>
        </div>
    </div>
</div>`,
			expectedComponents: []string{"Simple"},
			description:        "Component should be replaced, nested sibling div should remain as sibling",
		},
		{
			name: "Component with underscore naming",
			inputHTML: `<div class="container">
    <template id="component_my_component"></template>
    <p>Sibling paragraph</p>
</div>`,
			expectedOutput: `<div class="container">
    <MyComponent suppressHydrationWarning={true} />
    <p>Sibling paragraph</p>
</div>`,
			expectedComponents: []string{"MyComponent"},
			description:        "Component with underscore should be converted to camelCase and replaced",
		},
		{
			name: "No components",
			inputHTML: `<div class="container">
    <p>Just a paragraph</p>
    <span>Just a span</span>
</div>`,
			expectedOutput: `<div class="container">
    <p>Just a paragraph</p>
    <span>Just a span</span>
</div>`,
			expectedComponents: []string{},
			description:        "HTML without component divs should remain unchanged",
		},
		{
			name: "Component with attributes",
			inputHTML: `<div class="container">
    <template id="component-simple" class="my-class" data-test="value"></template>
    <span>Sibling element</span>
</div>`,
			expectedOutput: `<div class="container">
    <Simple suppressHydrationWarning={true} />
    <span>Sibling element</span>
</div>`,
			expectedComponents: []string{"Simple"},
			description:        "Component div with additional attributes should be replaced, attributes should be ignored",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary input file path for testing
			inputPath := "test.html"

			// Process the component imports
			result, importedComponents, err := processComponentImports(tt.inputHTML, inputPath)

			// Check for errors
			if err != nil {
				t.Errorf("processComponentImports() error = %v", err)
				return
			}

			// Check the output HTML
			if strings.TrimSpace(result) != strings.TrimSpace(tt.expectedOutput) {
				t.Errorf("processComponentImports() output mismatch")
				t.Errorf("Expected: %s", tt.expectedOutput)
				t.Errorf("Got:      %s", result)
			}

			// Check the imported components
			if len(importedComponents) != len(tt.expectedComponents) {
				t.Errorf("processComponentImports() imported components count mismatch")
				t.Errorf("Expected: %v", tt.expectedComponents)
				t.Errorf("Got:      %v", importedComponents)
				return
			}

			for i, expected := range tt.expectedComponents {
				if i >= len(importedComponents) || importedComponents[i] != expected {
					t.Errorf("processComponentImports() imported components mismatch")
					t.Errorf("Expected: %v", tt.expectedComponents)
					t.Errorf("Got:      %v", importedComponents)
					break
				}
			}

			// Additional check: ensure siblings are preserved
			if strings.Contains(tt.inputHTML, "Element beside a component") {
				if !strings.Contains(result, "Element beside a component") {
					t.Errorf("Sibling element 'Element beside a component' was lost during processing")
				}
			}

			if strings.Contains(tt.inputHTML, "Text after Simple") {
				if !strings.Contains(result, "Text after Simple") {
					t.Errorf("Sibling element 'Text after Simple' was lost during processing")
				}
			}
		})
	}
}

func TestProcessComponentImports_EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		inputHTML      string
		expectedOutput string
		description    string
	}{
		{
			name: "Empty component div",
			inputHTML: `<div class="container">
    <template id="component-simple"></template>
</div>`,
			expectedOutput: `<div class="container">
    <Simple suppressHydrationWarning={true} />
</div>`,
			description: "Empty component div should be replaced with self-closing component",
		},
		{
			name: "Component div with whitespace",
			inputHTML: `<div class="container">
    <template id="component-simple">   </template>
    <span>Sibling</span>
</div>`,
			expectedOutput: `<div class="container">
    <Simple suppressHydrationWarning={true} />
    <span>Sibling</span>
</div>`,
			description: "Component div with whitespace should be replaced, sibling should remain",
		},
		{
			name: "Multiple same components",
			inputHTML: `<div class="container">
    <template id="component-simple"></template>
    <template id="component-simple"></template>
    <span>Between components</span>
</div>`,
			expectedOutput: `<div class="container">
    <Simple suppressHydrationWarning={true} />
    <Simple suppressHydrationWarning={true} />
    <span>Between components</span>
</div>`,
			description: "Multiple same components should all be replaced, text between should remain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputPath := "test.html"

			result, _, err := processComponentImports(tt.inputHTML, inputPath)

			if err != nil {
				t.Errorf("processComponentImports() error = %v", err)
				return
			}

			if strings.TrimSpace(result) != strings.TrimSpace(tt.expectedOutput) {
				t.Errorf("processComponentImports() output mismatch")
				t.Errorf("Expected: %s", tt.expectedOutput)
				t.Errorf("Got:      %s", result)
			}
		})
	}
}

package transpile

import (
	"strings"
	"testing"
)

func TestNoDuplicateFunctionDeclarations(t *testing.T) {
	tests := []struct {
		name        string
		tsxContent  string
		description string
	}{
		{
			name: "SingleFunctionPattern",
			tsxContent: `export default function Test({page}: {page: any}) {
    return (
        <div>Hello World</div>
    );
}`,
			description: "Single function pattern should not have duplicate function declarations",
		},
		{
			name: "DualFunctionPattern",
			tsxContent: `export default function Test({page}: {page: any}) {
    // ╔══ 🔧 PAGE <script> TAG CONTENTS 🔧 ══
    let count = 0;
    const state = { count: 55 };
    
    // Call the JSX function with props and state
    return TestJSX(typeof props != 'undefined' ? props : {}, typeof state != 'undefined' ? state : {});
}

function TestJSX(props, state) {
    return (
        <div>Counter: {state.count}</div>
    );
}`,
			description: "Dual function pattern should not have duplicate function declarations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert TSX to JS
			jsContent := TSX2JSWithOptions(tt.tsxContent, true)
			
			// Check for duplicate function declarations
			functionCount := strings.Count(jsContent, "function Test({page})")
			if functionCount > 1 {
				t.Errorf("Found %d duplicate function declarations in %s", functionCount, tt.description)
				t.Logf("Generated JS content:\n%s", jsContent)
			}
			
			// Check for malformed syntax
			if strings.Contains(jsContent, "function Test({page}) {\n    return (\nfunction Test({page}) {") {
				t.Errorf("Found malformed nested function syntax in %s", tt.description)
				t.Logf("Generated JS content:\n%s", jsContent)
			}
			
			// Check for proper function structure
			if !strings.Contains(jsContent, "function Test({page}) {") {
				t.Errorf("Missing main function declaration in %s", tt.description)
				t.Logf("Generated JS content:\n%s", jsContent)
			}
			
			// For dual function pattern, check for TestJSX function
			if strings.Contains(tt.tsxContent, "TestJSX(") {
				if !strings.Contains(jsContent, "function TestJSX(") {
					t.Errorf("Missing TestJSX function in dual function pattern")
					t.Logf("Generated JS content:\n%s", jsContent)
				}
			}
			
			t.Logf("✅ %s: No duplicate function declarations found", tt.description)
		})
	}
}

func TestTSXToJSConversionStructure(t *testing.T) {
	// Test the specific issue with duplicate function declarations
	tsxContent := `export default function Test({page}: {page: any}) {
    return (
        <div>Hello World</div>
    );
}`

	jsContent := TSX2JSWithOptions(tsxContent, true)
	
	// Check that we have exactly one function declaration
	functionCount := strings.Count(jsContent, "function Test({page})")
	if functionCount != 1 {
		t.Errorf("Expected exactly 1 function declaration, got %d", functionCount)
		t.Logf("Generated JS content:\n%s", jsContent)
	}
	
	// Check that the function structure is correct
	if !strings.Contains(jsContent, "React.createElement('div', null, 'Hello World')") {
		t.Errorf("Generated JS does not contain expected React.createElement structure")
		t.Logf("Generated JS content:\n%s", jsContent)
	}
	
	t.Logf("✅ TSX to JS conversion structure is correct")
}

func TestDualFunctionPatternStructureDuplicate(t *testing.T) {
	// Test dual function pattern specifically
	tsxContent := `export default function Test({page}: {page: any}) {
    // ╔══ 🔧 PAGE <script> TAG CONTENTS 🔧 ══
    let count = 0;
    const state = { count: 55 };
    
    // Call the JSX function with props and state
    return TestJSX(typeof props != 'undefined' ? props : {}, typeof state != 'undefined' ? state : {});
}

function TestJSX(props, state) {
    return (
        <div>Counter: {state.count}</div>
    );
}`

	jsContent := TSX2JSWithOptions(tsxContent, true)
	
	// Check for main function
	if !strings.Contains(jsContent, "function Test({page})") {
		t.Errorf("Missing main Test function")
		t.Logf("Generated JS content:\n%s", jsContent)
	}
	
	// Check for JSX function
	if !strings.Contains(jsContent, "function TestJSX(") {
		t.Errorf("Missing TestJSX function")
		t.Logf("Generated JS content:\n%s", jsContent)
	}
	
	// Check for proper return statement
	if !strings.Contains(jsContent, "return TestJSX(") {
		t.Errorf("Missing TestJSX call in main function")
		t.Logf("Generated JS content:\n%s", jsContent)
	}
	
	// Check for no duplicate main functions
	mainFunctionCount := strings.Count(jsContent, "function Test({page})")
	if mainFunctionCount != 1 {
		t.Errorf("Expected exactly 1 main function, got %d", mainFunctionCount)
		t.Logf("Generated JS content:\n%s", jsContent)
	}
	
	t.Logf("✅ Dual function pattern structure is correct")
}

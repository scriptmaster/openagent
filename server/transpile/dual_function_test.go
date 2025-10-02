package transpile

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestDualFunctionPatternInGeneratedFiles(t *testing.T) {
	tests := []struct {
		name           string
		envValue       string
		expectDual     bool
		description    string
	}{
		{
			name:        "WAX_FORK_Unset_DualFunction",
			envValue:    "",
			expectDual:  true,
			description: "Should generate dual function pattern when WAX_FORK is unset",
		},
		{
			name:        "WAX_FORK_0_DualFunction",
			envValue:    "0",
			expectDual:  true,
			description: "Should generate dual function pattern when WAX_FORK=0",
		},
		{
			name:        "WAX_FORK_1_SingleFunction",
			envValue:    "1",
			expectDual:  false,
			description: "Should generate single function pattern when WAX_FORK=1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envValue != "" {
				os.Setenv("WAX_FORK", tt.envValue)
			} else {
				os.Unsetenv("WAX_FORK")
			}
			defer os.Unsetenv("WAX_FORK")

			// Test the environment variable logic
			useDualPattern := !getWaxForkSettingForTranspile()
			if useDualPattern != tt.expectDual {
				t.Errorf("Environment variable logic failed. Expected dual: %v, got: %v", tt.expectDual, useDualPattern)
			}

			// Test the dual function component creation
			componentName := "TestComponent"
			jsContent := "let count = 0; const inc = () => count++;"
			componentHTML := "<div>Counter: {count} <button onClick={inc}>+</button></div>"

			var result string
			if useDualPattern {
				result = createDualFunctionComponent(componentName, jsContent, componentHTML)
			} else {
				result = createSingleFunctionComponent(componentName, jsContent, componentHTML)
			}

			// Check for dual function pattern elements
			dualPatternElements := []string{
				"export default function TestComponent()",
				"// ╔══ 🔧 COMPONENT <script> TAG CONTENTS 🔧 ══",
				"let count = 0; const inc = () => count++;",
				"return TestComponentJSX(",
				"function TestComponentJSX(props, state)",
				"<div>Counter: {count} <button onClick={inc}>+</button></div>",
			}

			// Check for single function pattern elements
			singlePatternElements := []string{
				"export default function TestComponent()",
				"// ╔══ 🔧 COMPONENT <script> TAG CONTENTS 🔧 ══",
				"let count = 0; const inc = () => count++;",
				"return (",
				"<div>Counter: {count} <button onClick={inc}>+</button></div>",
			}

			if tt.expectDual {
				// Should contain dual function pattern elements
				for _, element := range dualPatternElements {
					if !strings.Contains(result, element) {
						t.Errorf("Dual function pattern missing element: %s", element)
					}
				}
				// Should NOT contain single function pattern elements
				if strings.Contains(result, "return (") && !strings.Contains(result, "TestComponentJSX(") {
					t.Errorf("Should not have single function pattern when dual is expected")
				}
			} else {
				// Should contain single function pattern elements
				for _, element := range singlePatternElements {
					if !strings.Contains(result, element) {
						t.Errorf("Single function pattern missing element: %s", element)
					}
				}
				// Should NOT contain dual function pattern elements
				if strings.Contains(result, "TestComponentJSX(") {
					t.Errorf("Should not have dual function pattern when single is expected")
				}
			}

			t.Logf("✅ %s: Pattern generation working correctly", tt.description)
			t.Logf("Generated component structure:\n%s", result)
		})
	}
}

func TestDualFunctionPatternStructure(t *testing.T) {
	// Test the structure of the dual function pattern
	componentName := "Counter"
	jsContent := `
    let count = 55;
    const state = {
        count: 55,
        inc: () => count++,
        dec: () => count--
    };`
	componentHTML := "<div><button onClick={state.dec}>-</button>Counter: {state.count}<button onClick={state.inc}>+</button></div>"

	// Create dual function component
	result := createDualFunctionComponent(componentName, jsContent, componentHTML)

	// Check for required dual function elements
	requiredElements := []string{
		"export default function Counter()",
		"// ╔══ 🔧 COMPONENT <script> TAG CONTENTS 🔧 ══",
		"let count = 55;",
		"const state = {",
		"count: 55,",
		"inc: () => count++,",
		"dec: () => count--",
		"return CounterJSX(",
		"typeof props != 'undefined' ? props : {}",
		"typeof state != 'undefined' ? state : {}",
		"function CounterJSX(props, state)",
		"<div><button onClick={state.dec}>-</button>Counter: {state.count}<button onClick={state.inc}>+</button></div>",
	}

	for _, element := range requiredElements {
		if !strings.Contains(result, element) {
			t.Errorf("Dual function pattern missing required element: %s", element)
		}
	}

	// Verify the structure is correct
	if !strings.Contains(result, "export default function Counter()") {
		t.Error("Missing main function declaration")
	}
	if !strings.Contains(result, "function CounterJSX(props, state)") {
		t.Error("Missing JSX function declaration")
	}
	if !strings.Contains(result, "return CounterJSX(") {
		t.Error("Missing JSX function call")
	}

	t.Log("✅ Dual function pattern structure is correct")
	t.Logf("Generated component:\n%s", result)
}

func TestSingleFunctionPatternStructure(t *testing.T) {
	// Test the structure of the single function pattern
	componentName := "Counter"
	jsContent := `
    let count = 55;
    const state = {
        count: 55,
        inc: () => count++,
        dec: () => count--
    };`
	componentHTML := "<div><button onClick={state.dec}>-</button>Counter: {state.count}<button onClick={state.inc}>+</button></div>"

	// Create single function component
	result := createSingleFunctionComponent(componentName, jsContent, componentHTML)

	// Check for required single function elements
	requiredElements := []string{
		"export default function Counter()",
		"// ╔══ 🔧 COMPONENT <script> TAG CONTENTS 🔧 ══",
		"let count = 55;",
		"const state = {",
		"count: 55,",
		"inc: () => count++,",
		"dec: () => count--",
		"return (",
		"<div><button onClick={state.dec}>-</button>Counter: {state.count}<button onClick={state.inc}>+</button></div>",
	}

	for _, element := range requiredElements {
		if !strings.Contains(result, element) {
			t.Errorf("Single function pattern missing required element: %s", element)
		}
	}

	// Should NOT contain dual function elements
	if strings.Contains(result, "CounterJSX(") {
		t.Error("Should not contain JSX function call in single function pattern")
	}
	if strings.Contains(result, "function CounterJSX(") {
		t.Error("Should not contain JSX function declaration in single function pattern")
	}

	t.Log("✅ Single function pattern structure is correct")
	t.Logf("Generated component:\n%s", result)
}

// createSingleFunctionComponent creates a component with single function pattern for testing
func createSingleFunctionComponent(componentName, jsContent, componentHTML string) string {
	componentNameCamel := convertToCamelCase(componentName)
	
	return fmt.Sprintf(`export default function %s() {
    // ╔══ 🔧 COMPONENT <script> TAG CONTENTS 🔧 ══
%s
    return (
        %s
    );
}`, componentNameCamel, jsContent, componentHTML)
}


package transpile

import (
	"strings"
	"testing"
)

func TestTSX2JS(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		contains    []string
		notContains []string
	}{
		{
			name: "Simple HTML elements",
			input: `<main>
				<div className="container">
					<h1>Hello World</h1>
					<p>This is a test paragraph.</p>
				</div>
			</main>`,
			expected: `React.createElement('div', {className: "container"}, React.createElement('h1', null, 'Hello World'), React.createElement('p', null, 'This is a test paragraph.'))`,
		},
		{
			name: "Nested elements with attributes",
			input: `<main>
				<div className="card" id="main-card">
					<header className="card-header">
						<h2 className="card-title">Card Title</h2>
					</header>
					<div className="card-body">
						<p className="text-muted">Card content goes here.</p>
						<button className="btn btn-primary" type="button">Click Me</button>
					</div>
				</div>
			</main>`,
			expected: `React.createElement('div', {className: "card", id: "main-card"}, React.createElement('header', {className: "card-header"}, React.createElement('h2', {className: "card-title"}, 'Card Title')), React.createElement('div', {className: "card-body"}, React.createElement('p', {className: "text-muted"}, 'Card content goes here.'), React.createElement('button', {className: "btn btn-primary", type: "button"}, 'Click Me')))`,
		},
		{
			name: "Custom React component",
			input: `<main>
				<div className="container">
					<Simple suppressHydrationWarning={true} />
					<Counter initialValue={0} />
				</div>
			</main>`,
			expected: `React.createElement('div', {className: "container"}, React.createElement(Simple, {suppressHydrationWarning: true}, React.createElement(Counter, {initialValue: 0})))`,
		},
		{
			name: "Self-closing tags",
			input: `<main>
				<div>
					<img src="/logo.png" alt="Logo" />
					<br />
					<hr className="divider" />
				</div>
			</main>`,
			expected: `React.createElement('div', null, React.createElement('img', {src: "/logo.png", alt: "Logo"}), React.createElement('br', null), React.createElement('hr', {className: "divider"}))`,
		},
		{
			name: "Mixed content with text and elements",
			input: `<main>
				<div>
					Welcome to 
					<strong>our website</strong>
					!
					<br />
					Please 
					<a href="/login">login</a>
					to continue.
				</div>
			</main>`,
			expected: `React.createElement('div', null, 'Welcome to', React.createElement('strong', null, 'our website'), '!', React.createElement('br', null), 'Please', React.createElement('a', {href: "/login"}, 'login'), 'to continue.')`,
		},
		{
			name: "Form elements",
			input: `<main>
				<form className="login-form">
					<div className="form-group">
						<label htmlFor="email">Email:</label>
						<input type="email" id="email" name="email" className="form-control" />
					</div>
					<div className="form-group">
						<label htmlFor="password">Password:</label>
						<input type="password" id="password" name="password" className="form-control" />
					</div>
					<button type="submit" className="btn btn-primary">Login</button>
				</form>
			</main>`,
			expected: `React.createElement('form', {className: "login-form"}, React.createElement('div', {className: "form-group"}, React.createElement('label', {htmlFor: "email"}, 'Email:'), React.createElement('input', {type: "email", id: "email", name: "email", className: "form-control"})), React.createElement('div', {className: "form-group"}, React.createElement('label', {htmlFor: "password"}, 'Password:'), React.createElement('input', {type: "password", id: "password", name: "password", className: "form-control"})), React.createElement('button', {type: "submit", className: "btn btn-primary"}, 'Login'))`,
		},
		{
			name: "List elements",
			input: `<main>
				<ul className="nav-list">
					<li className="nav-item">
						<a href="/home" className="nav-link">Home</a>
					</li>
					<li className="nav-item">
						<a href="/about" className="nav-link">About</a>
					</li>
					<li className="nav-item">
						<a href="/contact" className="nav-link">Contact</a>
					</li>
				</ul>
			</main>`,
			expected: `React.createElement('ul', {className: "nav-list"}, React.createElement('li', {className: "nav-item"}, React.createElement('a', {href: "/home", className: "nav-link"}, 'Home')), React.createElement('li', {className: "nav-item"}, React.createElement('a', {href: "/about", className: "nav-link"}, 'About')), React.createElement('li', {className: "nav-item"}, React.createElement('a', {href: "/contact", className: "nav-link"}, 'Contact')))`,
		},
		{
			name:     "Empty main tag",
			input:    `<main></main>`,
			expected: ``,
		},
		{
			name: "No main tag - direct content",
			input: `<div className="container">
				<h1>Direct Content</h1>
			</div>`,
			expected: `React.createElement('div', {className: "container"}, React.createElement('h1', null, 'Direct Content'))`,
		},
		{
			name: "Complex nested structure",
			input: `<main>
				<div className="app">
					<header className="app-header">
						<nav className="navbar">
							<div className="navbar-brand">
								<img src="/logo.svg" alt="Brand" className="logo" />
							</div>
							<ul className="navbar-nav">
								<li className="nav-item">
									<a href="/dashboard" className="nav-link active">Dashboard</a>
								</li>
								<li className="nav-item">
									<a href="/settings" className="nav-link">Settings</a>
								</li>
							</ul>
						</nav>
					</header>
					<main className="app-main">
						<div className="container-fluid">
							<div className="row">
								<div className="col-md-8">
									<div className="card">
										<div className="card-header">
											<h3 className="card-title">Main Content</h3>
										</div>
										<div className="card-body">
											<p>This is the main content area.</p>
											<button className="btn btn-success" onClick={handleClick}>
												Save Changes
											</button>
										</div>
									</div>
								</div>
								<div className="col-md-4">
									<div className="card">
										<div className="card-header">
											<h4 className="card-title">Sidebar</h4>
										</div>
										<div className="card-body">
											<p>Sidebar content here.</p>
										</div>
									</div>
								</div>
							</div>
						</div>
					</main>
				</div>
			</main>`,
			contains: []string{
				"React.createElement('div', {className: \"app\"}",
				"React.createElement('header', {className: \"app-header\"}",
				"React.createElement('nav', {className: \"navbar\"}",
				"React.createElement('img', {src: \"/logo.svg\", alt: \"Brand\", className: \"logo\"}",
				"React.createElement('ul', {className: \"navbar-nav\"}",
				"React.createElement('li', {className: \"nav-item\"}",
				"React.createElement('a', {href: \"/dashboard\", className: \"nav-link active\"}, 'Dashboard')",
				"React.createElement('a', {href: \"/settings\", className: \"nav-link\"}, 'Settings')",
				"React.createElement('main', {className: \"app-main\"}",
				"React.createElement('div', {className: \"container-fluid\"}",
				"React.createElement('div', {className: \"row\"}",
				"React.createElement('div', {className: \"col-md-8\"}",
				"React.createElement('div', {className: \"card\"}",
				"React.createElement('h3', {className: \"card-title\"}, 'Main Content')",
				"React.createElement('p', null, 'This is the main content area.')",
				"React.createElement('button', {className: \"btn btn-success\", onClick: handleClick}, 'Save Changes')",
				"React.createElement('div', {className: \"col-md-4\"}",
				"React.createElement('h4', {className: \"card-title\"}, 'Sidebar')",
				"React.createElement('p', null, 'Sidebar content here.')",
			},
			notContains: []string{
				"<script>",
				"<style>",
				"class=",
				"htmlFor=",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TSX2JS(tt.input)

			// Clean up the result for comparison (remove extra whitespace)
			result = strings.ReplaceAll(result, "\n", "")
			result = strings.ReplaceAll(result, "  ", " ")
			result = strings.TrimSpace(result)

			// Test exact match if expected is provided
			if tt.expected != "" {
				expected := strings.ReplaceAll(tt.expected, "\n", "")
				expected = strings.ReplaceAll(expected, "  ", " ")
				expected = strings.TrimSpace(expected)

				if result != expected {
					t.Errorf("TSX2JS() = %v, want %v", result, expected)
					t.Logf("Input: %s", tt.input)
					t.Logf("Got: %s", result)
					t.Logf("Expected: %s", expected)
				}
			}

			// Test contains
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("TSX2JS() result does not contain expected string: %v", expected)
					t.Logf("Input: %s", tt.input)
					t.Logf("Result: %s", result)
				}
			}

			// Test not contains
			for _, unwanted := range tt.notContains {
				if strings.Contains(result, unwanted) {
					t.Errorf("TSX2JS() result should not contain: %v", unwanted)
					t.Logf("Input: %s", tt.input)
					t.Logf("Result: %s", result)
				}
			}
		})
	}
}

func TestTSX2JS_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string // Check if result contains these strings
	}{
		{
			name:     "Empty input",
			input:    "",
			contains: []string{""},
		},
		{
			name:     "Only whitespace",
			input:    "   \n\t   ",
			contains: []string{""},
		},
		{
			name:     "Single text node",
			input:    "<main>Hello World</main>",
			contains: []string{"'Hello World'"},
		},
		{
			name:     "Multiple text nodes",
			input:    "<main>Hello <strong>World</strong>!</main>",
			contains: []string{"'Hello'", "React.createElement('strong'", "'!'"},
		},
		{
			name:     "Attributes with special characters",
			input:    `<main><div data-test="value with spaces" data-id="123" className="test-class"></div></main>`,
			contains: []string{"data-test: \"value with spaces\"", "data-id: \"123\"", "className: \"test-class\""},
		},
		{
			name:     "Boolean attributes",
			input:    `<main><input type="checkbox" checked disabled /></main>`,
			contains: []string{"type: \"checkbox\"", "checked: \"\"", "disabled: \"\""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TSX2JS(tt.input)

			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("TSX2JS() result does not contain expected string: %v", expected)
					t.Logf("Input: %s", tt.input)
					t.Logf("Result: %s", result)
				}
			}
		})
	}
}

func TestTSX2JS_WithImports(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "TSX with imports and export",
			input: `import React from 'react';
import Simple from '../components/Simple';

export default function TestPage({page}) {
    return (
        <main>
            <div className="container">
                <h1>Test Page</h1>
                <Simple />
            </div>
        </main>
    );
}`,
			expected: `React.createElement('div', {className: "container"}, React.createElement('h1', null, 'Test Page'), React.createElement(Simple, null))`,
		},
		{
			name: "TSX with multiple imports",
			input: `import React from 'react';
import { useState } from 'react';
import Simple from '../components/Simple';
import Counter from '../components/Counter';

export default function ComplexPage({page}) {
    return (
        <main>
            <div>
                <Simple />
                <Counter />
            </div>
        </main>
    );
}`,
			expected: `React.createElement('div', null, React.createElement(Simple, null, React.createElement(Counter, null)))`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TSX2JS(tt.input)

			// Clean up the result for comparison
			result = strings.ReplaceAll(result, "\n", "")
			result = strings.ReplaceAll(result, "  ", " ")
			result = strings.TrimSpace(result)

			// Clean up expected for comparison
			expected := strings.ReplaceAll(tt.expected, "\n", "")
			expected = strings.ReplaceAll(expected, "  ", " ")
			expected = strings.TrimSpace(expected)

			if result != expected {
				t.Errorf("TSX2JS() = %v, want %v", result, expected)
				t.Logf("Input: %s", tt.input)
				t.Logf("Got: %s", result)
				t.Logf("Expected: %s", expected)
			}
		})
	}
}

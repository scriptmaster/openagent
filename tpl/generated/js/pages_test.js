
// Embedded Component JS
///////////////////////////////

// Component: simple
function Simple() {
    return (
        React.createElement('div', null, 'Simple Component')
    );
}

///////////////////////////////

// Component prototype methods


///////////////////////////////

// Main Page JS

// Main Component JS (converted to JS from TSX)
function Test({page}) {
    return (
React.createElement('div', {className: 'container-xl'}, React.createElement('div', {className: 'row'}, React.createElement('div', {className: 'col-12'}, React.createElement('div', {className: 'page-header'}, React.createElement('h1', {className: 'page-title'}, 'React Tests - Interactive Components! 🚀'), React.createElement('p', {className: 'text-muted'}, 'Testing various React functionality with JSX transpilation')))), React.createElement('div', {className: 'row mb-4'}, React.createElement('div', {className: 'col-12'}, React.createElement('div', {className: 'alert alert-info'}, React.createElement('h4', {className: 'alert-heading'}, '🎉 React Transpilation Success!'), React.createElement('p', null, 'This page demonstrates our custom JSX-to-React transpilation system. All components below are written in JSX and automatically converted to React.createElement calls!'), React.createElement('hr', null), React.createElement('p', {className: 'mb-0'}, 'Check the browser console to see the transpiled React code in action.')))), React.createElement('div', {className: 'row mb-4'}, React.createElement('div', {className: 'col-12'}, React.createElement('h2', {className: 'h3 mb-3'}, 'Part 1: Basic React Demos'))), React.createElement('div', {className: 'row mb-4'}, React.createElement('div', {className: 'col-md-6'}, React.createElement('div', {className: 'card'}, React.createElement('div', {className: 'card-header'}, React.createElement('h5', {className: 'card-title mb-0'}, 'Demo 1: Counter')), React.createElement('div', {className: 'card-body'}, React.createElement('p', {className: 'card-text'}, 'Simple counter with increment/decrement buttons'), React.createElement(Simple, {}), React.createElement('div', {id: 'component-test-component'}))))), React.createElement('div', {className: 'row mb-4'}, React.createElement('div', {className: 'col-md-6'}, React.createElement('div', {className: 'card'}, React.createElement('div', {className: 'card-header'}, React.createElement('h5', {className: 'card-title mb-0'}, 'Demo 2: Toggle')), React.createElement('div', {className: 'card-body'}, React.createElement('p', {className: 'card-text'}, 'Toggle button that shows/hides content'), React.createElement('div', {id: 'demo2-toggle'}))))))
    );
}

///////////////////////////////

// Original JS content

        console.log('please wait...');
        
        // Extend Test component prototype with all demo functionality
        Object.assign(Test.prototype, {
            // Demo 1: Counter functionality
            counterState: { count: 0 },
            incrementCounter() {
                this.counterState.count++;
                this.forceUpdate();
            },
            decrementCounter() {
                this.counterState.count--;
                this.forceUpdate();
            },
            resetCounter() {
                this.counterState.count = 0;
                this.forceUpdate();
            },
            
            // Demo 2: Toggle functionality
            toggleState: { isVisible: false },
            toggleVisibility() {
                this.toggleState.isVisible = !this.toggleState.isVisible;
                this.forceUpdate();
            },
        });
    


///////////////////////////////

// Make component available globally for hydration
window.Test = Test;

// React hydration using common utilities
try {
    // Use the global hydration function from _common.js
    window.hydrateReactApp('Test', { 
        page: window.pageData || {},
        container: 'main',
		layout: React.createElement('div', {}, 'Layout placeholder')
    });
} catch(e) {
    console.error('React hydration error:', e);
}
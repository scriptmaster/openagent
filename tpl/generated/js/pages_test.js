
// ╔═════════════════════════════════════════════════════════════════════════════
// ║                           🧩 EMBEDDED COMPONENT JS 🧩                        
// ╚═════════════════════════════════════════════════════════════════════════════

// ┌─────────────────────────────────────────────────────────────────────────────
// │  🎯 COMPONENT: SIMPLE                                            
// └─────────────────────────────────────────────────────────────────────────────
function Simple() {
    return (
        React.createElement('div', null, React.createElement('span', null, 'Simple Component 2:'), React.createElement('button', null, 'Hey!'))
    );
}

// ╔══════════════════════════════════════════════════════════════════════════════
// ║                        🔧 COMPONENT PROTOTYPE METHODS 🔧                        
// ╚══════════════════════════════════════════════════════════════════════════════


    Object.assign({}, Simple.prototype, {
        hey() {
            alert('hey');
        }
    });



// ╔══════════════════════════════════════════════════════════════════════════════
// ║                            ⚛️  MAIN PAGE JS ⚛️                               
// ╚══════════════════════════════════════════════════════════════════════════════


// ╔══════════════════════════════════════════════════════════════════════════════
// ║                    ⚛️  MAIN COMPONENT JS (TSX → JS) ⚛️                      
// ╚══════════════════════════════════════════════════════════════════════════════
// Component imports

function Test({page}) {
    return (
React.createElement('div', {className: 'container-xl'}, React.createElement('div', {className: 'card-body'}, React.createElement('p', {className: 'card-text'}, 'Simple counter with increment/decrement buttons'), React.createElement(Simple, null)))
    );
}

// ╔══════════════════════════════════════════════════════════════════════════════
// ║                        📜 ORIGINAL JS CONTENT 📜                            
// ╚══════════════════════════════════════════════════════════════════════════════

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



// ╔══════════════════════════════════════════════════════════════════════════════
// ║                        💧 HYDRATION 💧                            
// ╚══════════════════════════════════════════════════════════════════════════════

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

// ╔═════════════════════════════════════════════════════════════════════════════
// ║                           🧩 EMBEDDED COMPONENT JS 🧩                        
// ╚═════════════════════════════════════════════════════════════════════════════

// ╔══════════════════════════════════════════════════════════════════════════════
// ║                            ⚛️  MAIN PAGE JS ⚛️                               
// ╚══════════════════════════════════════════════════════════════════════════════


// ╔══════════════════════════════════════════════════════════════════════════════
// ║                    ⚛️  MAIN COMPONENT JS (TSX → JS) ⚛️                      
// ╚══════════════════════════════════════════════════════════════════════════════
function Test({page}) {
    return (
React.createElement('div', {className: 'container-xl'}, React.createElement('div', {className: 'card-body'}, React.createElement('div', {id: 'component2-simple'})))
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
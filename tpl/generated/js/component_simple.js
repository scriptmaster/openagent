function Simple() {
    return (
        React.createElement('div', null, React.createElement('span', null, 'Simple Component 2:'))
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


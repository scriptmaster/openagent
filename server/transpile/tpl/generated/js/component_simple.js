function Simple() {
    return (
        React.createElement('div', {className: "simple-component"}, React.createElement('h2', null, 'Simple Component'), React.createElement('button', {onclick: handleClick}, 'Click me'))
    );
}

// ╔══════════════════════════════════════════════════════════════════════════════
// ║                        🔧 COMPONENT PROTOTYPE METHODS 🔧                        
// ╚══════════════════════════════════════════════════════════════════════════════


		Simple.prototype.handleClick = function() {
			console.log('Simple button clicked');
		};
		
		Simple.prototype.init = function() {
			this.active = true;
		};
	

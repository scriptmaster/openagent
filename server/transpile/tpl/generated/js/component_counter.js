function Counter() {
    return (
        React.createElement('div', {className: "counter-component"}, React.createElement('h2', null, 'Counter: {this.value}'), React.createElement('button', {onclick: increment}, '+'), React.createElement('button', {onclick: decrement}, '-'))
    );
}

// ╔══════════════════════════════════════════════════════════════════════════════
// ║                        🔧 COMPONENT PROTOTYPE METHODS 🔧                        
// ╚══════════════════════════════════════════════════════════════════════════════


		Counter.prototype.increment = function() {
			this.value++;
		};
		
		Counter.prototype.decrement = function() {
			this.value--;
		};
		
		Counter.prototype.init = function() {
			this.value = 0;
		};
	

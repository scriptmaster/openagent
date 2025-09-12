function Counter() {
    return (
        React.createElement('div', {className: 'mb-3'}, 
    React.createElement('span', {className: 'badge bg-primary fs-4'}, 'Count: ', this.counterState.count)
)
React.createElement('div', {className: 'btn-group', role: 'group'}, 
    <button className="btn btn-outline-danger" onClick={decrementCounter}>-</button>
    <button className="btn btn-outline-success" onClick={incrementCounter}>+</button>
    <button className="btn btn-outline-secondary" onClick={resetCounter}>Reset</button>
)
    );
}

///////////////////////////////

// Component prototype methods

    // Counter component prototype methods
    Object.assign(Counter.prototype, {
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
        }
    });


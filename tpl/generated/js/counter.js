
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


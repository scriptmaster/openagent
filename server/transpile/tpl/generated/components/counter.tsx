export default function Counter() {
    return (
        <div className="counter-component">
		<h2>Counter: {this.value}</h2>
		<button onClick={increment}>+</button>
		<button onClick={decrement}>-</button>
	</div>
	
    );
}
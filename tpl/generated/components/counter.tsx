export default function Counter() {
    return (
        <div className="mb-3">
    <span className="badge bg-primary fs-4">Count: {this.counterState.count}</span>
</div>
<div className="btn-group" role="group">
    <button className="btn btn-outline-danger" onClick={decrementCounter}>-</button>
    <button className="btn btn-outline-success" onClick={incrementCounter}>+</button>
    <button className="btn btn-outline-secondary" onClick={resetCounter}>Reset</button>
</div>



    );
}
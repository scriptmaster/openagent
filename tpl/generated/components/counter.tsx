export default function Counter() {
    return (
<div className="mb-3">
    <span className="badge bg-primary fs-4">Count: {this.counterState.count}</span>
</div>
<div className="btn-group" role="group">
    <button className="btn btn-outline-danger" onClick={() => this.decrementCounter()}>-</button>
    <button className="btn btn-outline-success" onClick={() => this.incrementCounter()}>+</button>
    <button className="btn btn-outline-secondary" onClick={() => this.resetCounter()}>Reset</button>
</div>



    );
}
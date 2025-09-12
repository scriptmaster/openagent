// Component imports
import Simple from '../components/simple';

export default function Test({page}) {
    return (
<main>
<div className="container-xl">
    <div className="card-body">
        <p className="card-text">Simple counter with increment/decrement buttons</p>
        <Simple />
    </div>
</div>
</main>
    );
}
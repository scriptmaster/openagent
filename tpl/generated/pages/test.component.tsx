// Component imports
import Simple from '../components/simple';

export default function Test({page}) {
    return (
<main>
<div className="container-xl">
    <div className="card-body">
        <Simple suppressHydrationWarning={true} />
    </div>
</div>
</main>
    );
}
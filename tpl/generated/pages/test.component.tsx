// Component imports
import Simple from '../components/simple';

export default function Test({page}: {page: any}) {
    return (
<div className="container-xl">
    <div className="card-body">
        <Simple suppressHydrationWarning={true} />
        <span>Element beside a component.</span>
    </div>
</div>
    );
}
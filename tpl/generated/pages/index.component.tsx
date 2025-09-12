export default function Index({page}) {
    return (
<main>
<div className="hero-section">
    <div className="hero-content">
        <h1 className="hero-title">
            <h2>{page.AppName}</h2>
        </h1>
        <p className="hero-subtitle">
            Welcome to OpenAgent - Your AI-powered development platform
        </p>
        <div className="cta-buttons">
            <a href="/login" className="btn-hero btn-primary">Get Started</a>
            <a href="/config" className="btn-hero btn-outline">Learn More</a>
        </div>
    </div>
</div>
</main>
    );
}
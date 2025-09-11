export default function Index({page}: {page: Page}) {
    return (
<meta charset="UTF-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0" />
<title>{page.PageTitle}</title>
{/* Tabler CSS */}
<link rel="stylesheet" href="/static/css/tabler.min.css" />
{/* Tabler Icons */}
<link rel="stylesheet" href="/static/css/tabler-icons.min.css" />
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
{/* Alpine.js */}
<script src="/tsx/js/index.js"></script>
    );
}
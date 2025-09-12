export default function Error404({page}) {
    return (
<main>
<div className="error-container">
    <div className="error-code">404</div>
    <div className="error-message">Page Not Found</div>
    <p>The requested page could not be found.</p>
    <a href="/" className="back-link">Return to Home</a>
</div>
</main>
    );
}
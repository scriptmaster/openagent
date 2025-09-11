export default function Error404({page}: {page: Page}) {
    return (
<>
<html lang="en">
<head>
    <meta charset="UTF-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
    <title>404 - Page Not Found</title>
    {/* Tabler CSS */}
    <link rel="stylesheet" href="/static/css/tabler.min.css"/>

    <link rel="stylesheet" href="/tsx/css/error_404.css" />
</head>
<body>
    <div className="error-container">
        <div className="error-code">404</div>
        <div className="error-message">Page Not Found</div>
        <p>The requested page could not be found.</p>
        <a href="/" className="back-link">Return to Home</a>
    </div>

<script src="/tsx/js/_common.js"></script>
</body>
</html>
</>
    );
}
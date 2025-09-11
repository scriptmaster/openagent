import Layout from '../layouts/layout_pages';

export default function Error404({page}: {page: Page}) {
    return (
        <Layout page={page} linkPaths={`/tsx/css/error_404.css`} scriptPaths={`/tsx/js/_common.js`}>
<div className="error-container">
    <div className="error-code">404</div>
    <div className="error-message">Page Not Found</div>
    <p>The requested page could not be found.</p>
    <a href="/" className="back-link">Return to Home</a>
</div>
        </Layout>
    );
}
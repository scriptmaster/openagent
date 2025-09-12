import Layout from '../../generated/layouts/layout_pages';

export default function Test({page}: {page: Page}) {
    return (
        <Layout page={page} linkPaths={``} scriptPaths={`/tsx/js/_common.js`}>
<div className="container">
    <h1>Custom View Override 4: {page.PageTitle}</h1>
    <p>This is a custom view that overrides the generated template!</p>
    <p>App: {page.AppName}</p>
    <p>Version: {page.AppVersion}</p>
    <div className="alert alert-info">
        <strong>View Override Working!</strong> This template is loaded from tpl/views/pages/test.tsx
    </div>
</div>
        </Layout>
    );
}
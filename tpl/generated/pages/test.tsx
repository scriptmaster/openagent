import Layout from '../layouts/layout_pages';

export default function Test({page}: {page: Page}) {
    return (
        <Layout page={page} linkTags={``} scriptTags={`<script src="/tsx/js/_common.js"></script>`}>
<div className="container">
    <h1>{page.PageTitle}</h1>
    <p>This is a test page to verify the template system is working.</p>
    <p>App: {page.AppName}</p>
    <p>Version: {page.AppVersion}</p>
</div>
        </Layout>
    );
}
import Layout from '../layouts/layout_landing';

export default function Index({page}: {page: Page}) {
    return (
        <Layout page={page} linkTags={``} scriptTags={`<script src="/tsx/js/_common.js"></script>`}>
<div className="container">
    <h1>{page.LandingTitle}</h1>
    <p>This is a landing page with specific landing page data.</p>
    <p>Hero Text: {page.HeroText}</p>
    <p>CTA Button: {page.CtaButton}</p>
    <div className="text-center">
        <button className="btn btn-primary btn-lg">{page.CtaButton}</button>
    </div>
</div>
        </Layout>
    );
}
import Layout from '../layouts/layout_landing';

export default function Test2.page.landing({page}: {page: Page}) {
    return (
        <Layout page={page} linkPaths={``} scriptPaths={`/tsx/js/_common.js`}>
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
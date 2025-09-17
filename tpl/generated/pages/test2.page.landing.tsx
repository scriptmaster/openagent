import Layout from '../layouts/layout_landing';
import App from './test2.page.landing.component';

export default function Test2.page.landingLayout({page}: {page: Page}) {
    return (
        <Layout page={page} linkPaths={`/tsx/css/test2.page.landing.css`} scriptPaths={`/tsx/js/_common.js`}>
            <App page={page} />
        </Layout>
    );
}
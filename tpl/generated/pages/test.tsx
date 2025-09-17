import Layout from '../layouts/layout_pages';
import App from './test.component';

export default function TestLayout({page}: {page: Page}) {
    return (
        <Layout page={page} linkPaths={`/tsx/css/test.css`} scriptPaths={`/tsx/js/_common.js,/tsx/js/pages_test.js`}>
            <App page={page} />
        </Layout>
    );
}
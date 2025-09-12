import Layout from '../layouts/layout_pages';
import App from './index.component';

export default function IndexLayout({page}: {page: Page}) {
    return (
        <Layout page={page} linkPaths={`/tsx/css/index.css`} scriptPaths={`/tsx/js/_common.js`}>
            <App page={page} />
        </Layout>
    );
}
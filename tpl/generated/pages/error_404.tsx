import Layout from '../layouts/layout_pages';
import App from './error_404.component';

export default function Error404Layout({page}: {page: Page}) {
    return (
        <Layout page={page} linkPaths={`/tsx/css/error_404.css`} scriptPaths={`/tsx/js/_common.js`}>
            <App page={page} />
        </Layout>
    );
}
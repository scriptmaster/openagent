import Layout from '../layouts/layout_pages';
import App from './config.component';

export default function ConfigLayout({page}: {page: Page}) {
    return (
        <Layout page={page} linkPaths={`/tsx/css/config.css`} scriptPaths={`/tsx/js/_common.js,/tsx/js/pages_config.js`}>
            <App page={page} />
        </Layout>
    );
}
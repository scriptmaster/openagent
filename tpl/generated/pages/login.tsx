import Layout from '../layouts/layout_pages';
import App from './login.component';

export default function LoginLayout({page}: {page: Page}) {
    return (
        <Layout page={page} linkPaths={`/tsx/css/login.css`} scriptPaths={`/tsx/js/_common.js`}>
            <App page={page} />
        </Layout>
    );
}
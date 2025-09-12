import Layout from '../layouts/layout_pages';
import App from './admin.component';

export default function AdminLayout({page}: {page: Page}) {
    return (
        <Layout page={page} linkPaths={``} scriptPaths={`/tsx/js/_common.js,/tsx/js/admin_admin.js`}>
            <App page={page} />
        </Layout>
    );
}
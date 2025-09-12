import Layout from '../layouts/layout_pages';
import App from './dashboard.component';

export default function DashboardLayout({page}: {page: Page}) {
    return (
        <Layout page={page} linkPaths={``} scriptPaths={`/tsx/js/_common.js,/tsx/js/app_dashboard.js`}>
            <App page={page} />
        </Layout>
    );
}
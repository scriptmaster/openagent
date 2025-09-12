import Layout from '../layouts/layout_pages';
import App from './maintenance-login.component';

export default function MaintenanceLoginLayout({page}: {page: Page}) {
    return (
        <Layout page={page} linkPaths={`/tsx/css/maintenance-login.css`} scriptPaths={`/tsx/js/_common.js,/tsx/js/pages_maintenance-login.js`}>
            <App page={page} />
        </Layout>
    );
}
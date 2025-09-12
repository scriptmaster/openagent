import Layout from '../layouts/layout_pages';
import App from './maintenance.component';

export default function MaintenanceLayout({page}: {page: Page}) {
    return (
        <Layout page={page} linkPaths={`/tsx/css/maintenance.css`} scriptPaths={`/tsx/js/_common.js,/tsx/js/pages_maintenance.js`}>
            <App page={page} />
        </Layout>
    );
}
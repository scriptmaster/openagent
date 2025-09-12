import Layout from '../layouts/layout_pages';
import App from './projects.component';

export default function ProjectsLayout({page}: {page: Page}) {
    return (
        <Layout page={page} linkPaths={``} scriptPaths={`/tsx/js/_common.js,/tsx/js/app_projects.js`}>
            <App page={page} />
        </Layout>
    );
}
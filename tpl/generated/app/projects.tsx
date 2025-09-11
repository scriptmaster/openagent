import Layout from '../layouts/layout_pages';

export default function Projects({page}: {page: Page}) {
    return (
        <Layout page={page} linkPaths={``} scriptPaths={`/tsx/js/_common.js,/tsx/js/projects.js`}>
<meta charset="UTF-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0" />
<title>Projects - {page.AppName}</title>
<link rel="stylesheet" href="/static/css/tabler.min.css" />
<link rel="stylesheet" href="/static/css/tabler-icons.min.css" />
<div data-x-data="projectsApp()" x-init="loadProjects()">
    <div className="page-header">
        <div className="container-xl">
            <div className="row g-2 align-items-center">
                <div className="col">
                    <div className="page-pretitle">
                        Project Management
                    </div>
                    <h2 className="page-title">
                        Projects
                    </h2>
                </div>
                <div className="col-auto ms-auto d-print-none">
                    <div className="btn-list">
                        <button className="btn btn-primary d-none d-sm-inline-block" data-click="showCreateModal = true">
                            <i className="ti ti-plus"></i>
                            Create New Project
                        </button>
                        <button className="btn btn-primary d-sm-none btn-icon" data-click="showCreateModal = true">
                            <i className="ti ti-plus"></i>
                        </button>
                    </div>
                </div>
            </div>
        </div>
    </div>
    <div className="page-body">
        <div className="container-xl">
            <div className="row row-deck row-cards">
                <div className="col-12">
                    <div className="card">
                        <div className="card-header">
                            <h3 className="card-title">Your Projects</h3>
                        </div>
                        <div className="card-body">
                            <div data-x-show="loading" className="text-center py-4">
                                <div className="spinner-border" role="status">
                                    <span className="visually-hidden">Loading...</span>
                                </div>
                            </div>
                            
                            <div data-x-show="!loading && projects.length === 0" className="text-center py-4">
                                <div className="empty">
                                    <div className="empty-icon">
                                        <i className="ti ti-folder-off"></i>
                                    </div>
                                    <p className="empty-title">No projects found</p>
                                    <p className="empty-subtitle text-muted">
                                        Get started by creating your first project.
                                    </p>
                                    <div className="empty-action">
                                        <button className="btn btn-primary" data-click="showCreateModal = true">
                                            <i className="ti ti-plus"></i>
                                            Create Project
                                        </button>
                                    </div>
                                </div>
                            </div>
                            <div data-x-show="!loading && projects.length > 0" className="row">
                                <template data-x-for="project in projects" data-key="project.id">
                                    <div className="col-md-6 col-lg-4 mb-3">
                                        <div className="card card-sm">
                                            <div className="card-body">
                                                <div className="row align-items-center">
                                                    <div className="col-auto">
                                                        <span className="bg-primary text-white avatar">
                                                            <i className="ti ti-folder"></i>
                                                        </span>
                                                    </div>
                                                    <div className="col">
                                                        <div className="font-weight-medium">
                                                            <span data-x-text="project.name"></span>
                                                        </div>
                                                        <div className="text-muted">
                                                            <span data-x-text="project.domain_name"></span>
                                                        </div>
                                                    </div>
                                                </div>
                                            </div>
                                            <div className="card-footer">
                                                <div className="row align-items-center">
                                                    <div className="col">
                                                        <div className="text-muted">
                                                            <i className="ti ti-database me-1"></i>
                                                            <span data-x-text="project.connection_count"></span> connections
                                                        </div>
                                                    </div>
                                                    <div className="col-auto">
                                                        <a data-href="`/projects/${project.id}`" className="btn btn-sm btn-primary">
                                                            View
                                                        </a>
                                                    </div>
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                </template>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</div>
        </Layout>
    );
}
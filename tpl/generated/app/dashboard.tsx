import Layout from '../layouts/layout_pages';

export default function Dashboard({page}: {page: Page}) {
    return (
        <Layout page={page} linkTags={``} scriptTags={`<script src="/tsx/js/_common.js"></script>
<script src="/tsx/js/dashboard.js"></script>`}>
<meta charset="UTF-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0" />
<title>Dashboard - {page.AppName}</title>
<link rel="stylesheet" href="/static/css/tabler.min.css" />
<link rel="stylesheet" href="/static/css/tabler-icons.min.css" />
<div className="page-header">
    <div className="container-xl">
        <div className="row g-2 align-items-center">
            <div className="col">
                <div className="page-pretitle">
                    Overview
                </div>
                <h2 className="page-title">
                    Dashboard
                </h2>
            </div>
        </div>
    </div>
</div>
<div className="page-body">
    <div className="container-xl">
        <div className="row row-deck row-cards">
            <div className="col-12">
                <div className="row row-cards">
                    <div className="col-sm-6 col-lg-3">
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
                                            Projects
                                        </div>
                                        <div className="text-muted">
                                            {page.ProjectCount}
                                        </div>
                                    </div>
                                </div>
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
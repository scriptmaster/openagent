export default function Layout({page}: {page: Page}) {
    return (
<>
<meta charset="UTF-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0" />
<title>{page.PageTitle || page.AppName}</title>
<link rel="stylesheet" href="/static/css/tabler.min.css" />
<link rel="stylesheet" href="/static/css/tabler-icons.min.css" />
<div className="page">
    <div className="page-wrapper">
        <div className="page-header d-print-none">
            <div className="container-xl">
                <div className="row g-2 align-items-center">
                    <div className="col">
                        <div className="page-pretitle">
                            {page.PageTitle}
                        </div>
                        <h2 className="page-title">
                            {page.PageTitle}
                        </h2>
                    </div>
                    <div className="col-auto ms-auto d-print-none">
                        <div className="btn-list">
                            {/* Page header actions can be added here */}
                        </div>
                    </div>
                </div>
            </div>
        </div>
        <div className="page-body">
            <div className="container-xl">
                {/* Main content area */}
                <div className="row row-deck row-cards">
                    <div className="col-12">
                        {/* Content will be inserted here */}
                    </div>
                </div>
            </div>
        </div>
        <footer className="footer footer-transparent d-print-none">
            <div className="container-xl">
                <div className="row text-center align-items-center flex-row-reverse">
                    <div className="col-lg-auto ms-lg-auto">
                        <ul className="list-inline list-inline-dots mb-0">
                            <li className="list-inline-item">
                                <a href="/config" className="link-secondary">Configuration</a>
                            </li>
                        </ul>
                    </div>
                    <div className="col-12 col-lg-auto mt-3 mt-lg-0">
                        <ul className="list-inline list-inline-dots mb-0">
                            <li className="list-inline-item">
                                &copy; 2025 {page.AppName} - v{page.AppVersion}
                            </li>
                        </ul>
                    </div>
                </div>
            </div>
        </footer>
    </div>
</div>
<script src="/tsx/js/_common.js"></script>
<script src="/tsx/js/layout.js"></script>
</>
    );
}
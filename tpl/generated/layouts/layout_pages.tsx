export default function LayoutPages({page, children, linkPaths, scriptPaths}: {page: any, children?: any, linkPaths?: string, scriptPaths?: string}) {
    return (
<>
<html lang="en">
<head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>{page.PageTitle? page.PageTitle + ' - ' + page.AppName: page.AppName}</title>
    {/* Favicon */}
<link rel="shortcut icon" type="image/x-icon" href="/static/favicon.ico" />
<link rel="icon" type="image/x-icon" href="/static/favicon.ico" />
<link rel="icon" type="image/svg+xml" href="/static/img/icon.svg" />
<link rel="icon" href="/static/favicon.ico" />
{/* Tabler CSS */}
<link rel="stylesheet" href="/static/css/tabler.min.css" />
{/* Tabler Icons */}
<link rel="stylesheet" href="/static/css/tabler-icons.min.css" />
{/* Custom Styles */}
<link rel="stylesheet" href="/static/css/custom.css" />
{/* <script src="/static/js/alpine.min.js" defer></script> */}
{/* <script src="/static/js/hyperscript.min.js"></script> */}
    {linkPaths && linkPaths.split(',').map((path, index) => (
    <link key={'gen-link-'+index} rel="stylesheet" href={path.trim()} />
))}
</head>
<body className="theme-pista">
    <div className="page">
        <div className="page-wrapper">
            <header className="navbar navbar-expand-md navbar-light d-print-none">
    <div className="container-xl">
        <button className="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbar-menu">
            <span className="navbar-toggler-icon"></span>
        </button>
        <h1 className="navbar-brand navbar-brand-autodark d-none-navbar-horizontal pe-0 pe-md-3">
            <a href="/">
                <img src="/static/img/logo.svg" height="36" alt="{page.AppName}"/>
            </a>
        </h1>
        <div className="navbar-nav flex-row order-md-last">
            {page.User ? (
                <>
                    {/* User is logged in */}
                    <div className="nav-item dropdown">
                        <a href="#" className="nav-link d-flex lh-1 text-reset p-0" data-bs-toggle="dropdown" aria-label="Open user menu">
                            <span className="avatar avatar-sm" style="background-image: url(/static/img/default-avatar.png)"></span>
                            <div className="d-none d-xl-block ps-2">
                                <div>{page.User.Email}</div>
                                <div className="mt-1 small text-muted">{page.User.IsAdmin ? 'Admin' : 'User'}</div>
                            </div>
                        </a>
                        <div className="dropdown-menu dropdown-menu-end dropdown-menu-arrow">
                            <a href="#" className="dropdown-item">Profile</a>
                            <div className="dropdown-divider"></div>
                            <a href="/logout" className="dropdown-item">Logout</a>
                        </div>
                    </div>
                </>
            ) : (
                <>
                    {/* User is not logged in */}
                    <div className="nav-item">
                        <a href="/login" className="btn btn-outline-primary">
                            <i className="ti ti-login me-1"></i>
                            Login
                        </a>
                    </div>
                </>
            )}
        </div>
        {/* Navigation Menu (Left Side) */}
        <div className="collapse navbar-collapse" id="navbar-menu">
            <div className="d-flex flex-column flex-md-row flex-fill align-items-stretch align-items-md-center">
                <ul className="navbar-nav">
                    <li className="nav-item">
                        <a className="nav-link" href="/dashboard">
                            <span className="nav-link-icon d-md-none d-lg-inline-block">
                                <i className="ti ti-dashboard"></i>
                            </span>
                            <span className="nav-link-title">
                                Dashboard
                            </span>
                        </a>
                    </li>
                     <li className="nav-item">
                        <a className="nav-link" href="/projects">
                            <span className="nav-link-icon d-md-none d-lg-inline-block">
                                <i className="ti ti-briefcase"></i>
                            </span>
                            <span className="nav-link-title">
                                Projects
                            </span>
                        </a>
                    </li>
                    {/* Add other nav items here */}
                </ul>
            </div>
        </div>
    </div>
</header>
            
            <div className="page-body">
                <div className="container-xl">
                    <div className="row row-cards">
                        <div className="col-12">
                            {/* Page content will be inserted here */}
                            {children}
                        </div>
                    </div>
                </div>
            </div>
            
            <footer className="footer footer-transparent d-print-none">
    <div className="container-xl">
        <div className="row text-center align-items-center flex-row-reverse">
            <div className="col-lg-auto ms-lg-auto">
                <ul className="list-inline list-inline-dots mb-0">
                    <li className="list-inline-item"><a href="/about" className="link-secondary">About</a></li>
                    <li className="list-inline-item"><a href="/contact" className="link-secondary">Contact</a></li>
                    <li className="list-inline-item"><a href="/privacy" className="link-secondary">Privacy</a></li>
                </ul>
            </div>
            <div className="col-12 col-lg-auto mt-3 mt-lg-0">
                <ul className="list-inline list-inline-dots mb-0">
                    <li className="list-inline-item">
                        Copyright © {new Date().getFullYear()}
                        &nbsp;
                        <a href="/" className="link-secondary">{page.AppName}</a>.
                        All rights reserved!
                    </li>
                    <li className="list-inline-item">
                        <a href="#" className="link-secondary" rel="noopener">
                            Version {page.AppVersion}
                        </a>
                    </li>
                </ul>
            </div>
        </div>
    </div>
</footer>
        </div>
    </div>
    {scriptPaths && scriptPaths.split(',').map((path, index) => (
    <script key={'gen-script-'+index} src={path.trim()}></script>
))}
</body>
</html>
</>
    );
}
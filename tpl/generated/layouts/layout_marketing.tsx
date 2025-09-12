export default function LayoutMarketing({page, children, linkPaths, scriptPaths}: {page: any, children: any, linkPaths?: string, scriptPaths?: string}) {
    return (
<>
<html lang="en">
<head>
    <meta charset="UTF-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
    <title>{page.Title}</title>
    {page.Head}
    {linkPaths && linkPaths.split(',').map((path, index) => (
    <link key={'gen-link-'+index} rel="stylesheet" href={path.trim()} />
))}
</head>
<body className="theme-pista">
    <div className="page">
        {/* Marketing Header */}
        <header className="navbar navbar-expand-md navbar-light">
            <div className="container-xl">
                <h1 className="navbar-brand navbar-brand-autodark">
                    <a href="/">
                        <img src="/static/img/logo.svg" width="110" height="32" alt="OpenAgent" className="navbar-brand-image"/>
                    </a>
                </h1>
                <div className="navbar-nav flex-row order-md-last">
                    <div className="nav-item">
                        <a href="/login" className="btn btn-outline-primary">Login</a>
                    </div>
                </div>
            </div>
        </header>
        {/* Main Content */}
        <div className="page-wrapper">
            <div className="page-body">
                <div className="container-xl">
                    {children}
                </div>
            </div>
        </div>
        {/* Marketing Footer */}
        <footer className="footer footer-transparent d-print-none">
            <div className="container-xl">
                <div className="row text-center align-items-center flex-row-reverse">
                    <div className="col-lg-auto ms-lg-auto">
                        <ul className="list-inline list-inline-dots mb-0">
                            <li className="list-inline-item"><a href="/" className="link-secondary">Home</a></li>
                            <li className="list-inline-item"><a href="/config" className="link-secondary">Setup</a></li>
                        </ul>
                    </div>
                    <div className="col-12 col-lg-auto mt-3 mt-lg-0">
                        <ul className="list-inline list-inline-dots mb-0">
                            <li className="list-inline-item">
                                Copyright © 2025 <a href="/" className="link-secondary">OpenAgent</a>. All rights reserved.
                            </li>
                        </ul>
                    </div>
                </div>
            </div>
        </footer>
    </div>
    {scriptPaths && scriptPaths.split(',').map((path, index) => (
    <script key={'gen-script-'+index} src={path.trim()}></script>
))}
</body>
</html>
</>
    );
}
export default function LayoutLanding({page, children, linkPaths, scriptPaths}: {page: any, children?: any, linkPaths?: string, scriptPaths?: string}) {
    return (
<>
<html lang="en">
<head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>{page.PageTitle || page.AppName}</title>
    <link rel="stylesheet" href="/static/css/tabler.min.css" />
    <link rel="stylesheet" href="/static/css/tabler-icons.min.css" />
    {linkPaths && linkPaths.split(',').map((path, index) => (
    <link key={'gen-link-'+index} rel="stylesheet" href={path.trim()} />
))}
</head>
<body className="theme-pista">
    <div className="page">
        <div className="page-wrapper">
            <div className="page-body">
                <div className="container-xl">
                    <div className="row row-deck row-cards">
                        <div className="col-12">
                            {/* Landing page content will be inserted here */}
                            {children}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
    <script src="/static/js/alpine.min.js" defer></script>
    {scriptPaths && scriptPaths.split(',').map((path, index) => (
    <script key={'gen-script-'+index} src={path.trim()}></script>
))}
</body>
</html>
</>
    );
}
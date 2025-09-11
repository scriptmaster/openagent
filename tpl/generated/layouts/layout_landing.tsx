export default function LayoutLanding({page, children, linkTags, scriptTags}: {page: any, children: any, linkTags?: string, scriptTags?: string}) {
    return (
<>
<html lang="en">
<head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>{page.PageTitle || page.AppName}</title>
    <link rel="stylesheet" href="/static/css/tabler.min.css" />
    <link rel="stylesheet" href="/static/css/tabler-icons.min.css" />
    {linkTags}
</head>
<body className="bg-gradient-primary">
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
    {scriptTags}
</body>
</html>
</>
    );
}
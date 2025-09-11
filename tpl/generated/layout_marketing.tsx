export default function LayoutMarketing({page}: {page: Page}) {
    return (
<>
<meta charset="UTF-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0" />
<title>{page.PageTitle || page.AppName}</title>
<link rel="stylesheet" href="/static/css/tabler.min.css" />
<link rel="stylesheet" href="/static/css/tabler-icons.min.css" />
<div className="page">
    <div className="page-wrapper">
        <div className="page-body">
            <div className="container-xl">
                <div className="row row-deck row-cards">
                    <div className="col-12">
                        {/* Marketing content */}
                        <div className="text-center py-5">
                            <h1 className="display-1 fw-bold">{page.AppName}</h1>
                            <p className="lead">Your AI-powered development platform</p>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</div>
<script src="/tsx/js/_common.js"></script>
<script src="/tsx/js/layout_marketing.js"></script>
</>
    );
}
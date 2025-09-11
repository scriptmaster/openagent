export default function LayoutLoggedin({page}: {page: Page}) {
    return (
<meta charset="UTF-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0" />
<title>{page.PageTitle || page.AppName}</title>
<link rel="stylesheet" href="/static/css/tabler.min.css" />
<link rel="stylesheet" href="/static/css/tabler-icons.min.css" />
<div className="page">
    <div className="page-wrapper">
        <div className="navbar navbar-expand-md navbar-light d-print-none">
            <div className="container-xl">
                <button className="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbar-menu">
                    <span className="navbar-toggler-icon"></span>
                </button>
                <div className="navbar-nav flex-row order-md-last">
                    <div className="nav-item dropdown">
                        <a href="#" className="nav-link d-flex lh-1 text-reset p-0" data-bs-toggle="dropdown">
                            <span className="avatar avatar-sm" style={{backgroundImage: `url(https://ui-avatars.com/api/?name=${page.User.Email}&background=206bc4&color=fff)`}></span>
                            <div className="d-none d-xl-block ps-2">
                                <div>{page.User.Email}</div>
                            </div>
                        </a>
                    </div>
                </div>
            </div>
        </div>
        <div className="page-body">
            <div className="container-xl">
                <div className="row row-deck row-cards">
                    <div className="col-12">
                        {/* Logged in user content */}
                    </div>
                </div>
            </div>
        </div>
    </div>
</div>
<script src="/tsx/js/layout_loggedin.js"></script>
    );
}
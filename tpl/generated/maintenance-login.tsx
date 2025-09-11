export default function MaintenanceLogin({page}: {page: Page}) {
    return (
<>
<html lang="en">
<head>
    <meta charset="UTF-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
    <title>Maintenance Mode - OpenAgent</title>
    {/* Tabler CSS */}
    <link rel="stylesheet" href="/static/css/tabler.min.css"/>
    {/* Tabler Icons */}
    <link rel="stylesheet" href="/static/css/tabler-icons.min.css"/>
    {/* Alpine.js */}
    
    <link rel="stylesheet" href="/tsx/css/maintenance-login.css" />
</head>
<body data-x-data="{ showAuth: false }">
    <div className="page page-center">
        <div className="container container-tight py-2">
            <div className="text-center mb-3">
                <h1 className="navbar-brand navbar-brand-autodark">
                    <img src="/static/img/logo.svg" width="130" alt="OpenAgent"/>
                </h1>
                <h2 className="h3 text-muted">OpenAgent</h2>
                <p className="text-muted">Maintenance Access</p>
            </div>
            
            <div className="maintenance-banner">
                <i className="ti ti-alert-triangle me-2"></i> SERVER IS CURRENTLY IN MAINTENANCE MODE
            </div>
            <div className="card card-md">
                <div className="card-body">
                    {/* Trending GIF */}
                    <div className="text-center mb-3">
                        <img src="https://media.tenor.com/images/36788ddef89ef20a91ebebc14c2454ad/tenor.gif" className="trending-gif" alt="Funny Tech GIF"/>
                    </div>
                    
                    <div className="alert alert-warning mb-3">
                        <i className="ti ti-alert-triangle me-2"></i> The server is temporarily unavailable due to maintenance.
                        <p className="mt-2 mb-0">We're working to improve the system and will be back online shortly.</p>
                    </div>
                    
                    {/* Authentication Form - Hidden Initially */}
                    <div data-x-show="showAuth" x-transition>
                        {page.Error && (
                        <div className="alert alert-danger" role="alert">
                            <i className="ti ti-alert-circle me-2"></i> {page.Error}
                        </div>
                        )}
                        <h2 className="card-title text-center mb-4">Maintenance Authentication</h2>
                        <form action="/maintenance/auth" method="post">
                            <div className="mb-3">
                                <label className="form-label">Maintenance Token</label>
                                <div className="input-group input-group-flat">
                                    <span className="input-group-text">
                                        <i className="ti ti-key"></i>
                                    </span>
                                    <input type="password" className="form-control" name="token" placeholder="Enter maintenance token" required/>
                                </div>
                            </div>
                            
                            <div className="form-footer">
                                <button type="submit" className="btn btn-primary w-100">
                                    Access Maintenance Panel
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            </div>
            
            <div className="text-center text-muted mt-3">
                <p>Need help? Contact your <a href="mailto:{page.AdminEmail}">system administrator</a>.</p>
                <p className="restricted-link" data-click="showAuth = true">Restricted: Maintenance Administration</p>
                <p className="text-muted small">Version {page.AppVersion}</p>
            </div>
        </div>
    </div>
    
<script src="/tsx/js/_common.js"></script>
<script src="/tsx/js/maintenance-login.js"></script>
</body>
</html>
</>
    );
}
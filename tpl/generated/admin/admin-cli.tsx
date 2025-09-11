export default function AdminCli({page}: {page: Page}) {
    return (
<>
<html lang="en"><head>
    <meta charset="UTF-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Admin CLI</title>
    <link rel="stylesheet" href="/static/css/tabler.min.css" />
    <link rel="stylesheet" href="/static/css/tabler-icons.min.css" />
    
</head>
<body>
    <div className="page" data-x-data="adminCLI()">
        
        <header className="navbar navbar-expand-md navbar-light d-print-none">
            <div className="container-xl">
                <button className="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbar-menu">
                    <span className="navbar-toggler-icon">
                </button>
                <h1 className="navbar-brand navbar-brand-autodark d-none-navbar-horizontal pe-0 pe-md-3">
                    <a href=".">
                        <img src="/static/img/logo.svg" width="110" height="32" alt={page.AppName} className="navbar-brand-image"/>
                    </a>
                </h1>
                <div className="navbar-nav flex-row order-md-last">
                    <div className="nav-item dropdown">
                        <a href="#" className="nav-link d-flex lh-1 text-reset p-0" data-bs-toggle="dropdown" aria-label="Open user menu">
                            <div className="d-none d-xl-block ps-2">
                                <div>{page.User.Email}</div>
                                <div className="mt-1 small text-muted">Administrator</div>
                            </div>
                            <span className="avatar avatar-sm bg-blue-lt ms-2">
                                <i className="ti ti-user">
                            </span>
                        </a>
                        <div className="dropdown-menu dropdown-menu-end dropdown-menu-arrow">
                            <a href="/profile" className="dropdown-item">Profile</a>
                            <div className="dropdown-divider">
                            <a href="/logout" className="dropdown-item">Logout</a>
                        </div>
                    </div>
                </div>
            </div>
        </header>
        
        
        <div className="navbar-expand-md">
            <div className="collapse navbar-collapse" id="navbar-menu">
                <div className="navbar navbar-light">
                    <div className="container-xl">
                        <div className="d-flex flex-column flex-md-row flex-fill align-items-stretch align-items-md-center">
                            <ul className="navbar-nav">
                            <li className="nav-item">
                                <a className="nav-link" href="/admin">
                                    <span className="nav-link-icon d-md-none d-lg-inline-block">
                                        <i className="ti ti-dashboard">
                                    </span>
                                    <span className="nav-link-title">Dashboard</span>
                                </a>
                            </li>
                            <li className="nav-item">
                                <a className="nav-link" href="/projects">
                                    <span className="nav-link-icon d-md-none d-lg-inline-block">
                                        <i className="ti ti-folder">
                                    </span>
                                    <span className="nav-link-title">Projects</span>
                                </a>
                            </li>
                            <li className="nav-item">
                                <a className="nav-link" href="/admin/connections">
                                    <span className="nav-link-icon d-md-none d-lg-inline-block">
                                        <i className="ti ti-database">
                                    </span>
                                    <span className="nav-link-title">Database Connections</span>
                                </a>
                            </li>
                            <li className="nav-item">
                                <a className="nav-link" href="/admin/tables">
                                    <span className="nav-link-icon d-md-none d-lg-inline-block">
                                        <i className="ti ti-table">
                                    </span>
                                    <span className="nav-link-title">Tables</span>
                                </a>
                            </li>
                            <li className="nav-item">
                                <a className="nav-link" href="/users">
                                    <span className="nav-link-icon d-md-none d-lg-inline-block">
                                        <i className="ti ti-users">
                                    </span>
                                    <span className="nav-link-title">Users</span>
                                </a>
                            </li>
                            <li className="nav-item active">
                                <a className="nav-link" href="/admin/cli">
                                    <span className="nav-link-icon d-md-none d-lg-inline-block">
                                        <i className="ti ti-terminal">
                                    </span>
                                    <span className="nav-link-title">CLI</span>
                                </a>
                            </li>
                            <li className="nav-item">
                                <a className="nav-link" href="/admin/settings">
                                    <span className="nav-link-icon d-md-none d-lg-inline-block">
                                        <i className="ti ti-settings">
                                    </span>
                                    <span className="nav-link-title">Settings</span>
                                </a>
                            </li>
                        </ul>
                        </div>
                    </div>
                </div>
            </div>
        </div>
        
        <div className="page-header d-print-none">
            <div className="container-xl">
                <div className="row g-2 align-items-center">
                    <div className="col">
                        <h2 className="page-title">
                            <i className="ti ti-terminal me-2">Admin CLI
                        </h2>
                        <div className="text-muted mt-1">Execute SQL queries through a web interface</div>
                    </div>
                </div>
            </div>
        </div>
        
        <div className="page-body">
            <div className="container-xl">
                <div className="row">
                    
                    <div className="col-md-4 col-lg-3">
                        <div className="card">
                            <div className="card-header">
                                <h3 className="card-title">
                                    <i className="ti ti-list me-2">Available Queries
                                </h3>
                            </div>
                            <div className="card-body p-0">
                                <div className="list-group list-group-flush" style="max-height: 600px; overflow-y: auto;">
                                    <template data-x-for="(queries, paramCount) in queryGroups" data-key="paramCount">
                                        <div>
                                            <div className="list-group-item bg-light fw-bold">
                                                <span data-x-text="paramCount == 0 ? '0 parameters' : paramCount + ' parameters'">
                                            </div>
                                            <template data-x-for="query in queries" data-key="query.name">
                                                <a href="#" className="list-group-item list-group-item-action" data-className="{ 'active': selectedQuery && selectedQuery.name === query.name }" data-click="selectQuery(query)">
                                                    <div className="fw-bold" data-x-text="query.name">
                                                    <div className="text-muted small" data-x-text="query.paramDetails || 'No parameters'">
                                                </a>
                                            </template>
                                        </div>
                                    </template>
                                </div>
                            </div>
                        </div>
                    </div>
                    
                    <div className="col-md-8 col-lg-9">
                        <div className="card">
                            <div className="card-header">
                                <h3 className="card-title">
                                    <i className="ti ti-play me-2">Query Execution
                                </h3>
                            </div>
                            <div className="card-body">
                                
                                <div data-x-show="selectedQuery" className="mb-4">
                                    <label className="form-label">Selected Query:</label>
                                    <div className="form-control-plaintext bg-light p-3 rounded" style="font-family: monospace; font-size: 0.9em; white-space: pre-wrap;" data-x-text="selectedQuery ? selectedQuery.query : ''">
                                </div>
                                
                                <div data-x-show="selectedQuery && selectedQuery.paramCount > 0" className="mb-4">
                                    <label className="form-label">Parameters:</label>
                                    <template data-x-for="(param, index) in selectedQuery.paramDetails.split(', ')" data-key="index">
                                        <div className="mb-3">
                                            <label className="form-label" data-x-text="param">
                                            <input type="text" className="form-control" data-placeholder="'Enter ' + param" data-x-model="queryParams[index]"/>
                                        </div>
                                    </template>
                                </div>
                                
                                <div className="mb-4">
                                    <button type="button" className="btn btn-primary" data-disabled="!selectedQuery || loading" data-click="executeQuery">
                                        <span data-x-show="!loading">
                                            <i className="ti ti-play me-2">Run Query
                                        </span>
                                        <span data-x-show="loading">
                                            <span className="spinner-border spinner-border-sm me-2" role="status">
                                            Executing...
                                        </span>
                                    </button>
                                </div>
                                
                                <div data-x-show="result" className="mt-4">
                                    <div className="alert" data-className="result.success ? 'alert-success' : 'alert-danger'">
                                        <div className="d-flex align-items-center">
                                            <i className="ti" data-className="result.success ? 'ti-check' : 'ti-alert-circle'" me-2="">
                                            <div>
                                                <div data-x-show="result.success">
                                                    <strong>Query executed successfully!</strong>
                                                    <div className="text-muted small">
                                                        <span data-x-text="result.rowCount"> rows returned in <span data-x-text="result.duration">
                                                    </div>
                                                </div>
                                                <div data-x-show="!result.success">
                                                    <strong>Error:</strong> <span data-x-text="result.error">
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                    
                                    <div data-x-show="result.success && result.columns && result.columns.length > 0" className="table-responsive">
                                        <table className="table table-striped table-hover">
                                            <thead>
                                                <tr>
                                                    <template data-x-for="column in result.columns" data-key="column">
                                                        <th data-x-text="column">
                                                    </template>
                                                </tr>
                                            </thead>
                                            <tbody>
                                                <template data-x-for="(row, rowIndex) in result.rows" data-key="rowIndex">
                                                    <tr>
                                                        <template data-x-for="(cell, cellIndex) in row" data-key="cellIndex">
                                                            <td data-x-text="cell === null ? 'NULL' : cell">
                                                        </template>
                                                    </tr>
                                                </template>
                                            </tbody>
                                        </table>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
    

<script src="/tsx/js/_common.js"></script>
<script src="/tsx/js/admin-cli.js"></script>
</body></html>
</>
    );
}
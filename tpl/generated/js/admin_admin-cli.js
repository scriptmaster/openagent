
// ╔══ ⚛️  MAIN COMPONENT JS (TSX → JS) ⚛️ ══
function Admin-Cli({page}) {
    return (
React.createElement('div', {className: "page", data-alpine-data: "adminCLI()"}, React.createElement('header', {className: "navbar navbar-expand-md navbar-light d-print-none"}, React.createElement('div', {className: "container-xl"}, React.createElement('button', {className: "navbar-toggler", type: "button", data-bs-toggle: "collapse", data-bs-target: "#navbar-menu"}, React.createElement('span', {className: "navbar-toggler-icon"})), React.createElement('h1', {className: "navbar-brand navbar-brand-autodark d-none-navbar-horizontal pe-0 pe-md-3"}, React.createElement('a', {href: "."}, React.createElement('img', {src: "/static/img/logo.svg", width: "110", height: "32", alt: page.AppName, className: "navbar-brand-image"}))), React.createElement('div', {className: "navbar-nav flex-row order-md-last"}, React.createElement('div', {className: "nav-item dropdown"}, React.createElement('a', {href: "#", className: "nav-link d-flex lh-1 text-reset p-0", data-bs-toggle: "dropdown", aria-label: "Open user menu"}, React.createElement('div', {className: "d-none d-xl-block ps-2"}, React.createElement('div', null, '{page.User.Email}'), React.createElement('div', {className: "mt-1 small text-muted"}, 'Administrator')), React.createElement('span', {className: "avatar avatar-sm bg-blue-lt ms-2"}, React.createElement('i', {className: "ti ti-user"})), React.createElement('i', {className: "ti ti-user"})), React.createElement('i', {className: "ti ti-user"}, React.createElement('div', {className: "dropdown-menu dropdown-menu-end dropdown-menu-arrow"}, React.createElement('a', {href: "/profile", className: "dropdown-item"}, 'Profile'), React.createElement('div', {className: "dropdown-divider"}, React.createElement('a', {href: "/logout", className: "dropdown-item"}, 'Logout'))))), React.createElement('i', {className: "ti ti-user"})), React.createElement('i', {className: "ti ti-user"}))), React.createElement('i', {className: "ti ti-user"}, React.createElement('div', {className: "navbar-expand-md"}, React.createElement('div', {className: "collapse navbar-collapse", id: "navbar-menu"}, React.createElement('div', {className: "navbar navbar-light"}, React.createElement('div', {className: "container-xl"}, React.createElement('div', {className: "d-flex flex-column flex-md-row flex-fill align-items-stretch align-items-md-center"}, React.createElement('ul', {className: "navbar-nav"}, React.createElement('li', {className: "nav-item"}, React.createElement('a', {className: "nav-link", href: "/admin"}, React.createElement('span', {className: "nav-link-icon d-md-none d-lg-inline-block"}, React.createElement('i', {className: "ti ti-dashboard"})), React.createElement('i', {className: "ti ti-dashboard"}, React.createElement('span', {className: "nav-link-title"}, 'Dashboard'))), React.createElement('i', {className: "ti ti-dashboard"})), React.createElement('i', {className: "ti ti-dashboard"}, React.createElement('li', {className: "nav-item"}, React.createElement('a', {className: "nav-link", href: "/projects"}, React.createElement('span', {className: "nav-link-icon d-md-none d-lg-inline-block"}, React.createElement('i', {className: "ti ti-folder"})), React.createElement('i', {className: "ti ti-folder"}, React.createElement('span', {className: "nav-link-title"}, 'Projects'))), React.createElement('i', {className: "ti ti-folder"})), React.createElement('i', {className: "ti ti-folder"}, React.createElement('li', {className: "nav-item"}, React.createElement('a', {className: "nav-link", href: "/admin/connections"}, React.createElement('span', {className: "nav-link-icon d-md-none d-lg-inline-block"}, React.createElement('i', {className: "ti ti-database"})), React.createElement('i', {className: "ti ti-database"}, React.createElement('span', {className: "nav-link-title"}, 'Database Connections'))), React.createElement('i', {className: "ti ti-database"})), React.createElement('i', {className: "ti ti-database"}, React.createElement('li', {className: "nav-item"}, React.createElement('a', {className: "nav-link", href: "/admin/tables"}, React.createElement('span', {className: "nav-link-icon d-md-none d-lg-inline-block"}, React.createElement('i', {className: "ti ti-table"})), React.createElement('i', {className: "ti ti-table"}, React.createElement('span', {className: "nav-link-title"}, 'Tables'))), React.createElement('i', {className: "ti ti-table"})), React.createElement('i', {className: "ti ti-table"}, React.createElement('li', {className: "nav-item"}, React.createElement('a', {className: "nav-link", href: "/users"}, React.createElement('span', {className: "nav-link-icon d-md-none d-lg-inline-block"}, React.createElement('i', {className: "ti ti-users"})), React.createElement('i', {className: "ti ti-users"}, React.createElement('span', {className: "nav-link-title"}, 'Users'))), React.createElement('i', {className: "ti ti-users"})), React.createElement('i', {className: "ti ti-users"}, React.createElement('li', {className: "nav-item active"}, React.createElement('a', {className: "nav-link", href: "/admin/cli"}, React.createElement('span', {className: "nav-link-icon d-md-none d-lg-inline-block"}, React.createElement('i', {className: "ti ti-terminal"})), React.createElement('i', {className: "ti ti-terminal"}, React.createElement('span', {className: "nav-link-title"}, 'CLI'))), React.createElement('i', {className: "ti ti-terminal"})), React.createElement('i', {className: "ti ti-terminal"}, React.createElement('li', {className: "nav-item"}, React.createElement('a', {className: "nav-link", href: "/admin/settings"}, React.createElement('span', {className: "nav-link-icon d-md-none d-lg-inline-block"}, React.createElement('i', {className: "ti ti-settings"})), React.createElement('i', {className: "ti ti-settings"}, React.createElement('span', {className: "nav-link-title"}, 'Settings'))), React.createElement('i', {className: "ti ti-settings"})), React.createElement('i', {className: "ti ti-settings"})))))))), React.createElement('i', {className: "ti ti-dashboard"}, React.createElement('i', {className: "ti ti-folder"}, React.createElement('i', {className: "ti ti-database"}, React.createElement('i', {className: "ti ti-table"}, React.createElement('i', {className: "ti ti-users"}, React.createElement('i', {className: "ti ti-terminal"}, React.createElement('i', {className: "ti ti-settings"})))))))), React.createElement('i', {className: "ti ti-dashboard"}, React.createElement('i', {className: "ti ti-folder"}, React.createElement('i', {className: "ti ti-database"}, React.createElement('i', {className: "ti ti-table"}, React.createElement('i', {className: "ti ti-users"}, React.createElement('i', {className: "ti ti-terminal"}, React.createElement('i', {className: "ti ti-settings"})))))))), React.createElement('i', {className: "ti ti-dashboard"}, React.createElement('i', {className: "ti ti-folder"}, React.createElement('i', {className: "ti ti-database"}, React.createElement('i', {className: "ti ti-table"}, React.createElement('i', {className: "ti ti-users"}, React.createElement('i', {className: "ti ti-terminal"}, React.createElement('i', {className: "ti ti-settings"})))))))), React.createElement('i', {className: "ti ti-dashboard"}, React.createElement('i', {className: "ti ti-folder"}, React.createElement('i', {className: "ti ti-database"}, React.createElement('i', {className: "ti ti-table"}, React.createElement('i', {className: "ti ti-users"}, React.createElement('i', {className: "ti ti-terminal"}, React.createElement('i', {className: "ti ti-settings"})))))))), React.createElement('i', {className: "ti ti-dashboard"}, React.createElement('i', {className: "ti ti-folder"}, React.createElement('i', {className: "ti ti-database"}, React.createElement('i', {className: "ti ti-table"}, React.createElement('i', {className: "ti ti-users"}, React.createElement('i', {className: "ti ti-terminal"}, React.createElement('i', {className: "ti ti-settings"})))))))), React.createElement('i', {className: "ti ti-dashboard"}, React.createElement('i', {className: "ti ti-folder"}, React.createElement('i', {className: "ti ti-database"}, React.createElement('i', {className: "ti ti-table"}, React.createElement('i', {className: "ti ti-users"}, React.createElement('i', {className: "ti ti-terminal"}, React.createElement('i', {className: "ti ti-settings"}, React.createElement('div', {className: "page-header d-print-none"}, React.createElement('div', {className: "container-xl"}, React.createElement('div', {className: "row g-2 align-items-center"}, React.createElement('div', {className: "col"}, React.createElement('h2', {className: "page-title"}, React.createElement('i', {className: "ti ti-terminal me-2"}, 'Admin CLI')), React.createElement('i', {className: "ti ti-terminal me-2"}, React.createElement('div', {className: "text-muted mt-1"}, 'Execute SQL queries through a web interface'))), React.createElement('i', {className: "ti ti-terminal me-2"})), React.createElement('i', {className: "ti ti-terminal me-2"})), React.createElement('i', {className: "ti ti-terminal me-2"})), React.createElement('i', {className: "ti ti-terminal me-2"}, React.createElement('div', {className: "page-body"}, React.createElement('div', {className: "container-xl"}, React.createElement('div', {className: "row"}, React.createElement('div', {className: "col-md-4 col-lg-3"}, React.createElement('div', {className: "card"}, React.createElement('div', {className: "card-header"}, React.createElement('h3', {className: "card-title"}, React.createElement('i', {className: "ti ti-list me-2"}, 'Available Queries')), React.createElement('i', {className: "ti ti-list me-2"})), React.createElement('i', {className: "ti ti-list me-2"}, React.createElement('div', {className: "card-body p-0"}, React.createElement('div', {className: "list-group list-group-flush", style: "max-height: 600px; overflow-y: auto;"}, React.createElement('template', {data-alpine-for: "(queries, paramCount) in queryGroups", :key: "paramCount"}, React.createElement('div', null, React.createElement('div', {className: "list-group-item bg-light fw-bold"}, React.createElement('span', {x-text: "paramCount == 0 ? '0 parameters' : paramCount + ' parameters'"})), React.createElement('template', {data-alpine-for: "query in queries", :key: "query.name"}, React.createElement('a', {href: "#", className: "list-group-item list-group-item-action", :className:  'active': selectedQuery && selectedQuery.name === query.name , @click: "selectQuery(query)"}), React.createElement('div', {className: "fw-bold", x-text: "query.name"}, React.createElement('a', {href: "#", className: "list-group-item list-group-item-action", :className:  'active': selectedQuery && selectedQuery.name === query.name , @click: "selectQuery(query)"}), React.createElement('div', {className: "text-muted small", x-text: "query.paramDetails || 'No parameters'"}, React.createElement('a', {href: "#", className: "list-group-item list-group-item-action", :className:  'active': selectedQuery && selectedQuery.name === query.name , @click: "selectQuery(query)"})))))))))), React.createElement('i', {className: "ti ti-list me-2"})), React.createElement('i', {className: "ti ti-list me-2"}, React.createElement('div', {className: "col-md-8 col-lg-9"}, React.createElement('div', {className: "card"}, React.createElement('div', {className: "card-header"}, React.createElement('h3', {className: "card-title"}, React.createElement('i', {className: "ti ti-play me-2"}, 'Query Execution')), React.createElement('i', {className: "ti ti-play me-2"})), React.createElement('i', {className: "ti ti-play me-2"}, React.createElement('div', {className: "card-body"}, React.createElement('div', {data-alpine-show: "selectedQuery", className: "mb-4"}, React.createElement('label', {className: "form-label"}, 'Selected Query:'), React.createElement('div', {className: "form-control-plaintext bg-light p-3 rounded", style: "font-family: monospace; font-size: 0.9em; white-space: pre-wrap;", x-text: "selectedQuery ? selectedQuery.query : ''"}), React.createElement('div', {data-alpine-show: "selectedQuery && selectedQuery.paramCount > 0", className: "mb-4"}, React.createElement('label', {className: "form-label"}, 'Parameters:'), React.createElement('template', {data-alpine-for: "(param, index) in selectedQuery.paramDetails.split(', ')", :key: "index"}, React.createElement('div', {className: "mb-3"}, React.createElement('label', {className: "form-label", x-text: "param"}, React.createElement('input', {type: "text", className: "form-control", :placeholder: "'Enter ' + param", data-alpine-model: "queryParams[index]"}))))), React.createElement('div', {className: "mb-4"}, React.createElement('button', {type: "button", className: "btn btn-primary", :disabled: "!selectedQuery || loading", @click: "executeQuery"}, React.createElement('span', {data-alpine-show: "!loading"}, React.createElement('i', {className: "ti ti-play me-2"}, 'Run Query')), React.createElement('i', {className: "ti ti-play me-2"}, React.createElement('span', {data-alpine-show: "loading"}, React.createElement('span', {className: "spinner-border spinner-border-sm me-2", role: "status"}, 'Executing...')))), React.createElement('i', {className: "ti ti-play me-2"})), React.createElement('i', {className: "ti ti-play me-2"}, React.createElement('div', {data-alpine-show: "result", className: "mt-4"}, React.createElement('div', {className: "alert", :className: "result.success ? 'alert-success' : 'alert-danger'"}, React.createElement('div', {className: "d-flex align-items-center"}, React.createElement('i', {className: "ti", :className: "result.success ? 'ti-check' : 'ti-alert-circle'", me-2: ""}, React.createElement('div', null, React.createElement('div', {data-alpine-show: "result.success"}, React.createElement('strong', null, 'Query executed successfully!'), React.createElement('div', {className: "text-muted small"}, React.createElement('span', {x-text: "result.rowCount"}, 'rows returned in', React.createElement('span', {x-text: "result.duration"})))), React.createElement('div', {data-alpine-show: "!result.success"}, React.createElement('strong', null, 'Error:'), React.createElement('span', {x-text: "result.error"}))))), React.createElement('i', {className: "ti", :className: "result.success ? 'ti-check' : 'ti-alert-circle'", me-2: ""})), React.createElement('i', {className: "ti", :className: "result.success ? 'ti-check' : 'ti-alert-circle'", me-2: ""}, React.createElement('div', {data-alpine-show: "result.success && result.columns && result.columns.length > 0", className: "table-responsive"}, React.createElement('table', {className: "table table-striped table-hover"}, React.createElement('thead', null, React.createElement('tr', null, React.createElement('template', {data-alpine-for: "column in result.columns", :key: "column"}, React.createElement('th', {x-text: "column"})))), React.createElement('tbody', null, React.createElement('template', {data-alpine-for: "(row, rowIndex) in result.rows", :key: "rowIndex"}, React.createElement('tr', null, React.createElement('template', {data-alpine-for: "(cell, cellIndex) in row", :key: "cellIndex"}, React.createElement('td', {x-text: "cell === null ? 'NULL' : cell"})))))))))))))))))))))))))))))
    );
}



// ╔══ 📜 ORIGINAL JS CONTENT 📜 ══


    
    <script>
        function adminCLI() {
            return {
                queryGroups: {},
                selectedQuery: null,
                queryParams: [],
                result: null,
                loading: false,

                async init() {
                    await this.loadQueries();
                },

                async loadQueries() {
                    try {
                        const response = await fetch('/admin/cli/api/queries');
                        if (response.ok) {
                            this.queryGroups = await response.json();
                        } else {
                            console.error('Failed to load queries');
                        }
                    } catch (error) {
                        console.error('Error loading queries:', error);
                    }
                },

                selectQuery(query) {
                    this.selectedQuery = query;
                    this.queryParams = new Array(query.paramCount).fill('');
                    this.result = null;
                },

                async executeQuery() {
                    if (!this.selectedQuery) return;

                    this.loading = true;
                    this.result = null;

                    try {
                        const params = this.queryParams.filter(p => p.trim() !== '');
                        
                        const response = await fetch('/admin/cli/api/execute', {
                            method: 'POST',
                            headers: {
                                'Content-Type': 'application/json',
                            },
                            body: JSON.stringify({
                                queryName: this.selectedQuery.name,
                                params: params
                            })
                        });

                        if (response.ok) {
                            this.result = await response.json();
                        } else {
                            this.result = {
                                success: false,
                                error: 'Failed to execute query'
                            };
                        }
                    } catch (error) {
                        this.result = {
                            success: false,
                            error: 'Network error: ' + error.message
                        };
                    } finally {
                        this.loading = false;
                    }
                }
            }
        }
    


// ╔══ 💧 HYDRATION 💧 ══

// Make component available globally for hydration
window.Admin-Cli = Admin-Cli;

// React hydration using common utilities
try {
    // Use the global hydration function from _common.js
    window.hydrateReactApp('Admin-Cli', { 
        page: window.pageData || {},
        container: 'main',
		layout: React.createElement('div', {}, 'Layout placeholder')
    });
} catch(e) {
    console.error('React hydration error:', e);
}
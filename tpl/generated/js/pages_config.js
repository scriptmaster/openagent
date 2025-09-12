
// Embedded Component JS
///////////////////////////////

// Component: simple
function Simple() {
    return (
        React.createElement('div', null, 'Simple Component')
    );
}

///////////////////////////////

// Component prototype methods


///////////////////////////////

// Main Page JS

// Main Component JS (converted to JS from TSX)
function Config({page}) {
    return (
React.createElement('div', {className: 'page'}, React.createElement('div', {className: 'container container-tight py-4'}, React.createElement('h1', {className: 'text-center mb-4'}, 'System Configuration'), React.createElement('div', {className: 'card card-md', data-x-data: 'configForm()'}, React.createElement('div', {className: 'card-body'}, React.createElement('div', {data-x-show: 'message', data-className: 'isError ? 'alert alert-danger' : 'alert alert-success'', data-x-text: 'message'}), React.createElement('form', {data-submit-prevent: 'submitConfig'}, React.createElement('h3', {className: 'mb-3'}, 'Admin Setup Token'), React.createElement('p', {className: 'text-muted'}, 'Enter the setup token displayed in the server console for today.'), React.createElement('div', {className: 'mb-3'}, React.createElement('label', {className: 'form-label required'}, 'Today's Admin Token'), React.createElement('input', {type: 'text', className: 'form-control', data-x-model: 'formData.admin_token', required: ''})), React.createElement('hr', {className: 'my-4'}), React.createElement('h3', {className: 'mb-3'}, 'Initial Project'), React.createElement('div', {className: 'mb-3'}, React.createElement('label', {className: 'form-label required'}, 'Project Name'), React.createElement('input', {type: 'text', className: 'form-control', data-x-model: 'formData.project_name', required: ''})), React.createElement('div', {className: 'mb-3'}, React.createElement('label', {className: 'form-label'}, 'Project Description'), React.createElement('textarea', {className: 'form-control', data-x-model: 'formData.project_desc', rows: '3'})), React.createElement('div', {className: 'mb-3'}, React.createElement('label', {className: 'form-label required'}, 'Primary Host'), React.createElement('input', {type: 'text', className: 'form-control', data-x-model: 'formData.primary_host', placeholder: 'e.g., myapp.com or localhost', required: ''}), React.createElement('small', {className: 'form-hint'}, 'The main domain name for this project (without http:// or port).')), React.createElement('div', {className: 'mb-3'}, React.createElement('label', {className: 'form-label'}, 'Redirect Hosts'), React.createElement('small', {className: 'form-hint d-block mb-2'}, 'Requests to these hosts will be redirected to the Primary Host.'), React.createElement('ul', {className: 'host-input-list'}, React.createElement('template', {data-x-for: '(host, index) in formData.redirect_hosts', data-key: 'index'}, React.createElement('li', null, React.createElement('input', {type: 'text', className: 'form-control form-control-sm', data-x-model: 'formData.redirect_hosts[index]', placeholder: 'e.g., www.myapp.com'}), React.createElement('button', {type: 'button', className: 'btn btn-sm btn-icon btn-outline-danger', data-click: 'removeHost('redirect', index)', aria-label: 'Remove host'}, React.createElement('i', {className: 'ti ti-x'}))))), React.createElement('button', {type: 'button', className: 'btn btn-sm btn-outline-secondary mt-2', data-click: 'addHost('redirect')', data-x-show: 'formData.redirect_hosts.length < 20'}, React.createElement('i', {className: 'ti ti-plus me-1'}), 'Add Redirect Host')), React.createElement('div', {className: 'mb-3'}, React.createElement('label', {className: 'form-label'}, 'Alias Hosts'), React.createElement('small', {className: 'form-hint d-block mb-2'}, 'Requests to these hosts will serve the same content as the Primary Host.'), React.createElement('ul', {className: 'host-input-list'}, React.createElement('template', {data-x-for: '(host, index) in formData.alias_hosts', data-key: 'index'}, React.createElement('li', null, React.createElement('input', {type: 'text', className: 'form-control form-control-sm', data-x-model: 'formData.alias_hosts[index]', placeholder: 'e.g., oldapp.com'}), React.createElement('button', {type: 'button', className: 'btn btn-sm btn-icon btn-outline-danger', data-click: 'removeHost('alias', index)', aria-label: 'Remove host'}, React.createElement('i', {className: 'ti ti-x'}))))), React.createElement('button', {type: 'button', className: 'btn btn-sm btn-outline-secondary mt-2', data-click: 'addHost('alias')', data-x-show: 'formData.alias_hosts.length < 20'}, React.createElement('i', {className: 'ti ti-plus me-1'}), 'Add Alias Host')), React.createElement('hr', {className: 'my-4'}), React.createElement('h3', {className: 'mb-3'}, 'Initial Admin User'), React.createElement('div', {className: 'mb-3'}, React.createElement('label', {className: 'form-label required'}, 'Admin Email'), React.createElement('input', {type: 'email', className: 'form-control', data-x-model: 'formData.admin_email', required: ''})), React.createElement('div', {className: 'mb-3'}, React.createElement('label', {className: 'form-label required'}, 'Admin Password'), React.createElement('input', {type: 'password', className: 'form-control', data-x-model: 'formData.admin_password', required: ''})), React.createElement('div', {className: 'form-footer mt-4'}, React.createElement('button', {type: 'submit', className: 'btn btn-primary w-100', data-disabled: 'loading'}, React.createElement('span', {data-x-show: 'loading', className: 'spinner-border spinner-border-sm me-2', role: 'status'}), React.createElement('span', {data-x-text: 'loading ? 'Configuring...' : 'Complete Setup''}))))))))
    );
}

///////////////////////////////

// Original JS content

    document.addEventListener('alpine:init', () => {
        Alpine.data('configForm', () => ({
            loading: false,
            message: '',
            isError: false,
            formData: {
                admin_token: '',
                project_name: '',
                project_desc: '',
                primary_host: '{page.CurrentHost}' || '', // Pre-fill if possible
                admin_email: '',
                admin_password: '',
                redirect_hosts: [''], // Start with one empty input
                alias_hosts: ['']     // Start with one empty input
            },
            addHost(type) {
                if (type === 'redirect' && this.formData.redirect_hosts.length < 20) {
                    this.formData.redirect_hosts.push('');
                } else if (type === 'alias' && this.formData.alias_hosts.length < 20) {
                    this.formData.alias_hosts.push('');
                }
            },
            removeHost(type, index) {
                if (type === 'redirect') {
                     // Prevent removing the last input if it's the only one
                    if (this.formData.redirect_hosts.length > 1) {
                         this.formData.redirect_hosts.splice(index, 1);
                    } else if (this.formData.redirect_hosts.length === 1) {
                            this.formData.redirect_hosts[0] = ''; // Clear the last one instead
                    }
                } else if (type === 'alias') {
                    if (this.formData.alias_hosts.length > 1) {
                        this.formData.alias_hosts.splice(index, 1);
                    } else if (this.formData.alias_hosts.length === 1) {
                             this.formData.alias_hosts[0] = ''; // Clear the last one
                    }
                }
            },
            async submitConfig() {
                this.loading = true;
                this.message = '';
                this.isError = false;

                // Filter out empty host strings before submitting
                const payload = {
                    ...this.formData,
                    redirect_hosts: this.formData.redirect_hosts.filter(h => h.trim() !== ''),
                    alias_hosts: this.formData.alias_hosts.filter(h => h.trim() !== '')
                };

                try {
                    const response = await fetch('/config/submit', { // Define submit endpoint
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify(payload)
                    });
                    const result = await response.json();

                    if (response.ok && result.success !== false) { // Check for explicit false success
                        this.message = result.message || 'Configuration successful! Redirecting...';
                        this.isError = false;
                        // Redirect on success, maybe to login?
                        setTimeout(() => window.location.href = '/login', 1500);
                    } else {
                        this.message = result.message || 'Configuration failed. Please check the form and server logs.';
                        this.isError = true;
                    }
                } catch (err) {
                    console.error("Config Submit Error:", err);
                    this.message = 'An unexpected error occurred during submission.';
                    this.isError = true;
                } finally {
                    this.loading = false;
                }
            }
        }));
    });



///////////////////////////////

// Make component available globally for hydration
window.Config = Config;

// React hydration using common utilities
try {
    // Use the global hydration function from _common.js
    window.hydrateReactApp('Config', { 
        page: window.pageData || {},
        container: 'main',
		layout: React.createElement('div', {}, 'Layout placeholder')
    });
} catch(e) {
    console.error('React hydration error:', e);
}
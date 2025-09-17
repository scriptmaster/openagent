export default function Config({page}: {page: any}) {
    return (
<div className="page">
    <div className="container container-tight py-4">
        <h1 className="text-center mb-4">System Configuration</h1>
        <div className="card card-md" data-alpine-data="configForm()">
            <div className="card-body">
                <div data-alpine-show="message" :className="isError ? 'alert alert-danger' : 'alert alert-success'" x-text="message"></div>
                <form @submit-prevent="submitConfig">
                    <h3 className="mb-3">Admin Setup Token</h3>
                    <p className="text-muted">Enter the setup token displayed in the server console for today.</p>
                    <div className="mb-3">
                        <label className="form-label required">Today's Admin Token</label>
                        <input type="text" className="form-control" data-alpine-model="formData.admin_token" required>
                    </div>
                    <hr className="my-4">
                    <h3 className="mb-3">Initial Project</h3>
                    <div className="mb-3">
                        <label className="form-label required">Project Name</label>
                        <input type="text" className="form-control" data-alpine-model="formData.project_name" required>
                    </div>
                    <div className="mb-3">
                        <label className="form-label">Project Description</label>
                        <textarea className="form-control" data-alpine-model="formData.project_desc" rows="3"></textarea>
                    </div>
                    <div className="mb-3">
                        <label className="form-label required">Primary Host</label>
                        <input type="text" className="form-control" data-alpine-model="formData.primary_host" placeholder="e.g., myapp.com or localhost" required>
                        <small className="form-hint">The main domain name for this project (without http:// or port).</small>
                    </div>
                    
                    <div className="mb-3">
                        <label className="form-label">Redirect Hosts</label>
                        <small className="form-hint d-block mb-2">Requests to these hosts will be redirected to the Primary Host.</small>
                        <ul className="host-input-list">
                            <template data-alpine-for="(host, index) in formData.redirect_hosts" :key="index">
                                <li>
                                    <input type="text" className="form-control form-control-sm" data-alpine-model="formData.redirect_hosts[index]" placeholder="e.g., www.myapp.com">
                                    <button type="button" className="btn btn-sm btn-icon btn-outline-danger" @click="removeHost('redirect', index)" aria-label="Remove host"><i className="ti ti-x"></i></button>
                                </li>
                            </template>
                        </ul>
                        <button type="button" className="btn btn-sm btn-outline-secondary mt-2" @click="addHost('redirect')" data-alpine-show="formData.redirect_hosts.length < 20">
                            <i className="ti ti-plus me-1"></i> Add Redirect Host
                        </button>
                    </div>
                    <div className="mb-3">
                        <label className="form-label">Alias Hosts</label>
                        <small className="form-hint d-block mb-2">Requests to these hosts will serve the same content as the Primary Host.</small>
                        <ul className="host-input-list">
                            <template data-alpine-for="(host, index) in formData.alias_hosts" :key="index">
                                <li>
                                    <input type="text" className="form-control form-control-sm" data-alpine-model="formData.alias_hosts[index]" placeholder="e.g., oldapp.com">
                                    <button type="button" className="btn btn-sm btn-icon btn-outline-danger" @click="removeHost('alias', index)" aria-label="Remove host"><i className="ti ti-x"></i></button>
                                </li>
                            </template>
                        </ul>
                        <button type="button" className="btn btn-sm btn-outline-secondary mt-2" @click="addHost('alias')" data-alpine-show="formData.alias_hosts.length < 20">
                            <i className="ti ti-plus me-1"></i> Add Alias Host
                        </button>
                    </div>
                    <hr className="my-4">
                    <h3 className="mb-3">Initial Admin User</h3>
                    <div className="mb-3">
                        <label className="form-label required">Admin Email</label>
                        <input type="email" className="form-control" data-alpine-model="formData.admin_email" required>
                    </div>
                    <div className="mb-3">
                        <label className="form-label required">Admin Password</label>
                        <input type="password" className="form-control" data-alpine-model="formData.admin_password" required>
                    </div>
                    <div className="form-footer mt-4">
                        <button type="submit" className="btn btn-primary w-100" :disabled="loading">
                            <span data-alpine-show="loading" className="spinner-border spinner-border-sm me-2" role="status"></span>
                            <span x-text="loading ? 'Configuring...' : 'Complete Setup'"></span>
                        </button>
                    </div>
                </form>
            </div>
        </div>
    </div>
</div>
    );
}
export default function Users({children}: {children: any}) {
    return (
<html><head><body>{{define "users.page"}}{{template "layout_marketing" .}}{{end}}

{{define "title"}}User Management - {{.AppName}}{{end}}

{{define "content"}}
<div className="page-body">
    <div className="container-xl">
        <div className="row">
            <div className="col-12">
                <div className="card">
                    <div className="card-header">
                        <h3 className="card-title">User Management</h3>
                    </div>
                    <div className="card-body">
                        <p className="text-muted">User management interface coming soon...</p>
                        <div className="alert alert-info">
                            <h4>Admin Access Required</h4>
                            <p>This page is only accessible to administrators.</p>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</div>
{{end}}
</body></html>
    );
}
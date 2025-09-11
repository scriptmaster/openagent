export default function Tables({children}: {children: any}) {
    return (
<html><head><body>{{/* Execute the layout template, passing the current data context */}}
{{template "layout.html" .}}

{{define "title"}}Database Tables - {{.AppName}}{{end}}

{{/* Define page title text (layout handles structure) */}}
{{define "page-title"}}Database Tables{{end}}

{{define "content"}}
<div className="page-body">
    <div className="container-xl">
        <div className="row">
            <div className="col-12">
                <div className="card">
                    <div className="card-header">
                        <h3 className="card-title">Database Tables</h3>
                    </div>
                    <div className="card-body">
                        <p className="text-muted">Database table interface coming soon...</p>
                        <div className="alert alert-info">
                            <h4>User Access</h4>
                            <p>This page is accessible to all authenticated users.</p>
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
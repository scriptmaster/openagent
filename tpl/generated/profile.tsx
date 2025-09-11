export default function Profile({children}: {children: any}) {
    return (
<html><head><body>{{/* Execute the layout template, passing the current data context */}}
{{template "layout.html" .}}

{{define "title"}}User Profile - {{.AppName}}{{end}}

{{/* Define page title text (layout handles structure) */}}
{{define "page-title"}}User Profile{{end}}

{{define "content"}}
<div className="page-body">
    <div className="container-xl">
        <div className="row">
            <div className="col-12">
                <div className="card">
                    <div className="card-header">
                        <h3 className="card-title">User Profile</h3>
                    </div>
                    <div className="card-body">
                        {{if .User}}
                        <div className="row">
                            <div className="col-md-6">
                                <h4>Profile Information</h4>
                                <table className="table table-borderless">
                                    <tbody><tr>
                                        <td><strong>Email:</strong></td>
                                        <td>{{.User.Email}}</td>
                                    </tr>
                                    <tr>
                                        <td><strong>Role:</strong></td>
                                        <td>
                                            {{if .User.IsAdmin}}
                                                <span className="badge bg-danger">Administrator</span>
                                            {{else}}
                                                <span className="badge bg-primary">User</span>
                                            {{end}}
                                        </td>
                                    </tr>
                                    <tr>
                                        <td><strong>User ID:</strong></td>
                                        <td>{{.User.ID}}</td>
                                    </tr>
                                </tbody></table>
                            </div>
                        </div>
                        {{else}}
                        <div className="alert alert-warning">
                            <h4>No User Information</h4>
                            <p>User information not available.</p>
                        </div>
                        {{end}}
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
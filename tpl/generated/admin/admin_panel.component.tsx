export default function Admin_panel({page}: {page: any}) {
    return (
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{page.PageTitle} - Admin Panel</title>
    <link rel="stylesheet" href="/static/css/tabler.min.css">
</head>
<body>
    <div className="container">
        <h1>{page.AdminTitle}</h1>
        <p>This is an admin panel with specific admin data.</p>
        <p>Admin Level: {page.AdminLevel}</p>
        <p>Permissions: {page.Permissions}</p>
    </div>
</body>
</html>
    );
}
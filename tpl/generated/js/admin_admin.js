
// ╔══ ⚛️  MAIN COMPONENT JS (TSX → JS) ⚛️ ══
function Admin({page}) {
    return (
React.createElement('div', {className: "page"}, React.createElement('div', {className: "page-wrapper"}, React.createElement('div', {className: "navbar navbar-expand-md navbar-light d-print-none"}, React.createElement('div', {className: "container-xl"}, React.createElement('div', {className: "navbar-nav flex-row order-md-last"}, React.createElement('div', {className: "nav-item dropdown"}, React.createElement('a', {href: "#", className: "nav-link d-flex lh-1 text-reset p-0", data-bs-toggle: "dropdown"}, React.createElement('span', {className: "avatar avatar-sm", style: "{{backgroundImage:", `url(https:: "", ui-avatars.com: "", api: "", ?name: "${page.User.Email}&background=dc3545&color=fff)`}"}), React.createElement('div', {className: "d-none d-xl-block ps-2"}, React.createElement('div', null, '{page.User.Email}'))))))), React.createElement('div', {className: "page-header"}, React.createElement('div', {className: "container-xl"}, React.createElement('div', {className: "row g-2 align-items-center"}, React.createElement('div', {className: "col"}, React.createElement('div', {className: "page-pretitle"}, 'Administration'), React.createElement('h2', {className: "page-title"}, 'Admin Dashboard'))))), React.createElement('div', {className: "page-body"}, React.createElement('div', {className: "container-xl"}, React.createElement('div', {className: "row row-deck row-cards"}, React.createElement('div', {className: "col-12"}, React.createElement('div', {className: "row row-cards"}, React.createElement('div', {className: "col-sm-6 col-lg-3"}, React.createElement('div', {className: "card card-sm"}, React.createElement('div', {className: "card-body"}, React.createElement('div', {className: "row align-items-center"}, React.createElement('div', {className: "col-auto"}, React.createElement('span', {className: "bg-primary text-white avatar"}, React.createElement('i', {className: "ti ti-folder"}))), React.createElement('div', {className: "col"}, React.createElement('div', {className: "font-weight-medium"}, 'Projects'), React.createElement('div', {className: "h1 mb-3"}, '{page.Stats.ProjectCount}')))))), React.createElement('div', {className: "col-sm-6 col-lg-3"}, React.createElement('div', {className: "card card-sm"}, React.createElement('div', {className: "card-body"}, React.createElement('div', {className: "row align-items-center"}, React.createElement('div', {className: "col-auto"}, React.createElement('span', {className: "bg-green text-white avatar"}, React.createElement('i', {className: "ti ti-database"}))), React.createElement('div', {className: "col"}, React.createElement('div', {className: "font-weight-medium"}, 'Connections'), React.createElement('div', {className: "h1 mb-3"}, '{page.Stats.ConnectionCount}')))))), React.createElement('div', {className: "col-sm-6 col-lg-3"}, React.createElement('div', {className: "card card-sm"}, React.createElement('div', {className: "card-body"}, React.createElement('div', {className: "row align-items-center"}, React.createElement('div', {className: "col-auto"}, React.createElement('span', {className: "bg-blue text-white avatar"}, React.createElement('i', {className: "ti ti-table"}))), React.createElement('div', {className: "col"}, React.createElement('div', {className: "font-weight-medium"}, 'Tables'), React.createElement('div', {className: "h1 mb-3"}, '{page.Stats.TableCount}')))))), React.createElement('div', {className: "col-sm-6 col-lg-3"}, React.createElement('div', {className: "card card-sm"}, React.createElement('div', {className: "card-body"}, React.createElement('div', {className: "row align-items-center"}, React.createElement('div', {className: "col-auto"}, React.createElement('span', {className: "bg-orange text-white avatar"}, React.createElement('i', {className: "ti ti-users"}))), React.createElement('div', {className: "col"}, React.createElement('div', {className: "font-weight-medium"}, 'Users'), React.createElement('div', {className: "h1 mb-3"}, '{page.Stats.UserCount}'))))))))))), React.createElement('footer', {className: "footer footer-transparent d-print-none"}, React.createElement('div', {className: "container-xl"}, React.createElement('div', {className: "row text-center align-items-center flex-row-reverse"}, React.createElement('div', {className: "col-12 col-lg-auto mt-3 mt-lg-0"}, React.createElement('ul', {className: "list-inline list-inline-dots mb-0"}, React.createElement('li', {className: "list-inline-item"}, '© {page.CurrentYear} {page.AppName}'))))))))
    );
}



// ╔══ 📜 ORIGINAL JS CONTENT 📜 ══



// ╔══ 💧 HYDRATION 💧 ══

// Make component available globally for hydration
window.Admin = Admin;

// React hydration using common utilities
try {
    // Use the global hydration function from _common.js
    window.hydrateReactApp('Admin', { 
        page: window.pageData || {},
        container: 'main',
		layout: React.createElement('div', {}, 'Layout placeholder')
    });
} catch(e) {
    console.error('React hydration error:', e);
}
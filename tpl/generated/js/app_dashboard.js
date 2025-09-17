
// ╔══ ⚛️  MAIN COMPONENT JS (TSX → JS) ⚛️ ══
function Dashboard({page}) {
    return (
React.createElement('div', {className: "page-header"}, React.createElement('div', {className: "container-xl"}, React.createElement('div', {className: "row g-2 align-items-center"}, React.createElement('div', {className: "col"}, React.createElement('div', {className: "page-pretitle"}, 'Overview'), React.createElement('h2', {className: "page-title"}, 'Dashboard')))))React.createElement('div', {className: "page-body"}, React.createElement('div', {className: "container-xl"}, React.createElement('div', {className: "row row-deck row-cards"}, React.createElement('div', {className: "col-12"}, React.createElement('div', {className: "row row-cards"}, React.createElement('div', {className: "col-sm-6 col-lg-3"}, React.createElement('div', {className: "card card-sm"}, React.createElement('div', {className: "card-body"}, React.createElement('div', {className: "row align-items-center"}, React.createElement('div', {className: "col-auto"}, React.createElement('span', {className: "bg-primary text-white avatar"}, React.createElement('i', {className: "ti ti-folder"}))), React.createElement('div', {className: "col"}, React.createElement('div', {className: "font-weight-medium"}, 'Projects'), React.createElement('div', {className: "text-muted"}, '{page.ProjectCount}')))))))))))
    );
}



// ╔══ 📜 ORIGINAL JS CONTENT 📜 ══



// ╔══ 💧 HYDRATION 💧 ══

// Make component available globally for hydration
window.Dashboard = Dashboard;

// React hydration using common utilities
try {
    // Use the global hydration function from _common.js
    window.hydrateReactApp('Dashboard', { 
        page: window.pageData || {},
        container: 'main',
		layout: React.createElement('div', {}, 'Layout placeholder')
    });
} catch(e) {
    console.error('React hydration error:', e);
}

// ╔══ ⚛️  MAIN COMPONENT JS (TSX → JS) ⚛️ ══
function Projects({page}) {
    return (
React.createElement('div', {data-alpine-data: "projectsApp()", x-init: "loadProjects()"}, React.createElement('div', {className: "page-header"}, React.createElement('div', {className: "container-xl"}, React.createElement('div', {className: "row g-2 align-items-center"}, React.createElement('div', {className: "col"}, React.createElement('div', {className: "page-pretitle"}, 'Project Management'), React.createElement('h2', {className: "page-title"}, 'Projects')), React.createElement('div', {className: "col-auto ms-auto d-print-none"}, React.createElement('div', {className: "btn-list"}, React.createElement('button', {className: "btn btn-primary d-none d-sm-inline-block", @click: "showCreateModal = true"}, React.createElement('i', {className: "ti ti-plus"}), 'Create New Project'), React.createElement('button', {className: "btn btn-primary d-sm-none btn-icon", @click: "showCreateModal = true"}, React.createElement('i', {className: "ti ti-plus"}))))))), React.createElement('div', {className: "page-body"}, React.createElement('div', {className: "container-xl"}, React.createElement('div', {className: "row row-deck row-cards"}, React.createElement('div', {className: "col-12"}, React.createElement('div', {className: "card"}, React.createElement('div', {className: "card-header"}, React.createElement('h3', {className: "card-title"}, 'Your Projects')), React.createElement('div', {className: "card-body"}, React.createElement('div', {data-alpine-show: "loading", className: "text-center py-4"}, React.createElement('div', {className: "spinner-border", role: "status"}, React.createElement('span', {className: "visually-hidden"}, 'Loading...'))), React.createElement('div', {data-alpine-show: "!loading && projects.length === 0", className: "text-center py-4"}, React.createElement('div', {className: "empty"}, React.createElement('div', {className: "empty-icon"}, React.createElement('i', {className: "ti ti-folder-off"})), React.createElement('p', {className: "empty-title"}, 'No projects found'), React.createElement('p', {className: "empty-subtitle text-muted"}, 'Get started by creating your first project.'), React.createElement('div', {className: "empty-action"}, React.createElement('button', {className: "btn btn-primary", @click: "showCreateModal = true"}, React.createElement('i', {className: "ti ti-plus"}), 'Create Project')))), React.createElement('div', {data-alpine-show: "!loading && projects.length > 0", className: "row"}, React.createElement('template', {data-alpine-for: "project in projects", :key: "project.id"}, React.createElement('div', {className: "col-md-6 col-lg-4 mb-3"}, React.createElement('div', {className: "card card-sm"}, React.createElement('div', {className: "card-body"}, React.createElement('div', {className: "row align-items-center"}, React.createElement('div', {className: "col-auto"}, React.createElement('span', {className: "bg-primary text-white avatar"}, React.createElement('i', {className: "ti ti-folder"}))), React.createElement('div', {className: "col"}, React.createElement('div', {className: "font-weight-medium"}, React.createElement('span', {x-text: "project.name"})), React.createElement('div', {className: "text-muted"}, React.createElement('span', {x-text: "project.domain_name"}))))), React.createElement('div', {className: "card-footer"}, React.createElement('div', {className: "row align-items-center"}, React.createElement('div', {className: "col"}, React.createElement('div', {className: "text-muted"}, React.createElement('i', {className: "ti ti-database me-1"}), React.createElement('span', {x-text: "project.connection_count"}), 'connections')), React.createElement('div', {className: "col-auto"}, React.createElement('a', {:href: "`/projects/${project.id}`", className: "btn btn-sm btn-primary"}, 'View')))))))))))))))
    );
}



// ╔══ 📜 ORIGINAL JS CONTENT 📜 ══



// ╔══ 💧 HYDRATION 💧 ══

// Make component available globally for hydration
window.Projects = Projects;

// React hydration using common utilities
try {
    // Use the global hydration function from _common.js
    window.hydrateReactApp('Projects', { 
        page: window.pageData || {},
        container: 'main',
		layout: React.createElement('div', {}, 'Layout placeholder')
    });
} catch(e) {
    console.error('React hydration error:', e);
}
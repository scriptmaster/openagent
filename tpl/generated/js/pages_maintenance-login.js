
// Component JS (converted to JS from TSX)
function MaintenanceLogin({page}) {
    return (
React.createElement('div', {className: 'page page-center', data-x-data: '{ showAuth: false }'}, React.createElement('div', {className: 'container container-tight py-2'}, React.createElement('div', {className: 'text-center mb-3'}, React.createElement('h1', {className: 'navbar-brand navbar-brand-autodark'}, React.createElement('img', {src: '/static/img/logo.svg', width: '130', alt: 'OpenAgent'})), React.createElement('h2', {className: 'h3 text-muted'}, 'OpenAgent'), React.createElement('p', {className: 'text-muted'}, 'Maintenance Access')), React.createElement('div', {className: 'maintenance-banner'}, React.createElement('i', {className: 'ti ti-alert-triangle me-2'}), 'SERVER IS CURRENTLY IN MAINTENANCE MODE'), React.createElement('div', {className: 'card card-md'}, React.createElement('div', {className: 'card-body'}, '{/* Trending GIF */}', React.createElement('div', {className: 'text-center mb-3'}, React.createElement('img', {src: 'https://media.tenor.com/images/36788ddef89ef20a91ebebc14c2454ad/tenor.gif', className: 'trending-gif', alt: 'Funny Tech GIF'})), React.createElement('div', {className: 'alert alert-warning mb-3'}, React.createElement('i', {className: 'ti ti-alert-triangle me-2'}), 'The server is temporarily unavailable due to maintenance.', React.createElement('p', {className: 'mt-2 mb-0'}, 'We're working to improve the system and will be back online shortly.')), '{/* Authentication Form - Hidden Initially */}', React.createElement('div', {data-x-show: 'showAuth', x-transition: ''}, '{page.Error && (', React.createElement('div', {className: 'alert alert-danger', role: 'alert'}, React.createElement('i', {className: 'ti ti-alert-circle me-2'}), '{page.Error}'), ')}', React.createElement('h2', {className: 'card-title text-center mb-4'}, 'Maintenance Authentication'), React.createElement('form', {action: '/maintenance/auth', method: 'post'}, React.createElement('div', {className: 'mb-3'}, React.createElement('label', {className: 'form-label'}, 'Maintenance Token'), React.createElement('div', {className: 'input-group input-group-flat'}, React.createElement('span', {className: 'input-group-text'}, React.createElement('i', {className: 'ti ti-key'})), React.createElement('input', {type: 'password', className: 'form-control', name: 'token', placeholder: 'Enter maintenance token', required: ''}))), React.createElement('div', {className: 'form-footer'}, React.createElement('button', {type: 'submit', className: 'btn btn-primary w-100'}, 'Access Maintenance Panel')))))), React.createElement('div', {className: 'text-center text-muted mt-3'}, React.createElement('p', null, 'Need help? Contact your', React.createElement('a', {href: 'mailto:{page.AdminEmail}'}, 'system administrator'), '.'), React.createElement('p', {className: 'restricted-link', data-click: 'showAuth = true'}, 'Restricted: Maintenance Administration'), React.createElement('p', {className: 'text-muted small'}, 'Version {page.AppVersion}'))))
    );
}

///////////////////////////////

// Original JS content

    // Load a random tech GIF from Tenor
    document.addEventListener('DOMContentLoaded', function() {
        // Tenor GIF API
        fetch('https://g.tenor.com/v1/random?q=tech+funny&key=LIVDSRZULELA&limit=1')
            .then(response => response.json())
            .then(data => {
                if (data.results && data.results.length > 0) {
                    document.querySelector('.trending-gif').src = data.results[0].media[0].gif.url;
                }
            })
            .catch(error => {
                console.error('Error fetching GIF:', error);
                // Keep the default GIF if there's an error
            });
    });



///////////////////////////////

// React hydration using common utilities
try {
    // Use the global hydration function from _common.js
    window.hydrateReactApp('MaintenanceLogin', { 
        page: window.pageData || {},
        container: 'main',
		layout: React.createElement('div', {}, 'Layout placeholder')
    });
} catch(e) {
    console.error('React hydration error:', e);
}
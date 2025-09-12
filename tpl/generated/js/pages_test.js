
// Component JS (converted to JS from TSX)
function Test({page}) {
    return (
React.createElement('div', {className: 'container-xl'}, React.createElement('div', {className: 'row'}, React.createElement('div', {className: 'col-12'}, React.createElement('div', {className: 'page-header'}, React.createElement('h1', {className: 'page-title'}, 'React Tests - Interactive Components! (No Imports)'), React.createElement('p', {className: 'text-muted'}, 'Testing various React functionality')))))
    );
}

///////////////////////////////

// Original JS content

        console.log('please wait...');
    


///////////////////////////////

// React hydration using common utilities
try {
    // Use the global hydration function from _common.js
    window.hydrateReactApp('Test', { 
        page: window.pageData || {},
        container: 'main',
		layout: React.createElement('div', {}, 'Layout placeholder')
    });
} catch(e) {
    console.error('React hydration error:', e);
}
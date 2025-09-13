// React imports - loaded once for all pages
// import React from '/static/js/react.production.min.js';
// import { hydrateRoot } from '/static/js/react-dom.production.min.js';

//#include /static/js/react.production.min.js
//#-include /static/js/react-dom.production.min.js
//#include /static/js/react-dom.development.js

// Common React utilities and functions
window.React = React;
window.hydrateRoot = React.hydrateRoot;

// Common React component utilities
window.createReactElement = React.createElement;
window.useState = React.useState;
window.useEffect = React.useEffect;

// Global error handler for React hydration
window.addEventListener('error', function(e) {
    if (e.message && e.message.includes('React')) {
        console.error('React Error:', e.message, e.filename, e.lineno);
    }
});

// Common Layout component
window.Layout = function({page, children, linkPaths, scriptPaths}) {
    return React.createElement('div', { className: 'page' },
        React.createElement('div', { className: 'page-wrapper' },
            children
        )
    );
};

// Common React hydration function
window.hydrateReactApp = function(componentName, props = {}) {
    try {
        const container = document.querySelector('main');
        if (container && window[componentName]) {
            const Component = window[componentName];
            ReactDOM.hydrateRoot(container, React.createElement(Component, props));
            console.log(`React app ${componentName} hydrated successfully on main tag`);
        } else {
            console.warn(`React hydration failed: main container or component ${componentName} not found`);
        }
    } catch(e) {
        console.error('React hydration error:', e);
    }
};

console.log('React common utilities loaded');

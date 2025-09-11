//#include "../../../static/js/alpine.min.js"

// Alpine.js plugin for data attribute support
document.addEventListener('alpine:init', () => {
    // Set Alpine prefix to 'data'
    Alpine.prefix('data');
    
    // Plugin to handle data-submit-prevent and data-click-prevent
    Alpine.plugin((Alpine) => {
        // Handle form submission with prevent default
        Alpine.directive('submit-prevent', (el, { expression }, { Alpine, effect, cleanup }) => {
            const handler = (e) => {
                e.preventDefault();
                if (expression) {
                    Alpine.evaluateLater(el, expression)(result => {
                        if (typeof result === 'function') {
                            result(e);
                        }
                    });
                }
            };
            
            el.addEventListener('submit', handler);
            cleanup(() => el.removeEventListener('submit', handler));
        });
        
        // Handle click events with prevent default
        Alpine.directive('click-prevent', (el, { expression }, { Alpine, effect, cleanup }) => {
            const handler = (e) => {
                e.preventDefault();
                if (expression) {
                    Alpine.evaluateLater(el, expression)(result => {
                        if (typeof result === 'function') {
                            result(e);
                        }
                    });
                }
            };
            
            el.addEventListener('click', handler);
            cleanup(() => el.removeEventListener('click', handler));
        });
        
        // Handle regular click events
        Alpine.directive('click', (el, { expression }, { Alpine, effect, cleanup }) => {
            const handler = (e) => {
                if (expression) {
                    Alpine.evaluateLater(el, expression)(result => {
                        if (typeof result === 'function') {
                            result(e);
                        }
                    });
                }
            };
            
            el.addEventListener('click', handler);
            cleanup(() => el.removeEventListener('click', handler));
        });
        
        // Handle form submission
        Alpine.directive('submit', (el, { expression }, { Alpine, effect, cleanup }) => {
            const handler = (e) => {
                if (expression) {
                    Alpine.evaluateLater(el, expression)(result => {
                        if (typeof result === 'function') {
                            result(e);
                        }
                    });
                }
            };
            
            el.addEventListener('submit', handler);
            cleanup(() => el.removeEventListener('submit', handler));
        });
    });
});

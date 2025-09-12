// Alpine.js plugin for data attribute support
document.addEventListener('alpine:init', () => {
    console.log('🔧 Alpine.js initializing with custom directives...');
    
    // Set Alpine prefix to 'data'
    Alpine.prefix('data');
    console.log('🔧 Alpine prefix set to "data"');
    
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
        
        // Handle show/hide directive
        Alpine.directive('show', (el, { expression }, { Alpine, effect, cleanup }) => {
            console.log(`🔧 data-show directive registered for element:`, el, `expression: ${expression}`);
            effect(() => {
                Alpine.evaluateLater(el, expression)(result => {
                    console.log(`🔧 data-show evaluation: ${expression} = ${result}`);
                    el.style.display = result ? '' : 'none';
                });
            });
        });
        
        // Handle model directive (two-way binding)
        Alpine.directive('model', (el, { expression }, { Alpine, effect, cleanup }) => {
            effect(() => {
                Alpine.evaluateLater(el, expression)(result => {
                    if (el.type === 'checkbox') {
                        el.checked = !!result;
                    } else {
                        el.value = result || '';
                    }
                });
            });
            
            const handler = (e) => {
                let value = e.target.value;
                if (e.target.type === 'checkbox') {
                    value = e.target.checked;
                }
                
                // Find the Alpine component and update the data
                const component = Alpine.$data(el);
                if (component && component[expression] !== undefined) {
                    component[expression] = value;
                }
            };
            
            el.addEventListener('input', handler);
            cleanup(() => el.removeEventListener('input', handler));
        });
        
        // Handle disabled directive
        Alpine.directive('disabled', (el, { expression }, { Alpine, effect, cleanup }) => {
            effect(() => {
                Alpine.evaluateLater(el, expression)(result => {
                    el.disabled = !!result;
                });
            });
        });
        
        // Handle text directive
        Alpine.directive('text', (el, { expression }, { Alpine, effect, cleanup }) => {
            effect(() => {
                Alpine.evaluateLater(el, expression)(result => {
                    el.textContent = result || '';
                });
            });
        });
        
        // Handle class directive
        Alpine.directive('class', (el, { expression }, { Alpine, effect, cleanup }) => {
            console.log(`🔧 data-class directive registered for element:`, el, `expression: ${expression}`);
            effect(() => {
                Alpine.evaluateLater(el, expression)(result => {
                    console.log(`🔧 data-class evaluation: ${expression} = ${result}`);
                    if (typeof result === 'object') {
                        Object.keys(result).forEach(className => {
                            if (result[className]) {
                                el.classList.add(className);
                            } else {
                                el.classList.remove(className);
                            }
                        });
                    } else if (typeof result === 'string') {
                        el.className = result;
                    }
                });
            });
        });
        
        // Handle required directive
        Alpine.directive('required', (el, { expression }, { Alpine, effect, cleanup }) => {
            effect(() => {
                Alpine.evaluateLater(el, expression)(result => {
                    el.required = !!result;
                });
            });
        });
    });
});

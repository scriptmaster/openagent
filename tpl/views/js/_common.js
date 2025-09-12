//#include /static/js/hyperscript.min.js

_hyperscript.addCommand("post", function(parser, runtime, tokens) {
    var formToken = tokens.matchAny("identifier", "idRef", "classRef");
    if (!formToken) parser.raiseParseError(tokens, "Expected form element identifier (e.g., #myForm).");
    
    tokens.requireToken("to");
    
    var urlToken = tokens.matchAny("string", "identifier");
    if (!urlToken) parser.raiseParseError(tokens, "Expected URL after 'to'.");
    
    return function(sender, evt, args) {
      var form = runtime.find(formToken.value, sender);
      var url = urlToken.value;
      
      if (!form) throw new Error("Could not find form element: " + formToken.value);
      
      const formData = new FormData(form);
      const data = Object.fromEntries(formData.entries());

      return fetch(url, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(data),
      })
      .then(response => response.json());
    };
});

_hyperscript.addCommand("loadingButton", function(parser, runtime, tokens) {
    var buttonSelector = tokens.matchAny("identifier", "idRef", "classRef");
    if (!buttonSelector) parser.raiseParseError(tokens, "Expected button selector (e.g., .btn-primary).");
    
    var loadingToken = tokens.matchAny("boolean", "identifier");
    if (!loadingToken) parser.raiseParseError(tokens, "Expected loading state (true/false).");
    
    var textToken = tokens.matchAny("string", "identifier");
    if (!textToken) parser.raiseParseError(tokens, "Expected button text.");
    
    return function(sender, evt, args) {
      var button = runtime.find(buttonSelector.value, sender);
      var isLoading = loadingToken.value === 'true' || loadingToken.value === true;
      var text = textToken.value;
      
      if (!button) throw new Error("Could not find button element: " + buttonSelector.value);
      
      // Find spinner and text elements within the button
      var spinner = button.querySelector('.spinner-border') || button.querySelector('[role="status"]') || button.querySelector('span:first-of-type');
      var textElement = button.querySelector('.btn-text') || button.querySelector('span:last-of-type');
      
      if (isLoading) {
        button.disabled = true;
        if (spinner) spinner.style.display = 'inline-block';
        if (textElement) textElement.textContent = text;
      } else {
        button.disabled = false;
        if (spinner) spinner.style.display = 'none';
        if (textElement) textElement.textContent = text;
      }
    };
});

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


    document.addEventListener('alpine:init', () => {
        Alpine.data('configForm', () => ({
            loading: false,
            message: '',
            isError: false,
            formData: {
                admin_token: '',
                project_name: '',
                project_desc: '',
                primary_host: '{page.CurrentHost}' || '', // Pre-fill if possible
                admin_email: '',
                admin_password: '',
                redirect_hosts: [''], // Start with one empty input
                alias_hosts: ['']     // Start with one empty input
            },
            addHost(type) {
                if (type === 'redirect' && this.formData.redirect_hosts.length < 20) {
                    this.formData.redirect_hosts.push('');
                } else if (type === 'alias' && this.formData.alias_hosts.length < 20) {
                    this.formData.alias_hosts.push('');
                }
            },
            removeHost(type, index) {
                if (type === 'redirect') {
                     // Prevent removing the last input if it's the only one
                    if (this.formData.redirect_hosts.length > 1) {
                         this.formData.redirect_hosts.splice(index, 1);
                    } else if (this.formData.redirect_hosts.length === 1) {
                            this.formData.redirect_hosts[0] = ''; // Clear the last one instead
                    }
                } else if (type === 'alias') {
                    if (this.formData.alias_hosts.length > 1) {
                        this.formData.alias_hosts.splice(index, 1);
                    } else if (this.formData.alias_hosts.length === 1) {
                             this.formData.alias_hosts[0] = ''; // Clear the last one
                    }
                }
            },
            async submitConfig() {
                this.loading = true;
                this.message = '';
                this.isError = false;

                // Filter out empty host strings before submitting
                const payload = {
                    ...this.formData,
                    redirect_hosts: this.formData.redirect_hosts.filter(h => h.trim() !== ''),
                    alias_hosts: this.formData.alias_hosts.filter(h => h.trim() !== '')
                };

                try {
                    const response = await fetch('/config/submit', { // Define submit endpoint
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify(payload)
                    });
                    const result = await response.json();

                    if (response.ok && result.success !== false) { // Check for explicit false success
                        this.message = result.message || 'Configuration successful! Redirecting...';
                        this.isError = false;
                        // Redirect on success, maybe to login?
                        setTimeout(() => window.location.href = '/login', 1500);
                    } else {
                        this.message = result.message || 'Configuration failed. Please check the form and server logs.';
                        this.isError = true;
                    }
                } catch (err) {
                    console.error("Config Submit Error:", err);
                    this.message = 'An unexpected error occurred during submission.';
                    this.isError = true;
                } finally {
                    this.loading = false;
                }
            }
        }));
    });


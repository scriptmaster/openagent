

    
    <script>
        function adminCLI() {
            return {
                queryGroups: {},
                selectedQuery: null,
                queryParams: [],
                result: null,
                loading: false,

                async init() {
                    await this.loadQueries();
                },

                async loadQueries() {
                    try {
                        const response = await fetch('/admin/cli/api/queries');
                        if (response.ok) {
                            this.queryGroups = await response.json();
                        } else {
                            console.error('Failed to load queries');
                        }
                    } catch (error) {
                        console.error('Error loading queries:', error);
                    }
                },

                selectQuery(query) {
                    this.selectedQuery = query;
                    this.queryParams = new Array(query.paramCount).fill('');
                    this.result = null;
                },

                async executeQuery() {
                    if (!this.selectedQuery) return;

                    this.loading = true;
                    this.result = null;

                    try {
                        const params = this.queryParams.filter(p => p.trim() !== '');
                        
                        const response = await fetch('/admin/cli/api/execute', {
                            method: 'POST',
                            headers: {
                                'Content-Type': 'application/json',
                            },
                            body: JSON.stringify({
                                queryName: this.selectedQuery.name,
                                params: params
                            })
                        });

                        if (response.ok) {
                            this.result = await response.json();
                        } else {
                            this.result = {
                                success: false,
                                error: 'Failed to execute query'
                            };
                        }
                    } catch (error) {
                        this.result = {
                            success: false,
                            error: 'Network error: ' + error.message
                        };
                    } finally {
                        this.loading = false;
                    }
                }
            }
        }
    

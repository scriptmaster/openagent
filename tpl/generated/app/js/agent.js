

        function agentApp() {
            return {
                goalInput: '',
                promptInput: '',
                agentStarted: false,
                isLoading: false, // Indicates an active request to the backend (/start, /next)
                agentState: { status: 'Idle', history: [], iteration: 0, maxIterations: 20, goal: '', lastOutput: '', lastError: '' },
                pollingInterval: null,
                isPolling: false, // Indicates if a background status poll is active

                init() {
                    console.log('Agent UI initialized');
                    this.fetchStatus(); // Fetch status once on load in case server restarted
                    this.startPolling();
                },

                startPolling() {
                    if (this.pollingInterval) clearInterval(this.pollingInterval);
                    this.pollingInterval = setInterval(async () => {
                        if (this.agentStarted && !this.isLoading && !this.isPolling) {
                            this.isPolling = true;
                            try {
                                await this.fetchStatus();
                            } finally {
                                this.isPolling = false;
                            }
                        }
                    }, 3000); // Poll every 3 seconds
                },

                 stopPolling() {
                    if (this.pollingInterval) {
                        clearInterval(this.pollingInterval);
                        this.pollingInterval = null;
                         console.log("Polling stopped.");
                    }
                },

                async fetchStatus() {
                    try {
                        const response = await fetch('/status');
                        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
                        const data = await response.json();

                        // Only update if data received, prevent clearing state on transient errors
                        if (data && data.status) {
                             // Scroll history to bottom if new messages arrived
                            const historyChanged = JSON.stringify(this.agentState.history) !== JSON.stringify(data.history);
                            this.agentState = data;

                            if (historyChanged) {
                                this.$nextTick(() => {
                                    const historyDiv = this.$el.querySelector('.history');
                                    if (historyDiv) historyDiv.scrollTop = historyDiv.scrollHeight;
                                });
                            }

                            // Check if agent session exists on server
                            if (this.agentState.goal && !this.agentStarted) {
                                this.agentStarted = true;
                                this.goalInput = this.agentState.goal;
                                console.log("Detected existing agent session on server.");
                            } else if (!this.agentState.goal && this.agentStarted) {
                                // Agent was reset on server? Reset UI.
                                this.resetUI();
                                console.log("Agent session ended on server.");
                            }

                             // Stop polling if finished/error/blocked
                            if (this.agentState.status === 'Finished' || this.agentState.status === 'Error' || this.agentState.status === 'Command Blocked (Safety)') {
                               this.stopPolling();
                               this.isLoading = false; // Ensure loading indicator is off
                            }
                        }

                    } catch (error) {
                        console.error("Error fetching agent status:", error);
                        // Don't clear the whole state, maybe just show a temporary fetch error message
                        // this.agentState.lastError = `Failed to fetch status: ${error.message}`;
                    }
                },

                async startAgent() {
                    if (!this.goalInput.trim() || this.isLoading) return;
                    this.isLoading = true;
                    this.agentStarted = false;
                    this.resetUIState(); // Clear visual state
                    console.log("Starting agent with goal:", this.goalInput);
                    try {
                        const response = await fetch('/start', {
                            method: 'POST',
                            headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
                            body: new URLSearchParams({ 'goal': this.goalInput })
                        });
                        if (!response.ok) {
                             const errorText = await response.text();
                            throw new Error(`HTTP error! status: ${response.status} - ${errorText}`);
                        }
                        const data = await response.json();
                        this.agentState = data;
                        this.agentStarted = true;
                        console.log("Agent started successfully");
                        this.startPolling(); // Ensure polling is active
                        this.$nextTick(() => { // Scroll history after initial messages load
                             const historyDiv = this.$el.querySelector('.history');
                             if (historyDiv) historyDiv.scrollTop = historyDiv.scrollHeight;
                        });
                    } catch (error) {
                        console.error("Error starting agent:", error);
                        this.agentState.status = 'Error';
                        this.agentState.lastError = `Failed to start agent: ${error.message}`;
                         this.agentStarted = false;
                         this.stopPolling();
                    } finally {
                        this.isLoading = false;
                    }
                },

                async nextStep() {
                    if (!this.canProceed() || this.isLoading) return;
                    this.isLoading = true;
                    console.log("Triggering next step...");
                    try {
                        const response = await fetch('/next', { 
                            method: 'POST',
                            headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
                            body: new URLSearchParams({ 'prompt': this.promptInput })
                        });
                        if (!response.ok) {
                            const errorText = await response.text();
                            throw new Error(`HTTP error! status: ${response.status} - ${errorText}`);
                        }
                        const data = await response.json();
                        this.agentState = data;
                        this.promptInput = ''; // Clear prompt after sending
                        console.log("Next step triggered, agent status:", this.agentState.status);
                    } catch (error) {
                        console.error("Error triggering next step:", error);
                        this.agentState.status = 'Error';
                        this.agentState.lastError = `Failed to trigger next step: ${error.message}`;
                        this.stopPolling();
                    } finally {
                        this.isLoading = false;
                    }
                },

                 canProceed() {
                    const proceedStates = ['Awaiting Next Step'];
                    return this.agentStarted && proceedStates.includes(this.agentState.status);
                },

                resetUIState() {
                    this.agentState = { status: 'Initializing...', history: [], iteration: 0, maxIterations: 20, goal: '', lastOutput: '', lastError: '' };
                },
                resetUI() {
                    this.agentStarted = false;
                    this.goalInput = '';
                    this.resetUIState();
                    this.stopPolling();
                },
                async retryLastAction() {
                    if (this.isLoading) return;
                    this.isLoading = true;
                    try {
                        await new Promise(resolve => setTimeout(resolve, 2000));
                        
                        if (this.agentState.goal) {
                            // If we have a goal, retry starting the agent with the prompt
                            const response = await fetch('/start', {
                                method: 'POST',
                                headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
                                body: new URLSearchParams({ 
                                    'goal': this.agentState.goal,
                                    'prompt': this.promptInput 
                                })
                            });
                            if (!response.ok) {
                                const errorText = await response.text();
                                throw new Error(`HTTP error! status: ${response.status} - ${errorText}`);
                            }
                            const data = await response.json();
                            this.agentState = data;
                            this.promptInput = ''; // Clear prompt after sending
                            this.agentStarted = true;
                            this.startPolling();
                        } else {
                            // Otherwise retry the next step with the prompt
                            await this.nextStep();
                        }
                    } finally {
                        this.isLoading = false;
                    }
                }
            }
        }
    

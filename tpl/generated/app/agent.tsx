export default function Agent({page}: {page: any}) {
    return (
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Go Agent</title>
    
</head>
<body data-alpine-data="agentApp()" x-init="init()">
    <div className="container">
        <h1>Go Agent</h1>
        <div className="goal-input">
            <input type="text" data-alpine-model="goalInput" placeholder="Enter agent's goal..." :disabled="agentStarted">
            <button @click="startAgent()" :disabled="agentStarted || goalInput.trim() === '' || isLoading">Start Agent</button>
             <div data-alpine-show="isLoading && !agentStarted" className="loading"><div className="loader"></div></div>
        </div>
        <template data-alpine-if="agentStarted">
            <div className="status-bar">
                <div>Status:
                    <span className="status-text"
                          :className="{
                              'status-error': agentState.status === 'Error',
                              'status-blocked': agentState.status === 'Command Blocked (Safety)',
                              'status-finished': agentState.status === 'Finished'
                          }"
                          x-text="agentState.status || 'Initializing...'">
                    </span>
                     (<span x-text="agentState.iteration"></span>/<span x-text="agentState.maxIterations"></span> iterations)
                </div>
                <div className="action-button">
                    <div className="prompt-input">
                        <input type="text" data-alpine-model="promptInput" placeholder="Add a tip or prompt..." 
                               :disabled="!canProceed() || isLoading">
                    </div>
                    <button @click="nextStep()" :disabled="!canProceed() || isLoading">
                         Next Step
                    </button>
                    <button @click="retryLastAction()" 
                            data-alpine-show="agentState.status === 'Error' && (agentState.lastError.includes('model') || agentState.lastError.includes('timeout'))" 
                            className="retry-button" :disabled="isLoading">
                        Retry
                    </button>
                    <div data-alpine-show="isLoading && canProceed()" className="loading"><div className="loader"></div></div>
                </div>
            </div>
             <div data-alpine-show="agentState.lastError" className="error-box">
                <strong>Last Error:</strong> <span x-text="agentState.lastError"></span>
             </div>
        </template>
        <div className="history-container" data-alpine-show="agentStarted">
            <div className="history">
                <h2>Agent Log</h2>
                <template data-alpine-if="agentState.history && agentState.history.length > 0">
                     <template data-alpine-for="(msg, index) in agentState.history" :key="index">
                        <div className="message" :className="msg.role">
                            <strong>[<span x-text="msg.role.toUpperCase()"></span>] <span x-text="new Date(msg.timestamp).toLocaleString()"></span></strong>
                            <pre x-text="msg.content"></pre>
                        </div>
                    </template>
                </template>
                <template data-alpine-if="!agentState.history || agentState.history.length === 0">
                    <p>Agent log is empty.</p>
                </template>
            </div>
            <div data-alpine-show="agentState.lastOutput && (agentState.status == 'Finished' || agentState.status == 'Awaiting Next Step' || agentState.status == 'Blocked')" className="output-box">
                 <h2>Last Output / Final Answer</h2>
                 <pre x-text="agentState.lastOutput"></pre>
            </div>
        </div>
    </div>
    <div className="text-center text-muted mt-3">
        <p className="text-muted small">Version {page.AppVersion}</p>
    </div>
</body>
</html>
    );
}
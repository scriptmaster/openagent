export default function Agent({page}) {
    return (
<main>
<html lang="en">
<head>
    <meta charset="UTF-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
    <title>Go Agent</title>
    

    <link rel="stylesheet" href="/tsx/css/agent.css" />
</head>
<body data-x-data="agentApp()" x-init="init()">
    <div className="container">
        <h1>Go Agent</h1>
        <div className="goal-input">
            <input type="text" data-x-model="goalInput" placeholder="Enter agent's goal..." data-disabled="agentStarted"/>
            <button data-click="startAgent()" data-disabled="agentStarted || goalInput.trim() === '' || isLoading">Start Agent</button>
             <div data-x-show="isLoading && !agentStarted" className="loading"><div className="loader"></div></div>
        </div>
        <template data-x-if="agentStarted">
            <div className="status-bar">
                <div>Status:
                    <span className="status-text"
                          data-className="{
                              'status-error': agentState.status === 'Error',
                              'status-blocked': agentState.status === 'Command Blocked (Safety)',
                              'status-finished': agentState.status === 'Finished'
                          }"
                          data-x-text="agentState.status || 'Initializing...'">
                    </span>
                     (<span data-x-text="agentState.iteration"></span>/<span data-x-text="agentState.maxIterations"></span> iterations)
                </div>
                <div className="action-button">
                    <div className="prompt-input">
                        <input type="text" data-x-model="promptInput" placeholder="Add a tip or prompt..." 
                               data-disabled="!canProceed() || isLoading"/>
                    </div>
                    <button data-click="nextStep()" data-disabled="!canProceed() || isLoading">
                         Next Step
                    </button>
                    <button data-click="retryLastAction()" 
                            data-x-show="agentState.status === 'Error' && (agentState.lastError.includes('model') || agentState.lastError.includes('timeout'))" 
                            className="retry-button" data-disabled="isLoading">
                        Retry
                    </button>
                    <div data-x-show="isLoading && canProceed()" className="loading"><div className="loader"></div></div>
                </div>
            </div>
             <div data-x-show="agentState.lastError" className="error-box">
                <strong>Last Error:</strong> <span data-x-text="agentState.lastError"></span>
             </div>
        </template>
        <div className="history-container" data-x-show="agentStarted">
            <div className="history">
                <h2>Agent Log</h2>
                <template data-x-if="agentState.history && agentState.history.length > 0">
                     <template data-x-for="(msg, index) in agentState.history" data-key="index">
                        <div className="message" data-className="msg.role">
                            <strong>[<span data-x-text="msg.role.toUpperCase()"></span>] <span data-x-text="new Date(msg.timestamp).toLocaleString()"></span></strong>
                            <pre data-x-text="msg.content"></pre>
                        </div>
                    </template>
                </template>
                <template data-x-if="!agentState.history || agentState.history.length === 0">
                    <p>Agent log is empty.</p>
                </template>
            </div>
            <div data-x-show="agentState.lastOutput && (agentState.status == 'Finished' || agentState.status == 'Awaiting Next Step' || agentState.status == 'Blocked')" className="output-box">
                 <h2>Last Output / Final Answer</h2>
                 <pre data-x-text="agentState.lastOutput"></pre>
            </div>
        </div>
    </div>
    <div className="text-center text-muted mt-3">
        <p className="text-muted small">Version {page.AppVersion}</p>
    </div>

<script src="/tsx/js/_common.js"></script>
<script src="/tsx/js/app_agent.js"></script>
</body>
</html>
</main>
    );
}
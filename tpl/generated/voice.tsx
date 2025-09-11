export default function Voice({page}: {page: Page}) {
    return (
<>
<html lang="en">
<head>
    <meta charset="UTF-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
    <title>Voice Agent</title>
    {/* WebLLM Integration */}
     {/* Placeholder for the web worker script */}
     
    <link rel="stylesheet" href="/tsx/css/voice.css" />
</head>
<body>
    <div className="container">
        <div className="chat-panel">
            <h3>Chat</h3>
            <div id="chat-history">
                {/* Chat messages will appear here */}
                 <div className="agent-message">Hello! How can I help you today?</div>
            </div>
            <input type="text" id="chat-input" placeholder="Type or speak your message..."/>
        </div>
        <div className="main-content">
            <h2>Voice Input</h2>
            <div id="voice-meter-container">
                 <div id="voice-meter"></div>
            </div>
            <button id="record-button">Start Recording</button>
            <div id="status">Loading...</div>
        </div>
    </div>
<script src="/tsx/js/_common.js"></script>
<script src="/tsx/js/voice.js"></script>
</body>
</html>
</>
    );
}
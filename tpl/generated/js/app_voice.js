
// ╔══ ⚛️  MAIN COMPONENT JS (TSX → JS) ⚛️ ══
function Voice({page}) {
    return (
React.createElement('div', {className: "container"}, React.createElement('div', {className: "chat-panel"}, React.createElement('h3', null, 'Chat'), React.createElement('div', {id: "chat-history"}, React.createElement('div', {className: "agent-message"}, 'Hello! How can I help you today?')), React.createElement('input', {type: "text", id: "chat-input", placeholder: "Type or speak your message..."})), React.createElement('div', {className: "main-content"}, React.createElement('h2', null, 'Voice Input'), React.createElement('div', {id: "voice-meter-container"}, React.createElement('div', {id: "voice-meter"})), React.createElement('button', {id: "record-button"}, 'Start Recording'), React.createElement('div', {id: "status"}, 'Loading...')))
    );
}



// ╔══ 📜 ORIGINAL JS CONTENT 📜 ══

        import { CreateWebWorkerMLCEngine } from "https://esm.run/@mlc-ai/web-llm";

        const statusElement = document.getElementById("status");
        const chatHistory = document.getElementById("chat-history");
        const chatInput = document.getElementById("chat-input");

        // Use a small model, e.g., "Llama-3-8B-Instruct-q4f32_1-MLC-1k" or "Phi-3-mini-4k-instruct-q4f32_1-MLC"
        // See https://llm.mlc.ai/docs/prebuilt_models.html for options
        const SELECTED_MODEL = "Phi-3-mini-4k-instruct-q4f32_1-MLC";

        let engine;
        let recognition; // SpeechRecognition instance
        let isRecording = false;
        const recordButton = document.getElementById('record-button');
        const voiceMeter = document.getElementById('voice-meter');
        let audioContext;
        let analyser;
        let microphone;
        let javascriptNode;


        async function initializeWebLLM() {
            statusElement.textContent = "Loading WebLLM engine...";
            try {
                engine = await CreateWebWorkerMLCEngine(
                    new Worker(
                        new URL('/static/js/web-llm-worker.js', import.meta.url),
                        { type: 'module' }
                    ),
                    SELECTED_MODEL,
                    {
                        initProgressCallback: (report) => {
                             statusElement.textContent = `Loading model: ${report.text}`;
                             console.log(report);
                        }
                    }
                );
                 statusElement.textContent = "WebLLM Engine and Model Ready.";
            } catch (error) {
                 statusElement.textContent = `Error initializing WebLLM: ${error}`;
                console.error("Error initializing WebLLM:", error);
            }
        }

        function appendMessage(text, sender) {
            const messageDiv = document.createElement("div");
            messageDiv.textContent = text;
            messageDiv.classList.add(sender === 'user' ? 'user-message' : 'agent-message');
            chatHistory.appendChild(messageDiv);
            chatHistory.scrollTop = chatHistory.scrollHeight; // Scroll to bottom
        }

        async function handleChatInput(event) {
            if (event.key === 'Enter' && chatInput.value.trim() !== '') {
                const userInput = chatInput.value.trim();
                appendMessage(userInput, 'user');
                chatInput.value = ''; // Clear input
                await generateResponse(userInput);
            }
        }

        async function generateResponse(prompt) {
             if (!engine) {
                 appendMessage("Error: LLM Engine not ready.", 'agent');
                 return;
             }
             statusElement.textContent = "Agent thinking...";
             try {
                const chunks = await engine.chat.completions.create({
                    stream: true,
                    messages: [{ role: "user", content: prompt }],
                    // Consider adding temperature, max_gen_len etc. if needed
                });

                let reply = "";
                const agentMessageDiv = document.createElement("div");
                agentMessageDiv.classList.add('agent-message');
                chatHistory.appendChild(agentMessageDiv);

                for await (const chunk of chunks) {
                    const deltaContent = chunk.choices[0]?.delta?.content || "";
                    reply += deltaContent;
                    // Update incrementally for streaming effect
                     agentMessageDiv.textContent = reply;
                     chatHistory.scrollTop = chatHistory.scrollHeight;
                }
                statusElement.textContent = "Ready.";
                await engine.runtimeStatsText(); // Optional: log stats

             } catch (error) {
                appendMessage(`Error generating response: ${error}`, 'agent');
                statusElement.textContent = "Error.";
                console.error("LLM Generation Error:", error);
             }
        }

        // --- Voice Recognition & Meter ---

        function setupAudioMeter() {
            audioContext = new (window.AudioContext || window.webkitAudioContext)();
            analyser = audioContext.createAnalyser();
            analyser.fftSize = 256; // Smaller FFT size for volume detection is fine
            javascriptNode = audioContext.createScriptProcessor(2048, 1, 1);

            javascriptNode.onaudioprocess = function() {
                const array = new Uint8Array(analyser.frequencyBinCount);
                analyser.getByteFrequencyData(array);
                let values = 0;
                const length = array.length;
                for (let i = 0; i < length; i++) {
                    values += (array[i]);
                }
                const average = values / length;
                // Scale the average volume to a percentage for the meter width
                const volumePercent = Math.min(100, Math.max(0, average * 2)); // Adjust multiplier as needed
                voiceMeter.style.width = volumePercent + '%';
            }
        }

        function connectMicrophoneToMeter() {
            if (microphone) {
                 microphone.connect(analyser);
                 analyser.connect(javascriptNode);
                 javascriptNode.connect(audioContext.destination);
             }
        }
        function disconnectMicrophoneFromMeter() {
             if (microphone) {
                 javascriptNode.disconnect();
                 analyser.disconnect();
                 microphone.disconnect();
                 voiceMeter.style.width = '0%'; // Reset meter
             }
        }


        function setupSpeechRecognition() {
            const SpeechRecognition = window.SpeechRecognition || window.webkitSpeechRecognition;
            if (!SpeechRecognition) {
                 statusElement.textContent = "Speech Recognition API not supported in this browser.";
                recordButton.disabled = true;
                return;
            }

            recognition = new SpeechRecognition();
            recognition.continuous = false; // Process after silence
            recognition.lang = 'en-US';
            recognition.interimResults = false; // We only want final results
            recognition.maxAlternatives = 1;

            recognition.onresult = (event) => {
                const transcript = event.results[event.results.length - 1][0].transcript.trim();
                 statusElement.textContent = `Recognized: "${transcript}"`;
                console.log('Voice Result:', transcript);
                if (transcript) {
                    appendMessage(transcript, 'user');
                    generateResponse(transcript); // Send transcribed text to LLM
                }
            };

            recognition.onspeechend = () => {
                // Optionally stop recording automatically when speech ends
                // stopRecording();
                 statusElement.textContent = "Processing speech...";
            };

            recognition.onnomatch = () => {
                 statusElement.textContent = "Speech not recognized.";
            };

            recognition.onerror = (event) => {
                 statusElement.textContent = `Speech recognition error: ${event.error}`;
                console.error("Speech recognition error:", event.error);
                stopRecording(); // Ensure recording stops on error
            };
            
             recognition.onaudiostart = () => {
                statusElement.textContent = "Listening...";
                 connectMicrophoneToMeter();
            };
            recognition.onaudioend = () => {
                 disconnectMicrophoneFromMeter();
                 // statusElement.textContent = "Finished listening."; // Status updated by other events
            };
        }

        async function startRecording() {
            if (isRecording || !recognition) return;
            
             // Request microphone access if not already granted / setup audio context
             if (!audioContext) {
                 setupAudioMeter();
             }
             
            try {
                 // Re-request stream each time to ensure it's active
                const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
                microphone = audioContext.createMediaStreamSource(stream);

                 isRecording = true;
                recordButton.textContent = 'Stop Recording';
                recordButton.classList.add('recording');
                recognition.start();

             } catch (err) {
                 console.error("Error accessing microphone:", err);
                 statusElement.textContent = `Microphone Error: ${err.message}`;
             }
        }

        function stopRecording() {
             if (!isRecording || !recognition) return;
             recognition.stop();
             isRecording = false;
            recordButton.textContent = 'Start Recording';
            recordButton.classList.remove('recording');
             // No need to manually stop the stream tracks here, recognition.stop() handles it
             // Also, stopping tracks would prevent re-use if needed quickly again
             disconnectMicrophoneFromMeter(); // Disconnect meter processing
        }

        recordButton.addEventListener('click', () => {
             if (isRecording) {
                stopRecording();
            } else {
                startRecording();
            }
        });


        // Initialization
        initializeWebLLM();
        setupSpeechRecognition();
        chatInput.addEventListener('keypress', handleChatInput);

    



// ╔══ 💧 HYDRATION 💧 ══

// Make component available globally for hydration
window.Voice = Voice;

// React hydration using common utilities
try {
    // Use the global hydration function from _common.js
    window.hydrateReactApp('Voice', { 
        page: window.pageData || {},
        container: 'main',
		layout: React.createElement('div', {}, 'Layout placeholder')
    });
} catch(e) {
    console.error('React hydration error:', e);
}
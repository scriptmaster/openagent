// web-llm-worker.js
// This serves as the entry point for the WebLLM worker.
// It primarily imports and initializes the actual worker logic provided by the library.
try {
    self.importScripts("https://esm.run/@mlc-ai/web-llm/lib"); // Adjust path if hosting locally
} catch (e) {
    console.error("Error importing WebLLM worker script:", e);
    // Optionally, post a message back to the main thread indicating failure
    self.postMessage({ type: "error", message: "Failed to load WebLLM worker script." });
}

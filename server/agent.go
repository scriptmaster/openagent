package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Configuration constants (can be overridden by environment variables)
const (
	defaultOllamaURL = "http://host.docker.internal:11434" // Default for Docker Desktop, override for Linux
	defaultModel     = "llama3:8b"                         // Change to your preferred downloaded model
	defaultPort      = ":8800"                             // Port the agent's web server listens on inside the container
	maxIterations    = 20                                  // Safety limit for agent iterations
	requestTimeout   = 120 * time.Second                   // Timeout for Ollama requests
	execTimeout      = 60 * time.Second                    // Timeout for command execution
	htmlTemplatePath = "/app/tpl/agent.html"               // Path inside the container
	dataDir          = "/app/data"                         // Agent's working data directory inside the container
)

// --- Agent State ---
type AgentState string

const (
	StateIdle         AgentState = "Idle"
	StateThinking     AgentState = "Thinking..."
	StateExecuting    AgentState = "Executing Command..."
	StateAwaitingStep AgentState = "Awaiting Next Step"
	StateFinished     AgentState = "Finished"
	StateBlocked      AgentState = "Command Blocked (Safety)"
	StateError        AgentState = "Error"
)

// --- Agent Definition ---
type Agent struct {
	sync.Mutex // To protect concurrent access

	OllamaURL     string
	ModelName     string
	Goal          string
	History       []Message
	Iteration     int
	MaxIterations int
	HttpClient    *http.Client
	State         AgentState
	LastOutput    string
	LastError     string
}

type Message struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// --- Global Agent Instance ---
var globalAgent *Agent
var agentMutex sync.Mutex

// --- Ollama Structs ---
type OllamaRequest struct {
	Model   string                 `json:"model"`
	Prompt  string                 `json:"prompt"`
	System  string                 `json:"system,omitempty"`
	Stream  bool                   `json:"stream"`
	Options map[string]interface{} `json:"options,omitempty"`
}

type OllamaResponse struct {
	Model           string    `json:"model"`
	CreatedAt       time.Time `json:"created_at"`
	Response        string    `json:"response"`
	Done            bool      `json:"done"`
	TotalDuration   int64     `json:"total_duration"`
	LoadDuration    int64     `json:"load_duration"`
	PromptEvalCount int       `json:"prompt_eval_count"`
	EvalCount       int       `json:"eval_count"`
	EvalDuration    int64     `json:"eval_duration"`
}

// --- Config Loading ---
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	log.Printf("Using default for %s: %s", key, fallback)
	return fallback
}

// --- Agent Methods ---

func InitializeAgent(goal, ollamaURL, modelName string) {
	agentMutex.Lock()
	defer agentMutex.Unlock()

	globalAgent = &Agent{
		OllamaURL:     ollamaURL,
		ModelName:     modelName,
		Goal:          goal,
		History:       make([]Message, 0),
		Iteration:     0,
		MaxIterations: maxIterations,
		HttpClient: &http.Client{
			Timeout: requestTimeout,
		},
		State:      StateIdle,
		LastOutput: "",
		LastError:  "",
	}
	log.Printf("Agent initialized. Goal: '%s', Ollama: %s, Model: %s", goal, ollamaURL, modelName)
}

func GetAgentState() map[string]interface{} {
	agentMutex.Lock() // Lock the specific agent init mutex
	if globalAgent == nil {
		agentMutex.Unlock()
		return map[string]interface{}{
			"status":  StateIdle,
			"history": []Message{},
			"goal":    "",
			"error":   "Agent not initialized",
		}
	}
	agentMutex.Unlock() // Unlock agent init mutex

	// Lock the agent instance for reading state
	globalAgent.Lock()
	defer globalAgent.Unlock()

	historyCopy := make([]Message, len(globalAgent.History))
	copy(historyCopy, globalAgent.History)

	return map[string]interface{}{
		"status":        globalAgent.State,
		"history":       historyCopy,
		"iteration":     globalAgent.Iteration,
		"maxIterations": globalAgent.MaxIterations,
		"goal":          globalAgent.Goal,
		"lastOutput":    globalAgent.LastOutput,
		"lastError":     globalAgent.LastError,
	}
}

func (a *Agent) buildPrompt() string {
	var promptBuilder strings.Builder
	for _, msg := range a.History {
		promptBuilder.WriteString(fmt.Sprintf("[%s]\n%s\n\n", strings.ToUpper(msg.Role), msg.Content))
	}
	promptBuilder.WriteString("[ASSISTANT]\nWhat is the next single command to execute within the '/app/data' directory or the final answer? Respond ONLY with 'COMMAND: <command>' OR 'FINAL_ANSWER: <answer>'.")
	return promptBuilder.String()
}

func (a *Agent) buildSystemPrompt() string {
	// Updated prompt to mention the working directory constraint
	return fmt.Sprintf(`You are an autonomous AI agent running inside a restricted Docker container. Your goal is: %s
You can execute shell commands on the Linux system within the container to achieve the goal.
IMPORTANT: All file operations and commands that create output should target the '/app/data' directory. Do NOT attempt to write outside this directory. For example, use 'ls /app/data', 'mkdir /app/data/newdir', 'echo "hello" > /app/data/file.txt'.
Dangerous commands (like rm, dd, mkfs, shutdown, direct redirection overwrite > outside /app/data) are blocked. Use pipes | carefully.
Think step-by-step. Plan your actions.
Based on the history and the goal, decide the single next best shell command to execute.
Respond ONLY in one of the following two formats:
1. To execute a command: COMMAND: <command_to_execute> (Ensure paths are within /app/data where appropriate)
2. To provide the final answer: FINAL_ANSWER: <your_final_answer>

Do NOT provide explanations, apologies, or any text other than the chosen format.
If a command is blocked, analyze the error and try a different, safe approach within '/app/data'.
Use 'tee /app/data/...' or '>> /app/data/...' for writing files safely. Avoid '>' if possible, especially outside /app/data.
If the goal is achieved, provide the FINAL_ANSWER.
Current Date/Time: %s`, a.Goal, time.Now().Format(time.RFC3339))
}

func (a *Agent) addToHistory(role, content string) {
	// (History truncation logic - same as before)
	const maxHistoryTokens = 3500
	currentTokens := 0
	for _, msg := range a.History {
		currentTokens += len(strings.Fields(msg.Content))
	}
	tokenLenContent := len(strings.Fields(content))

	for (currentTokens+tokenLenContent) > maxHistoryTokens && len(a.History) > 1 {
		if a.History[0].Role == "system" && len(a.History) > 2 {
			removedMsg := a.History[1]
			a.History = append(a.History[:1], a.History[2:]...)
			currentTokens -= len(strings.Fields(removedMsg.Content))
			log.Printf("Truncated history, removed oldest non-system message (%s)", removedMsg.Role)
		} else if a.History[0].Role != "system" {
			removedMsg := a.History[0]
			a.History = a.History[1:]
			currentTokens -= len(strings.Fields(removedMsg.Content))
			log.Printf("Truncated history, removed oldest message (%s)", removedMsg.Role)
		} else {
			log.Println("Warning: Cannot truncate history further.")
			break
		}
	}

	log.Printf("Adding to History - Role: %s", role)
	a.History = append(a.History, Message{Role: role, Content: content, Timestamp: time.Now()})
}

// --- Security Check ---
var forbiddenCommandPrefixes = []string{
	"rm ", "sudo rm ", "mv /", "dd ", "mkfs", "shutdown", "reboot",
	":(){ :|:& };:", "chmod -R 000 /", "chown -R ",
	// Block writing directly to sensitive root dirs
	"> /etc", "> /dev", "> /bin", "> /sbin", "> /usr", "> /root", "> /var", "> /tmp", "> /run",
	"| sh", "| bash", "| zsh",
}
var allowedWritePrefix = dataDir + "/" // Allow writing inside the data directory

// isCommandSafe performs basic checks. NOT FOOLPROOF.
func isCommandSafe(commandStr string) (bool, string) {
	trimmedCmd := strings.TrimSpace(commandStr)
	lowerCmd := strings.ToLower(trimmedCmd)

	for _, prefix := range forbiddenCommandPrefixes {
		if strings.HasPrefix(lowerCmd, prefix) {
			return false, fmt.Sprintf("Command blocked: Starts with forbidden pattern '%s'", prefix)
		}
	}

	// Check for file operations outside the allowed directory
	// This is a basic check looking for redirection '>' or '>>' and common file manipulation commands
	if strings.Contains(trimmedCmd, ">") || strings.Contains(trimmedCmd, "touch ") || strings.Contains(trimmedCmd, "mkdir ") || strings.Contains(trimmedCmd, "cp ") || strings.Contains(trimmedCmd, "mv ") {
		// Extract potential file paths (very simplistic parsing)
		fields := strings.Fields(trimmedCmd)
		for _, field := range fields {
			// Check if a field looks like an absolute path or relative path starting with '/' or './' or '../'
			// And it's NOT within the allowed data directory
			if strings.Contains(field, "/") && !strings.HasPrefix(field, allowedWritePrefix) && field != ">" && field != ">>" && field != "|" && field != "&" {
				// Allow paths relative to current dir IF current dir is /app/data (or /app, needs check)
				// For simplicity, let's just block absolute paths outside /app/data for now
				if strings.HasPrefix(field, "/") {
					isPipeTarget := false
					pipeParts := strings.Split(trimmedCmd, "|")
					if len(pipeParts) > 1 && strings.Contains(pipeParts[len(pipeParts)-1], field) {
						isPipeTarget = true // Allow piping to tools like `grep /etc/passwd` (read-only) - risky!
					}

					// More refinement needed here. Let's block direct writes outside allowed prefix.
					if (strings.Contains(trimmedCmd, " > "+field) || strings.Contains(trimmedCmd, " >> "+field) || strings.HasPrefix(trimmedCmd, "touch "+field) || strings.HasPrefix(trimmedCmd, "mkdir "+field)) && !isPipeTarget {
						return false, fmt.Sprintf("Command blocked: Attempting file operation on '%s' which is outside the allowed '%s' directory.", field, dataDir)
					}
				}
			}
		}
	}

	// Allow 'cd /app/data' but not 'cd /' or 'cd /etc' etc.
	if strings.HasPrefix(lowerCmd, "cd ") {
		targetDir := strings.TrimSpace(strings.TrimPrefix(lowerCmd, "cd "))
		if targetDir != dataDir && !strings.HasPrefix(targetDir, dataDir+"/") && targetDir != "." && targetDir != ".." {
			// Allow relative cd inside dataDir - complex to check robustly here.
			// Safest bet is to perhaps ONLY allow `cd /app/data`.
			if targetDir != "/app/data" {
				return false, fmt.Sprintf("Command blocked: 'cd' is only allowed to '%s'. Attempted: '%s'", dataDir, targetDir)
			}
		}
	}

	return true, ""
}

// --- Agent Step Logic ---

func (a *Agent) Step() {
	a.Lock() // Lock the agent instance
	// Check state conditions (same as before)
	if a.State == StateFinished || a.State == StateBlocked || a.State == StateError || a.State == StateThinking || a.State == StateExecuting {
		log.Printf("Agent step requested but agent not in AwaitingStep state (current: %s).", a.State)
		a.Unlock()
		return
	}
	if a.Iteration >= a.MaxIterations {
		log.Printf("Agent reached max iterations (%d).\n", a.MaxIterations)
		a.State = StateFinished
		a.LastOutput = "Stopped: Reached maximum iteration limit."
		a.Unlock()
		return
	}

	a.Iteration++
	a.LastError = ""
	log.Printf("--- Agent Step: Iteration %d ---", a.Iteration)
	a.State = StateThinking
	a.Unlock() // *Unlock* before potentially long-running think/execute

	// Run think/execute in a separate goroutine to avoid blocking status updates
	go func() {
		// Re-lock within the goroutine to modify agent state
		a.Lock()
		// Ensure state was still Thinking when we re-acquired lock
		if a.State != StateThinking {
			log.Printf("Agent state changed unexpectedly before thinkInternal started (State: %s)", a.State)
			a.Unlock()
			return
		}
		action, err := a.thinkInternal() // Handles history update
		if err != nil {
			log.Printf("Error during thinking: %v", err)
			a.State = StateError
			a.LastError = fmt.Sprintf("Thinking error: %v", err)
			a.addToHistory("system", fmt.Sprintf("System Error during thinking phase: %v. Please analyze and proceed.", err))
			a.Unlock()
			return
		}

		// Now execute
		a.State = StateExecuting
		observation, isFinal, blockReason := a.executeInternal(action) // Handles history update for result

		// Update state based on execution outcome
		if blockReason != "" {
			log.Printf("Command blocked: %s", blockReason)
			a.State = StateBlocked
			a.LastError = blockReason
			// History already updated in executeInternal for blocked commands
		} else {
			a.LastOutput = observation
			if isFinal {
				log.Println("Agent received final answer.")
				a.State = StateFinished
			} else {
				a.State = StateAwaitingStep
				log.Println("Agent is awaiting the next step.")
			}
		}
		a.Unlock() // Unlock after state update
	}()
}

// thinkInternal (requires agent lock held)
func (a *Agent) thinkInternal() (string, error) {
	log.Println("Agent thinking...")
	// Ensure system prompt (same as before)
	if len(a.History) == 0 || a.History[0].Role != "system" {
		systemPrompt := a.buildSystemPrompt()
		if len(a.History) == 0 {
			a.addToHistory("system", systemPrompt)
		} else {
			a.History = append([]Message{{Role: "system", Content: systemPrompt, Timestamp: time.Now()}}, a.History...)
		}
	}

	fullPromptString := a.buildPrompt()
	requestPayload := OllamaRequest{
		Model:   a.ModelName,
		Prompt:  fullPromptString,
		Stream:  false,
		Options: map[string]interface{}{"temperature": 0.5},
	}
	jsonData, err := json.Marshal(requestPayload)
	// ... (Ollama request/response handling - same as before) ...
	if err != nil {
		return "", fmt.Errorf("error marshalling ollama request: %w", err)
	}
	req, err := http.NewRequest("POST", a.OllamaURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating ollama request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.HttpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request to ollama: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading ollama response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var ollamaResp OllamaResponse
	err = json.Unmarshal(body, &ollamaResp)
	if err != nil {
		log.Printf("Raw Ollama Response on Unmarshal Error:\n%s\n", string(body))
		return "", fmt.Errorf("error unmarshalling ollama response: %w. Body: %s", err, string(body))
	}
	assistantResponse := strings.TrimSpace(ollamaResp.Response)
	log.Printf("Ollama Response Received.") // Don't log full response here by default

	// Add assistant's *intended* action to history
	a.addToHistory("assistant", assistantResponse)
	return assistantResponse, nil
}

// executeInternal (requires agent lock held)
func (a *Agent) executeInternal(action string) (string, bool, string) {
	log.Printf("Attempting to execute action: %s", action) // Log action attempt

	if strings.HasPrefix(action, "FINAL_ANSWER:") {
		finalAnswer := strings.TrimSpace(strings.TrimPrefix(action, "FINAL_ANSWER:"))
		log.Printf("Final Answer Received: %s", finalAnswer)
		a.addToHistory("user", fmt.Sprintf("Result of action: Final Answer Provided\n%s", finalAnswer))
		return finalAnswer, true, ""
	}

	if strings.HasPrefix(action, "COMMAND:") {
		commandStr := strings.TrimSpace(strings.TrimPrefix(action, "COMMAND:"))
		if commandStr == "" {
			errorMsg := "Error: Empty command received."
			a.addToHistory("user", fmt.Sprintf("Result of action: %s", errorMsg))
			return errorMsg, false, ""
		}

		safe, reason := isCommandSafe(commandStr)
		if !safe {
			log.Printf("Safety Check Failed: %s (Command: %s)", reason, commandStr)
			// Add info about blocking to history for the LLM to see
			a.addToHistory("user", fmt.Sprintf("Command '%s' was blocked by safety filter: %s. Propose a different, safe command within '/app/data'.", commandStr, reason))
			return fmt.Sprintf("Command blocked by safety filter: %s", reason), false, reason
		}

		log.Printf("Executing safe command: %s (in dir: %s)", commandStr, dataDir)
		ctx, cancel := context.WithTimeout(context.Background(), execTimeout)
		defer cancel()

		cmd := exec.CommandContext(ctx, "sh", "-c", commandStr)
		cmd.Dir = dataDir // *** Execute command within the data directory ***
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		startTime := time.Now()
		err := cmd.Run()
		duration := time.Since(startTime)
		output := stdout.String()
		errMsg := stderr.String()
		result := ""

		if ctx.Err() == context.DeadlineExceeded {
			result = fmt.Sprintf("Command timed out after %s.\nSTDOUT:\n%s\nSTDERR:\n%s", duration, output, errMsg)
			log.Println("Command execution timed out.")
		} else if err != nil {
			result = fmt.Sprintf("Command failed (Duration: %s).\nError: %s\nSTDOUT:\n%s\nSTDERR:\n%s", duration, err, output, errMsg)
			log.Println("Command execution failed.")
		} else {
			result = fmt.Sprintf("Command executed successfully (Duration: %s).\nSTDOUT:\n%s\nSTDERR:\n%s", duration, output, errMsg)
			log.Println("Command execution succeeded.")
		}
		// Add execution result to history
		a.addToHistory("user", fmt.Sprintf("Result of action '%s':\n%s", commandStr, result))
		return result, false, ""
	}

	// Incorrect format
	log.Printf("⚠️ Assistant response did not match expected format: %s", action)
	errorMsg := fmt.Sprintf("Error: Invalid action format received from assistant: '%s'. Please respond ONLY with 'COMMAND: <command>' or 'FINAL_ANSWER: <answer>'.", action)
	a.addToHistory("user", fmt.Sprintf("Result of action: %s", errorMsg))
	return errorMsg, false, ""
}

// --- Web Server Handlers ---

// Global template variable
var tpl *template.Template

// InitAgentTemplates initializes the agent templates from the main template engine
func InitAgentTemplates(mainTemplates *template.Template) {
	tpl = mainTemplates
}

// HandleAgent renders the main agent interface page.
// It now uses the global templates variable initialized in routes.go.
func HandleAgent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Get the current app version
	appVersion := AppVersion

	// Create template data with version
	data := struct {
		AppVersion string
		Project    interface{}
	}{
		AppVersion: appVersion,
		Project:    nil,
	}

	// Use the imported template from main.go rather than the local tpl
	if err := tpl.ExecuteTemplate(w, "agent.html", data); err != nil {
		http.Error(w, "Internal Server Error: Could not execute template", http.StatusInternalServerError)
		log.Printf("Error executing template: %v", err)
	}
}

// HandleStart initializes the agent with a new goal.
func HandleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request: Could not parse form", http.StatusBadRequest)
		return
	}
	goal := r.FormValue("goal")
	if goal == "" {
		http.Error(w, "Bad Request: Goal cannot be empty", http.StatusBadRequest)
		return
	}

	prompt := r.FormValue("prompt") // Get optional prompt

	ollamaURL := getEnv("OLLAMA_URL", defaultOllamaURL)
	modelName := getEnv("OLLAMA_MODEL", defaultModel)
	InitializeAgent(goal, ollamaURL, modelName)

	// Add initial user prompt to history
	globalAgent.Lock()
	initialUserPrompt := fmt.Sprintf("My goal is: %s. What is the first safe shell command I should execute within the '%s' directory?", goal, dataDir)
	if prompt != "" {
		initialUserPrompt = fmt.Sprintf("%s\nAdditional context: %s", initialUserPrompt, prompt)
	}
	globalAgent.addToHistory("user", initialUserPrompt)
	globalAgent.State = StateAwaitingStep
	globalAgent.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GetAgentState())
}

// HandleNextStep triggers the agent to perform its next thinking/execution cycle.
func HandleNextStep(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request: Could not parse form", http.StatusBadRequest)
		return
	}

	prompt := r.FormValue("prompt") // Get optional prompt

	agentMutex.Lock() // Lock agent init mutex
	if globalAgent == nil {
		agentMutex.Unlock()
		http.Error(w, "Bad Request: Agent not initialized", http.StatusBadRequest)
		return
	}
	agentMutex.Unlock() // Unlock agent init mutex

	// Add prompt to history if provided
	if prompt != "" {
		globalAgent.Lock()
		globalAgent.addToHistory("user", fmt.Sprintf("Additional context: %s", prompt))
		globalAgent.Unlock()
	}

	// globalAgent.Step() now runs the core logic in a goroutine
	globalAgent.Step()

	w.Header().Set("Content-Type", "application/json")
	// Give a brief moment for the state to potentially update to Thinking/Executing
	time.Sleep(50 * time.Millisecond)
	json.NewEncoder(w).Encode(GetAgentState())
}

// HandleStatus returns the current state of the agent.
func HandleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GetAgentState())
}

// --- Main Function ---
// StartAgent is intended to be called if the agent runs as a standalone service.
// In the integrated setup, the handlers are registered by the main server's routes.
func StartAgent() {
	log.Println("--- Go Web Agent Starting (Standalone Mode) ---")

	// Load configuration from environment variables
	ollamaURL := getEnv("OLLAMA_URL", defaultOllamaURL)
	modelName := getEnv("OLLAMA_MODEL", defaultModel)
	port := getEnv("PORT", defaultPort)
	if !strings.HasPrefix(port, ":") {
		port = ":" + port // Ensure port starts with ':'
	}

	log.Printf("Ollama URL: %s", ollamaURL)
	log.Printf("Using Model: %s", modelName)
	log.Printf("Agent Data Directory: %s", dataDir)
	log.Printf("Web server starting on port %s", port)
	log.Println("Basic command safety checks implemented.")

	// Ensure data directory exists (might be created by volume mount, but good practice)
	err := os.MkdirAll(dataDir, 0755)
	if err != nil {
		log.Fatalf("Failed to create data directory '%s': %v", dataDir, err)
	} else {
		log.Printf("Ensured data directory exists: %s", dataDir)
	}

	// Parse the HTML template file
	tpl, err = template.ParseFiles(htmlTemplatePath)
	if err != nil {
		log.Fatalf("FATAL: Failed to parse template file '%s': %v", htmlTemplatePath, err)
	} else {
		log.Printf("Successfully parsed template: %s", htmlTemplatePath)
	}

	// Initialize with a nil agent
	globalAgent = nil

	// In standalone mode, register handlers directly.
	// In integrated mode, these are registered via server/routes.go.
	http.HandleFunc("/", HandleAgent)
	http.HandleFunc("/start", HandleStart)
	http.HandleFunc("/next", HandleNextStep)
	http.HandleFunc("/status", HandleStatus)

	log.Fatal(http.ListenAndServe(port, nil))
}

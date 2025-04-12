package main

// AgentRequest represents a request to the agent
type AgentRequest struct {
	Prompt string `json:"prompt"`
}

// AgentResponse represents a response from the agent
type AgentResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// Note: Agent handlers (handleAgent, handleStart, handleNextStep, handleStatus)
// are implemented in agent.go to avoid duplication.

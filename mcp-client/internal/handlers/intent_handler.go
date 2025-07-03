package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/bchamber/taskman/mcp-client/internal/client"
)

// IntentHandler handles MCP JSON intents
type IntentHandler struct {
	mcpClient *client.MCPClient
	logger    *slog.Logger
}

// Intent represents a structured JSON intent as specified by MCP protocol
type Intent struct {
	Method string      `json:"method"`
	Params interface{} `json:"params,omitempty"`
}

// ToolCallParams represents parameters for tool execution
type ToolCallParams struct {
	Name      string      `json:"name"`
	Arguments interface{} `json:"arguments,omitempty"`
}

// PromptGetParams represents parameters for prompt retrieval
type PromptGetParams struct {
	Name      string      `json:"name"`
	Arguments interface{} `json:"arguments,omitempty"`
}

// NewIntentHandler creates a new intent handler
func NewIntentHandler(mcpClient *client.MCPClient, logger *slog.Logger) *IntentHandler {
	return &IntentHandler{
		mcpClient: mcpClient,
		logger:    logger,
	}
}

// ProcessIntent processes a JSON intent according to MCP specification
func (h *IntentHandler) ProcessIntent(ctx context.Context, intentJSON string) (interface{}, error) {
	h.logger.Info("Processing intent", "json", intentJSON)
	
	// Parse the intent JSON
	var intent Intent
	if err := json.Unmarshal([]byte(intentJSON), &intent); err != nil {
		h.logger.Error("Failed to parse intent JSON", "error", err)
		return nil, fmt.Errorf("failed to parse intent JSON: %w", err)
	}
	
	// Process the intent based on method
	switch intent.Method {
	case "tools/list":
		return h.handleListTools(ctx)
	case "tools/call":
		return h.handleExecuteTool(ctx, intent.Params)
	case "prompts/list":
		return h.handleListPrompts(ctx)
	case "prompts/get":
		return h.handleGetPrompt(ctx, intent.Params)
	default:
		h.logger.Error("Unsupported intent method", "method", intent.Method)
		return nil, fmt.Errorf("unsupported intent method: %s", intent.Method)
	}
}

// handleListTools handles tools/list intent
func (h *IntentHandler) handleListTools(ctx context.Context) (interface{}, error) {
	h.logger.Info("Handling list tools intent")
	
	resp, err := h.mcpClient.ListTools(ctx)
	if err != nil {
		h.logger.Error("Failed to list tools", "error", err)
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}
	
	if resp.Error != nil {
		h.logger.Error("MCP server returned error for list tools", "error", resp.Error.Message)
		return nil, fmt.Errorf("MCP server error: %s", resp.Error.Message)
	}
	
	return resp.Result, nil
}

// handleExecuteTool handles tools/call intent
func (h *IntentHandler) handleExecuteTool(ctx context.Context, params interface{}) (interface{}, error) {
	h.logger.Info("Handling execute tool intent")
	
	// Parse tool call parameters
	paramsBytes, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tool params: %w", err)
	}
	
	var toolParams ToolCallParams
	if err := json.Unmarshal(paramsBytes, &toolParams); err != nil {
		h.logger.Error("Failed to parse tool call parameters", "error", err)
		return nil, fmt.Errorf("failed to parse tool call parameters: %w", err)
	}
	
	if toolParams.Name == "" {
		return nil, fmt.Errorf("tool name is required")
	}
	
	h.logger.Info("Executing tool", "name", toolParams.Name)
	
	resp, err := h.mcpClient.ExecuteTool(ctx, toolParams.Name, toolParams.Arguments)
	if err != nil {
		h.logger.Error("Failed to execute tool", "tool", toolParams.Name, "error", err)
		return nil, fmt.Errorf("failed to execute tool %s: %w", toolParams.Name, err)
	}
	
	if resp.Error != nil {
		h.logger.Error("MCP server returned error for tool execution", "tool", toolParams.Name, "error", resp.Error.Message)
		return nil, fmt.Errorf("MCP server error executing tool %s: %s", toolParams.Name, resp.Error.Message)
	}
	
	return resp.Result, nil
}

// handleListPrompts handles prompts/list intent
func (h *IntentHandler) handleListPrompts(ctx context.Context) (interface{}, error) {
	h.logger.Info("Handling list prompts intent")
	
	resp, err := h.mcpClient.ListPrompts(ctx)
	if err != nil {
		h.logger.Error("Failed to list prompts", "error", err)
		return nil, fmt.Errorf("failed to list prompts: %w", err)
	}
	
	if resp.Error != nil {
		h.logger.Error("MCP server returned error for list prompts", "error", resp.Error.Message)
		return nil, fmt.Errorf("MCP server error: %s", resp.Error.Message)
	}
	
	return resp.Result, nil
}

// handleGetPrompt handles prompts/get intent
func (h *IntentHandler) handleGetPrompt(ctx context.Context, params interface{}) (interface{}, error) {
	h.logger.Info("Handling get prompt intent")
	
	// Parse prompt get parameters
	paramsBytes, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal prompt params: %w", err)
	}
	
	var promptParams PromptGetParams
	if err := json.Unmarshal(paramsBytes, &promptParams); err != nil {
		h.logger.Error("Failed to parse prompt get parameters", "error", err)
		return nil, fmt.Errorf("failed to parse prompt get parameters: %w", err)
	}
	
	if promptParams.Name == "" {
		return nil, fmt.Errorf("prompt name is required")
	}
	
	h.logger.Info("Getting prompt", "name", promptParams.Name)
	
	resp, err := h.mcpClient.GetPrompt(ctx, promptParams.Name, promptParams.Arguments)
	if err != nil {
		h.logger.Error("Failed to get prompt", "prompt", promptParams.Name, "error", err)
		return nil, fmt.Errorf("failed to get prompt %s: %w", promptParams.Name, err)
	}
	
	if resp.Error != nil {
		h.logger.Error("MCP server returned error for prompt get", "prompt", promptParams.Name, "error", resp.Error.Message)
		return nil, fmt.Errorf("MCP server error getting prompt %s: %s", promptParams.Name, resp.Error.Message)
	}
	
	return resp.Result, nil
}
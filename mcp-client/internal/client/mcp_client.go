package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// MCPClient represents a client for communicating with an MCP server via HTTP
type MCPClient struct {
	serverURL  string
	httpClient *http.Client
	logger     *slog.Logger
	sessionID  string
	initialized bool
}

// MCPRequest represents a standard MCP request in JSON-RPC 2.0 format
type MCPRequest struct {
	JSONRpc string      `json:"jsonrpc"`
	ID      string      `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// MCPResponse represents a standard MCP response in JSON-RPC 2.0 format
type MCPResponse struct {
	JSONRpc string      `json:"jsonrpc"`
	ID      string      `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

// MCPError represents an MCP error response
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// NewMCPClient creates a new MCP client
func NewMCPClient(serverURL string, logger *slog.Logger) *MCPClient {
	return &MCPClient{
		serverURL: serverURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// SendRequest sends an MCP request and returns the response
func (c *MCPClient) SendRequest(ctx context.Context, req MCPRequest) (*MCPResponse, error) {
	c.logger.Info("Sending MCP request", "method", req.Method, "id", req.ID)
	
	// Marshal request to JSON
	reqBody, err := json.Marshal(req)
	if err != nil {
		c.logger.Error("Failed to marshal request", "error", err)
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.serverURL, bytes.NewBuffer(reqBody))
	if err != nil {
		c.logger.Error("Failed to create HTTP request", "error", err)
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")
	
	// Add session ID if we have one
	if c.sessionID != "" {
		httpReq.Header.Set("Mcp-Session-Id", c.sessionID)
	}
	
	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		c.logger.Error("Failed to send HTTP request", "error", err)
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()
	
	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		// Read error body for logging
		errorBody, _ := io.ReadAll(resp.Body)
		c.logger.Error("HTTP request failed", "status", resp.StatusCode, "body", string(errorBody))
		return nil, fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(errorBody))
	}
	
	// Extract session ID from headers if present
	if sessionID := resp.Header.Get("Mcp-Session-Id"); sessionID != "" {
		c.sessionID = sessionID
		c.logger.Info("Updated session ID", "session_id", sessionID)
	}
	
	// Parse SSE response
	sseData, err := ParseSSEResponse(resp.Body)
	if err != nil {
		c.logger.Error("Failed to parse SSE response", "error", err)
		return nil, fmt.Errorf("failed to parse SSE response: %w", err)
	}
	
	// Parse MCP response from SSE data
	var mcpResp MCPResponse
	if err := json.Unmarshal([]byte(sseData), &mcpResp); err != nil {
		c.logger.Error("Failed to unmarshal MCP response", "error", err, "data", sseData)
		return nil, fmt.Errorf("failed to unmarshal MCP response: %w", err)
	}
	
	c.logger.Info("Received MCP response", "id", mcpResp.ID, "hasResult", mcpResp.Result != nil, "hasError", mcpResp.Error != nil)
	
	return &mcpResp, nil
}

// Initialize initializes the MCP session with the server
func (c *MCPClient) Initialize(ctx context.Context) error {
	if c.initialized {
		return nil // Already initialized
	}
	
	c.logger.Info("Initializing MCP session")
	
	req := MCPRequest{
		JSONRpc: "2.0",
		ID:      generateID(),
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2025-03-26",
			"capabilities": map[string]interface{}{
				"tools":     map[string]interface{}{},
				"prompts":   map[string]interface{}{},
				"resources": map[string]interface{}{},
			},
			"clientInfo": map[string]interface{}{
				"name":    "taskman-mcp-client",
				"version": "1.0.0",
			},
		},
	}
	
	resp, err := c.SendRequest(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to initialize MCP session: %w", err)
	}
	
	if resp.Error != nil {
		return fmt.Errorf("MCP initialization error: %s", resp.Error.Message)
	}
	
	c.initialized = true
	c.logger.Info("MCP session initialized successfully")
	
	// Send initialized notification
	notificationReq := MCPRequest{
		JSONRpc: "2.0",
		Method:  "initialized",
		Params:  map[string]interface{}{},
	}
	
	// Note: initialized is a notification, so we don't expect a response
	_, err = c.SendRequest(ctx, notificationReq)
	if err != nil {
		c.logger.Warn("Failed to send initialized notification", "error", err)
		// Don't fail initialization for this
	}
	
	return nil
}

// ensureInitialized ensures the MCP session is initialized
func (c *MCPClient) ensureInitialized(ctx context.Context) error {
	if !c.initialized {
		return c.Initialize(ctx)
	}
	return nil
}

// ListTools lists available tools from the MCP server
func (c *MCPClient) ListTools(ctx context.Context) (*MCPResponse, error) {
	if err := c.ensureInitialized(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize before listing tools: %w", err)
	}
	
	req := MCPRequest{
		JSONRpc: "2.0",
		ID:      generateID(),
		Method:  "tools/list",
	}
	
	return c.SendRequest(ctx, req)
}

// ExecuteTool executes a tool with the given name and parameters
func (c *MCPClient) ExecuteTool(ctx context.Context, toolName string, params interface{}) (*MCPResponse, error) {
	if err := c.ensureInitialized(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize before executing tool: %w", err)
	}
	
	req := MCPRequest{
		JSONRpc: "2.0",
		ID:      generateID(),
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      toolName,
			"arguments": params,
		},
	}
	
	return c.SendRequest(ctx, req)
}

// ListPrompts lists available prompts from the MCP server
func (c *MCPClient) ListPrompts(ctx context.Context) (*MCPResponse, error) {
	if err := c.ensureInitialized(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize before listing prompts: %w", err)
	}
	
	req := MCPRequest{
		JSONRpc: "2.0",
		ID:      generateID(),
		Method:  "prompts/list",
	}
	
	return c.SendRequest(ctx, req)
}

// GetPrompt gets a specific prompt with the given name and parameters
func (c *MCPClient) GetPrompt(ctx context.Context, promptName string, params interface{}) (*MCPResponse, error) {
	if err := c.ensureInitialized(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize before getting prompt: %w", err)
	}
	
	req := MCPRequest{
		JSONRpc: "2.0",
		ID:      generateID(),
		Method:  "prompts/get",
		Params: map[string]interface{}{
			"name":      promptName,
			"arguments": params,
		},
	}
	
	return c.SendRequest(ctx, req)
}

// generateID generates a unique ID for MCP requests
func generateID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}
package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/bchamber/taskman-mcp/internal/client"
	"github.com/bchamber/taskman-mcp/internal/config"
)

type Server struct {
	mcpServer  *mcp.Server
	apiClient  *client.APIClient
	config     *config.Config
	httpServer *http.Server
}

func NewServer(cfg *config.Config) *Server {
	slog.Info("Creating new MCP server", 
		"server_name", cfg.ServerName,
		"server_version", cfg.ServerVersion,
	)
	
	// Create MCP server instance with proper options
	serverOptions := &mcp.ServerOptions{
		Instructions: "Taskman MCP Server - Provides task management tools, resources, and prompts",
		InitializedHandler: func(ctx context.Context, session *mcp.ServerSession, params *mcp.InitializedParams) {
			slog.Info("MCP client initialized", "session", session)
		},
		PageSize:  100,
		KeepAlive: 30 * time.Second,
	}
	
	mcpServer := mcp.NewServer(cfg.ServerName, cfg.ServerVersion, serverOptions)
	
	// Create API client
	apiClient := client.NewAPIClient(cfg.APIBaseURL, cfg.APITimeout)
	
	server := &Server{
		mcpServer: mcpServer,
		apiClient: apiClient,
		config:    cfg,
	}
	
	// Set up HTTP server if needed
	if cfg.TransportMode == "http" || cfg.TransportMode == "both" {
		server.setupHTTPServer()
	}
	
	// Register tools, resources, and prompts
	server.registerTools()
	server.registerResources()
	server.registerPrompts()
	
	slog.Info("MCP server created successfully")
	return server
}

func (s *Server) registerTools() {
	slog.Info("Registering MCP tools")
	
	// Register a basic health check tool to demonstrate functionality
	healthTool := mcp.NewServerTool(
		"health_check",
		"Check the health of the taskman API server",
		s.handleHealthCheck,
		// No input parameters needed for health check
	)
	
	s.mcpServer.AddTools(healthTool)
	
	slog.Info("Tools registration completed", "tool_count", 1)
}

// Health check tool handler
func (s *Server) handleHealthCheck(
	ctx context.Context,
	session *mcp.ServerSession,
	params *mcp.CallToolParamsFor[map[string]any],
) (*mcp.CallToolResultFor[map[string]any], error) {
	slog.Info("Executing health_check tool")
	
	// Make API call to health endpoint
	resp, err := s.apiClient.Get(ctx, "/health")
	if err != nil {
		slog.Error("Health check failed", "error", err)
		return &mcp.CallToolResultFor[map[string]any]{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Health check failed: %v", err),
				},
			},
		}, nil
	}
	
	result := map[string]any{
		"status":    "healthy",
		"api_response": string(resp),
		"timestamp": time.Now().Format(time.RFC3339),
	}
	
	slog.Info("Health check completed successfully")
	
	return &mcp.CallToolResultFor[map[string]any]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("API Health Check: %s", result["status"]),
			},
		},
		Meta: result,
	}, nil
}

func (s *Server) registerResources() {
	slog.Info("Registering MCP resources")
	
	// Register a basic API status resource
	statusResource := &mcp.ServerResource{
		Resource: &mcp.Resource{
			URI:         "taskman://api/status",
			Name:        "API Status",
			Description: "Current status of the taskman API server",
			MIMEType:    "application/json",
		},
		Handler: s.handleStatusResource,
	}
	
	s.mcpServer.AddResources(statusResource)
	
	slog.Info("Resources registration completed", "resource_count", 1)
}

// Status resource handler
func (s *Server) handleStatusResource(
	ctx context.Context,
	session *mcp.ServerSession,
	params *mcp.ReadResourceParams,
) (*mcp.ReadResourceResult, error) {
	slog.Info("Reading status resource", "uri", params.URI)
	
	// Make API call to health endpoint
	resp, err := s.apiClient.Get(ctx, "/health")
	if err != nil {
		slog.Error("Failed to read API status", "error", err)
		return nil, fmt.Errorf("failed to read API status: %w", err)
	}
	
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      params.URI,
				MIMEType: "application/json",
				Text:     string(resp),
			},
		},
	}, nil
}

func (s *Server) registerPrompts() {
	slog.Info("Registering MCP prompts")
	
	// Register a basic task creation prompt template
	taskPrompt := &mcp.ServerPrompt{
		Prompt: &mcp.Prompt{
			Name:        "create_task",
			Description: "Template for creating a new task with proper context",
			Arguments: []*mcp.PromptArgument{
				{
					Name:        "task_name",
					Description: "Name of the task to create",
					Required:    true,
				},
				{
					Name:        "project_id",
					Description: "Optional project ID to associate with the task",
					Required:    false,
				},
			},
		},
		Handler: s.handleCreateTaskPrompt,
	}
	
	s.mcpServer.AddPrompts(taskPrompt)
	
	slog.Info("Prompts registration completed", "prompt_count", 1)
}

// Create task prompt handler
func (s *Server) handleCreateTaskPrompt(
	ctx context.Context,
	session *mcp.ServerSession,
	params *mcp.GetPromptParams,
) (*mcp.GetPromptResult, error) {
	slog.Info("Generating create_task prompt", "name", params.Name)
	
	// Extract arguments
	taskName := ""
	projectID := ""
	
	if params.Arguments != nil {
		if name, ok := params.Arguments["task_name"]; ok {
			taskName = name
		}
		if pid, ok := params.Arguments["project_id"]; ok {
			projectID = pid
		}
	}
	
	// Generate prompt text
	promptText := fmt.Sprintf(`Create a new task with the following details:

Task Name: %s`, taskName)
	
	if projectID != "" {
		promptText += fmt.Sprintf(`
Project ID: %s`, projectID)
	}
	
	promptText += `

Please provide:
1. A detailed description for this task
2. Appropriate priority level (Low, Medium, High)
3. Estimated completion timeline
4. Any dependencies or prerequisites
5. Success criteria for completion`
	
	return &mcp.GetPromptResult{
		Description: "Task creation guidance prompt",
		Messages: []*mcp.PromptMessage{
			{
				Role: "user",
				Content: &mcp.TextContent{
					Text: promptText,
				},
			},
		},
	}, nil
}

func (s *Server) setupHTTPServer() {
	mux := http.NewServeMux()
	
	// Create SSE handler that provides access to our server
	sseHandler := mcp.NewSSEHandler(func(r *http.Request) *mcp.Server {
		return s.mcpServer
	})
	
	// Set up SSE endpoint for streaming connections
	mux.Handle("/sse", sseHandler)
	
	// Set up streamable HTTP handler for HTTP transport
	streamableHandler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return s.mcpServer
	}, nil)
	
	// Set up streamable HTTP endpoint  
	mux.Handle("/mcp", streamableHandler)
	
	addr := fmt.Sprintf("%s:%s", s.config.HTTPHost, s.config.HTTPPort)
	s.httpServer = &http.Server{
		Addr:           addr,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	
	slog.Info("HTTP server configured", 
		"address", addr,
		"sse_endpoint", "/sse",
		"http_endpoint", "/mcp",
	)
}

func (s *Server) Run(ctx context.Context) error {
	slog.Info("Starting MCP server", "transport_mode", s.config.TransportMode)
	
	var wg sync.WaitGroup
	errCh := make(chan error, 2)
	
	// Start stdio transport if needed
	if s.config.TransportMode == "stdio" || s.config.TransportMode == "both" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			slog.Info("Starting MCP server with stdio transport")
			
			transport := mcp.NewStdioTransport()
			if err := s.mcpServer.Run(ctx, transport); err != nil {
				slog.Error("Stdio MCP server failed", "error", err)
				errCh <- fmt.Errorf("stdio transport error: %w", err)
			}
		}()
	}
	
	// Start HTTP server if needed
	if s.config.TransportMode == "http" || s.config.TransportMode == "both" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			slog.Info("Starting HTTP server", "address", s.httpServer.Addr)
			
			if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.Error("HTTP server failed", "error", err)
				errCh <- fmt.Errorf("HTTP server error: %w", err)
			}
		}()
		
		// Handle graceful HTTP server shutdown
		go func() {
			<-ctx.Done()
			if s.httpServer != nil {
				slog.Info("Shutting down HTTP server")
				if err := s.httpServer.Shutdown(context.Background()); err != nil {
					slog.Error("HTTP server shutdown error", "error", err)
				}
			}
		}()
	}
	
	// Wait for either completion or error
	go func() {
		wg.Wait()
		close(errCh)
	}()
	
	// Return first error if any
	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case <-ctx.Done():
		slog.Info("Server stopped by context cancellation")
	}
	
	slog.Info("MCP server stopped")
	return nil
}
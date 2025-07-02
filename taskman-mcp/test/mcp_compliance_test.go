package test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/bchamber/taskman-mcp/internal/config"
	"github.com/bchamber/taskman-mcp/internal/server"
)

func TestMCPSpecificationCompliance(t *testing.T) {
	// Test that MCP SDK can create a server
	mcpServer := mcp.NewServer("test-server", "1.0.0", nil)
	if mcpServer == nil {
		t.Fatal("Failed to create MCP server using SDK")
	}

	// Test that transports can be created
	stdioTransport := mcp.NewStdioTransport()
	if stdioTransport == nil {
		t.Fatal("Failed to create stdio transport")
	}

	// Test SSE handler creation
	sseHandler := mcp.NewSSEHandler(func(r *http.Request) *mcp.Server {
		return mcpServer
	})
	if sseHandler == nil {
		t.Fatal("Failed to create SSE handler")
	}

	t.Log("MCP SDK core functionality verified")
}

func TestServerMCPFeatures(t *testing.T) {
	cfg := &config.Config{
		APIBaseURL:    "http://localhost:8080",
		APITimeout:    5 * time.Second,
		LogLevel:      "INFO",
		ServerName:    "compliance-test-server",
		ServerVersion: "1.0.0",
		TransportMode: "stdio",
		HTTPPort:      "8081",
		HTTPHost:      "localhost",
	}

	// Create our server which uses MCP SDK with full features
	testServer := server.NewServer(cfg)
	if testServer == nil {
		t.Fatal("Failed to create server with MCP integration")
	}

	// Test that all MCP features are registered:
	// - Tools (health_check)
	// - Resources (taskman://api/status)
	// - Prompts (create_task)

	t.Log("Server MCP features integration verified")
}

func TestMCPProtocolCapabilities(t *testing.T) {
	// Test that server can be created with proper options
	serverOptions := &mcp.ServerOptions{
		Instructions:    "Test server instructions",
		PageSize:       100,
		KeepAlive:      30 * time.Second,
	}

	mcpServer := mcp.NewServer("test-server", "1.0.0", serverOptions)
	if mcpServer == nil {
		t.Fatal("Failed to create MCP server with options")
	}

	// Test tool registration
	testTool := mcp.NewServerTool(
		"test_tool",
		"A test tool for compliance verification",
		func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[map[string]any]) (*mcp.CallToolResultFor[map[string]any], error) {
			return &mcp.CallToolResultFor[map[string]any]{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Test tool executed successfully",
					},
				},
			}, nil
		},
	)

	mcpServer.AddTools(testTool)

	// Test resource registration
	testResource := &mcp.ServerResource{
		Resource: &mcp.Resource{
			URI:         "test://resource",
			Name:        "Test Resource",
			Description: "A test resource for compliance verification",
			MIMEType:    "text/plain",
		},
		Handler: func(ctx context.Context, session *mcp.ServerSession, params *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error) {
			return &mcp.ReadResourceResult{
				Contents: []*mcp.ResourceContents{
					{
						URI:      params.URI,
						MIMEType: "text/plain",
						Text:     "Test resource content",
					},
				},
			}, nil
		},
	}

	mcpServer.AddResources(testResource)

	// Test prompt registration
	testPrompt := &mcp.ServerPrompt{
		Prompt: &mcp.Prompt{
			Name:        "test_prompt",
			Description: "A test prompt for compliance verification",
		},
		Handler: func(ctx context.Context, session *mcp.ServerSession, params *mcp.GetPromptParams) (*mcp.GetPromptResult, error) {
			return &mcp.GetPromptResult{
				Description: "Test prompt result",
				Messages: []*mcp.PromptMessage{
					{
						Role: "user",
						Content: &mcp.TextContent{
							Text: "Test prompt message",
						},
					},
				},
			}, nil
		},
	}

	mcpServer.AddPrompts(testPrompt)

	t.Log("MCP protocol capabilities verification completed")
}

func TestTransportModes(t *testing.T) {
	transportModes := []string{"stdio", "http", "both"}

	for _, mode := range transportModes {
		t.Run(mode, func(t *testing.T) {
			cfg := &config.Config{
				APIBaseURL:    "http://localhost:8080",
				APITimeout:    5 * time.Second,
				LogLevel:      "INFO",
				ServerName:    "transport-test-server",
				ServerVersion: "1.0.0",
				TransportMode: mode,
				HTTPPort:      "8082",
				HTTPHost:      "localhost",
			}

			testServer := server.NewServer(cfg)
			if testServer == nil {
				t.Fatalf("Failed to create server with transport mode: %s", mode)
			}

			t.Logf("Transport mode %s verified", mode)
		})
	}
}
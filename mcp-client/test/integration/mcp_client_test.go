package integration

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/bchamber/taskman/mcp-client/internal/client"
	"github.com/bchamber/taskman/mcp-client/internal/handlers"
)

func getTestMCPServerURL() string {
	if url := os.Getenv("MCP_SERVER_URL"); url != "" {
		return url
	}
	return "http://localhost:3000"
}

func TestMCPClientIntegration(t *testing.T) {
	// Skip if MCP server is not available
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Integration tests skipped")
	}
	
	// Setup logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	
	// Create MCP client
	mcpClient := client.NewMCPClient(getTestMCPServerURL(), logger)
	
	// Create intent handler
	intentHandler := handlers.NewIntentHandler(mcpClient, logger)
	
	ctx := context.Background()
	
	t.Run("ListTools", func(t *testing.T) {
		intent := `{"method": "tools/list"}`
		
		result, err := intentHandler.ProcessIntent(ctx, intent)
		if err != nil {
			t.Fatalf("Failed to list tools: %v", err)
		}
		
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		
		t.Logf("Tools result: %+v", result)
	})
	
	t.Run("ExecuteTool", func(t *testing.T) {
		intent := `{
			"method": "tools/call",
			"params": {
				"name": "get_task_overview",
				"arguments": {}
			}
		}`
		
		result, err := intentHandler.ProcessIntent(ctx, intent)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}
		
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		
		t.Logf("Tool execution result: %+v", result)
	})
	
	t.Run("ListPrompts", func(t *testing.T) {
		intent := `{"method": "prompts/list"}`
		
		result, err := intentHandler.ProcessIntent(ctx, intent)
		if err != nil {
			t.Fatalf("Failed to list prompts: %v", err)
		}
		
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		
		t.Logf("Prompts result: %+v", result)
	})
	
	t.Run("GetPrompt", func(t *testing.T) {
		intent := `{
			"method": "prompts/get",
			"params": {
				"name": "task_planning",
				"arguments": {}
			}
		}`
		
		result, err := intentHandler.ProcessIntent(ctx, intent)
		if err != nil {
			t.Fatalf("Failed to get prompt: %v", err)
		}
		
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		
		t.Logf("Prompt result: %+v", result)
	})
}

func TestMCPClientDirect(t *testing.T) {
	// Skip if MCP server is not available
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Integration tests skipped")
	}
	
	// Setup logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	
	// Create MCP client
	mcpClient := client.NewMCPClient(getTestMCPServerURL(), logger)
	
	ctx := context.Background()
	
	t.Run("DirectListTools", func(t *testing.T) {
		resp, err := mcpClient.ListTools(ctx)
		if err != nil {
			t.Fatalf("Failed to list tools: %v", err)
		}
		
		if resp == nil {
			t.Fatal("Expected response, got nil")
		}
		
		if resp.Error != nil {
			t.Fatalf("MCP server returned error: %s", resp.Error.Message)
		}
		
		if resp.Result == nil {
			t.Fatal("Expected result, got nil")
		}
		
		t.Logf("Direct tools result: %+v", resp.Result)
	})
	
	t.Run("DirectExecuteTool", func(t *testing.T) {
		resp, err := mcpClient.ExecuteTool(ctx, "get_task_overview", map[string]interface{}{})
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}
		
		if resp == nil {
			t.Fatal("Expected response, got nil")
		}
		
		if resp.Error != nil {
			t.Fatalf("MCP server returned error: %s", resp.Error.Message)
		}
		
		if resp.Result == nil {
			t.Fatal("Expected result, got nil")
		}
		
		t.Logf("Direct tool execution result: %+v", resp.Result)
	})
}
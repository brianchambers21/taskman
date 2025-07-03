package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/bchamber/taskman/mcp-client/internal/client"
	"github.com/bchamber/taskman/mcp-client/internal/config"
	"github.com/bchamber/taskman/mcp-client/internal/handlers"
)

func main() {
	// Parse command line flags
	var (
		serverURL = flag.String("server", "", "MCP server URL (overrides MCP_SERVER_URL)")
		logLevel  = flag.String("log-level", "", "Log level: debug, info, warn, error (overrides LOG_LEVEL)")
		interactive = flag.Bool("interactive", false, "Run in interactive mode")
		intent    = flag.String("intent", "", "JSON intent to process")
	)
	flag.Parse()

	// Load configuration
	cfg := config.LoadConfig()
	if *serverURL != "" {
		cfg.MCPServerURL = *serverURL
	}
	if *logLevel != "" {
		cfg.LogLevel = *logLevel
	}

	// Setup logger
	logger := setupLogger(cfg.LogLevel)
	
	// Create MCP client
	mcpClient := client.NewMCPClient(cfg.MCPServerURL, logger)
	
	// Create intent handler
	intentHandler := handlers.NewIntentHandler(mcpClient, logger)
	
	ctx := context.Background()
	
	// Handle different modes
	if *interactive {
		runInteractiveMode(ctx, intentHandler, logger)
	} else if *intent != "" {
		runSingleIntent(ctx, intentHandler, *intent, logger)
	} else {
		// Handle subcommands
		args := flag.Args()
		if len(args) == 0 {
			printUsage()
			os.Exit(1)
		}
		
		command := args[0]
		switch command {
		case "list-tools":
			runListTools(ctx, intentHandler, logger)
		case "execute-tool":
			if len(args) < 2 {
				fmt.Fprintf(os.Stderr, "Error: tool name required\n")
				os.Exit(1)
			}
			toolName := args[1]
			var toolArgs interface{}
			if len(args) > 2 {
				if err := json.Unmarshal([]byte(args[2]), &toolArgs); err != nil {
					fmt.Fprintf(os.Stderr, "Error: invalid tool arguments JSON: %v\n", err)
					os.Exit(1)
				}
			}
			runExecuteTool(ctx, intentHandler, toolName, toolArgs, logger)
		case "list-prompts":
			runListPrompts(ctx, intentHandler, logger)
		case "get-prompt":
			if len(args) < 2 {
				fmt.Fprintf(os.Stderr, "Error: prompt name required\n")
				os.Exit(1)
			}
			promptName := args[1]
			var promptArgs interface{}
			if len(args) > 2 {
				if err := json.Unmarshal([]byte(args[2]), &promptArgs); err != nil {
					fmt.Fprintf(os.Stderr, "Error: invalid prompt arguments JSON: %v\n", err)
					os.Exit(1)
				}
			}
			runGetPrompt(ctx, intentHandler, promptName, promptArgs, logger)
		default:
			fmt.Fprintf(os.Stderr, "Error: unknown command: %s\n", command)
			printUsage()
			os.Exit(1)
		}
	}
}

func setupLogger(level string) *slog.Logger {
	var logLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}
	
	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	
	return slog.New(slog.NewTextHandler(os.Stderr, opts))
}

func runInteractiveMode(ctx context.Context, handler *handlers.IntentHandler, logger *slog.Logger) {
	fmt.Println("MCP Client Interactive Mode")
	fmt.Println("Enter JSON intents (press Ctrl+D to exit):")
	fmt.Println()
	
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		
		result, err := handler.ProcessIntent(ctx, line)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		
		output, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Printf("Error formatting result: %v\n", err)
			continue
		}
		
		fmt.Printf("Result:\n%s\n\n", output)
	}
}

func runSingleIntent(ctx context.Context, handler *handlers.IntentHandler, intentJSON string, logger *slog.Logger) {
	result, err := handler.ProcessIntent(ctx, intentJSON)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	
	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting result: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("%s\n", output)
}

func runListTools(ctx context.Context, handler *handlers.IntentHandler, logger *slog.Logger) {
	intent := `{"method": "tools/list"}`
	runSingleIntent(ctx, handler, intent, logger)
}

func runExecuteTool(ctx context.Context, handler *handlers.IntentHandler, toolName string, args interface{}, logger *slog.Logger) {
	intentData := map[string]interface{}{
		"method": "tools/call",
		"params": map[string]interface{}{
			"name":      toolName,
			"arguments": args,
		},
	}
	
	intentJSON, err := json.Marshal(intentData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating intent: %v\n", err)
		os.Exit(1)
	}
	
	runSingleIntent(ctx, handler, string(intentJSON), logger)
}

func runListPrompts(ctx context.Context, handler *handlers.IntentHandler, logger *slog.Logger) {
	intent := `{"method": "prompts/list"}`
	runSingleIntent(ctx, handler, intent, logger)
}

func runGetPrompt(ctx context.Context, handler *handlers.IntentHandler, promptName string, args interface{}, logger *slog.Logger) {
	intentData := map[string]interface{}{
		"method": "prompts/get",
		"params": map[string]interface{}{
			"name":      promptName,
			"arguments": args,
		},
	}
	
	intentJSON, err := json.Marshal(intentData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating intent: %v\n", err)
		os.Exit(1)
	}
	
	runSingleIntent(ctx, handler, string(intentJSON), logger)
}

func printUsage() {
	fmt.Printf(`MCP Client - Model Context Protocol client

Usage:
  %s [flags] <command> [args...]
  %s [flags] -intent '<json>'
  %s [flags] -interactive

Commands:
  list-tools                     List available tools
  execute-tool <name> [args]     Execute a tool with optional JSON arguments
  list-prompts                   List available prompts  
  get-prompt <name> [args]       Get a prompt with optional JSON arguments

Flags:
  -server <url>                  MCP server URL (default: $MCP_SERVER_URL or http://localhost:3000)
  -log-level <level>             Log level: debug, info, warn, error (default: info)
  -intent '<json>'               Process a single JSON intent
  -interactive                   Run in interactive mode

Examples:
  %s list-tools
  %s execute-tool get_task_overview
  %s execute-tool create_task_with_context '{"task_name": "Test", "description": "Test task"}'
  %s -intent '{"method": "tools/list"}'
  %s -interactive

Environment Variables:
  MCP_SERVER_URL                 Default MCP server URL
  LOG_LEVEL                      Default log level
`, os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0])
}
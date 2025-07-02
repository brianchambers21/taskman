package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

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
	
	// Create MCP server instance
	mcpServer := mcp.NewServer(cfg.ServerName, cfg.ServerVersion, nil)
	
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
	
	// Register tools (will be implemented in future tasks)
	server.registerTools()
	
	slog.Info("MCP server created successfully")
	return server
}

func (s *Server) registerTools() {
	slog.Info("Registering MCP tools")
	// TODO: Register tools in future tasks
	slog.Info("Tools registration completed (none yet)")
}

func (s *Server) setupHTTPServer() {
	mux := http.NewServeMux()
	
	// Set up SSE endpoint
	mux.HandleFunc("/sse", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Handling SSE connection", "remote_addr", r.RemoteAddr)
		
		// Create SSE transport for this connection
		sseTransport := mcp.NewSSEServerTransport("/sse", w)
		
		// Run MCP server with this transport
		ctx := r.Context()
		if err := s.mcpServer.Run(ctx, sseTransport); err != nil {
			slog.Error("SSE MCP server failed", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	})
	
	// Set up JSON-RPC endpoint for POST requests
	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		slog.Info("Handling MCP JSON-RPC request", "remote_addr", r.RemoteAddr)
		
		// Create SSE transport for JSON-RPC
		sseTransport := mcp.NewSSEServerTransport("/mcp", w)
		
		// Run MCP server with this transport
		ctx := r.Context()
		if err := s.mcpServer.Run(ctx, sseTransport); err != nil {
			slog.Error("JSON-RPC MCP server failed", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	})
	
	addr := fmt.Sprintf("%s:%s", s.config.HTTPHost, s.config.HTTPPort)
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	
	slog.Info("HTTP server configured", "address", addr)
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
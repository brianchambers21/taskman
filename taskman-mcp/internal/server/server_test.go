package server

import (
	"testing"
	"time"

	"github.com/bchamber/taskman-mcp/internal/config"
)

func TestNewServer(t *testing.T) {
	tests := []struct {
		name          string
		transportMode string
		expectHTTP    bool
	}{
		{
			name:          "stdio transport",
			transportMode: "stdio",
			expectHTTP:    false,
		},
		{
			name:          "http transport",
			transportMode: "http",
			expectHTTP:    true,
		},
		{
			name:          "both transports",
			transportMode: "both",
			expectHTTP:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				APIBaseURL:    "http://localhost:8080",
				APITimeout:    30 * time.Second,
				LogLevel:      "INFO",
				ServerName:    "test-server",
				ServerVersion: "1.0.0",
				TransportMode: tt.transportMode,
				HTTPPort:      "8081",
				HTTPHost:      "localhost",
			}

			server := NewServer(cfg)

			if server == nil {
				t.Fatal("Expected server to be created, got nil")
			}

			if server.mcpServer == nil {
				t.Error("Expected mcpServer to be initialized")
			}

			if server.apiClient == nil {
				t.Error("Expected apiClient to be initialized")
			}

			if server.config != cfg {
				t.Error("Expected config to be set correctly")
			}

			if tt.expectHTTP && server.httpServer == nil {
				t.Error("Expected httpServer to be initialized for HTTP transport")
			}

			if !tt.expectHTTP && server.httpServer != nil {
				t.Error("Expected httpServer to be nil for stdio-only transport")
			}
		})
	}
}

func TestServer_RegisterTools(t *testing.T) {
	cfg := &config.Config{
		APIBaseURL:    "http://localhost:8080",
		APITimeout:    30 * time.Second,
		LogLevel:      "INFO",
		ServerName:    "test-server",
		ServerVersion: "1.0.0",
		TransportMode: "stdio",
		HTTPPort:      "8081",
		HTTPHost:      "localhost",
	}

	server := NewServer(cfg)

	// Test that registerTools doesn't panic or error
	// In the current implementation, it just logs since no tools are implemented yet
	server.registerTools()

	// This test mainly ensures the method exists and can be called safely
	// More comprehensive testing will be added when tools are implemented in future tasks
}
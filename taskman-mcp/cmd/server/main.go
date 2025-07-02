package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/bchamber/taskman-mcp/internal/config"
	"github.com/bchamber/taskman-mcp/internal/server"
)

func main() {
	// Load configuration
	cfg := config.Load()
	
	// Set up structured logging
	setupLogging(cfg.LogLevel)
	
	slog.Info("Starting Taskman MCP Server",
		"server_name", cfg.ServerName,
		"server_version", cfg.ServerVersion,
		"api_base_url", cfg.APIBaseURL,
	)
	
	// Create server
	mcpServer := server.NewServer(cfg)
	
	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Handle shutdown signals
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		
		slog.Info("Shutdown signal received, stopping server...")
		cancel()
	}()
	
	// Run server
	if err := mcpServer.Run(ctx); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
	
	slog.Info("Server stopped gracefully")
}

func setupLogging(level string) {
	var logLevel slog.Level
	switch level {
	case "DEBUG":
		logLevel = slog.LevelDebug
	case "INFO":
		logLevel = slog.LevelInfo
	case "WARN":
		logLevel = slog.LevelWarn
	case "ERROR":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}
	
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	})
	
	logger := slog.New(handler)
	slog.SetDefault(logger)
	
	slog.Info("Logging initialized", "level", level)
}
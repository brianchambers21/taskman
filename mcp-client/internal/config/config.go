package config

import (
	"os"
)

// Config holds the configuration for the MCP client
type Config struct {
	MCPServerURL string
	LogLevel     string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() Config {
	return Config{
		MCPServerURL: getEnv("MCP_SERVER_URL", "http://localhost:3000"),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
	}
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
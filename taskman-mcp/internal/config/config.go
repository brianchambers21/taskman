package config

import (
	"log/slog"
	"os"
	"time"
)

type Config struct {
	APIBaseURL    string
	APITimeout    time.Duration
	LogLevel      string
	ServerName    string
	ServerVersion string

	// Transport configuration
	TransportMode string // "stdio", "http", "both"
	HTTPPort      string
	HTTPHost      string
}

func Load() *Config {
	slog.Info("Loading MCP server configuration")

	config := &Config{
		APIBaseURL:    getEnv("TASKMAN_API_BASE_URL", "http://localhost:8080"),
		APITimeout:    getEnvDuration("TASKMAN_API_TIMEOUT", 30*time.Second),
		LogLevel:      getEnv("TASKMAN_LOG_LEVEL", "INFO"),
		ServerName:    getEnv("TASKMAN_MCP_SERVER_NAME", "taskman-mcp"),
		ServerVersion: getEnv("TASKMAN_MCP_SERVER_VERSION", "1.0.0"),

		TransportMode: getEnv("TASKMAN_MCP_TRANSPORT", "stdio"),
		HTTPPort:      getEnv("TASKMAN_MCP_HTTP_PORT", "8081"),
		HTTPHost:      getEnv("TASKMAN_MCP_HTTP_HOST", "localhost"),
	}

	slog.Info("MCP server configuration loaded",
		"api_base_url", config.APIBaseURL,
		"api_timeout", config.APITimeout,
		"log_level", config.LogLevel,
		"server_name", config.ServerName,
		"server_version", config.ServerVersion,
		"transport_mode", config.TransportMode,
		"http_port", config.HTTPPort,
		"http_host", config.HTTPHost,
	)

	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
		slog.Warn("Invalid duration in environment variable, using default",
			"key", key,
			"value", value,
			"default", defaultValue,
		)
	}
	return defaultValue
}

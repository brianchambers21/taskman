package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected *Config
	}{
		{
			name:    "default configuration",
			envVars: map[string]string{},
			expected: &Config{
				APIBaseURL:    "http://localhost:8080",
				APITimeout:    30 * time.Second,
				LogLevel:      "INFO",
				ServerName:    "taskman-mcp",
				ServerVersion: "1.0.0",
				TransportMode: "stdio",
				HTTPPort:      "8081",
				HTTPHost:      "localhost",
			},
		},
		{
			name: "custom configuration",
			envVars: map[string]string{
				"TASKMAN_API_BASE_URL":       "http://api.example.com:9000",
				"TASKMAN_API_TIMEOUT":        "60s",
				"TASKMAN_LOG_LEVEL":          "DEBUG",
				"TASKMAN_MCP_SERVER_NAME":    "custom-mcp",
				"TASKMAN_MCP_SERVER_VERSION": "2.0.0",
				"TASKMAN_MCP_TRANSPORT":      "http",
				"TASKMAN_MCP_HTTP_PORT":      "9001",
				"TASKMAN_MCP_HTTP_HOST":      "0.0.0.0",
			},
			expected: &Config{
				APIBaseURL:    "http://api.example.com:9000",
				APITimeout:    60 * time.Second,
				LogLevel:      "DEBUG",
				ServerName:    "custom-mcp",
				ServerVersion: "2.0.0",
				TransportMode: "http",
				HTTPPort:      "9001",
				HTTPHost:      "0.0.0.0",
			},
		},
		{
			name: "invalid timeout falls back to default",
			envVars: map[string]string{
				"TASKMAN_API_TIMEOUT": "invalid",
			},
			expected: &Config{
				APIBaseURL:    "http://localhost:8080",
				APITimeout:    30 * time.Second,
				LogLevel:      "INFO",
				ServerName:    "taskman-mcp",
				ServerVersion: "1.0.0",
				TransportMode: "stdio",
				HTTPPort:      "8081",
				HTTPHost:      "localhost",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment
			originalEnv := make(map[string]string)
			for key := range tt.envVars {
				originalEnv[key] = os.Getenv(key)
			}

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Restore environment after test
			defer func() {
				for key, originalValue := range originalEnv {
					if originalValue == "" {
						os.Unsetenv(key)
					} else {
						os.Setenv(key, originalValue)
					}
				}
			}()

			// Test configuration loading
			config := Load()

			// Verify results
			if config.APIBaseURL != tt.expected.APIBaseURL {
				t.Errorf("Expected APIBaseURL %s, got %s", tt.expected.APIBaseURL, config.APIBaseURL)
			}
			if config.APITimeout != tt.expected.APITimeout {
				t.Errorf("Expected APITimeout %v, got %v", tt.expected.APITimeout, config.APITimeout)
			}
			if config.LogLevel != tt.expected.LogLevel {
				t.Errorf("Expected LogLevel %s, got %s", tt.expected.LogLevel, config.LogLevel)
			}
			if config.ServerName != tt.expected.ServerName {
				t.Errorf("Expected ServerName %s, got %s", tt.expected.ServerName, config.ServerName)
			}
			if config.ServerVersion != tt.expected.ServerVersion {
				t.Errorf("Expected ServerVersion %s, got %s", tt.expected.ServerVersion, config.ServerVersion)
			}
			if config.TransportMode != tt.expected.TransportMode {
				t.Errorf("Expected TransportMode %s, got %s", tt.expected.TransportMode, config.TransportMode)
			}
			if config.HTTPPort != tt.expected.HTTPPort {
				t.Errorf("Expected HTTPPort %s, got %s", tt.expected.HTTPPort, config.HTTPPort)
			}
			if config.HTTPHost != tt.expected.HTTPHost {
				t.Errorf("Expected HTTPHost %s, got %s", tt.expected.HTTPHost, config.HTTPHost)
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		value        string
		defaultValue string
		expected     string
	}{
		{
			name:         "environment variable exists",
			key:          "TEST_VAR_EXISTS",
			value:        "test_value",
			defaultValue: "default",
			expected:     "test_value",
		},
		{
			name:         "environment variable does not exist",
			key:          "TEST_VAR_NOT_EXISTS",
			value:        "",
			defaultValue: "default",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original value
			originalValue := os.Getenv(tt.key)
			defer func() {
				if originalValue == "" {
					os.Unsetenv(tt.key)
				} else {
					os.Setenv(tt.key, originalValue)
				}
			}()

			// Set test value
			if tt.value != "" {
				os.Setenv(tt.key, tt.value)
			} else {
				os.Unsetenv(tt.key)
			}

			// Test function
			result := getEnv(tt.key, tt.defaultValue)

			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGetEnvDuration(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		value        string
		defaultValue time.Duration
		expected     time.Duration
	}{
		{
			name:         "valid duration",
			key:          "TEST_DURATION_VALID",
			value:        "45s",
			defaultValue: 30 * time.Second,
			expected:     45 * time.Second,
		},
		{
			name:         "invalid duration uses default",
			key:          "TEST_DURATION_INVALID",
			value:        "invalid",
			defaultValue: 30 * time.Second,
			expected:     30 * time.Second,
		},
		{
			name:         "missing duration uses default",
			key:          "TEST_DURATION_MISSING",
			value:        "",
			defaultValue: 30 * time.Second,
			expected:     30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original value
			originalValue := os.Getenv(tt.key)
			defer func() {
				if originalValue == "" {
					os.Unsetenv(tt.key)
				} else {
					os.Setenv(tt.key, originalValue)
				}
			}()

			// Set test value
			if tt.value != "" {
				os.Setenv(tt.key, tt.value)
			} else {
				os.Unsetenv(tt.key)
			}

			// Test function
			result := getEnvDuration(tt.key, tt.defaultValue)

			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
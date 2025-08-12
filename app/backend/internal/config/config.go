package config

import (
	"os"
)

// Config holds all configuration for the SOHOAAS backend
type Config struct {
	Port         string
	Environment  string
	LogLevel     string
	WorkflowsDir string
	OpenAI       OpenAIConfig
	MCP          MCPConfig
	OAuth2       OAuth2Config
	Genkit       GenkitConfig
}

// OpenAIConfig holds OpenAI-specific configuration
type OpenAIConfig struct {
	APIKey string
}

// MCPConfig holds MCP service configuration
type MCPConfig struct {
	BaseURL      string
	AuthEndpoint string
}

// OAuth2Config holds OAuth2 configuration
type OAuth2Config struct {
	GoogleClientID     string
	GoogleClientSecret string
}

// GenkitConfig holds Genkit-specific configuration
type GenkitConfig struct {
	Environment string
}

// New creates a new configuration instance from environment variables
func New() *Config {
	return &Config{
		Port:         getEnv("PORT", "8080"),
		Environment:  getEnv("ENVIRONMENT", "development"),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
		WorkflowsDir: getEnv("WORKFLOWS_DIR", "./generated_workflows"),
		OpenAI: OpenAIConfig{
			APIKey: getEnv("GOOGLE_API_KEY", ""),
		},
		MCP: MCPConfig{
			BaseURL:      getEnv("MCP_SERVICE_URL", "http://localhost:8080"),
			AuthEndpoint: getEnv("MCP_AUTH_ENDPOINT", "/api/auth/token"),
		},
		OAuth2: OAuth2Config{
			GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
			GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		},
		Genkit: GenkitConfig{
			Environment: getEnv("GENKIT_ENV", "dev"),
		},
	}
}

// getEnv gets an environment variable with a default fallback
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

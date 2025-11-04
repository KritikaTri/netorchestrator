package config

import (
	"os"

	"github.com/joho/godotenv"
)

// OpenAIConfig holds OpenAI API configuration
type OpenAIConfig struct {
	APIKey string `json:"api_key" yaml:"api_key"`
	Model  string `json:"model" yaml:"model"`
}

// LoadOpenAIConfig loads OpenAI configuration from environment and .env file
func LoadOpenAIConfig() *OpenAIConfig {
	// Try to load .env file if it exists
	_ = godotenv.Load() // Ignore errors - .env is optional

	config := &OpenAIConfig{
		APIKey: os.Getenv("OPENAI_API_KEY"),
		Model:  os.Getenv("OPENAI_MODEL"),
	}

	// Set defaults if not specified
	if config.Model == "" {
		config.Model = "gpt-4"
	}

	return config
}

// IsConfigured returns true if OpenAI is properly configured
func (c *OpenAIConfig) IsConfigured() bool {
	return c.APIKey != ""
}

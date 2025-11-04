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

// AIConfig holds AI configuration for multiple providers
type AIConfig struct {
	Provider    string  `json:"provider" yaml:"provider"` // "openai", "gemini", "auto"
	OpenAIKey   string  `json:"openai_key" yaml:"openai_key"`
	GeminiKey   string  `json:"gemini_key" yaml:"gemini_key"`
	Model       string  `json:"model" yaml:"model"`
	Temperature float64 `json:"temperature" yaml:"temperature"`
	MaxTokens   int     `json:"max_tokens" yaml:"max_tokens"`
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
		config.Model = "gpt-3.5-turbo"
	}

	return config
}

// LoadAIConfig loads AI configuration supporting multiple providers
func LoadAIConfig() *AIConfig {
	// Try to load .env file if it exists
	_ = godotenv.Load() // Ignore errors - .env is optional

	config := &AIConfig{
		Provider:    os.Getenv("AI_PROVIDER"),
		OpenAIKey:   os.Getenv("OPENAI_API_KEY"),
		GeminiKey:   os.Getenv("GEMINI_API_KEY"),
		Model:       os.Getenv("AI_MODEL"),
		Temperature: 0.7,
		MaxTokens:   2000,
	}

	// Auto-detect provider if not specified
	if config.Provider == "" {
		if config.GeminiKey != "" {
			config.Provider = "gemini"
			config.Model = "gemini-pro"
		} else if config.OpenAIKey != "" {
			config.Provider = "openai"
			config.Model = "gpt-3.5-turbo"
		} else {
			config.Provider = "mock"
			config.Model = "mock-model"
		}
	}

	// Set model defaults based on provider
	if config.Model == "" {
		switch config.Provider {
		case "gemini":
			config.Model = "gemini-pro"
		case "openai":
			config.Model = "gpt-3.5-turbo"
		default:
			config.Model = "mock-model"
		}
	}

	return config
}

// IsConfigured returns true if AI is properly configured
func (c *AIConfig) IsConfigured() bool {
	switch c.Provider {
	case "gemini":
		return c.GeminiKey != ""
	case "openai":
		return c.OpenAIKey != ""
	default:
		return false
	}
}

// IsConfigured returns true if OpenAI is properly configured
func (c *OpenAIConfig) IsConfigured() bool {
	return c.APIKey != ""
}

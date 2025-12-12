package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

const (
	defaultGeminiModel = "gemini-2.5-flash"
)

// Config holds all application configuration
type Config struct {
	GitHubUser        string
	GitHubToken       string
	GeminiAPIKey      string
	GeminiModel       string
	DiscordWebhookURL string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Try to load .env file (ignore error if a file doesn't exist)
	_ = godotenv.Load() //nolint:errcheck // .env file is optional

	geminiModel := os.Getenv("GEMINI_MODEL")
	if geminiModel == "" {
		geminiModel = defaultGeminiModel
	}

	cfg := &Config{
		GitHubUser:        os.Getenv("GH_USER"),
		GitHubToken:       os.Getenv("GH_TOKEN"),
		GeminiAPIKey:      os.Getenv("GEMINI_API_KEY"),
		GeminiModel:       geminiModel,
		DiscordWebhookURL: os.Getenv("DISCORD_WEBHOOK_URL"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validate checks if all required configuration is present
func (c *Config) validate() error {
	if c.GitHubUser == "" {
		return fmt.Errorf("GH_USER environment variable is required")
	}
	if c.GitHubToken == "" {
		return fmt.Errorf("GH_TOKEN environment variable is required")
	}
	if c.GeminiAPIKey == "" {
		return fmt.Errorf("GEMINI_API_KEY environment variable is required")
	}
	if c.DiscordWebhookURL == "" {
		return fmt.Errorf("DISCORD_WEBHOOK_URL environment variable is required")
	}
	return nil
}

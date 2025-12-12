package config

import (
	"testing"
)

func TestLoad_Success(t *testing.T) {
	// Set up test environment variables
	t.Setenv("GH_USER", "testuser")
	t.Setenv("GH_TOKEN", "testtoken")
	t.Setenv("GEMINI_API_KEY", "testkey")
	t.Setenv("DISCORD_WEBHOOK_URL", "https://discord.com/webhook")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if cfg.GitHubUser != "testuser" {
		t.Errorf("Expected GitHubUser=testuser, got: %s", cfg.GitHubUser)
	}
	if cfg.GitHubToken != "testtoken" {
		t.Errorf("Expected GitHubToken=testtoken, got: %s", cfg.GitHubToken)
	}
	if cfg.GeminiAPIKey != "testkey" {
		t.Errorf("Expected GeminiAPIKey=testkey, got: %s", cfg.GeminiAPIKey)
	}
	if cfg.GeminiModel != "gemini-2.5-flash" {
		t.Errorf("Expected GeminiModel=gemini-2.5-flash (default), got: %s", cfg.GeminiModel)
	}
	if cfg.DiscordWebhookURL != "https://discord.com/webhook" {
		t.Errorf(
			"Expected DiscordWebhookURL=https://discord.com/webhook, got: %s",
			cfg.DiscordWebhookURL,
		)
	}
}

func TestLoad_CustomGeminiModel(t *testing.T) {
	t.Setenv("GH_USER", "testuser")
	t.Setenv("GH_TOKEN", "testtoken")
	t.Setenv("GEMINI_API_KEY", "testkey")
	t.Setenv("GEMINI_MODEL", "gemini-1.5-pro")
	t.Setenv("DISCORD_WEBHOOK_URL", "https://discord.com/webhook")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if cfg.GeminiModel != "gemini-1.5-pro" {
		t.Errorf("Expected GeminiModel=gemini-1.5-pro, got: %s", cfg.GeminiModel)
	}
}

func TestLoad_MissingGitHubUser(t *testing.T) {
	t.Setenv("GH_TOKEN", "testtoken")
	t.Setenv("GEMINI_API_KEY", "testkey")
	t.Setenv("DISCORD_WEBHOOK_URL", "https://discord.com/webhook")

	_, err := Load()
	if err == nil {
		t.Fatal("Expected error for missing GH_USER, got nil")
	}

	expected := "GH_USER environment variable is required"
	if err.Error() != expected {
		t.Errorf("Expected error: %s, got: %s", expected, err.Error())
	}
}

func TestLoad_MissingGitHubToken(t *testing.T) {
	t.Setenv("GH_USER", "testuser")
	t.Setenv("GEMINI_API_KEY", "testkey")
	t.Setenv("DISCORD_WEBHOOK_URL", "https://discord.com/webhook")

	_, err := Load()
	if err == nil {
		t.Fatal("Expected error for missing GH_TOKEN, got nil")
	}

	expected := "GH_TOKEN environment variable is required"
	if err.Error() != expected {
		t.Errorf("Expected error: %s, got: %s", expected, err.Error())
	}
}

func TestLoad_MissingGeminiAPIKey(t *testing.T) {
	t.Setenv("GH_USER", "testuser")
	t.Setenv("GH_TOKEN", "testtoken")
	t.Setenv("DISCORD_WEBHOOK_URL", "https://discord.com/webhook")

	_, err := Load()
	if err == nil {
		t.Fatal("Expected error for missing GEMINI_API_KEY, got nil")
	}

	expected := "GEMINI_API_KEY environment variable is required"
	if err.Error() != expected {
		t.Errorf("Expected error: %s, got: %s", expected, err.Error())
	}
}

func TestLoad_MissingDiscordWebhookURL(t *testing.T) {
	t.Setenv("GH_USER", "testuser")
	t.Setenv("GH_TOKEN", "testtoken")
	t.Setenv("GEMINI_API_KEY", "testkey")

	_, err := Load()
	if err == nil {
		t.Fatal("Expected error for missing DISCORD_WEBHOOK_URL, got nil")
	}

	expected := "DISCORD_WEBHOOK_URL environment variable is required"
	if err.Error() != expected {
		t.Errorf("Expected error: %s, got: %s", expected, err.Error())
	}
}

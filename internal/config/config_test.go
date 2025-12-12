package config

import (
	"os"
	"testing"
)

func TestLoad_Success(t *testing.T) {
	// Set up test environment variables
	os.Setenv("GH_USER", "testuser")
	os.Setenv("GH_TOKEN", "testtoken")
	os.Setenv("GEMINI_API_KEY", "testkey")
	os.Setenv("DISCORD_WEBHOOK_URL", "https://discord.com/webhook")
	defer cleanupEnv()

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
	if cfg.DiscordWebhookURL != "https://discord.com/webhook" {
		t.Errorf(
			"Expected DiscordWebhookURL=https://discord.com/webhook, got: %s",
			cfg.DiscordWebhookURL,
		)
	}
}

func TestLoad_MissingGitHubUser(t *testing.T) {
	os.Setenv("GH_TOKEN", "testtoken")
	os.Setenv("GEMINI_API_KEY", "testkey")
	os.Setenv("DISCORD_WEBHOOK_URL", "https://discord.com/webhook")
	defer cleanupEnv()

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
	os.Setenv("GH_USER", "testuser")
	os.Setenv("GEMINI_API_KEY", "testkey")
	os.Setenv("DISCORD_WEBHOOK_URL", "https://discord.com/webhook")
	defer cleanupEnv()

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
	os.Setenv("GH_USER", "testuser")
	os.Setenv("GH_TOKEN", "testtoken")
	os.Setenv("DISCORD_WEBHOOK_URL", "https://discord.com/webhook")
	defer cleanupEnv()

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
	os.Setenv("GH_USER", "testuser")
	os.Setenv("GH_TOKEN", "testtoken")
	os.Setenv("GEMINI_API_KEY", "testkey")
	defer cleanupEnv()

	_, err := Load()
	if err == nil {
		t.Fatal("Expected error for missing DISCORD_WEBHOOK_URL, got nil")
	}

	expected := "DISCORD_WEBHOOK_URL environment variable is required"
	if err.Error() != expected {
		t.Errorf("Expected error: %s, got: %s", expected, err.Error())
	}
}

func cleanupEnv() {
	os.Unsetenv("GH_USER")
	os.Unsetenv("GH_TOKEN")
	os.Unsetenv("GEMINI_API_KEY")
	os.Unsetenv("DISCORD_WEBHOOK_URL")
}

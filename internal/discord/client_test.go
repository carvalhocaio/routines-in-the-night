package discord

import (
	"errors"
	"testing"
)

func TestNewClient(t *testing.T) {
	webhookURL := "https://discord.com/api/webhooks/123/abc"
	client := NewClient(webhookURL)

	if client.webhookURL != webhookURL {
		t.Errorf("Expected webhookURL=%s, got: %s", webhookURL, client.webhookURL)
	}
	if client.httpClient == nil {
		t.Error("Expected httpClient to be initialized")
	}
}

func TestEmbedStructure(t *testing.T) {
	embed := Embed{
		Title:       "Test Title",
		Description: "Test Description",
		Color:       colorBlue,
		Timestamp:   "2024-01-01T00:00:00Z",
		Footer: &EmbedFooter{
			Text: "Test Footer",
		},
	}

	if embed.Title != "Test Title" {
		t.Errorf("Expected Title='Test Title', got: %s", embed.Title)
	}
	if embed.Description != "Test Description" {
		t.Errorf("Expected Description='Test Description', got: %s", embed.Description)
	}
	if embed.Color != colorBlue {
		t.Errorf("Expected Color=%d, got: %d", colorBlue, embed.Color)
	}
	if embed.Footer.Text != "Test Footer" {
		t.Errorf("Expected Footer.Text='Test Footer', got: %s", embed.Footer.Text)
	}
}

func TestWebhookPayload(t *testing.T) {
	embed := Embed{
		Title:       "Daily Report",
		Description: "Test report",
		Color:       colorBlue,
	}

	payload := WebhookPayload{
		Embeds: []Embed{embed},
	}

	if len(payload.Embeds) != 1 {
		t.Errorf("Expected 1 embed, got: %d", len(payload.Embeds))
	}
	if payload.Embeds[0].Title != "Daily Report" {
		t.Errorf("Expected Title='Daily Report', got: %s", payload.Embeds[0].Title)
	}
}

func TestErrorMessage(t *testing.T) {
	testErr := errors.New("test error message")

	embed := Embed{
		Title:       "GitHub Daily Reporter - Error",
		Description: testErr.Error(),
		Color:       0xFF0000,
	}

	expectedDesc := "test error message"
	if embed.Description != expectedDesc {
		t.Errorf("Expected Description='%s', got: %s", expectedDesc, embed.Description)
	}
	if embed.Color != 0xFF0000 {
		t.Errorf("Expected red color (0xFF0000), got: %d", embed.Color)
	}
}

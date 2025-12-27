package discord

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestTruncateMessage(t *testing.T) {
	tests := []struct {
		name      string
		message   string
		maxLength int
		expected  string
	}{
		{
			name:      "short message",
			message:   "This is a short message.",
			maxLength: 100,
			expected:  "This is a short message.",
		},
		{
			name:      "exactly at limit",
			message:   "12345",
			maxLength: 5,
			expected:  "12345",
		},
		{
			name:      "over limit with period",
			message:   "First sentence. Second sentence. Third sentence.",
			maxLength: 35,
			expected:  "First sentence. Second sentence.",
		},
		{
			name:      "over limit no period",
			message:   "This is a long message without periods",
			maxLength: 20,
			expected:  "This is a long messa",
		},
		{
			name:      "period at very beginning",
			message:   ".rest of message is long",
			maxLength: 10,
			expected:  ".rest of m", // period at index 0 is not > 0, so truncates at limit
		},
		{
			name:      "empty message",
			message:   "",
			maxLength: 100,
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateMessage(tt.message, tt.maxLength)
			if result != tt.expected {
				t.Errorf("truncateMessage(%q, %d) = %q, expected %q",
					tt.message, tt.maxLength, result, tt.expected)
			}
		})
	}
}

func TestSendDailyReport_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Verify payload
		var payload WebhookPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("Failed to decode payload: %v", err)
		}

		if len(payload.Embeds) != 1 {
			t.Errorf("Expected 1 embed, got %d", len(payload.Embeds))
		}

		embed := payload.Embeds[0]
		if !strings.HasPrefix(embed.Title, "GitHub Daily - ") {
			t.Errorf("Expected title to start with 'GitHub Daily - ', got %s", embed.Title)
		}
		if embed.Description != "Test message" {
			t.Errorf("Expected description 'Test message', got %s", embed.Description)
		}
		if embed.Color != colorBlue {
			t.Errorf("Expected color %d, got %d", colorBlue, embed.Color)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.SendDailyReport("Test message")

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestSendDailyReport_LongMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload WebhookPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("Failed to decode payload: %v", err)
		}

		// Verify message was truncated
		if len(payload.Embeds[0].Description) > maxDescriptionLen {
			t.Errorf("Description exceeds max length: %d > %d",
				len(payload.Embeds[0].Description), maxDescriptionLen)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	// Create a message longer than maxDescriptionLen
	longMessage := strings.Repeat("This is a sentence. ", 300)
	client := NewClient(server.URL)
	err := client.SendDailyReport(longMessage)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestSendDailyReport_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.SendDailyReport("Test message")

	if err == nil {
		t.Error("Expected error for server error response")
	}
	if !strings.Contains(err.Error(), "unexpected status code: 500") {
		t.Errorf("Expected status code error, got: %v", err)
	}
}

func TestSendDailyReport_NetworkError(t *testing.T) {
	client := NewClient("http://localhost:99999")
	err := client.SendDailyReport("Test message")

	if err == nil {
		t.Error("Expected error for network failure")
	}
	if !strings.Contains(err.Error(), "failed to send request") {
		t.Errorf("Expected network error, got: %v", err)
	}
}

func TestSendError_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload WebhookPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("Failed to decode payload: %v", err)
		}

		embed := payload.Embeds[0]
		if embed.Title != "GitHub Daily Reporter - Error" {
			t.Errorf("Expected error title, got %s", embed.Title)
		}
		if !strings.Contains(embed.Description, "test error") {
			t.Errorf("Expected error description to contain 'test error', got %s", embed.Description)
		}
		if embed.Color != 0xFF0000 {
			t.Errorf("Expected red color, got %d", embed.Color)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.SendError(errors.New("test error"))

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestSendError_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.SendError(errors.New("test error"))

	if err == nil {
		t.Error("Expected error for bad request response")
	}
}

func TestEmbedStructure(t *testing.T) {
	timestamp := "2024-01-01T00:00:00Z"
	embed := Embed{
		Title:       "Test Title",
		Description: "Test Description",
		Color:       colorBlue,
		Timestamp:   timestamp,
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
	if embed.Timestamp != timestamp {
		t.Errorf("Expected Timestamp='%s', got: %s", timestamp, embed.Timestamp)
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

func TestSendEmbed_InvalidURL(t *testing.T) {
	client := NewClient("://invalid-url")
	err := client.SendDailyReport("Test")

	if err == nil {
		t.Error("Expected error for invalid URL")
	}
	if !strings.Contains(err.Error(), "failed to create request") {
		t.Errorf("Expected request creation error, got: %v", err)
	}
}

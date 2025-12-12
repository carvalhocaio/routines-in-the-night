package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	colorBlue         = 0x7289DA
	maxDescriptionLen = 4096 // Discord embed description limit
)

// Client handles Discord webhook interactions
type Client struct {
	webhookURL string
	httpClient *http.Client
}

// Embed represents a Discord embed message
type Embed struct {
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Color       int          `json:"color"`
	Timestamp   string       `json:"timestamp"`
	Footer      *EmbedFooter `json:"footer,omitempty"`
}

// EmbedFooter represents the footer of a Discord embed
type EmbedFooter struct {
	Text string `json:"text"`
}

// WebhookPayload represents the payload sent to Discord
type WebhookPayload struct {
	Embeds []Embed `json:"embeds"`
}

// NewClient creates a new Discord webhook client
func NewClient(webhookURL string) *Client {
	return &Client{
		webhookURL: webhookURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendDailyReport sends the daily GitHub report to Discord
func (c *Client) SendDailyReport(message string) error {
	embed := Embed{
		Title:       "GitHub Daily",
		Description: truncateMessage(message, maxDescriptionLen),
		Color:       colorBlue,
		Timestamp:   time.Now().Format(time.RFC3339),
		Footer: &EmbedFooter{
			Text: "GitHub Daily Reporter",
		},
	}

	return c.sendEmbed(embed)
}

// truncateMessage ensures the message fits within Discord's limits
// while preserving complete sentences
func truncateMessage(message string, maxLength int) string {
	if len(message) <= maxLength {
		return message
	}

	// Find the last period before the limit
	truncated := message[:maxLength]
	lastPeriod := -1
	for i := len(truncated) - 1; i >= 0; i-- {
		if truncated[i] == '.' {
			lastPeriod = i
			break
		}
	}

	if lastPeriod > 0 {
		return message[:lastPeriod+1]
	}

	// No period found, truncate at limit
	return truncated
}

// SendError sends an error message to Discord
func (c *Client) SendError(err error) error {
	embed := Embed{
		Title:       "GitHub Daily Reporter - Error",
		Description: fmt.Sprintf("Error occurred: %v", err),
		Color:       0xFF0000, // Red color
		Timestamp:   time.Now().Format(time.RFC3339),
		Footer: &EmbedFooter{
			Text: "GitHub Daily Reporter",
		},
	}

	return c.sendEmbed(embed)
}

// sendEmbed sends an embed to Discord via webhook
func (c *Client) sendEmbed(embed Embed) error {
	payload := WebhookPayload{
		Embeds: []Embed{embed},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		c.webhookURL,
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close() //nolint:errcheck // defer close is best effort
	}()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf(
			"unexpected status code: %d",
			resp.StatusCode,
		)
	}

	return nil
}

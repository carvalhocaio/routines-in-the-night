package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	colorBlue = 0x7289DA
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
		Description: message,
		Color:       colorBlue,
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf(
			"unexpected status code: %d",
			resp.StatusCode,
		)
	}

	return nil
}

package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/carvalhocaio/routines-in-the-night/internal/github"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

const (
	maxTokens       = 8192
	temperature     = 1.2
	maxSummaryChars = 4096 // Discord embed description limit (max is 4096)
)

// GenerativeModel defines the interface for AI model operations
type GenerativeModel interface {
	GenerateContent(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error)
	SetTemperature(temp float32)
	SetMaxOutputTokens(tokens int32)
}

// GenAIClient defines the interface for the Gemini client
type GenAIClient interface {
	GenerativeModel(name string) GenerativeModel
	Close() error
}

// genaiClientWrapper wraps the real genai.Client to implement GenAIClient
type genaiClientWrapper struct {
	client *genai.Client
}

func (w *genaiClientWrapper) GenerativeModel(name string) GenerativeModel {
	return w.client.GenerativeModel(name)
}

func (w *genaiClientWrapper) Close() error {
	return w.client.Close()
}

// ClientFactory creates GenAI clients
type ClientFactory func(ctx context.Context, apiKey string) (GenAIClient, error)

// defaultClientFactory creates real Gemini clients
func defaultClientFactory(ctx context.Context, apiKey string) (GenAIClient, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}
	return &genaiClientWrapper{client: client}, nil
}

// Client handles Gemini API interactions
type Client struct {
	apiKey        string
	modelName     string
	clientFactory ClientFactory
}

// NewClient creates a new Gemini API client
func NewClient(apiKey, modelName string) *Client {
	return &Client{
		apiKey:        apiKey,
		modelName:     modelName,
		clientFactory: defaultClientFactory,
	}
}

// NewClientWithFactory creates a new Gemini API client with a custom factory (for testing)
func NewClientWithFactory(apiKey, modelName string, factory ClientFactory) *Client {
	return &Client{
		apiKey:        apiKey,
		modelName:     modelName,
		clientFactory: factory,
	}
}

// GenerateDailySummary creates an AI-generated summary of GitHub events
func (c *Client) GenerateDailySummary(
	events []github.FormattedEvent,
) (string, error) {
	if len(events) == 0 {
		return "Hoje foi um dia de planejamento e reflexão no código.", nil
	}

	ctx := context.Background()

	// Initialize Gemini Client
	client, err := c.clientFactory(ctx, c.apiKey)
	if err != nil {
		return "", fmt.Errorf("failed to create Gemini client: %w", err)
	}
	defer func() {
		_ = client.Close() //nolint:errcheck // defer close is best effort
	}()

	// Configure the model
	model := client.GenerativeModel(c.modelName)
	model.SetTemperature(float32(temperature))
	model.SetMaxOutputTokens(int32(maxTokens))

	// Build the prompt
	prompt, err := c.buildPrompt(events)
	if err != nil {
		return "", fmt.Errorf("failed to build prompt: %w", err)
	}

	// Generate content
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate content; %w", err)
	}

	// Extract text from response
	summary := c.extractText(resp)
	if summary == "" {
		return "", fmt.Errorf("empty response from Gemini")
	}

	// Truncate if necessary
	return c.truncateSummary(summary), nil
}

// buildPrompt creates the prompt for Gemini based on events
func (c *Client) buildPrompt(events []github.FormattedEvent) (string, error) {
	eventsJSON, err := json.MarshalIndent(events, "", " ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal events: %w", err)
	}

	prompt := fmt.Sprintf(dailySummaryPromptTemplate, eventsJSON)

	return prompt, nil
}

// extractText extracts the text content from Gemini response
func (c *Client) extractText(resp *genai.GenerateContentResponse) string {
	var builder strings.Builder

	for _, candidate := range resp.Candidates {
		if candidate.Content != nil {
			for _, part := range candidate.Content.Parts {
				if text, ok := part.(genai.Text); ok {
					builder.WriteString(string(text))
				}
			}
		}
	}

	return strings.TrimSpace(builder.String())
}

// truncateSummary ensures the summary fits within Discord's limits
func (c *Client) truncateSummary(summary string) string {
	if len(summary) <= maxSummaryChars {
		return summary
	}

	// Find the latest period before the limit
	truncated := summary[:maxSummaryChars]
	lastPeriod := strings.LastIndex(truncated, ".")

	if lastPeriod > 0 {
		return summary[:lastPeriod+1]
	}

	// No period found, truncate at limit
	return truncated
}

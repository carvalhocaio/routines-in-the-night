package gemini

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/carvalhocaio/routines-in-the-night/internal/github"
	"github.com/google/generative-ai-go/genai"
)

// Mock implementations for testing

type mockGenerativeModel struct {
	response *genai.GenerateContentResponse
	err      error
}

func (m *mockGenerativeModel) GenerateContent(
	_ context.Context,
	_ ...genai.Part,
) (*genai.GenerateContentResponse, error) {
	return m.response, m.err
}

func (m *mockGenerativeModel) SetTemperature(_ float32) {}

func (m *mockGenerativeModel) SetMaxOutputTokens(_ int32) {}

type mockGenAIClient struct {
	model    *mockGenerativeModel
	closeErr error
}

func (m *mockGenAIClient) GenerativeModel(_ string) GenerativeModel {
	return m.model
}

func (m *mockGenAIClient) Close() error {
	return m.closeErr
}

func TestNewClient(t *testing.T) {
	apiKey := "test-api-key"
	modelName := "gemini-2.5-flash"
	client := NewClient(apiKey, modelName)

	if client.apiKey != apiKey {
		t.Errorf("Expected apiKey=%s, got: %s", apiKey, client.apiKey)
	}
	if client.modelName != modelName {
		t.Errorf("Expected modelName=%s, got: %s", modelName, client.modelName)
	}
}

func TestTruncateSummary(t *testing.T) {
	client := NewClient("test-key", "gemini-2.5-flash")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "short text",
			input:    "This is a short summary.",
			expected: "This is a short summary.",
		},
		{
			name:     "exactly at limit",
			input:    strings.Repeat("a", maxSummaryChars),
			expected: strings.Repeat("a", maxSummaryChars),
		},
		{
			name: "over limit with period",
			input: strings.Repeat("a", maxSummaryChars-10) +
				" sentence." +
				strings.Repeat("b", 50),
			expected: strings.Repeat("a", maxSummaryChars-10) + " sentence.",
		},
		{
			name:     "over limit no period",
			input:    strings.Repeat("a", maxSummaryChars+100),
			expected: strings.Repeat("a", maxSummaryChars),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.truncateSummary(tt.input)
			if result != tt.expected {
				t.Errorf(
					"truncateSummary() length=%d, expected length=%d",
					len(result),
					len(tt.expected),
				)
			}
			if len(result) > maxSummaryChars {
				t.Errorf(
					"Result exceeds max length: %d > %d",
					len(result),
					maxSummaryChars,
				)
			}
		})
	}
}

func TestGenerateDailySummary_EmptyEvents(t *testing.T) {
	client := NewClient("test-key", "gemini-2.5-flash")

	summary, err := client.GenerateDailySummary([]github.FormattedEvent{})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	expected := "Hoje foi um dia de planejamento e reflexão no código."
	if summary != expected {
		t.Errorf("Expected default message, got: %s", summary)
	}
}

func TestBuildPrompt(t *testing.T) {
	client := NewClient("test-key", "gemini-2.5-flash")

	events := []github.FormattedEvent{
		{
			Type:           "PushEvent",
			Repo:           "user/repo",
			CreatedAt:      time.Now(),
			IsPrivate:      false,
			Branch:         "main",
			Commits:        2,
			CommitMessages: []string{"Initial commit", "Add feature"},
		},
		{
			Type:      "CreateEvent",
			Repo:      "user/repo2",
			CreatedAt: time.Now(),
			IsPrivate: true,
			RefType:   "branch",
			Ref:       "feature/new",
		},
	}

	prompt, err := client.buildPrompt(events)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify prompt contains expected elements
	if !strings.Contains(prompt, "PushEvent") {
		t.Error("Prompt should contain PushEvent")
	}
	if !strings.Contains(prompt, "CreateEvent") {
		t.Error("Prompt should contain CreateEvent")
	}
	if !strings.Contains(prompt, "user/repo") {
		t.Error("Prompt should contain repository name")
	}
	if !strings.Contains(prompt, "Initial commit") {
		t.Error("Prompt should contain commit message")
	}
	if !strings.Contains(prompt, "REQUISITOS OBRIGATÓRIOS") {
		t.Error("Prompt should contain requirements section")
	}
}

func TestBuildPrompt_SingleEvent(t *testing.T) {
	client := NewClient("test-key", "gemini-2.5-flash")

	events := []github.FormattedEvent{
		{
			Type:      "IssuesEvent",
			Repo:      "user/repo",
			CreatedAt: time.Now(),
			IsPrivate: false,
			Action:    "opened",
		},
	}

	prompt, err := client.buildPrompt(events)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !strings.Contains(prompt, "IssuesEvent") {
		t.Error("Prompt should contain IssuesEvent")
	}
	if !strings.Contains(prompt, "opened") {
		t.Error("Prompt should contain action")
	}
}

func TestExtractText(t *testing.T) {
	client := NewClient("test-key", "gemini-2.5-flash")

	tests := []struct {
		name     string
		response *genai.GenerateContentResponse
		expected string
	}{
		{
			name: "single candidate with text",
			response: &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{
								genai.Text("This is the summary."),
							},
						},
					},
				},
			},
			expected: "This is the summary.",
		},
		{
			name: "multiple parts",
			response: &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{
								genai.Text("Part 1. "),
								genai.Text("Part 2."),
							},
						},
					},
				},
			},
			expected: "Part 1. Part 2.",
		},
		{
			name: "empty candidates",
			response: &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{},
			},
			expected: "",
		},
		{
			name: "nil content",
			response: &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{
					{
						Content: nil,
					},
				},
			},
			expected: "",
		},
		{
			name: "text with whitespace",
			response: &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []genai.Part{
								genai.Text("  Summary with spaces  "),
							},
						},
					},
				},
			},
			expected: "Summary with spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.extractText(tt.response)
			if result != tt.expected {
				t.Errorf("extractText() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestTruncateSummary_EdgeCases(t *testing.T) {
	client := NewClient("test-key", "gemini-2.5-flash")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "single character",
			input:    "a",
			expected: "a",
		},
		{
			name:     "period only over limit",
			input:    strings.Repeat("a", maxSummaryChars+10),
			expected: strings.Repeat("a", maxSummaryChars),
		},
		{
			name:     "period at position 0 (not truncated at period)",
			input:    "." + strings.Repeat("a", maxSummaryChars+10),
			expected: "." + strings.Repeat("a", maxSummaryChars-1), // period at 0 is not > 0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.truncateSummary(tt.input)
			if result != tt.expected {
				t.Errorf("truncateSummary(%q) = %q, expected %q",
					tt.input[:minInt(20, len(tt.input))],
					result[:minInt(20, len(result))],
					tt.expected[:minInt(20, len(tt.expected))])
			}
		})
	}
}

func TestBuildPrompt_AllEventTypes(t *testing.T) {
	client := NewClient("test-key", "gemini-2.5-flash")

	now := time.Now()
	events := []github.FormattedEvent{
		{
			Type:           "PushEvent",
			Repo:           "user/repo",
			CreatedAt:      now,
			Branch:         "main",
			Commits:        1,
			CommitMessages: []string{"commit"},
		},
		{
			Type:      "CreateEvent",
			Repo:      "user/repo",
			CreatedAt: now,
			RefType:   "branch",
			Ref:       "new-branch",
		},
		{
			Type:      "DeleteEvent",
			Repo:      "user/repo",
			CreatedAt: now,
			RefType:   "branch",
			Ref:       "old-branch",
		},
		{
			Type:      "IssuesEvent",
			Repo:      "user/repo",
			CreatedAt: now,
			Action:    "opened",
		},
		{
			Type:      "PullRequestEvent",
			Repo:      "user/repo",
			CreatedAt: now,
			Action:    "merged",
			PRTitle:   "Feature PR",
		},
	}

	prompt, err := client.buildPrompt(events)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify all event types are in the prompt
	eventTypes := []string{"PushEvent", "CreateEvent", "DeleteEvent", "IssuesEvent", "PullRequestEvent"}
	for _, eventType := range eventTypes {
		if !strings.Contains(prompt, eventType) {
			t.Errorf("Prompt should contain %s", eventType)
		}
	}

	// Verify the prompt template is included
	if !strings.Contains(prompt, "Atividades do dia:") {
		t.Error("Prompt should contain 'Atividades do dia:'")
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestGenerateDailySummary_Success(t *testing.T) {
	mockModel := &mockGenerativeModel{
		response: &genai.GenerateContentResponse{
			Candidates: []*genai.Candidate{
				{
					Content: &genai.Content{
						Parts: []genai.Part{
							genai.Text("This is a generated summary."),
						},
					},
				},
			},
		},
		err: nil,
	}

	mockClient := &mockGenAIClient{
		model:    mockModel,
		closeErr: nil,
	}

	factory := func(_ context.Context, _ string) (GenAIClient, error) {
		return mockClient, nil
	}

	client := NewClientWithFactory("test-key", "gemini-2.5-flash", factory)

	events := []github.FormattedEvent{
		{
			Type:      "PushEvent",
			Repo:      "user/repo",
			CreatedAt: time.Now(),
			Branch:    "main",
			Commits:   1,
		},
	}

	summary, err := client.GenerateDailySummary(events)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	expected := "This is a generated summary."
	if summary != expected {
		t.Errorf("Expected summary %q, got %q", expected, summary)
	}
}

func TestGenerateDailySummary_ClientFactoryError(t *testing.T) {
	factory := func(_ context.Context, _ string) (GenAIClient, error) {
		return nil, errors.New("failed to create client")
	}

	client := NewClientWithFactory("test-key", "gemini-2.5-flash", factory)

	events := []github.FormattedEvent{
		{
			Type:      "PushEvent",
			Repo:      "user/repo",
			CreatedAt: time.Now(),
		},
	}

	_, err := client.GenerateDailySummary(events)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !strings.Contains(err.Error(), "failed to create Gemini client") {
		t.Errorf("Expected error to contain 'failed to create Gemini client', got: %v", err)
	}
}

func TestGenerateDailySummary_GenerateContentError(t *testing.T) {
	mockModel := &mockGenerativeModel{
		response: nil,
		err:      errors.New("API error"),
	}

	mockClient := &mockGenAIClient{
		model:    mockModel,
		closeErr: nil,
	}

	factory := func(_ context.Context, _ string) (GenAIClient, error) {
		return mockClient, nil
	}

	client := NewClientWithFactory("test-key", "gemini-2.5-flash", factory)

	events := []github.FormattedEvent{
		{
			Type:      "PushEvent",
			Repo:      "user/repo",
			CreatedAt: time.Now(),
		},
	}

	_, err := client.GenerateDailySummary(events)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !strings.Contains(err.Error(), "failed to generate content") {
		t.Errorf("Expected error to contain 'failed to generate content', got: %v", err)
	}
}

func TestGenerateDailySummary_EmptyResponse(t *testing.T) {
	mockModel := &mockGenerativeModel{
		response: &genai.GenerateContentResponse{
			Candidates: []*genai.Candidate{},
		},
		err: nil,
	}

	mockClient := &mockGenAIClient{
		model:    mockModel,
		closeErr: nil,
	}

	factory := func(_ context.Context, _ string) (GenAIClient, error) {
		return mockClient, nil
	}

	client := NewClientWithFactory("test-key", "gemini-2.5-flash", factory)

	events := []github.FormattedEvent{
		{
			Type:      "PushEvent",
			Repo:      "user/repo",
			CreatedAt: time.Now(),
		},
	}

	_, err := client.GenerateDailySummary(events)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !strings.Contains(err.Error(), "empty response from Gemini") {
		t.Errorf("Expected error to contain 'empty response from Gemini', got: %v", err)
	}
}

func TestGenerateDailySummary_TruncatesLongResponse(t *testing.T) {
	// Create a response that exceeds maxSummaryChars
	longText := strings.Repeat("a", maxSummaryChars-10) + ". " + strings.Repeat("b", 100)

	mockModel := &mockGenerativeModel{
		response: &genai.GenerateContentResponse{
			Candidates: []*genai.Candidate{
				{
					Content: &genai.Content{
						Parts: []genai.Part{
							genai.Text(longText),
						},
					},
				},
			},
		},
		err: nil,
	}

	mockClient := &mockGenAIClient{
		model:    mockModel,
		closeErr: nil,
	}

	factory := func(_ context.Context, _ string) (GenAIClient, error) {
		return mockClient, nil
	}

	client := NewClientWithFactory("test-key", "gemini-2.5-flash", factory)

	events := []github.FormattedEvent{
		{
			Type:      "PushEvent",
			Repo:      "user/repo",
			CreatedAt: time.Now(),
		},
	}

	summary, err := client.GenerateDailySummary(events)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(summary) > maxSummaryChars {
		t.Errorf("Summary exceeds max length: %d > %d", len(summary), maxSummaryChars)
	}

	// Should be truncated at the period
	expectedLen := maxSummaryChars - 10 + 1 // includes the period
	if len(summary) != expectedLen {
		t.Errorf("Expected summary length %d, got %d", expectedLen, len(summary))
	}
}

func TestNewClientWithFactory(t *testing.T) {
	factoryCalled := false
	factory := func(_ context.Context, _ string) (GenAIClient, error) {
		factoryCalled = true
		return nil, errors.New("test")
	}

	client := NewClientWithFactory("key", "model", factory)

	if client.apiKey != "key" {
		t.Errorf("Expected apiKey 'key', got %s", client.apiKey)
	}
	if client.modelName != "model" {
		t.Errorf("Expected modelName 'model', got %s", client.modelName)
	}
	if client.clientFactory == nil {
		t.Error("Expected clientFactory to be set")
	}

	// Verify factory is the one we provided by calling it
	//nolint:errcheck,gosec // Intentionally ignoring error in test
	client.clientFactory(context.Background(), "")
	if !factoryCalled {
		t.Error("Expected custom factory to be called")
	}
}

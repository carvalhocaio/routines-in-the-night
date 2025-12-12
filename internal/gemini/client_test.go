package gemini

import (
	"strings"
	"testing"

	"github.com/carvalhocaio/routines-in-the-night/internal/github"
)

func TestNewClient(t *testing.T) {
	apiKey := "test-api-key"
	client := NewClient(apiKey)

	if client.apiKey != apiKey {
		t.Errorf("Expected apiKey=%s, got: %s", apiKey, client.apiKey)
	}
}

func TestTruncateSummary(t *testing.T) {
	client := NewClient("test-key")

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
	client := NewClient("test-key")

	summary, err := client.GenerateDailySummary([]github.FormattedEvent{})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	expected := "Hoje foi um dia de planejamento e reflexão no código."
	if summary != expected {
		t.Errorf("Expected default message, got: %s", summary)
	}
}

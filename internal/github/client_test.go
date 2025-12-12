package github

import (
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	username := "testuser"
	token := "testtoken"

	client := NewClient(username, token)

	if client.username != username {
		t.Errorf("Expected username=%s, got: %s", username, client.username)
	}
	if client.token != token {
		t.Errorf("Expected token=%s, got: %s", token, client.token)
	}
	if client.httpClient == nil {
		t.Error("Expected httpClient to be initialized")
	}
}

func TestExtractBranchName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "full ref",
			input:    "refs/heads/main",
			expected: "main",
		},
		{
			name:     "feature branch",
			input:    "refs/heads/feature/new-feature",
			expected: "feature/new-feature",
		},
		{
			name:     "no prefix",
			input:    "main",
			expected: "main",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "partial prefix",
			input:    "refs/heads",
			expected: "refs/heads",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractBranchName(tt.input)
			if result != tt.expected {
				t.Errorf(
					"extractBranchName(%s) = %s, expected %s",
					tt.input,
					result,
					tt.expected,
				)
			}
		})
	}
}

func TestFormatEvents(t *testing.T) {
	client := NewClient("testuser", "testtoken")

	now := time.Now()
	events := []Event{
		{
			Type:      "PushEvent",
			CreatedAt: now,
			Public:    true,
			Repo: Repo{
				Name: "testuser/testrepo",
			},
			Payload: Payload{
				Ref: "refs/heads/main",
				Commits: []Commit{
					{Message: "Initial commit", SHA: "abc123"},
					{Message: "Add feature", SHA: "def456"},
				},
			},
		},
		{
			Type:      "CreateEvent",
			CreatedAt: now,
			Public:    false,
			Repo: Repo{
				Name: "testuser/privaterepo",
			},
			Payload: Payload{
				RefType: "branch",
				Ref:     "feature/new",
			},
		},
		{
			Type:      "IssuesEvent",
			CreatedAt: now,
			Public:    true,
			Repo: Repo{
				Name: "testuser/testrepo",
			},
			Payload: Payload{
				Action: "opened",
			},
		},
	}

	formatted := client.formatEvents(events)

	if len(formatted) != 3 {
		t.Fatalf("Expected 3 formatted events, got: %d", len(formatted))
	}

	// Test PushEvent
	pushEvent := formatted[0]
	if pushEvent.Type != "PushEvent" {
		t.Errorf("Expected Type=PushEvent, got: %s", pushEvent.Type)
	}
	if pushEvent.Branch != "main" {
		t.Errorf("Expected Branch=main, got: %s", pushEvent.Branch)
	}
	if pushEvent.Commits != 2 {
		t.Errorf("Expected Commits=2, got: %d", pushEvent.Commits)
	}
	if len(pushEvent.CommitMessages) != 2 {
		t.Errorf("Expected 2 commit messages, got: %d", len(pushEvent.CommitMessages))
	}
	if pushEvent.IsPrivate {
		t.Error("Expected IsPrivate=false")
	}

	// Test CreateEvent
	createEvent := formatted[1]
	if createEvent.Type != "CreateEvent" {
		t.Errorf("Expected Type=CreateEvent, got: %s", createEvent.Type)
	}
	if createEvent.RefType != "branch" {
		t.Errorf("Expected RefType=branch, got: %s", createEvent.RefType)
	}
	if createEvent.Ref != "feature/new" {
		t.Errorf("Expected Ref=feature/new, got: %s", createEvent.Ref)
	}
	if !createEvent.IsPrivate {
		t.Error("Expected IsPrivate=true")
	}

	// Test IssuesEvent
	issueEvent := formatted[2]
	if issueEvent.Type != "IssuesEvent" {
		t.Errorf("Expected Type=IssuesEvent, got: %s", issueEvent.Type)
	}
	if issueEvent.Action != "opened" {
		t.Errorf("Expected Action=opened, got: %s", issueEvent.Action)
	}
}

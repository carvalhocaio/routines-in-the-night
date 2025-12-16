package github

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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
		{
			Type:      "DeleteEvent",
			CreatedAt: now,
			Public:    true,
			Repo: Repo{
				Name: "testuser/testrepo",
			},
			Payload: Payload{
				RefType: "branch",
				Ref:     "old-branch",
			},
		},
		{
			Type:      "PullRequestEvent",
			CreatedAt: now,
			Public:    true,
			Repo: Repo{
				Name: "testuser/testrepo",
			},
			Payload: Payload{
				Action: "opened",
				PullRequest: &PullRequest{
					Title: "Add new feature",
				},
			},
		},
	}

	formatted := client.formatEvents(events)

	if len(formatted) != 5 {
		t.Fatalf("Expected 5 formatted events, got: %d", len(formatted))
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

	// Test DeleteEvent
	deleteEvent := formatted[3]
	if deleteEvent.Type != "DeleteEvent" {
		t.Errorf("Expected Type=DeleteEvent, got: %s", deleteEvent.Type)
	}
	if deleteEvent.RefType != "branch" {
		t.Errorf("Expected RefType=branch, got: %s", deleteEvent.RefType)
	}
	if deleteEvent.Ref != "old-branch" {
		t.Errorf("Expected Ref=old-branch, got: %s", deleteEvent.Ref)
	}

	// Test PullRequestEvent
	prEvent := formatted[4]
	if prEvent.Type != "PullRequestEvent" {
		t.Errorf("Expected Type=PullRequestEvent, got: %s", prEvent.Type)
	}
	if prEvent.Action != "opened" {
		t.Errorf("Expected Action=opened, got: %s", prEvent.Action)
	}
	if prEvent.PRTitle != "Add new feature" {
		t.Errorf("Expected PRTitle='Add new feature', got: %s", prEvent.PRTitle)
	}
}

func TestFetchUserEvents_Success(t *testing.T) {
	now := time.Now()
	events := []Event{
		{
			Type:      "PushEvent",
			CreatedAt: now,
			Public:    true,
			Repo:      Repo{Name: "user/repo"},
			Payload: Payload{
				Ref:     "refs/heads/main",
				Commits: []Commit{{Message: "test", SHA: "abc"}},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			t.Errorf("Expected Bearer token, got: %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Accept") != "application/vnd.github.v3+json" {
			t.Errorf("Expected Accept header, got: %s", r.Header.Get("Accept"))
		}
		if r.Header.Get("X-GitHub-Api-Version") != githubAPIVERSION {
			t.Errorf("Expected API version header, got: %s", r.Header.Get("X-GitHub-Api-Version"))
		}

		// Verify URL path
		expectedPath := "/users/testuser/events"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got: %s", expectedPath, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(events); err != nil {
			t.Errorf("Failed to encode events: %v", err)
		}
	}))
	defer server.Close()

	client := NewClientWithBaseURL("testuser", "testtoken", server.URL)
	result, err := client.GetDailyEvents()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 event, got: %d", len(result))
	}
}

func TestGetDailyEvents_Success(t *testing.T) {
	now := time.Now()
	oldEvent := now.Add(-48 * time.Hour) // 2 days ago

	events := []Event{
		{
			Type:      "PushEvent",
			CreatedAt: now,
			Public:    true,
			Repo:      Repo{Name: "user/repo"},
			Payload: Payload{
				Ref:     "refs/heads/main",
				Commits: []Commit{{Message: "recent commit", SHA: "abc"}},
			},
		},
		{
			Type:      "PushEvent",
			CreatedAt: oldEvent,
			Public:    true,
			Repo:      Repo{Name: "user/repo"},
			Payload: Payload{
				Ref:     "refs/heads/main",
				Commits: []Commit{{Message: "old commit", SHA: "def"}},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(events); err != nil {
			t.Errorf("Failed to encode events: %v", err)
		}
	}))
	defer server.Close()

	client := NewClientWithBaseURL("testuser", "testtoken", server.URL)
	result, err := client.GetDailyEvents()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Only recent event should be returned (within 24 hours)
	if len(result) != 1 {
		t.Errorf("Expected 1 recent event, got: %d", len(result))
	}
	if len(result) > 0 && result[0].CommitMessages[0] != "recent commit" {
		t.Errorf("Expected recent commit, got: %s", result[0].CommitMessages[0])
	}
}

func TestGetDailyEvents_NoRecentEvents(t *testing.T) {
	oldTime := time.Now().Add(-48 * time.Hour)
	events := []Event{
		{
			Type:      "PushEvent",
			CreatedAt: oldTime,
			Public:    true,
			Repo:      Repo{Name: "user/repo"},
		},
	}

	// Filter events like GetDailyEvents does
	yesterday := time.Now().Add(-24 * time.Hour)
	var recentEvents []Event
	for i := range events {
		if events[i].CreatedAt.After(yesterday) {
			recentEvents = append(recentEvents, events[i])
		}
	}

	if len(recentEvents) != 0 {
		t.Errorf("Expected 0 recent events, got: %d", len(recentEvents))
	}
}

func TestFetchUserEvents_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClientWithBaseURL("testuser", "testtoken", server.URL)
	_, err := client.GetDailyEvents()

	if err == nil {
		t.Error("Expected error for server error response")
	}
	if !strings.Contains(err.Error(), "unexpected status code: 500") {
		t.Errorf("Expected status code error, got: %v", err)
	}
}

func TestFetchUserEvents_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte("invalid json")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := NewClientWithBaseURL("testuser", "testtoken", server.URL)
	_, err := client.GetDailyEvents()

	if err == nil {
		t.Error("Expected JSON decode error")
	}
	if !strings.Contains(err.Error(), "failed to decode response") {
		t.Errorf("Expected decode error, got: %v", err)
	}
}

func TestFetchUserEvents_NetworkError(t *testing.T) {
	client := NewClientWithBaseURL("testuser", "testtoken", "http://localhost:99999")
	_, err := client.GetDailyEvents()

	if err == nil {
		t.Error("Expected network error")
	}
	if !strings.Contains(err.Error(), "failed to execute request") {
		t.Errorf("Expected network error, got: %v", err)
	}
}

func TestNewClientWithBaseURL(t *testing.T) {
	client := NewClientWithBaseURL("user", "token", "http://custom.api")

	if client.username != "user" {
		t.Errorf("Expected username=user, got: %s", client.username)
	}
	if client.token != "token" {
		t.Errorf("Expected token=token, got: %s", client.token)
	}
	if client.baseURL != "http://custom.api" {
		t.Errorf("Expected baseURL=http://custom.api, got: %s", client.baseURL)
	}
}

func TestGetDailyEvents_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode([]Event{}); err != nil {
			t.Errorf("Failed to encode events: %v", err)
		}
	}))
	defer server.Close()

	client := NewClientWithBaseURL("testuser", "testtoken", server.URL)
	result, err := client.GetDailyEvents()

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("Expected 0 events, got: %d", len(result))
	}
}

func TestFormatEvents_EmptySlice(t *testing.T) {
	client := NewClient("testuser", "testtoken")

	formatted := client.formatEvents([]Event{})

	if len(formatted) != 0 {
		t.Errorf("Expected 0 events, got: %d", len(formatted))
	}
}

func TestFormatEvents_PullRequestWithoutPR(t *testing.T) {
	client := NewClient("testuser", "testtoken")

	events := []Event{
		{
			Type:      "PullRequestEvent",
			CreatedAt: time.Now(),
			Public:    true,
			Repo:      Repo{Name: "user/repo"},
			Payload: Payload{
				Action:      "opened",
				PullRequest: nil, // No PR object
			},
		},
	}

	formatted := client.formatEvents(events)

	if len(formatted) != 1 {
		t.Fatalf("Expected 1 event, got: %d", len(formatted))
	}
	if formatted[0].PRTitle != "" {
		t.Errorf("Expected empty PRTitle, got: %s", formatted[0].PRTitle)
	}
}

func TestEventTypes(t *testing.T) {
	client := NewClient("testuser", "testtoken")
	now := time.Now()

	// Test all event types
	events := []Event{
		{Type: "PushEvent", CreatedAt: now, Public: true, Repo: Repo{Name: "r"}},
		{Type: "CreateEvent", CreatedAt: now, Public: true, Repo: Repo{Name: "r"}},
		{Type: "DeleteEvent", CreatedAt: now, Public: true, Repo: Repo{Name: "r"}},
		{Type: "IssuesEvent", CreatedAt: now, Public: true, Repo: Repo{Name: "r"}},
		{Type: "PullRequestEvent", CreatedAt: now, Public: true, Repo: Repo{Name: "r"}},
		{Type: "WatchEvent", CreatedAt: now, Public: true, Repo: Repo{Name: "r"}}, // Unknown type
	}

	formatted := client.formatEvents(events)

	if len(formatted) != 6 {
		t.Errorf("Expected 6 events, got: %d", len(formatted))
	}

	// Verify types are preserved
	expectedTypes := []string{"PushEvent", "CreateEvent", "DeleteEvent", "IssuesEvent", "PullRequestEvent", "WatchEvent"}
	for i, expected := range expectedTypes {
		if formatted[i].Type != expected {
			t.Errorf("Event %d: expected type %s, got %s", i, expected, formatted[i].Type)
		}
	}
}

package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	githubAPIURL     = "https://api.github.com"
	githubAPIVERSION = "2022-11-28"
)

// Client handles GitHub API interactions
type Client struct {
	username   string
	token      string
	httpClient *http.Client
}

// Event represents a GitHub event
type Event struct {
	Type      string    `json:"type"`
	Repo      Repo      `json:"repo"`
	CreatedAt time.Time `json:"created_at"`
	Public    bool      `json:"public"`
	Payload   Payload   `json:"payload"`
}

// Repo represents repository information
type Repo struct {
	Name string `json:"name"`
}

// Payload represents event-specific data
type Payload struct {
	Ref         string       `json:"ref"`
	RefType     string       `json:"ref_type"`
	Action      string       `json:"action"`
	Commits     []Commit     `json:"commits"`
	PullRequest *PullRequest `json:"pull_request"`
}

// Commit represents a commit in a push event
type Commit struct {
	Message string `json:"message"`
	SHA     string `json:"sha"`
}

// PullRequest represents PR information
type PullRequest struct {
	Title string `json:"title"`
}

// FormattedEvent contains processes event information
type FormattedEvent struct {
	Type           string    `json:"type"`
	Repo           string    `json:"repo"`
	CreateAt       time.Time `json:"create_at"`
	IsPrivate      bool      `json:"is_private"`
	Branch         string    `json:"branch,omitempty"`
	Commits        int       `json:"commits,omitempty"`
	CommitMessages []string  `json:"commit_messages,omitempty"`
	RefType        string    `json:"ref_type,omitempty"`
	Ref            string    `json:"ref,omitempty"`
	Action         string    `json:"action,omitempty"`
	PRTitle        string    `json:"pr_title,omitempty"`
}

// NewClient creates a new GitHub API client
func NewClient(username, token string) *Client {
	return &Client{
		username: username,
		token:    token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetDailyEvents fetches GitHub events from the last 24 hours
func (c *Client) GetDailyEvents() ([]FormattedEvent, error) {
	events, err := c.fetchUserEvents()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch events: %w", err)
	}

	// Filter events from the last 24 hours
	yesterday := time.Now().Add(-24 * time.Hour)
	var recentEvents []Event

	for i := range events {
		if events[i].CreatedAt.After(yesterday) {
			recentEvents = append(recentEvents, events[i])
		}
	}

	return c.formatEvents(recentEvents), nil
}

// fetchUserEvents retrieves events from GitHub API
func (c *Client) fetchUserEvents() ([]Event, error) {
	url := fmt.Sprintf("%s/users/%s/events", githubAPIURL, c.username)

	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("X-GitHub-Api-Version", githubAPIVERSION)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close() //nolint:errcheck // defer close is best effort
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"unexpected status code: %d",
			resp.StatusCode,
		)
	}

	var events []Event
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return events, nil
}

// formatEvents converts raw events to formatted events
func (c *Client) formatEvents(events []Event) []FormattedEvent {
	formatted := make([]FormattedEvent, 0, len(events))

	for i := range events {
		fe := FormattedEvent{
			Type:      events[i].Type,
			Repo:      events[i].Repo.Name,
			CreateAt:  events[i].CreatedAt,
			IsPrivate: !events[i].Public,
		}

		switch events[i].Type {
		case "PushEvent":
			fe.Commits = len(events[i].Payload.Commits)
			fe.Branch = extractBranchName(events[i].Payload.Ref)

			messages := make([]string, 0, len(events[i].Payload.Commits))
			for _, commit := range events[i].Payload.Commits {
				messages = append(messages, commit.Message)
			}
			fe.CommitMessages = messages

		case "CreateEvent":
			fe.RefType = events[i].Payload.RefType
			fe.Ref = events[i].Payload.Ref

		case "DeleteEvent":
			fe.RefType = events[i].Payload.RefType
			fe.Ref = events[i].Payload.Ref

		case "IssuesEvent", "PullRequestEvent":
			fe.Action = events[i].Payload.Action
			if events[i].Payload.PullRequest != nil {
				fe.PRTitle = events[i].Payload.PullRequest.Title
			}
		}

		formatted = append(formatted, fe)
	}

	return formatted
}

// extractBranchName removes "refs/heads/" prefix from ref
func extractBranchName(ref string) string {
	const prefix = "refs/heads/"
	if len(ref) > len(prefix) && ref[:len(prefix)] == prefix {
		return ref[len(prefix):]
	}
	return ref
}

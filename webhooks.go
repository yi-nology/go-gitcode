package gitcode

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Webhook struct {
	ID                  int64     `json:"id"`
	URL                 string    `json:"url"`
	Events              []string  `json:"events"`
	PushEvents          bool      `json:"push_events"`
	TagPushEvents       bool      `json:"tag_push_events"`
	IssuesEvents        bool      `json:"issues_events"`
	MergeRequestsEvents bool      `json:"merge_requests_events"`
	NoteEvents          bool      `json:"note_events"`
	Active              bool      `json:"active"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type CreateWebhookOptions struct {
	URL    string   `json:"url"`
	Secret string   `json:"secret,omitempty"`
	Events []string `json:"events,omitempty"`
	Active *bool    `json:"active,omitempty"`
}

type UpdateWebhookOptions struct {
	URL    string   `json:"url,omitempty"`
	Secret string   `json:"secret,omitempty"`
	Events []string `json:"events,omitempty"`
	Active *bool    `json:"active,omitempty"`
}

func (c *Client) ListWebhooks(ctx context.Context, owner, repo string) ([]*Webhook, error) {
	var hooks []*Webhook
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/hooks", owner, repo), nil, &hooks)
	if err != nil {
		return nil, err
	}
	return hooks, nil
}

func (c *Client) GetWebhook(ctx context.Context, owner, repo string, hookID int64) (*Webhook, error) {
	var hook Webhook
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/hooks/%d", owner, repo, hookID), nil, &hook)
	if err != nil {
		return nil, err
	}
	return &hook, nil
}

func (c *Client) CreateWebhook(ctx context.Context, owner, repo string, opts CreateWebhookOptions) (*Webhook, error) {
	var hook Webhook
	err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/hooks", owner, repo), opts, &hook)
	if err != nil {
		return nil, err
	}
	return &hook, nil
}

func (c *Client) UpdateWebhook(ctx context.Context, owner, repo string, hookID int64, opts UpdateWebhookOptions) (*Webhook, error) {
	var hook Webhook
	err := c.doRequest(ctx, http.MethodPatch, fmt.Sprintf("/repos/%s/%s/hooks/%d", owner, repo, hookID), opts, &hook)
	if err != nil {
		return nil, err
	}
	return &hook, nil
}

func (c *Client) DeleteWebhook(ctx context.Context, owner, repo string, hookID int64) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/repos/%s/%s/hooks/%d", owner, repo, hookID), nil, nil)
}

func (c *Client) TestWebhook(ctx context.Context, owner, repo string, hookID int64) error {
	return c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/hooks/%d/tests", owner, repo, hookID), nil, nil)
}

type WebhookEvent struct {
	Ref        string      `json:"ref"`
	Before     string      `json:"before"`
	After      string      `json:"after"`
	Repository *Repository `json:"repository"`
	Commits    []*Commit   `json:"commits"`
	Sender     *User       `json:"sender"`
}

type PullRequestWebhookEvent struct {
	Action      string       `json:"action"`
	Number      int          `json:"number"`
	PullRequest *PullRequest `json:"pull_request"`
	Repository  *Repository  `json:"repository"`
	Sender      *User        `json:"sender"`
}

type IssueWebhookEvent struct {
	Action     string      `json:"action"`
	Issue      *Issue      `json:"issue"`
	Repository *Repository `json:"repository"`
	Sender     *User       `json:"sender"`
}

type NoteWebhookEvent struct {
	NoteType   string      `json:"noteable_type"`
	NoteID     int64       `json:"id"`
	Body       string      `json:"body"`
	Author     *User       `json:"author"`
	Repository *Repository `json:"repository"`
	Sender     *User       `json:"sender"`
}

type TagPushEvent struct {
	Ref        string      `json:"ref"`
	Before     string      `json:"before"`
	After      string      `json:"after"`
	Repository *Repository `json:"repository"`
	Sender     *User       `json:"sender"`
}

type PushEvent struct {
	Ref        string      `json:"ref"`
	Before     string      `json:"before"`
	After      string      `json:"after"`
	Repository *Repository `json:"repository"`
	Commits    []*Commit   `json:"commits"`
	Sender     *User       `json:"sender"`
}

func (c *Client) ParsePushEvent(payload []byte) (*PushEvent, error) {
	var event PushEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, err
	}
	return &event, nil
}

func (c *Client) ParsePullRequestEvent(payload []byte) (*PullRequestWebhookEvent, error) {
	var event PullRequestWebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, err
	}
	return &event, nil
}

func (c *Client) ParseIssueEvent(payload []byte) (*IssueWebhookEvent, error) {
	var event IssueWebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, err
	}
	return &event, nil
}

func (c *Client) ParseNoteEvent(payload []byte) (*NoteWebhookEvent, error) {
	var event NoteWebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, err
	}
	return &event, nil
}

func (c *Client) ParseTagPushEvent(payload []byte) (*TagPushEvent, error) {
	var event TagPushEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, err
	}
	return &event, nil
}

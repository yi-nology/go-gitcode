package gitcode

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Webhook mirrors GitCode's documented repository hook object. The wire
// expresses event subscriptions as per-family flags (push_events,
// issues_events, ...) — there is no "events" array. The flags use FlexBool
// because the documented response schema types them as integers (0/1).
type Webhook struct {
	ID                  int64     `json:"id"`
	URL                 string    `json:"url"`
	Password            string    `json:"password,omitempty"`
	EncryptionType      *int      `json:"encryption_type,omitempty"`
	ProjectID           int64     `json:"project_id,omitempty"`
	PushEvents          FlexBool  `json:"push_events"`
	TagPushEvents       FlexBool  `json:"tag_push_events"`
	IssuesEvents        FlexBool  `json:"issues_events"`
	MergeRequestsEvents FlexBool  `json:"merge_requests_events"`
	NoteEvents          FlexBool  `json:"note_events"`
	Active              FlexBool  `json:"active"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// CreateWebhookOptions is the body of POST /repos/{owner}/{repo}/hooks.
// Event subscriptions are per-family flags: nil leaves the family at the
// server default, a pointer (use Bool) states it explicitly.
type CreateWebhookOptions struct {
	URL      string `json:"url"`
	Password string `json:"password,omitempty"`
	// EncryptionType: 0 = password (default), 1 = signature key.
	EncryptionType      *int  `json:"encryption_type,omitempty"`
	PushEvents          *bool `json:"push_events,omitempty"`
	TagPushEvents       *bool `json:"tag_push_events,omitempty"`
	IssuesEvents        *bool `json:"issues_events,omitempty"`
	NoteEvents          *bool `json:"note_events,omitempty"`
	MergeRequestsEvents *bool `json:"merge_requests_events,omitempty"`
	Active              *bool `json:"active,omitempty"`
}

// UpdateWebhookOptions is the body of PATCH
// /repos/{owner}/{repo}/hooks/{id}. Nil fields stay off the wire
// ("unchanged"); a pointer set through Bool(false) turns a family off.
type UpdateWebhookOptions struct {
	URL                 string `json:"url,omitempty"`
	Password            string `json:"password,omitempty"`
	EncryptionType      *int   `json:"encryption_type,omitempty"`
	PushEvents          *bool  `json:"push_events,omitempty"`
	TagPushEvents       *bool  `json:"tag_push_events,omitempty"`
	IssuesEvents        *bool  `json:"issues_events,omitempty"`
	NoteEvents          *bool  `json:"note_events,omitempty"`
	MergeRequestsEvents *bool  `json:"merge_requests_events,omitempty"`
	Active              *bool  `json:"active,omitempty"`
}

// Bool returns a pointer to v, for the webhook option flag fields.
func Bool(v bool) *bool { return &v }

// EventFlags maps the platform-neutral event names used by this SDK's
// callers onto GitCode's per-family flag fields.
var EventFlags = map[string]string{
	"push":          "push_events",
	"tag_push":      "tag_push_events",
	"issues":        "issues_events",
	"note":          "note_events",
	"comment":       "note_events",
	"pull_request":  "merge_requests_events",
	"merge_request": "merge_requests_events",
}

// ListWebhooks lists a repository's hooks (server-default first page). Use
// ListWebhooksWithOptions to drive pagination.
func (c *Client) ListWebhooks(ctx context.Context, owner, repo string) ([]*Webhook, error) {
	return c.ListWebhooksWithOptions(ctx, owner, repo, ListOptions{})
}

// ListWebhooksWithOptions lists a repository's hooks with explicit
// pagination (GET /repos/{owner}/{repo}/hooks).
func (c *Client) ListWebhooksWithOptions(ctx context.Context, owner, repo string, opts ListOptions) ([]*Webhook, error) {
	var hooks []*Webhook
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/hooks?%s", owner, repo, opts.toQuery()), nil, &hooks)
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
	NoteType   string      `json:"noteable_type"` //nolint:misspell // API field name
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

package gitcode

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// OrgWebhook represents an organization webhook.
type OrgWebhook struct {
	ID                  int64          `json:"id"`
	URL                 string         `json:"url"`
	Events              []string       `json:"events"`
	Active              bool           `json:"active"`
	Config              *WebhookConfig `json:"config,omitempty"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	PushEvents          bool           `json:"push_events"`
	TagPushEvents       bool           `json:"tag_push_events"`
	IssuesEvents        bool           `json:"issues_events"`
	NoteEvents          bool           `json:"note_events"`
	MergeRequestsEvents bool           `json:"merge_requests_events"`
}

// WebhookConfig represents the configuration of a webhook.
type WebhookConfig struct {
	URL         string `json:"url"`
	ContentType string `json:"content_type,omitempty"`
	Secret      string `json:"secret,omitempty"`
	InsecureSSL bool   `json:"insecure_ssl,omitempty"`
}

// CreateOrgWebhookOptions specifies options for creating an organization webhook.
type CreateOrgWebhookOptions struct {
	URL    string         `json:"url"`
	Secret string         `json:"secret,omitempty"`
	Events []string       `json:"events,omitempty"`
	Active *bool          `json:"active,omitempty"`
	Config *WebhookConfig `json:"config,omitempty"`
}

// UpdateOrgWebhookOptions specifies options for updating an organization webhook.
type UpdateOrgWebhookOptions struct {
	URL    string         `json:"url,omitempty"`
	Secret string         `json:"secret,omitempty"`
	Events []string       `json:"events,omitempty"`
	Active *bool          `json:"active,omitempty"`
	Config *WebhookConfig `json:"config,omitempty"`
}

// ListOrgWebhooks lists all webhooks of an organization.
//
// GET /orgs/{org}/hooks
func (c *Client) ListOrgWebhooks(ctx context.Context, org string, opts ListOptions) ([]*OrgWebhook, error) {
	var hooks []*OrgWebhook
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/orgs/%s/hooks?%s", org, opts.toQuery()), nil, &hooks)
	if err != nil {
		return nil, err
	}
	return hooks, nil
}

// GetOrgWebhook gets a single organization webhook.
//
// GET /orgs/{org}/hooks/{id}
func (c *Client) GetOrgWebhook(ctx context.Context, org string, hookID int64) (*OrgWebhook, error) {
	var hook OrgWebhook
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/orgs/%s/hooks/%d", org, hookID), nil, &hook)
	if err != nil {
		return nil, err
	}
	return &hook, nil
}

// CreateOrgWebhook creates a new organization webhook.
//
// POST /orgs/{org}/hooks
func (c *Client) CreateOrgWebhook(ctx context.Context, org string, opts CreateOrgWebhookOptions) (*OrgWebhook, error) {
	var hook OrgWebhook
	err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/orgs/%s/hooks", org), opts, &hook)
	if err != nil {
		return nil, err
	}
	return &hook, nil
}

// UpdateOrgWebhook updates an organization webhook.
//
// PATCH /orgs/{org}/hooks/{id}
func (c *Client) UpdateOrgWebhook(ctx context.Context, org string, hookID int64, opts UpdateOrgWebhookOptions) (*OrgWebhook, error) {
	var hook OrgWebhook
	err := c.doRequest(ctx, http.MethodPatch, fmt.Sprintf("/orgs/%s/hooks/%d", org, hookID), opts, &hook)
	if err != nil {
		return nil, err
	}
	return &hook, nil
}

// DeleteOrgWebhook deletes an organization webhook.
//
// DELETE /orgs/{org}/hooks/{id}
func (c *Client) DeleteOrgWebhook(ctx context.Context, org string, hookID int64) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/orgs/%s/hooks/%d", org, hookID), nil, nil)
}

package gitcode

import (
	"context"
	"net/http"
	"time"
)

// IssueTimelineEvent represents an event in an issue's timeline.
type IssueTimelineEvent struct {
	ID        int64      `json:"id"`
	Action    string     `json:"action"`
	Body      string     `json:"body,omitempty"`
	User      *User      `json:"user"`
	Author    *User      `json:"author"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	CommitID  string     `json:"commit_id,omitempty"`
	EventType string     `json:"event,omitempty"`
	Label     *Label     `json:"label,omitempty"`
	Milestone *Milestone `json:"milestone,omitempty"`
	Rename    *struct {
		From string `json:"from"`
		To   string `json:"to"`
	} `json:"rename,omitempty"`
}

// ListIssueTimelineEvents lists all timeline events for an issue.
//
// GET /repos/{owner}/{repo}/issues/{index}/timeline
func (c *Client) ListIssueTimelineEvents(ctx context.Context, owner, repo string, number int, opts ListOptions) ([]*IssueTimelineEvent, error) {
	var events []*IssueTimelineEvent
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/issues/%d/timeline?%s", owner, repo, number, opts.toQuery()), nil, &events)
	if err != nil {
		return nil, err
	}
	return events, nil
}

// IssueSubscriber represents a user subscribed to an issue.
type IssueSubscriber struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

// ListIssueSubscribers lists all subscribers of an issue.
//
// GET /repos/{owner}/{repo}/issues/{index}/subscribers
func (c *Client) ListIssueSubscribers(ctx context.Context, owner, repo string, number int, opts ListOptions) ([]*IssueSubscriber, error) {
	var subscribers []*IssueSubscriber
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/issues/%d/subscribers?%s", owner, repo, number, opts.toQuery()), nil, &subscribers)
	if err != nil {
		return nil, err
	}
	return subscribers, nil
}

// SubscribeToIssue subscribes the authenticated user to an issue.
//
// PUT /repos/{owner}/{repo}/issues/{index}/subscribers/{username}
func (c *Client) SubscribeToIssue(ctx context.Context, owner, repo string, number int, username string) error {
	return c.doRequest(ctx, http.MethodPut, pathf("/repos/%s/%s/issues/%d/subscribers/%s", owner, repo, number, username), nil, nil)
}

// UnsubscribeFromIssue unsubscribes the authenticated user from an issue.
//
// DELETE /repos/{owner}/{repo}/issues/{index}/subscribers/{username}
func (c *Client) UnsubscribeFromIssue(ctx context.Context, owner, repo string, number int, username string) error {
	return c.doRequest(ctx, http.MethodDelete, pathf("/repos/%s/%s/issues/%d/subscribers/%s", owner, repo, number, username), nil, nil)
}

// IssueDependency represents a dependency between issues.
type IssueDependency struct {
	ID    int64  `json:"id"`
	Issue *Issue `json:"issue"`
}

// ListIssueDependencies lists all dependencies of an issue.
//
// GET /repos/{owner}/{repo}/issues/{index}/dependencies
func (c *Client) ListIssueDependencies(ctx context.Context, owner, repo string, number int, opts ListOptions) ([]*IssueDependency, error) {
	var deps []*IssueDependency
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/issues/%d/dependencies?%s", owner, repo, number, opts.toQuery()), nil, &deps)
	if err != nil {
		return nil, err
	}
	return deps, nil
}

// CreateIssueDependency creates a dependency between two issues.
//
// POST /repos/{owner}/{repo}/issues/{index}/dependencies
func (c *Client) CreateIssueDependency(ctx context.Context, owner, repo string, number int, dependsOnNumber int) (*IssueDependency, error) {
	var dep IssueDependency
	err := c.doRequest(ctx, http.MethodPost, pathf("/repos/%s/%s/issues/%d/dependencies", owner, repo, number), map[string]int{"depends_on": dependsOnNumber}, &dep)
	if err != nil {
		return nil, err
	}
	return &dep, nil
}

// DeleteIssueDependency removes a dependency between two issues.
//
// DELETE /repos/{owner}/{repo}/issues/{index}/dependencies/{dependency}
func (c *Client) DeleteIssueDependency(ctx context.Context, owner, repo string, number int, dependencyNumber int) error {
	return c.doRequest(ctx, http.MethodDelete, pathf("/repos/%s/%s/issues/%d/dependencies/%d", owner, repo, number, dependencyNumber), nil, nil)
}

// ListIssueBlockingIssues lists all issues that this issue blocks.
//
// GET /repos/{owner}/{repo}/issues/{index}/blocks
func (c *Client) ListIssueBlockingIssues(ctx context.Context, owner, repo string, number int, opts ListOptions) ([]*Issue, error) {
	var issues []*Issue
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/issues/%d/blocks?%s", owner, repo, number, opts.toQuery()), nil, &issues)
	if err != nil {
		return nil, err
	}
	return issues, nil
}

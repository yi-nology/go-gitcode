package gitcode

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type IssueState string

const (
	IssueStateOpen   IssueState = "open"
	IssueStateClosed IssueState = "closed"
)

type Issue struct {
	ID        int64      `json:"id"`
	Number    FlexInt    `json:"number"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	State     IssueState `json:"state"`
	User      *User      `json:"user"`
	Author    *User      `json:"author"`
	Assignees []*User    `json:"assignees"`
	Labels    []*Label   `json:"labels"`
	Milestone *Milestone `json:"milestone"`
	HTMLURL   string     `json:"html_url"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	ClosedAt  *time.Time `json:"closed_at"`
}

type Label struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// Milestone mirrors GitCode's documented milestone object: the wire
// identifies milestones by `number` (older deployments may also emit `id`),
// and `due_on` is a date-only string such as "2026-01-31" — deliberately a
// string, since Go's time.Time rejects that shape.
type Milestone struct {
	ID           int64     `json:"id"`
	Number       int64     `json:"number"`
	RepositoryID int64     `json:"repository_id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	State        string    `json:"state"`
	DueOn        string    `json:"due_on,omitempty"`
	URL          string    `json:"url,omitempty"`
	OpenIssues   int       `json:"open_issues"`
	ClosedIssues int       `json:"closed_issues"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type IssueComment struct {
	ID        int64     `json:"id"`
	Body      string    `json:"body"`
	User      *User     `json:"user"`
	Author    *User     `json:"author"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListIssuesOptions struct {
	ListOptions
	State     IssueState `json:"state,omitempty"`
	Assignee  string     `json:"assignee,omitempty"`
	Creator   string     `json:"creator,omitempty"`
	Milestone string     `json:"milestone,omitempty"`
	Labels    string     `json:"labels,omitempty"`
	Sort      string     `json:"sort,omitempty"`
	Direction string     `json:"direction,omitempty"`
	Since     string     `json:"since,omitempty"`
}

type CreateIssueOptions struct {
	Title     string   `json:"title"`
	Body      string   `json:"body,omitempty"`
	Assignee  string   `json:"assignee,omitempty"`
	Assignees []string `json:"assignees,omitempty"`
	Milestone int64    `json:"milestone,omitempty"`
	Labels    []string `json:"labels,omitempty"`
}

func (o CreateIssueOptions) MarshalJSON() ([]byte, error) {
	type Alias CreateIssueOptions
	a := Alias(o)
	raw, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	if labels, ok := m["labels"].([]interface{}); ok {
		parts := make([]string, len(labels))
		for i, l := range labels {
			parts[i] = fmt.Sprint(l)
		}
		m["labels"] = strings.Join(parts, ",")
	}
	return json.Marshal(m)
}

type UpdateIssueOptions struct {
	Title      string     `json:"title,omitempty"`
	Body       string     `json:"body,omitempty"`
	State      IssueState `json:"state,omitempty"`
	StateEvent string     `json:"state_event,omitempty"`
	Assignee   string     `json:"assignee,omitempty"`
	Assignees  []string   `json:"assignees,omitempty"`
	Milestone  int64      `json:"milestone,omitempty"`
	Labels     []string   `json:"labels,omitempty"`
}

func (c *Client) ListIssues(ctx context.Context, owner, repo string, opts ListIssuesOptions) ([]*Issue, error) {
	var issues []*Issue
	q := url.Values{}
	q.Set("page", strconv.Itoa(opts.withDefaults().Page))
	q.Set("per_page", strconv.Itoa(opts.withDefaults().PerPage))
	if opts.State != "" {
		q.Set("state", string(opts.State))
	}
	if opts.Assignee != "" {
		q.Set("assignee", opts.Assignee)
	}
	if opts.Creator != "" {
		q.Set("creator", opts.Creator)
	}
	if opts.Milestone != "" {
		q.Set("milestone", opts.Milestone)
	}
	if opts.Labels != "" {
		q.Set("labels", opts.Labels)
	}
	if opts.Sort != "" {
		q.Set("sort", opts.Sort)
	}
	if opts.Direction != "" {
		q.Set("direction", opts.Direction)
	}
	if opts.Since != "" {
		q.Set("since", opts.Since)
	}
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/issues?%s", owner, repo, q.Encode()), nil, &issues)
	if err != nil {
		return nil, err
	}
	return issues, nil
}

func (c *Client) GetIssue(ctx context.Context, owner, repo string, number int) (*Issue, error) {
	var issue Issue
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/issues/%d", owner, repo, number), nil, &issue)
	if err != nil {
		return nil, err
	}
	return &issue, nil
}

func (c *Client) CreateIssue(ctx context.Context, owner, repo string, opts CreateIssueOptions) (*Issue, error) {
	var issue Issue
	err := c.doRequest(ctx, http.MethodPost, pathf("/repos/%s/%s/issues", owner, repo), opts, &issue)
	if err != nil {
		return nil, err
	}
	return &issue, nil
}

func (c *Client) UpdateIssue(ctx context.Context, owner, repo string, number int, opts UpdateIssueOptions) (*Issue, error) {
	var issue Issue
	err := c.doRequest(ctx, http.MethodPatch, pathf("/repos/%s/%s/issues/%d", owner, repo, number), opts, &issue)
	if err != nil {
		return nil, err
	}
	return &issue, nil
}

func (c *Client) CloseIssue(ctx context.Context, owner, repo string, number int) (*Issue, error) {
	issue, err := c.GetIssue(ctx, owner, repo, number)
	if err != nil {
		return nil, err
	}
	return c.UpdateIssue(ctx, owner, repo, number, UpdateIssueOptions{
		Title:      issue.Title,
		StateEvent: "close",
	})
}

func (c *Client) ReopenIssue(ctx context.Context, owner, repo string, number int) (*Issue, error) {
	issue, err := c.GetIssue(ctx, owner, repo, number)
	if err != nil {
		return nil, err
	}
	return c.UpdateIssue(ctx, owner, repo, number, UpdateIssueOptions{
		Title:      issue.Title,
		StateEvent: "reopen",
	})
}

// ListIssueCommentsOptions carries the documented query parameters of the
// issue-comment list endpoint (page/per_page plus order/since).
type ListIssueCommentsOptions struct {
	ListOptions
	Order string `json:"order,omitempty"` // "asc" (default) or "desc"
	Since string `json:"since,omitempty"` // RFC3339 timestamp; '+' must be URL-encoded
}

// ListIssueComments lists the comments of an issue (server-default first
// page). Use ListIssueCommentsWithOptions to drive pagination.
func (c *Client) ListIssueComments(ctx context.Context, owner, repo string, number int) ([]*IssueComment, error) {
	return c.ListIssueCommentsWithOptions(ctx, owner, repo, number, ListIssueCommentsOptions{})
}

// ListIssueCommentsWithOptions lists the comments of an issue with explicit
// pagination (GET /repos/{owner}/{repo}/issues/{number}/comments).
func (c *Client) ListIssueCommentsWithOptions(ctx context.Context, owner, repo string, number int, opts ListIssueCommentsOptions) ([]*IssueComment, error) {
	var comments []*IssueComment
	q := url.Values{}
	q.Set("page", strconv.Itoa(opts.withDefaults().Page))
	q.Set("per_page", strconv.Itoa(opts.withDefaults().PerPage))
	if opts.Order != "" {
		q.Set("order", opts.Order)
	}
	if opts.Since != "" {
		q.Set("since", opts.Since)
	}
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/issues/%d/comments?%s", owner, repo, number, q.Encode()), nil, &comments)
	if err != nil {
		return nil, err
	}
	return comments, nil
}

func (c *Client) CreateIssueComment(ctx context.Context, owner, repo string, number int, body string) (*IssueComment, error) {
	var comment IssueComment
	err := c.doRequest(ctx, http.MethodPost, pathf("/repos/%s/%s/issues/%d/comments", owner, repo, number), map[string]string{"body": body}, &comment)
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (c *Client) UpdateIssueComment(ctx context.Context, owner, repo string, commentID int64, body string) (*IssueComment, error) {
	var comment IssueComment
	err := c.doRequest(ctx, http.MethodPatch, pathf("/repos/%s/%s/issues/comments/%d", owner, repo, commentID), map[string]string{"body": body}, &comment)
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (c *Client) DeleteIssueComment(ctx context.Context, owner, repo string, commentID int64) error {
	return c.doRequest(ctx, http.MethodDelete, pathf("/repos/%s/%s/issues/comments/%d", owner, repo, commentID), nil, nil)
}

// ListIssueLabels lists a repository's labels (server-default first page).
// Use ListIssueLabelsWithOptions to drive pagination.
func (c *Client) ListIssueLabels(ctx context.Context, owner, repo string) ([]*Label, error) {
	return c.ListIssueLabelsWithOptions(ctx, owner, repo, ListOptions{})
}

// ListIssueLabelsWithOptions lists a repository's labels with explicit
// pagination (GET /repos/{owner}/{repo}/labels).
func (c *Client) ListIssueLabelsWithOptions(ctx context.Context, owner, repo string, opts ListOptions) ([]*Label, error) {
	var labels []*Label
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/labels?%s", owner, repo, opts.toQuery()), nil, &labels)
	if err != nil {
		return nil, err
	}
	return labels, nil
}

func (c *Client) CreateIssueLabel(ctx context.Context, owner, repo string, name, color string) (*Label, error) {
	var label Label
	params := url.Values{"name": {name}, "color": {"#" + strings.TrimPrefix(color, "#")}}
	err := c.doFormRequest(ctx, http.MethodPost, pathf("/repos/%s/%s/labels", owner, repo), params, &label)
	if err != nil {
		return nil, err
	}
	return &label, nil
}

func (c *Client) DeleteIssueLabel(ctx context.Context, owner, repo string, name string) error {
	return c.doRequest(ctx, http.MethodDelete, pathf("/repos/%s/%s/labels/%s", owner, repo, name), nil, nil)
}

func (c *Client) AddIssueLabels(ctx context.Context, owner, repo string, number int, labels []string) error {
	return c.doRawBodyRequest(ctx, http.MethodPost, pathf("/repos/%s/%s/issues/%d/labels", owner, repo, number), labels, nil)
}

func (c *Client) RemoveIssueLabel(ctx context.Context, owner, repo string, number int, name string) error {
	return c.doRequest(ctx, http.MethodDelete, pathf("/repos/%s/%s/issues/%d/labels/%s", owner, repo, number, name), nil, nil)
}

func (c *Client) ListMilestones(ctx context.Context, owner, repo string) ([]*Milestone, error) {
	return c.ListMilestonesWithOptions(ctx, owner, repo, ListMilestonesOptions{})
}

func (c *Client) CreateMilestone(ctx context.Context, owner, repo string, title, description string) (*Milestone, error) {
	return c.CreateMilestoneWithOptions(ctx, owner, repo, CreateMilestoneOptions{Title: title, Description: description})
}

func (c *Client) DeleteMilestone(ctx context.Context, owner, repo string, number int) error {
	return c.doRequest(ctx, http.MethodDelete, pathf("/repos/%s/%s/milestones/%d", owner, repo, number), nil, nil)
}

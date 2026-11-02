package gitcode

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type ListUserIssuesOptions struct {
	ListOptions
	Filter    string `json:"filter,omitempty"`
	State     string `json:"state,omitempty"`
	Labels    string `json:"labels,omitempty"`
	Sort      string `json:"sort,omitempty"`
	Direction string `json:"direction,omitempty"`
	Since     string `json:"since,omitempty"`
}

// userIssuesQuery renders the shared filter set with escaped values.
func userIssuesQuery(opts ListUserIssuesOptions) string {
	q := url.Values{}
	q.Set("page", strconv.Itoa(opts.withDefaults().Page))
	q.Set("per_page", strconv.Itoa(opts.withDefaults().PerPage))
	setIf := func(key, val string) {
		if val != "" {
			q.Set(key, val)
		}
	}
	setIf("filter", opts.Filter)
	setIf("state", opts.State)
	setIf("labels", opts.Labels)
	setIf("sort", opts.Sort)
	setIf("direction", opts.Direction)
	setIf("since", opts.Since)
	return q.Encode()
}

func (c *Client) ListUserIssues(ctx context.Context, opts ListUserIssuesOptions) ([]*Issue, error) {
	var issues []*Issue
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/user/issues?%s", userIssuesQuery(opts)), nil, &issues)
	if err != nil {
		return nil, err
	}
	return issues, nil
}

func (c *Client) ListOrgIssues(ctx context.Context, org string, opts ListUserIssuesOptions) ([]*Issue, error) {
	var issues []*Issue
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/orgs/%s/issues?%s", org, userIssuesQuery(opts)), nil, &issues)
	if err != nil {
		return nil, err
	}
	return issues, nil
}

func (c *Client) ListRepoAllIssueComments(ctx context.Context, owner, repo string, opts ListOptions) ([]*IssueComment, error) {
	var comments []*IssueComment
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/issues/comments?%s", owner, repo, opts.toQuery()), nil, &comments)
	if err != nil {
		return nil, err
	}
	return comments, nil
}

func (c *Client) GetIssueComment(ctx context.Context, owner, repo string, commentID int64) (*IssueComment, error) {
	var comment IssueComment
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/issues/comments/%d", owner, repo, commentID), nil, &comment)
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

type IssueOperateLog struct {
	ID        int64  `json:"id"`
	Action    string `json:"action"`
	CreatedAt string `json:"created_at"`
	User      *User  `json:"user"`
}

func (c *Client) GetIssueOperateLogs(ctx context.Context, owner, repo string, number int) ([]*IssueOperateLog, error) {
	var logs []*IssueOperateLog
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/issues/%d/operate_logs", owner, repo, number), nil, &logs)
	if err != nil {
		return nil, err
	}
	return logs, nil
}

func (c *Client) GetIssueLinkedPRs(ctx context.Context, owner, repo string, number int) ([]*PullRequest, error) {
	var prs []*PullRequest
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/issues/%d/pull_requests", owner, repo, number), nil, &prs)
	if err != nil {
		return nil, err
	}
	return prs, nil
}

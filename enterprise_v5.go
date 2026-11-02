package gitcode

// v5-legacy enterprise endpoints. The Enterprise family moved to the v8
// wire (enterprise.go), but four capabilities remain documented ONLY at v5
// (docs.gitcode.com v5 slug provenance in each method comment): the single
// issue fetch and the issue comment/label/status lists. They ride the v5
// base and the v5 issue shape; the rest of the family (members, labels
// CRUD, milestones, custom fields, issue filtering) lives on enterprise.go.

import (
	"context"
	"fmt"
	"net/http"
)

// EnterpriseIssueStatus represents an enterprise-level issue status.
//
// v5 wire: GET /enterprises/{enterprise}/issue-statuses (docs slug
// get-api-v-5-enterprises-enterprise-issue-statuses); no v8 page exists.
type EnterpriseIssueStatus struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Color     string `json:"color,omitempty"`
	Sort      int    `json:"sort,omitempty"`
	IsDefault bool   `json:"is_default,omitempty"`
}

// ListEnterpriseIssueStatuses lists all enterprise issue statuses.
//
// GET /enterprises/{enterprise}/issue-statuses (v5).
func (c *Client) ListEnterpriseIssueStatuses(ctx context.Context, enterprise string) ([]*EnterpriseIssueStatus, error) {
	var statuses []*EnterpriseIssueStatus
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/enterprises/%s/issue-statuses", enterprise), nil, &statuses)
	if err != nil {
		return nil, err
	}
	return statuses, nil
}

// GetEnterpriseIssue fetches a single enterprise issue. The v8 family has
// no single-issue GET; this rides the documented v5 shape.
//
// GET /enterprises/{enterprise}/issues/{number} (v5, docs slug
// get-api-v-5-enterprises-enterprise-issues-number).
func (c *Client) GetEnterpriseIssue(ctx context.Context, enterprise string, number int) (*Issue, error) {
	var issue Issue
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/enterprises/%s/issues/%d", enterprise, number), nil, &issue)
	if err != nil {
		return nil, err
	}
	return &issue, nil
}

// ListEnterpriseIssueComments lists an enterprise issue's comments.
//
// GET /enterprises/{enterprise}/issues/{number}/comments (v5, docs slug
// get-api-v-5-enterprises-enterprise-issues-number-comments).
func (c *Client) ListEnterpriseIssueComments(ctx context.Context, enterprise string, number int, opts ListOptions) ([]*IssueComment, error) {
	var comments []*IssueComment
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/enterprises/%s/issues/%d/comments?%s", enterprise, number, opts.toQuery()), nil, &comments)
	if err != nil {
		return nil, err
	}
	return comments, nil
}

// ListEnterpriseIssueLabels lists the labels attached to an enterprise
// issue.
//
// GET /enterprises/{enterprise}/issues/{issue_id}/labels (v5, docs slug
// get-api-v-5-enterprises-enterprise-issues-issue-id-labels).
func (c *Client) ListEnterpriseIssueLabels(ctx context.Context, enterprise string, issueID int64) ([]*Label, error) {
	var labels []*Label
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/enterprises/%s/issues/%d/labels", enterprise, issueID), nil, &labels)
	if err != nil {
		return nil, err
	}
	return labels, nil
}

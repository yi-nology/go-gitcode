package gitcode

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// --- Issue Related Branches ---

// IssueRelatedBranch represents a branch related to an issue.
type IssueRelatedBranch struct {
	BranchName string `json:"branch_name"`
	RepoName   string `json:"repo_name,omitempty"`
}

// ListIssueRelatedBranches lists all branches related to an issue.
//
// GET /repos/{owner}/{repo}/issues/{number}/related-branches
func (c *Client) ListIssueRelatedBranches(ctx context.Context, owner, repo string, number int) ([]*IssueRelatedBranch, error) {
	var branches []*IssueRelatedBranch
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/issues/%d/related-branches", owner, repo, number), nil, &branches)
	if err != nil {
		return nil, err
	}
	return branches, nil
}

// SetIssueRelatedBranches sets the branches related to an issue.
//
// PUT /repos/{owner}/{repo}/issues/{number}/related-branches
func (c *Client) SetIssueRelatedBranches(ctx context.Context, owner, repo string, number int, branches []string) error {
	return c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/repos/%s/%s/issues/%d/related-branches", owner, repo, number), map[string]interface{}{"branches": branches}, nil)
}

// --- Issue Kanban Values ---

// KanbanValue represents a kanban field value for an issue.
type KanbanValue struct {
	FieldID   int64  `json:"field_id"`
	FieldName string `json:"field_name"`
	ValueID   int64  `json:"value_id"`
	ValueName string `json:"value_name"`
}

// UpdateIssueKanbanValues updates the kanban field values of an issue.
//
// PUT /repos/{owner}/{repo}/issues/{number}/kanban-values
func (c *Client) UpdateIssueKanbanValues(ctx context.Context, owner, repo string, number int, values []KanbanValue) error {
	return c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/repos/%s/%s/issues/%d/kanban-values", owner, repo, number), values, nil)
}

// --- Issue Modify History ---

// ModifyHistoryEntry represents a modification history entry.
type ModifyHistoryEntry struct {
	ID        int64     `json:"id"`
	Action    string    `json:"action"`
	Field     string    `json:"field,omitempty"`
	OldValue  string    `json:"old_value,omitempty"`
	NewValue  string    `json:"new_value,omitempty"`
	User      *User     `json:"user"`
	Author    *User     `json:"author"`
	CreatedAt time.Time `json:"created_at"`
}

// ListIssueModifyHistory lists the modification history of an issue.
//
// GET /repos/{owner}/{repo}/issues/{number}/modify-history
func (c *Client) ListIssueModifyHistory(ctx context.Context, owner, repo string, number int, opts ListOptions) ([]*ModifyHistoryEntry, error) {
	var history []*ModifyHistoryEntry
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/issues/%d/modify-history?%s", owner, repo, number, opts.toQuery()), nil, &history)
	if err != nil {
		return nil, err
	}
	return history, nil
}

// ListIssueCommentModifyHistory lists the modification history of an issue comment.
//
// GET /repos/{owner}/{repo}/issues/comment/{comment_id}/modify-history
func (c *Client) ListIssueCommentModifyHistory(ctx context.Context, owner, repo string, commentID int64, opts ListOptions) ([]*ModifyHistoryEntry, error) {
	var history []*ModifyHistoryEntry
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/issues/comment/%d/modify-history?%s", owner, repo, commentID, opts.toQuery()), nil, &history)
	if err != nil {
		return nil, err
	}
	return history, nil
}

// --- Issue User Reactions (separate from emoji reactions) ---

// IssueUserReaction represents a user's reaction to an issue.
type IssueUserReaction struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	User      *User     `json:"user"`
	CreatedAt time.Time `json:"created_at"`
}

// ListIssueUserReactions lists user reactions for an issue.
//
// GET /repos/{owner}/{repo}/issues/{number}/user-reactions
func (c *Client) ListIssueUserReactions(ctx context.Context, owner, repo string, number int, opts ListOptions) ([]*IssueUserReaction, error) {
	var reactions []*IssueUserReaction
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/issues/%d/user-reactions?%s", owner, repo, number, opts.toQuery()), nil, &reactions)
	if err != nil {
		return nil, err
	}
	return reactions, nil
}

// ListIssueCommentUserReactions lists user reactions for an issue comment.
//
// GET /repos/{owner}/{repo}/issues/comment/{comment_id}/user-reactions
func (c *Client) ListIssueCommentUserReactions(ctx context.Context, owner, repo string, commentID int64, opts ListOptions) ([]*IssueUserReaction, error) {
	var reactions []*IssueUserReaction
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/issues/comment/%d/user-reactions?%s", owner, repo, commentID, opts.toQuery()), nil, &reactions)
	if err != nil {
		return nil, err
	}
	return reactions, nil
}

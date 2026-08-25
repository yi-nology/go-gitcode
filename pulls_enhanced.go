package gitcode

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// --- PR Linked Issues ---

// LinkPullRequestIssue links an issue to a pull request.
//
// POST /repos/{owner}/{repo}/pulls/{number}/linked-issues
func (c *Client) LinkPullRequestIssue(ctx context.Context, owner, repo string, number int, issueNumber int) error {
	return c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/pulls/%d/linked-issues", owner, repo, number), map[string]int{"issue_number": issueNumber}, nil)
}

// UnlinkPullRequestIssue removes an issue link from a pull request.
//
// DELETE /repos/{owner}/{repo}/pulls/{number}/issues
func (c *Client) UnlinkPullRequestIssue(ctx context.Context, owner, repo string, number int, issueNumber int) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/repos/%s/%s/pulls/%d/issues?issue_number=%d", owner, repo, number, issueNumber), nil, nil)
}

// --- PR Tester Management ---

// UnassignPullRequestTesters removes testers from a pull request.
//
// DELETE /repos/{owner}/{repo}/pulls/{number}/testers
func (c *Client) UnassignPullRequestTesters(ctx context.Context, owner, repo string, number int, testers string) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/repos/%s/%s/pulls/%d/testers?testers=%s", owner, repo, number, testers), nil, nil)
}

// ListPullRequestAvailableTesters lists users available as testers for a pull request.
//
// GET /repos/{owner}/{repo}/pulls/{number}/option-approval-testers
func (c *Client) ListPullRequestAvailableTesters(ctx context.Context, owner, repo string, number int, opts ListOptions) ([]*User, error) {
	var users []*User
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/pulls/%d/option-approval-testers?%s", owner, repo, number, opts.toQuery()), nil, &users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// --- PR Reviewer Management (approval-reviewers) ---

// AssignPullRequestApprovalReviewers assigns users as approval reviewers for a pull request.
//
// POST /repos/{owner}/{repo}/pulls/{number}/approval-reviewers
func (c *Client) AssignPullRequestApprovalReviewers(ctx context.Context, owner, repo string, number int, reviewers string) error {
	return c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/pulls/%d/approval-reviewers", owner, repo, number), map[string]string{"reviewers": reviewers}, nil)
}

// UnassignPullRequestApprovalReviewers removes approval reviewers from a pull request.
//
// DELETE /repos/{owner}/{repo}/pulls/{number}/approval-reviewers
func (c *Client) UnassignPullRequestApprovalReviewers(ctx context.Context, owner, repo string, number int, reviewers string) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/repos/%s/%s/pulls/%d/approval-reviewers?reviewers=%s", owner, repo, number, reviewers), nil, nil)
}

// ListPullRequestAvailableReviewers lists users available as approval reviewers for a pull request.
//
// GET /repos/{owner}/{repo}/pulls/{number}/option-approval-reviewers
func (c *Client) ListPullRequestAvailableReviewers(ctx context.Context, owner, repo string, number int, opts ListOptions) ([]*User, error) {
	var users []*User
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/pulls/%d/option-approval-reviewers?%s", owner, repo, number, opts.toQuery()), nil, &users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// --- PR Discussion Comments ---

// PRDiscussionComment represents a reply to a PR review comment.
type PRDiscussionComment struct {
	ID           int64     `json:"id"`
	Body         string    `json:"body"`
	User         *User     `json:"user"`
	Author       *User     `json:"author"`
	DiscussionID string    `json:"discussion_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ReplyPullRequestComment replies to a pull request review comment.
//
// POST /repos/{owner}/{repo}/pulls/{number}/discussions/{discussions_id}/comments
func (c *Client) ReplyPullRequestComment(ctx context.Context, owner, repo string, number int, discussionID string, body string) (*PRDiscussionComment, error) {
	var comment PRDiscussionComment
	err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/pulls/%d/discussions/%s/comments", owner, repo, number, discussionID), map[string]string{"body": body}, &comment)
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

// ResolvePullRequestDiscussion resolves or unresolves a review discussion.
//
// PUT /repos/{owner}/{repo}/pulls/{number}/comments/discussions/{id}
func (c *Client) ResolvePullRequestDiscussion(ctx context.Context, owner, repo string, number int, discussionID string, resolved bool) error {
	return c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/repos/%s/%s/pulls/%d/comments/discussions/%s", owner, repo, number, discussionID), map[string]bool{"resolved": resolved}, nil)
}

// --- PR Modify History ---

// ListPullRequestModifyHistory lists the modification history of a pull request.
//
// GET /repos/{owner}/{repo}/pulls/{number}/modify-history
func (c *Client) ListPullRequestModifyHistory(ctx context.Context, owner, repo string, number int, opts ListOptions) ([]*ModifyHistoryEntry, error) {
	var history []*ModifyHistoryEntry
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/pulls/%d/modify-history?%s", owner, repo, number, opts.toQuery()), nil, &history)
	if err != nil {
		return nil, err
	}
	return history, nil
}

// ListPullRequestCommentModifyHistory lists the modification history of a PR comment.
//
// GET /repos/{owner}/{repo}/pulls/comment/{comment_id}/modify-history
func (c *Client) ListPullRequestCommentModifyHistory(ctx context.Context, owner, repo string, commentID int64, opts ListOptions) ([]*ModifyHistoryEntry, error) {
	var history []*ModifyHistoryEntry
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/pulls/comment/%d/modify-history?%s", owner, repo, commentID, opts.toQuery()), nil, &history)
	if err != nil {
		return nil, err
	}
	return history, nil
}

// --- PR User Reactions ---

// ListPullRequestUserReactions lists user reactions for a pull request.
//
// GET /repos/{owner}/{repo}/pulls/{number}/user-reactions
func (c *Client) ListPullRequestUserReactions(ctx context.Context, owner, repo string, number int, opts ListOptions) ([]*IssueUserReaction, error) {
	var reactions []*IssueUserReaction
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/pulls/%d/user-reactions?%s", owner, repo, number, opts.toQuery()), nil, &reactions)
	if err != nil {
		return nil, err
	}
	return reactions, nil
}

// ListPullRequestCommentUserReactions lists user reactions for a PR comment.
//
// GET /repos/{owner}/{repo}/pulls/comment/{comment_id}/user-reactions
func (c *Client) ListPullRequestCommentUserReactions(ctx context.Context, owner, repo string, commentID int64, opts ListOptions) ([]*IssueUserReaction, error) {
	var reactions []*IssueUserReaction
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/pulls/comment/%d/user-reactions?%s", owner, repo, commentID, opts.toQuery()), nil, &reactions)
	if err != nil {
		return nil, err
	}
	return reactions, nil
}

// --- PR Refresh Position ---

// RefreshPullRequestCommentPosition refreshes the position/expired status of PR comments.
//
// POST /repos/{owner}/{repo}/pulls/{number}/refresh-position
func (c *Client) RefreshPullRequestCommentPosition(ctx context.Context, owner, repo string, number int) error {
	return c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/pulls/%d/refresh-position", owner, repo, number), nil, nil)
}

// --- PR Files JSON ---

// PullRequestFileChange represents a file change in a pull request (JSON format).
type PullRequestFileChange struct {
	Filename         string `json:"filename"`
	PreviousFilename string `json:"previous_filename,omitempty"`
	Status           string `json:"status"`
	Additions        int    `json:"additions"`
	Deletions        int    `json:"deletions"`
	Changes          int    `json:"changes"`
	BlobURL          string `json:"blob_url,omitempty"`
	RawURL           string `json:"raw_url,omitempty"`
	ContentsURL      string `json:"contents_url,omitempty"`
	Patch            string `json:"patch,omitempty"`
}

// ListPullRequestFilesJSON lists files changed in a pull request in JSON format.
//
// GET /repos/{owner}/{repo}/pulls/{number}/files.json
func (c *Client) ListPullRequestFilesJSON(ctx context.Context, owner, repo string, number int, opts ListOptions) ([]*PullRequestFileChange, error) {
	var files []*PullRequestFileChange
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/pulls/%d/files.json?%s", owner, repo, number, opts.toQuery()), nil, &files)
	if err != nil {
		return nil, err
	}
	return files, nil
}

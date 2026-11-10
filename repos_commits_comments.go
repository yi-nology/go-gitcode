package gitcode

import (
	"context"
	"net/http"
	"time"
)

// CommitComment represents a comment on a commit.
type CommitComment struct {
	ID        int64     `json:"id"`
	Body      string    `json:"body"`
	Path      string    `json:"path,omitempty"`
	Position  int       `json:"position,omitempty"`
	Line      int       `json:"line,omitempty"`
	CommitID  string    `json:"commit_id"`
	User      *User     `json:"user"`
	Author    *User     `json:"author"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	HTMLURL   string    `json:"html_url,omitempty"`
}

// CreateCommitCommentOptions specifies options for creating a commit comment.
type CreateCommitCommentOptions struct {
	Body     string `json:"body"`
	Path     string `json:"path,omitempty"`
	Position int    `json:"position,omitempty"`
	Line     int    `json:"line,omitempty"`
}

// UpdateCommitCommentOptions specifies options for updating a commit comment.
type UpdateCommitCommentOptions struct {
	Body string `json:"body"`
}

// ListCommitComments lists all comments for a commit.
//
// GET /repos/{owner}/{repo}/commits/{sha}/comments
func (c *Client) ListCommitComments(ctx context.Context, owner, repo, sha string, opts ListOptions) ([]*CommitComment, error) {
	var comments []*CommitComment
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/commits/%s/comments?%s", owner, repo, sha, opts.toQuery()), nil, &comments)
	if err != nil {
		return nil, err
	}
	return comments, nil
}

// ListRepoCommitComments lists all commit comments for a repository.
//
// GET /repos/{owner}/{repo}/comments
func (c *Client) ListRepoCommitComments(ctx context.Context, owner, repo string, opts ListOptions) ([]*CommitComment, error) {
	var comments []*CommitComment
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/comments?%s", owner, repo, opts.toQuery()), nil, &comments)
	if err != nil {
		return nil, err
	}
	return comments, nil
}

// GetCommitComment gets a single commit comment.
//
// GET /repos/{owner}/{repo}/comments/{id}
func (c *Client) GetCommitComment(ctx context.Context, owner, repo string, commentID int64) (*CommitComment, error) {
	var comment CommitComment
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/comments/%d", owner, repo, commentID), nil, &comment)
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

// CreateCommitComment creates a new commit comment.
//
// POST /repos/{owner}/{repo}/commits/{sha}/comments
func (c *Client) CreateCommitComment(ctx context.Context, owner, repo, sha string, opts CreateCommitCommentOptions) (*CommitComment, error) {
	var comment CommitComment
	err := c.doRequest(ctx, http.MethodPost, pathf("/repos/%s/%s/commits/%s/comments", owner, repo, sha), opts, &comment)
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

// UpdateCommitComment updates a commit comment.
//
// PATCH /repos/{owner}/{repo}/comments/{id}
func (c *Client) UpdateCommitComment(ctx context.Context, owner, repo string, commentID int64, opts UpdateCommitCommentOptions) (*CommitComment, error) {
	var comment CommitComment
	err := c.doRequest(ctx, http.MethodPatch, pathf("/repos/%s/%s/comments/%d", owner, repo, commentID), opts, &comment)
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

// DeleteCommitComment deletes a commit comment.
//
// DELETE /repos/{owner}/{repo}/comments/{id}
func (c *Client) DeleteCommitComment(ctx context.Context, owner, repo string, commentID int64) error {
	return c.doRequest(ctx, http.MethodDelete, pathf("/repos/%s/%s/comments/%d", owner, repo, commentID), nil, nil)
}

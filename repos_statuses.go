package gitcode

import (
	"context"
	"net/http"
	"time"
)

// CommitStatus represents a commit status.
type CommitStatus struct {
	ID          int64     `json:"id"`
	SHA         string    `json:"sha"`
	State       string    `json:"state"` // pending, success, error, failure
	TargetURL   string    `json:"target_url"`
	Description string    `json:"description"`
	Context     string    `json:"context"`
	Creator     *User     `json:"creator"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateCommitStatusOptions specifies options for creating a commit status.
type CreateCommitStatusOptions struct {
	State       string `json:"state"` // pending, success, error, failure
	TargetURL   string `json:"target_url,omitempty"`
	Description string `json:"description,omitempty"`
	Context     string `json:"context,omitempty"`
}

// CombinedStatus represents the combined status for a commit.
type CombinedStatus struct {
	SHA        string          `json:"sha"`
	TotalCount int             `json:"total_count"`
	Statuses   []*CommitStatus `json:"statuses"`
	Repository *Repository     `json:"repository"`
	CommitURL  string          `json:"commit_url"`
	URL        string          `json:"url"`
}

// ListCommitStatuses lists all statuses for a commit.
//
// GET /repos/{owner}/{repo}/statuses/{sha}
func (c *Client) ListCommitStatuses(ctx context.Context, owner, repo, sha string, opts ListOptions) ([]*CommitStatus, error) {
	var statuses []*CommitStatus
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/statuses/%s?%s", owner, repo, sha, opts.toQuery()), nil, &statuses)
	if err != nil {
		return nil, err
	}
	return statuses, nil
}

// CreateCommitStatus creates a commit status.
//
// POST /repos/{owner}/{repo}/statuses/{sha}
func (c *Client) CreateCommitStatus(ctx context.Context, owner, repo, sha string, opts CreateCommitStatusOptions) (*CommitStatus, error) {
	var status CommitStatus
	err := c.doRequest(ctx, http.MethodPost, pathf("/repos/%s/%s/statuses/%s", owner, repo, sha), opts, &status)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

// GetCombinedStatus gets the combined status for a commit.
//
// GET /repos/{owner}/{repo/commits/{sha}/status
func (c *Client) GetCombinedStatus(ctx context.Context, owner, repo, sha string) (*CombinedStatus, error) {
	var status CombinedStatus
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/commits/%s/status", owner, repo, sha), nil, &status)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

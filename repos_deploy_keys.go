package gitcode

import (
	"context"
	"net/http"
	"time"
)

// DeployKey represents a repository deploy key.
type DeployKey struct {
	ID        int64     `json:"id"`
	Key       string    `json:"key"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	ReadOnly  bool      `json:"read_only"`
	URL       string    `json:"url,omitempty"`
}

// CreateDeployKeyOptions specifies options for creating a deploy key.
type CreateDeployKeyOptions struct {
	Title    string `json:"title"`
	Key      string `json:"key"`
	ReadOnly *bool  `json:"read_only,omitempty"`
}

// ListDeployKeys lists all deploy keys of a repository.
//
// GET /repos/{owner}/{repo}/keys
func (c *Client) ListDeployKeys(ctx context.Context, owner, repo string, opts ListOptions) ([]*DeployKey, error) {
	var keys []*DeployKey
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/keys?%s", owner, repo, opts.toQuery()), nil, &keys)
	if err != nil {
		return nil, err
	}
	return keys, nil
}

// GetDeployKey gets a single deploy key.
//
// GET /repos/{owner}/{repo}/keys/{id}
func (c *Client) GetDeployKey(ctx context.Context, owner, repo string, keyID int64) (*DeployKey, error) {
	var key DeployKey
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/keys/%d", owner, repo, keyID), nil, &key)
	if err != nil {
		return nil, err
	}
	return &key, nil
}

// CreateDeployKey creates a new deploy key for a repository.
//
// POST /repos/{owner}/{repo}/keys
func (c *Client) CreateDeployKey(ctx context.Context, owner, repo string, opts CreateDeployKeyOptions) (*DeployKey, error) {
	var key DeployKey
	err := c.doRequest(ctx, http.MethodPost, pathf("/repos/%s/%s/keys", owner, repo), opts, &key)
	if err != nil {
		return nil, err
	}
	return &key, nil
}

// DeleteDeployKey deletes a deploy key.
//
// DELETE /repos/{owner}/{repo}/keys/{id}
func (c *Client) DeleteDeployKey(ctx context.Context, owner, repo string, keyID int64) error {
	return c.doRequest(ctx, http.MethodDelete, pathf("/repos/%s/%s/keys/%d", owner, repo, keyID), nil, nil)
}

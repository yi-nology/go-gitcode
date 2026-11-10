package gitcode

import (
	"context"
	"net/http"
	"time"
)

// ProtectedTag represents a protected tag rule.
type ProtectedTag struct {
	ID          int64     `json:"id"`
	NamePattern string    `json:"name_pattern"`
	Owners      []*User   `json:"owners,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateProtectedTagOptions specifies options for creating a protected tag.
type CreateProtectedTagOptions struct {
	NamePattern string   `json:"name_pattern"`
	Owners      []string `json:"owners,omitempty"`
}

// UpdateProtectedTagOptions specifies options for updating a protected tag.
type UpdateProtectedTagOptions struct {
	NamePattern string   `json:"name_pattern,omitempty"`
	Owners      []string `json:"owners,omitempty"`
}

// ListProtectedTags lists all protected tag rules for a repository.
//
// GET /repos/{owner}/{repo}/protected-tags
func (c *Client) ListProtectedTags(ctx context.Context, owner, repo string, opts ListOptions) ([]*ProtectedTag, error) {
	var tags []*ProtectedTag
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/protected-tags?%s", owner, repo, opts.toQuery()), nil, &tags)
	if err != nil {
		return nil, err
	}
	return tags, nil
}

// GetProtectedTag gets a single protected tag rule.
//
// GET /repos/{owner}/{repo}/protected-tags/{tag}
func (c *Client) GetProtectedTag(ctx context.Context, owner, repo, tag string) (*ProtectedTag, error) {
	var pt ProtectedTag
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/protected-tags/%s", owner, repo, tag), nil, &pt)
	if err != nil {
		return nil, err
	}
	return &pt, nil
}

// CreateProtectedTag creates a new protected tag rule.
//
// POST /repos/{owner}/{repo}/protected-tags
func (c *Client) CreateProtectedTag(ctx context.Context, owner, repo string, opts CreateProtectedTagOptions) (*ProtectedTag, error) {
	var pt ProtectedTag
	err := c.doRequest(ctx, http.MethodPost, pathf("/repos/%s/%s/protected-tags", owner, repo), opts, &pt)
	if err != nil {
		return nil, err
	}
	return &pt, nil
}

// UpdateProtectedTag updates a protected tag rule.
//
// PUT /repos/{owner}/{repo}/protected-tags/{tag}
func (c *Client) UpdateProtectedTag(ctx context.Context, owner, repo, tag string, opts UpdateProtectedTagOptions) (*ProtectedTag, error) {
	var pt ProtectedTag
	err := c.doRequest(ctx, http.MethodPut, pathf("/repos/%s/%s/protected-tags/%s", owner, repo, tag), opts, &pt)
	if err != nil {
		return nil, err
	}
	return &pt, nil
}

// DeleteProtectedTag deletes a protected tag rule.
//
// DELETE /repos/{owner}/{repo}/protected-tags/{tag}
func (c *Client) DeleteProtectedTag(ctx context.Context, owner, repo, tag string) error {
	return c.doRequest(ctx, http.MethodDelete, pathf("/repos/%s/%s/protected-tags/%s", owner, repo, tag), nil, nil)
}

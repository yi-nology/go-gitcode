package gitcode

import (
	"context"
	"net/http"
)

// GitReference represents a git reference.
type GitReference struct {
	Ref    string     `json:"ref"`
	URL    string     `json:"url"`
	Object *GitObject `json:"object"`
}

// GitObject represents a git object (commit, tree, blob, tag).
type GitObject struct {
	SHA  string `json:"sha"`
	Type string `json:"type"`
	URL  string `json:"url"`
}

// CreateReferenceOptions specifies options for creating a reference.
type CreateReferenceOptions struct {
	Ref string `json:"ref"` // e.g. "refs/heads/feature" or "refs/tags/v1.0"
	SHA string `json:"sha"` // The SHA of the object to point to
}

// UpdateReferenceOptions specifies options for updating a reference.
type UpdateReferenceOptions struct {
	SHA   string `json:"sha"`
	Force bool   `json:"force,omitempty"`
}

// ListGitReferences lists all references of a repository.
//
// GET /repos/{owner}/{repo}/git/refs
func (c *Client) ListGitReferences(ctx context.Context, owner, repo string, opts ListOptions) ([]*GitReference, error) {
	var refs []*GitReference
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/git/refs?%s", owner, repo, opts.toQuery()), nil, &refs)
	if err != nil {
		return nil, err
	}
	return refs, nil
}

// GetGitReference gets a single reference.
//
// GET /repos/{owner}/{repo}/git/refs/{ref}
func (c *Client) GetGitReference(ctx context.Context, owner, repo, ref string) (*GitReference, error) {
	var gitRef GitReference
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/git/refs/%s", owner, repo, ref), nil, &gitRef)
	if err != nil {
		return nil, err
	}
	return &gitRef, nil
}

// CreateGitReference creates a new reference.
//
// POST /repos/{owner}/{repo}/git/refs
func (c *Client) CreateGitReference(ctx context.Context, owner, repo string, opts CreateReferenceOptions) (*GitReference, error) {
	var gitRef GitReference
	err := c.doRequest(ctx, http.MethodPost, pathf("/repos/%s/%s/git/refs", owner, repo), opts, &gitRef)
	if err != nil {
		return nil, err
	}
	return &gitRef, nil
}

// UpdateGitReference updates a reference.
//
// PATCH /repos/{owner}/{repo}/git/refs/{ref}
func (c *Client) UpdateGitReference(ctx context.Context, owner, repo, ref string, opts UpdateReferenceOptions) (*GitReference, error) {
	var gitRef GitReference
	err := c.doRequest(ctx, http.MethodPatch, pathf("/repos/%s/%s/git/refs/%s", owner, repo, ref), opts, &gitRef)
	if err != nil {
		return nil, err
	}
	return &gitRef, nil
}

// DeleteGitReference deletes a reference.
//
// DELETE /repos/{owner}/{repo}/git/refs/{ref}
func (c *Client) DeleteGitReference(ctx context.Context, owner, repo, ref string) error {
	return c.doRequest(ctx, http.MethodDelete, pathf("/repos/%s/%s/git/refs/%s", owner, repo, ref), nil, nil)
}

// ListGitRefSubPaths lists references filtered by prefix (e.g. "heads/", "tags/").
//
// GET /repos/{owner}/{repo}/git/refs/{refPrefix}
func (c *Client) ListGitRefSubPaths(ctx context.Context, owner, repo, refPrefix string) ([]*GitReference, error) {
	var refs []*GitReference
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/git/refs/%s", owner, repo, refPrefix), nil, &refs)
	if err != nil {
		return nil, err
	}
	return refs, nil
}

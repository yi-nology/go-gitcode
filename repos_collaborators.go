package gitcode

import (
	"context"
	"fmt"
	"net/http"
)

// Collaborator represents a repository collaborator.
type Collaborator struct {
	ID         int64  `json:"id"`
	Login      string `json:"login"`
	Name       string `json:"name"`
	AvatarURL  string `json:"avatar_url"`
	HTMLURL    string `json:"html_url"`
	Permission string `json:"permission,omitempty"`
}

// CollaboratorPermission represents a collaborator's permission level.
type CollaboratorPermission struct {
	Permission string `json:"permission"`
	RoleName   string `json:"role_name"`
}

// AddCollaboratorOptions specifies options for adding a collaborator.
type AddCollaboratorOptions struct {
	Permission string `json:"permission,omitempty"` // pull, push, admin
}

// ListCollaborators lists all collaborators of a repository.
//
// GET /repos/{owner}/{repo}/collaborators
func (c *Client) ListCollaborators(ctx context.Context, owner, repo string, opts ListOptions) ([]*Collaborator, error) {
	var collaborators []*Collaborator
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/collaborators?%s", owner, repo, opts.toQuery()), nil, &collaborators)
	if err != nil {
		return nil, err
	}
	return collaborators, nil
}

// IsCollaborator checks if a user is a collaborator of a repository.
//
// GET /repos/{owner}/{repo}/collaborators/{username}
func (c *Client) IsCollaborator(ctx context.Context, owner, repo, username string) (bool, error) {
	_, err := c.doRawRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/collaborators/%s", owner, repo, username))
	if err != nil {
		if isNotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// AddCollaborator adds a collaborator to a repository.
//
// PUT /repos/{owner}/{repo}/collaborators/{username}
func (c *Client) AddCollaborator(ctx context.Context, owner, repo, username string, opts *AddCollaboratorOptions) error {
	return c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/repos/%s/%s/collaborators/%s", owner, repo, username), opts, nil)
}

// RemoveCollaborator removes a collaborator from a repository.
//
// DELETE /repos/{owner}/{repo}/collaborators/{username}
func (c *Client) RemoveCollaborator(ctx context.Context, owner, repo, username string) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/repos/%s/%s/collaborators/%s", owner, repo, username), nil, nil)
}

// GetCollaboratorPermission gets the permission of a collaborator for a repository.
//
// GET /repos/{owner}/{repo}/collaborators/{username}/permission
func (c *Client) GetCollaboratorPermission(ctx context.Context, owner, repo, username string) (*CollaboratorPermission, error) {
	var perm CollaboratorPermission
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/collaborators/%s/permission", owner, repo, username), nil, &perm)
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

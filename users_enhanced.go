package gitcode

import (
	"context"
	"net/http"
)

// --- User Subscriptions (Watched Repos) ---

// ListUserWatchedRepositories lists repositories watched by a user.
//
// GET /users/{username}/subscriptions
func (c *Client) ListUserWatchedRepositories(ctx context.Context, username string, opts ListOptions) ([]*Repository, error) {
	var repos []*Repository
	err := c.doRequest(ctx, http.MethodGet, pathf("/users/%s/subscriptions?%s", username, opts.toQuery()), nil, &repos)
	if err != nil {
		return nil, err
	}
	return repos, nil
}

// ListCurrentUserWatchedRepositories lists repositories watched by the authenticated user.
//
// GET /user/subscriptions
func (c *Client) ListCurrentUserWatchedRepositories(ctx context.Context, opts ListOptions) ([]*Repository, error) {
	var repos []*Repository
	err := c.doRequest(ctx, http.MethodGet, pathf("/user/subscriptions?%s", opts.toQuery()), nil, &repos)
	if err != nil {
		return nil, err
	}
	return repos, nil
}

// --- Update Current User Profile ---

// UpdateCurrentUserOptions specifies options for updating the authenticated user's profile.
type UpdateCurrentUserOptions struct {
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
	Bio      string `json:"bio,omitempty"`
	Location string `json:"location,omitempty"`
	Website  string `json:"website,omitempty"`
}

// UpdateCurrentUser updates the authenticated user's profile.
//
// PATCH /user
func (c *Client) UpdateCurrentUser(ctx context.Context, opts UpdateCurrentUserOptions) (*User, error) {
	var user User
	err := c.doRequest(ctx, http.MethodPatch, "/user", opts, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// --- User Pull Requests ---

// ListUserPullRequests lists pull requests for the authenticated user.
//
// GET /users/merge-requests
func (c *Client) ListUserPullRequests(ctx context.Context, opts ListPullRequestsOptions) ([]*PullRequest, error) {
	var prs []*PullRequest
	query := opts.toQuery()
	if opts.State != "" {
		query += "&state=" + string(opts.State)
	}
	if opts.Sort != "" {
		query += "&sort=" + opts.Sort
	}
	if opts.Direction != "" {
		query += "&direction=" + opts.Direction
	}
	err := c.doRequest(ctx, http.MethodGet, pathf("/users/merge-requests?%s", query), nil, &prs)
	if err != nil {
		return nil, err
	}
	return prs, nil
}

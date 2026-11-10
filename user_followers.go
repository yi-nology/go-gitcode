package gitcode

import (
	"context"
	"net/http"
)

// ListUserFollowers lists the followers of a user.
//
// GET /users/{username}/followers
func (c *Client) ListUserFollowers(ctx context.Context, username string, opts ListOptions) ([]*User, error) {
	var users []*User
	err := c.doRequest(ctx, http.MethodGet, pathf("/users/%s/followers?%s", username, opts.toQuery()), nil, &users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// ListCurrentUserFollowers lists the followers of the authenticated user.
//
// GET /user/followers
func (c *Client) ListCurrentUserFollowers(ctx context.Context, opts ListOptions) ([]*User, error) {
	var users []*User
	err := c.doRequest(ctx, http.MethodGet, pathf("/user/followers?%s", opts.toQuery()), nil, &users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// ListUserFollowing lists the users that a user is following.
//
// GET /users/{username}/following
func (c *Client) ListUserFollowing(ctx context.Context, username string, opts ListOptions) ([]*User, error) {
	var users []*User
	err := c.doRequest(ctx, http.MethodGet, pathf("/users/%s/following?%s", username, opts.toQuery()), nil, &users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// ListCurrentUserFollowing lists the users that the authenticated user is following.
//
// GET /user/following
func (c *Client) ListCurrentUserFollowing(ctx context.Context, opts ListOptions) ([]*User, error) {
	var users []*User
	err := c.doRequest(ctx, http.MethodGet, pathf("/user/following?%s", opts.toQuery()), nil, &users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// IsFollowing checks if the authenticated user is following a user.
//
// GET /user/following/{username}
func (c *Client) IsFollowing(ctx context.Context, username string) (bool, error) {
	_, err := c.doRawRequest(ctx, http.MethodGet, pathf("/user/following/%s", username))
	if err != nil {
		if isNotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// FollowUser follows a user.
//
// PUT /user/following/{username}
func (c *Client) FollowUser(ctx context.Context, username string) error {
	return c.doRequest(ctx, http.MethodPut, pathf("/user/following/%s", username), nil, nil)
}

// UnfollowUser unfollows a user.
//
// DELETE /user/following/{username}
func (c *Client) UnfollowUser(ctx context.Context, username string) error {
	return c.doRequest(ctx, http.MethodDelete, pathf("/user/following/%s", username), nil, nil)
}

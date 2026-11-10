package gitcode

import (
	"context"
	"net/http"
)

// ListOrgPublicMembers lists all public members of an organization.
//
// GET /orgs/{org}/public_members
func (c *Client) ListOrgPublicMembers(ctx context.Context, org string, opts ListOptions) ([]*User, error) {
	var users []*User
	err := c.doRequest(ctx, http.MethodGet, pathf("/orgs/%s/public_members?%s", org, opts.toQuery()), nil, &users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// IsOrgPublicMember checks if a user is a public member of an organization.
//
// GET /orgs/{org}/public_members/{username}
func (c *Client) IsOrgPublicMember(ctx context.Context, org, username string) (bool, error) {
	_, err := c.doRawRequest(ctx, http.MethodGet, pathf("/orgs/%s/public_members/%s", org, username))
	if err != nil {
		if isNotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// PublicizeOrgMembership makes the authenticated user's membership public.
//
// PUT /orgs/{org}/public_members/{username}
func (c *Client) PublicizeOrgMembership(ctx context.Context, org, username string) error {
	return c.doRequest(ctx, http.MethodPut, pathf("/orgs/%s/public_members/%s", org, username), nil, nil)
}

// ConcealOrgMembership makes the authenticated user's membership private.
//
// DELETE /orgs/{org}/public_members/{username}
func (c *Client) ConcealOrgMembership(ctx context.Context, org, username string) error {
	return c.doRequest(ctx, http.MethodDelete, pathf("/orgs/%s/public_members/%s", org, username), nil, nil)
}

// ListOrgBlockedUsers lists all blocked users of an organization.
//
// GET /orgs/{org}/blocks
func (c *Client) ListOrgBlockedUsers(ctx context.Context, org string, opts ListOptions) ([]*User, error) {
	var users []*User
	err := c.doRequest(ctx, http.MethodGet, pathf("/orgs/%s/blocks?%s", org, opts.toQuery()), nil, &users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// IsOrgBlockedUser checks if a user is blocked by an organization.
//
// GET /orgs/{org}/blocks/{username}
func (c *Client) IsOrgBlockedUser(ctx context.Context, org, username string) (bool, error) {
	_, err := c.doRawRequest(ctx, http.MethodGet, pathf("/orgs/%s/blocks/%s", org, username))
	if err != nil {
		if isNotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// BlockOrgUser blocks a user from an organization.
//
// PUT /orgs/{org}/blocks/{username}
func (c *Client) BlockOrgUser(ctx context.Context, org, username string) error {
	return c.doRequest(ctx, http.MethodPut, pathf("/orgs/%s/blocks/%s", org, username), nil, nil)
}

// UnblockOrgUser unblocks a user from an organization.
//
// DELETE /orgs/{org}/blocks/{username}
func (c *Client) UnblockOrgUser(ctx context.Context, org, username string) error {
	return c.doRequest(ctx, http.MethodDelete, pathf("/orgs/%s/blocks/%s", org, username), nil, nil)
}

// CreateOrganization creates a new organization.
//
// POST /orgs
func (c *Client) CreateOrganization(ctx context.Context, opts CreateOrgOptions) (*Organization, error) {
	var org Organization
	err := c.doRequest(ctx, http.MethodPost, "/orgs", opts, &org)
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// CreateOrgOptions specifies options for creating an organization.
type CreateOrgOptions struct {
	Username    string `json:"username"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Email       string `json:"email,omitempty"`
	Location    string `json:"location,omitempty"`
	Website     string `json:"website,omitempty"`
	Visibility  string `json:"visibility,omitempty"` // public, limited, private
}

// DeleteOrganization deletes an organization.
//
// DELETE /orgs/{org}
func (c *Client) DeleteOrganization(ctx context.Context, org string) error {
	return c.doRequest(ctx, http.MethodDelete, pathf("/orgs/%s", org), nil, nil)
}

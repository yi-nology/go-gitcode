package gitcode

import (
	"context"
	"net/http"
)

func (c *Client) ListUserOrganizations(ctx context.Context, username string, opts ListOptions) ([]*Organization, error) {
	var orgs []*Organization
	err := c.doRequest(ctx, http.MethodGet, pathf("/users/%s/orgs?%s", username, opts.toQuery()), nil, &orgs)
	if err != nil {
		return nil, err
	}
	return orgs, nil
}

func (c *Client) ListOrganizationsWithOptions(ctx context.Context, admin bool, opts ListOptions) ([]*Organization, error) {
	var orgs []*Organization
	query := opts.toQuery()
	if admin {
		query += "&admin=true"
	}
	err := c.doRequest(ctx, http.MethodGet, pathf("/users/orgs?%s", query), nil, &orgs)
	if err != nil {
		return nil, err
	}
	return orgs, nil
}

type OrgMemberDetail struct {
	ID        int64  `json:"id"`
	Path      string `json:"path"`
	Name      string `json:"name"`
	URL       string `json:"url"`
	AvatarURL string `json:"avatar_url"`
	User      *User  `json:"user"`
}

func (c *Client) GetOrgMemberDetail(ctx context.Context, org, username string) (*OrgMemberDetail, error) {
	var detail OrgMemberDetail
	err := c.doRequest(ctx, http.MethodGet, pathf("/orgs/%s/members/%s", org, username), nil, &detail)
	if err != nil {
		return nil, err
	}
	return &detail, nil
}

func (c *Client) GetOrgInfo(ctx context.Context, org string) (*Organization, error) {
	var o Organization
	err := c.doRequest(ctx, http.MethodGet, pathf("/orgs/%s", org), nil, &o)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

type UpdateOrgOptions struct {
	Name        string `json:"name,omitempty"`
	Email       string `json:"email,omitempty"`
	Location    string `json:"location,omitempty"`
	Description string `json:"description,omitempty"`
	HTMLURL     string `json:"html_url,omitempty"`
}

func (c *Client) UpdateOrganization(ctx context.Context, org string, opts UpdateOrgOptions) (*Organization, error) {
	var o Organization
	err := c.doRequest(ctx, http.MethodPatch, pathf("/orgs/%s", org), opts, &o)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

type InviteMemberOptions struct {
	Permission string `json:"permission,omitempty"`
	RoleID     string `json:"role_id,omitempty"`
}

func (c *Client) InviteOrgMember(ctx context.Context, org, username string, opts InviteMemberOptions) (*User, error) {
	var user User
	err := c.doRequest(ctx, http.MethodPost, pathf("/orgs/%s/memberships/%s", org, username), opts, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *Client) RemoveOrgMember(ctx context.Context, org, username string) error {
	return c.doRequest(ctx, http.MethodDelete, pathf("/orgs/%s/memberships/%s", org, username), nil, nil)
}

type OrgFollowers struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	WatchAt   string `json:"watch_at"`
}

func (c *Client) ListOrgFollowers(ctx context.Context, org string, opts ListOptions) ([]*OrgFollowers, error) {
	var followers []*OrgFollowers
	err := c.doRequest(ctx, http.MethodGet, pathf("/orgs/%s/followers?%s", org, opts.toQuery()), nil, &followers)
	if err != nil {
		return nil, err
	}
	return followers, nil
}

type OrgMember struct {
	AvatarURL  string `json:"avatar_url"`
	HTMLURL    string `json:"html_url"`
	ID         string `json:"id"`
	Login      string `json:"login"`
	MemberRole string `json:"member_role"`
	Name       string `json:"name"`
	Type       string `json:"type"`
}

func (c *Client) ListOrgMembers(ctx context.Context, org, role string, opts ListOptions) ([]*OrgMember, error) {
	var members []*OrgMember
	query := opts.toQuery()
	if role != "" {
		query += "&role=" + role
	}
	err := c.doRequest(ctx, http.MethodGet, pathf("/orgs/%s/members?%s", org, query), nil, &members)
	if err != nil {
		return nil, err
	}
	return members, nil
}

type UserMembership struct {
	ID           int64  `json:"id"`
	Path         string `json:"path"`
	Name         string `json:"name"`
	URL          string `json:"url"`
	AvatarURL    string `json:"avatar_url"`
	User         *User  `json:"user"`
	Active       bool   `json:"active"`
	Role         string `json:"role"`
	Organization *struct {
		ID    int64  `json:"id"`
		Login string `json:"login"`
		Name  string `json:"name"`
	} `json:"organization"`
}

func (c *Client) GetUserMembership(ctx context.Context, org string) (*UserMembership, error) {
	var membership UserMembership
	err := c.doRequest(ctx, http.MethodGet, pathf("/user/memberships/orgs/%s", org), nil, &membership)
	if err != nil {
		return nil, err
	}
	return &membership, nil
}

func (c *Client) ExitOrganization(ctx context.Context, org string) error {
	return c.doRequest(ctx, http.MethodDelete, pathf("/user/memberships/orgs/%s", org), nil, nil)
}

type IssueExtendSetting struct {
	TypeName string `json:"type_name"`
	TypeID   int    `json:"type_id"`
	TypeDesc string `json:"type_desc"`
	Status   []*struct {
		StatusName         string `json:"status_name"`
		StatusID           int    `json:"status_id"`
		StatusDesc         string `json:"status_desc"`
		GitcodeIssueStatus int    `json:"gitcode_issue_status"`
	} `json:"status"`
}

func (c *Client) GetOrgIssueExtendSettings(ctx context.Context, org string) ([]*IssueExtendSetting, error) {
	var settings []*IssueExtendSetting
	err := c.doRequest(ctx, http.MethodGet, pathf("/orgs/%s/issue/extend/settings", org), nil, &settings)
	if err != nil {
		return nil, err
	}
	return settings, nil
}

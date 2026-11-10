package gitcode

import (
	"context"
	"net/http"
	"time"
)

// Team represents an organization team.
type Team struct {
	ID               int64  `json:"id"`
	Name             string `json:"name"`
	Description      string `json:"description,omitempty"`
	Permission       string `json:"permission,omitempty"` // read, write, admin
	Privacy          string `json:"privacy,omitempty"`    // closed, secret
	CanCreateOrgRepo bool   `json:"can_create_org_repo,omitempty"`
	HTMLURL          string `json:"html_url,omitempty"`
	Parent           *Team  `json:"parent,omitempty"`
}

// CreateTeamOptions specifies options for creating a team.
type CreateTeamOptions struct {
	Name             string `json:"name"`
	Description      string `json:"description,omitempty"`
	Permission       string `json:"permission,omitempty"` // none, read, write, admin
	Privacy          string `json:"privacy,omitempty"`    // closed, secret
	CanCreateOrgRepo *bool  `json:"can_create_org_repo,omitempty"`
	ParentTeamID     int64  `json:"parent_team_id,omitempty"`
}

// UpdateTeamOptions specifies options for updating a team.
type UpdateTeamOptions struct {
	Name             string `json:"name,omitempty"`
	Description      string `json:"description,omitempty"`
	Permission       string `json:"permission,omitempty"`
	Privacy          string `json:"privacy,omitempty"`
	CanCreateOrgRepo *bool  `json:"can_create_org_repo,omitempty"`
	ParentTeamID     int64  `json:"parent_team_id,omitempty"`
}

// TeamMember represents a team member.
type TeamMember struct {
	ID        int64     `json:"id"`
	Login     string    `json:"login"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	AvatarURL string    `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
}

// ListOrgTeams lists all teams in an organization.
//
// GET /orgs/{org}/teams
func (c *Client) ListOrgTeams(ctx context.Context, org string, opts ListOptions) ([]*Team, error) {
	var teams []*Team
	err := c.doRequest(ctx, http.MethodGet, pathf("/orgs/%s/teams?%s", org, opts.toQuery()), nil, &teams)
	if err != nil {
		return nil, err
	}
	return teams, nil
}

// GetTeam gets a team by ID.
//
// GET /teams/{team_id}
func (c *Client) GetTeam(ctx context.Context, teamID int64) (*Team, error) {
	var team Team
	err := c.doRequest(ctx, http.MethodGet, pathf("/teams/%d", teamID), nil, &team)
	if err != nil {
		return nil, err
	}
	return &team, nil
}

// CreateTeam creates a new team in an organization.
//
// POST /orgs/{org}/teams
func (c *Client) CreateTeam(ctx context.Context, org string, opts CreateTeamOptions) (*Team, error) {
	var team Team
	err := c.doRequest(ctx, http.MethodPost, pathf("/orgs/%s/teams", org), opts, &team)
	if err != nil {
		return nil, err
	}
	return &team, nil
}

// UpdateTeam updates a team.
//
// PATCH /teams/{team_id}
func (c *Client) UpdateTeam(ctx context.Context, teamID int64, opts UpdateTeamOptions) (*Team, error) {
	var team Team
	err := c.doRequest(ctx, http.MethodPatch, pathf("/teams/%d", teamID), opts, &team)
	if err != nil {
		return nil, err
	}
	return &team, nil
}

// DeleteTeam deletes a team.
//
// DELETE /teams/{team_id}
func (c *Client) DeleteTeam(ctx context.Context, teamID int64) error {
	return c.doRequest(ctx, http.MethodDelete, pathf("/teams/%d", teamID), nil, nil)
}

// ListTeamMembers lists all members of a team.
//
// GET /teams/{team_id}/members
func (c *Client) ListTeamMembers(ctx context.Context, teamID int64, opts ListOptions) ([]*TeamMember, error) {
	var members []*TeamMember
	err := c.doRequest(ctx, http.MethodGet, pathf("/teams/%d/members?%s", teamID, opts.toQuery()), nil, &members)
	if err != nil {
		return nil, err
	}
	return members, nil
}

// GetTeamMember gets a team member.
//
// GET /teams/{team_id}/members/{username}
func (c *Client) GetTeamMember(ctx context.Context, teamID int64, username string) (*User, error) {
	var user User
	err := c.doRequest(ctx, http.MethodGet, pathf("/teams/%d/members/%s", teamID, username), nil, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// AddTeamMember adds a member to a team.
//
// PUT /teams/{team_id}/members/{username}
func (c *Client) AddTeamMember(ctx context.Context, teamID int64, username string) error {
	return c.doRequest(ctx, http.MethodPut, pathf("/teams/%d/members/%s", teamID, username), nil, nil)
}

// RemoveTeamMember removes a member from a team.
//
// DELETE /teams/{team_id}/members/{username}
func (c *Client) RemoveTeamMember(ctx context.Context, teamID int64, username string) error {
	return c.doRequest(ctx, http.MethodDelete, pathf("/teams/%d/members/%s", teamID, username), nil, nil)
}

// ListTeamRepositories lists all repositories of a team.
//
// GET /teams/{team_id}/repos
func (c *Client) ListTeamRepositories(ctx context.Context, teamID int64, opts ListOptions) ([]*Repository, error) {
	var repos []*Repository
	err := c.doRequest(ctx, http.MethodGet, pathf("/teams/%d/repos?%s", teamID, opts.toQuery()), nil, &repos)
	if err != nil {
		return nil, err
	}
	return repos, nil
}

// AddTeamRepository adds a repository to a team.
//
// PUT /teams/{team_id}/repos/{org}/{repo}
func (c *Client) AddTeamRepository(ctx context.Context, teamID int64, org, repo string) error {
	return c.doRequest(ctx, http.MethodPut, pathf("/teams/%d/repos/%s/%s", teamID, org, repo), nil, nil)
}

// RemoveTeamRepository removes a repository from a team.
//
// DELETE /teams/{team_id}/repos/{org}/{repo}
func (c *Client) RemoveTeamRepository(ctx context.Context, teamID int64, org, repo string) error {
	return c.doRequest(ctx, http.MethodDelete, pathf("/teams/%d/repos/%s/%s", teamID, org, repo), nil, nil)
}

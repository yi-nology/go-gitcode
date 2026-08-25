package gitcode

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// --- Enterprise Members ---

// InviteEnterpriseMember invites a user to an enterprise.
//
// POST /enterprises/{enterprise}/members
func (c *Client) InviteEnterpriseMember(ctx context.Context, enterprise string, username string, opts InviteMemberOptions) error {
	return c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/enterprises/%s/members", enterprise), map[string]interface{}{
		"username":   username,
		"permission": opts.Permission,
		"role_id":    opts.RoleID,
	}, nil)
}

// DeleteEnterpriseMember removes a member from an enterprise.
//
// DELETE /enterprises/{enterprise}/members/{username}
func (c *Client) DeleteEnterpriseMember(ctx context.Context, enterprise, username string) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/enterprises/%s/members/%s", enterprise, username), nil, nil)
}

// --- Org Enterprise ---

// OrgEnterprise represents the enterprise associated with an organization.
type OrgEnterprise struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
	URL  string `json:"url"`
}

// GetOrgEnterprise gets the enterprise associated with an organization.
//
// GET /orgs/{org}/enterprise
func (c *Client) GetOrgEnterprise(ctx context.Context, org string) (*OrgEnterprise, error) {
	var ent OrgEnterprise
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/orgs/%s/enterprise", org), nil, &ent)
	if err != nil {
		return nil, err
	}
	return &ent, nil
}

// --- Enterprise Customized Roles ---

// EnterpriseCustomizedRole represents a customized role in an enterprise.
type EnterpriseCustomizedRole struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Permission  string `json:"permission,omitempty"`
}

// ListEnterpriseCustomizedRoles lists all customized roles for an enterprise.
//
// GET /enterprises/{enterprise}/customized-roles
func (c *Client) ListEnterpriseCustomizedRoles(ctx context.Context, enterprise string) ([]*EnterpriseCustomizedRole, error) {
	var roles []*EnterpriseCustomizedRole
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/enterprises/%s/customized-roles", enterprise), nil, &roles)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// --- Enterprise Milestones ---

// EnterpriseMilestone represents an enterprise milestone.
type EnterpriseMilestone struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	State       string     `json:"state"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CreateEnterpriseMilestoneOptions specifies options for creating an enterprise milestone.
type CreateEnterpriseMilestoneOptions struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	DueOn       string `json:"due_on,omitempty"`
}

// UpdateEnterpriseMilestoneOptions specifies options for updating an enterprise milestone.
type UpdateEnterpriseMilestoneOptions struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	State       string `json:"state,omitempty"`
	DueOn       string `json:"due_on,omitempty"`
}

// ListEnterpriseMilestones lists all milestones for an enterprise.
//
// GET /enterprises/{enterprise}/milestones
func (c *Client) ListEnterpriseMilestones(ctx context.Context, enterprise string, opts ListMilestonesOptions) ([]*EnterpriseMilestone, error) {
	var milestones []*EnterpriseMilestone
	query := opts.toQuery()
	if opts.State != "" {
		query += "&state=" + opts.State
	}
	if opts.Sort != "" {
		query += "&sort=" + opts.Sort
	}
	if opts.Direction != "" {
		query += "&direction=" + opts.Direction
	}
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/enterprises/%s/milestones?%s", enterprise, query), nil, &milestones)
	if err != nil {
		return nil, err
	}
	return milestones, nil
}

// GetEnterpriseMilestone gets a single enterprise milestone.
//
// GET /enterprises/{enterprise}/milestones/{id}
func (c *Client) GetEnterpriseMilestone(ctx context.Context, enterprise string, milestoneID int64) (*EnterpriseMilestone, error) {
	var milestone EnterpriseMilestone
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/enterprises/%s/milestones/%d", enterprise, milestoneID), nil, &milestone)
	if err != nil {
		return nil, err
	}
	return &milestone, nil
}

// CreateEnterpriseMilestone creates a new enterprise milestone.
//
// POST /enterprises/{enterprise}/milestones
func (c *Client) CreateEnterpriseMilestone(ctx context.Context, enterprise string, opts CreateEnterpriseMilestoneOptions) (*EnterpriseMilestone, error) {
	var milestone EnterpriseMilestone
	err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/enterprises/%s/milestones", enterprise), opts, &milestone)
	if err != nil {
		return nil, err
	}
	return &milestone, nil
}

// UpdateEnterpriseMilestone updates an enterprise milestone.
//
// PATCH /enterprises/{enterprise}/milestones/{id}
func (c *Client) UpdateEnterpriseMilestone(ctx context.Context, enterprise string, milestoneID int64, opts UpdateEnterpriseMilestoneOptions) (*EnterpriseMilestone, error) {
	var milestone EnterpriseMilestone
	err := c.doRequest(ctx, http.MethodPatch, fmt.Sprintf("/enterprises/%s/milestones/%d", enterprise, milestoneID), opts, &milestone)
	if err != nil {
		return nil, err
	}
	return &milestone, nil
}

// DeleteEnterpriseMilestone deletes an enterprise milestone.
//
// DELETE /enterprises/{enterprise}/milestones/{id}
func (c *Client) DeleteEnterpriseMilestone(ctx context.Context, enterprise string, milestoneID int64) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/enterprises/%s/milestones/%d", enterprise, milestoneID), nil, nil)
}

// ListEnterpriseMilestoneRepos lists repositories that can be associated with an enterprise milestone.
//
// GET /enterprises/{enterprise}/milestones/{id}/repos
func (c *Client) ListEnterpriseMilestoneRepos(ctx context.Context, enterprise string, milestoneID int64, opts ListOptions) ([]*Repository, error) {
	var repos []*Repository
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/enterprises/%s/milestones/%d/repos?%s", enterprise, milestoneID, opts.toQuery()), nil, &repos)
	if err != nil {
		return nil, err
	}
	return repos, nil
}

// --- Enterprise Issue Custom Fields ---

// IssueCustomField represents a custom field for enterprise issues.
type IssueCustomField struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
	Options  []struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"options,omitempty"`
}

// ListEnterpriseIssueCustomFields lists all custom fields for enterprise issues.
//
// GET /enterprises/{enterprise}/issue-custom-fields
func (c *Client) ListEnterpriseIssueCustomFields(ctx context.Context, enterprise string) ([]*IssueCustomField, error) {
	var fields []*IssueCustomField
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/enterprises/%s/issue-custom-fields", enterprise), nil, &fields)
	if err != nil {
		return nil, err
	}
	return fields, nil
}

// --- Enterprise Labels (enhanced) ---

// CreateEnterpriseLabelOptions specifies options for creating an enterprise label.
type CreateEnterpriseLabelOptions struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// UpdateEnterpriseLabelOptions specifies options for updating an enterprise label.
type UpdateEnterpriseLabelOptions struct {
	Name  string `json:"name,omitempty"`
	Color string `json:"color,omitempty"`
}

// CreateEnterpriseLabel creates a new enterprise label.
//
// POST /enterprises/{enterprise}/labels
func (c *Client) CreateEnterpriseLabel(ctx context.Context, enterprise string, opts CreateEnterpriseLabelOptions) (*EnterpriseLabel, error) {
	var label EnterpriseLabel
	err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/enterprises/%s/labels", enterprise), opts, &label)
	if err != nil {
		return nil, err
	}
	return &label, nil
}

// UpdateEnterpriseLabel updates an enterprise label.
//
// PATCH /enterprises/{enterprise}/labels/{id}
func (c *Client) UpdateEnterpriseLabel(ctx context.Context, enterprise string, labelID int64, opts UpdateEnterpriseLabelOptions) (*EnterpriseLabel, error) {
	var label EnterpriseLabel
	err := c.doRequest(ctx, http.MethodPatch, fmt.Sprintf("/enterprises/%s/labels/%d", enterprise, labelID), opts, &label)
	if err != nil {
		return nil, err
	}
	return &label, nil
}

// DeleteEnterpriseLabel deletes an enterprise label.
//
// DELETE /enterprises/{enterprise}/labels/{id}
func (c *Client) DeleteEnterpriseLabel(ctx context.Context, enterprise string, labelID int64) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/enterprises/%s/labels/%d", enterprise, labelID), nil, nil)
}

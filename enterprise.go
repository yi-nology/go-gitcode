package gitcode

// Enterprise client, speaking the documented GitCode **v8** wire
// (docs.gitcode.com "Enterprise" category — the only version the platform
// serves; the module previously carried /api/v5 name-addressed paths that
// 404 against real servers).
//
// Addressing model: enterprise-scoped endpoints take the numeric enterprise
// id (string on the path), obtainable via GetOrgEnterprise for
// org-linked enterprises. Docs-page provenance is annotated per method —
// where a page's response examples are corrupt (the label endpoints carry
// copy-pasted issue/field samples), the model follows the documented
// REQUEST contract plus the minimal response fields, and says so.
//
// Removed with the rewrite (no v8 documentation exists, so no honest wire
// shape is implementable): enterprise issue comments, enterprise issue
// labels, milestone-repos listing, and enterprise issue statuses. The v8
// docs also add UpdateEnterpriseMemberPermission, implemented here.

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// OrgEnterprise is the enterprise linked to an organization (v8 response:
// id + name).
type OrgEnterprise struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// GetOrgEnterprise returns the enterprise linked to an organization; its
// ID is the enterprise_id every other enterprise endpoint addresses.
//
// GET /v8/org/{org}/enterprise.
func (c *Client) GetOrgEnterprise(ctx context.Context, org string) (*OrgEnterprise, error) {
	var ent OrgEnterprise
	err := c.doV8Request(ctx, http.MethodGet, fmt.Sprintf("/org/%s/enterprise", org), nil, &ent)
	if err != nil {
		return nil, err
	}
	return &ent, nil
}

// --- Customized roles ---

// EnterpriseCustomizedRole is one enterprise custom role (v8 schema: string
// id/product_id/role_type, integer access_level/mgnt_mode/member_count).
type EnterpriseCustomizedRole struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	NameCN      string `json:"name_cn"`
	ProductID   string `json:"product_id"`
	AccessLevel int    `json:"access_level"`
	CreatedAt   string `json:"created_at"`
	RoleType    string `json:"role_type"`
	UpdatedAt   string `json:"updated_at"`
	Description string `json:"description"`
	MgntMode    int    `json:"mgnt_mode"`
	MemberCount int    `json:"member_count"`
}

// ListEnterpriseCustomizedRoles lists the enterprise's custom roles. This
// endpoint lives under the singular /enterprise/ prefix in the docs (the
// rest of the family uses /enterprises/).
//
// GET /v8/enterprise/{enterprise_id}/customized_roles.
func (c *Client) ListEnterpriseCustomizedRoles(ctx context.Context, enterpriseID string) ([]*EnterpriseCustomizedRole, error) {
	var roles []*EnterpriseCustomizedRole
	err := c.doV8Request(ctx, http.MethodGet, fmt.Sprintf("/enterprise/%s/customized_roles", enterpriseID), nil, &roles)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// --- Milestones ---

// EnterpriseMilestone is an enterprise milestone. created_at/updated_at and
// the date fields stay strings as served (dates are "%Y-%m-%d").
type EnterpriseMilestone struct {
	ID          int64    `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	State       string   `json:"state"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
	DueDate     string   `json:"due_date,omitempty"`
	StartDate   string   `json:"start_date,omitempty"`
	Projects    []string `json:"projects"`
}

// CreateEnterpriseMilestoneOptions is the create body. Dates are
// "%Y-%m-%d"; Projects lists the enterprise project ids to link (the
// linkable set comes from the enterprise's groups/projects endpoint).
type CreateEnterpriseMilestoneOptions struct {
	Title       string  `json:"title"`
	Description string  `json:"description,omitempty"`
	StartDate   string  `json:"start_date,omitempty"`
	DueDate     string  `json:"due_date,omitempty"`
	Projects    []int64 `json:"projects,omitempty"`
}

// UpdateEnterpriseMilestoneOptions is the PUT body; the docs do not spell
// out the update fields separately, so it mirrors the create shape with
// everything optional.
type UpdateEnterpriseMilestoneOptions struct {
	Title       string  `json:"title,omitempty"`
	Description string  `json:"description,omitempty"`
	StartDate   string  `json:"start_date,omitempty"`
	DueDate     string  `json:"due_date,omitempty"`
	State       string  `json:"state,omitempty"`
	Projects    []int64 `json:"projects,omitempty"`
}

// ListEnterpriseMilestones lists the enterprise's milestones. The list
// page's schema is not separately documented; the envelope follows the
// bare-array style of the family's other v8 lists (roles, custom fields).
//
// GET /v8/enterprises/{enterprise_id}/milestones.
func (c *Client) ListEnterpriseMilestones(ctx context.Context, enterpriseID string, opts ListOptions) ([]*EnterpriseMilestone, error) {
	var milestones []*EnterpriseMilestone
	err := c.doV8Request(ctx, http.MethodGet, fmt.Sprintf("/enterprises/%s/milestones?%s", enterpriseID, opts.toQuery()), nil, &milestones)
	if err != nil {
		return nil, err
	}
	return milestones, nil
}

// GetEnterpriseMilestone fetches one milestone.
//
// GET /v8/enterprises/{enterprise_id}/milestones/{milestone_id}.
func (c *Client) GetEnterpriseMilestone(ctx context.Context, enterpriseID, milestoneID string) (*EnterpriseMilestone, error) {
	var milestone EnterpriseMilestone
	err := c.doV8Request(ctx, http.MethodGet, fmt.Sprintf("/enterprises/%s/milestones/%s", enterpriseID, milestoneID), nil, &milestone)
	if err != nil {
		return nil, err
	}
	return &milestone, nil
}

// CreateEnterpriseMilestone creates an enterprise milestone.
//
// POST /v8/enterprises/{enterprise_id}/milestones.
func (c *Client) CreateEnterpriseMilestone(ctx context.Context, enterpriseID string, opts CreateEnterpriseMilestoneOptions) (*EnterpriseMilestone, error) {
	var milestone EnterpriseMilestone
	err := c.doV8Request(ctx, http.MethodPost, fmt.Sprintf("/enterprises/%s/milestones", enterpriseID), opts, &milestone)
	if err != nil {
		return nil, err
	}
	return &milestone, nil
}

// UpdateEnterpriseMilestone updates an enterprise milestone (PUT verb, per
// the v8 family convention).
//
// PUT /v8/enterprises/{enterprise_id}/milestones/{milestone_id}.
func (c *Client) UpdateEnterpriseMilestone(ctx context.Context, enterpriseID, milestoneID string, opts UpdateEnterpriseMilestoneOptions) (*EnterpriseMilestone, error) {
	var milestone EnterpriseMilestone
	err := c.doV8Request(ctx, http.MethodPut, fmt.Sprintf("/enterprises/%s/milestones/%s", enterpriseID, milestoneID), opts, &milestone)
	if err != nil {
		return nil, err
	}
	return &milestone, nil
}

// DeleteEnterpriseMilestone deletes an enterprise milestone.
//
// DELETE /v8/enterprises/{enterprise_id}/milestones/{milestone_id}.
func (c *Client) DeleteEnterpriseMilestone(ctx context.Context, enterpriseID, milestoneID string) error {
	return c.doV8Request(ctx, http.MethodDelete, fmt.Sprintf("/enterprises/%s/milestones/%s", enterpriseID, milestoneID), nil, nil)
}

// --- Labels ---

// EnterpriseLabel is an enterprise issue label. The v8 docs' response
// examples for the label endpoints are corrupt (issue/custom-field samples
// pasted under the label pages), so this model carries the minimal label
// contract implied by the documented request bodies; extend fields only
// after live verification.
type EnterpriseLabel struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Color       string `json:"color,omitempty"`
	Description string `json:"description,omitempty"`
}

// CreateEnterpriseLabelOptions is the documented create body: name and
// description are required, color is "#"-prefixed 6-digit hex (optional).
type CreateEnterpriseLabelOptions struct {
	Name        string `json:"name"`
	Color       string `json:"color,omitempty"`
	Description string `json:"description"`
}

// UpdateEnterpriseLabelOptions is the PUT body (fields not separately
// documented; mirrors the create shape with everything optional).
type UpdateEnterpriseLabelOptions struct {
	Name        string `json:"name,omitempty"`
	Color       string `json:"color,omitempty"`
	Description string `json:"description,omitempty"`
}

// CreateEnterpriseLabel creates an enterprise label.
//
// POST /v8/enterprises/{enterprise_id}/label (singular on the wire).
func (c *Client) CreateEnterpriseLabel(ctx context.Context, enterpriseID string, opts CreateEnterpriseLabelOptions) (*EnterpriseLabel, error) {
	var label EnterpriseLabel
	err := c.doV8Request(ctx, http.MethodPost, fmt.Sprintf("/enterprises/%s/label", enterpriseID), opts, &label)
	if err != nil {
		return nil, err
	}
	return &label, nil
}

// UpdateEnterpriseLabel updates an enterprise label.
//
// PUT /v8/enterprises/{enterprise_id}/label/{label_id}.
func (c *Client) UpdateEnterpriseLabel(ctx context.Context, enterpriseID, labelID string, opts UpdateEnterpriseLabelOptions) (*EnterpriseLabel, error) {
	var label EnterpriseLabel
	err := c.doV8Request(ctx, http.MethodPut, fmt.Sprintf("/enterprises/%s/label/%s", enterpriseID, labelID), opts, &label)
	if err != nil {
		return nil, err
	}
	return &label, nil
}

// DeleteEnterpriseLabel deletes an enterprise label.
//
// DELETE /v8/enterprises/{enterprise_id}/label/{label_id}.
func (c *Client) DeleteEnterpriseLabel(ctx context.Context, enterpriseID, labelID string) error {
	return c.doV8Request(ctx, http.MethodDelete, fmt.Sprintf("/enterprises/%s/label/%s", enterpriseID, labelID), nil, nil)
}

// --- Members ---

// EnterpriseMemberPermissions is the member permission map (schema says
// integers, the example shows booleans — FlexBool tolerates both).
type EnterpriseMemberPermissions struct {
	Admin      FlexBool `json:"admin"`
	Customized FlexBool `json:"customized"`
	Push       FlexBool `json:"push"`
	Pull       FlexBool `json:"pull"`
}

// EnterpriseMember is an enterprise member user object with its permission
// flags.
type EnterpriseMember struct {
	FollowersURL string                       `json:"followers_url,omitempty"`
	HTMLURL      string                       `json:"html_url,omitempty"`
	ID           string                       `json:"id,omitempty"`
	Login        string                       `json:"login,omitempty"`
	Name         string                       `json:"name,omitempty"`
	Type         string                       `json:"type,omitempty"`
	URL          string                       `json:"url,omitempty"`
	Permissions  *EnterpriseMemberPermissions `json:"permissions,omitempty"`
}

// EnterpriseMemberOptions carries the member permission body shared by the
// invite and permission-update calls. Permission is one of pull (参与
// 者) / push (开发者) / admin (维护者) / customized (自定义角色, requires
// RoleID); empty defaults to the server's push.
type EnterpriseMemberOptions struct {
	Permission string `json:"permission,omitempty"`
	RoleID     string `json:"role_id,omitempty"`
}

// InviteEnterpriseMember invites a user into the enterprise.
//
// POST /v8/enterprises/{enterprise_id}/memberships/{username}.
func (c *Client) InviteEnterpriseMember(ctx context.Context, enterpriseID, username string, opts EnterpriseMemberOptions) (*EnterpriseMember, error) {
	var member EnterpriseMember
	err := c.doV8Request(ctx, http.MethodPost, fmt.Sprintf("/enterprises/%s/memberships/%s", enterpriseID, username), opts, &member)
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// UpdateEnterpriseMemberPermission changes a member's permission (added in
// the v8 docs; the old surface had no equivalent).
//
// PUT /v8/enterprises/{enterprise_id}/members/{username}.
func (c *Client) UpdateEnterpriseMemberPermission(ctx context.Context, enterpriseID, username string, opts EnterpriseMemberOptions) (*EnterpriseMember, error) {
	var member EnterpriseMember
	err := c.doV8Request(ctx, http.MethodPut, fmt.Sprintf("/enterprises/%s/members/%s", enterpriseID, username), opts, &member)
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// ListEnterpriseMembers lists the enterprise's members. The list page's
// schema is not separately documented; the envelope follows the bare-array
// style of the family's other v8 lists with the per-member user shape.
//
// GET /v8/enterprises/{enterprise_id}/members.
func (c *Client) ListEnterpriseMembers(ctx context.Context, enterpriseID string, opts ListOptions) ([]*EnterpriseMember, error) {
	var members []*EnterpriseMember
	err := c.doV8Request(ctx, http.MethodGet, fmt.Sprintf("/enterprises/%s/members?%s", enterpriseID, opts.toQuery()), nil, &members)
	if err != nil {
		return nil, err
	}
	return members, nil
}

// GetEnterpriseMember fetches one member.
//
// GET /v8/enterprises/{enterprise_id}/members/{username}.
func (c *Client) GetEnterpriseMember(ctx context.Context, enterpriseID, username string) (*EnterpriseMember, error) {
	var member EnterpriseMember
	err := c.doV8Request(ctx, http.MethodGet, fmt.Sprintf("/enterprises/%s/members/%s", enterpriseID, username), nil, &member)
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// DeleteEnterpriseMember removes members (the wire's usernames segment is
// plural — several logins, comma-joined).
//
// DELETE /v8/enterprises/{enterprise_id}/members/{usernames}.
func (c *Client) DeleteEnterpriseMember(ctx context.Context, enterpriseID string, usernames ...string) error {
	return c.doV8Request(ctx, http.MethodDelete, fmt.Sprintf("/enterprises/%s/members/%s", enterpriseID, strings.Join(usernames, ",")), nil, nil)
}

// --- Issue custom fields ---

// EnterpriseCustomFieldOption is one choice of a select-type field.
type EnterpriseCustomFieldOption struct {
	ID            int64  `json:"id"`
	ExtendFieldID int64  `json:"extend_field_id"`
	SortNum       int    `json:"sort_num"`
	FieldValue    string `json:"field_value"`
	CreatedTime   string `json:"created_time"`
	UpdatedTime   string `json:"updated_time"`
}

// EnterpriseCustomFieldType links a field to an enterprise issue type.
type EnterpriseCustomFieldType struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Desc    string `json:"desc"`
	Enabled bool   `json:"enabled"`
	SortNum int    `json:"sort_num"`
	Kind    int    `json:"kind"`
}

// EnterpriseCustomField is an enterprise issue custom field. FieldType is
// one of text_single / text_multi / integer / decimal / date /
// select_single / select_multi / member_single / member_multi;
// FieldVisibility 0 = admin-only, 1 = everyone. Options only appear on
// choice-type fields.
type EnterpriseCustomField struct {
	ID              int64                          `json:"id"`
	Name            string                         `json:"name"`
	Description     string                         `json:"description"`
	FieldType       string                         `json:"field_type"`
	FieldVisibility int                            `json:"field_visibility"`
	CreatedTime     string                         `json:"created_time"`
	UpdatedTime     string                         `json:"updated_time"`
	Options         []*EnterpriseCustomFieldOption `json:"options,omitempty"`
	Types           []*EnterpriseCustomFieldType   `json:"types,omitempty"`
}

// ListEnterpriseIssueCustomFields lists the enterprise's issue custom
// fields (documented path uses the underscored issue_extend_field).
//
// GET /v8/enterprises/{enterprise_id}/issue_extend_field.
func (c *Client) ListEnterpriseIssueCustomFields(ctx context.Context, enterpriseID string, opts ListOptions) ([]*EnterpriseCustomField, error) {
	var fields []*EnterpriseCustomField
	err := c.doV8Request(ctx, http.MethodGet, fmt.Sprintf("/enterprises/%s/issue_extend_field?%s", enterpriseID, opts.toQuery()), nil, &fields)
	if err != nil {
		return nil, err
	}
	return fields, nil
}

// --- Enterprise issues ---

// EnterpriseIssueRepository is the repository reference on an enterprise
// issue.
type EnterpriseIssueRepository struct {
	ID        int64  `json:"id"`
	FullName  string `json:"full_name"`
	HumanName string `json:"human_name"`
	Path      string `json:"path"`
	Name      string `json:"name"`
	URL       string `json:"url"`
}

// EnterpriseIssueLabel is a label reference on an enterprise issue.
type EnterpriseIssueLabel struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// EnterpriseIssueCustomFieldValue is one custom-field value on an issue.
type EnterpriseIssueCustomFieldValue struct {
	FieldName   string   `json:"field_name"`
	FieldValues []string `json:"field_values"`
}

// EnterpriseLinkedPull is a pull request linked from an enterprise issue
// (core fields of the documented linked_pulls shape; merged_by's shape is
// not typed by the docs and is skipped).
type EnterpriseLinkedPull struct {
	ID              int64        `json:"id"`
	Number          string       `json:"number"`
	HTMLURL         string       `json:"html_url"`
	Title           string       `json:"title"`
	Body            string       `json:"body"`
	State           string       `json:"state"`
	User            *ActionActor `json:"user"`
	ProjectID       int64        `json:"project_id"`
	SourceProjectID int64        `json:"source_project_id"`
	TargetProjectID int64        `json:"target_project_id"`
	SourceBranch    string       `json:"source_branch"`
	TargetBranch    string       `json:"target_branch"`
	CreatedAt       string       `json:"created_at"`
	UpdatedAt       string       `json:"updated_at"`
	MergedAt        string       `json:"merged_at,omitempty"`
	ClosedAt        string       `json:"closed_at,omitempty"`
}

// EnterpriseIssue is one issue from the enterprise issue list. Visibility
// reasons: public / confidential / other.
type EnterpriseIssue struct {
	ID               int64                              `json:"id"`
	HTMLURL          string                             `json:"html_url"`
	Number           string                             `json:"number"`
	State            string                             `json:"state"`
	Title            string                             `json:"title"`
	Body             string                             `json:"body"`
	User             *ActionActor                       `json:"user"`
	Repository       *EnterpriseIssueRepository         `json:"repository"`
	CreatedAt        string                             `json:"created_at"`
	UpdatedAt        string                             `json:"updated_at"`
	FinishedAt       string                             `json:"finished_at,omitempty"`
	Deadline         string                             `json:"deadline,omitempty"`
	Labels           []*EnterpriseIssueLabel            `json:"labels,omitempty"`
	IssueState       string                             `json:"issue_state,omitempty"`
	Comments         int                                `json:"comments"`
	CustomFields     []*EnterpriseIssueCustomFieldValue `json:"custom_fields,omitempty"`
	IssueType        string                             `json:"issue_type,omitempty"`
	IssuePriority    string                             `json:"issue_priority,omitempty"`
	URL              string                             `json:"url,omitempty"`
	RepositoryURL    string                             `json:"repository_url,omitempty"`
	VisibilityReason string                             `json:"visibility_reason,omitempty"`
	RelatedBranches  []string                           `json:"related_branches,omitempty"`
	LinkedPulls      []*EnterpriseLinkedPull            `json:"linked_pulls,omitempty"`
}

// EnterpriseCustomFieldFilter is one custom-field filter of the issue
// list body. Operation: search (fuzzy), greater_equal, less_equal,
// between, in (exact).
type EnterpriseCustomFieldFilter struct {
	FieldName string   `json:"field_name"`
	Values    []string `json:"values"`
	Operation string   `json:"operation,omitempty"`
}

// ListEnterpriseIssuesOptions is the POST body of the enterprise issue
// list: filters plus page/per_page. Sort is created / updated_at.
type ListEnterpriseIssuesOptions struct {
	Page         int                            `json:"page,omitempty"`
	PerPage      int                            `json:"per_page,omitempty"`
	Labels       string                         `json:"labels,omitempty"`
	Sort         string                         `json:"sort,omitempty"`
	Direction    string                         `json:"direction,omitempty"`
	Since        string                         `json:"since,omitempty"`
	Assignees    []string                       `json:"assignees,omitempty"`
	MilestoneIDs []int64                        `json:"milestone_ids,omitempty"`
	ProjectIDs   []int64                        `json:"project_ids,omitempty"`
	CreateAt     string                         `json:"create_at,omitempty"`
	CreatedBefor string                         `json:"created_before,omitempty"`
	Search       string                         `json:"search,omitempty"`
	IssueTypes   []string                       `json:"issue_types,omitempty"`
	IssueStates  []string                       `json:"issue_states,omitempty"`
	CustomFields []*EnterpriseCustomFieldFilter `json:"custom_fields,omitempty"`
}

// ListEnterpriseIssues lists an enterprise's issues. The v8 wire makes this
// a POST (query-by-filter body); totals ride the total_count/total_page
// response headers.
//
// POST /v8/enterprises/{enterprise_id}/issues.
func (c *Client) ListEnterpriseIssues(ctx context.Context, enterpriseID string, opts ListEnterpriseIssuesOptions) ([]*EnterpriseIssue, error) {
	if opts.Page == 0 {
		opts.Page = 1
	}
	if opts.PerPage == 0 {
		opts.PerPage = 20
	}
	var issues []*EnterpriseIssue
	err := c.doV8Request(ctx, http.MethodPost, fmt.Sprintf("/enterprises/%s/issues", enterpriseID), opts, &issues)
	if err != nil {
		return nil, err
	}
	return issues, nil
}

package gitcode

import (
	"context"
	"net/http"
	"time"
)

// OrgLabel represents an organization-level label.
type OrgLabel struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	Exclusive bool      `json:"exclusive,omitempty"`
	Template  bool      `json:"template,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateOrgLabelOptions specifies options for creating an organization label.
type CreateOrgLabelOptions struct {
	Name      string `json:"name"`
	Color     string `json:"color"`
	Exclusive *bool  `json:"exclusive,omitempty"`
	Template  *bool  `json:"template,omitempty"`
}

// UpdateOrgLabelOptions specifies options for updating an organization label.
type UpdateOrgLabelOptions struct {
	Name      string `json:"name,omitempty"`
	Color     string `json:"color,omitempty"`
	Exclusive *bool  `json:"exclusive,omitempty"`
	Template  *bool  `json:"template,omitempty"`
}

// ListOrgLabels lists all labels of an organization.
//
// GET /orgs/{org}/labels
func (c *Client) ListOrgLabels(ctx context.Context, org string, opts ListOptions) ([]*OrgLabel, error) {
	var labels []*OrgLabel
	err := c.doRequest(ctx, http.MethodGet, pathf("/orgs/%s/labels?%s", org, opts.toQuery()), nil, &labels)
	if err != nil {
		return nil, err
	}
	return labels, nil
}

// GetOrgLabel gets a single organization label.
//
// GET /orgs/{org}/labels/{id}
func (c *Client) GetOrgLabel(ctx context.Context, org string, labelID int64) (*OrgLabel, error) {
	var label OrgLabel
	err := c.doRequest(ctx, http.MethodGet, pathf("/orgs/%s/labels/%d", org, labelID), nil, &label)
	if err != nil {
		return nil, err
	}
	return &label, nil
}

// CreateOrgLabel creates a new organization label.
//
// POST /orgs/{org}/labels
func (c *Client) CreateOrgLabel(ctx context.Context, org string, opts CreateOrgLabelOptions) (*OrgLabel, error) {
	var label OrgLabel
	err := c.doRequest(ctx, http.MethodPost, pathf("/orgs/%s/labels", org), opts, &label)
	if err != nil {
		return nil, err
	}
	return &label, nil
}

// UpdateOrgLabel updates an organization label.
//
// PATCH /orgs/{org}/labels/{id}
func (c *Client) UpdateOrgLabel(ctx context.Context, org string, labelID int64, opts UpdateOrgLabelOptions) (*OrgLabel, error) {
	var label OrgLabel
	err := c.doRequest(ctx, http.MethodPatch, pathf("/orgs/%s/labels/%d", org, labelID), opts, &label)
	if err != nil {
		return nil, err
	}
	return &label, nil
}

// DeleteOrgLabel deletes an organization label.
//
// DELETE /orgs/{org}/labels/{id}
func (c *Client) DeleteOrgLabel(ctx context.Context, org string, labelID int64) error {
	return c.doRequest(ctx, http.MethodDelete, pathf("/orgs/%s/labels/%d", org, labelID), nil, nil)
}

package gitcode

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type ListMilestonesOptions struct {
	ListOptions
	State     string `json:"state,omitempty"`
	Sort      string `json:"sort,omitempty"`
	Direction string `json:"direction,omitempty"`
}

func (c *Client) ListMilestonesWithOptions(ctx context.Context, owner, repo string, opts ListMilestonesOptions) ([]*Milestone, error) {
	var milestones []*Milestone
	q := url.Values{}
	q.Set("page", fmt.Sprint(opts.withDefaults().Page))
	q.Set("per_page", fmt.Sprint(opts.withDefaults().PerPage))
	if opts.State != "" {
		q.Set("state", opts.State)
	}
	if opts.Sort != "" {
		q.Set("sort", opts.Sort)
	}
	if opts.Direction != "" {
		q.Set("direction", opts.Direction)
	}
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/milestones?%s", owner, repo, q.Encode()), nil, &milestones)
	if err != nil {
		return nil, err
	}
	return milestones, nil
}

func (c *Client) GetMilestone(ctx context.Context, owner, repo string, number int) (*Milestone, error) {
	var milestone Milestone
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/milestones/%d", owner, repo, number), nil, &milestone)
	if err != nil {
		return nil, err
	}
	return &milestone, nil
}

// UpdateMilestoneOptions is the body of PATCH
// /repos/{owner}/{repo}/milestones/{number}. The docs mark title and
// due_on as required with minLength 1; in practice the server accepts
// their omission, so unset fields stay off the wire (omitempty) instead
// of wiping values with empty strings — an explicit "" due_on would clear
// the milestone's due date.
type UpdateMilestoneOptions struct {
	Title       string `json:"title,omitempty"`
	State       string `json:"state,omitempty"`
	Description string `json:"description,omitempty"`
	DueOn       string `json:"due_on,omitempty"`
}

func (c *Client) UpdateMilestone(ctx context.Context, owner, repo string, number int, opts UpdateMilestoneOptions) (*Milestone, error) {
	var milestone Milestone
	err := c.doRequest(ctx, http.MethodPatch, fmt.Sprintf("/repos/%s/%s/milestones/%d", owner, repo, number), opts, &milestone)
	if err != nil {
		return nil, err
	}
	return &milestone, nil
}

type CreateMilestoneOptions struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	// DueOn is a date such as "2026-01-31" (the documented wire shape).
	DueOn string `json:"due_on,omitempty"`
}

func (c *Client) CreateMilestoneWithOptions(ctx context.Context, owner, repo string, opts CreateMilestoneOptions) (*Milestone, error) {
	var milestone Milestone
	err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/milestones", owner, repo), opts, &milestone)
	if err != nil {
		return nil, err
	}
	return &milestone, nil
}

package gitcode

import (
	"context"
	"fmt"
	"net/http"
)

type UpdateLabelOptions struct {
	Name  string `json:"name,omitempty"`
	Color string `json:"color,omitempty"`
}

func (c *Client) UpdateIssueLabel(ctx context.Context, owner, repo, originalName string, opts UpdateLabelOptions) (*Label, error) {
	var label Label
	err := c.doRequest(ctx, http.MethodPatch, fmt.Sprintf("/repos/%s/%s/labels/%s", owner, repo, originalName), opts, &label)
	if err != nil {
		return nil, err
	}
	return &label, nil
}

func (c *Client) ReplaceIssueLabels(ctx context.Context, owner, repo string, number int, labels []string) error {
	return c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/repos/%s/%s/issues/%d/labels", owner, repo, number), labels, nil)
}

func (c *Client) RemoveAllIssueLabels(ctx context.Context, owner, repo string, number int) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/repos/%s/%s/issues/%d/labels", owner, repo, number), nil, nil)
}

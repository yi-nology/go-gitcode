package gitcode

import (
	"context"
	"net/http"
)

// IssueTemplate represents a repository issue template.
type IssueTemplate struct {
	Name        string   `json:"name"`
	FileName    string   `json:"file_name,omitempty"`
	Title       string   `json:"title"`
	Body        string   `json:"body"`
	Labels      []string `json:"labels,omitempty"`
	Assignees   []string `json:"assignees,omitempty"`
	Description string   `json:"description,omitempty"`
}

// PRMergeTemplate represents a pull request merge template.
type PRMergeTemplate struct {
	Name string `json:"name"`
	Body string `json:"body"`
}

// ListIssueTemplates lists all issue templates of a repository.
//
// GET /repos/{owner}/{repo}/issue_templates
func (c *Client) ListIssueTemplates(ctx context.Context, owner, repo string) ([]*IssueTemplate, error) {
	var templates []*IssueTemplate
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/issue_templates", owner, repo), nil, &templates)
	if err != nil {
		return nil, err
	}
	return templates, nil
}

// GetIssueTemplate gets a single issue template by name.
//
// GET /repos/{owner}/{repo}/issue_templates/{name}
func (c *Client) GetIssueTemplate(ctx context.Context, owner, repo, name string) (*IssueTemplate, error) {
	var template IssueTemplate
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/issue_templates/%s", owner, repo, name), nil, &template)
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// ListPullRequestMergeTemplates lists all pull request merge templates of a repository.
//
// GET /repos/{owner}/{repo}/merge_templates
func (c *Client) ListPullRequestMergeTemplates(ctx context.Context, owner, repo string) ([]*PRMergeTemplate, error) {
	var templates []*PRMergeTemplate
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/merge_templates", owner, repo), nil, &templates)
	if err != nil {
		return nil, err
	}
	return templates, nil
}

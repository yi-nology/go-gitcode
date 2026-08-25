package gitcode

import (
	"context"
	"fmt"
	"net/http"
)

// WikiPage represents a wiki page.
type WikiPage struct {
	Title      string `json:"title"`
	Content    string `json:"content,omitempty"`
	HTMLURL    string `json:"html_url,omitempty"`
	CommitSHA  string `json:"commit_sha,omitempty"`
	Sidebar    string `json:"sidebar,omitempty"`
	Footer     string `json:"footer,omitempty"`
	LastCommit *struct {
		SHA     string `json:"sha"`
		Message string `json:"message"`
	} `json:"last_commit,omitempty"`
}

// CreateWikiPageOptions specifies options for creating a wiki page.
type CreateWikiPageOptions struct {
	Title         string `json:"title"`
	ContentBase64 string `json:"content_base64,omitempty"` // base64 encoded content
	Message       string `json:"message,omitempty"`
}

// UpdateWikiPageOptions specifies options for updating a wiki page.
type UpdateWikiPageOptions struct {
	ContentBase64 string `json:"content_base64,omitempty"` // base64 encoded content
	Message       string `json:"message,omitempty"`
}

// ListWikiPages lists all wiki pages of a repository.
//
// GET /repos/{owner}/{repo}/wiki/pages
func (c *Client) ListWikiPages(ctx context.Context, owner, repo string, opts ListOptions) ([]*WikiPage, error) {
	var pages []*WikiPage
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/wiki/pages?%s", owner, repo, opts.toQuery()), nil, &pages)
	if err != nil {
		return nil, err
	}
	return pages, nil
}

// GetWikiPage gets a single wiki page by title.
//
// GET /repos/{owner}/{repo}/wiki/page/{pageName}
func (c *Client) GetWikiPage(ctx context.Context, owner, repo, pageName string) (*WikiPage, error) {
	var page WikiPage
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/wiki/page/%s", owner, repo, pageName), nil, &page)
	if err != nil {
		return nil, err
	}
	return &page, nil
}

// CreateWikiPage creates a new wiki page.
//
// POST /repos/{owner}/{repo}/wiki/new
func (c *Client) CreateWikiPage(ctx context.Context, owner, repo string, opts CreateWikiPageOptions) (*WikiPage, error) {
	var page WikiPage
	err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/wiki/new", owner, repo), opts, &page)
	if err != nil {
		return nil, err
	}
	return &page, nil
}

// UpdateWikiPage updates an existing wiki page.
//
// PATCH /repos/{owner}/{repo}/wiki/page/{pageName}
func (c *Client) UpdateWikiPage(ctx context.Context, owner, repo, pageName string, opts UpdateWikiPageOptions) (*WikiPage, error) {
	var page WikiPage
	err := c.doRequest(ctx, http.MethodPatch, fmt.Sprintf("/repos/%s/%s/wiki/page/%s", owner, repo, pageName), opts, &page)
	if err != nil {
		return nil, err
	}
	return &page, nil
}

// DeleteWikiPage deletes a wiki page.
//
// DELETE /repos/{owner}/{repo}/wiki/page/{pageName}
func (c *Client) DeleteWikiPage(ctx context.Context, owner, repo, pageName string) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/repos/%s/%s/wiki/page/%s", owner, repo, pageName), nil, nil)
}

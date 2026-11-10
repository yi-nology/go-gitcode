package gitcode

import (
	"context"
	"net/http"
)

// --- Gitignore Templates ---

// ListGitignoreTemplates lists all available gitignore templates.
//
// GET /gitignore/templates
func (c *Client) ListGitignoreTemplates(ctx context.Context) ([]string, error) {
	var templates []string
	err := c.doRequest(ctx, http.MethodGet, "/gitignore/templates", nil, &templates)
	if err != nil {
		return nil, err
	}
	return templates, nil
}

// GetGitignoreTemplate gets a gitignore template by name.
//
// GET /gitignore/templates/{name}
func (c *Client) GetGitignoreTemplate(ctx context.Context, name string) (*GitignoreTemplate, error) {
	var template GitignoreTemplate
	err := c.doRequest(ctx, http.MethodGet, pathf("/gitignore/templates/%s", name), nil, &template)
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// GitignoreTemplate represents a gitignore template.
type GitignoreTemplate struct {
	Name   string `json:"name"`
	Source string `json:"source"`
}

// --- License Templates ---

// ListLicenseTemplates lists all available license templates.
//
// GET /licenses
func (c *Client) ListLicenseTemplates(ctx context.Context) ([]*LicenseTemplate, error) {
	var templates []*LicenseTemplate
	err := c.doRequest(ctx, http.MethodGet, "/licenses", nil, &templates)
	if err != nil {
		return nil, err
	}
	return templates, nil
}

// GetLicenseTemplate gets a license template by key.
//
// GET /licenses/{name}
func (c *Client) GetLicenseTemplate(ctx context.Context, name string) (*LicenseTemplate, error) {
	var template LicenseTemplate
	err := c.doRequest(ctx, http.MethodGet, pathf("/licenses/%s", name), nil, &template)
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// LicenseTemplate represents a license template.
type LicenseTemplate struct {
	Key         string   `json:"key"`
	Name        string   `json:"name"`
	URL         string   `json:"url,omitempty"`
	HTMLURL     string   `json:"html_url,omitempty"`
	HTMLContent string   `json:"html_content,omitempty"`
	Body        string   `json:"body,omitempty"`
	Featured    bool     `json:"featured,omitempty"`
	Conditions  []string `json:"conditions,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	Limitations []string `json:"limitations,omitempty"`
}

// --- Label Templates ---

// ListLabelTemplates lists all available label templates.
//
// GET /label/templates
func (c *Client) ListLabelTemplates(ctx context.Context) ([]string, error) {
	var templates []string
	err := c.doRequest(ctx, http.MethodGet, "/label/templates", nil, &templates)
	if err != nil {
		return nil, err
	}
	return templates, nil
}

// GetLabelTemplate gets a label template by name.
//
// GET /label/templates/{name}
func (c *Client) GetLabelTemplate(ctx context.Context, name string) ([]*Label, error) {
	var labels []*Label
	err := c.doRequest(ctx, http.MethodGet, pathf("/label/templates/%s", name), nil, &labels)
	if err != nil {
		return nil, err
	}
	return labels, nil
}

// --- Markdown Rendering ---

// RenderMarkdown renders a markdown document as HTML.
//
// POST /markdown
func (c *Client) RenderMarkdown(ctx context.Context, text, mode, context string) (string, error) {
	body := map[string]string{
		"text": text,
	}
	if mode != "" {
		body["mode"] = mode // gfm, comment
	}
	if context != "" {
		body["context"] = context // owner/repo for context
	}
	var result string
	err := c.doRequest(ctx, http.MethodPost, "/markdown", body, &result)
	if err != nil {
		return "", err
	}
	return result, nil
}

// RenderMarkdownRaw renders raw markdown as HTML.
//
// POST /markdown/raw
func (c *Client) RenderMarkdownRaw(ctx context.Context, markdown string) (string, error) {
	var result string
	err := c.doRequest(ctx, http.MethodPost, "/markdown/raw", markdown, &result)
	if err != nil {
		return "", err
	}
	return result, nil
}

// --- User Repositories ---

// ListCurrentUserRepositories lists the authenticated user's repositories.
//
// GET /user/repos
func (c *Client) ListCurrentUserRepositories(ctx context.Context, opts ListRepositoriesOptions) ([]*Repository, error) {
	var repos []*Repository
	// opts.toQuery() is ListRepositoriesOptions.toQuery, which already
	// carries type/sort — the same helper ListRepositories uses, so the
	// two entry points cannot diverge again (issue #10).
	err := c.doRequest(ctx, http.MethodGet, pathf("/user/repos?%s", opts.toQuery()), nil, &repos)
	if err != nil {
		return nil, err
	}
	return repos, nil
}

// ListUserRepositories lists a user's public repositories.
//
// GET /users/{username}/repos
func (c *Client) ListUserRepositories(ctx context.Context, username string, opts ListOptions) ([]*Repository, error) {
	var repos []*Repository
	err := c.doRequest(ctx, http.MethodGet, pathf("/users/%s/repos?%s", username, opts.toQuery()), nil, &repos)
	if err != nil {
		return nil, err
	}
	return repos, nil
}

// --- User Info ---

// ExtendedUser represents extended user information.
type ExtendedUser struct {
	ID                int64  `json:"id"`
	Login             string `json:"login"`
	Name              string `json:"name"`
	Email             string `json:"email"`
	AvatarURL         string `json:"avatar_url"`
	HTMLURL           string `json:"html_url"`
	Type              string `json:"type"`
	Bio               string `json:"bio,omitempty"`
	Location          string `json:"location,omitempty"`
	Website           string `json:"website,omitempty"`
	FullName          string `json:"full_name,omitempty"`
	FollowersCount    int    `json:"followers_count,omitempty"`
	FollowingCount    int    `json:"following_count,omitempty"`
	StarredReposCount int    `json:"starred_repos_count,omitempty"`
	Username          string `json:"username,omitempty"`
}

// GetExtendedUser gets extended information about a user.
//
// GET /users/{username}
func (c *Client) GetExtendedUser(ctx context.Context, username string) (*ExtendedUser, error) {
	var user ExtendedUser
	err := c.doRequest(ctx, http.MethodGet, pathf("/users/%s", username), nil, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// --- Repo Visibility ---

// GetRepoVisibility returns whether a repository is public or private.
//
// GET /repos/{owner}/{repo}
func (c *Client) GetRepoVisibility(ctx context.Context, owner, repo string) (bool, error) {
	r, err := c.GetRepository(ctx, owner, repo)
	if err != nil {
		return false, err
	}
	return r.Private, nil
}

// SetRepoVisibility sets a repository's visibility (public/private).
//
// PATCH /repos/{owner}/{repo}
func (c *Client) SetRepoVisibility(ctx context.Context, owner, repo string, private bool) (*Repository, error) {
	return c.UpdateRepository(ctx, owner, repo, UpdateRepositoryOptions{
		Private: &private,
	})
}

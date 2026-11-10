package gitcode

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

type Email struct {
	Email string `json:"email"`
	State string `json:"state"`
}

func (c *Client) ListEmails(ctx context.Context) ([]*Email, error) {
	var emails []*Email
	err := c.doRequest(ctx, http.MethodGet, "/emails", nil, &emails)
	if err != nil {
		return nil, err
	}
	return emails, nil
}

type SSHKey struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Key       string    `json:"key"`
	CreatedAt time.Time `json:"created_at"`
	URL       string    `json:"url"`
}

type CreateSSHKeyOptions struct {
	Title string `json:"title"`
	Key   string `json:"key"`
}

func (c *Client) ListSSHKeys(ctx context.Context, opts ListOptions) ([]*SSHKey, error) {
	var keys []*SSHKey
	err := c.doRequest(ctx, http.MethodGet, pathf("/user/keys?%s", opts.toQuery()), nil, &keys)
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (c *Client) GetSSHKey(ctx context.Context, id int64) (*SSHKey, error) {
	var key SSHKey
	err := c.doRequest(ctx, http.MethodGet, pathf("/user/keys/%d", id), nil, &key)
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func (c *Client) CreateSSHKey(ctx context.Context, opts CreateSSHKeyOptions) (*SSHKey, error) {
	var key SSHKey
	err := c.doRequest(ctx, http.MethodPost, "/user/keys", opts, &key)
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func (c *Client) DeleteSSHKey(ctx context.Context, id int64) error {
	return c.doRequest(ctx, http.MethodDelete, pathf("/user/keys/%d", id), nil, nil)
}

type UserEvent struct {
	Action         int    `json:"action"`
	ActionName     string `json:"action_name"`
	AuthorID       int64  `json:"author_id"`
	AuthorUsername string `json:"author_username"`
	ProjectID      int64  `json:"project_id"`
	ProjectName    string `json:"project_name"`
	CreatedAt      string `json:"created_at"`
	PushData       *struct {
		CommitCount int    `json:"commit_count"`
		Action      string `json:"action"`
		RefType     string `json:"ref_type"`
		Ref         string `json:"ref"`
		CommitFrom  string `json:"commit_from"`
		CommitTo    string `json:"commit_to"`
		CommitTitle string `json:"commit_title"`
	} `json:"push_data"`
}

type UserEventsResponse struct {
	Events map[string][]*UserEvent `json:"events"`
	Next   string                  `json:"next"`
}

func (c *Client) GetUserEvents(ctx context.Context, username, year, next string) (*UserEventsResponse, error) {
	var resp UserEventsResponse
	q := url.Values{}
	if year != "" {
		q.Set("year", year)
	}
	if next != "" {
		q.Set("next", next)
	}
	path := pathf("/users/%s/events", username)
	if len(q) > 0 {
		path += "?" + q.Encode()
	}
	err := c.doRequest(ctx, http.MethodGet, path, nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListStarredReposOptions struct {
	ListOptions
	Sort      string `json:"sort,omitempty"`
	Direction string `json:"direction,omitempty"`
}

func (c *Client) ListStarredRepositories(ctx context.Context, opts ListStarredReposOptions) ([]*Repository, error) {
	var repos []*Repository
	query := opts.toQuery()
	if opts.Sort != "" {
		query += "&sort=" + url.QueryEscape(opts.Sort)
	}
	if opts.Direction != "" {
		query += "&direction=" + url.QueryEscape(opts.Direction)
	}
	err := c.doRequest(ctx, http.MethodGet, pathf("/user/starred?%s", query), nil, &repos)
	if err != nil {
		return nil, err
	}
	return repos, nil
}

type Namespace struct {
	ID      int64  `json:"id"`
	Path    string `json:"path"`
	Name    string `json:"name"`
	HTMLURL string `json:"html_url"`
	Type    string `json:"type"`
}

func (c *Client) GetNamespace(ctx context.Context, path string) (*Namespace, error) {
	var ns Namespace
	err := c.doRequest(ctx, http.MethodGet, pathf("/user/namespace?path=%s", url.QueryEscape(path)), nil, &ns)
	if err != nil {
		return nil, err
	}
	return &ns, nil
}

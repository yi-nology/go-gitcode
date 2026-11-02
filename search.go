package gitcode

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type SearchRepositoriesOptions struct {
	ListOptions
	Query    string `json:"q"`
	Sort     string `json:"sort,omitempty"`
	Order    string `json:"order,omitempty"`
	Owner    string `json:"owner,omitempty"`
	Fork     string `json:"fork,omitempty"`
	Language string `json:"language,omitempty"`
}

type SearchRepositoryResult struct {
	ID              int64     `json:"id"`
	FullName        string    `json:"full_name"`
	HumanName       string    `json:"human_name"`
	Path            string    `json:"path"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	SSHURLToRepo    string    `json:"ssh_url_to_repo"`
	HTTPURLToRepo   string    `json:"http_url_to_repo"`
	WebURL          string    `json:"web_url"`
	ForksCount      int       `json:"forks_count"`
	StargazersCount int       `json:"stargazers_count"`
	WatchersCount   int       `json:"watchers_count"`
	DefaultBranch   string    `json:"default_branch"`
	OpenIssuesCount int       `json:"open_issues_count"`
	Private         bool      `json:"private"`
	Public          bool      `json:"public"`
	Fork            bool      `json:"fork"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	PushedAt        string    `json:"pushed_at"`
	Owner           *User     `json:"owner"`
	Namespace       *struct {
		ID      int64  `json:"id"`
		Type    string `json:"type"`
		Name    string `json:"name"`
		Path    string `json:"path"`
		HTMLURL string `json:"html_url"`
	} `json:"namespace"`
}

func (c *Client) SearchRepositories(ctx context.Context, opts SearchRepositoriesOptions) ([]*SearchRepositoryResult, error) {
	var repos []*SearchRepositoryResult
	query := fmt.Sprintf("q=%s&%s", opts.Query, opts.toQuery())
	if opts.Sort != "" {
		query += "&sort=" + opts.Sort
	}
	if opts.Order != "" {
		query += "&order=" + opts.Order
	}
	if opts.Owner != "" {
		query += "&owner=" + opts.Owner
	}
	if opts.Fork != "" {
		query += "&fork=" + opts.Fork
	}
	if opts.Language != "" {
		query += "&language=" + opts.Language
	}
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/search/repositories?%s", query), nil, &repos)
	if err != nil {
		return nil, err
	}
	return repos, nil
}

type SearchIssuesOptions struct {
	ListOptions
	Query string `json:"q"`
	Sort  string `json:"sort,omitempty"`
	Order string `json:"order,omitempty"`
	Repo  string `json:"repo,omitempty"`
	State string `json:"state,omitempty"`
}

type SearchIssueResult struct {
	ID         int64     `json:"id"`
	HTMLURL    string    `json:"html_url"`
	Number     string    `json:"number"`
	State      string    `json:"state"`
	Title      string    `json:"title"`
	Body       string    `json:"body"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Labels     []*Label  `json:"labels"`
	Priority   int       `json:"priority"`
	Comments   int       `json:"comments"`
	Repository *struct {
		ID        int64  `json:"id"`
		FullName  string `json:"full_name"`
		HumanName string `json:"human_name"`
		Path      string `json:"path"`
		Name      string `json:"name"`
		URL       string `json:"url"`
		Owner     *User  `json:"owner"`
	} `json:"repository"`
}

func (c *Client) SearchIssues(ctx context.Context, opts SearchIssuesOptions) ([]*SearchIssueResult, error) {
	var issues []*SearchIssueResult
	query := fmt.Sprintf("q=%s&%s", opts.Query, opts.toQuery())
	if opts.Sort != "" {
		query += "&sort=" + opts.Sort
	}
	if opts.Order != "" {
		query += "&order=" + opts.Order
	}
	if opts.Repo != "" {
		query += "&repo=" + opts.Repo
	}
	if opts.State != "" {
		query += "&state=" + opts.State
	}
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/search/issues?%s", query), nil, &issues)
	if err != nil {
		return nil, err
	}
	return issues, nil
}

type SearchUsersOptions struct {
	ListOptions
	Query string `json:"q"`
	Sort  string `json:"sort,omitempty"`
	Order string `json:"order,omitempty"`
}

type SearchUserResult struct {
	ID        string    `json:"id"`
	Login     string    `json:"login"`
	Name      string    `json:"name"`
	AvatarURL string    `json:"avatar_url"`
	HTMLURL   string    `json:"html_url"`
	CreatedAt time.Time `json:"created_at"`
}

func (c *Client) SearchUsers(ctx context.Context, opts SearchUsersOptions) ([]*SearchUserResult, error) {
	var users []*SearchUserResult
	query := fmt.Sprintf("q=%s&%s", opts.Query, opts.toQuery())
	if opts.Sort != "" {
		query += "&sort=" + opts.Sort
	}
	if opts.Order != "" {
		query += "&order=" + opts.Order
	}
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/search/users?%s", query), nil, &users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

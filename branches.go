package gitcode

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type Branch struct {
	Name          string              `json:"name"`
	Commit        *BranchCommitDetail `json:"commit"`
	DefaultBranch bool                `json:"default_branch,omitempty"`
	Protected     bool                `json:"protected"`
}

type Commit struct {
	SHA       string    `json:"sha"`
	Message   string    `json:"message"`
	Author    *User     `json:"author"`
	Committer *User     `json:"committer"`
	CreatedAt time.Time `json:"created_at"`
	URL       string    `json:"url,omitempty"`
}

type BranchCommitDetail struct {
	SHA    string `json:"sha"`
	URL    string `json:"url"`
	Commit *struct {
		Author    *CommitAuthor `json:"author"`
		Committer *CommitAuthor `json:"committer"`
		Message   string        `json:"message"`
	} `json:"commit"`
}

type CommitAuthor struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Date  string `json:"date"`
}

type BranchProtection struct {
	Enabled                  bool `json:"enabled"`
	RequiredStatusChecks     bool `json:"required_status_checks"`
	RequiredApprovingReviews int  `json:"required_approving_reviews"`
	AllowForcePushes         bool `json:"allow_force_pushes"`
	AllowDeletions           bool `json:"allow_deletions"`
}

func (c *Client) ListBranches(ctx context.Context, owner, repo string) ([]*Branch, error) {
	var branches []*Branch
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/branches?per_page=100", owner, repo), nil, &branches)
	if err != nil {
		return nil, err
	}
	return branches, nil
}

func (c *Client) GetBranch(ctx context.Context, owner, repo, branch string) (*Branch, error) {
	var b Branch
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/branches/%s", owner, repo, branch), nil, &b)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

type CreateBranchOptions struct {
	BranchName string `json:"branch_name"`
	Refs       string `json:"refs"`
	Ref        string `json:"ref,omitempty"`
}

func (c *Client) CreateBranch(ctx context.Context, owner, repo string, opts CreateBranchOptions) (*Branch, error) {
	var b Branch
	err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/branches", owner, repo), opts, &b)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (c *Client) DeleteBranch(ctx context.Context, owner, repo, branch string) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/repos/%s/%s/branches/%s", owner, repo, branch), nil, nil)
}

type ListBranchProtectionOptions struct {
	ListOptions
}

func (c *Client) DeleteBranchProtection(ctx context.Context, owner, repo, name string) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/repos/%s/%s/branches/%s/setting", owner, repo, name), nil, nil)
}

type CommitComparison struct {
	TotalCommits int                `json:"total_commits"`
	AheadBy      int                `json:"ahead_by"`
	BehindBy     int                `json:"behind_by"`
	Commits      []*Commit          `json:"commits"`
	Files        []*PullRequestFile `json:"files"`
}

func (c *Client) CompareCommits(ctx context.Context, owner, repo, base, head string) (*CommitComparison, error) {
	var cmp CommitComparison
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/compare/%s...%s", owner, repo, base, head), nil, &cmp)
	if err != nil {
		return nil, err
	}
	return &cmp, nil
}

func (c *Client) GetCommit(ctx context.Context, owner, repo, sha string) (*Commit, error) {
	var commit Commit
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/commits/%s", owner, repo, sha), nil, &commit)
	if err != nil {
		return nil, err
	}
	return &commit, nil
}

type ListCommitsOptions struct {
	ListOptions
	Branch string `json:"branch,omitempty"`
	Since  string `json:"since,omitempty"`
	Until  string `json:"until,omitempty"`
}

func (c *Client) ListCommits(ctx context.Context, owner, repo string, opts ListCommitsOptions) ([]*Commit, error) {
	var commits []*Commit
	query := opts.toQuery()
	if opts.Branch != "" {
		query += "&sha=" + opts.Branch
	}
	if opts.Since != "" {
		query += "&since=" + opts.Since
	}
	if opts.Until != "" {
		query += "&until=" + opts.Until
	}
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/commits?%s", owner, repo, query), nil, &commits)
	if err != nil {
		return nil, err
	}
	return commits, nil
}

type UpdateBranchProtectionOptions struct {
	Pusher string `json:"pusher"`
	Merger string `json:"merger"`
}

func (c *Client) UpdateBranchProtection(ctx context.Context, owner, repo, wildcard string, opts UpdateBranchProtectionOptions) error {
	return c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/repos/%s/%s/branches/%s/setting", owner, repo, wildcard), opts, nil)
}

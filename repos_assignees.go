package gitcode

import (
	"context"
	"net/http"
)

// ListRepoAssignees lists all available assignees for issues in a repository.
//
// GET /repos/{owner}/{repo}/assignees
func (c *Client) ListRepoAssignees(ctx context.Context, owner, repo string, opts ListOptions) ([]*User, error) {
	var users []*User
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/assignees?%s", owner, repo, opts.toQuery()), nil, &users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// IsRepoAssignee checks if a user is an assignee for issues in a repository.
//
// GET /repos/{owner}/{repo}/assignees/{username}
func (c *Client) IsRepoAssignee(ctx context.Context, owner, repo, username string) (bool, error) {
	_, err := c.doRawRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/assignees/%s", owner, repo, username))
	if err != nil {
		if isNotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// IssueAssignee represents an issue assignee operation result.
type IssueAssigneeResult struct {
	Assignees []*User `json:"assignees"`
}

// AddIssueAssignees adds assignees to an issue.
//
// POST /repos/{owner}/{repo}/issues/{index}/assignees
func (c *Client) AddIssueAssignees(ctx context.Context, owner, repo string, number int, assignees []string) (*Issue, error) {
	var issue Issue
	err := c.doRequest(ctx, http.MethodPost, pathf("/repos/%s/%s/issues/%d/assignees", owner, repo, number), map[string]interface{}{"assignees": assignees}, &issue)
	if err != nil {
		return nil, err
	}
	return &issue, nil
}

// RemoveIssueAssignees removes assignees from an issue.
//
// DELETE /repos/{owner}/{repo}/issues/{index}/assignees
func (c *Client) RemoveIssueAssignees(ctx context.Context, owner, repo string, number int, assignees []string) (*Issue, error) {
	var issue Issue
	err := c.doRequest(ctx, http.MethodDelete, pathf("/repos/%s/%s/issues/%d/assignees", owner, repo, number), map[string]interface{}{"assignees": assignees}, &issue)
	if err != nil {
		return nil, err
	}
	return &issue, nil
}

// RepoBranchInfo represents branch information for listing.
type RepoBranchInfo struct {
	Name   string `json:"name"`
	Commit *struct {
		SHA string `json:"sha"`
		URL string `json:"url"`
	} `json:"commit"`
	Protected bool `json:"protected"`
}

// ListRepoBranches lists all branches of a repository with pagination.
//
// GET /repos/{owner}/{repo}/branches
func (c *Client) ListRepoBranchesPaginated(ctx context.Context, owner, repo string, opts ListOptions) ([]*RepoBranchInfo, error) {
	var branches []*RepoBranchInfo
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/branches?%s", owner, repo, opts.toQuery()), nil, &branches)
	if err != nil {
		return nil, err
	}
	return branches, nil
}

// GetRepoBranch gets a single branch with full details.
//
// GET /repos/{owner}/{repo}/branches/{branch}
func (c *Client) GetRepoBranchDetail(ctx context.Context, owner, repo, branch string) (*RepoBranchInfo, error) {
	var b RepoBranchInfo
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/branches/%s", owner, repo, branch), nil, &b)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

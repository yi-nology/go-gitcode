package gitcode

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type PullRequestState string

const (
	PullRequestStateOpen   PullRequestState = "open"
	PullRequestStateOpened PullRequestState = "opened"
	PullRequestStateClosed PullRequestState = "closed"
	PullRequestStateMerged PullRequestState = "merged"
)

type PullRequest struct {
	ID          int64              `json:"id"`
	Number      int                `json:"number"`
	IID         int                `json:"iid,omitempty"`
	Title       string             `json:"title"`
	Body        string             `json:"body"`
	Description string             `json:"description,omitempty"`
	State       PullRequestState   `json:"state"`
	User        *User              `json:"user"`
	Author      *User              `json:"author"`
	Head        *PullRequestBranch `json:"head"`
	Base        *PullRequestBranch `json:"base"`
	Merged      bool               `json:"merged"`
	Mergeable   *bool              `json:"mergeable"`
	HTMLURL     string             `json:"html_url"`
	DiffURL     string             `json:"diff_url,omitempty"`
	PatchURL    string             `json:"patch_url,omitempty"`
	Draft       bool               `json:"draft,omitempty"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	ClosedAt    NullableTime       `json:"closed_at,omitempty"`
	MergedAt    NullableTime       `json:"merged_at,omitempty"`
}

type PullRequestBranch struct {
	Ref  string `json:"ref"`
	SHA  string `json:"sha"`
	Repo *struct {
		FullName string `json:"full_name"`
		Name     string `json:"name"`
		Path     string `json:"path"`
		HTMLURL  string `json:"html_url"`
	} `json:"repo"`
}

type PullRequestFile struct {
	Filename         string      `json:"filename"`
	PreviousFilename string      `json:"previous_filename,omitempty"`
	Status           string      `json:"status"`
	Additions        int         `json:"additions"`
	Deletions        int         `json:"deletions"`
	Changes          int         `json:"changes"`
	Patch            interface{} `json:"patch,omitempty"`
}

type PullRequestComment struct {
	ID        FlexString `json:"id"`
	Body      string     `json:"body"`
	User      *User      `json:"user"`
	Author    *User      `json:"author"`
	Path      string     `json:"path"`
	Position  int        `json:"position"`
	Line      int        `json:"line"`
	Side      string     `json:"side"`
	CommitID  string     `json:"commit_id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type CreatePullRequestInlineCommentOptions struct {
	Body     string `json:"body"`
	Path     string `json:"path"`
	Line     int    `json:"line"`
	Side     string `json:"side"`
	CommitID string `json:"commit_id,omitempty"`
}

type PullRequestReview struct {
	ID        int64     `json:"id"`
	Body      string    `json:"body"`
	State     string    `json:"state"`
	User      *User     `json:"user"`
	Author    *User     `json:"author"`
	CommitID  string    `json:"commit_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListPullRequestsOptions struct {
	ListOptions
	State     PullRequestState `json:"state,omitempty"`
	Sort      string           `json:"sort,omitempty"`
	Direction string           `json:"direction,omitempty"`
	Head      string           `json:"head,omitempty"`
	Base      string           `json:"base,omitempty"`
}

type CreatePullRequestOptions struct {
	Title string `json:"title"`
	Body  string `json:"body,omitempty"`
	Head  string `json:"head"`
	Base  string `json:"base"`
	Draft bool   `json:"draft,omitempty"`
}

type UpdatePullRequestOptions struct {
	Title      string           `json:"title,omitempty"`
	Body       string           `json:"body,omitempty"`
	State      PullRequestState `json:"state,omitempty"`
	StateEvent string           `json:"state_event,omitempty"`
	Base       string           `json:"base,omitempty"`
}

func (c *Client) ListPullRequests(ctx context.Context, owner, repo string, opts ListPullRequestsOptions) ([]*PullRequest, error) {
	var prs []*PullRequest
	query := opts.toQuery()
	if opts.State != "" {
		query += "&state=" + string(opts.State)
	}
	if opts.Sort != "" {
		query += "&sort=" + opts.Sort
	}
	if opts.Direction != "" {
		query += "&direction=" + opts.Direction
	}
	if opts.Head != "" {
		query += "&head=" + opts.Head
	}
	if opts.Base != "" {
		query += "&base=" + opts.Base
	}
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/pulls?%s", owner, repo, query), nil, &prs)
	if err != nil {
		return nil, err
	}
	return prs, nil
}

func (c *Client) GetPullRequest(ctx context.Context, owner, repo string, number int) (*PullRequest, error) {
	var pr PullRequest
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/pulls/%d", owner, repo, number), nil, &pr)
	if err != nil {
		return nil, err
	}
	return &pr, nil
}

func (c *Client) CreatePullRequest(ctx context.Context, owner, repo string, opts CreatePullRequestOptions) (*PullRequest, error) {
	var pr PullRequest
	err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/pulls", owner, repo), opts, &pr)
	if err != nil {
		return nil, err
	}
	return &pr, nil
}

func (c *Client) UpdatePullRequest(ctx context.Context, owner, repo string, number int, opts UpdatePullRequestOptions) (*PullRequest, error) {
	var pr PullRequest
	err := c.doRequest(ctx, http.MethodPatch, fmt.Sprintf("/repos/%s/%s/pulls/%d", owner, repo, number), opts, &pr)
	if err != nil {
		return nil, err
	}
	return &pr, nil
}

func (c *Client) ClosePullRequest(ctx context.Context, owner, repo string, number int) (*PullRequest, error) {
	pr, err := c.GetPullRequest(ctx, owner, repo, number)
	if err != nil {
		return nil, err
	}
	return c.UpdatePullRequest(ctx, owner, repo, number, UpdatePullRequestOptions{
		Title:      pr.Title,
		StateEvent: "close",
	})
}

func (c *Client) ReopenPullRequest(ctx context.Context, owner, repo string, number int) (*PullRequest, error) {
	pr, err := c.GetPullRequest(ctx, owner, repo, number)
	if err != nil {
		return nil, err
	}
	return c.UpdatePullRequest(ctx, owner, repo, number, UpdatePullRequestOptions{
		Title:      pr.Title,
		StateEvent: "reopen",
	})
}

type MergePullRequestOptions struct {
	CommitMessage string `json:"commit_message,omitempty"`
	Squash        bool   `json:"squash,omitempty"`
}

func (c *Client) MergePullRequest(ctx context.Context, owner, repo string, number int, opts *MergePullRequestOptions) error {
	body := map[string]interface{}{}
	if opts != nil {
		if opts.CommitMessage != "" {
			body["commit_message"] = opts.CommitMessage
		}
		if opts.Squash {
			body["merge_method"] = "squash"
		}
	}
	return c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/repos/%s/%s/pulls/%d/merge", owner, repo, number), body, nil)
}

func (c *Client) IsPullRequestMerged(ctx context.Context, owner, repo string, number int) (bool, error) {
	_, err := c.doRawRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/pulls/%d/merge", owner, repo, number))
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *Client) ListPullRequestFiles(ctx context.Context, owner, repo string, number int) ([]*PullRequestFile, error) {
	var files []*PullRequestFile
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/pulls/%d/files", owner, repo, number), nil, &files)
	if err != nil {
		return nil, err
	}
	return files, nil
}

func (c *Client) ListPullRequestComments(ctx context.Context, owner, repo string, number int) ([]*PullRequestComment, error) {
	var comments []*PullRequestComment
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/pulls/%d/comments", owner, repo, number), nil, &comments)
	if err != nil {
		return nil, err
	}
	return comments, nil
}

func (c *Client) CreatePullRequestComment(ctx context.Context, owner, repo string, number int, body, path string, position, commitID string) (*PullRequestComment, error) {
	var comment PullRequestComment
	err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/pulls/%d/comments", owner, repo, number), map[string]interface{}{
		"body":      body,
		"path":      path,
		"position":  position,
		"commit_id": commitID,
	}, &comment)
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (c *Client) CreatePullRequestInlineComment(ctx context.Context, owner, repo string, number int, opts CreatePullRequestInlineCommentOptions) (*PullRequestComment, error) {
	var comment PullRequestComment
	err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/pulls/%d/comments", owner, repo, number), opts, &comment)
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (c *Client) UpdatePullRequestComment(ctx context.Context, owner, repo string, commentID string, body string) (*PullRequestComment, error) {
	var comment PullRequestComment
	err := c.doRequest(ctx, http.MethodPatch, fmt.Sprintf("/repos/%s/%s/pulls/comments/%s", owner, repo, commentID), map[string]string{"body": body}, &comment)
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (c *Client) DeletePullRequestComment(ctx context.Context, owner, repo string, commentID string) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/repos/%s/%s/pulls/comments/%s", owner, repo, commentID), nil, nil)
}

func (c *Client) ListPullRequestReviews(ctx context.Context, owner, repo string, number int) ([]*PullRequestReview, error) {
	var reviews []*PullRequestReview
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/pulls/%d/reviews", owner, repo, number), nil, &reviews)
	if err != nil {
		return nil, err
	}
	return reviews, nil
}

func (c *Client) CreatePullRequestReview(ctx context.Context, owner, repo string, number int, body, event string) (*PullRequestReview, error) {
	var review PullRequestReview
	err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/pulls/%d/reviews", owner, repo, number), map[string]interface{}{
		"body":  body,
		"event": event,
	}, &review)
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (c *Client) ListPullRequestCommits(ctx context.Context, owner, repo string, number int) ([]*Commit, error) {
	var commits []*Commit
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/pulls/%d/commits", owner, repo, number), nil, &commits)
	if err != nil {
		return nil, err
	}
	return commits, nil
}

func (c *Client) ListPullRequestLabels(ctx context.Context, owner, repo string, number int) ([]*Label, error) {
	var labels []*Label
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/pulls/%d/labels", owner, repo, number), nil, &labels)
	if err != nil {
		return nil, err
	}
	return labels, nil
}

func (c *Client) AddPullRequestLabels(ctx context.Context, owner, repo string, number int, labels []string) error {
	return c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/pulls/%d/labels", owner, repo, number), labels, nil)
}

func (c *Client) RemovePullRequestLabel(ctx context.Context, owner, repo string, number int, name string) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/repos/%s/%s/pulls/%d/labels/%s", owner, repo, number, name), nil, nil)
}

func (c *Client) ReplacePullRequestLabels(ctx context.Context, owner, repo string, number int, labels []string) error {
	return c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/repos/%s/%s/pulls/%d/labels", owner, repo, number), labels, nil)
}

func (c *Client) AssignPullRequestReviewers(ctx context.Context, owner, repo string, number int, assignees string) error {
	return c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/pulls/%d/assignees", owner, repo, number), map[string]string{"assignees": assignees}, nil)
}

func (c *Client) UnassignPullRequestReviewers(ctx context.Context, owner, repo string, number int, assignees string) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/repos/%s/%s/pulls/%d/assignees?assignees=%s", owner, repo, number, assignees), nil, nil)
}

func (c *Client) AssignPullRequestTesters(ctx context.Context, owner, repo string, number int, testers string) error {
	return c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/pulls/%d/testers", owner, repo, number), map[string]string{"testers": testers}, nil)
}

func (c *Client) HandlePullRequestTest(ctx context.Context, owner, repo string, number int, force bool) error {
	body := map[string]interface{}{}
	if force {
		body["force"] = true
	}
	return c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/pulls/%d/test", owner, repo, number), body, nil)
}

func (c *Client) HandlePullRequestReview(ctx context.Context, owner, repo string, number int, force bool) error {
	body := map[string]interface{}{}
	if force {
		body["force"] = true
	}
	return c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/pulls/%d/review", owner, repo, number), body, nil)
}

func (c *Client) ResetPullRequestTestStatus(ctx context.Context, owner, repo string, number int, resetAll bool) error {
	body := map[string]interface{}{}
	if resetAll {
		body["reset_all"] = true
	}
	return c.doRequest(ctx, http.MethodPatch, fmt.Sprintf("/repos/%s/%s/pulls/%d/testers", owner, repo, number), body, nil)
}

func (c *Client) ResetPullRequestReviewStatus(ctx context.Context, owner, repo string, number int, resetAll bool) error {
	body := map[string]interface{}{}
	if resetAll {
		body["reset_all"] = true
	}
	return c.doRequest(ctx, http.MethodPatch, fmt.Sprintf("/repos/%s/%s/pulls/%d/assignees", owner, repo, number), body, nil)
}

type PROperateLog struct {
	ID        int64  `json:"id"`
	Action    string `json:"action"`
	CreatedAt string `json:"created_at"`
	User      *User  `json:"user"`
}

func (c *Client) GetPullRequestOperateLogs(ctx context.Context, owner, repo string, number int) ([]*PROperateLog, error) {
	var logs []*PROperateLog
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/pulls/%d/operate_logs", owner, repo, number), nil, &logs)
	if err != nil {
		return nil, err
	}
	return logs, nil
}

func (c *Client) GetPullRequestLinkedIssues(ctx context.Context, owner, repo string, number int, opts ListOptions) ([]*Issue, error) {
	var issues []*Issue
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/pulls/%d/issues?%s", owner, repo, number, opts.toQuery()), nil, &issues)
	if err != nil {
		return nil, err
	}
	return issues, nil
}

func (c *Client) GetPullRequestComment(ctx context.Context, owner, repo string, commentID int64) (*PullRequestComment, error) {
	var comment PullRequestComment
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/pulls/comments/%d", owner, repo, commentID), nil, &comment)
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

type ListEnterprisePRsOptions struct {
	ListOptions
	State       string `json:"state,omitempty"`
	IssueNumber int    `json:"issue_number,omitempty"`
	Sort        string `json:"sort,omitempty"`
	Direction   string `json:"direction,omitempty"`
}

func (c *Client) ListEnterprisePullRequests(ctx context.Context, enterprise string, opts ListEnterprisePRsOptions) ([]*PullRequest, error) {
	var prs []*PullRequest
	query := opts.toQuery()
	if opts.State != "" {
		query += "&state=" + opts.State
	}
	if opts.IssueNumber > 0 {
		query += fmt.Sprintf("&issue_number=%d", opts.IssueNumber)
	}
	if opts.Sort != "" {
		query += "&sort=" + opts.Sort
	}
	if opts.Direction != "" {
		query += "&direction=" + opts.Direction
	}
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/enterprises/%s/pull-requests?%s", enterprise, query), nil, &prs)
	if err != nil {
		return nil, err
	}
	return prs, nil
}

func (c *Client) ListOrgPullRequests(ctx context.Context, org string, opts ListEnterprisePRsOptions) ([]*PullRequest, error) {
	var prs []*PullRequest
	query := opts.toQuery()
	if opts.State != "" {
		query += "&state=" + opts.State
	}
	if opts.IssueNumber > 0 {
		query += fmt.Sprintf("&issue_number=%d", opts.IssueNumber)
	}
	if opts.Sort != "" {
		query += "&sort=" + opts.Sort
	}
	if opts.Direction != "" {
		query += "&direction=" + opts.Direction
	}
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/org/%s/pull-requests?%s", org, query), nil, &prs)
	if err != nil {
		return nil, err
	}
	return prs, nil
}

func (c *Client) GetEnterpriseIssueLinkedPRs(ctx context.Context, enterprise string, number int) ([]*PullRequest, error) {
	var prs []*PullRequest
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/enterprises/%s/issues/%d/pull-requests", enterprise, number), nil, &prs)
	if err != nil {
		return nil, err
	}
	return prs, nil
}

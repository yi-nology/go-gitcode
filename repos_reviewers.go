package gitcode

import (
	"context"
	"net/http"
)

// Reviewer represents a potential reviewer for a pull request.
type Reviewer struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

// ListRepoReviewers lists all available reviewers for a repository.
//
// GET /repos/{owner}/{repo}/reviewers
func (c *Client) ListRepoReviewers(ctx context.Context, owner, repo string, opts ListOptions) ([]*Reviewer, error) {
	var reviewers []*Reviewer
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/reviewers?%s", owner, repo, opts.toQuery()), nil, &reviewers)
	if err != nil {
		return nil, err
	}
	return reviewers, nil
}

// PullRequestReviewRequest represents a review request.
type PullRequestReviewRequest struct {
	Reviewers     []string `json:"reviewers,omitempty"`
	TeamReviewers []string `json:"team_reviewers,omitempty"`
}

// ListPullRequestReviewers lists all reviewers requested for a pull request.
//
// GET /repos/{owner}/{repo}/pulls/{index}/requested_reviewers
func (c *Client) ListPullRequestReviewers(ctx context.Context, owner, repo string, number int) ([]*Reviewer, error) {
	var reviewers []*Reviewer
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/pulls/%d/requested_reviewers", owner, repo, number), nil, &reviewers)
	if err != nil {
		return nil, err
	}
	return reviewers, nil
}

// RequestPullRequestReviewers requests reviewers for a pull request.
//
// POST /repos/{owner}/{repo}/pulls/{index}/requested_reviewers
func (c *Client) RequestPullRequestReviewers(ctx context.Context, owner, repo string, number int, opts PullRequestReviewRequest) error {
	return c.doRequest(ctx, http.MethodPost, pathf("/repos/%s/%s/pulls/%d/requested_reviewers", owner, repo, number), opts, nil)
}

// RemovePullRequestReviewer removes a requested reviewer from a pull request.
//
// DELETE /repos/{owner}/{repo}/pulls/{index}/requested_reviewers
func (c *Client) RemovePullRequestReviewer(ctx context.Context, owner, repo string, number int, opts PullRequestReviewRequest) error {
	return c.doRequest(ctx, http.MethodDelete, pathf("/repos/%s/%s/pulls/%d/requested_reviewers", owner, repo, number), opts, nil)
}

// DismissPullRequestReview dismisses a pull request review.
//
// PUT /repos/{owner}/{repo}/pulls/{index}/reviews/{id}/dismissals
func (c *Client) DismissPullRequestReview(ctx context.Context, owner, repo string, number int, reviewID int64, message string) error {
	return c.doRequest(ctx, http.MethodPut, pathf("/repos/%s/%s/pulls/%d/reviews/%d/dismissals", owner, repo, number, reviewID), map[string]string{"message": message}, nil)
}

// SubmitPullRequestReview submits a pull request review.
//
// POST /repos/{owner}/{repo}/pulls/{index}/reviews/{id}
func (c *Client) SubmitPullRequestReview(ctx context.Context, owner, repo string, number int, reviewID int64, body, event string) (*PullRequestReview, error) {
	var review PullRequestReview
	err := c.doRequest(ctx, http.MethodPost, pathf("/repos/%s/%s/pulls/%d/reviews/%d", owner, repo, number, reviewID), map[string]string{
		"body":  body,
		"event": event,
	}, &review)
	if err != nil {
		return nil, err
	}
	return &review, nil
}

// GetPullRequestReview gets a single pull request review.
//
// GET /repos/{owner}/{repo}/pulls/{index}/reviews/{id}
func (c *Client) GetPullRequestReview(ctx context.Context, owner, repo string, number int, reviewID int64) (*PullRequestReview, error) {
	var review PullRequestReview
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/pulls/%d/reviews/%d", owner, repo, number, reviewID), nil, &review)
	if err != nil {
		return nil, err
	}
	return &review, nil
}

// PullRequestDiff represents the diff of a pull request.
type PullRequestDiff struct {
	Content string `json:"content"`
}

// GetPullRequestDiff returns the diff of a pull request.
//
// GET /repos/{owner}/{repo}/pulls/{index}.diff
func (c *Client) GetPullRequestDiff(ctx context.Context, owner, repo string, number int) ([]byte, error) {
	return c.doRawRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/pulls/%d.diff", owner, repo, number))
}

// GetPullRequestPatch returns the patch of a pull request.
//
// GET /repos/{owner}/{repo}/pulls/{index}.patch
func (c *Client) GetPullRequestPatch(ctx context.Context, owner, repo string, number int) ([]byte, error) {
	return c.doRawRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/pulls/%d.patch", owner, repo, number))
}

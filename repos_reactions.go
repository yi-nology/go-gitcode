package gitcode

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Reaction represents an emoji reaction.
type Reaction struct {
	ID        int64     `json:"id"`
	User      *User     `json:"user"`
	Content   string    `json:"content"` // +1, -1, laugh, confused, heart, hooray, rocket, eyes
	CreatedAt time.Time `json:"created_at"`
}

// ReactionContent constants for reaction types.
const (
	ReactionPlusOne  = "+1"
	ReactionMinusOne = "-1"
	ReactionLaugh    = "laugh"
	ReactionConfused = "confused"
	ReactionHeart    = "heart"
	ReactionHooray   = "hooray"
	ReactionRocket   = "rocket"
	ReactionEyes     = "eyes"
)

// CreateReactionOptions specifies options for creating a reaction.
type CreateReactionOptions struct {
	Content string `json:"content"`
}

// --- Issue Reactions ---

// ListIssueReactions lists all reactions for an issue.
//
// GET /repos/{owner}/{repo}/issues/{index}/reactions
func (c *Client) ListIssueReactions(ctx context.Context, owner, repo string, number int, opts ListOptions) ([]*Reaction, error) {
	var reactions []*Reaction
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/issues/%d/reactions?%s", owner, repo, number, opts.toQuery()), nil, &reactions)
	if err != nil {
		return nil, err
	}
	return reactions, nil
}

// CreateIssueReaction adds a reaction to an issue.
//
// POST /repos/{owner}/{repo}/issues/{index}/reactions
func (c *Client) CreateIssueReaction(ctx context.Context, owner, repo string, number int, content string) (*Reaction, error) {
	var reaction Reaction
	err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/issues/%d/reactions", owner, repo, number), CreateReactionOptions{Content: content}, &reaction)
	if err != nil {
		return nil, err
	}
	return &reaction, nil
}

// DeleteIssueReaction removes a reaction from an issue.
//
// DELETE /repos/{owner}/{repo}/issues/{index}/reactions/{id}
func (c *Client) DeleteIssueReaction(ctx context.Context, owner, repo string, number int, reactionID int64) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/repos/%s/%s/issues/%d/reactions/%d", owner, repo, number, reactionID), nil, nil)
}

// --- Issue Comment Reactions ---

// ListIssueCommentReactions lists all reactions for an issue comment.
//
// GET /repos/{owner}/{repo}/issues/comments/{id}/reactions
func (c *Client) ListIssueCommentReactions(ctx context.Context, owner, repo string, commentID int64, opts ListOptions) ([]*Reaction, error) {
	var reactions []*Reaction
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/issues/comments/%d/reactions?%s", owner, repo, commentID, opts.toQuery()), nil, &reactions)
	if err != nil {
		return nil, err
	}
	return reactions, nil
}

// CreateIssueCommentReaction adds a reaction to an issue comment.
//
// POST /repos/{owner}/{repo}/issues/comments/{id}/reactions
func (c *Client) CreateIssueCommentReaction(ctx context.Context, owner, repo string, commentID int64, content string) (*Reaction, error) {
	var reaction Reaction
	err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/issues/comments/%d/reactions", owner, repo, commentID), CreateReactionOptions{Content: content}, &reaction)
	if err != nil {
		return nil, err
	}
	return &reaction, nil
}

// DeleteIssueCommentReaction removes a reaction from an issue comment.
//
// DELETE /repos/{owner}/{repo}/issues/comments/{id}/reactions/{id}
func (c *Client) DeleteIssueCommentReaction(ctx context.Context, owner, repo string, commentID, reactionID int64) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/repos/%s/%s/issues/comments/%d/reactions/%d", owner, repo, commentID, reactionID), nil, nil)
}

// --- Pull Request Comment Reactions ---

// ListPullRequestCommentReactions lists all reactions for a pull request comment.
//
// GET /repos/{owner}/{repo}/pulls/comments/{id}/reactions
func (c *Client) ListPullRequestCommentReactions(ctx context.Context, owner, repo string, commentID int64, opts ListOptions) ([]*Reaction, error) {
	var reactions []*Reaction
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/pulls/comments/%d/reactions?%s", owner, repo, commentID, opts.toQuery()), nil, &reactions)
	if err != nil {
		return nil, err
	}
	return reactions, nil
}

// CreatePullRequestCommentReaction adds a reaction to a pull request comment.
//
// POST /repos/{owner}/{repo}/pulls/comments/{id}/reactions
func (c *Client) CreatePullRequestCommentReaction(ctx context.Context, owner, repo string, commentID int64, content string) (*Reaction, error) {
	var reaction Reaction
	err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/pulls/comments/%d/reactions", owner, repo, commentID), CreateReactionOptions{Content: content}, &reaction)
	if err != nil {
		return nil, err
	}
	return &reaction, nil
}

// DeletePullRequestCommentReaction removes a reaction from a pull request comment.
//
// DELETE /repos/{owner}/{repo}/pulls/comments/{id}/reactions/{id}
func (c *Client) DeletePullRequestCommentReaction(ctx context.Context, owner, repo string, commentID, reactionID int64) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/repos/%s/%s/pulls/comments/%d/reactions/%d", owner, repo, commentID, reactionID), nil, nil)
}

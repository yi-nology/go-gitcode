package gitcode

import (
	"context"
	"net/http"
	"time"
)

// Discussion represents a repository discussion.
type Discussion struct {
	ID        int64     `json:"id"`
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	State     string    `json:"state"`
	User      *User     `json:"user"`
	Author    *User     `json:"author"`
	Labels    []*Label  `json:"labels"`
	HTMLURL   string    `json:"html_url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DiscussionComment represents a comment on a discussion.
type DiscussionComment struct {
	ID        int64     `json:"id"`
	Body      string    `json:"body"`
	User      *User     `json:"user"`
	Author    *User     `json:"author"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DiscussionCommentReply represents a reply to a discussion comment.
type DiscussionCommentReply struct {
	ID        int64     `json:"id"`
	Body      string    `json:"body"`
	User      *User     `json:"user"`
	Author    *User     `json:"author"`
	ReplyToID int64     `json:"reply_to_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ListDiscussions lists all discussions for a repository.
//
// GET /repos/{owner}/{repo}/discuss
func (c *Client) ListDiscussions(ctx context.Context, owner, repo string, opts ListOptions) ([]*Discussion, error) {
	var discussions []*Discussion
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/discuss?%s", owner, repo, opts.toQuery()), nil, &discussions)
	if err != nil {
		return nil, err
	}
	return discussions, nil
}

// GetDiscussion gets a single discussion by number.
//
// GET /repos/{owner}/{repo}/discuss/{number}
func (c *Client) GetDiscussion(ctx context.Context, owner, repo string, number int) (*Discussion, error) {
	var discussion Discussion
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/discuss/%d", owner, repo, number), nil, &discussion)
	if err != nil {
		return nil, err
	}
	return &discussion, nil
}

// ListDiscussionComments lists all comments for a discussion.
//
// GET /repos/{owner}/{repo}/discuss/{number}/comment
func (c *Client) ListDiscussionComments(ctx context.Context, owner, repo string, number int, opts ListOptions) ([]*DiscussionComment, error) {
	var comments []*DiscussionComment
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/discuss/%d/comment?%s", owner, repo, number, opts.toQuery()), nil, &comments)
	if err != nil {
		return nil, err
	}
	return comments, nil
}

// ListDiscussionCommentReplies lists all replies for a discussion comment.
//
// GET /repos/{owner}/{repo}/discuss/{number}/comment/{comment_id}/reply
func (c *Client) ListDiscussionCommentReplies(ctx context.Context, owner, repo string, number int, commentID int64, opts ListOptions) ([]*DiscussionCommentReply, error) {
	var replies []*DiscussionCommentReply
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/discuss/%d/comment/%d/reply?%s", owner, repo, number, commentID, opts.toQuery()), nil, &replies)
	if err != nil {
		return nil, err
	}
	return replies, nil
}

// --- Fork Sync ---

// ForkSyncStatus represents the sync status of a forked repository.
type ForkSyncStatus struct {
	Synced    bool   `json:"synced"`
	BehindBy  int    `json:"behind_by"`
	AheadBy   int    `json:"ahead_by"`
	MergeBase string `json:"merge_base,omitempty"`
}

// GetForkSyncStatus checks the sync status of a forked repository.
//
// GET /repos/{owner}/{repo}/sync-repo
func (c *Client) GetForkSyncStatus(ctx context.Context, owner, repo string) (*ForkSyncStatus, error) {
	var status ForkSyncStatus
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/sync-repo", owner, repo), nil, &status)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

// SyncForkRepository syncs a forked repository with its upstream.
//
// PUT /repos/{owner}/{repo}/sync-repo
func (c *Client) SyncForkRepository(ctx context.Context, owner, repo string) error {
	return c.doRequest(ctx, http.MethodPut, pathf("/repos/%s/%s/sync-repo", owner, repo), nil, nil)
}

// --- Remote Mirrors ---

// RemoteMirror represents a remote mirror configuration.
type RemoteMirror struct {
	ID            int64  `json:"id"`
	RemoteURL     string `json:"remote_url"`
	Enabled       bool   `json:"enabled"`
	OnlyProtected bool   `json:"only_protected"`
	LastError     string `json:"last_error,omitempty"`
	LastUpdateAt  string `json:"last_update_at,omitempty"`
}

// GetRepoRemoteMirror gets the remote mirror configuration of a repository.
//
// GET /repos/{owner}/{repo}/repo-remote-mirror
func (c *Client) GetRepoRemoteMirror(ctx context.Context, owner, repo string) (*RemoteMirror, error) {
	var mirror RemoteMirror
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/repo-remote-mirror", owner, repo), nil, &mirror)
	if err != nil {
		return nil, err
	}
	return &mirror, nil
}

// ListPushRemoteMirrors lists all push remote mirrors of a repository.
//
// GET /repos/{owner}/{repo}/push-remote-mirrors
func (c *Client) ListPushRemoteMirrors(ctx context.Context, owner, repo string, opts ListOptions) ([]*RemoteMirror, error) {
	var mirrors []*RemoteMirror
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/push-remote-mirrors?%s", owner, repo, opts.toQuery()), nil, &mirrors)
	if err != nil {
		return nil, err
	}
	return mirrors, nil
}

// --- License & CLA ---

// RepoLicense represents a repository license.
type RepoLicense struct {
	Key     string `json:"key"`
	Name    string `json:"name"`
	URL     string `json:"url,omitempty"`
	HTMLURL string `json:"html_url,omitempty"`
	Body    string `json:"body,omitempty"`
}

// GetRepoLicense gets the license of a repository.
//
// GET /repos/{owner}/{repo}/license
func (c *Client) GetRepoLicense(ctx context.Context, owner, repo string) (*RepoLicense, error) {
	var license RepoLicense
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/license", owner, repo), nil, &license)
	if err != nil {
		return nil, err
	}
	return &license, nil
}

// RepoCLA represents a repository CLA (Contributor License Agreement).
type RepoCLA struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Content string `json:"content,omitempty"`
	Enabled bool   `json:"enabled"`
}

// ListRepoCLAs lists all CLAs for a repository.
//
// GET /repos/{owner}/{repo}/clas
func (c *Client) ListRepoCLAs(ctx context.Context, owner, repo string) ([]*RepoCLA, error) {
	var claList []*RepoCLA                                                                          //nolint:misspell // "clas" is the API endpoint name
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/clas", owner, repo), nil, &claList) //nolint:misspell // API endpoint
	if err != nil {
		return nil, err
	}
	return claList, nil
}

// ConfigureRepoCLA configures the CLA for a repository.
//
// PUT /repos/{owner}/{repo}/clas
func (c *Client) ConfigureRepoCLA(ctx context.Context, owner, repo string, opts *RepoCLA) (*RepoCLA, error) {
	var cla RepoCLA
	err := c.doRequest(ctx, http.MethodPut, pathf("/repos/%s/%s/clas", owner, repo), opts, &cla) //nolint:misspell // API endpoint
	if err != nil {
		return nil, err
	}
	return &cla, nil
}

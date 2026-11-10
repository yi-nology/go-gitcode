package gitcode

import (
	"context"
	"net/http"
	"time"
)

// --- Organization Customized Roles ---

// OrgCustomizedRole represents a customized role in an organization.
type OrgCustomizedRole struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Permission  string `json:"permission,omitempty"`
}

// ListOrgCustomizedRoles lists all customized roles for an organization.
//
// GET /org/{org}/customized-roles
func (c *Client) ListOrgCustomizedRoles(ctx context.Context, org string) ([]*OrgCustomizedRole, error) {
	var roles []*OrgCustomizedRole
	err := c.doRequest(ctx, http.MethodGet, pathf("/org/%s/customized-roles", org), nil, &roles)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// --- Organization Discussions ---

// OrgDiscussion represents a discussion in an organization.
type OrgDiscussion struct {
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

// OrgDiscussionComment represents a comment on an organization discussion.
type OrgDiscussionComment struct {
	ID        int64     `json:"id"`
	Body      string    `json:"body"`
	User      *User     `json:"user"`
	Author    *User     `json:"author"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// OrgDiscussionCommentReply represents a reply to an organization discussion comment.
type OrgDiscussionCommentReply struct {
	ID        int64     `json:"id"`
	Body      string    `json:"body"`
	User      *User     `json:"user"`
	Author    *User     `json:"author"`
	ReplyToID int64     `json:"reply_to_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ListOrgDiscussions lists all discussions for an organization.
//
// GET /orgs/{org}/discuss
func (c *Client) ListOrgDiscussions(ctx context.Context, org string, opts ListOptions) ([]*OrgDiscussion, error) {
	var discussions []*OrgDiscussion
	err := c.doRequest(ctx, http.MethodGet, pathf("/orgs/%s/discuss?%s", org, opts.toQuery()), nil, &discussions)
	if err != nil {
		return nil, err
	}
	return discussions, nil
}

// GetOrgDiscussion gets a single organization discussion by number.
//
// GET /orgs/{org}/discuss/{number}
func (c *Client) GetOrgDiscussion(ctx context.Context, org string, number int) (*OrgDiscussion, error) {
	var discussion OrgDiscussion
	err := c.doRequest(ctx, http.MethodGet, pathf("/orgs/%s/discuss/%d", org, number), nil, &discussion)
	if err != nil {
		return nil, err
	}
	return &discussion, nil
}

// ListOrgDiscussionComments lists all comments for an organization discussion.
//
// GET /orgs/{org}/discuss/{number}/comment
func (c *Client) ListOrgDiscussionComments(ctx context.Context, org string, number int, opts ListOptions) ([]*OrgDiscussionComment, error) {
	var comments []*OrgDiscussionComment
	err := c.doRequest(ctx, http.MethodGet, pathf("/orgs/%s/discuss/%d/comment?%s", org, number, opts.toQuery()), nil, &comments)
	if err != nil {
		return nil, err
	}
	return comments, nil
}

// ListOrgDiscussionCommentReplies lists all replies for an organization discussion comment.
//
// GET /orgs/{org}/discuss/{number}/comment/{comment_id}/reply
func (c *Client) ListOrgDiscussionCommentReplies(ctx context.Context, org string, number int, commentID int64, opts ListOptions) ([]*OrgDiscussionCommentReply, error) {
	var replies []*OrgDiscussionCommentReply
	err := c.doRequest(ctx, http.MethodGet, pathf("/orgs/%s/discuss/%d/comment/%d/reply?%s", org, number, commentID, opts.toQuery()), nil, &replies)
	if err != nil {
		return nil, err
	}
	return replies, nil
}

package gitcode

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// RepoInvitation represents a repository invitation.
type RepoInvitation struct {
	ID         int64       `json:"id"`
	Repo       *Repository `json:"repo"`
	Invitee    *User       `json:"invitee"`
	Inviter    *User       `json:"inviter"`
	Permission string      `json:"permission"`
	CreatedAt  time.Time   `json:"created_at"`
	URL        string      `json:"url"`
	HTMLURL    string      `json:"html_url,omitempty"`
}

// AcceptRepoInvitation accepts a repository invitation.
//
// PATCH /user/repository_invitations/{id}
func (c *Client) AcceptRepoInvitation(ctx context.Context, invitationID int64) error {
	return c.doRequest(ctx, http.MethodPatch, fmt.Sprintf("/user/repository_invitations/%d", invitationID), nil, nil)
}

// DeclineRepoInvitation declines a repository invitation.
//
// DELETE /user/repository_invitations/{id}
func (c *Client) DeclineRepoInvitation(ctx context.Context, invitationID int64) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/user/repository_invitations/%d", invitationID), nil, nil)
}

// ListPendingRepoInvitations lists pending repository invitations for the authenticated user.
//
// GET /user/repository_invitations
func (c *Client) ListPendingRepoInvitations(ctx context.Context, opts ListOptions) ([]*RepoInvitation, error) {
	var invitations []*RepoInvitation
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/user/repository_invitations?%s", opts.toQuery()), nil, &invitations)
	if err != nil {
		return nil, err
	}
	return invitations, nil
}

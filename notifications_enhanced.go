package gitcode

import (
	"context"
	"net/http"
	"time"
)

// NotificationThread represents a detailed notification thread.
type NotificationThread struct {
	ID         int64                `json:"id"`
	Unread     bool                 `json:"unread"`
	Pinned     bool                 `json:"pinned,omitempty"`
	UpdatedAt  time.Time            `json:"updated_at"`
	Subject    *NotificationSubject `json:"subject"`
	Repository *Repository          `json:"repository"`
	URL        string               `json:"url"`
}

// NotificationSubject represents the subject of a notification.
type NotificationSubject struct {
	Title            string `json:"title"`
	URL              string `json:"url"`
	LatestCommentURL string `json:"latest_comment_url"`
	Type             string `json:"type"` // Issue, PullRequest, Commit, Repository
}

// ListNotificationsOptions specifies options for listing notifications.
type ListNotificationsOptions struct {
	ListOptions
	All    bool   `json:"all,omitempty"`    // Include read notifications
	Since  string `json:"since,omitempty"`  // Only notifications updated after this time
	Before string `json:"before,omitempty"` // Only notifications updated before this time
	Status string `json:"status,omitempty"` // unread, read, pinned
}

// MarkNotificationsOptions specifies options for marking notifications.
type MarkNotificationsOptions struct {
	LastReadAt string `json:"last_read_at,omitempty"` // ISO 8601 timestamp
	All        bool   `json:"all,omitempty"`
}

// ListNotificationsWithOptions lists the current user's notifications with options.
//
// GET /notifications
func (c *Client) ListNotificationsWithOptions(ctx context.Context, opts ListNotificationsOptions) ([]*NotificationThread, error) {
	var notifications []*NotificationThread
	query := opts.toQuery()
	if opts.All {
		query += "&all=true"
	}
	if opts.Since != "" {
		query += "&since=" + opts.Since
	}
	if opts.Before != "" {
		query += "&before=" + opts.Before
	}
	if opts.Status != "" {
		query += "&status=" + opts.Status
	}
	err := c.doRequest(ctx, http.MethodGet, pathf("/notifications?%s", query), nil, &notifications)
	if err != nil {
		return nil, err
	}
	return notifications, nil
}

// GetNotificationThread gets a single notification thread.
//
// GET /notifications/threads/{id}
func (c *Client) GetNotificationThread(ctx context.Context, threadID int64) (*NotificationThread, error) {
	var thread NotificationThread
	err := c.doRequest(ctx, http.MethodGet, pathf("/notifications/threads/%d", threadID), nil, &thread)
	if err != nil {
		return nil, err
	}
	return &thread, nil
}

// MarkNotificationThreadAsRead marks a single notification thread as read.
//
// PATCH /notifications/threads/{id}
func (c *Client) MarkNotificationThreadAsRead(ctx context.Context, threadID int64) error {
	return c.doRequest(ctx, http.MethodPatch, pathf("/notifications/threads/%d", threadID), nil, nil)
}

// MarkNotificationsAsRead marks all notifications as read.
//
// PUT /notifications
func (c *Client) MarkNotificationsAsRead(ctx context.Context, opts MarkNotificationsOptions) error {
	return c.doRequest(ctx, http.MethodPut, "/notifications", opts, nil)
}

// ListRepoNotifications lists notifications for a repository.
//
// GET /repos/{owner}/{repo}/notifications
func (c *Client) ListRepoNotifications(ctx context.Context, owner, repo string, opts ListNotificationsOptions) ([]*NotificationThread, error) {
	var notifications []*NotificationThread
	query := opts.toQuery()
	if opts.All {
		query += "&all=true"
	}
	if opts.Since != "" {
		query += "&since=" + opts.Since
	}
	if opts.Before != "" {
		query += "&before=" + opts.Before
	}
	if opts.Status != "" {
		query += "&status=" + opts.Status
	}
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/notifications?%s", owner, repo, query), nil, &notifications)
	if err != nil {
		return nil, err
	}
	return notifications, nil
}

// MarkRepoNotificationsAsRead marks all notifications for a repository as read.
//
// PUT /repos/{owner}/{repo}/notifications
func (c *Client) MarkRepoNotificationsAsRead(ctx context.Context, owner, repo string, opts MarkNotificationsOptions) error {
	return c.doRequest(ctx, http.MethodPut, pathf("/repos/%s/%s/notifications", owner, repo), opts, nil)
}

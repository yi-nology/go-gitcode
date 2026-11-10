package gitcode

// Dashboard kanban client, speaking the live-verified GitCode wire
// (probe round 2026-11-04 against the yi-nology org: every route below hit
// GitCode business logic — KANBAN_NON_EXISTENT for a dummy board id —
// while the previously implemented /orgs/{org}/kanban/boards family
// returned framework 404s; docs slugs all use the singular /org/ prefix).
// Wire quirks: board ids are strings on the wire, list is a counted
// envelope, remove_item takes an array body, and the old
// UpdateKanbanBoardItem route has no live equivalent (the repo-scoped
// /repos/{o}/{r}/issues/{n}/kanban-values route — covered by
// UpdateIssueKanbanValues — is the documented item-value writer).

import (
	"context"
	"fmt"
	"net/http"
)

// KanbanBoard is one org dashboard board. ID is a string on the wire;
// Iid is the numeric sequence.
type KanbanBoard struct {
	ID          string `json:"id"`
	Iid         int64  `json:"iid"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      int    `json:"status"`     // 0 open / 1 closed
	Visibility  int    `json:"visibility"` // 1 public / 2 private
	UpdatedAt   string `json:"updated_at"`
}

// KanbanBoardList is the counted envelope of the board list.
type KanbanBoardList struct {
	CloseCount int            `json:"close_count"`
	OpenCount  int            `json:"open_count"`
	AllCount   int            `json:"all_count"`
	Content    []*KanbanBoard `json:"content"`
}

// ListKanbanBoardsOptions carries the documented list filters.
type ListKanbanBoardsOptions struct {
	ListOptions
	Status     int    `json:"status,omitempty"`     // 0 open / 1 closed / unset = all
	Sort       string `json:"sort,omitempty"`       // newest / oldest
	Visibility int    `json:"visibility,omitempty"` // 1 public / 2 private
	Search     string `json:"search,omitempty"`
}

// ListKanbanBoards lists an organization's dashboard boards.
//
// GET /v5/org/{org}/kanban/list (live-verified).
func (c *Client) ListKanbanBoards(ctx context.Context, org string, opts ListKanbanBoardsOptions) (*KanbanBoardList, error) {
	q := urlValuesFromListOptions(opts.ListOptions)
	if opts.Status != 0 {
		q.Set("status", fmt.Sprint(opts.Status))
	}
	if opts.Sort != "" {
		q.Set("sort", opts.Sort)
	}
	if opts.Visibility != 0 {
		q.Set("visibility", fmt.Sprint(opts.Visibility))
	}
	if opts.Search != "" {
		q.Set("search", opts.Search)
	}
	var list KanbanBoardList
	err := c.doRequest(ctx, http.MethodGet, pathf("/org/%s/kanban/list?%s", org, q.Encode()), nil, &list)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

// GetKanbanBoard fetches one board's detail.
//
// GET /v5/org/{org}/kanban/{kanban_id}/detail (live-verified).
func (c *Client) GetKanbanBoard(ctx context.Context, org, kanbanID string) (*KanbanBoard, error) {
	var board KanbanBoard
	err := c.doRequest(ctx, http.MethodGet, pathf("/org/%s/kanban/%s/detail", org, kanbanID), nil, &board)
	if err != nil {
		return nil, err
	}
	return &board, nil
}

// KanbanBoardItem is one item on a board.
type KanbanBoardItem struct {
	ID         int64  `json:"id"`
	Number     int64  `json:"number"`
	Title      string `json:"title"`
	SourceType string `json:"source_type"` // issue / merge_request
	Status     string `json:"status"`
	HTMLURL    string `json:"html_url"`
	URL        string `json:"url"`
	Repo       *struct {
		ID        int64  `json:"id"`
		FullName  string `json:"full_name"`
		HumanName string `json:"human_name"`
		URL       string `json:"url"`
		Name      string `json:"name"`
	} `json:"repo"`
	Labels []struct {
		ID    int64  `json:"id"`
		Name  string `json:"name"`
		Color string `json:"color"`
	} `json:"labels"`
	Assignees []struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
		HTMLURL   string `json:"html_url"`
		URL       string `json:"url"`
	} `json:"assignees"`
}

// ListKanbanBoardItemsOptions carries the documented item-list filters.
type ListKanbanBoardItemsOptions struct {
	ListOptions
	Repo         string `json:"repo,omitempty"`
	SourceType   string `json:"source_type,omitempty"`   // issue / merge_request
	SourceStatus string `json:"source_status,omitempty"` // opened / closed / merged
	Title        string `json:"title,omitempty"`
	SourceIIDs   string `json:"source_iids,omitempty"`
}

// ListKanbanBoardItems lists a board's items.
//
// GET /v5/org/{org}/kanban/{kanban_id}/item_list (live-verified).
func (c *Client) ListKanbanBoardItems(ctx context.Context, org, kanbanID string, opts ListKanbanBoardItemsOptions) ([]*KanbanBoardItem, error) {
	q := urlValuesFromListOptions(opts.ListOptions)
	setIf := func(key, val string) {
		if val != "" {
			q.Set(key, val)
		}
	}
	setIf("repo", opts.Repo)
	setIf("source_type", opts.SourceType)
	setIf("source_status", opts.SourceStatus)
	setIf("title", opts.Title)
	setIf("source_iids", opts.SourceIIDs)
	var items []*KanbanBoardItem
	err := c.doRequest(ctx, http.MethodGet, pathf("/org/%s/kanban/%s/item_list?%s", org, kanbanID, q.Encode()), nil, &items)
	if err != nil {
		return nil, err
	}
	return items, nil
}

// KanbanItemRef is the repo-scoped item reference shared by add/remove:
// repo is the repository path; IssueIIDs/PrIIDs list the in-repo iids.
type KanbanItemRef struct {
	Repo      string  `json:"repo"`
	IssueIIDs []int64 `json:"issue_iids"`
	PrIIDs    []int64 `json:"pr_iids"`
}

// AddToKanbanBoard adds repo issues/PRs to a board.
//
// POST /v5/org/{org}/kanban/{kanban_id}/add_item (live-verified; the
// response is an empty object per the docs).
func (c *Client) AddToKanbanBoard(ctx context.Context, org, kanbanID string, items []KanbanItemRef) error {
	return c.doRequest(ctx, http.MethodPost, pathf("/org/%s/kanban/%s/add_item", org, kanbanID), items, nil)
}

// UpdateKanbanBoardStatus sets a board open/closed (state: "open" or
// "closed" on the wire).
//
// PUT /v5/org/{org}/kanban/{kanban_id}/state (live-verified).
func (c *Client) UpdateKanbanBoardStatus(ctx context.Context, org, kanbanID, state string) error {
	return c.doRequest(ctx, http.MethodPut, pathf("/org/%s/kanban/%s/state", org, kanbanID), map[string]string{"state": state}, nil)
}

// RemoveFromKanbanBoard removes repo issues/PRs from a board.
//
// DELETE /v5/org/{org}/kanban/{kanban_id}/remove_item (live-verified; the
// body is an array per the docs).
func (c *Client) RemoveFromKanbanBoard(ctx context.Context, org, kanbanID string, items []KanbanItemRef) error {
	return c.doRequest(ctx, http.MethodDelete, pathf("/org/%s/kanban/%s/remove_item", org, kanbanID), items, nil)
}

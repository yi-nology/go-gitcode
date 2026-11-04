package gitcode

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// kanbanV8Capture mirrors the actions_v8 capture helper: records method,
// path, query, body per request and answers with a fixed body.
func kanbanCapture(t *testing.T, resp string, capture func(*v8Capture)) *Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		if capture != nil {
			q, _ := url.ParseQuery(r.URL.RawQuery)
			capture(&v8Capture{Method: r.Method, Path: r.URL.Path, Query: q, Body: string(raw)})
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	t.Cleanup(server.Close)
	return NewClientWithBaseURL(server.URL+"/api/v5", "test-token")
}

// All kanban routes ride the singular /org/ prefix (live-verified: the
// /orgs/.../boards family 404s framework-style while /org/.../kanban/...
// hits KANBAN_NON_EXISTENT business logic; probe round 2026-11-04).
func TestKanbanRoutes_LiveVerifiedShapes(t *testing.T) {
	listBody := `{"close_count":0,"open_count":1,"all_count":1,"content":[{"id":"kb-1","iid":3,"name":"Sprint","status":0,"visibility":1,"updated_at":"2026-01-01"}]}`
	var cap v8Capture
	serve := func(resp string) *Client {
		return kanbanCapture(t, resp, func(c *v8Capture) { cap = *c })
	}
	client := serve(listBody)

	list, err := client.ListKanbanBoards(context.Background(), "my-org", ListKanbanBoardsOptions{
		ListOptions: ListOptions{Page: 2, PerPage: 10},
		Status:      1, Sort: "newest", Search: "sprint",
	})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if cap.Path != "/api/v5/org/my-org/kanban/list" {
		t.Errorf("list path = %s", cap.Path)
	}
	if cap.Query.Get("status") != "1" || cap.Query.Get("search") != "sprint" {
		t.Errorf("list query: %v", cap.Query)
	}
	if list.AllCount != 1 || list.Content[0].ID != "kb-1" || list.Content[0].Iid != 3 {
		t.Errorf("list envelope: %+v", list)
	}

	board, err := serve(`{"id":"kb-1","name":"Sprint","description":"d","status":0,"visibility":1}`).GetKanbanBoard(context.Background(), "my-org", "kb-1")
	if err != nil {
		t.Fatalf("detail: %v", err)
	}
	if cap.Path != "/api/v5/org/my-org/kanban/kb-1/detail" {
		t.Errorf("detail path = %s", cap.Path)
	}
	if board.Name != "Sprint" {
		t.Errorf("detail board: %+v", board)
	}

	items, err := serve(`[{"id":9,"number":4,"title":"t","source_type":"issue","status":"opened","repo":{"id":1,"full_name":"o/r"}}]`).ListKanbanBoardItems(context.Background(), "my-org", "kb-1", ListKanbanBoardItemsOptions{
		SourceType: "issue", SourceStatus: "opened", SourceIIDs: "4,5",
	})
	if err != nil {
		t.Fatalf("items: %v", err)
	}
	if cap.Path != "/api/v5/org/my-org/kanban/kb-1/item_list" {
		t.Errorf("items path = %s", cap.Path)
	}
	if cap.Query.Get("source_type") != "issue" || cap.Query.Get("source_iids") != "4,5" {
		t.Errorf("items query: %v", cap.Query)
	}
	_ = items

	addClient := serve(`{}`)
	if err := addClient.AddToKanbanBoard(context.Background(), "my-org", "kb-1", []KanbanItemRef{{Repo: "o/r", IssueIIDs: []int64{4}, PrIIDs: []int64{}}}); err != nil {
		t.Fatalf("add: %v", err)
	}
	if cap.Path != "/api/v5/org/my-org/kanban/kb-1/add_item" || cap.Method != http.MethodPost {
		t.Errorf("add %s %s", cap.Method, cap.Path)
	}
	var addBody []map[string]any
	if err := json.Unmarshal([]byte(cap.Body), &addBody); err != nil || len(addBody) != 1 || addBody[0]["repo"] != "o/r" {
		t.Errorf("add body: %v", cap.Body)
	}

	stateClient := serve(`{}`)
	if err := stateClient.UpdateKanbanBoardStatus(context.Background(), "my-org", "kb-1", "closed"); err != nil {
		t.Fatalf("state: %v", err)
	}
	if cap.Path != "/api/v5/org/my-org/kanban/kb-1/state" || cap.Method != http.MethodPut {
		t.Errorf("state %s %s", cap.Method, cap.Path)
	}
	if !strings.Contains(cap.Body, `"state":"closed"`) {
		t.Errorf("state body: %v", cap.Body)
	}

	rmClient := serve(`{}`)
	if err := rmClient.RemoveFromKanbanBoard(context.Background(), "my-org", "kb-1", []KanbanItemRef{{Repo: "o/r", IssueIIDs: []int64{4}}}); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if cap.Path != "/api/v5/org/my-org/kanban/kb-1/remove_item" || cap.Method != http.MethodDelete {
		t.Errorf("remove %s %s", cap.Method, cap.Path)
	}
	var rmBody []map[string]any
	if err := json.Unmarshal([]byte(cap.Body), &rmBody); err != nil || len(rmBody) != 1 || rmBody[0]["repo"] != "o/r" {
		t.Errorf("remove body: %v", cap.Body)
	}
}

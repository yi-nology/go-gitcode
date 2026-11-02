package gitcode

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// newWireTestClient returns a client pointed at a recorder server; the
// handler receives each request and replies with body.
func newWireTestClient(t *testing.T, body string, capture func(r *http.Request, reqBody string)) *Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		if capture != nil {
			capture(r, string(raw))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	return NewClientWithBaseURL(server.URL+"/api/v5", "test-token")
}

// The GitCode wire marks due_on as a date string with minLength 1; an unset
// due date must stay off the wire, not go out as "" (which clears or errors).
func TestCreateMilestoneDueOnOmittedWhenEmpty(t *testing.T) {
	var reqBody string
	client := newWireTestClient(t, `{"id":1,"number":3,"title":"m","state":"open"}`, func(r *http.Request, b string) {
		reqBody = b
	})
	ms, err := client.CreateMilestoneWithOptions(context.Background(), "o", "r", CreateMilestoneOptions{Title: "m"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ms.Number != 3 {
		t.Errorf("expected Number 3, got %d", ms.Number)
	}
	if strings.Contains(reqBody, "due_on") {
		t.Errorf("empty due_on must be omitted, got body %s", reqBody)
	}

	if _, err := client.CreateMilestoneWithOptions(context.Background(), "o", "r", CreateMilestoneOptions{Title: "m", DueOn: "2026-01-31"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(reqBody, `"due_on":"2026-01-31"`) {
		t.Errorf("set due_on must be sent verbatim, got body %s", reqBody)
	}
}

func TestUpdateMilestoneOmitsUnsetFields(t *testing.T) {
	var reqBody string
	client := newWireTestClient(t, `{"number":3}`, func(r *http.Request, b string) {
		reqBody = b
	})
	if _, err := client.UpdateMilestone(context.Background(), "o", "r", 3, UpdateMilestoneOptions{State: "closed"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(reqBody, "due_on") || strings.Contains(reqBody, `"title"`) {
		t.Errorf("unset title/due_on must stay off the wire, got body %s", reqBody)
	}
	if !strings.Contains(reqBody, `"state":"closed"`) {
		t.Errorf("state must be sent, got body %s", reqBody)
	}
}

// Milestone must parse the documented response shape: date-only due_on,
// integer number, no id/due_date fields.
func TestMilestoneParsesDocumentedShape(t *testing.T) {
	client := newWireTestClient(t, `{
		"closed_issues": 0,
		"created_at": "2024-10-08T10:58:16+08:00",
		"description": "你好中国",
		"due_on": "2024-11-08",
		"number": 4914,
		"open_issues": 2,
		"repository_id": 4066481,
		"state": "active",
		"title": "你好中国",
		"updated_at": "2024-10-08T10:58:16+08:00",
		"url": "https://gitcode.com/o/r/milestones/1"
	}`, nil)
	ms, err := client.GetMilestone(context.Background(), "o", "r", 4914)
	if err != nil {
		t.Fatalf("documented milestone shape must unmarshal: %v", err)
	}
	if ms.Number != 4914 || ms.DueOn != "2024-11-08" || ms.Title != "你好中国" {
		t.Errorf("unexpected milestone: %+v", ms)
	}
	if ms.RepositoryID != 4066481 || ms.OpenIssues != 2 || ms.ClosedIssues != 0 {
		t.Errorf("unexpected milestone counters: %+v", ms)
	}
}

// The hook wire has no "events" array: the body carries password plus the
// per-family event flags.
func TestCreateWebhookSendsPasswordAndFlags(t *testing.T) {
	var reqBody string
	client := newWireTestClient(t, `{"id":7,"url":"https://example.com/hook"}`, func(r *http.Request, b string) {
		reqBody = b
	})
	hook, err := client.CreateWebhook(context.Background(), "o", "r", CreateWebhookOptions{
		URL:                 "https://example.com/hook",
		Password:            "s3cret",
		PushEvents:          Bool(true),
		MergeRequestsEvents: Bool(true),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hook.ID != 7 {
		t.Errorf("expected hook id 7, got %d", hook.ID)
	}
	for _, forbidden := range []string{`"secret"`, `"events"`} {
		if strings.Contains(reqBody, forbidden) {
			t.Errorf("body must not carry %s, got %s", forbidden, reqBody)
		}
	}
	for _, required := range []string{`"password":"s3cret"`, `"push_events":true`, `"merge_requests_events":true`} {
		if !strings.Contains(reqBody, required) {
			t.Errorf("body must carry %s, got %s", required, reqBody)
		}
	}
}

// Nil flags stay off the wire (unchanged); an explicit Bool(false) is sent
// so turning a family off is expressible.
func TestUpdateWebhookExpressesExplicitFlags(t *testing.T) {
	var reqBody string
	client := newWireTestClient(t, `{"id":7}`, func(r *http.Request, b string) {
		reqBody = b
	})
	if _, err := client.UpdateWebhook(context.Background(), "o", "r", 7, UpdateWebhookOptions{
		URL:        "https://example.com/hook2",
		PushEvents: Bool(false),
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(reqBody, `"push_events":false`) {
		t.Errorf("explicit false must be sent, got %s", reqBody)
	}
	if strings.Contains(reqBody, "issues_events") {
		t.Errorf("nil flags must stay off the wire, got %s", reqBody)
	}
}

// The documented hook response types the flags as integers on some
// deployments; the client must tolerate 0/1, true/false, and strings.
func TestWebhookResponseFlagShapes(t *testing.T) {
	client := newWireTestClient(t, `{
		"id": 7,
		"url": "https://example.com/hook",
		"password": "s3cret",
		"project_id": 42,
		"push_events": 1,
		"tag_push_events": 0,
		"issues_events": true,
		"note_events": "false",
		"merge_requests_events": 1,
		"active": 1
	}`, nil)
	hook, err := client.GetWebhook(context.Background(), "o", "r", 7)
	if err != nil {
		t.Fatalf("integer-flag response must unmarshal: %v", err)
	}
	if !hook.PushEvents || hook.TagPushEvents || !hook.IssuesEvents || hook.NoteEvents || !hook.MergeRequestsEvents || !hook.Active {
		t.Errorf("unexpected flags: %+v", hook)
	}
	if hook.Password != "s3cret" || hook.ProjectID != 42 {
		t.Errorf("unexpected hook: %+v", hook)
	}
}

func TestListIssueCommentsPaginationParams(t *testing.T) {
	var rawQuery string
	client := newWireTestClient(t, `[]`, func(r *http.Request, _ string) {
		rawQuery = r.URL.RawQuery
	})
	if _, err := client.ListIssueCommentsWithOptions(context.Background(), "o", "r", 5, ListIssueCommentsOptions{
		ListOptions: ListOptions{Page: 2, PerPage: 50},
		Order:       "desc",
		Since:       "2024-11-10T08:10:30.000+08:00",
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	q, err := url.ParseQuery(rawQuery)
	if err != nil {
		t.Fatalf("query must stay parseable: %v", err)
	}
	if q.Get("page") != "2" || q.Get("per_page") != "50" || q.Get("order") != "desc" {
		t.Errorf("unexpected query: %s", rawQuery)
	}
	if q.Get("since") != "2024-11-10T08:10:30.000+08:00" {
		t.Errorf("since must be URL-encoded intact, got %s", rawQuery)
	}
}

func TestListIssueLabelsPaginationParams(t *testing.T) {
	var rawQuery string
	client := newWireTestClient(t, `[{"id":1,"name":"bug","color":"#FF0000","repository_id":9}]`, func(r *http.Request, _ string) {
		rawQuery = r.URL.RawQuery
	})
	labels, err := client.ListIssueLabelsWithOptions(context.Background(), "o", "r", ListOptions{Page: 3, PerPage: 100})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(labels) != 1 || labels[0].Name != "bug" {
		t.Errorf("unexpected labels: %+v", labels)
	}
	q, err := url.ParseQuery(rawQuery)
	if err != nil {
		t.Fatalf("query must stay parseable: %v", err)
	}
	if q.Get("page") != "3" || q.Get("per_page") != "100" {
		t.Errorf("unexpected query: %s", rawQuery)
	}
}

func TestListWebhooksPaginationParams(t *testing.T) {
	var rawQuery string
	client := newWireTestClient(t, `[]`, func(r *http.Request, _ string) {
		rawQuery = r.URL.RawQuery
	})
	if _, err := client.ListWebhooksWithOptions(context.Background(), "o", "r", ListOptions{Page: 2, PerPage: 20}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	q, err := url.ParseQuery(rawQuery)
	if err != nil {
		t.Fatalf("query must stay parseable: %v", err)
	}
	if q.Get("page") != "2" || q.Get("per_page") != "20" {
		t.Errorf("unexpected query: %s", rawQuery)
	}
}

// Filter values with spaces or '&' must be escaped, not spliced raw into
// the query string.
func TestListIssuesEscapesQueryValues(t *testing.T) {
	var rawQuery string
	client := newWireTestClient(t, `[]`, func(r *http.Request, _ string) {
		rawQuery = r.URL.RawQuery
	})
	if _, err := client.ListIssues(context.Background(), "o", "r", ListIssuesOptions{
		Labels: "good first issue",
		Since:  "2024-11-10T08:10:30.000+08:00",
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	q, err := url.ParseQuery(rawQuery)
	if err != nil {
		t.Fatalf("query must stay parseable: %v", err)
	}
	if q.Get("labels") != "good first issue" {
		t.Errorf("labels value must survive round-trip, got %s", rawQuery)
	}
	if q.Get("since") != "2024-11-10T08:10:30.000+08:00" {
		t.Errorf("since must be URL-encoded intact, got %s", rawQuery)
	}
}

func TestFlexBoolStringAndNumberForms(t *testing.T) {
	for tc, want := range map[string]bool{
		`true`: true, `1`: true, `"1"`: true, `"true"`: true, `"yes"`: true,
		`false`: false, `0`: false, `"0"`: false, `"false"`: false, `""`: false,
	} {
		var fb FlexBool
		if err := fb.UnmarshalJSON([]byte(tc)); err != nil {
			t.Fatalf("FlexBool(%s): unexpected error: %v", tc, err)
		}
		if bool(fb) != want {
			t.Errorf("FlexBool(%s) = %v, want %v", tc, fb, want)
		}
	}
	var fb FlexBool
	if err := fb.UnmarshalJSON([]byte(`"bogus"`)); err == nil {
		t.Error("FlexBool should reject non-boolean strings")
	}
}

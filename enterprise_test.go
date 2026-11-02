package gitcode

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetOrgEnterprise_V8Wire(t *testing.T) {
	var cap v8Capture
	client := newV8TestClient(t, `{"id":42,"name":"acme"}`, 200, func(c *v8Capture) { cap = *c })
	ent, err := client.GetOrgEnterprise(context.Background(), "my-org")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The v8 wire uses the singular /org/ segment.
	if cap.Path != "/api/v8/org/my-org/enterprise" {
		t.Errorf("path = %s", cap.Path)
	}
	if ent.ID != 42 || ent.Name != "acme" {
		t.Errorf("unexpected enterprise: %+v", ent)
	}
}

// The roles endpoint lives under the singular /enterprise/ prefix per its
// docs page (the rest of the family uses /enterprises/).
func TestListEnterpriseCustomizedRoles_V8Wire(t *testing.T) {
	var cap v8Capture
	client := newV8TestClient(t, `[{"id":"r1","name":"QA","name_cn":"测试","access_level":30,"member_count":5}]`, 200, func(c *v8Capture) { cap = *c })
	roles, err := client.ListEnterpriseCustomizedRoles(context.Background(), "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.Path != "/api/v8/enterprise/42/customized_roles" {
		t.Errorf("path = %s", cap.Path)
	}
	if len(roles) != 1 || roles[0].ID != "r1" || roles[0].NameCN != "测试" || roles[0].MemberCount != 5 {
		t.Errorf("unexpected roles: %+v", roles)
	}
}

func TestEnterpriseMilestones_V8Wire(t *testing.T) {
	var cap v8Capture
	client := newV8TestClient(t, `{"id":7,"title":"v1","description":"d","state":"active","created_at":"2026-01-01","updated_at":"2026-01-02","due_date":"2026-03-01","start_date":"2026-01-01","projects":["p1"]}`, 200,
		func(c *v8Capture) { cap = *c })

	// Create posts the documented body shape (dates + project ids).
	ms, err := client.CreateEnterpriseMilestone(context.Background(), "42", CreateEnterpriseMilestoneOptions{
		Title: "v1", StartDate: "2026-01-01", DueDate: "2026-03-01", Projects: []int64{9, 10},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if cap.Method != http.MethodPost || cap.Path != "/api/v8/enterprises/42/milestones" {
		t.Errorf("create %s %s", cap.Method, cap.Path)
	}
	var body map[string]any
	_ = json.Unmarshal([]byte(cap.Body), &body)
	if body["title"] != "v1" || body["due_date"] != "2026-03-01" {
		t.Errorf("create body: %v", cap.Body)
	}
	projects, _ := body["projects"].([]any)
	if len(projects) != 2 || projects[0] != float64(9) {
		t.Errorf("projects mismatch: %v", cap.Body)
	}
	if ms.ID != 7 || ms.DueDate != "2026-03-01" {
		t.Errorf("unexpected milestone: %+v", ms)
	}

	// Update is a PUT.
	if _, err := client.UpdateEnterpriseMilestone(context.Background(), "42", "7", UpdateEnterpriseMilestoneOptions{Title: "v1-renamed"}); err != nil {
		t.Fatalf("update: %v", err)
	}
	if cap.Method != http.MethodPut || cap.Path != "/api/v8/enterprises/42/milestones/7" {
		t.Errorf("update %s %s", cap.Method, cap.Path)
	}

	// Delete.
	if err := client.DeleteEnterpriseMilestone(context.Background(), "42", "7"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if cap.Method != http.MethodDelete || cap.Path != "/api/v8/enterprises/42/milestones/7" {
		t.Errorf("delete %s %s", cap.Method, cap.Path)
	}
}

// The label wire segment is singular; the documented-required description
// ships even when empty (the docs mark it required).
func TestEnterpriseLabels_V8Wire(t *testing.T) {
	var cap v8Capture
	client := newV8TestClient(t, `{"id":3,"name":"bug","color":"#FF0000"}`, 200, func(c *v8Capture) { cap = *c })
	label, err := client.CreateEnterpriseLabel(context.Background(), "42", CreateEnterpriseLabelOptions{Name: "bug", Color: "#FF0000"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if cap.Path != "/api/v8/enterprises/42/label" {
		t.Errorf("path = %s", cap.Path)
	}
	var body map[string]any
	_ = json.Unmarshal([]byte(cap.Body), &body)
	if body["name"] != "bug" || body["color"] != "#FF0000" {
		t.Errorf("create body: %v", cap.Body)
	}
	if _, has := body["description"]; !has {
		t.Errorf("documented-required description must ship, got %v", cap.Body)
	}
	if label.Name != "bug" {
		t.Errorf("unexpected label: %+v", label)
	}

	if _, err := client.UpdateEnterpriseLabel(context.Background(), "42", "3", UpdateEnterpriseLabelOptions{Color: "#00FF00"}); err != nil {
		t.Fatalf("update: %v", err)
	}
	if cap.Method != http.MethodPut || cap.Path != "/api/v8/enterprises/42/label/3" {
		t.Errorf("update %s %s", cap.Method, cap.Path)
	}
}

func TestEnterpriseMembers_V8Wire(t *testing.T) {
	memberBody := `{"id":"u1","login":"dev","name":"Dev","permissions":{"admin":false,"customized":true,"push":1,"pull":true}}`
	var cap v8Capture
	client := newV8TestClient(t, memberBody, 200, func(c *v8Capture) { cap = *c })

	// Invite rides /memberships and carries permission+role_id.
	member, err := client.InviteEnterpriseMember(context.Background(), "42", "dev", EnterpriseMemberOptions{Permission: "customized", RoleID: "r1"})
	if err != nil {
		t.Fatalf("invite: %v", err)
	}
	if cap.Method != http.MethodPost || cap.Path != "/api/v8/enterprises/42/memberships/dev" {
		t.Errorf("invite %s %s", cap.Method, cap.Path)
	}
	var body map[string]any
	_ = json.Unmarshal([]byte(cap.Body), &body)
	if body["permission"] != "customized" || body["role_id"] != "r1" {
		t.Errorf("invite body: %v", cap.Body)
	}
	if member.Permissions == nil || !member.Permissions.Customized || !member.Permissions.Push {
		t.Errorf("permissions must tolerate boolean and integer forms: %+v", member.Permissions)
	}

	// Permission update is a PUT on members.
	if _, err := client.UpdateEnterpriseMemberPermission(context.Background(), "42", "dev", EnterpriseMemberOptions{Permission: "admin"}); err != nil {
		t.Fatalf("update permission: %v", err)
	}
	if cap.Method != http.MethodPut || cap.Path != "/api/v8/enterprises/42/members/dev" {
		t.Errorf("update %s %s", cap.Method, cap.Path)
	}

	// Delete joins several usernames onto the plural segment.
	if err := client.DeleteEnterpriseMember(context.Background(), "42", "a", "b"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if cap.Method != http.MethodDelete || cap.Path != "/api/v8/enterprises/42/members/a,b" {
		t.Errorf("delete %s %s", cap.Method, cap.Path)
	}
}

func TestListEnterpriseIssueCustomFields_V8Wire(t *testing.T) {
	var cap v8Capture
	client := newV8TestClient(t, `[{"id":1,"name":"环境","field_type":"select_multi","field_visibility":1,"options":[{"id":5,"extend_field_id":1,"sort_num":1,"field_value":"prod"}],"types":[{"id":2,"name":"缺陷","enabled":true,"sort_num":1,"kind":0}]}]`, 200,
		func(c *v8Capture) { cap = *c })
	fields, err := client.ListEnterpriseIssueCustomFields(context.Background(), "42", ListOptions{Page: 1, PerPage: 50})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.Path != "/api/v8/enterprises/42/issue_extend_field" {
		t.Errorf("path = %s", cap.Path)
	}
	if cap.Query.Get("per_page") != "50" {
		t.Errorf("pagination missing: %v", cap.Query)
	}
	if len(fields) != 1 || fields[0].FieldType != "select_multi" || fields[0].Options[0].FieldValue != "prod" || !fields[0].Types[0].Enabled {
		t.Errorf("unexpected fields: %+v", fields)
	}
}

// The v8 enterprise issue list is a POST (query-by-filter body) with
// page/per_page defaulted when unset.
func TestListEnterpriseIssues_V8PostBody(t *testing.T) {
	var cap v8Capture
	client := newV8TestClient(t, `[{"id":11,"number":"3","title":"bug in prod","state":"open","user":{"id":"u1","login":"dev"},"labels":[{"id":3,"name":"bug","color":"#FF0000"}],"comments":2}]`, 200,
		func(c *v8Capture) { cap = *c })
	issues, err := client.ListEnterpriseIssues(context.Background(), "42", ListEnterpriseIssuesOptions{
		PerPage: 30, Search: "prod", Assignees: []string{"dev"}, MilestoneIDs: []int64{7},
		CustomFields: []*EnterpriseCustomFieldFilter{{FieldName: "环境", Values: []string{"prod"}, Operation: "in"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.Method != http.MethodPost || cap.Path != "/api/v8/enterprises/42/issues" {
		t.Errorf("method/path = %s %s", cap.Method, cap.Path)
	}
	var body map[string]any
	if err := json.Unmarshal([]byte(cap.Body), &body); err != nil {
		t.Fatalf("body must be JSON: %v", err)
	}
	if body["search"] != "prod" || body["page"] != float64(1) || body["per_page"] != float64(30) {
		t.Errorf("filter body mismatch: %v", cap.Body)
	}
	if _, has := body["custom_fields"]; !has {
		t.Errorf("custom_fields filter missing: %v", cap.Body)
	}
	if len(issues) != 1 || issues[0].Number != "3" || issues[0].Labels[0].Name != "bug" || issues[0].Comments != 2 {
		t.Errorf("unexpected issues: %+v", issues)
	}
	if !strings.HasPrefix(cap.Path, "/api/v8/") {
		t.Errorf("enterprise must hit the v8 base: %s", cap.Path)
	}
}

// --- v5-legacy enterprise endpoints (enterprise_v5.go) ---

func TestEnterpriseV5Legacy_Wire(t *testing.T) {
	newV5 := func(body string, capture func(*http.Request)) *Client {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if capture != nil {
				capture(r)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(body))
		}))
		t.Cleanup(server.Close)
		return NewClientWithBaseURL(server.URL+"/api/v5", "test-token")
	}

	ctx := context.Background()

	statuses, err := newV5(`[{"id":1,"name":"open","is_default":true}]`, func(r *http.Request) {
		if r.URL.Path != "/api/v5/enterprises/42/issue-statuses" {
			t.Errorf("statuses path = %s", r.URL.Path)
		}
	}).ListEnterpriseIssueStatuses(ctx, "42")
	if err != nil || len(statuses) != 1 || !statuses[0].IsDefault {
		t.Errorf("statuses: %v %+v", err, statuses)
	}

	issue, err := newV5(`{"number":3,"title":"bug"}`, func(r *http.Request) {
		if r.URL.Path != "/api/v5/enterprises/42/issues/3" {
			t.Errorf("issue path = %s", r.URL.Path)
		}
	}).GetEnterpriseIssue(ctx, "42", 3)
	if err != nil || issue.Title != "bug" {
		t.Errorf("issue: %v %+v", err, issue)
	}

	comments, err := newV5(`[{"id":9,"body":"hi"}]`, func(r *http.Request) {
		if r.URL.Path != "/api/v5/enterprises/42/issues/3/comments" {
			t.Errorf("comments path = %s", r.URL.Path)
		}
	}).ListEnterpriseIssueComments(ctx, "42", 3, ListOptions{})
	if err != nil || len(comments) != 1 || comments[0].Body != "hi" {
		t.Errorf("comments: %v %+v", err, comments)
	}

	labels, err := newV5(`[{"id":3,"name":"bug"}]`, func(r *http.Request) {
		if r.URL.Path != "/api/v5/enterprises/42/issues/7/labels" {
			t.Errorf("labels path = %s", r.URL.Path)
		}
	}).ListEnterpriseIssueLabels(ctx, "42", 7)
	if err != nil || len(labels) != 1 || labels[0].Name != "bug" {
		t.Errorf("labels: %v %+v", err, labels)
	}
}

// The issue→PR links use the documented hyphen spelling (v0.12.0: was the
// Gitee-style underscore).
func TestIssueLinkedPRs_HyphenSpelling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v5/repos/o/r/issues/3/pull-requests" {
			t.Errorf("path = %s (want the documented pull-requests spelling)", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()
	client := NewClientWithBaseURL(server.URL+"/api/v5", "test-token")
	if _, err := client.GetIssueLinkedPRs(context.Background(), "o", "r", 3); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

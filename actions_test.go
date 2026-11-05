package gitcode

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// v8TestClient mounts a client over a recorder server. Every request is
// captured (method, path incl. the /api/v8 or /api/v5 prefix, query, body)
// and answered with body (or 201 when code is set); the handler answers
// with `resp` for every request.
type v8Capture struct {
	Method string
	Path   string
	Query  url.Values
	Body   string
}

func newV8TestClient(t *testing.T, resp string, code int, capture func(*v8Capture)) *Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		if capture != nil {
			q, _ := url.ParseQuery(r.URL.RawQuery)
			capture(&v8Capture{Method: r.Method, Path: r.URL.Path, Query: q, Body: string(raw)})
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_, _ = w.Write([]byte(resp))
	}))
	t.Cleanup(server.Close)
	return NewClientWithBaseURL(server.URL+"/api/v5", "test-token")
}

// The v8 base derives from the configured base by swapping the /v5 suffix;
// a base without /v5 passes through unchanged.
func TestV8BaseURLDerivation(t *testing.T) {
	if got := NewClientWithBaseURL("https://api.gitcode.com/api/v5", "t").v8BaseURL(); got != "https://api.gitcode.com/api/v8" {
		t.Errorf("v8BaseURL() = %q, want the /v5→/v8 swap", got)
	}
	if got := NewClientWithBaseURL("https://git.example.com/api", "t").v8BaseURL(); got != "https://git.example.com/api" {
		t.Errorf("v8BaseURL() = %q, want passthrough without a /v5 suffix", got)
	}
}

func TestListActionWorkflows_V8Wire(t *testing.T) {
	var cap v8Capture
	client := newV8TestClient(t, `{"total_count":1,"workflows":[{"workflow_id":"wf-1","file_path":".workflow/ci.yml","name":"CI","state":"ACTIVE"}]}`, 200,
		func(c *v8Capture) { cap = *c })
	resp, err := client.ListActionWorkflows(context.Background(), "o", "r", ListOptions{Page: 2, PerPage: 30})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(cap.Path, "/api/v8/") {
		t.Errorf("actions must hit the derived v8 base, got %s", cap.Path)
	}
	if cap.Path != "/api/v8/repos/o/r/actions/workflows" {
		t.Errorf("path = %s", cap.Path)
	}
	if cap.Query.Get("page") != "2" || cap.Query.Get("per_page") != "30" {
		t.Errorf("pagination missing: %v", cap.Query)
	}
	if resp.TotalCount != 1 || len(resp.Workflows) != 1 || resp.Workflows[0].WorkflowID != "wf-1" {
		t.Errorf("unexpected envelope: %+v", resp)
	}
}

func TestListActionRuns_V8Filters(t *testing.T) {
	var cap v8Capture
	client := newV8TestClient(t, `{"total_count":0,"workflow_runs":[]}`, 200, func(c *v8Capture) { cap = *c })
	_, err := client.ListActionRuns(context.Background(), "o", "r", ListWorkflowRunsOptions{
		ListOptions: ListOptions{Page: 1, PerPage: 20},
		Event:       "MR", Status: "FAILED", Branch: "main", WorkflowID: "wf-9",
		StartTime: 1700000000, EndTime: 1700009999,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for key, want := range map[string]string{
		"event": "MR", "status": "FAILED", "branch": "main",
		"workflow_id": "wf-9", "startTime": "1700000000", "endTime": "1700009999",
	} {
		if cap.Query.Get(key) != want {
			t.Errorf("query[%s] = %q, want %q", key, cap.Query.Get(key), want)
		}
	}
}

// The run detail carries the full stage → job → step tree; documented null
// values must unmarshal as zero values without error.
func TestGetActionRun_V8Tree(t *testing.T) {
	client := newV8TestClient(t, `{
		"workflow_run_id": "run-1", "workflow_id": "wf-1", "workflow_name": "CI",
		"file_path": ".workflow/ci.yml", "title": "CI on main", "status": "RUNNING",
		"event": "Push", "run_number": 3, "head_branch": "main", "head_sha": "abc123",
		"actor": {"id": "u1", "object_id": "42", "login": "dev", "name": "Dev"},
		"start_time": 1700000000, "end_time": 0, "pause_time": 0,
		"attempt": 1, "exist_in_default_branch": true,
		"stages": [{
			"id": "s1", "name": "build", "sequence": 1, "depends_on": null,
			"status": "RUNNING", "pre": null, "post": null,
			"jobs": [{
				"id": "j1", "name": "compile", "identifier": null, "depends_on": null,
				"status": "RUNNING", "steps": [{
					"name": "go build", "task": "shell@latest", "id": "st1",
					"inputs": null, "env": null, "status": "RUNNING", "sequence": 1
				}]
			}]
		}]
	}`, 200, nil)
	run, err := client.GetActionRun(context.Background(), "o", "r", "run-1")
	if err != nil {
		t.Fatalf("null-heavy v8 tree must unmarshal: %v", err)
	}
	if run.Status != "RUNNING" || run.Attempt != 1 || !run.ExistInDefaultBranch {
		t.Errorf("unexpected run: %+v", run.WorkflowRun)
	}
	if len(run.Stages) != 1 || len(run.Stages[0].Jobs) != 1 || len(run.Stages[0].Jobs[0].Steps) != 1 {
		t.Fatalf("stage/job/step tree truncated: %+v", run.Stages)
	}
	if run.Stages[0].Jobs[0].Steps[0].Name != "go build" {
		t.Errorf("step name = %q", run.Stages[0].Jobs[0].Steps[0].Name)
	}
}

func TestDispatchActionWorkflow_V8Body(t *testing.T) {
	var cap v8Capture
	client := newV8TestClient(t, `{"workflow_id":"wf-1","workflow_run_id":"run-9"}`, 201, func(c *v8Capture) { cap = *c })
	resp, err := client.DispatchActionWorkflow(context.Background(), "o", "r", "wf-1", DispatchActionWorkflowOptions{Ref: "main"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.Method != http.MethodPost || cap.Path != "/api/v8/repos/o/r/actions/workflows/wf-1/dispatches" {
		t.Errorf("method/path = %s %s", cap.Method, cap.Path)
	}
	var body map[string]any
	if err := json.Unmarshal([]byte(cap.Body), &body); err != nil {
		t.Fatalf("body must be JSON: %v", err)
	}
	if body["ref"] != "main" {
		t.Errorf("ref = %v", body["ref"])
	}
	inputs, _ := body["inputs"].(map[string]any)
	if inputs == nil {
		t.Errorf("documented-required inputs must ship (as {} when unset), got %v", cap.Body)
	}
	if resp.WorkflowRunID != "run-9" {
		t.Errorf("workflow_run_id = %q", resp.WorkflowRunID)
	}
}

func TestValidateActionWorkflowYAML_V8Base64(t *testing.T) {
	var cap v8Capture
	client := newV8TestClient(t, `{"valid":true,"diagnostics":[]}`, 200, func(c *v8Capture) { cap = *c })
	v, err := client.ValidateActionWorkflowYAML(context.Background(), "o", "r", []byte("version: 1\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !v.Valid {
		t.Errorf("valid = false")
	}
	var body map[string]any
	_ = json.Unmarshal([]byte(cap.Body), &body)
	if body["base64_content"] != base64.StdEncoding.EncodeToString([]byte("version: 1\n")) {
		t.Errorf("base64_content mismatch: %v", body["base64_content"])
	}
}

func TestActionRunMutations_V8SuccessBody(t *testing.T) {
	mutators := []struct {
		name       string
		call       func(*Client) (bool, error)
		wantSuffix string
	}{
		{"stop", func(c *Client) (bool, error) { return c.StopActionRun(context.Background(), "o", "r", "run-1") }, "/actions/runs/run-1/stop"},
		{"retry", func(c *Client) (bool, error) { return c.RetryActionRun(context.Background(), "o", "r", "run-1") }, "/actions/runs/run-1/retry"},
		{"rerun", func(c *Client) (bool, error) { return c.RerunActionRun(context.Background(), "o", "r", "run-1") }, "/actions/runs/run-1/rerun"},
	}
	for _, tc := range mutators {
		var cap v8Capture
		client := newV8TestClient(t, `{"success":true}`, 200, func(c *v8Capture) { cap = *c })
		ok, err := tc.call(client)
		if err != nil || !ok {
			t.Fatalf("%s: ok=%v err=%v", tc.name, ok, err)
		}
		if !strings.HasSuffix(cap.Path, tc.wantSuffix) || cap.Method != http.MethodPost {
			t.Errorf("%s: %s %s", tc.name, cap.Method, cap.Path)
		}
	}
}

func TestListActionRunJobs_V8Envelope(t *testing.T) {
	var cap v8Capture
	client := newV8TestClient(t, `{"total_count":1,"jobs":[{"id":"j1","name":"build","status":"COMPLETED","steps":[{"id":"st1","name":"go build","status":"COMPLETED"}]}]}`, 200,
		func(c *v8Capture) { cap = *c })
	resp, err := client.ListActionRunJobs(context.Background(), "o", "r", "run-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.Path != "/api/v8/repos/o/r/actions/runs/run-1/jobs" {
		t.Errorf("path = %s", cap.Path)
	}
	if resp.TotalCount != 1 || resp.Jobs[0].Steps[0].Name != "go build" {
		t.Errorf("unexpected envelope: %+v", resp)
	}
}

// v8 addresses jobs under their run: runs/{run_id}/jobs/{job_id}.
func TestGetActionJob_V8Addressing(t *testing.T) {
	var cap v8Capture
	client := newV8TestClient(t, `{"id":"j1","name":"build","status":"FAILED"}`, 200, func(c *v8Capture) { cap = *c })
	job, err := client.GetActionJob(context.Background(), "o", "r", "run-1", "j1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.Path != "/api/v8/repos/o/r/actions/runs/run-1/jobs/j1" {
		t.Errorf("path = %s", cap.Path)
	}
	if job.Status != "FAILED" {
		t.Errorf("status = %q", job.Status)
	}
}

func TestGetActionJobLogs_V8Body(t *testing.T) {
	var cap v8Capture
	client := newV8TestClient(t, `{"has_more":true,"start_offset":0,"end_offset":100,"log":"line1\nline2"}`, 200, func(c *v8Capture) { cap = *c })
	logs, err := client.GetActionJobLogs(context.Background(), "o", "r", "run-1", "j1", GetActionJobLogsOptions{
		StepID: "st1", Offset: 0, Limit: 100, Sort: "asc",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !logs.HasMore || logs.EndOffset != 100 {
		t.Errorf("unexpected logs envelope: %+v", logs)
	}
	var body map[string]any
	_ = json.Unmarshal([]byte(cap.Body), &body)
	if body["step_id"] != "st1" || body["offset"] != float64(0) || body["limit"] != float64(100) || body["sort"] != "asc" {
		t.Errorf("log request body mismatch: %v", cap.Body)
	}
}

func TestDownloadActionJobLogs_V8Raw(t *testing.T) {
	var cap v8Capture
	client := newV8TestClient(t, "raw log bytes", 200, func(c *v8Capture) { cap = *c })
	raw, err := client.DownloadActionJobLogs(context.Background(), "o", "r", "run-1", "j1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(raw) != "raw log bytes" {
		t.Errorf("raw body = %q", raw)
	}
	if !strings.HasSuffix(cap.Path, "/actions/runs/run-1/jobs/j1/download-log") {
		t.Errorf("path = %s", cap.Path)
	}
}

func TestListActionArtifacts_V8QueryAndEnvelope(t *testing.T) {
	var cap v8Capture
	client := newV8TestClient(t, `{"total_count":1,"artifacts":[{"id":"a1","name":"bin","size_bytes":2048,"digest":"deadbeef","created_at":"1700000000000"}]}`, 200,
		func(c *v8Capture) { cap = *c })
	resp, err := client.ListActionArtifacts(context.Background(), "o", "r", ListActionArtifactsOptions{
		ListOptions: ListOptions{Page: 1, PerPage: 10}, Name: "bin", Direction: "desc",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.Path != "/api/v8/repos/o/r/actions/artifacts" {
		t.Errorf("path = %s", cap.Path)
	}
	if cap.Query.Get("name") != "bin" || cap.Query.Get("direction") != "desc" {
		t.Errorf("filters missing: %v", cap.Query)
	}
	if resp.Artifacts[0].SizeBytes != 2048 || resp.Artifacts[0].Digest != "deadbeef" {
		t.Errorf("unexpected artifact: %+v", resp.Artifacts[0])
	}
}

func TestListActionRunArtifacts_V8Path(t *testing.T) {
	var cap v8Capture
	client := newV8TestClient(t, `{"total_count":0,"artifacts":[]}`, 200, func(c *v8Capture) { cap = *c })
	if _, err := client.ListActionRunArtifacts(context.Background(), "o", "r", "run-1", ListOptions{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.Path != "/api/v8/repos/o/r/actions/runs/run-1/artifacts" {
		t.Errorf("path = %s", cap.Path)
	}
}

func TestDownloadActionArtifact_V8Format(t *testing.T) {
	var cap v8Capture
	client := newV8TestClient(t, "PK\x03\x04", 200, func(c *v8Capture) { cap = *c })
	raw, err := client.DownloadActionArtifact(context.Background(), "o", "r", "a1", "zip")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(raw) != "PK\x03\x04" {
		t.Errorf("raw = %q", raw)
	}
	if !strings.HasSuffix(cap.Path, "/actions/artifacts/a1/zip") {
		t.Errorf("path = %s", cap.Path)
	}
}

func TestListRepoActionRunners_V8AndVariants(t *testing.T) {
	respBody := `{"total_count":1,"runners":[{"id":"r1","runner_name":"host-1","name":"runner-1","status":"IDLE","labels":[{"label_name":"os","label_value":"linux","label_color":"#000"}]}]}`
	var cap v8Capture
	client := newV8TestClient(t, respBody, 200, func(c *v8Capture) { cap = *c })
	resp, err := client.ListRepoActionRunners(context.Background(), "o", "r", ListActionRunnersOptions{Keyword: "host"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.Path != "/api/v8/repos/o/r/actions/runners" || cap.Query.Get("keyword") != "host" {
		t.Errorf("path/query = %s %v", cap.Path, cap.Query)
	}
	if resp.Runners[0].Status != "IDLE" || resp.Runners[0].Labels[0].LabelName != "os" {
		t.Errorf("unexpected runner: %+v", resp.Runners[0])
	}
	for _, tc := range []struct {
		call   func(*Client) error
		suffix string
	}{
		{func(c *Client) error {
			_, err := c.ListRepoSharedActionRunners(context.Background(), "o", "r", ListActionRunnersOptions{})
			return err
		}, "/actions/runners/shared-runners"},
		{func(c *Client) error {
			_, err := c.ListRepoK8SRunners(context.Background(), "o", "r", ListActionRunnersOptions{})
			return err
		}, "/actions/runner-sets"},
		{func(c *Client) error {
			_, err := c.ListRepoSharedK8SRunners(context.Background(), "o", "r", ListActionRunnersOptions{})
			return err
		}, "/actions/shared-runner-sets"},
	} {
		if err := tc.call(client); err != nil {
			t.Fatalf("variant %s: %v", tc.suffix, err)
		}
		if !strings.HasSuffix(cap.Path, tc.suffix) {
			t.Errorf("variant path = %s, want suffix %s", cap.Path, tc.suffix)
		}
	}
}

func TestOrgActionRunnerGroups_V8(t *testing.T) {
	var cap v8Capture
	client := newV8TestClient(t, `{"total_count":1,"runner_groups":[{"id":"g1","name":"default","runner_count":3,"share_all":false}]}`, 200,
		func(c *v8Capture) { cap = *c })
	resp, err := client.ListOrgActionRunnerGroups(context.Background(), "my-org", ListActionRunnersOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.Path != "/api/v8/orgs/my-org/actions/runner-groups" {
		t.Errorf("path = %s", cap.Path)
	}
	if resp.RunnerGroups[0].RunnerCount != 3 {
		t.Errorf("unexpected group: %+v", resp.RunnerGroups[0])
	}
	getClient := newV8TestClient(t, `{"id":"g1","name":"default","runner_count":3,"share_all":false}`, 200, func(c *v8Capture) { cap = *c })
	group, err := getClient.GetActionRunnerGroup(context.Background(), "my-org", "g1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.Path != "/api/v8/orgs/my-org/actions/runner-groups/g1" || group.ID != "g1" || group.ShareAll {
		t.Errorf("get group: %s %+v", cap.Path, group)
	}
}

// FindActionRunByHeadSHA: the v8 wire has no head_sha filter, so the finder
// pages the documented filters and matches client-side. Hit on page 1 must
// ship prune filters and per_page=100 in the query.
func TestFindActionRunByHeadSHAHit(t *testing.T) {
	var cap v8Capture
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		q, _ := url.ParseQuery(r.URL.RawQuery)
		cap = v8Capture{Method: r.Method, Path: r.URL.Path, Query: q, Body: string(raw)}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"total_count":1,"workflow_runs":[{"workflow_run_id":"r1","head_sha":"abc123","status":"FAILED"}]}`))
	}))
	t.Cleanup(server.Close)
	client := NewClientWithBaseURL(server.URL+"/api/v5", "test-token")

	run, err := client.FindActionRunByHeadSHA(context.Background(), "o", "r", "abc123", ListWorkflowRunsOptions{Branch: "main", Status: "FAILED"})
	if err != nil {
		t.Fatal(err)
	}
	if run.WorkflowRunID != "r1" {
		t.Fatalf("expected run r1, got %q", run.WorkflowRunID)
	}
	if !strings.HasPrefix(cap.Path, "/api/v8/repos/o/r/actions/runs") {
		t.Fatalf("finder must hit the v8 runs list, got %s", cap.Path)
	}
	if cap.Query.Get("per_page") != "100" || cap.Query.Get("page") != "1" {
		t.Fatalf("finder must scan at per_page=100 from page 1, got page=%s per_page=%s", cap.Query.Get("page"), cap.Query.Get("per_page"))
	}
	if cap.Query.Get("branch") != "main" || cap.Query.Get("status") != "FAILED" {
		t.Fatalf("prune filters must reach the wire, got %v", cap.Query)
	}
}

// No matching SHA across the full list (full page, then short page) must
// surface ErrActionRunNotFound, not a fabricated run or a hang.
func TestFindActionRunByHeadSHANotFound(t *testing.T) {
	full := `{"total_count":200,"workflow_runs":[` +
		`{"workflow_run_id":"r1","head_sha":"aaa"}` + strings.Repeat(`,{"workflow_run_id":"rx","head_sha":"bbb"}`, 98) + `]}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q, _ := url.ParseQuery(r.URL.RawQuery)
		w.Header().Set("Content-Type", "application/json")
		if q.Get("page") == "2" {
			_, _ = w.Write([]byte(`{"total_count":200,"workflow_runs":[{"workflow_run_id":"r2","head_sha":"ccc"}]}`))
			return
		}
		_, _ = w.Write([]byte(full))
	}))
	t.Cleanup(server.Close)
	client := NewClientWithBaseURL(server.URL+"/api/v5", "test-token")

	run, err := client.FindActionRunByHeadSHA(context.Background(), "o", "r", "deadbeef", ListWorkflowRunsOptions{})
	if run != nil {
		t.Fatalf("expected nil run, got %+v", run)
	}
	if !errors.Is(err, ErrActionRunNotFound) {
		t.Fatalf("expected ErrActionRunNotFound, got %v", err)
	}
}

// A match on a later page must be found (walker does not stop at page 1).
func TestFindActionRunByHeadSHASecondPage(t *testing.T) {
	page1 := `{"total_count":101,"workflow_runs":[` + strings.Repeat(`{"workflow_run_id":"rx","head_sha":"aaa"},`, 99) + `{"workflow_run_id":"r99","head_sha":"bbb"}]}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q, _ := url.ParseQuery(r.URL.RawQuery)
		w.Header().Set("Content-Type", "application/json")
		if q.Get("page") == "2" {
			_, _ = w.Write([]byte(`{"total_count":101,"workflow_runs":[{"workflow_run_id":"r100","head_sha":"deadbeef"}]}`))
			return
		}
		_, _ = w.Write([]byte(page1))
	}))
	t.Cleanup(server.Close)
	client := NewClientWithBaseURL(server.URL+"/api/v5", "test-token")

	run, err := client.FindActionRunByHeadSHA(context.Background(), "o", "r", "deadbeef", ListWorkflowRunsOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if run.WorkflowRunID != "r100" {
		t.Fatalf("expected page-2 run r100, got %q", run.WorkflowRunID)
	}
}

// Empty sha must be rejected before any wire traffic.
func TestFindActionRunByHeadSHARequiresSHA(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++ }))
	t.Cleanup(server.Close)
	client := NewClientWithBaseURL(server.URL+"/api/v5", "test-token")

	if _, err := client.FindActionRunByHeadSHA(context.Background(), "o", "r", "", ListWorkflowRunsOptions{}); err == nil {
		t.Fatal("expected an error for empty sha")
	}
	if calls != 0 {
		t.Fatalf("empty sha must not hit the wire, got %d calls", calls)
	}
}

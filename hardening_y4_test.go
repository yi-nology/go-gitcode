package gitcode

// Regression tests for issues #12-#14 (sweep lane Y4): structural 404
// detection, scan-truncation distinguishability, and Multiplier default.

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// --- issue #12: 404 probes must judge by status code, not message text ---

// sweepY4ProbeCalls drives every public boolean probe that used to swallow
// errors on a "404" substring match (5 via isNotFoundError + 2 raw
// strings.Contains sites). Each gets path segments containing "404" so the
// OLD implementation would misjudge any status as Not Found.
func sweepY4ProbeCalls(c *Client) map[string]func() (bool, error) {
	return map[string]func() (bool, error){
		"IsCollaborator":      func() (bool, error) { return c.IsCollaborator(context.Background(), "org404", "repo", "user404") },
		"IsRepoAssignee":      func() (bool, error) { return c.IsRepoAssignee(context.Background(), "org404", "repo", "user404") },
		"IsOrgPublicMember":   func() (bool, error) { return c.IsOrgPublicMember(context.Background(), "org404", "user404") },
		"IsOrgBlockedUser":    func() (bool, error) { return c.IsOrgBlockedUser(context.Background(), "org404", "user404") },
		"IsFollowing":         func() (bool, error) { return c.IsFollowing(context.Background(), "user404") },
		"IsRepositoryStarred": func() (bool, error) { return c.IsRepositoryStarred(context.Background(), "org404", "repo404") },
		"IsPullRequestMerged": func() (bool, error) { return c.IsPullRequestMerged(context.Background(), "org404", "repo", 404) },
	}
}

func TestSweepY4_Probes403With404InPathSurfacesError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"message":"insufficient permissions"}`)
	}))
	defer srv.Close()

	c := NewClientWithBaseURL(srv.URL+"/api/v5", "tok")
	for name, call := range sweepY4ProbeCalls(c) {
		t.Run(name, func(t *testing.T) {
			ok, err := call()
			if err == nil {
				t.Fatalf("403 must surface (issue #12: it used to be swallowed as not-found on paths containing 404), got ok=%v err=nil", ok)
			}
			if ok {
				t.Fatal("ok must be false on error")
			}
			if !IsForbidden(err) {
				t.Fatalf("surfaced error should be the typed ForbiddenError, got %v", err)
			}
		})
	}
}

func TestSweepY4_Probes500With404InBodySurfacesError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"message":"upstream returned 404"}`)
	}))
	defer srv.Close()

	c := NewClientWithBaseURL(srv.URL+"/api/v5", "tok")
	for name, call := range sweepY4ProbeCalls(c) {
		t.Run(name, func(t *testing.T) {
			ok, err := call()
			if err == nil {
				t.Fatalf("500 with a 404 substring in the body must surface, got ok=%v err=nil", ok)
			}
			if ok {
				t.Fatal("ok must be false on error")
			}
		})
	}
}

func TestSweepY4_ProbesReal404ReturnsFalseNil(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"not found"}`)
	}))
	defer srv.Close()

	c := NewClientWithBaseURL(srv.URL+"/api/v5", "tok")
	for name, call := range sweepY4ProbeCalls(c) {
		t.Run(name, func(t *testing.T) {
			ok, err := call()
			if err != nil {
				t.Fatalf("a genuine 404 still reads as (false, nil): %v", err)
			}
			if ok {
				t.Fatal("ok must be false for not-found")
			}
		})
	}
}

func TestSweepY4_IsNotFoundErrorStructural(t *testing.T) {
	// 类型化 404：true
	if !isNotFoundError(&NotFoundError{APIError: &APIError{StatusCode: http.StatusNotFound}}) {
		t.Fatal("typed NotFoundError must match")
	}
	// 裸 APIError 携带 404 状态码：true（errors.As 分流）
	if !isNotFoundError(&APIError{StatusCode: http.StatusNotFound, Message: "nope"}) {
		t.Fatal("raw APIError with 404 status must match")
	}
	// 包装层：errors.As 要穿透 wrap
	wrapped := fmt.Errorf("probe: %w", &NotFoundError{APIError: &APIError{StatusCode: http.StatusNotFound}})
	if !isNotFoundError(wrapped) {
		t.Fatal("wrapped NotFoundError must match through errors.As")
	}
	// 非 404 一律 false——即使消息文本里带 "404"（旧实现的误判源）
	if isNotFoundError(&ForbiddenError{APIError: &APIError{StatusCode: http.StatusForbidden, URL: "/repos/org404/repo"}}) {
		t.Fatal("403 must NOT match, even with 404 in the URL")
	}
	if isNotFoundError(&APIError{StatusCode: http.StatusInternalServerError, Message: "upstream returned 404"}) {
		t.Fatal("500 must NOT match, even with 404 in the body")
	}
	if isNotFoundError(errors.New("plain network error mentioning 404")) {
		t.Fatal("non-API errors must NOT match on message text")
	}
	if isNotFoundError(nil) {
		t.Fatal("nil must not match")
	}
}

// --- issue #13: scan truncation must not masquerade as not-found ---

func fullRunsPage(n int, sha string) string {
	body := `{"total_count":99999,"workflow_runs":[`
	for i := 0; i < n; i++ {
		if i > 0 {
			body += ","
		}
		body += fmt.Sprintf(`{"workflow_run_id":"%d","status":"COMPLETED","head_sha":"other-sha-%d"}`, i, i)
	}
	return body + `]}`
}

func TestSweepY4_FindActionRunByHeadSHA_TruncatedScanIsDistinguishable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, fullRunsPage(100, "target-sha")) // 恒返满页：游走必然触顶
	}))
	defer srv.Close()

	c := NewClientWithBaseURL(srv.URL+"/api/v5", "tok")
	_, err := c.FindActionRunByHeadSHA(context.Background(), "o", "r", "target-sha", ListWorkflowRunsOptions{})
	if err == nil {
		t.Fatal("a capped scan must not end silently")
	}
	if !errors.Is(err, ErrTooManyPages) {
		t.Fatalf("err = %v, want errors.Is(err, ErrTooManyPages) — same sentinel as Paginate (issue #5 discipline)", err)
	}
	if errors.Is(err, ErrActionRunNotFound) {
		t.Fatal("a truncated scan must NOT surface as ErrActionRunNotFound — 'not scanned' is not 'does not exist'")
	}
}

func TestSweepY4_FindActionRunByHeadSHA_ShortPageStillNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"total_count":1,"workflow_runs":[{"workflow_run_id":"1","status":"COMPLETED","head_sha":"other-sha"}]}`)
	}))
	defer srv.Close()

	c := NewClientWithBaseURL(srv.URL+"/api/v5", "tok")
	_, err := c.FindActionRunByHeadSHA(context.Background(), "o", "r", "target-sha", ListWorkflowRunsOptions{})
	if !errors.Is(err, ErrActionRunNotFound) {
		t.Fatalf("short-page exhaustion should stay ErrActionRunNotFound, got %v", err)
	}
}

func TestSweepY4_FindActionRunByHeadSHA_HitOnLaterPage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("page") == "2" {
			fmt.Fprint(w, `{"total_count":150,"workflow_runs":[{"workflow_run_id":"42","status":"COMPLETED","head_sha":"target-sha"}]}`)
			return
		}
		fmt.Fprint(w, fullRunsPage(100, "target-sha"))
	}))
	defer srv.Close()

	c := NewClientWithBaseURL(srv.URL+"/api/v5", "tok")
	run, err := c.FindActionRunByHeadSHA(context.Background(), "o", "r", "target-sha", ListWorkflowRunsOptions{})
	if err != nil {
		t.Fatalf("hit on page 2 must be returned: %v", err)
	}
	if run == nil || run.WorkflowRunID != "42" {
		t.Fatalf("wrong run: %+v", run)
	}
}

// --- issue #14: unset Multiplier must not collapse backoff to zero ---

func TestSweepY4_RetryPolicyMultiplierDefault(t *testing.T) {
	// 对照：显式 2.0 与漏填的退避序列必须一致（文档缺省 2.0 归一）
	gap := func(policy RetryPolicy) []time.Duration {
		var attempts int
		var stamps []time.Time
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			stamps = append(stamps, time.Now())
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprint(w, `{"message":"down"}`)
		}))
		defer srv.Close()

		c := NewClientWithBaseURL(srv.URL+"/api/v5", "tok")
		policy.RetryableStatusCodes = []int{503}
		c.SetRetryPolicy(policy)
		_, _ = c.GetCurrentUser(context.Background())
		var gaps []time.Duration
		for i := 1; i < len(stamps); i++ {
			gaps = append(gaps, stamps[i].Sub(stamps[i-1]))
		}
		return gaps
	}

	unset := gap(RetryPolicy{MaxRetries: 2, InitialBackoff: 200 * time.Millisecond})
	explicit := gap(RetryPolicy{MaxRetries: 2, InitialBackoff: 200 * time.Millisecond, Multiplier: 2.0})
	if len(unset) != 2 || len(explicit) != 2 {
		t.Fatalf("attempts = %d/%d, want 2 gaps each", len(unset), len(explicit))
	}
	for i := range unset {
		if unset[i] <= 0 {
			t.Fatalf("issue #14 regression: retry %d gap = %v with unset Multiplier (backoff collapsed to zero, busy-wait)", i+1, unset[i])
		}
		// 归一后漏填 ≈ 显式 2.0（±抖动余量）
		if unset[i] < explicit[i]/2 || unset[i] > explicit[i]*2 {
			t.Fatalf("retry %d gap: unset=%v explicit=%v — should match the documented 2.0 default", i+1, unset[i], explicit[i])
		}
	}
	if !(unset[1] > unset[0]) {
		t.Fatalf("backoff must grow exponentially: %v vs %v", unset[0], unset[1])
	}

	// Multiplier=1.0 是合法显式选择（恒定退避），不得被归一吞掉
	constant := gap(RetryPolicy{MaxRetries: 2, InitialBackoff: 200 * time.Millisecond, Multiplier: 1.0})
	if constant[1] > constant[0]*3/2 {
		t.Fatalf("Multiplier=1.0 should give constant backoff: %v then %v", constant[0], constant[1])
	}
}

package gitcode

// Hardening tests for issues #4-#10: Retry-After gates, paginate page cap,
// path-segment escaping, response body limits, client field read locking,
// upload pipeline reuse, and ListRepositories filter params.

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

// --- issue #4: Retry-After three gates ---

func TestRetryAfterDuration(t *testing.T) {
	now := time.Date(2026, 11, 11, 12, 0, 0, 0, time.UTC)
	future := now.Add(time.Hour)
	past := now.Add(-time.Hour)

	tests := []struct {
		name   string
		header string
		want   time.Duration
		ok     bool
	}{
		{"delay-seconds", "5", 5 * time.Second, true},
		{"zero seconds rejected", "0", 0, false},
		{"negative seconds rejected", "-5", 0, false},
		{"not a number", "soon", 0, false},
		{"empty-ish garbage", " ", 0, false},
		{"http-date future", future.Format(http.TimeFormat), time.Hour, true},
		{"http-date past rejected", past.Format(http.TimeFormat), 0, false},
		{"http-date exactly now rejected", now.Format(http.TimeFormat), 0, false},
		{"overflowing seconds clamped", "9223372036854", time.Duration(math.MaxInt64), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{Header: http.Header{}}
			if tt.header != "" {
				resp.Header.Set("Retry-After", tt.header)
			}
			got, ok := retryAfterDuration(resp, now)
			if ok != tt.ok {
				t.Fatalf("ok = %v, want %v", ok, tt.ok)
			}
			if ok && got != tt.want {
				t.Errorf("duration = %v, want %v", got, tt.want)
			}
		})
	}

	if _, ok := retryAfterDuration(nil, now); ok {
		t.Error("nil response must not parse")
	}
	resp := &http.Response{Header: http.Header{}}
	if _, ok := retryAfterDuration(resp, now); ok {
		t.Error("missing header must not parse")
	}
}

func TestGetRetryWaitTimeRetryAfter(t *testing.T) {
	clampedPolicy := RetryPolicy{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     30 * time.Second,
		Multiplier:     2.0,
	}
	future := time.Now().Add(2 * time.Hour).UTC()

	tests := []struct {
		name    string
		policy  RetryPolicy
		header  string
		attempt int
		want    time.Duration
	}{
		{"retry-after within bounds", clampedPolicy, "5", 0, 5 * time.Second},
		{"retry-after clamped to MaxBackoff", clampedPolicy, "3600", 0, 30 * time.Second},
		{"negative falls back to exponential", clampedPolicy, "-5", 0, 1 * time.Second},
		{"zero falls back to exponential", clampedPolicy, "0", 0, 1 * time.Second},
		{"garbage falls back to exponential", clampedPolicy, "soon", 0, 1 * time.Second},
		{"http-date clamped to MaxBackoff", clampedPolicy, future.Format(http.TimeFormat), 0, 30 * time.Second},
		{"unset MaxBackoff uses documented default", RetryPolicy{MaxRetries: 3, InitialBackoff: 1 * time.Second, Multiplier: 2}, "3600", 0, 30 * time.Second},
		{"exponential attempt 0", clampedPolicy, "", 0, 1 * time.Second},
		{"exponential attempt 3", clampedPolicy, "", 3, 8 * time.Second},
		{"exponential clamped at attempt 10", clampedPolicy, "", 10, 30 * time.Second},
		{"zero InitialBackoff normalized", RetryPolicy{MaxRetries: 3, MaxBackoff: 30 * time.Second, Multiplier: 2}, "", 0, 1 * time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient("tok")
			c.SetRetryPolicy(tt.policy)
			resp := &http.Response{Header: http.Header{}}
			if tt.header != "" {
				resp.Header.Set("Retry-After", tt.header)
			}
			got := c.getRetryWaitTime(tt.attempt, resp)
			if got != tt.want {
				t.Errorf("getRetryWaitTime() = %v, want %v", got, tt.want)
			}
		})
	}

	c := NewClient("tok") // no policy
	if got := c.getRetryWaitTime(0, nil); got != 0 {
		t.Errorf("no policy: got %v, want 0", got)
	}
}

// TestRetryAfterHeaderHonoredOn429 wires a real 429-then-200 exchange and
// asserts the SDK waits for the (small) Retry-After instead of busy-looping.
func TestRetryAfterHeaderHonoredOn429(t *testing.T) {
	var attempts int
	var waits []time.Duration
	start := time.Now()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprint(w, `{"message":"rate limited"}`)
			return
		}
		waits = append(waits, time.Since(start))
		fmt.Fprint(w, `{"login":"u"}`)
	}))
	defer srv.Close()

	c := NewClientWithBaseURL(srv.URL+"/api/v5", "tok")
	c.SetRetryPolicy(RetryPolicy{MaxRetries: 1, RetryableStatusCodes: []int{429}})
	if _, err := c.GetCurrentUser(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
	if len(waits) != 1 || waits[0] < 900*time.Millisecond {
		t.Errorf("second attempt came too early: %v (want >= ~1s Retry-After)", waits)
	}
}

// --- issue #5: paginate page cap ---

func TestPaginatePageCap(t *testing.T) {
	calls := 0
	fetch := func(opts ListOptions) ([]*Issue, error) {
		calls++
		// Server pathology: always a full page, cursor never advances.
		out := make([]*Issue, opts.PerPage)
		for i := range out {
			out[i] = &Issue{}
		}
		return out, nil
	}

	var pages int
	var lastErr error
	for items, err := range Paginate(context.Background(), fetch) {
		if err != nil {
			lastErr = err
			break
		}
		pages++
		_ = items
	}
	if pages != maxPaginatePages {
		t.Errorf("pages yielded = %d, want %d", pages, maxPaginatePages)
	}
	if calls != maxPaginatePages {
		t.Errorf("fetch calls = %d, want %d", calls, maxPaginatePages)
	}
	if !errors.Is(lastErr, ErrTooManyPages) {
		t.Errorf("terminal error = %v, want ErrTooManyPages", lastErr)
	}
}

func TestPaginateEndsNormally(t *testing.T) {
	fetch := func(opts ListOptions) ([]string, error) {
		if opts.Page > 1 {
			return nil, nil
		}
		return []string{"a", "b"}, nil // short page: end of list
	}
	var got []string
	var errs int
	for items, err := range Paginate(context.Background(), fetch) {
		if err != nil {
			errs++
			continue
		}
		got = append(got, items...)
	}
	if errs != 0 {
		t.Errorf("errors = %d, want 0", errs)
	}
	if strings.Join(got, ",") != "a,b" {
		t.Errorf("items = %v, want [a b]", got)
	}
}

// --- issue #6: path segment escaping ---

func TestPathfEscaping(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"plain identifiers unchanged", pathf("/repos/%s/%s", "owner", "repo"), "/repos/owner/repo"},
		{"slash confined to segment", pathf("/repos/%s/%s", "evil/../admin", "repo"), "/repos/evil%2F..%2Fadmin/repo"},
		{"question mark escaped", pathf("/orgs/%s/hooks", "org?x=1"), "/orgs/org%3Fx=1/hooks"},
		{"fragment escaped", pathf("/repos/%s", "o#w"), "/repos/o%23w"},
		{"space escaped", pathf("/repos/%s/%s", "a b", "r"), "/repos/a%20b/r"},
		{"int verb untouched", pathf("/repos/%s/pulls/%d", "o", 42), "/repos/o/pulls/42"},
		{"query verbs untouched", pathf("/repos/%s?%s", "o", "page=1&x=%2F"), "/repos/o?page=1&x=%2F"},
		{"multi-segment file path", pathf("/repos/%s/contents/%s", "o", pathSegs("docs/READ ME.md")), "/repos/o/contents/docs/READ%20ME.md"},
		{"pathSegs on plain segment", string(pathSegs("plain")), "plain"},
		{"literal percent", pathf("/a%%b/%s", "c"), "/a%b/c"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestPathSegEquivalence(t *testing.T) {
	if pathSeg("x") != string(pathSegs("x")) {
		t.Error("single-segment inputs should agree between pathSeg and pathSegs")
	}
}

// TestEndpointPathIsolation asserts the wire form: a caller-controlled
// identifier stays inside its own path segment instead of adding levels or
// spilling into the query string.
func TestEndpointPathIsolation(t *testing.T) {
	tests := []struct {
		name           string
		owner          string
		wantRequestURI string
	}{
		{"plain", "good-org", "/api/v5/repos/good-org/repo"},
		{"dotdot slash", "evil/../admin", "/api/v5/repos/evil%2F..%2Fadmin/repo"},
		{"query injection", "org?x=1", "/api/v5/repos/org%3Fx=1/repo"},
		{"fragment", "o#w", "/api/v5/repos/o%23w/repo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotURI string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotURI = r.RequestURI
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, `{}`)
			}))
			defer srv.Close()

			c := NewClientWithBaseURL(srv.URL+"/api/v5", "tok")
			_, _ = c.GetRepository(context.Background(), tt.owner, "repo")
			if gotURI != tt.wantRequestURI {
				t.Errorf("request URI = %q, want %q", gotURI, tt.wantRequestURI)
			}
		})
	}
}

func TestRawFilePathSegments(t *testing.T) {
	var gotURI string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURI = r.RequestURI
		fmt.Fprint(w, "raw")
	}))
	defer srv.Close()

	c := NewClientWithBaseURL(srv.URL+"/api/v5", "tok")
	if _, err := c.GetRawFile(context.Background(), "o", "r", "docs/READ ME.md", "main"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "/api/v5/repos/o/r/raw/docs/READ%20ME.md?ref=main"
	if gotURI != want {
		t.Errorf("request URI = %q, want %q", gotURI, want)
	}
}

// --- issue #7: response body limits ---

func TestResponseBodyLimit(t *testing.T) {
	bigBody := strings.Repeat("x", 1<<20) // 1 MiB, far over the small limits below

	tests := []struct {
		name    string
		limit   int64
		call    func(*Client) error
		wantBig bool
	}{
		{
			name:  "json endpoint over custom limit",
			limit: 1024,
			call: func(c *Client) error {
				_, err := c.GetRepository(context.Background(), "o", "r")
				return err
			},
			wantBig: true,
		},
		{
			name:  "raw download over custom limit",
			limit: 1024,
			call: func(c *Client) error {
				_, err := c.GetRawFile(context.Background(), "o", "r", "f", "main")
				return err
			},
			wantBig: true,
		},
		{
			name:  "upload response over custom limit",
			limit: 8,
			call: func(c *Client) error {
				_, err := c.UploadFileBytes(context.Background(), "o", "r", "f.txt", []byte("data"))
				return err
			},
			wantBig: true,
		},
		{
			name:  "raw download under limit passes",
			limit: int64(len(bigBody)) + 10,
			call: func(c *Client) error {
				b, err := c.GetRawFile(context.Background(), "o", "r", "f", "")
				if err == nil && len(b) != len(bigBody) {
					return fmt.Errorf("short read: %d", len(b))
				}
				return err
			},
			wantBig: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/octet-stream")
				fmt.Fprint(w, bigBody)
			}))
			defer srv.Close()

			c := NewClientWithBaseURL(srv.URL+"/api/v5", "tok")
			c.SetMaxResponseBodySize(tt.limit)
			err := tt.call(c)
			if tt.wantBig {
				if !IsResponseBodyTooLarge(err) {
					t.Fatalf("error = %v, want ResponseBodyTooLargeError", err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestMaxResponseBodySizeSetter(t *testing.T) {
	c := NewClient("tok")
	if got := c.MaxResponseBodySize(); got != 0 {
		t.Errorf("default limit = %d, want 0 (tiered defaults)", got)
	}
	c.SetMaxResponseBodySize(4096)
	if got := c.MaxResponseBodySize(); got != 4096 {
		t.Errorf("limit = %d, want 4096", got)
	}
	if got := c.jsonBodyLimit(); got != 4096 {
		t.Errorf("jsonBodyLimit = %d, want 4096", got)
	}
	if got := c.rawBodyLimit(); got != 4096 {
		t.Errorf("rawBodyLimit = %d, want 4096", got)
	}
	c.SetMaxResponseBodySize(0)
	if got := c.jsonBodyLimit(); got != defaultMaxJSONBodyBytes {
		t.Errorf("jsonBodyLimit after reset = %d, want %d", got, defaultMaxJSONBodyBytes)
	}
	if got := c.rawBodyLimit(); got != defaultMaxRawBodyBytes {
		t.Errorf("rawBodyLimit after reset = %d, want %d", got, defaultMaxRawBodyBytes)
	}
}

func TestReadBodyLimited(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		limit   int64
		wantErr bool
	}{
		{"under limit", "abc", 10, false},
		{"exactly at limit", "abcdefghij", 10, false},
		{"over limit", "abcdefghijk", 10, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := readBodyLimited(strings.NewReader(tt.input), tt.limit)
			if tt.wantErr {
				if !IsResponseBodyTooLarge(err) {
					t.Fatalf("error = %v, want ResponseBodyTooLargeError", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(b) != tt.input {
				t.Errorf("body = %q, want %q", b, tt.input)
			}
		})
	}
}

// --- issue #8: concurrent setters vs requests ---

// TestConcurrentSettersAndRequests must be run with -race: before the
// requestSnapshot fix, SetHTTPClient/SetBaseURL (locked writes) raced the
// bare reads of c.baseURL / c.httpClient on the request path.
func TestConcurrentSettersAndRequests(t *testing.T) {
	srv := httptest.NewServer(jsonHandler(http.StatusOK, map[string]string{"login": "u"}))
	defer srv.Close()

	c := NewClientWithBaseURL(srv.URL+"/api/v5", "tok")

	const rounds = 100
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := range rounds {
			c.SetHTTPClient(&http.Client{Timeout: 10 * time.Second})
			c.SetBaseURL(srv.URL + "/api/v5")
			_ = i
		}
	}()
	go func() {
		defer wg.Done()
		for range rounds {
			if _, err := c.GetCurrentUser(context.Background()); err != nil {
				t.Errorf("GetCurrentUser: %v", err)
				return
			}
		}
	}()
	wg.Wait()
}

// --- issue #9: uploads through the shared pipeline ---

func TestUploadRunsHooksAndRetries(t *testing.T) {
	var attempts, reqHooks, respHooks int
	var lastFileContent string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/v5/repos/o/r/file/upload") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, `{"error_msg":"boom"}`)
			return
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer file.Close()
		raw, err := io.ReadAll(file)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		lastFileContent = string(raw)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"file_path":"/uploads/a.txt"}`)
	}))
	defer srv.Close()

	c := NewClientWithBaseURL(srv.URL+"/api/v5", "tok")
	c.SetRetryPolicy(RetryPolicy{
		MaxRetries:           2,
		InitialBackoff:       time.Millisecond,
		MaxBackoff:           2 * time.Millisecond,
		RetryableStatusCodes: []int{http.StatusInternalServerError},
	})
	c.AddRequestHook(func(*http.Request) error { reqHooks++; return nil })
	c.AddResponseHook(func(*http.Response) error { respHooks++; return nil })

	res, err := c.UploadFileBytes(context.Background(), "o", "r", "a.txt", []byte("hello upload"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempts != 3 {
		t.Errorf("server attempts = %d, want 3 (1 + 2 retries with replayed body)", attempts)
	}
	if lastFileContent != "hello upload" {
		t.Errorf("uploaded content on final attempt = %q, want %q", lastFileContent, "hello upload")
	}
	if reqHooks != 1 {
		t.Errorf("request hook calls = %d, want 1", reqHooks)
	}
	if respHooks != 1 {
		t.Errorf("response hook calls = %d, want 1", respHooks)
	}
	if res.FilePath != "/uploads/a.txt" {
		t.Errorf("FilePath = %q, want /uploads/a.txt", res.FilePath)
	}
}

func TestUploadImageRunsHooks(t *testing.T) {
	var reqHooks int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqHooks++
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"file_path":"/img/x.png"}`)
	}))
	defer srv.Close()

	c := NewClientWithBaseURL(srv.URL+"/api/v5", "tok")
	c.AddRequestHook(func(*http.Request) error { return nil })
	hooked := 0
	c.AddRequestHook(func(*http.Request) error { hooked++; return nil })

	if _, err := c.UploadImageReader(context.Background(), "o", "r", "x.png", strings.NewReader("png")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hooked != 1 {
		t.Errorf("request hook calls = %d, want 1 (image upload used to bypass hooks entirely)", hooked)
	}
}

// TestJSONBodyReplayOnRetry guards the retry body reset for regular JSON
// POSTs: the retried request must carry the same body as the first.
func TestJSONBodyReplayOnRetry(t *testing.T) {
	var bodies []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(b)
		bodies = append(bodies, string(b))
		if len(bodies) < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"number":1}`)
	}))
	defer srv.Close()

	c := NewClientWithBaseURL(srv.URL+"/api/v5", "tok")
	c.SetRetryPolicy(RetryPolicy{
		MaxRetries:           2,
		InitialBackoff:       time.Millisecond,
		MaxBackoff:           2 * time.Millisecond,
		RetryableStatusCodes: []int{http.StatusInternalServerError},
	})
	if _, err := c.CreateIssue(context.Background(), "o", "r", CreateIssueOptions{Title: "t"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bodies) != 3 {
		t.Fatalf("attempts = %d, want 3", len(bodies))
	}
	for i, b := range bodies {
		if b == "" || b == "null" {
			t.Errorf("attempt %d sent an empty body: %q", i+1, b)
		}
		if b != bodies[0] {
			t.Errorf("attempt %d body %q differs from attempt 1 %q", i+1, b, bodies[0])
		}
	}
}

// --- issue #10: ListRepositories Type/Sort ---

func TestListRepositoriesFiltersOnWire(t *testing.T) {
	tests := []struct {
		name string
		call func(*Client) error
		// wantPath is the endpoint path prefix the request must hit.
		wantPath string
	}{
		{
			name: "ListRepositories authenticated user",
			call: func(c *Client) error {
				_, err := c.ListRepositories(context.Background(), ListRepositoriesOptions{Type: "owner", Sort: "created"})
				return err
			},
			wantPath: "/api/v5/user/repos",
		},
		{
			name: "ListRepositories owner branch",
			call: func(c *Client) error {
				_, err := c.ListRepositories(context.Background(), ListRepositoriesOptions{Owner: "someuser", Type: "owner", Sort: "created"})
				return err
			},
			wantPath: "/api/v5/users/someuser/repos",
		},
		{
			name: "ListCurrentUserRepositories",
			call: func(c *Client) error {
				_, err := c.ListCurrentUserRepositories(context.Background(), ListRepositoriesOptions{Type: "owner", Sort: "created"})
				return err
			},
			wantPath: "/api/v5/user/repos",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath string
			var gotQuery url.Values
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				gotQuery = r.URL.Query()
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, `[]`)
			}))
			defer srv.Close()

			c := NewClientWithBaseURL(srv.URL+"/api/v5", "tok")
			if err := tt.call(c); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotPath != tt.wantPath {
				t.Errorf("path = %q, want %q", gotPath, tt.wantPath)
			}
			if got := gotQuery.Get("type"); got != "owner" {
				t.Errorf("type filter = %q, want %q (silently dropped before issue #10 fix)", got, "owner")
			}
			if got := gotQuery.Get("sort"); got != "created" {
				t.Errorf("sort filter = %q, want %q", got, "created")
			}
			if got := gotQuery.Get("page"); got != "1" {
				t.Errorf("page = %q, want 1", got)
			}
			if got := gotQuery.Get("per_page"); got != "20" {
				t.Errorf("per_page = %q, want 20", got)
			}
		})
	}
}

func TestListRepositoriesOptionsNoFilters(t *testing.T) {
	var gotQuery url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[]`)
	}))
	defer srv.Close()

	c := NewClientWithBaseURL(srv.URL+"/api/v5", "tok")
	if _, err := c.ListRepositories(context.Background(), ListRepositoriesOptions{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := gotQuery["type"]; ok {
		t.Errorf("unexpected type filter: %v", gotQuery["type"])
	}
	if _, ok := gotQuery["sort"]; ok {
		t.Errorf("unexpected sort filter: %v", gotQuery["sort"])
	}
}

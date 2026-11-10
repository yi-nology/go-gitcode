package gitcode

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	DefaultBaseURL = "https://api.gitcode.com/api/v5"

	// EnvGitBaseURL is the environment variable for a custom GitCode API base URL.
	// Set this to your private instance URL (e.g. "https://gitcode.example.com/api/v5").
	EnvGitBaseURL = "GITCODE_URL"

	// EnvGitToken is the environment variable for the GitCode API token.
	EnvGitToken = "GITCODE_TOKEN"

	// defaultMaxJSONBodyBytes caps how much of a JSON API response is read
	// into memory. JSON payloads are small; anything past this means a
	// broken or hostile server, and refusing beats OOM-ing the process
	// (issue #7).
	defaultMaxJSONBodyBytes int64 = 16 << 20 // 16 MiB

	// defaultMaxRawBodyBytes caps binary downloads (raw files, release
	// assets, artifacts, log archives). These are legitimately large, so
	// the ceiling is far higher — but still a ceiling.
	defaultMaxRawBodyBytes int64 = 1 << 30 // 1 GiB
)

// ResponseBodyTooLargeError is returned when a response body exceeds the
// configured read limit instead of being read unbounded into memory.
type ResponseBodyTooLargeError struct {
	// Limit is the maximum number of bytes that would have been read.
	Limit int64
}

func (e *ResponseBodyTooLargeError) Error() string {
	return fmt.Sprintf("gitcode: response body exceeds %d-byte limit", e.Limit)
}

// IsResponseBodyTooLarge checks if an error is a ResponseBodyTooLargeError.
func IsResponseBodyTooLarge(err error) bool {
	var target *ResponseBodyTooLargeError
	return errors.As(err, &target)
}

// readBodyLimited reads r fully into memory but never more than limit
// bytes; past the limit it stops and reports ResponseBodyTooLargeError so a
// broken or hostile server cannot trigger unbounded allocation (issue #7).
func readBodyLimited(r io.Reader, limit int64) ([]byte, error) {
	b, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return b, err
	}
	if int64(len(b)) > limit {
		return b, &ResponseBodyTooLargeError{Limit: limit}
	}
	return b, nil
}

// Client is the GitCode API client.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
	authStyle  AuthStyle
	mu         sync.RWMutex

	// maxBodyBytes overrides the per-response body read limit for every
	// request; 0 means "use the tiered defaults" (defaultMaxJSONBodyBytes
	// for JSON endpoints, defaultMaxRawBodyBytes for binary downloads).
	maxBodyBytes int64

	// Retry policy for failed requests.
	retryPolicy *RetryPolicy

	// Request/response hooks for middleware.
	requestHooks  []RequestHook
	responseHooks []ResponseHook
}

// AuthStyle defines how authentication tokens are sent.
type AuthStyle int

const (
	// AuthStyleBearer sends "Authorization: Bearer <token>" header.
	AuthStyleBearer AuthStyle = iota
	// AuthStylePrivateToken sends "PRIVATE-TOKEN: <token>" header.
	AuthStylePrivateToken
	// AuthStyleAccessToken sends "?access_token=<token>" query parameter.
	AuthStyleAccessToken
)

// NewClient creates a new GitCode API client with the given token.
func NewClient(token string) *Client {
	return &Client{
		baseURL: DefaultBaseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		authStyle: AuthStyleBearer,
	}
}

// NewClientWithBaseURL creates a new GitCode API client with a custom base URL.
func NewClientWithBaseURL(baseURL, token string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		authStyle: AuthStyleBearer,
	}
}

// NewClientFromEnv creates a client from GITCODE_TOKEN and optionally GITCODE_URL
// environment variables. Returns nil if GITCODE_TOKEN is not set.
//
// This is the recommended way to create a client for CLI tools and scripts that
// support both public and private GitCode instances via environment configuration.
func NewClientFromEnv() *Client {
	token := os.Getenv(EnvGitToken)
	if token == "" {
		return nil
	}
	if baseURL := os.Getenv(EnvGitBaseURL); baseURL != "" {
		return NewClientWithBaseURL(baseURL, token)
	}
	return NewClient(token)
}

// SetAuthStyle sets the authentication style for the client.
func (c *Client) SetAuthStyle(style AuthStyle) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.authStyle = style
}

// SetHTTPClient sets a custom HTTP client.
func (c *Client) SetHTTPClient(client *http.Client) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.httpClient = client
}

// SetBaseURL sets a custom base URL for the client.
// This is useful for switching to a private GitCode deployment after construction.
func (c *Client) SetBaseURL(baseURL string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.baseURL = strings.TrimRight(baseURL, "/")
}

// BaseURL returns the current base URL of the client.
func (c *Client) BaseURL() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.baseURL
}

// SetMaxResponseBodySize caps how many bytes of any response body are read
// into memory. Values <= 0 restore the tiered defaults (16 MiB for JSON
// endpoints, 1 GiB for binary downloads). The limit is enforced on every
// request path; exceeding it fails with ResponseBodyTooLargeError (issue #7).
func (c *Client) SetMaxResponseBodySize(n int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if n <= 0 {
		n = 0
	}
	c.maxBodyBytes = n
}

// MaxResponseBodySize returns the configured response body limit; 0 means
// the tiered defaults are in effect.
func (c *Client) MaxResponseBodySize() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.maxBodyBytes
}

// jsonBodyLimit is the read limit for JSON API responses.
func (c *Client) jsonBodyLimit() int64 {
	if n := c.MaxResponseBodySize(); n > 0 {
		return n
	}
	return defaultMaxJSONBodyBytes
}

// rawBodyLimit is the read limit for binary download responses.
func (c *Client) rawBodyLimit() int64 {
	if n := c.MaxResponseBodySize(); n > 0 {
		return n
	}
	return defaultMaxRawBodyBytes
}

// requestSnapshot captures the request-time mutable settings under the read
// lock. Every read of baseURL/httpClient on the request path goes through
// it so that concurrent SetHTTPClient/SetBaseURL calls cannot race the
// readers — the setters write under the same mutex, and until now the
// readers did not hold it (issue #8).
func (c *Client) requestSnapshot() (base string, hc *http.Client) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.baseURL, c.httpClient
}

// setAuthHeader sets the authentication header on the request.
func (c *Client) setAuthHeader(req *http.Request) {
	c.mu.RLock()
	style := c.authStyle
	token := c.token
	c.mu.RUnlock()

	switch style {
	case AuthStylePrivateToken:
		req.Header.Set("PRIVATE-TOKEN", token)
	case AuthStyleAccessToken:
		q := req.URL.Query()
		q.Set("access_token", token)
		req.URL.RawQuery = q.Encode()
	default:
		req.Header.Set("Authorization", "Bearer "+token)
	}
}

// doRequest executes an HTTP request and unmarshals the result.
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	base, _ := c.requestSnapshot()
	return c.doRequestAt(ctx, base, method, path, body, result)
}

// doRequestAt executes an HTTP request against an explicit base URL and
// unmarshals the result — the shared execution core behind the v5 API
// (doRequest) and the v8 API (actions_v8.go), whose paths differ only by
// the base.
func (c *Client) doRequestAt(ctx context.Context, base, method, path string, body interface{}, result interface{}) error {
	var raw []byte
	if body != nil {
		var err error
		raw, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}
	return c.doBufferedRequestAt(ctx, base, method, path, "application/json", raw, result, c.jsonBodyLimit())
}

// doBufferedRequestAt is the single shared execution pipeline:
// auth header → request hooks → retry loop → response hooks →
// size-limited body read → status check → JSON decode.
//
// Every request shape — JSON bodies, form encoding, multipart uploads —
// funnels through it, so request/response hooks and the retry policy apply
// uniformly (issue #9: uploads used to bypass both). body may be nil (no
// request body); limit caps the response body read (issue #7).
func (c *Client) doBufferedRequestAt(ctx context.Context, base, method, path, contentType string, body []byte, result interface{}, limit int64) error {
	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, base+path, reqBody)
	if err != nil {
		return err
	}

	c.setAuthHeader(req)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	// Run request hooks
	if err := c.runRequestHooks(req); err != nil {
		return fmt.Errorf("request hook error: %w", err)
	}

	// Execute request with retry
	resp, err := c.doRequestWithRetry(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Run response hooks
	if err := c.runResponseHooks(resp); err != nil {
		return fmt.Errorf("response hook error: %w", err)
	}

	respBody, err := readBodyLimited(resp.Body, limit)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return newAPIError(method, path, resp.StatusCode, string(respBody))
	}

	if result != nil && resp.StatusCode != http.StatusNoContent && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return err
		}
	}

	return nil
}

// doRawRequestAt executes an HTTP request against an explicit base URL and
// returns the raw response body (binary downloads: artifacts, log archives).
func (c *Client) doRawRequestAt(ctx context.Context, base, method, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, base+path, nil)
	if err != nil {
		return nil, err
	}
	return c.doRawRequestWith(ctx, req, method, path)
}

// doRawRequest executes a raw HTTP request and returns the response body.
func (c *Client) doRawRequest(ctx context.Context, method, path string) ([]byte, error) {
	base, _ := c.requestSnapshot()
	req, err := http.NewRequestWithContext(ctx, method, base+path, nil)
	if err != nil {
		return nil, err
	}
	return c.doRawRequestWith(ctx, req, method, path)
}

// doRawRequestWith runs an already-built request through the shared
// pipeline (hooks + retry) and returns the raw, size-limited body.
func (c *Client) doRawRequestWith(ctx context.Context, req *http.Request, method, path string) ([]byte, error) {
	c.setAuthHeader(req)
	if err := c.runRequestHooks(req); err != nil {
		return nil, fmt.Errorf("request hook error: %w", err)
	}
	resp, err := c.doRequestWithRetry(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := c.runResponseHooks(resp); err != nil {
		return nil, fmt.Errorf("response hook error: %w", err)
	}
	body, err := readBodyLimited(resp.Body, c.rawBodyLimit())
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, newAPIError(method, path, resp.StatusCode, string(body[:min(len(body), 200)]))
	}
	return body, nil
}

// doFormRequest executes a form-encoded HTTP request.
func (c *Client) doFormRequest(ctx context.Context, method, path string, params url.Values, result interface{}) error {
	var raw []byte
	if params != nil {
		raw = []byte(params.Encode())
	}
	base, _ := c.requestSnapshot()
	return c.doBufferedRequestAt(ctx, base, method, path, "application/x-www-form-urlencoded", raw, result, c.jsonBodyLimit())
}

// doRawBodyRequest executes an HTTP request with a raw JSON body.
func (c *Client) doRawBodyRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	base, _ := c.requestSnapshot()
	return c.doBufferedRequestAt(ctx, base, method, path, "application/json", raw, result, c.jsonBodyLimit())
}

// isNotFoundError checks if an error is a 404 Not Found response.
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	if IsNotFound(err) {
		return true
	}
	return contains(err.Error(), "404")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// User represents a GitCode user.
type User struct {
	ID        FlexString `json:"id"`
	Login     string     `json:"login"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	AvatarURL string     `json:"avatar_url"`
	HTMLURL   string     `json:"html_url,omitempty"`
	Type      string     `json:"type,omitempty"`
}

// GetCurrentUser returns the authenticated user's profile.
func (c *Client) GetCurrentUser(ctx context.Context) (*User, error) {
	var user User
	err := c.doRequest(ctx, http.MethodGet, "/user", nil, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUser returns a user by username.
func (c *Client) GetUser(ctx context.Context, username string) (*User, error) {
	var user User
	err := c.doRequest(ctx, http.MethodGet, pathf("/users/%s", username), nil, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ListOptions contains pagination parameters for list requests.
type ListOptions struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

func (o ListOptions) withDefaults() ListOptions {
	if o.Page == 0 {
		o.Page = 1
	}
	if o.PerPage == 0 {
		o.PerPage = 20
	}
	return o
}

func (o ListOptions) toQuery() string {
	opts := o.withDefaults()
	return fmt.Sprintf("page=%d&per_page=%d", opts.Page, opts.PerPage)
}

// RateLimit contains rate limit information.
type RateLimit struct {
	Limit     int       `json:"limit"`
	Remaining int       `json:"remaining"`
	Reset     time.Time `json:"reset"`
}

// GetRateLimit returns the current rate limit status.
func (c *Client) GetRateLimit(ctx context.Context) (*RateLimit, error) {
	var rl RateLimit
	err := c.doRequest(ctx, http.MethodGet, "/rate_limit", nil, &rl)
	if err != nil {
		return nil, err
	}
	return &rl, nil
}

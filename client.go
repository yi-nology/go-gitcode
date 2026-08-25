package gitcode

import (
	"bytes"
	"context"
	"encoding/json"
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
)

// Client is the GitCode API client.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
	authStyle  AuthStyle
	mu         sync.RWMutex

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
	_, err := c.doRequestWithHeaders(ctx, method, path, body, result)
	return err
}

// doRequestWithHeaders executes an HTTP request and returns the response headers.
func (c *Client) doRequestWithHeaders(ctx context.Context, method, path string, body interface{}, result interface{}) (http.Header, error) {
	var reqBody io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, err
	}

	c.setAuthHeader(req)
	req.Header.Set("Content-Type", "application/json")

	// Run request hooks
	if err := c.runRequestHooks(req); err != nil {
		return nil, fmt.Errorf("request hook error: %w", err)
	}

	// Execute request with retry
	resp, err := c.doRequestWithRetry(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Run response hooks
	if err := c.runResponseHooks(resp); err != nil {
		return nil, fmt.Errorf("response hook error: %w", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, newAPIError(method, path, resp.StatusCode, string(respBody))
	}

	if result != nil && resp.StatusCode != http.StatusNoContent {
		if len(respBody) > 0 {
			if err := json.Unmarshal(respBody, result); err != nil {
				return nil, err
			}
		}
	}

	return resp.Header, nil
}

// doRawRequest executes a raw HTTP request and returns the response body.
func (c *Client) doRawRequest(ctx context.Context, method, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}

	c.setAuthHeader(req)

	// Run request hooks
	if err := c.runRequestHooks(req); err != nil {
		return nil, fmt.Errorf("request hook error: %w", err)
	}

	// Execute request with retry
	resp, err := c.doRequestWithRetry(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Run response hooks
	if err := c.runResponseHooks(resp); err != nil {
		return nil, fmt.Errorf("response hook error: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
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
	var reqBody io.Reader
	if params != nil {
		reqBody = strings.NewReader(params.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return err
	}

	c.setAuthHeader(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return newAPIError(method, path, resp.StatusCode, string(respBody))
	}

	if result != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.Unmarshal(respBody, result); err != nil {
			return err
		}
	}

	return nil
}

// doRawBodyRequest executes an HTTP request with a raw JSON body.
func (c *Client) doRawBodyRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(raw))
	if err != nil {
		return err
	}

	c.setAuthHeader(req)
	req.Header.Set("Content-Type", "application/json")

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

	respBody, err := io.ReadAll(resp.Body)
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/users/%s", username), nil, &user)
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

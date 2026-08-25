package gitcode

import (
	"net/http"
)

// RequestHook is a function called before each HTTP request.
// It can modify the request or return an error to cancel the request.
type RequestHook func(req *http.Request) error

// ResponseHook is a function called after each HTTP response.
type ResponseHook func(resp *http.Response) error

// AddRequestHook adds a hook that runs before each HTTP request.
func (c *Client) AddRequestHook(hook RequestHook) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.requestHooks = append(c.requestHooks, hook)
}

// AddResponseHook adds a hook that runs after each HTTP response.
func (c *Client) AddResponseHook(hook ResponseHook) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.responseHooks = append(c.responseHooks, hook)
}

// ClearHooks removes all request and response hooks.
func (c *Client) ClearHooks() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.requestHooks = nil
	c.responseHooks = nil
}

// runRequestHooks executes all registered request hooks.
func (c *Client) runRequestHooks(req *http.Request) error {
	c.mu.RLock()
	hooks := make([]RequestHook, len(c.requestHooks))
	copy(hooks, c.requestHooks)
	c.mu.RUnlock()

	for _, hook := range hooks {
		if err := hook(req); err != nil {
			return err
		}
	}
	return nil
}

// runResponseHooks executes all registered response hooks.
func (c *Client) runResponseHooks(resp *http.Response) error {
	c.mu.RLock()
	hooks := make([]ResponseHook, len(c.responseHooks))
	copy(hooks, c.responseHooks)
	c.mu.RUnlock()

	for _, hook := range hooks {
		if err := hook(resp); err != nil {
			return err
		}
	}
	return nil
}

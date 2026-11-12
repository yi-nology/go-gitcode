package gitcode

import (
	"context"
	"math"
	"net/http"
	"strconv"
	"time"
)

// RetryPolicy defines the retry behavior for failed requests.
type RetryPolicy struct {
	// MaxRetries is the maximum number of retry attempts. Default: 0 (no retry).
	MaxRetries int
	// InitialBackoff is the initial wait time before the first retry. Default: 1s.
	InitialBackoff time.Duration
	// MaxBackoff is the maximum wait time between retries. Default: 30s.
	// The clamp applies to the exponential backoff **and** to a server-sent
	// Retry-After: a hostile or misbehaving server cannot park the client
	// for an arbitrary amount of time (issue #4).
	MaxBackoff time.Duration
	// Multiplier is the backoff multiplier for exponential backoff. Default: 2.0.
	Multiplier float64
	// RetryableStatusCodes defines which HTTP status codes should trigger a retry.
	// Default: [429, 500, 502, 503, 504]
	RetryableStatusCodes []int
}

// DefaultRetryPolicy returns a sensible default retry policy.
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxRetries:           3,
		InitialBackoff:       1 * time.Second,
		MaxBackoff:           30 * time.Second,
		Multiplier:           2.0,
		RetryableStatusCodes: []int{429, 500, 502, 503, 504},
	}
}

// SetRetryPolicy sets the retry policy for the client.
func (c *Client) SetRetryPolicy(policy RetryPolicy) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.retryPolicy = &policy
}

// shouldRetry checks if a request should be retried based on the status code.
func (c *Client) shouldRetry(statusCode int) bool {
	c.mu.RLock()
	policy := c.retryPolicy
	c.mu.RUnlock()

	if policy == nil {
		return false
	}

	for _, code := range policy.RetryableStatusCodes {
		if code == statusCode {
			return true
		}
	}
	return false
}

// defaultInitialBackoff / defaultMaxBackoff / defaultMultiplier fill in the
// documented defaults when a custom policy leaves the corresponding field
// unset (issue #4: a zero MaxBackoff used to clamp every wait to 0;
// issue #14: a zero Multiplier made math.Pow(0, n>0) = 0 — every retry
// after the first had zero backoff, the busy-wait pathology #4 fixed
// re-entering through another path).
const (
	defaultInitialBackoff = 1 * time.Second
	defaultMaxBackoff     = 30 * time.Second
	defaultMultiplier     = 2.0
)

// maxRetryAfterSeconds bounds delay-seconds before the Duration conversion
// so a huge Retry-After cannot overflow time.Duration.
const maxRetryAfterSeconds = math.MaxInt64 / int64(time.Second)

// retryAfterDuration parses the Retry-After header (issue #4). It accepts
// both RFC 9110 forms — delay-seconds and HTTP-date — and applies three
// gates before the value is used:
//
//  1. non-positive values (negative or zero seconds, a date in the past)
//     are rejected instead of producing a negative duration, which used to
//     make time.After fire immediately and turn rate-limit backoff into a
//     busy loop;
//  2. the parsed value is clamped by the caller to MaxBackoff;
//  3. unparseable values report ok=false and the caller falls back to
//     exponential backoff.
func retryAfterDuration(resp *http.Response, now time.Time) (time.Duration, bool) {
	if resp == nil {
		return 0, false
	}
	v := resp.Header.Get("Retry-After")
	if v == "" {
		return 0, false
	}
	if seconds, err := strconv.Atoi(v); err == nil {
		if seconds <= 0 {
			return 0, false
		}
		if int64(seconds) > maxRetryAfterSeconds {
			return time.Duration(math.MaxInt64), true
		}
		return time.Duration(seconds) * time.Second, true
	}
	// HTTP-date form: wait until the given instant.
	if t, err := http.ParseTime(v); err == nil {
		if d := t.Sub(now); d > 0 {
			return d, true
		}
		return 0, false
	}
	return 0, false
}

// getRetryWaitTime calculates the wait time for a given retry attempt.
func (c *Client) getRetryWaitTime(attempt int, resp *http.Response) time.Duration {
	c.mu.RLock()
	policy := c.retryPolicy
	c.mu.RUnlock()

	if policy == nil {
		return 0
	}

	initial := policy.InitialBackoff
	if initial <= 0 {
		initial = defaultInitialBackoff
	}
	maxBackoff := policy.MaxBackoff
	if maxBackoff <= 0 {
		maxBackoff = defaultMaxBackoff
	}
	multiplier := policy.Multiplier
	if multiplier <= 0 {
		multiplier = defaultMultiplier // 1.0（恒定退避）是合法显式选择，仅归一未填的零值
	}

	// Check Retry-After header first (for 429 responses): a valid
	// server-provided delay wins over the exponential curve, but shares
	// the same MaxBackoff ceiling (issue #4).
	if d, ok := retryAfterDuration(resp, time.Now()); ok {
		return min(d, maxBackoff)
	}

	// Exponential backoff
	backoff := time.Duration(float64(initial) * math.Pow(multiplier, float64(attempt)))
	return min(backoff, maxBackoff)
}

// doRequestWithRetry executes an HTTP request with retry logic.
func (c *Client) doRequestWithRetry(ctx context.Context, req *http.Request) (*http.Response, error) {
	c.mu.RLock()
	policy := c.retryPolicy
	c.mu.RUnlock()

	maxRetries := 0
	if policy != nil {
		maxRetries = policy.MaxRetries
	}

	// The HTTP client is read under the same lock the setters write under;
	// a bare read here raced SetHTTPClient (issue #8).
	_, hc := c.requestSnapshot()

	var resp *http.Response
	var err error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Replay the request body: the previous attempt consumed it.
			// http.NewRequestWithContext wires GetBody for *bytes.Reader /
			// *strings.Reader bodies, which is exactly what every request
			// builder in this package produces.
			if req.GetBody != nil {
				if body, bodyErr := req.GetBody(); bodyErr == nil {
					req.Body = body
				}
			}
			waitTime := c.getRetryWaitTime(attempt-1, resp)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(waitTime):
			}
		}

		resp, err = hc.Do(req)
		if err != nil {
			return nil, err
		}

		// Check if we should retry
		if c.shouldRetry(resp.StatusCode) && attempt < maxRetries {
			resp.Body.Close()
			continue
		}

		return resp, nil
	}

	return resp, nil
}

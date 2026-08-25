package gitcode

import (
	"errors"
	"fmt"
	"net/http"
)

// APIError represents a structured GitCode API error.
type APIError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	URL        string `json:"url,omitempty"`
	Method     string `json:"method,omitempty"`
	Body       string `json:"body,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("GitCode API %s %s returned %d: %s", e.Method, e.URL, e.StatusCode, e.Message)
}

// NotFoundError represents a 404 Not Found error.
type NotFoundError struct {
	*APIError
	Resource string `json:"resource,omitempty"` // e.g. "repository", "issue", "user"
}

func (e *NotFoundError) Error() string {
	if e.Resource != "" {
		return fmt.Sprintf("GitCode API: %s not found (%s %s)", e.Resource, e.Method, e.URL)
	}
	return e.APIError.Error()
}

// UnauthorizedError represents a 401 Unauthorized error.
type UnauthorizedError struct {
	*APIError
}

func (e *UnauthorizedError) Error() string {
	return fmt.Sprintf("GitCode API: unauthorized (%s %s) - check your token", e.Method, e.URL)
}

// ForbiddenError represents a 403 Forbidden error.
type ForbiddenError struct {
	*APIError
}

func (e *ForbiddenError) Error() string {
	return fmt.Sprintf("GitCode API: forbidden (%s %s) - insufficient permissions", e.Method, e.URL)
}

// RateLimitError represents a 429 Too Many Requests error.
type RateLimitError struct {
	*APIError
	RetryAfter int `json:"retry_after,omitempty"` // seconds to wait
}

func (e *RateLimitError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("GitCode API: rate limited (%s %s) - retry after %d seconds", e.Method, e.URL, e.RetryAfter)
	}
	return fmt.Sprintf("GitCode API: rate limited (%s %s)", e.Method, e.URL)
}

// ConflictError represents a 409 Conflict error.
type ConflictError struct {
	*APIError
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("GitCode API: conflict (%s %s) - resource already exists", e.Method, e.URL)
}

// ValidationError represents a 422 Unprocessable Entity error.
type ValidationError struct {
	*APIError
	Errors []FieldError `json:"errors,omitempty"`
}

func (e *ValidationError) Error() string {
	if len(e.Errors) > 0 {
		return fmt.Sprintf("GitCode API: validation failed (%s %s): %s", e.Method, e.URL, e.Errors[0].Message)
	}
	return e.APIError.Error()
}

// FieldError represents a field-level validation error.
type FieldError struct {
	Resource string `json:"resource"`
	Field    string `json:"field"`
	Code     string `json:"code"`
	Message  string `json:"message,omitempty"`
}

// IsNotFound checks if an error is a NotFoundError.
func IsNotFound(err error) bool {
	var target *NotFoundError
	return errors.As(err, &target)
}

// IsUnauthorized checks if an error is an UnauthorizedError.
func IsUnauthorized(err error) bool {
	var target *UnauthorizedError
	return errors.As(err, &target)
}

// IsRateLimit checks if an error is a RateLimitError.
func IsRateLimit(err error) bool {
	var target *RateLimitError
	return errors.As(err, &target)
}

// IsConflict checks if an error is a ConflictError.
func IsConflict(err error) bool {
	var target *ConflictError
	return errors.As(err, &target)
}

// IsForbidden checks if an error is a ForbiddenError.
func IsForbidden(err error) bool {
	var target *ForbiddenError
	return errors.As(err, &target)
}

// IsValidation checks if an error is a ValidationError.
func IsValidation(err error) bool {
	var target *ValidationError
	return errors.As(err, &target)
}

// newAPIError creates a structured error from an HTTP response.
func newAPIError(method, url string, statusCode int, body string) error {
	apiErr := &APIError{
		StatusCode: statusCode,
		Message:    body,
		URL:        url,
		Method:     method,
		Body:       body,
	}

	switch statusCode {
	case http.StatusNotFound:
		return &NotFoundError{APIError: apiErr}
	case http.StatusUnauthorized:
		return &UnauthorizedError{APIError: apiErr}
	case http.StatusForbidden:
		return &ForbiddenError{APIError: apiErr}
	case http.StatusTooManyRequests:
		return &RateLimitError{APIError: apiErr}
	case http.StatusConflict:
		return &ConflictError{APIError: apiErr}
	case http.StatusUnprocessableEntity:
		return &ValidationError{APIError: apiErr}
	default:
		return apiErr
	}
}

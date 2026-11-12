package gitcode

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStructuredErrors(t *testing.T) {
	// Test NotFoundError
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"404 Not Found"}`))
	}))
	defer server.Close()

	client := NewClientWithBaseURL(server.URL+"/api/v5", "test-token")
	_, err := client.GetRepository(context.Background(), "owner", "repo")
	if err == nil {
		t.Fatal("expected error")
	}

	if !IsNotFound(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}

	// Test UnauthorizedError
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"401 Unauthorized"}`))
	}))
	defer server2.Close()

	client2 := NewClientWithBaseURL(server2.URL+"/api/v5", "bad-token")
	_, err = client2.GetCurrentUser(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}

	if !IsUnauthorized(err) {
		t.Errorf("expected UnauthorizedError, got %T: %v", err, err)
	}

	// Test RateLimitError
	server3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"message":"rate limit exceeded"}`))
	}))
	defer server3.Close()

	client3 := NewClientWithBaseURL(server3.URL+"/api/v5", "test-token")
	_, err = client3.GetCurrentUser(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}

	if !IsRateLimit(err) {
		t.Errorf("expected RateLimitError, got %T: %v", err, err)
	}

	// Test ConflictError
	server4 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"message":"already exists"}`))
	}))
	defer server4.Close()

	client4 := NewClientWithBaseURL(server4.URL+"/api/v5", "test-token")
	_, err = client4.CreateRepository(context.Background(), CreateRepositoryOptions{Name: "existing"})
	if err == nil {
		t.Fatal("expected error")
	}

	if !IsConflict(err) {
		t.Errorf("expected ConflictError, got %T: %v", err, err)
	}
}

func TestRetryPolicy(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"server error"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"login":"test","name":"Test User"}`))
	}))
	defer server.Close()

	client := NewClientWithBaseURL(server.URL+"/api/v5", "test-token")
	client.SetRetryPolicy(RetryPolicy{
		MaxRetries:           3,
		InitialBackoff:       1, // 1ms for fast test
		MaxBackoff:           10,
		Multiplier:           1.0,
		RetryableStatusCodes: []int{500},
	})

	user, err := client.GetCurrentUser(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.Login != "test" {
		t.Errorf("expected login 'test', got '%s'", user.Login)
	}

	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestRequestHooks(t *testing.T) {
	var hookCalled bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"login":"test","name":"Test"}`))
	}))
	defer server.Close()

	client := NewClientWithBaseURL(server.URL+"/api/v5", "test-token")
	client.AddRequestHook(func(req *http.Request) error {
		hookCalled = true
		req.Header.Set("X-Custom", "test-value")
		return nil
	})

	_, err := client.GetCurrentUser(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !hookCalled {
		t.Error("request hook was not called")
	}
}

func TestRequestHookError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClientWithBaseURL(server.URL+"/api/v5", "test-token")
	client.AddRequestHook(func(req *http.Request) error {
		return errors.New("hook rejected request")
	})

	_, err := client.GetCurrentUser(context.Background())
	if err == nil {
		t.Fatal("expected error from hook")
	}

	if !strings.Contains(err.Error(), "hook rejected request") {
		t.Errorf("expected hook error, got: %v", err)
	}
}

func TestResponseHooks(t *testing.T) {
	var hookCalled bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"login":"test","name":"Test"}`))
	}))
	defer server.Close()

	client := NewClientWithBaseURL(server.URL+"/api/v5", "test-token")
	client.AddResponseHook(func(resp *http.Response) error {
		hookCalled = true
		return nil
	})

	_, err := client.GetCurrentUser(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !hookCalled {
		t.Error("response hook was not called")
	}
}

func TestConcurrentSafety(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"login":"test","name":"Test"}`))
	}))
	defer server.Close()

	client := NewClientWithBaseURL(server.URL+"/api/v5", "test-token")

	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			// Concurrent SetAuthStyle and SetHTTPClient
			client.SetAuthStyle(AuthStyleBearer)
			client.SetHTTPClient(&http.Client{})
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestWebhookSignatureVerification(t *testing.T) {
	payload := []byte(`{"action":"opened","number":1}`)
	secret := "test-secret"

	// Compute signature
	sig := ComputeWebhookSignature(payload, secret)

	// Verify signature
	if !VerifyWebhookSignature(payload, secret, sig) {
		t.Error("valid signature should pass verification")
	}

	// Verify with wrong secret
	if VerifyWebhookSignature(payload, "wrong-secret", sig) {
		t.Error("wrong secret should fail verification")
	}

	// Verify with wrong payload
	if VerifyWebhookSignature([]byte("wrong payload"), secret, sig) {
		t.Error("wrong payload should fail verification")
	}

	// Verify with empty inputs
	if VerifyWebhookSignature(payload, "", sig) {
		t.Error("empty secret should fail verification")
	}

	if VerifyWebhookSignature(payload, secret, "") {
		t.Error("empty signature should fail verification")
	}
}

func TestValidation(t *testing.T) {
	err := validateOwnerRepo("", "repo")
	if err == nil {
		t.Error("expected validation error for empty owner")
	}

	err = validateOwnerRepo("owner", "")
	if err == nil {
		t.Error("expected validation error for empty repo")
	}

	err = validateOwnerRepo("owner", "repo")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = validateNonEmpty("field", "")
	if err == nil {
		t.Error("expected validation error for empty field")
	}

	if !IsValidationError(validateOwnerRepo("", "repo")) {
		t.Error("expected IsValidationError to return true")
	}
}

func TestCollectAll(t *testing.T) {
	// Simulate a list that has 2 full pages then an empty page
	fetch := func(opts ListOptions) ([]string, error) {
		if opts.Page > 2 {
			return nil, nil // nil signals end (len == 0)
		}
		// Return perPage items to simulate full pages
		items := make([]string, opts.PerPage)
		for i := range items {
			items[i] = "item"
		}
		return items, nil
	}

	all, err := CollectAll(context.Background(), fetch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(all) != 200 { // 2 pages * 100 per page
		t.Errorf("expected 200 items, got %d", len(all))
	}
}

func TestClearHooks(t *testing.T) {
	client := NewClient("test")
	client.AddRequestHook(func(req *http.Request) error { return nil })
	client.AddResponseHook(func(resp *http.Response) error { return nil })

	client.ClearHooks()

	if len(client.requestHooks) != 0 {
		t.Errorf("expected 0 request hooks, got %d", len(client.requestHooks))
	}
	if len(client.responseHooks) != 0 {
		t.Errorf("expected 0 response hooks, got %d", len(client.responseHooks))
	}
}

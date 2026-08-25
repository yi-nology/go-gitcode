package gitcode

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestListCommitStatuses(t *testing.T) {
	client, server := newTestServer(jsonResponse(200, []*CommitStatus{
		{ID: 1, State: "success", Context: "ci/build"},
		{ID: 2, State: "pending", Context: "ci/test"},
	}))
	defer server.Close()

	statuses, err := client.ListCommitStatuses(context.Background(), "owner", "repo", "abc123", ListOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(statuses) != 2 {
		t.Errorf("expected 2 statuses, got %d", len(statuses))
	}
}

func TestCreateCommitStatus(t *testing.T) {
	var receivedBody CreateCommitStatusOptions
	client, server := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedBody)
		json.NewEncoder(w).Encode(&CommitStatus{ID: 1, State: "success"})
	}))
	defer server.Close()

	status, err := client.CreateCommitStatus(context.Background(), "owner", "repo", "abc123", CreateCommitStatusOptions{
		State:       "success",
		TargetURL:   "https://ci.example.com",
		Description: "Build passed",
		Context:     "ci/build",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.State != "success" {
		t.Errorf("expected state 'success', got '%s'", status.State)
	}
}

func TestGetCombinedStatus(t *testing.T) {
	client, server := newTestServer(jsonResponse(200, &CombinedStatus{
		SHA:        "abc123",
		TotalCount: 2,
		Statuses: []*CommitStatus{
			{State: "success", Context: "ci/build"},
			{State: "pending", Context: "ci/test"},
		},
	}))
	defer server.Close()

	combined, err := client.GetCombinedStatus(context.Background(), "owner", "repo", "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if combined.TotalCount != 2 {
		t.Errorf("expected total count 2, got %d", combined.TotalCount)
	}
}

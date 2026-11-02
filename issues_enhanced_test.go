package gitcode

import (
	"context"
	"testing"
)

func TestListIssueRelatedBranches(t *testing.T) {
	client, server := newTestServer(jsonResponse(200, []*IssueRelatedBranch{
		{BranchName: "feature-branch"},
	}))
	defer server.Close()

	branches, err := client.ListIssueRelatedBranches(context.Background(), "owner", "repo", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(branches) != 1 {
		t.Errorf("expected 1 branch, got %d", len(branches))
	}
}

func TestSetIssueRelatedBranches(t *testing.T) {
	client, server := newTestServer(jsonResponse(200, nil))
	defer server.Close()

	err := client.SetIssueRelatedBranches(context.Background(), "owner", "repo", 1, []string{"feature-branch"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListIssueModifyHistory(t *testing.T) {
	client, server := newTestServer(jsonResponse(200, []*ModifyHistoryEntry{
		{ID: 1, Action: "update", Field: "title", OldValue: "old", NewValue: "new"},
	}))
	defer server.Close()

	history, err := client.ListIssueModifyHistory(context.Background(), "owner", "repo", 1, ListOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(history) != 1 {
		t.Errorf("expected 1 history entry, got %d", len(history))
	}
}

func TestListIssueUserReactions(t *testing.T) {
	client, server := newTestServer(jsonResponse(200, []*IssueUserReaction{
		{ID: 1, Content: "+1"},
	}))
	defer server.Close()

	reactions, err := client.ListIssueUserReactions(context.Background(), "owner", "repo", 1, ListOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reactions) != 1 {
		t.Errorf("expected 1 reaction, got %d", len(reactions))
	}
}

package gitcode

import (
	"context"
	"testing"
)

func TestListDiscussions(t *testing.T) {
	client, server := newTestServer(jsonResponse(200, []*Discussion{
		{ID: 1, Number: 1, Title: "Discussion 1"},
	}))
	defer server.Close()

	discussions, err := client.ListDiscussions(context.Background(), "owner", "repo", ListOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(discussions) != 1 {
		t.Errorf("expected 1 discussion, got %d", len(discussions))
	}
}

func TestGetDiscussion(t *testing.T) {
	client, server := newTestServer(jsonResponse(200, &Discussion{
		ID: 1, Number: 1, Title: "Test Discussion",
	}))
	defer server.Close()

	disc, err := client.GetDiscussion(context.Background(), "owner", "repo", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if disc.Title != "Test Discussion" {
		t.Errorf("expected title 'Test Discussion', got '%s'", disc.Title)
	}
}

func TestListDiscussionComments(t *testing.T) {
	client, server := newTestServer(jsonResponse(200, []*DiscussionComment{
		{ID: 1, Body: "Comment 1"},
	}))
	defer server.Close()

	comments, err := client.ListDiscussionComments(context.Background(), "owner", "repo", 1, ListOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(comments) != 1 {
		t.Errorf("expected 1 comment, got %d", len(comments))
	}
}

func TestListDiscussionCommentReplies(t *testing.T) {
	client, server := newTestServer(jsonResponse(200, []*DiscussionCommentReply{
		{ID: 1, Body: "Reply 1"},
	}))
	defer server.Close()

	replies, err := client.ListDiscussionCommentReplies(context.Background(), "owner", "repo", 1, 100, ListOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(replies) != 1 {
		t.Errorf("expected 1 reply, got %d", len(replies))
	}
}

func TestGetForkSyncStatus(t *testing.T) {
	client, server := newTestServer(jsonResponse(200, &ForkSyncStatus{
		Synced:   true,
		BehindBy: 0,
		AheadBy:  5,
	}))
	defer server.Close()

	status, err := client.GetForkSyncStatus(context.Background(), "owner", "fork-repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !status.Synced {
		t.Error("expected synced to be true")
	}
}

func TestSyncForkRepository(t *testing.T) {
	client, server := newTestServer(jsonResponse(200, nil))
	defer server.Close()

	err := client.SyncForkRepository(context.Background(), "owner", "fork-repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetRepoLicense(t *testing.T) {
	client, server := newTestServer(jsonResponse(200, &RepoLicense{
		Key: "mit", Name: "MIT License",
	}))
	defer server.Close()

	license, err := client.GetRepoLicense(context.Background(), "owner", "repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if license.Key != "mit" {
		t.Errorf("expected key 'mit', got '%s'", license.Key)
	}
}

func TestListRepoCLAs(t *testing.T) {
	client, server := newTestServer(jsonResponse(200, []*RepoCLA{
		{ID: 1, Name: "CLA", Enabled: true},
	}))
	defer server.Close()

	claList, err := client.ListRepoCLAs(context.Background(), "owner", "repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(claList) != 1 {
		t.Errorf("expected 1 CLA, got %d", len(claList))
	}
}

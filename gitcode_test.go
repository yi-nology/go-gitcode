package gitcode

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)


// getTestClient returns a client for the live-API tests. The suite is
// env-gated: without GITCODE_TOKEN every live test skips, so plain
// `go test` (and CI's unit step) stays hermetic — live credentials must
// never be committed to the repo.
func getTestClient(t *testing.T) *Client {
	t.Helper()
	token := os.Getenv(EnvGitToken)
	if token == "" {
		t.Skip("live API test: set GITCODE_TOKEN to enable")
	}
	if baseURL := os.Getenv(EnvGitBaseURL); baseURL != "" {
		return NewClientWithBaseURL(baseURL, token)
	}
	return NewClient(token)
}

func getTestOwner(t *testing.T, client *Client) string {
	t.Helper()
	ctx := context.Background()
	user, err := client.GetCurrentUser(ctx)
	if err != nil {
		t.Fatalf("Failed to get current user: %v", err)
	}
	return user.Login
}

func getFileTestBranch(t *testing.T, client *Client, owner, repoName string) string {
	t.Helper()
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		branches, err := client.ListBranches(ctx, owner, repoName)
		if err == nil && len(branches) > 0 {
			return branches[0].Name
		}
		time.Sleep(2 * time.Second)
	}
	t.Fatal("Timed out waiting for branch to appear")
	return ""
}

func getRepoDefaultBranch(t *testing.T, client *Client, owner, repoName string, repo *Repository) string {
	t.Helper()
	if repo != nil && repo.DefaultBranch != "" {
		return repo.DefaultBranch
	}
	return getFileTestBranch(t, client, owner, repoName)
}

func generateRepoName() string {
	return fmt.Sprintf("test-api-%d", time.Now().UnixNano())
}

func generateBranchName() string {
	return fmt.Sprintf("test-branch-%d", time.Now().UnixNano())
}

// ============================================================
// Client & Auth Tests
// ============================================================

func TestNewClient(t *testing.T) {
	client := NewClient("test-token")
	if client.baseURL != DefaultBaseURL {
		t.Errorf("Expected baseURL %s, got %s", DefaultBaseURL, client.baseURL)
	}
	if client.token != "test-token" {
		t.Errorf("Expected token 'test-token', got %s", client.token)
	}
	if client.authStyle != AuthStyleBearer {
		t.Errorf("Expected AuthStyleBearer, got %d", client.authStyle)
	}
	if client.httpClient.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", client.httpClient.Timeout)
	}
}

func TestNewClientWithBaseURL(t *testing.T) {
	client := NewClientWithBaseURL("https://custom.api.com/v5/", "test-token")
	if client.baseURL != "https://custom.api.com/v5" {
		t.Errorf("Expected trimmed baseURL, got %s", client.baseURL)
	}
}

func TestSetAuthStyle(t *testing.T) {
	client := NewClient("test")
	client.SetAuthStyle(AuthStylePrivateToken)
	if client.authStyle != AuthStylePrivateToken {
		t.Errorf("Expected AuthStylePrivateToken, got %d", client.authStyle)
	}
}

func TestSetHTTPClient(t *testing.T) {
	client := NewClient("test")
	custom := &http.Client{Timeout: 10 * time.Second}
	client.SetHTTPClient(custom)
	if client.httpClient != custom {
		t.Error("Expected custom HTTP client to be set")
	}
}

func TestListOptionsWithDefaults(t *testing.T) {
	opts := ListOptions{}
	defaulted := opts.withDefaults()
	if defaulted.Page != 1 {
		t.Errorf("Expected default Page=1, got %d", defaulted.Page)
	}
	if defaulted.PerPage != 20 {
		t.Errorf("Expected default PerPage=20, got %d", defaulted.PerPage)
	}
}

func TestListOptionsToQuery(t *testing.T) {
	opts := ListOptions{Page: 3, PerPage: 50}
	q := opts.toQuery()
	expected := "page=3&per_page=50"
	if q != expected {
		t.Errorf("Expected query %q, got %q", expected, q)
	}
}

// ============================================================
// Live API: User Tests
// ============================================================

func TestGetCurrentUser(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	user, err := client.GetCurrentUser(ctx)
	if err != nil {
		t.Fatalf("GetCurrentUser failed: %v", err)
	}
	if user.ID == "" {
		t.Error("Expected non-empty user ID")
	}
	if user.Login == "" {
		t.Error("Expected non-empty login")
	}
	t.Logf("Current user: %s (ID: %s, Name: %s)", user.Login, user.ID, user.Name)
}

func TestGetUser(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	me, err := client.GetCurrentUser(ctx)
	if err != nil {
		t.Fatalf("GetCurrentUser failed: %v", err)
	}

	user, err := client.GetUser(ctx, me.Login)
	if err != nil {
		t.Fatalf("GetUser(%s) failed: %v", me.Login, err)
	}
	if user.Login != me.Login {
		t.Errorf("Expected login %s, got %s", me.Login, user.Login)
	}
}

// ============================================================
// Live API: Repository CRUD Tests
// ============================================================

func TestRepositoryCRUD(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	owner := getTestOwner(t, client)
	repoName := generateRepoName()

	t.Logf("Testing repo CRUD: %s/%s", owner, repoName)

	// Create
	repo, err := client.CreateRepository(ctx, CreateRepositoryOptions{
		Name:        repoName,
		Description: "Test repo for API testing",
		AutoInit:    boolPtr(true),
		Private:     boolPtr(false),
	})
	if err != nil {
		t.Fatalf("CreateRepository failed: %v", err)
	}
	if repo.Name != repoName {
		t.Errorf("Expected repo name %s, got %s", repoName, repo.Name)
	}
	if repo.FullName != fmt.Sprintf("%s/%s", owner, repoName) {
		t.Errorf("Unexpected full_name: %s", repo.FullName)
	}
	t.Logf("Created repo: %s (ID: %d)", repo.FullName, repo.ID)

	// Wait for repo to be fully initialized
	time.Sleep(3 * time.Second)

	// Get
	repo, err = client.GetRepository(ctx, owner, repoName)
	if err != nil {
		t.Fatalf("GetRepository failed: %v", err)
	}
	if repo.Description != "Test repo for API testing" {
		t.Errorf("Unexpected description: %s", repo.Description)
	}

	// Update
	repo, err = client.UpdateRepository(ctx, owner, repoName, UpdateRepositoryOptions{
		Description: "Updated description",
	})
	if err != nil {
		t.Fatalf("UpdateRepository failed: %v", err)
	}
	if repo.Description != "Updated description" {
		t.Errorf("Expected updated description, got: %s", repo.Description)
	}

	// List repos
	repos, err := client.ListRepositories(ctx, ListRepositoriesOptions{
		ListOptions: ListOptions{Page: 1, PerPage: 5},
	})
	if err != nil {
		t.Fatalf("ListRepositories failed: %v", err)
	}
	if len(repos) == 0 {
		t.Error("Expected at least one repository")
	}
	t.Logf("Listed %d repositories", len(repos))

	// Cleanup: Delete
	err = client.DeleteRepository(ctx, owner, repoName)
	if err != nil {
		t.Fatalf("DeleteRepository failed: %v", err)
	}
	t.Logf("Deleted repo: %s/%s", owner, repoName)
}

// ============================================================
// Live API: File Operations Tests
// ============================================================

func TestFileOperations(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	owner := getTestOwner(t, client)
	repoName := generateRepoName()

	repo, err := client.CreateRepository(ctx, CreateRepositoryOptions{
		Name:     repoName,
		AutoInit: boolPtr(true),
		Private:  boolPtr(true),
	})
	if err != nil {
		t.Fatalf("CreateRepository failed: %v", err)
	}
	t.Logf("Created test repo: %s", repo.FullName)
	defer client.DeleteRepository(ctx, owner, repoName)

	defaultBranch := getFileTestBranch(t, client, owner, repoName)
	t.Logf("Using branch: %s", defaultBranch)

	content := base64.StdEncoding.EncodeToString([]byte("Hello, GitCode API!"))
	fileResult, err := client.CreateFile(ctx, owner, repoName, "test-file.txt", CreateFileOptions{
		Message: "Add test file",
		Content: content,
		Branch:  defaultBranch,
	})
	if err != nil {
		t.Fatalf("CreateFile failed: %v", err)
	}
	if fileResult.Content == nil {
		t.Log("Warning: file result content is nil (API may not return content)")
	} else {
		t.Logf("Created file: %s (SHA: %s)", fileResult.Content.Path, fileResult.Content.SHA)
	}

	fileContent, err := client.GetRepositoryContent(ctx, owner, repoName, "test-file.txt", defaultBranch)
	if err != nil {
		t.Fatalf("GetRepositoryContent failed: %v", err)
	}
	if fileContent.Name != "test-file.txt" {
		t.Errorf("Expected file name test-file.txt, got %s", fileContent.Name)
	}

	newContent := base64.StdEncoding.EncodeToString([]byte("Updated content!"))
	updateResult, err := client.UpdateFile(ctx, owner, repoName, "test-file.txt", UpdateFileOptions{
		Message: "Update test file",
		Content: newContent,
		SHA:     fileContent.SHA,
		Branch:  defaultBranch,
	})
	if err != nil {
		t.Fatalf("UpdateFile failed: %v", err)
	}
	var updateSHA string
	if updateResult.Content != nil {
		updateSHA = updateResult.Content.SHA
		t.Logf("Updated file, new SHA: %s", updateSHA)
	} else {
		updated, err := client.GetRepositoryContent(ctx, owner, repoName, "test-file.txt", defaultBranch)
		if err != nil {
			t.Fatalf("GetRepositoryContent after update failed: %v", err)
		}
		updateSHA = updated.SHA
		t.Logf("Updated file, fetched SHA: %s", updateSHA)
	}

	_, err = client.DeleteFile(ctx, owner, repoName, "test-file.txt", DeleteFileOptions{
		Message: "Delete test file",
		SHA:     updateSHA,
		Branch:  defaultBranch,
	})
	if err != nil {
		t.Fatalf("DeleteFile failed: %v", err)
	}
	t.Logf("Deleted file successfully")

	contents, err := client.ListRepositoryContents(ctx, owner, repoName, "", defaultBranch)
	if err != nil {
		t.Fatalf("ListRepositoryContents failed: %v", err)
	}
	t.Logf("Root contents: %d items", len(contents))
}

// ============================================================
// Live API: Branch Operations Tests
// ============================================================

func TestBranchOperations(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	owner := getTestOwner(t, client)
	repoName := generateRepoName()

	repo, err := client.CreateRepository(ctx, CreateRepositoryOptions{
		Name:     repoName,
		AutoInit: boolPtr(true),
		Private:  boolPtr(true),
	})
	if err != nil {
		t.Fatalf("CreateRepository failed: %v", err)
	}
	t.Logf("Created test repo: %s", repo.FullName)
	defer client.DeleteRepository(ctx, owner, repoName)

	time.Sleep(3 * time.Second)

	defaultBranch := getRepoDefaultBranch(t, client, owner, repoName, repo)

	// List branches
	branches, err := client.ListBranches(ctx, owner, repoName)
	if err != nil {
		t.Fatalf("ListBranches failed: %v", err)
	}
	if len(branches) == 0 {
		t.Error("Expected at least one branch")
	}
	t.Logf("Branches: %v", func() []string {
		names := make([]string, len(branches))
		for i, b := range branches {
			names[i] = b.Name
		}
		return names
	}())

	// Create branch
	branchName := generateBranchName()
	branch, err := client.CreateBranch(ctx, owner, repoName, CreateBranchOptions{
		BranchName: branchName,
		Refs:       defaultBranch,
	})
	if err != nil {
		t.Fatalf("CreateBranch failed: %v", err)
	}
	t.Logf("Created branch: %s", branch.Name)

	// Get branch
	branch, err = client.GetBranch(ctx, owner, repoName, branchName)
	if err != nil {
		t.Fatalf("GetBranch failed: %v", err)
	}
	if branch.Name != branchName {
		t.Errorf("Expected branch name %s, got %s", branchName, branch.Name)
	}

	// Compare commits
	cmp, err := client.CompareCommits(ctx, owner, repoName, defaultBranch, branchName)
	if err != nil {
		t.Logf("CompareCommits: %v (may be OK if same)", err)
	} else {
		t.Logf("Compare: ahead=%d, behind=%d, total=%d", cmp.AheadBy, cmp.BehindBy, cmp.TotalCommits)
	}

	// List commits
	commits, err := client.ListCommits(ctx, owner, repoName, ListCommitsOptions{
		ListOptions: ListOptions{PerPage: 5},
		Branch:      defaultBranch,
	})
	if err != nil {
		t.Fatalf("ListCommits failed: %v", err)
	}
	if len(commits) == 0 {
		t.Error("Expected at least one commit")
	}
	t.Logf("Found %d commits", len(commits))

	// Get single commit
	if len(commits) > 0 && commits[0].SHA != "" {
		commit, err := client.GetCommit(ctx, owner, repoName, commits[0].SHA)
		if err != nil {
			t.Fatalf("GetCommit failed: %v", err)
		}
		t.Logf("Commit %s: %s", commit.SHA[:8], commit.Message)
	}

	// Delete branch
	err = client.DeleteBranch(ctx, owner, repoName, branchName)
	if err != nil {
		t.Fatalf("DeleteBranch failed: %v", err)
	}
	t.Logf("Deleted branch: %s", branchName)
}

// ============================================================
// Live API: Tag & Release Tests
// ============================================================

func TestTagsAndReleases(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	owner := getTestOwner(t, client)
	repoName := generateRepoName()

	_, err := client.CreateRepository(ctx, CreateRepositoryOptions{
		Name:     repoName,
		AutoInit: boolPtr(true),
		Private:  boolPtr(true),
	})
	if err != nil {
		t.Fatalf("CreateRepository failed: %v", err)
	}
	defer client.DeleteRepository(ctx, owner, repoName)
	time.Sleep(3 * time.Second)

	// List tags (should be empty)
	tags, err := client.ListTags(ctx, owner, repoName)
	if err != nil {
		t.Fatalf("ListTags failed: %v", err)
	}
	t.Logf("Initial tags: %d", len(tags))

	// Create release (creates tag automatically)
	release, err := client.CreateRelease(ctx, owner, repoName, CreateReleaseOptions{
		TagName: "v0.1.0",
		Title:   "Test Release v0.1.0",
		Body:    "This is a test release created by the API test suite",
	})
	if err != nil {
		t.Fatalf("CreateRelease failed: %v", err)
	}
	t.Logf("Created release: %s (ID: %d)", release.TagName, release.ID)

	// List releases
	releases, err := client.ListReleases(ctx, owner, repoName)
	if err != nil {
		t.Fatalf("ListReleases failed: %v", err)
	}
	if len(releases) == 0 {
		t.Error("Expected at least one release")
	}

	// List tags again
	tags, err = client.ListTags(ctx, owner, repoName)
	if err != nil {
		t.Fatalf("ListTags (after release) failed: %v", err)
	}
	if len(tags) == 0 {
		t.Error("Expected at least one tag after creating release")
	}
	t.Logf("Tags after release: %d", len(tags))

	// Delete release
	err = client.DeleteRelease(ctx, owner, repoName, release.TagName)
	if err != nil {
		t.Logf("DeleteRelease failed (may not be supported): %v", err)
	} else {
		t.Logf("Deleted release: %s", release.TagName)
	}
}

// ============================================================
// Live API: Issue CRUD Tests
// ============================================================

func TestIssueCRUD(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	owner := getTestOwner(t, client)
	repoName := generateRepoName()

	_, err := client.CreateRepository(ctx, CreateRepositoryOptions{
		Name:     repoName,
		AutoInit: boolPtr(true),
		Private:  boolPtr(true),
	})
	if err != nil {
		t.Fatalf("CreateRepository failed: %v", err)
	}
	defer client.DeleteRepository(ctx, owner, repoName)
	time.Sleep(3 * time.Second)

	// Create issue
	issue, err := client.CreateIssue(ctx, owner, repoName, CreateIssueOptions{
		Title: "Test Issue from API",
		Body:  "This is a test issue created by the API test suite",
	})
	if err != nil {
		t.Fatalf("CreateIssue failed: %v", err)
	}
	t.Logf("Created issue #%d: %s", int(issue.Number), issue.Title)

	// Get issue
	issue, err = client.GetIssue(ctx, owner, repoName, int(issue.Number))
	if err != nil {
		t.Fatalf("GetIssue failed: %v", err)
	}
	if issue.State != IssueStateOpen {
		t.Errorf("Expected open state, got %s", issue.State)
	}

	// Update issue
	issue, err = client.UpdateIssue(ctx, owner, repoName, int(issue.Number), UpdateIssueOptions{
		Title: "Updated Issue Title",
		Body:  "Updated body",
	})
	if err != nil {
		t.Fatalf("UpdateIssue failed: %v", err)
	}
	if issue.Title != "Updated Issue Title" {
		t.Errorf("Expected updated title, got: %s", issue.Title)
	}

	// List issues
	issues, err := client.ListIssues(ctx, owner, repoName, ListIssuesOptions{
		State: IssueStateOpen,
	})
	if err != nil {
		t.Fatalf("ListIssues failed: %v", err)
	}
	if len(issues) == 0 {
		t.Error("Expected at least one issue")
	}
	t.Logf("Found %d open issues", len(issues))

	// Issue comments
	comment, err := client.CreateIssueComment(ctx, owner, repoName, int(issue.Number), "Test comment from API")
	if err != nil {
		t.Fatalf("CreateIssueComment failed: %v", err)
	}
	t.Logf("Created comment ID: %d", comment.ID)

	// List comments
	comments, err := client.ListIssueComments(ctx, owner, repoName, int(issue.Number))
	if err != nil {
		t.Fatalf("ListIssueComments failed: %v", err)
	}
	if len(comments) == 0 {
		t.Error("Expected at least one comment")
	}

	// Update comment
	updatedComment, err := client.UpdateIssueComment(ctx, owner, repoName, comment.ID, "Updated comment")
	if err != nil {
		t.Logf("UpdateIssueComment: %v (API may return empty body)", err)
	} else if updatedComment != nil && updatedComment.Body != "" {
		if updatedComment.Body != "Updated comment" {
			t.Errorf("Expected 'Updated comment', got: %s", updatedComment.Body)
		}
	}

	// Delete comment
	err = client.DeleteIssueComment(ctx, owner, repoName, comment.ID)
	if err != nil {
		t.Fatalf("DeleteIssueComment failed: %v", err)
	}

	// Close issue
	issue, err = client.CloseIssue(ctx, owner, repoName, int(issue.Number))
	if err != nil {
		t.Fatalf("CloseIssue failed: %v", err)
	}
	t.Logf("Closed issue #%d (state: %s)", int(issue.Number), issue.State)

	// Reopen issue
	issue, err = client.ReopenIssue(ctx, owner, repoName, int(issue.Number))
	if err != nil {
		t.Fatalf("ReopenIssue failed: %v", err)
	}
	t.Logf("Reopened issue #%d (state: %s)", int(issue.Number), issue.State)
}

// ============================================================
// Live API: Issue Labels Tests
// ============================================================

func TestIssueLabels(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	owner := getTestOwner(t, client)
	repoName := generateRepoName()

	_, err := client.CreateRepository(ctx, CreateRepositoryOptions{
		Name:     repoName,
		AutoInit: boolPtr(true),
		Private:  boolPtr(true),
	})
	if err != nil {
		t.Fatalf("CreateRepository failed: %v", err)
	}
	defer client.DeleteRepository(ctx, owner, repoName)
	time.Sleep(3 * time.Second)

	// Create label
	label, err := client.CreateIssueLabel(ctx, owner, repoName, "test-label", "#ff0000")
	if err != nil {
		t.Fatalf("CreateIssueLabel failed: %v", err)
	}
	t.Logf("Created label: %s (ID: %d)", label.Name, label.ID)

	// List labels
	labels, err := client.ListIssueLabels(ctx, owner, repoName)
	if err != nil {
		t.Fatalf("ListIssueLabels failed: %v", err)
	}
	if len(labels) == 0 {
		t.Error("Expected at least one label")
	}

	// Create issue with label
	issue, err := client.CreateIssue(ctx, owner, repoName, CreateIssueOptions{
		Title: "Test issue with label",
		Body:  "Test body for label issue",
	})
	if err != nil {
		t.Fatalf("CreateIssue failed: %v", err)
	}
	t.Logf("Created issue #%d", int(issue.Number))

	// Add labels to issue
	err = client.AddIssueLabels(ctx, owner, repoName, int(issue.Number), []string{"test-label"})
	if err != nil {
		t.Logf("AddIssueLabels: %v (may already have it)", err)
	}

	// Remove label
	err = client.RemoveIssueLabel(ctx, owner, repoName, int(issue.Number), "test-label")
	if err != nil {
		t.Logf("RemoveIssueLabel: %v", err)
	}

	// Delete label
	err = client.DeleteIssueLabel(ctx, owner, repoName, "test-label")
	if err != nil {
		t.Fatalf("DeleteIssueLabel failed: %v", err)
	}
	t.Logf("Deleted label: test-label")
}

// ============================================================
// Live API: Milestone Tests
// ============================================================

func TestMilestones(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	owner := getTestOwner(t, client)
	repoName := generateRepoName()

	_, err := client.CreateRepository(ctx, CreateRepositoryOptions{
		Name:     repoName,
		AutoInit: boolPtr(true),
		Private:  boolPtr(true),
	})
	if err != nil {
		t.Fatalf("CreateRepository failed: %v", err)
	}
	defer client.DeleteRepository(ctx, owner, repoName)
	time.Sleep(3 * time.Second)

	milestone, err := client.CreateMilestone(ctx, owner, repoName, "v1.0", "First milestone")
	if err != nil {
		t.Skipf("CreateMilestone not supported: %v", err)
	}
	t.Logf("Created milestone: %s (ID: %d)", milestone.Title, milestone.ID)

	milestones, err := client.ListMilestones(ctx, owner, repoName)
	if err != nil {
		t.Fatalf("ListMilestones failed: %v", err)
	}
	if len(milestones) == 0 {
		t.Error("Expected at least one milestone")
	}

	err = client.DeleteMilestone(ctx, owner, repoName, int(milestone.ID))
	if err != nil {
		t.Logf("DeleteMilestone failed: %v", err)
	} else {
		t.Logf("Deleted milestone: %d", milestone.ID)
	}
}

// ============================================================
// Live API: Pull Request Tests
// ============================================================

func TestPullRequestCRUD(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	owner := getTestOwner(t, client)
	repoName := generateRepoName()

	repo, err := client.CreateRepository(ctx, CreateRepositoryOptions{
		Name:     repoName,
		AutoInit: boolPtr(true),
		Private:  boolPtr(true),
	})
	if err != nil {
		t.Fatalf("CreateRepository failed: %v", err)
	}
	defer client.DeleteRepository(ctx, owner, repoName)
	time.Sleep(3 * time.Second)

	defaultBranch := getRepoDefaultBranch(t, client, owner, repoName, repo)

	// Create branch for PR
	branchName := generateBranchName()
	_, err = client.CreateBranch(ctx, owner, repoName, CreateBranchOptions{
		BranchName: branchName,
		Refs:       defaultBranch,
	})
	if err != nil {
		t.Fatalf("CreateBranch for PR failed: %v", err)
	}

	// Create a file on the branch to have a diff
	content := base64.StdEncoding.EncodeToString([]byte("PR test content"))
	_, err = client.CreateFile(ctx, owner, repoName, "pr-test-file.txt", CreateFileOptions{
		Message: "Add file for PR",
		Content: content,
		Branch:  branchName,
	})
	if err != nil {
		t.Fatalf("CreateFile on branch failed: %v", err)
	}

	// Create PR
	pr, err := client.CreatePullRequest(ctx, owner, repoName, CreatePullRequestOptions{
		Title: "Test Pull Request from API",
		Body:  "This is a test PR created by the API test suite",
		Head:  branchName,
		Base:  defaultBranch,
	})
	if err != nil {
		t.Fatalf("CreatePullRequest failed: %v", err)
	}
	t.Logf("Created PR #%d: %s", pr.Number, pr.Title)

	// Get PR
	pr, err = client.GetPullRequest(ctx, owner, repoName, pr.Number)
	if err != nil {
		t.Fatalf("GetPullRequest failed: %v", err)
	}
	if pr.State != PullRequestStateOpen {
		t.Errorf("Expected open state, got %s", pr.State)
	}

	// List PRs
	prs, err := client.ListPullRequests(ctx, owner, repoName, ListPullRequestsOptions{
		State: PullRequestStateOpen,
	})
	if err != nil {
		t.Fatalf("ListPullRequests failed: %v", err)
	}
	if len(prs) == 0 {
		t.Error("Expected at least one PR")
	}
	t.Logf("Found %d open PRs", len(prs))

	// List PR files
	files, err := client.ListPullRequestFiles(ctx, owner, repoName, pr.Number)
	if err != nil {
		t.Logf("ListPullRequestFiles: %v", err)
	} else {
		t.Logf("PR files: %d", len(files))
		for _, f := range files {
			t.Logf("  - %s (+%d/-%d)", f.Filename, f.Additions, f.Deletions)
		}
	}

	// List PR commits
	prCommits, err := client.ListPullRequestCommits(ctx, owner, repoName, pr.Number)
	if err != nil {
		t.Logf("ListPullRequestCommits: %v", err)
	} else {
		t.Logf("PR commits: %d", len(prCommits))
	}

	// PR comment
	prComment, err := client.CreatePullRequestComment(ctx, owner, repoName, pr.Number, "Test PR comment", "", "", "")
	if err != nil {
		t.Logf("CreatePullRequestComment: %v", err)
	} else {
		t.Logf("Created PR comment ID: %s", string(prComment.ID))
	}

	// PR reviews
	reviews, err := client.ListPullRequestReviews(ctx, owner, repoName, pr.Number)
	if err != nil {
		t.Logf("ListPullRequestReviews: %v", err)
	} else {
		t.Logf("PR reviews: %d", len(reviews))
	}

	// Check if merged
	merged, err := client.IsPullRequestMerged(ctx, owner, repoName, pr.Number)
	if err != nil {
		t.Logf("IsPullRequestMerged: %v", err)
	} else {
		t.Logf("PR merged: %v", merged)
	}

	// Update PR
	prUpdated, err := client.UpdatePullRequest(ctx, owner, repoName, pr.Number, UpdatePullRequestOptions{
		Title: "Updated PR Title",
		Body:  "Updated body",
	})
	if err != nil {
		t.Fatalf("UpdatePullRequest failed: %v", err)
	}
	if prUpdated.Title != "Updated PR Title" {
		t.Errorf("Expected updated title, got: %s", prUpdated.Title)
	}
	t.Logf("Updated PR #%d title: %s", pr.Number, prUpdated.Title)

	// Close PR
	_, err = client.ClosePullRequest(ctx, owner, repoName, pr.Number)
	if err != nil {
		t.Fatalf("ClosePullRequest failed: %v", err)
	}
	t.Logf("Closed PR #%d", pr.Number)
}

// ============================================================
// Live API: Webhook Tests
// ============================================================

func TestWebhookCRUD(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	owner := getTestOwner(t, client)
	repoName := generateRepoName()

	_, err := client.CreateRepository(ctx, CreateRepositoryOptions{
		Name:     repoName,
		AutoInit: boolPtr(true),
		Private:  boolPtr(true),
	})
	if err != nil {
		t.Fatalf("CreateRepository failed: %v", err)
	}
	defer client.DeleteRepository(ctx, owner, repoName)
	time.Sleep(3 * time.Second)

	// List webhooks (should be empty)
	hooks, err := client.ListWebhooks(ctx, owner, repoName)
	if err != nil {
		t.Fatalf("ListWebhooks failed: %v", err)
	}
	t.Logf("Initial webhooks: %d", len(hooks))

	// Create webhook
	active := true
	hook, err := client.CreateWebhook(ctx, owner, repoName, CreateWebhookOptions{
		URL:                 "https://example.com/webhook/test",
		Password:            "test-secret",
		PushEvents:          Bool(true),
		MergeRequestsEvents: Bool(true),
		Active:              &active,
	})
	if err != nil {
		t.Fatalf("CreateWebhook failed: %v", err)
	}
	t.Logf("Created webhook ID: %d", hook.ID)

	// Get webhook
	hook, err = client.GetWebhook(ctx, owner, repoName, hook.ID)
	if err != nil {
		t.Fatalf("GetWebhook failed: %v", err)
	}
	if !hook.Active {
		t.Error("Expected webhook to be active")
	}

	// Update webhook
	newURL := "https://example.com/webhook/updated"
	hook, err = client.UpdateWebhook(ctx, owner, repoName, hook.ID, UpdateWebhookOptions{
		URL: newURL,
	})
	if err != nil {
		t.Fatalf("UpdateWebhook failed: %v", err)
	}

	// List webhooks again
	hooks, err = client.ListWebhooks(ctx, owner, repoName)
	if err != nil {
		t.Fatalf("ListWebhooks failed: %v", err)
	}
	if len(hooks) != 1 {
		t.Errorf("Expected 1 webhook, got %d", len(hooks))
	}

	// Delete webhook
	err = client.DeleteWebhook(ctx, owner, repoName, hook.ID)
	if err != nil {
		t.Fatalf("DeleteWebhook failed: %v", err)
	}
	t.Logf("Deleted webhook: %d", hook.ID)
}

// ============================================================
// Live API: Organization & Star Tests
// ============================================================

func TestOrganizations(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	orgs, err := client.ListOrganizations(ctx)
	if err != nil {
		t.Skipf("ListOrganizations not available (GitCode API does not support /user/orgs): %v", err)
	}
	t.Logf("User organizations: %d", len(orgs))
	for _, org := range orgs {
		t.Logf("  - %s (%s)", org.Login, org.Name)
	}

	if len(orgs) > 0 {
		org, err := client.GetOrganization(ctx, orgs[0].Login)
		if err != nil {
			t.Fatalf("GetOrganization failed: %v", err)
		}
		t.Logf("Org details: %s (ID: %d)", org.Login, org.ID)

		members, err := client.ListOrganizationMembers(ctx, orgs[0].Login)
		if err != nil {
			t.Logf("ListOrganizationMembers failed: %v", err)
		} else {
			t.Logf("Org members: %d", len(members))
		}
	}
}

func TestStarOperations(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	owner := getTestOwner(t, client)
	repoName := generateRepoName()

	_, err := client.CreateRepository(ctx, CreateRepositoryOptions{
		Name:     repoName,
		AutoInit: boolPtr(true),
		Private:  boolPtr(true),
	})
	if err != nil {
		t.Fatalf("CreateRepository failed: %v", err)
	}
	defer client.DeleteRepository(ctx, owner, repoName)
	time.Sleep(3 * time.Second)

	err = client.StarRepository(ctx, owner, repoName)
	if err != nil {
		t.Skipf("StarRepository not available (GitCode API does not support /user/starred): %v", err)
	}
	t.Logf("Starred repo: %s/%s", owner, repoName)

	starred, err := client.IsRepositoryStarred(ctx, owner, repoName)
	if err != nil {
		t.Fatalf("IsRepositoryStarred failed: %v", err)
	}
	if !starred {
		t.Error("Expected repo to be starred")
	}

	err = client.UnstarRepository(ctx, owner, repoName)
	if err != nil {
		t.Fatalf("UnstarRepository failed: %v", err)
	}
	t.Logf("Unstarred repo: %s/%s", owner, repoName)

	starred, err = client.IsRepositoryStarred(ctx, owner, repoName)
	if err != nil {
		t.Fatalf("IsRepositoryStarred (after unstar) failed: %v", err)
	}
	if starred {
		t.Error("Expected repo to NOT be starred after unstar")
	}
}

// ============================================================
// Live API: Contributors & Languages Tests
// ============================================================

func TestContributorsAndLanguages(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	owner := getTestOwner(t, client)
	repoName := generateRepoName()

	_, err := client.CreateRepository(ctx, CreateRepositoryOptions{
		Name:     repoName,
		AutoInit: boolPtr(true),
		Private:  boolPtr(true),
	})
	if err != nil {
		t.Fatalf("CreateRepository failed: %v", err)
	}
	defer client.DeleteRepository(ctx, owner, repoName)
	time.Sleep(3 * time.Second)

	// List contributors
	contributors, err := client.ListContributors(ctx, owner, repoName)
	if err != nil {
		t.Logf("ListContributors failed: %v", err)
	} else {
		t.Logf("Contributors: %d", len(contributors))
	}

	// Get languages
	langs, err := client.GetLanguages(ctx, owner, repoName)
	if err != nil {
		t.Logf("GetLanguages failed: %v", err)
	} else {
		t.Logf("Languages: %v", langs)
	}
}

// ============================================================
// Live API: Notifications Tests
// ============================================================

func TestNotifications(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	_, err := client.ListNotifications(ctx)
	if err != nil {
		t.Skipf("ListNotifications not available (GitCode API does not support /notifications): %v", err)
	}
}

// ============================================================
// Live API: Rate Limit Tests
// ============================================================

func TestRateLimit(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()

	rl, err := client.GetRateLimit(ctx)
	if err != nil {
		t.Logf("GetRateLimit failed: %v (endpoint may not exist)", err)
	} else {
		t.Logf("Rate limit: %d remaining / %d limit", rl.Remaining, rl.Limit)
	}
}

// ============================================================
// Webhook Event Parsing Tests
// ============================================================

func TestParsePushEvent(t *testing.T) {
	client := getTestClient(t)
	payload := []byte(`{
		"ref": "refs/heads/main",
		"before": "abc123",
		"after": "def456",
		"repository": {"id": 1, "full_name": "test/repo", "name": "repo"},
		"commits": [{"sha": "def456", "message": "test commit"}],
		"sender": {"id": 1, "login": "testuser"}
	}`)

	event, err := client.ParsePushEvent(payload)
	if err != nil {
		t.Fatalf("ParsePushEvent failed: %v", err)
	}
	if event.Ref != "refs/heads/main" {
		t.Errorf("Expected ref refs/heads/main, got %s", event.Ref)
	}
	if event.Repository == nil || event.Repository.FullName != "test/repo" {
		t.Error("Expected repository to be parsed correctly")
	}
	if len(event.Commits) != 1 {
		t.Errorf("Expected 1 commit, got %d", len(event.Commits))
	}
	if event.Sender.Login != "testuser" {
		t.Errorf("Expected sender testuser, got %s", event.Sender.Login)
	}
}

func TestParsePullRequestEvent(t *testing.T) {
	client := getTestClient(t)
	payload := []byte(`{
		"action": "opened",
		"number": 42,
		"pull_request": {"id": 1, "number": 42, "title": "Test PR"},
		"repository": {"id": 1, "full_name": "test/repo"},
		"sender": {"id": 1, "login": "testuser"}
	}`)

	event, err := client.ParsePullRequestEvent(payload)
	if err != nil {
		t.Fatalf("ParsePullRequestEvent failed: %v", err)
	}
	if event.Action != "opened" {
		t.Errorf("Expected action opened, got %s", event.Action)
	}
	if event.Number != 42 {
		t.Errorf("Expected number 42, got %d", event.Number)
	}
}

func TestParseIssueEvent(t *testing.T) {
	client := getTestClient(t)
	payload := []byte(`{
		"action": "closed",
		"issue": {"id": 1, "number": 7, "title": "Test Issue", "state": "closed"},
		"repository": {"id": 1, "full_name": "test/repo"},
		"sender": {"id": 1, "login": "testuser"}
	}`)

	event, err := client.ParseIssueEvent(payload)
	if err != nil {
		t.Fatalf("ParseIssueEvent failed: %v", err)
	}
	if event.Action != "closed" {
		t.Errorf("Expected action closed, got %s", event.Action)
	}
	if event.Issue.Number != 7 {
		t.Errorf("Expected issue number 7, got %d", event.Issue.Number)
	}
}

func TestParsePushEventInvalid(t *testing.T) {
	client := getTestClient(t)
	_, err := client.ParsePushEvent([]byte(`invalid json`))
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

// ============================================================
// Timestamp Unmarshal Test
// ============================================================

func TestTimestampUnmarshal(t *testing.T) {
	ts := &Timestamp{}
	err := ts.UnmarshalJSON([]byte(`"2024-01-15T10:30:00Z"`))
	if err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}
	if ts.Year() != 2024 || ts.Month() != time.January || ts.Day() != 15 {
		t.Errorf("Unexpected date: %v", ts.Time)
	}
}

func TestTimestampUnmarshalInvalid(t *testing.T) {
	ts := &Timestamp{}
	err := ts.UnmarshalJSON([]byte(`"not-a-date"`))
	if err == nil {
		t.Error("Expected error for invalid date")
	}
}

// ============================================================
// Error Type Test
// ============================================================

func TestErrorType(t *testing.T) {
	e := &Error{Message: "test error"}
	if e.Error() != "test error" {
		t.Errorf("Expected 'test error', got '%s'", e.Error())
	}
}

// ============================================================
// Branch Protection Tests (create + list + delete)
// ============================================================

func TestBranchProtection(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	owner := getTestOwner(t, client)
	repoName := generateRepoName()

	repo, err := client.CreateRepository(ctx, CreateRepositoryOptions{
		Name:     repoName,
		AutoInit: boolPtr(true),
		Private:  boolPtr(true),
	})
	if err != nil {
		t.Fatalf("CreateRepository failed: %v", err)
	}
	defer client.DeleteRepository(ctx, owner, repoName)
	time.Sleep(3 * time.Second)

	defaultBranch := getRepoDefaultBranch(t, client, owner, repoName, repo)

	// Create branch protection
	rule, err := client.CreateBranchProtection(ctx, owner, repoName, CreateBranchProtectionOptions{
		Name:                     defaultBranch,
		RequiredStatusChecks:     false,
		RequiredApprovingReviews: 1,
		AllowForcePushes:         false,
		AllowDeletions:           false,
	})
	if err != nil {
		t.Logf("CreateBranchProtection failed: %v (may not be supported)", err)
	} else {
		t.Logf("Created branch protection: %s", rule.Name)

		// List protections
		rules, err := client.ListBranchProtections(ctx, owner, repoName)
		if err != nil {
			t.Logf("ListBranchProtections failed: %v", err)
		} else {
			t.Logf("Branch protections: %d", len(rules))
		}

		// Delete protection
		err = client.DeleteBranchProtection(ctx, owner, repoName, defaultBranch)
		if err != nil {
			t.Logf("DeleteBranchProtection failed: %v", err)
		} else {
			t.Logf("Deleted branch protection for: %s", defaultBranch)
		}
	}
}

// ============================================================
// Merge PR Test
// ============================================================

func TestMergePullRequest(t *testing.T) {
	client := getTestClient(t)
	ctx := context.Background()
	owner := getTestOwner(t, client)
	repoName := generateRepoName()

	repo, err := client.CreateRepository(ctx, CreateRepositoryOptions{
		Name:     repoName,
		AutoInit: boolPtr(true),
		Private:  boolPtr(true),
	})
	if err != nil {
		t.Fatalf("CreateRepository failed: %v", err)
	}
	defer client.DeleteRepository(ctx, owner, repoName)
	time.Sleep(3 * time.Second)

	defaultBranch := getRepoDefaultBranch(t, client, owner, repoName, repo)

	// Create branch and add a file
	branchName := generateBranchName()
	_, err = client.CreateBranch(ctx, owner, repoName, CreateBranchOptions{
		BranchName: branchName,
		Refs:       defaultBranch,
	})
	if err != nil {
		t.Fatalf("CreateBranch failed: %v", err)
	}

	content := base64.StdEncoding.EncodeToString([]byte("Merge test content"))
	_, err = client.CreateFile(ctx, owner, repoName, "merge-test.txt", CreateFileOptions{
		Message: "File for merge test",
		Content: content,
		Branch:  branchName,
	})
	if err != nil {
		t.Fatalf("CreateFile failed: %v", err)
	}

	// Create PR
	pr, err := client.CreatePullRequest(ctx, owner, repoName, CreatePullRequestOptions{
		Title: "Merge Test PR",
		Body:  "PR to test merge functionality",
		Head:  branchName,
		Base:  defaultBranch,
	})
	if err != nil {
		t.Fatalf("CreatePullRequest failed: %v", err)
	}
	t.Logf("Created PR #%d for merge test", pr.Number)

	// Merge PR
	err = client.MergePullRequest(ctx, owner, repoName, pr.Number, &MergePullRequestOptions{
		CommitMessage: "Merge test PR",
	})
	if err != nil {
		t.Fatalf("MergePullRequest failed: %v", err)
	}
	t.Logf("Merged PR #%d", pr.Number)

	// Verify merged
	merged, err := client.IsPullRequestMerged(ctx, owner, repoName, pr.Number)
	if err != nil {
		t.Logf("IsPullRequestMerged check: %v", err)
	} else if !merged {
		t.Error("Expected PR to be merged")
	} else {
		t.Logf("PR is confirmed merged")
	}
}

// ============================================================
// Helper
// ============================================================

func boolPtr(b bool) *bool {
	return &b
}

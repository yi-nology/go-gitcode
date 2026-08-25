package gitcode

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// --- Workflows ---

// ActionWorkflow represents a CI/CD workflow.
type ActionWorkflow struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	State     string    `json:"state"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ListActionWorkflows lists all workflows for a repository.
//
// GET /repos/{owner}/{repo}/actions/workflows
func (c *Client) ListActionWorkflows(ctx context.Context, owner, repo string, opts ListOptions) ([]*ActionWorkflow, error) {
	var workflows []*ActionWorkflow
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/actions/workflows?%s", owner, repo, opts.toQuery()), nil, &workflows)
	if err != nil {
		return nil, err
	}
	return workflows, nil
}

// --- Workflow Runs ---

// ActionRun represents a workflow run.
type ActionRun struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	WorkflowID   int64     `json:"workflow_id"`
	Status       string    `json:"status"`
	Conclusion   string    `json:"conclusion,omitempty"`
	Branch       string    `json:"head_branch"`
	SHA          string    `json:"head_sha"`
	Event        string    `json:"event"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	RunStartedAt time.Time `json:"run_started_at,omitempty"`
	HTMLURL      string    `json:"html_url,omitempty"`
}

// ListActionRunsOptions specifies options for listing action runs.
type ListActionRunsOptions struct {
	ListOptions
	Status string `json:"status,omitempty"` // queued, in_progress, completed
	Branch string `json:"branch,omitempty"`
	Event  string `json:"event,omitempty"`
}

// ListActionRuns lists all workflow runs for a repository.
//
// GET /repos/{owner}/{repo}/actions/runs
func (c *Client) ListActionRuns(ctx context.Context, owner, repo string, opts ListActionRunsOptions) ([]*ActionRun, error) {
	var runs []*ActionRun
	query := opts.toQuery()
	if opts.Status != "" {
		query += "&status=" + opts.Status
	}
	if opts.Branch != "" {
		query += "&branch=" + opts.Branch
	}
	if opts.Event != "" {
		query += "&event=" + opts.Event
	}
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/actions/runs?%s", owner, repo, query), nil, &runs)
	if err != nil {
		return nil, err
	}
	return runs, nil
}

// GetActionRun gets a single workflow run.
//
// GET /repos/{owner}/{repo}/actions/runs/{id}
func (c *Client) GetActionRun(ctx context.Context, owner, repo string, runID int64) (*ActionRun, error) {
	var run ActionRun
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/actions/runs/%d", owner, repo, runID), nil, &run)
	if err != nil {
		return nil, err
	}
	return &run, nil
}

// RunActionWorkflowOptions specifies options for manually running a workflow.
type RunActionWorkflowOptions struct {
	Ref    string            `json:"ref"`
	Inputs map[string]string `json:"inputs,omitempty"`
}

// RunActionWorkflow manually triggers a workflow run.
//
// POST /repos/{owner}/{repo}/actions/run
func (c *Client) RunActionWorkflow(ctx context.Context, owner, repo string, opts RunActionWorkflowOptions) error {
	return c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/actions/run", owner, repo), opts, nil)
}

// --- Workflow Jobs ---

// ActionJob represents a workflow job.
type ActionJob struct {
	ID          int64         `json:"id"`
	RunID       int64         `json:"run_id"`
	Name        string        `json:"name"`
	Status      string        `json:"status"`
	Conclusion  string        `json:"conclusion,omitempty"`
	StartedAt   time.Time     `json:"started_at"`
	CompletedAt *time.Time    `json:"completed_at,omitempty"`
	RunnerID    int64         `json:"runner_id,omitempty"`
	RunnerName  string        `json:"runner_name,omitempty"`
	Steps       []*ActionStep `json:"steps,omitempty"`
}

// ActionStep represents a step in a workflow job.
type ActionStep struct {
	Name        string     `json:"name"`
	Status      string     `json:"status"`
	Conclusion  string     `json:"conclusion,omitempty"`
	Number      int        `json:"number"`
	StartedAt   time.Time  `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// ListActionJobs lists all jobs for a workflow run.
//
// GET /repos/{owner}/{repo}/actions/runs/{id}/jobs
func (c *Client) ListActionJobs(ctx context.Context, owner, repo string, runID int64, opts ListOptions) ([]*ActionJob, error) {
	var jobs []*ActionJob
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/actions/runs/%d/jobs?%s", owner, repo, runID, opts.toQuery()), nil, &jobs)
	if err != nil {
		return nil, err
	}
	return jobs, nil
}

// GetActionJob gets a single workflow job.
//
// GET /repos/{owner}/{repo}/actions/jobs/{id}
func (c *Client) GetActionJob(ctx context.Context, owner, repo string, jobID int64) (*ActionJob, error) {
	var job ActionJob
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/actions/jobs/%d", owner, repo, jobID), nil, &job)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

// --- Job Logs ---

// GetActionJobLogs gets the logs for a workflow job.
//
// GET /repos/{owner}/{repo}/actions/jobs/{id}/logs
func (c *Client) GetActionJobLogs(ctx context.Context, owner, repo string, jobID int64) ([]byte, error) {
	return c.doRawRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/actions/jobs/%d/logs", owner, repo, jobID))
}

// GetActionJobStepLogs gets the step-level logs for a workflow job.
//
// GET /repos/{owner}/{repo}/actions/jobs/{id}/logs/step
func (c *Client) GetActionJobStepLogs(ctx context.Context, owner, repo string, jobID int64) ([]byte, error) {
	return c.doRawRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/actions/jobs/%d/logs/step", owner, repo, jobID))
}

// --- Artifacts ---

// ActionArtifact represents a workflow artifact.
type ActionArtifact struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Size       int64     `json:"size"`
	ArchiveURL string    `json:"archive_download_url,omitempty"`
	Expired    bool      `json:"expired"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	RunID      int64     `json:"run_id,omitempty"`
}

// ListActionArtifacts lists all artifacts for a repository.
//
// GET /repos/{owner}/{repo}/actions/artifacts
func (c *Client) ListActionArtifacts(ctx context.Context, owner, repo string, opts ListOptions) ([]*ActionArtifact, error) {
	var artifacts []*ActionArtifact
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/actions/artifacts?%s", owner, repo, opts.toQuery()), nil, &artifacts)
	if err != nil {
		return nil, err
	}
	return artifacts, nil
}

// ListActionRunArtifacts lists all artifacts for a specific workflow run.
//
// GET /repos/{owner}/{repo}/actions/runs/{id}/artifacts
func (c *Client) ListActionRunArtifacts(ctx context.Context, owner, repo string, runID int64, opts ListOptions) ([]*ActionArtifact, error) {
	var artifacts []*ActionArtifact
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/actions/runs/%d/artifacts?%s", owner, repo, runID, opts.toQuery()), nil, &artifacts)
	if err != nil {
		return nil, err
	}
	return artifacts, nil
}

// GetActionArtifact gets a single artifact.
//
// GET /repos/{owner}/{repo}/actions/artifacts/{id}
func (c *Client) GetActionArtifact(ctx context.Context, owner, repo string, artifactID int64) (*ActionArtifact, error) {
	var artifact ActionArtifact
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/actions/artifacts/%d", owner, repo, artifactID), nil, &artifact)
	if err != nil {
		return nil, err
	}
	return &artifact, nil
}

// DeleteActionArtifact deletes an artifact.
//
// DELETE /repos/{owner}/{repo}/actions/artifacts/{id}
func (c *Client) DeleteActionArtifact(ctx context.Context, owner, repo string, artifactID int64) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/repos/%s/%s/actions/artifacts/%d", owner, repo, artifactID), nil, nil)
}

// DownloadActionArtifact downloads an artifact.
//
// GET /repos/{owner}/{repo}/actions/artifacts/{id}/download
func (c *Client) DownloadActionArtifact(ctx context.Context, owner, repo string, artifactID int64) ([]byte, error) {
	return c.doRawRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/actions/artifacts/%d/download", owner, repo, artifactID))
}

// --- YAML Validation ---

// ValidateWorkflowYAML validates a workflow YAML file.
//
// POST /repos/{owner}/{repo}/actions/validate
func (c *Client) ValidateWorkflowYAML(ctx context.Context, owner, repo string, yamlContent string) error {
	return c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/actions/validate", owner, repo), map[string]string{"yaml": yamlContent}, nil)
}

// --- Runners ---

// ActionRunner represents a CI/CD runner.
type ActionRunner struct {
	ID         int64    `json:"id"`
	Name       string   `json:"name"`
	Status     string   `json:"status"`
	OS         string   `json:"os,omitempty"`
	Arch       string   `json:"arch,omitempty"`
	Labels     []string `json:"labels,omitempty"`
	RunnerType string   `json:"runner_type,omitempty"`
}

// ListRepoHostRunners lists all host runners for a repository.
//
// GET /repos/{owner}/{repo}/actions/runners/host
func (c *Client) ListRepoHostRunners(ctx context.Context, owner, repo string, opts ListOptions) ([]*ActionRunner, error) {
	var runners []*ActionRunner
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/actions/runners/host?%s", owner, repo, opts.toQuery()), nil, &runners)
	if err != nil {
		return nil, err
	}
	return runners, nil
}

// ListRepoSharedHostRunners lists all shared host runners for a repository.
//
// GET /repos/{owner}/{repo}/actions/runners/shared-host
func (c *Client) ListRepoSharedHostRunners(ctx context.Context, owner, repo string, opts ListOptions) ([]*ActionRunner, error) {
	var runners []*ActionRunner
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/actions/runners/shared-host?%s", owner, repo, opts.toQuery()), nil, &runners)
	if err != nil {
		return nil, err
	}
	return runners, nil
}

// ListRepoK8SRunners lists all K8S runners for a repository.
//
// GET /repos/{owner}/{repo}/actions/runners/k8s
func (c *Client) ListRepoK8SRunners(ctx context.Context, owner, repo string, opts ListOptions) ([]*ActionRunner, error) {
	var runners []*ActionRunner
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/actions/runners/k8s?%s", owner, repo, opts.toQuery()), nil, &runners)
	if err != nil {
		return nil, err
	}
	return runners, nil
}

// ListRepoSharedK8SRunners lists all shared K8S runners for a repository.
//
// GET /repos/{owner}/{repo}/actions/runners/shared-k8s
func (c *Client) ListRepoSharedK8SRunners(ctx context.Context, owner, repo string, opts ListOptions) ([]*ActionRunner, error) {
	var runners []*ActionRunner
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/actions/runners/shared-k8s?%s", owner, repo, opts.toQuery()), nil, &runners)
	if err != nil {
		return nil, err
	}
	return runners, nil
}

// --- Runner Groups ---

// RunnerGroup represents a runner group.
type RunnerGroup struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Visibility  string `json:"visibility,omitempty"`
	RunnersURL  string `json:"runners_url,omitempty"`
}

// ListOrgRunnerGroups lists all runner groups for an organization.
//
// GET /orgs/{org}/actions/runner-groups
func (c *Client) ListOrgRunnerGroups(ctx context.Context, org string, opts ListOptions) ([]*RunnerGroup, error) {
	var groups []*RunnerGroup
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/orgs/%s/actions/runner-groups?%s", org, opts.toQuery()), nil, &groups)
	if err != nil {
		return nil, err
	}
	return groups, nil
}

// GetRunnerGroup gets a single runner group.
//
// GET /orgs/{org}/actions/runner-groups/{id}
func (c *Client) GetRunnerGroup(ctx context.Context, org string, groupID int64) (*RunnerGroup, error) {
	var group RunnerGroup
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/orgs/%s/actions/runner-groups/%d", org, groupID), nil, &group)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// ListRunnerGroupHostRunners lists all host runners in a runner group.
//
// GET /orgs/{org}/actions/runner-groups/{id}/host-runners
func (c *Client) ListRunnerGroupHostRunners(ctx context.Context, org string, groupID int64, opts ListOptions) ([]*ActionRunner, error) {
	var runners []*ActionRunner
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/orgs/%s/actions/runner-groups/%d/host-runners?%s", org, groupID, opts.toQuery()), nil, &runners)
	if err != nil {
		return nil, err
	}
	return runners, nil
}

// ListRunnerGroupK8SRunners lists all K8S runners in a runner group.
//
// GET /orgs/{org}/actions/runner-groups/{id}/k8s-runners
func (c *Client) ListRunnerGroupK8SRunners(ctx context.Context, org string, groupID int64, opts ListOptions) ([]*ActionRunner, error) {
	var runners []*ActionRunner
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/orgs/%s/actions/runner-groups/%d/k8s-runners?%s", org, groupID, opts.toQuery()), nil, &runners)
	if err != nil {
		return nil, err
	}
	return runners, nil
}

// ListRunnerGroupRepos lists all repositories that have access to a runner group.
//
// GET /orgs/{org}/actions/runner-groups/{id}/repos
func (c *Client) ListRunnerGroupRepos(ctx context.Context, org string, groupID int64, opts ListOptions) ([]*Repository, error) {
	var repos []*Repository
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/orgs/%s/actions/runner-groups/%d/repos?%s", org, groupID, opts.toQuery()), nil, &repos)
	if err != nil {
		return nil, err
	}
	return repos, nil
}

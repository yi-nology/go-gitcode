package gitcode

// Actions (流水线 / CI) client, speaking the documented GitCode **v8** wire
// (docs.gitcode.com, "Actions" category — the only version the platform
// serves; the module previously carried GitHub-shaped /api/v5 paths that
// 404 against real servers).
//
// The v8 domain model differs fundamentally from GitHub Actions: pipelines
// have stages → jobs → steps, IDs are strings, timestamps are epoch
// seconds, statuses are COMPLETED / RUNNING / FAILED / CANCELED / IGNORED /
// PAUSED / SUSPEND, and trigger events are MR / Push / Manual. JSON nulls
// unmarshal as zero values, so the nullable fields the docs show are plain
// Go types here.
//
// The v8 base URL is derived from the configured base URL by swapping the
// trailing /v5 for /v8 (see v8BaseURL); deployments must expose v8 at the
// same derived path.
//
// Undocumented variants are annotated at their methods: the shared/K8S
// runner lists reuse the documented repo-host-runners envelope, and the
// org runner-group *sub*-resources of the old surface are gone (their v8
// paths were never documented).

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// v8BaseURL derives the v8 API base from the configured base URL: the v8
// API lives next to v5 under the same host (…/api/v5 → …/api/v8). A base
// URL without a /v5 suffix is returned unchanged.
func (c *Client) v8BaseURL() string {
	base := c.BaseURL()
	if strings.HasSuffix(base, "/v5") {
		return strings.TrimSuffix(base, "/v5") + "/v8"
	}
	return base
}

func (c *Client) doV8Request(ctx context.Context, method, path string, body, result interface{}) error {
	return c.doRequestAt(ctx, c.v8BaseURL(), method, path, body, result)
}

// urlValuesFromListOptions renders page/per_page as escaped query values.
func urlValuesFromListOptions(o ListOptions) url.Values {
	opts := o.withDefaults()
	q := url.Values{}
	q.Set("page", strconv.Itoa(opts.Page))
	q.Set("per_page", strconv.Itoa(opts.PerPage))
	return q
}

// --- Workflows ---

// Workflow is one pipeline definition (v8: string IDs, state vocabulary).
type Workflow struct {
	WorkflowID string `json:"workflow_id"`
	FilePath   string `json:"file_path"`
	Name       string `json:"name"`
	State      string `json:"state"`
}

type ListWorkflowsResponse struct {
	TotalCount int         `json:"total_count"`
	Workflows  []*Workflow `json:"workflows"`
}

// ListActionWorkflows lists the repository's pipelines.
//
// GET /v8/repos/{owner}/{repo}/actions/workflows (page/per_page).
func (c *Client) ListActionWorkflows(ctx context.Context, owner, repo string, opts ListOptions) (*ListWorkflowsResponse, error) {
	var resp ListWorkflowsResponse
	err := c.doRequestAt(ctx, c.v8BaseURL(), http.MethodGet, fmt.Sprintf("/repos/%s/%s/actions/workflows?%s", owner, repo, opts.toQuery()), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// DispatchActionWorkflowOptions is the body of the manual-dispatch call.
// Ref is the branch or tag name; Inputs carries the pipeline's custom
// key-value parameters (documented required — always ships, as {} when nil).
type DispatchActionWorkflowOptions struct {
	Ref    string            `json:"ref"`
	Inputs map[string]string `json:"inputs"`
}

// DispatchActionWorkflowResult echoes the dispatched pipeline and run IDs.
type DispatchActionWorkflowResult struct {
	WorkflowID    string `json:"workflow_id"`
	WorkflowRunID string `json:"workflow_run_id"`
}

// DispatchActionWorkflow manually triggers a pipeline on ref. workflowID
// is the pipeline ID or a bare .yml/.yaml filename (not a full path).
//
// POST /v8/repos/{owner}/{repo}/actions/workflows/{workflow_id}/dispatches
// → 201 {workflow_id, workflow_run_id}.
func (c *Client) DispatchActionWorkflow(ctx context.Context, owner, repo, workflowID string, opts DispatchActionWorkflowOptions) (*DispatchActionWorkflowResult, error) {
	if opts.Inputs == nil {
		opts.Inputs = map[string]string{}
	}
	var resp DispatchActionWorkflowResult
	err := c.doV8Request(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/actions/workflows/%s/dispatches", owner, repo, workflowID), opts, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// WorkflowValidation is the result of the YAML validator.
type WorkflowPosition struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

type WorkflowDiagnosticRange struct {
	Start    WorkflowPosition `json:"start"`
	End      WorkflowPosition `json:"end"`
	Severity string           `json:"severity,omitempty"`
	Message  string           `json:"message,omitempty"`
}

type WorkflowDiagnostic struct {
	Range WorkflowDiagnosticRange `json:"range"`
}

type WorkflowValidation struct {
	Valid       bool                 `json:"valid"`
	Diagnostics []WorkflowDiagnostic `json:"diagnostics"`
}

// ValidateActionWorkflowYAML validates pipeline YAML content server-side
// (content is base64-encoded for the wire, per the documented contract).
//
// POST /v8/repos/{owner}/{repo}/actions/workflows/validate.
func (c *Client) ValidateActionWorkflowYAML(ctx context.Context, owner, repo string, yamlContent []byte) (*WorkflowValidation, error) {
	body := map[string]string{"base64_content": base64.StdEncoding.EncodeToString(yamlContent)}
	var resp WorkflowValidation
	err := c.doV8Request(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/actions/workflows/validate", owner, repo), body, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// --- Runs ---

// ActionActor is the trigger user on run objects (v8: string IDs).
type ActionActor struct {
	ID       string `json:"id"`
	ObjectID string `json:"object_id"`
	Login    string `json:"login"`
	Name     string `json:"name"`
}

// WorkflowRun is one pipeline run. Status: COMPLETED / RUNNING / FAILED /
// CANCELED / IGNORED / PAUSED / SUSPEND; event: MR / Push / Manual;
// timestamps are epoch seconds.
type WorkflowRun struct {
	WorkflowRunID string       `json:"workflow_run_id"`
	WorkflowID    string       `json:"workflow_id"`
	WorkflowName  string       `json:"workflow_name"`
	FilePath      string       `json:"file_path"`
	Title         string       `json:"title"`
	Status        string       `json:"status"`
	Event         string       `json:"event"`
	RunNumber     int          `json:"run_number"`
	HeadBranch    string       `json:"head_branch"`
	HeadSHA       string       `json:"head_sha"`
	Actor         *ActionActor `json:"actor"`
	StartTime     int64        `json:"start_time"`
	EndTime       int64        `json:"end_time"`
	PauseTime     int64        `json:"pause_time"`
}

type ListWorkflowRunsOptions struct {
	ListOptions
	Event         string `json:"event,omitempty"`  // MR / Push / Manual
	Status        string `json:"status,omitempty"` // run status vocabulary
	Branch        string `json:"branch,omitempty"`
	Executor      string `json:"executor,omitempty"`
	PullRequestID string `json:"pull_request_id,omitempty"`
	WorkflowID    string `json:"workflow_id,omitempty"`
	WorkflowName  string `json:"workflow_name,omitempty"`
	StartTime     int64  `json:"startTime,omitempty"` // epoch seconds
	EndTime       int64  `json:"endTime,omitempty"`
}

func (o ListWorkflowRunsOptions) toQuery() string {
	q := urlValuesFromListOptions(o.ListOptions)
	setIfNotEmpty := func(key, val string) {
		if val != "" {
			q.Set(key, val)
		}
	}
	setIfNotEmpty("event", o.Event)
	setIfNotEmpty("status", o.Status)
	setIfNotEmpty("branch", o.Branch)
	setIfNotEmpty("executor", o.Executor)
	setIfNotEmpty("pull_request_id", o.PullRequestID)
	setIfNotEmpty("workflow_id", o.WorkflowID)
	setIfNotEmpty("workflow_name", o.WorkflowName)
	if o.StartTime != 0 {
		q.Set("startTime", strconv.FormatInt(o.StartTime, 10))
	}
	if o.EndTime != 0 {
		q.Set("endTime", strconv.FormatInt(o.EndTime, 10))
	}
	return q.Encode()
}

type ListWorkflowRunsResponse struct {
	TotalCount   int            `json:"total_count"`
	WorkflowRuns []*WorkflowRun `json:"workflow_runs"`
}

// ListActionRuns lists pipeline runs with the documented filters.
//
// GET /v8/repos/{owner}/{repo}/actions/runs.
func (c *Client) ListActionRuns(ctx context.Context, owner, repo string, opts ListWorkflowRunsOptions) (*ListWorkflowRunsResponse, error) {
	var resp ListWorkflowRunsResponse
	err := c.doV8Request(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/actions/runs?%s", owner, repo, opts.toQuery()), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// WorkflowRunDetail is a run with its full stage → job → step tree. Pre
// steps hang off stages with their own (task-shaped) item model.
type ActionPreStep struct {
	Task           string   `json:"task"`
	Inputs         []string `json:"inputs"`
	ID             string   `json:"id"`
	JobRunID       string   `json:"job_run_id"`
	LastDispatchID string   `json:"last_dispatch_id"`
	Status         string   `json:"status"`
	EndTime        int64    `json:"end_time"`
}

type ActionStep struct {
	Name       string     `json:"name"`
	Task       string     `json:"task"`
	ID         string     `json:"id"`
	JobRunID   string     `json:"job_run_id"`
	Sequence   int        `json:"sequence"`
	Identifier string     `json:"identifier"`
	Status     string     `json:"status"`
	Message    string     `json:"message"`
	StartTime  int64      `json:"start_time"`
	EndTime    int64      `json:"end_time"`
	Inputs     []ActionKV `json:"inputs"`
	Env        []string   `json:"env"`
}

type ActionKV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ActionJob struct {
	ID              string        `json:"id"`
	Sequence        int           `json:"sequence"`
	Name            string        `json:"name"`
	Identifier      string        `json:"identifier"`
	DependsOn       []string      `json:"depends_on"`
	Condition       string        `json:"condition"`
	Resource        string        `json:"resource"`
	IsSelect        bool          `json:"is_select"`
	LastDispatchID  string        `json:"last_dispatch_id"`
	ExecuteCostTime int64         `json:"execute_cost_time"`
	Status          string        `json:"status"`
	Message         string        `json:"message"`
	StartTime       int64         `json:"start_time"`
	EndTime         int64         `json:"end_time"`
	ApplyTime       int64         `json:"apply_time"`
	ExecID          string        `json:"exec_id"`
	Steps           []*ActionStep `json:"steps"`
	HasSummary      bool          `json:"has_summary"`
	JobType         string        `json:"job_type"`
	SecretsInherit  bool          `json:"secrets_inherit"`
	IsShadow        bool          `json:"is_shadow"`
}

type ActionStage struct {
	ID             string           `json:"id"`
	Category       string           `json:"category"`
	Name           string           `json:"name"`
	Identifier     string           `json:"identifier"`
	Sequence       int              `json:"sequence"`
	DependsOn      []string         `json:"depends_on"`
	Status         string           `json:"status"`
	StartTime      int64            `json:"start_time"`
	EndTime        int64            `json:"end_time"`
	PauseTime      int64            `json:"pause_time"`
	Pre            []*ActionPreStep `json:"pre"`
	Post           []string         `json:"post"`
	Jobs           []*ActionJob     `json:"jobs"`
	LastDispatchID string           `json:"last_dispatch_id"`
}

type WorkflowRunDetail struct {
	WorkflowRun
	Attempt              int            `json:"attempt"`
	ExistInDefaultBranch bool           `json:"exist_in_default_branch"`
	Stages               []*ActionStage `json:"stages"`
}

// GetActionRun fetches one run with its full stage/job/step tree.
//
// GET /v8/repos/{owner}/{repo}/actions/runs/{run_id}.
func (c *Client) GetActionRun(ctx context.Context, owner, repo, runID string) (*WorkflowRunDetail, error) {
	var resp WorkflowRunDetail
	err := c.doV8Request(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/actions/runs/%s", owner, repo, runID), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// stopRetryRerunResult is the {success} body of the run-mutation verbs.
type stopRetryRerunResult struct {
	Success bool `json:"success"`
}

// StopActionRun stops a running pipeline.
//
// POST /v8/repos/{owner}/{repo}/actions/runs/{run_id}/stop → {success}.
func (c *Client) StopActionRun(ctx context.Context, owner, repo, runID string) (bool, error) {
	var resp stopRetryRerunResult
	err := c.doV8Request(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/actions/runs/%s/stop", owner, repo, runID), nil, &resp)
	return resp.Success, err
}

// RetryActionRun retries a failed run.
//
// POST /v8/repos/{owner}/{repo}/actions/runs/{run_id}/retry → {success}.
func (c *Client) RetryActionRun(ctx context.Context, owner, repo, runID string) (bool, error) {
	var resp stopRetryRerunResult
	err := c.doV8Request(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/actions/runs/%s/retry", owner, repo, runID), nil, &resp)
	return resp.Success, err
}

// RerunActionRun reruns a pipeline instance.
//
// POST /v8/repos/{owner}/{repo}/actions/runs/{run_id}/rerun → {success}.
func (c *Client) RerunActionRun(ctx context.Context, owner, repo, runID string) (bool, error) {
	var resp stopRetryRerunResult
	err := c.doV8Request(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/actions/runs/%s/rerun", owner, repo, runID), nil, &resp)
	return resp.Success, err
}

// --- Jobs & logs ---

type ListActionJobsResponse struct {
	TotalCount int          `json:"total_count"`
	Jobs       []*ActionJob `json:"jobs"`
}

// ListActionRunJobs lists a run's jobs with their steps.
//
// GET /v8/repos/{owner}/{repo}/actions/runs/{run_id}/jobs.
func (c *Client) ListActionRunJobs(ctx context.Context, owner, repo, runID string) (*ListActionJobsResponse, error) {
	var resp ListActionJobsResponse
	err := c.doV8Request(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/actions/runs/%s/jobs", owner, repo, runID), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetActionJob fetches one job of a run (jobs are addressed as
// runs/{run_id}/jobs/{job_id} on the v8 wire).
//
// GET /v8/repos/{owner}/{repo}/actions/runs/{run_id}/jobs/{job_id}.
func (c *Client) GetActionJob(ctx context.Context, owner, repo, runID, jobID string) (*ActionJob, error) {
	var job ActionJob
	err := c.doV8Request(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/actions/runs/%s/jobs/%s", owner, repo, runID, jobID), nil, &job)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

// GetActionJobLogsOptions drives the incremental log pagination: the
// first page uses Offset 0; forward paging continues from the previous
// EndOffset, reverse paging from StartOffset. Sort is "asc" or "desc".
type GetActionJobLogsOptions struct {
	StepID string `json:"step_id"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
	Sort   string `json:"sort"`
}

// ActionJobLogs is one page of job (or step) logs: Log covers
// [StartOffset, EndOffset); HasMore reports whether the window continues.
type ActionJobLogs struct {
	HasMore     bool   `json:"has_more"`
	StartOffset int    `json:"start_offset"`
	EndOffset   int    `json:"end_offset"`
	Log         string `json:"log"`
}

// GetActionJobLogs fetches one page of job/step logs.
//
// POST /v8/repos/{owner}/{repo}/actions/runs/{run_id}/jobs/{job_id}/logs.
func (c *Client) GetActionJobLogs(ctx context.Context, owner, repo, runID, jobID string, opts GetActionJobLogsOptions) (*ActionJobLogs, error) {
	var resp ActionJobLogs
	err := c.doV8Request(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/actions/runs/%s/jobs/%s/logs", owner, repo, runID, jobID), opts, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// DownloadActionJobLogs downloads a job's full log archive.
//
// GET /v8/repos/{owner}/{repo}/actions/runs/{run_id}/jobs/{job_id}/download-log.
func (c *Client) DownloadActionJobLogs(ctx context.Context, owner, repo, runID, jobID string) ([]byte, error) {
	return c.doRawRequestAt(ctx, c.v8BaseURL(), http.MethodGet,
		fmt.Sprintf("/repos/%s/%s/actions/runs/%s/jobs/%s/download-log", owner, repo, runID, jobID))
}

// --- Artifacts ---

// ActionArtifact is one pipeline artifact. The *_at fields are millisecond
// timestamp strings as served; Digest is the artifact SHA256.
type ActionArtifact struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	SizeBytes     int64  `json:"size_bytes"`
	WorkflowID    string `json:"workflow_id"`
	WorkflowRunID string `json:"workflow_run_id"`
	Digest        string `json:"digest"`
	ExpiresAt     string `json:"expires_at"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type ListActionArtifactsOptions struct {
	ListOptions
	Name      string `json:"name,omitempty"`      // fuzzy filter
	Sort      string `json:"sort,omitempty"`      // only "created" documented
	Direction string `json:"direction,omitempty"` // asc / desc
}

func (o ListActionArtifactsOptions) toQuery() string {
	q := urlValuesFromListOptions(o.ListOptions)
	if o.Name != "" {
		q.Set("name", o.Name)
	}
	if o.Sort != "" {
		q.Set("sort", o.Sort)
	}
	if o.Direction != "" {
		q.Set("direction", o.Direction)
	}
	return q.Encode()
}

type ListActionArtifactsResponse struct {
	TotalCount int               `json:"total_count"`
	Artifacts  []*ActionArtifact `json:"artifacts"`
}

// ListActionArtifacts lists the repository's artifacts (name fuzzy filter,
// created sort).
//
// GET /v8/repos/{owner}/{repo}/actions/artifacts.
func (c *Client) ListActionArtifacts(ctx context.Context, owner, repo string, opts ListActionArtifactsOptions) (*ListActionArtifactsResponse, error) {
	var resp ListActionArtifactsResponse
	err := c.doV8Request(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/actions/artifacts?%s", owner, repo, opts.toQuery()), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListActionRunArtifacts lists one run's artifacts.
//
// GET /v8/repos/{owner}/{repo}/actions/runs/{run_id}/artifacts.
func (c *Client) ListActionRunArtifacts(ctx context.Context, owner, repo, runID string, opts ListOptions) (*ListActionArtifactsResponse, error) {
	var resp ListActionArtifactsResponse
	err := c.doV8Request(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/actions/runs/%s/artifacts?%s", owner, repo, runID, opts.toQuery()), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetActionArtifact fetches one artifact's details.
//
// GET /v8/repos/{owner}/{repo}/actions/artifacts/{artifact_id}.
func (c *Client) GetActionArtifact(ctx context.Context, owner, repo, artifactID string) (*ActionArtifact, error) {
	var artifact ActionArtifact
	err := c.doV8Request(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/actions/artifacts/%s", owner, repo, artifactID), nil, &artifact)
	if err != nil {
		return nil, err
	}
	return &artifact, nil
}

// DeleteActionArtifact deletes an artifact.
//
// DELETE /v8/repos/{owner}/{repo}/actions/artifacts/{artifact_id}.
func (c *Client) DeleteActionArtifact(ctx context.Context, owner, repo, artifactID string) error {
	return c.doV8Request(ctx, http.MethodDelete, fmt.Sprintf("/repos/%s/%s/actions/artifacts/%s", owner, repo, artifactID), nil, nil)
}

// DownloadActionArtifact downloads an artifact archive in format (the
// documented path segment; "zip" is the platform's format).
//
// GET /v8/repos/{owner}/{repo}/actions/artifacts/{artifact_id}/{archive_format}.
func (c *Client) DownloadActionArtifact(ctx context.Context, owner, repo, artifactID, format string) ([]byte, error) {
	return c.doRawRequestAt(ctx, c.v8BaseURL(), http.MethodGet,
		fmt.Sprintf("/repos/%s/%s/actions/artifacts/%s/%s", owner, repo, artifactID, format))
}

// --- Runners ---

// ActionRunnerLabel is one runner capability label.
type ActionRunnerLabel struct {
	LabelName  string `json:"label_name"`
	LabelValue string `json:"label_value"`
	LabelColor string `json:"label_color"`
}

// ActionRunner is a pipeline runner. Status: IDLE / BUSY / OFFLINE /
// DISABLE; memory is percent used, disk is GB free.
type ActionRunner struct {
	ID            string              `json:"id"`
	RunnerGroupID string              `json:"runner_group_id"`
	RunnerName    string              `json:"runner_name"`
	Name          string              `json:"name"`
	WorkDir       string              `json:"work_dir"`
	Labels        []ActionRunnerLabel `json:"labels"`
	Status        string              `json:"status"`
	CreateTime    int64               `json:"create_time"`
	UpdateTime    int64               `json:"update_time"`
	Memory        float64             `json:"memory"`
	Disk          float64             `json:"disk"`
}

type ListActionRunnersOptions struct {
	ListOptions
	Keyword string `json:"keyword,omitempty"`
}

func (o ListActionRunnersOptions) toQuery() string {
	q := urlValuesFromListOptions(o.ListOptions)
	if o.Keyword != "" {
		q.Set("keyword", o.Keyword)
	}
	return q.Encode()
}

type ListActionRunnersResponse struct {
	TotalCount int             `json:"total_count"`
	Runners    []*ActionRunner `json:"runners"`
}

// ListRepoActionRunners lists the repository's host runners.
//
// GET /v8/repos/{owner}/{repo}/actions/runners.
func (c *Client) ListRepoActionRunners(ctx context.Context, owner, repo string, opts ListActionRunnersOptions) (*ListActionRunnersResponse, error) {
	var resp ListActionRunnersResponse
	err := c.doV8Request(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/actions/runners?%s", owner, repo, opts.toQuery()), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// listRepoRunnersVariant hits one of the documented runner list variants
// (shared host runners / K8S runner sets / shared K8S runner sets). The
// platform documents these paths under the same Actions category; their
// envelopes mirror the repo host-runners shape.
func (c *Client) listRepoRunnersVariant(ctx context.Context, owner, repo, suffix string, opts ListActionRunnersOptions) (*ListActionRunnersResponse, error) {
	var resp ListActionRunnersResponse
	err := c.doV8Request(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/actions/%s?%s", owner, repo, suffix, opts.toQuery()), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListRepoSharedActionRunners lists host runners shared with the repository.
//
// GET /v8/repos/{owner}/{repo}/actions/runners/shared-runners.
func (c *Client) ListRepoSharedActionRunners(ctx context.Context, owner, repo string, opts ListActionRunnersOptions) (*ListActionRunnersResponse, error) {
	return c.listRepoRunnersVariant(ctx, owner, repo, "runners/shared-runners", opts)
}

// ListRepoK8SRunners lists the repository's K8S runner sets.
//
// GET /v8/repos/{owner}/{repo}/actions/runner-sets.
func (c *Client) ListRepoK8SRunners(ctx context.Context, owner, repo string, opts ListActionRunnersOptions) (*ListActionRunnersResponse, error) {
	return c.listRepoRunnersVariant(ctx, owner, repo, "runner-sets", opts)
}

// ListRepoSharedK8SRunners lists K8S runner sets shared with the repository.
//
// GET /v8/repos/{owner}/{repo}/actions/shared-runner-sets.
func (c *Client) ListRepoSharedK8SRunners(ctx context.Context, owner, repo string, opts ListActionRunnersOptions) (*ListActionRunnersResponse, error) {
	return c.listRepoRunnersVariant(ctx, owner, repo, "shared-runner-sets", opts)
}

// ActionRunnerGroup is an org runner group.
type ActionRunnerGroup struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	RunnerGroupName string `json:"runner_group_name"`
	NamespaceID     string `json:"namespace_id"`
	Creator         string `json:"creator"`
	CreateTime      int64  `json:"create_time"`
	RunnerCount     int    `json:"runner_count"`
	NamespaceType   string `json:"namespace_type"`
	ShareAll        bool   `json:"share_all"`
}

type ListOrgRunnerGroupsResponse struct {
	TotalCount   int                  `json:"total_count"`
	RunnerGroups []*ActionRunnerGroup `json:"runner_groups"`
}

// ListOrgActionRunnerGroups lists the organization's runner groups.
//
// GET /v8/orgs/{org}/actions/runner-groups.
func (c *Client) ListOrgActionRunnerGroups(ctx context.Context, org string, opts ListActionRunnersOptions) (*ListOrgRunnerGroupsResponse, error) {
	var resp ListOrgRunnerGroupsResponse
	err := c.doV8Request(ctx, http.MethodGet, fmt.Sprintf("/orgs/%s/actions/runner-groups?%s", org, opts.toQuery()), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetActionRunnerGroup fetches one org runner group.
//
// GET /v8/orgs/{org}/actions/runner-groups/{group_id}.
func (c *Client) GetActionRunnerGroup(ctx context.Context, org, groupID string) (*ActionRunnerGroup, error) {
	var group ActionRunnerGroup
	err := c.doV8Request(ctx, http.MethodGet, fmt.Sprintf("/orgs/%s/actions/runner-groups/%s", org, groupID), nil, &group)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

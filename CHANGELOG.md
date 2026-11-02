# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.10.0] - 2026-11-02

### Fixed / Changed (breaking)
- **Actions/CI 面整体重写为 v8 wire（对照 docs.gitcode.com "Actions" 分类）**。
  GitCode 的 Actions API 只有 v8 版本——旧面的 GitHub 形状 `/api/v5/...`
  路径对真实服务只会 404（同 org_hooks 的虚构形态，但端点本身有文档、
  因此重写而非删除）。v8 是完全不同的领域模型：流水线 stages→jobs→steps
  树、字符串 ID、秒级时间戳、状态词表 COMPLETED/RUNNING/FAILED/CANCELED/
  IGNORED/PAUSED/SUSPEND、触发事件 MR/Push/Manual。
  - v8 base 由配置 base 自动派生（`…/api/v5` → `…/api/v8`，见
    `v8BaseURL`；无 /v5 后缀的自有部署按原样透传并在文档注明前提）；
  - 方法面（22 个）：`ListActionWorkflows`（信封 `{total_count, workflows}`）、
    `DispatchActionWorkflow`（替代 `RunActionWorkflow`，body
    `{ref, inputs}`，inputs 文档 required 恒发）、`ValidateActionWorkflowYAML`
    （`[]byte` 自动 base64；替代 `ValidateWorkflowYAML`）、
    `ListActionRuns`（event/status/branch/executor/pull_request_id/
    workflow_id/workflow_name/startTime/endTime 过滤）、`GetActionRun`
    （含 stage 树）、`StopActionRun`/`RetryActionRun`/`RerunActionRun`
    （`{success}` 回包）、`ListActionRunJobs`、`GetActionJob`（v8 把 job
    挂在 runs/{run_id}/jobs/{job_id} 下——旧 `/actions/jobs/{id}` 路径
    不存在）、`GetActionJobLogs`（增量分页 body
    `{step_id, offset, limit, sort}`；替代旧 `GetActionJobLogs`/
    `GetActionJobStepLogs`）、`DownloadActionJobLogs`、
    `ListActionArtifacts`（name 模糊过滤/sort/direction）、
    `ListActionRunArtifacts`、`GetActionArtifact`、`DeleteActionArtifact`、
    `DownloadActionArtifact`（`{archive_format}` 路径段）、
    `ListRepoActionRunners`/`ListRepoSharedActionRunners`/
    `ListRepoK8SRunners`/`ListRepoSharedK8SRunners`、
    `ListOrgActionRunnerGroups`/`GetActionRunnerGroup`；
  - **移除**：旧 `ListRunnerGroupHostRunners`/`ListRunnerGroupK8SRunners`/
    `ListRunnerGroupRepos`（v8 从未文档化这些子资源路径，不留猜测形状）；
    shared/K8S runner 变体沿用 repo host-runners 信封（注释标注为同信封
    假设，schema 页未单列）。
  - client 执行核心去重：`doRequestAt(base,…)`/`doRawRequestAt(base,…)`
    为 v5/v8 共用（旧 `doRequestWithHeaders` 随之删除）。
  - 22 个 httptest 用例钉住 v8 base 派生、路径、信封解析（含大量 null
    字段的容错）、关键请求体（dispatch inputs 恒发、validate base64、
    日志分页 body）。

### Migration notes (v0.9.x → v0.10.0)
- Actions 旧类型/方法（GitHub 形状 `ActionRun{ID int64, Conclusion}` 等）
  → v8 类型（`WorkflowRun{WorkflowRunID string, Status}` 等）。
  `RunActionWorkflow` → `DispatchActionWorkflow`；`ValidateWorkflowYAML`
  → `ValidateActionWorkflowYAML`；`ListActionJobs`/`GetActionJob` →
  `ListActionRunJobs`/`GetActionJob`（增加 runID 参数）；
  runner-group 子资源三方法无 v8 等价物。

## [v0.9.0] - 2026-11-02

### Removed
- **`org_hooks.go` (organization webhooks) deleted entirely.** Neither the
  documented GitCode v5 API (docs.gitcode.com/docs/apis lists only
  repo-scoped `repos/{owner}/{repo}/hooks`) nor the companion AtomGit API
  docs have an `orgs/{org}/hooks` endpoint — the module was a GitHub-shaped
  invention (config/secret/events, the same wire fictions removed from the
  repo-level webhook in v0.8.0) that could only ever 404 against real
  servers. Removing beats shipping a plausible-looking lie.

### Migration notes (v0.8.x → v0.9.0)
- `ListOrgWebhooks` / `GetOrgWebhook` / `CreateOrgWebhook` /
  `UpdateOrgWebhook` / `DeleteOrgWebhook`, the `OrgWebhook` /
  `WebhookConfig` / `CreateOrgWebhookOptions` / `UpdateOrgWebhookOptions`
  types are gone. There is no org-level hook API to migrate to: use
  repo-level webhooks (`ListWebhooksWithOptions` family), or call the
  endpoint raw if a private deployment actually exposes one.

## [v0.8.0] - 2026-11-02

Wire-shape corrections against the documented GitCode v5 API
(docs.gitcode.com), driven by the needs of go-git-platform's gitcode
backend. 0.x carries the breaking renames in a minor per policy.

### Fixed
- **Webhook secret never took effect**: create/update bodies carried
  `secret`, but the documented wire field is `password` (plus
  `encryption_type`, 0 = password / 1 = signature key). `Secret` is
  replaced by `Password`.
- **Webhook event subscriptions were fictional**: GitCode's hook API has
  no `events` array — event families are the per-family flags
  `push_events` / `tag_push_events` / `issues_events` / `note_events` /
  `merge_requests_events`. The `Events []string` field is removed from
  `Webhook`, `CreateWebhookOptions`, and `UpdateWebhookOptions`; the
  options now expose `*bool` flags (`nil` = unchanged, `Bool(false)` =
  family off) and the response flags use the tolerant `FlexBool`
  (accepts 0/1, true/false, strings) since the documented response
  schema types them as integers.
- **Milestone wire shape**: the response identifies milestones by
  `number` and carries `due_on` as a date-only string ("2026-01-31").
  `Milestone` gains `Number` / `RepositoryID` / `URL` / `OpenIssues` /
  `ClosedIssues`; `DueDate *time.Time` (`json:"due_date"`, never
  populated and unable to parse date-only values) is replaced by
  `DueOn string` (`json:"due_on"`). Create/update option structs now
  marshal `due_on` with `omitempty`, so an unset due date stays off the
  wire instead of going out as `""` (which clears or errors).
- **Query injection through unescaped values**: filter values (labels
  with spaces, `since` timestamps containing `+`, ...) are now escaped
  via `url.Values` in `ListIssues` and `ListMilestonesWithOptions`
  instead of being spliced raw into the query string.
- `.golangci.yml` migrated to golangci-lint v2 config syntax.

### Added
- `ListIssueCommentsOptions{ListOptions, Order, Since}` and
  `ListIssueCommentsWithOptions` — the issue-comment list endpoint takes
  `page` / `per_page` / `order` / `since`; `ListIssueComments` remains
  as the first-page convenience wrapper.
- `ListIssueLabelsWithOptions(ListOptions)` and
  `ListWebhooksWithOptions(ListOptions)` for paginating the repository
  label and hook lists (both default to 100 per page server-side).
- `Bool(v bool) *bool` helper for the webhook flag fields, and
  `EventFlags` mapping neutral event names ("push", "pull_request",
  "issues", "note", "comment", "tag_push") onto the flag fields.
- `FlexBool` tolerant boolean type.
- `Milestone` now also surfaces the documented `state` counter fields
  (`open_issues` / `closed_issues`).

### Migration notes (v0.7.x → v0.8.0)
- `CreateWebhookOptions.Secret` / `UpdateWebhookOptions.Secret` →
  `Password`.
- `Webhook.Events` / `CreateWebhookOptions.Events` /
  `UpdateWebhookOptions.Events` → per-family flags
  (`PushEvents` etc.); map neutral names via `EventFlags`.
- `Milestone.DueDate` (`*time.Time`) → `Milestone.DueOn` (`string`,
  date-only as served); milestone identity: prefer `Number` (falls back
  to `ID` on older deployments).

## [v0.7.2] - 2026-08-25

### Fixed
- CI: all source files formatted with `gofmt -s` (30 files)
- CI: golangci-lint now built from source (`goinstall` mode) for Go 1.26 compatibility
- Replaced direct type assertions with `errors.As` in all `Is*` error check functions (`errors.go`, `validate.go`)
- Removed unused `hookMu` type and `sync` import from `hooks.go`
- Suppressed false-positive lint warnings on API field names (`noteable_type`, `clas` endpoint)

## [v0.7.1] - 2026-08-25

### Added
- `SetBaseURL()` / `BaseURL()` on `Client` for runtime base URL switching
- `NewClientFromEnv()` constructor — reads `GITCODE_URL` and `GITCODE_TOKEN` environment variables
- `NewOAuthClientWithBaseURL()` for creating OAuth clients pointing to private instances
- `OAuthClient.NewClient(token)` method — preserves the OAuth client's base URL (fixes lost URL bug in `NewClientFromOAuthToken`)
- `EnvGitBaseURL` / `EnvGitToken` exported constants for environment variable names

### Changed
- Module path renamed from `github.com/yi-nology/gitcode_api` to `github.com/yi-nology/go-gitcode`
- `getTestClient()` in tests now respects `GITCODE_URL` env var for private instance testing
- README and examples updated with private deployment documentation

### Deprecated
- `NewClientFromOAuthToken()` — use `OAuthClient.NewClient()` instead to preserve base URL

## [v0.7.0] - 2026-08-15

### Added
- Comprehensive mock HTTP server tests for all 30+ previously untested modules
- `test_helpers_test.go` with shared test utilities (`newTestServer`, `jsonResponse`, etc.)
- `Makefile` with build/test/lint/vet/fmt/tidy/check/ci/release-check targets
- `.golangci.yml` with comprehensive lint configuration
- `CHANGELOG.md` with version history
- `.github/workflows/ci.yml` GitHub Actions CI pipeline

### Test Coverage
- Added ~200 new unit tests covering all API modules
- Tests use `httptest.NewServer` for mock HTTP responses
- Each module now has dedicated test file with CRUD operation coverage

## [v0.6.0] - 2026-08-15

### Added
- `errors.go`: Structured error types (`NotFoundError`, `UnauthorizedError`, `RateLimitError`, `ConflictError`, `ForbiddenError`, `ValidationError`) with type-check helpers (`IsNotFound`, `IsUnauthorized`, `IsRateLimit`, etc.)
- `retry.go`: Automatic retry with exponential backoff for 429/5xx errors, configurable `RetryPolicy` (MaxRetries, InitialBackoff, MaxBackoff, Multiplier)
- `pagination.go`: Generic auto-pagination iterator using Go 1.23+ `iter` package (`CollectAll`, `Paginate`)
- `hooks.go`: Request/response middleware hooks (`AddRequestHook`, `AddResponseHook`, `ClearHooks`)
- `validate.go`: Client-side input validation for required fields (`owner`, `repo`)
- `webhook_verify.go`: HMAC-SHA256 webhook signature verification (`VerifyWebhookSignature`, `ComputeWebhookSignature`)
- `upload.go`: Multipart file upload via `io.Reader` (`UploadFileReader`, `UploadImageReader`, `UploadFileBytes`)
- `client_enhanced_test.go`: Unit tests for all new infrastructure features
- Updated `examples/main.go` with comprehensive usage examples

### Changed
- `client.go`: Integrated `sync.RWMutex` for concurrent safety on `SetAuthStyle`, `SetHTTPClient`, hooks
- All HTTP error responses now return structured error types instead of plain `fmt.Errorf`
- All HTTP request methods now run request/response hooks and support retry policy

## [v0.5.0] - 2026-08-15

### Added
- `commits_enhanced.go`: Commit diff/patch endpoints (2)
- `repos_protected_tags.go`: Protected tags CRUD (5)
- `releases_enhanced.go`: Latest release, upload URL, download asset (3)
- `enterprise.go`: Enterprise members, milestones, labels, roles, custom fields (14)
- `dashboard_kanban.go`: Kanban boards CRUD, board items, board status (7)
- `actions.go`: CI/CD workflows, runs, jobs, logs, artifacts, runners, runner groups (23)

### Coverage
- 100% coverage of official API documentation at docs.gitcode.com/docs/apis/
- 428 Client methods across 48 source files

## [v0.4.0] - 2026-08-15

### Added
- `repos_discussions.go`: Repository discussions, fork sync, remote mirrors, license, CLA (11)
- `issues_enhanced.go`: Issue related branches, kanban values, modify history, enterprise statuses (8)
- `pulls_enhanced.go`: PR linked issues, tester/reviewer management, discussions, modify history (15)
- `users_enhanced.go`: Watched repos, update profile, user PR list (4)
- `orgs_enhanced.go`: Organization customized roles, discussions (5)

### Coverage
- Cross-referenced with official docs.gitcode.com/docs/apis/ sidebar

## [v0.3.0] - 2026-08-15

### Added
- 22 new source files with major API expansion
- Repository collaborators, topics, commit statuses, deploy keys
- Git references, reactions, wiki, user followers
- Organization teams, hooks, labels, public members, block/unblock
- Enhanced notifications, repository invitations
- Issue/PR templates, commit comments, repository archive/stats
- Annotated tags, release assets, PR reviewers, diff/patch
- Gitignore/license/label templates, markdown rendering

### Coverage
- 330 Client methods across 36 source files

## [v0.2.0] - 2026-08-15

### Added
- Initial API client implementation
- Authentication (Bearer, PRIVATE-TOKEN, access_token, OAuth 2.0)
- Repository CRUD, file operations, tree/blob, tags, releases
- Issue CRUD, comments, labels, milestones
- Pull request CRUD, merge, reviews, comments
- Branch management, protection rules, commits
- Webhook CRUD with event parsing
- User management, SSH keys, emails
- Organization management, enterprise members
- Search (repositories, issues, users)
- Star/watch/fork operations
- Comprehensive README documentation

### Coverage
- ~150 Client methods across 14 source files

## [v0.1.0] - 2026-08-15

### Added
- Initial project setup
- Basic client structure

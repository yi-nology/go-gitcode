# GitCode API Client

Go 语言的 [GitCode](https://gitcode.com) / AtomGit API 客户端库,提供对 GitCode 平台几乎所有 REST API (`/api/v5`) 的类型安全访问。

[![Go Reference](https://pkg.go.dev/badge/github.com/yi-nology/go-gitcode.svg)](https://pkg.go.dev/github.com/yi-nology/go-gitcode)
[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](https://go.dev)

## 功能特性

- **认证** — Bearer Token / PRIVATE-TOKEN Header / `access_token` Query 三种鉴权 + OAuth 2.0 授权码流程
- **仓库** — 创建/更新/删除/Fork/归档/转让,组织仓库,文件 CRUD,Tree/Blob,Raw,文件&图片上传,Fork 同步,远程镜像,许可协议,CLA
- **Issue** — 仓库 Issue CRUD、评论、标签、里程碑、操作日志、关联 PR、时间线、订阅者、依赖关系、关联分支、看板字段、修改历史;用户/组织/企业级 Issue 查询
- **Pull Request** — CRUD、合并、文件/提交/评论/审查、标签、审查人/测试人分配、操作日志、关联 Issue、Diff/Patch、讨论回复、检视意见、修改历史
- **讨论** — 仓库/组织级讨论、评论、回复
- **分支** — 列表/创建/删除,保护分支规则,提交比较,提交历史
- **Webhook** — CRUD、测试推送,Push/TagPull/Issue/PullRequest/Note 事件解析
- **用户** — 当前用户/指定用户、SSH 公钥、邮箱、动态、Star/Watch 仓库、Namespace、关注/取关、更新资料、用户 PR 列表
- **组织 / 企业** — 组织信息/成员/关注者/公开成员/屏蔽用户,企业成员、企业 Issue/PR、Issue 扩展状态、企业标签、组织自定义角色、组织讨论
- **组织团队** — 团队 CRUD、成员管理、仓库关联
- **组织 Webhook** — 组织级 Webhook CRUD
- **组织标签** — 组织级标签 CRUD
- **搜索** — 仓库 / Issue / 用户
- **协作者** — 仓库协作者 CRUD、权限查询
- **Topics** — 仓库主题管理
- **Commit Status** — 提交状态、合并状态 (CI/CD 集成)
- **Deploy Keys** — 部署密钥 CRUD
- **Git References** — Git 引用 CRUD
- **Git Tags** — 轻量/注释标签、Release 资产管理
- **Reactions** — Issue/评论/PR 表情回应
- **Wiki** — Wiki 页面 CRUD
- **通知** — 增强通知操作(标记已读、线程详情、仓库通知)
- **仓库邀请** — 接受/拒绝仓库邀请
- **模板** — Issue 模板、PR 合并模板、Gitignore/License/Label 模板
- **Markdown** — Markdown 渲染
- **仓库统计** — 参与度、代码频率、提交活动、Punch Card
- **仓库归档** — 下载仓库压缩包
- **仓库设置** — Push 配置、PR 设置、模块开关、审查规则、下载统计
- **保护标签** — 保护标签规则 CRUD
- **Release 增强** — 最新 Release、上传 URL、下载附件
- **企业** — 企业成员/里程碑/标签/自定义角色/Issue 自定义字段/组织关联企业
- **看板 (Dashboard)** — 组织看板 CRUD、看板条目管理
- **Actions/CI** — 工作流/运行记录/Jobs/日志/Artifacts/Runners/Runner Groups
- **类型友好** — `FlexInt`/`FlexString`/`NullableTime` 等 JSON 容错类型,适配 GitCode 返回值类型漂移

## 环境要求

- Go ≥ 1.26(`go.mod` 已锁定)
- 一个 GitCode 个人访问令牌(PAT)或 OAuth 应用凭据

## 安装

```bash
go get github.com/yi-nology/go-gitcode@latest
```

当前最新版本为 `v0.8.0`,完整版本历史见 [releases](https://github.com/yi-nology/go-gitcode/releases)。

## 快速开始

```go
package main

import (
    "context"
    "fmt"
    "log"

    gitcode "github.com/yi-nology/go-gitcode"
)

func main() {
    client := gitcode.NewClient("your-gitcode-token")
    ctx := context.Background()

    user, err := client.GetCurrentUser(ctx)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("当前用户: %s (%s)\n", user.Name, user.Login)

    repos, err := client.ListRepositories(ctx, gitcode.ListRepositoriesOptions{
        ListOptions: gitcode.ListOptions{Page: 1, PerPage: 10},
    })
    if err != nil {
        log.Fatal(err)
    }
    for _, r := range repos {
        fmt.Printf("- %s\n", r.FullName)
    }
}
```

更完整的示例参见 [`examples/main.go`](examples/main.go)。

## 客户端与认证

### 获取 Token

访问 [GitCode → 个人设置 → 私人令牌](https://gitcode.com/profile/personal_access_tokens) 创建 Token。

### 三种鉴权方式

`Client` 默认使用 `Authorization: Bearer <token>`,可通过 `SetAuthStyle` 切换:

```go
client := gitcode.NewClient("your-token")                       // Bearer (默认)
client.SetAuthStyle(gitcode.AuthStylePrivateToken)              // PRIVATE-TOKEN Header
client.SetAuthStyle(gitcode.AuthStyleAccessToken)               // access_token Query 参数
```

### 私有部署 / 自定义 HTTP 客户端

#### 方式一：构造时指定

```go
client := gitcode.NewClientWithBaseURL("https://your-gitcode.com/api/v5", "token")
client.SetHTTPClient(&http.Client{Timeout: 60 * time.Second})   // 自定义超时
```

#### 方式二：构造后修改

```go
client := gitcode.NewClient("token")
client.SetBaseURL("https://your-gitcode.com/api/v5")            // 切换到私有实例
fmt.Println(client.BaseURL())                                    // 查看当前 URL
```

#### 方式三：环境变量（推荐用于 CLI / CI）

设置以下环境变量后，使用 `NewClientFromEnv()` 自动读取:

| 环境变量 | 说明 | 示例 |
|----------|------|------|
| `GITCODE_URL` | 私有实例 API 地址 | `https://gitcode.example.com/api/v5` |
| `GITCODE_TOKEN` | 认证 Token | `your-token` |

```bash
export GITCODE_URL="https://gitcode.example.com/api/v5"
export GITCODE_TOKEN="your-token"
```

```go
client := gitcode.NewClientFromEnv()  // 自动读取环境变量
if client == nil {
    log.Fatal("请设置 GITCODE_TOKEN 环境变量")
}
```

优先级: **显式参数 > 环境变量 > 默认值** (`https://api.gitcode.com/api/v5`)。

### OAuth 2.0

适用于需要用户授权而非固定 Token 的场景:

```go
oauth := gitcode.NewOAuthClient("client-id", "client-secret", "https://app.example.com/callback")

// 1. 引导用户跳转授权
url := oauth.AuthorizeURL("user_info projects", "random-state")
http.Redirect(w, r, url, http.StatusFound)

// 2. 用回调拿到的 code 换 token
token, err := oauth.ExchangeToken(ctx, code)

// 3. 用 token 创建 API 客户端
client := oauth.NewClient(token)  // 推荐：保留 OAuth 客户端的 base URL

// 4. 后续可用 refresh token 续期
token, err = oauth.RefreshToken(ctx, token.RefreshToken)
```

#### OAuth 私有部署

```go
// 指定私有实例地址
oauth := gitcode.NewOAuthClientWithBaseURL(
    "client-id", "client-secret",
    "https://app.example.com/callback",
    "https://gitcode.example.com",  // 私有实例地址
)

// 后续流程与上面相同，oauth.NewClient(token) 会自动使用私有实例的 API 地址
client := oauth.NewClient(token)
```

## API 示例

> 下文示例中 `ctx := context.Background()`,`owner`/`repo` 为仓库所有者和路径名。

### 仓库与文件

```go
// 创建仓库(私有)
private := false
repo, _ := client.CreateRepository(ctx, gitcode.CreateRepositoryOptions{
    Name:        "my-repo",
    Description: "描述",
    Private:     &private,
    AutoInit:    &private, // true,初始化 README
})

// 更新 / 删除 / Fork
client.UpdateRepository(ctx, "owner", "my-repo", gitcode.UpdateRepositoryOptions{Description: "新描述"})
client.DeleteRepository(ctx, "owner", "my-repo")
client.ForkRepository(ctx, "owner", "repo", nil)

// 组织下创建仓库
client.CreateOrgRepository(ctx, "my-org", gitcode.CreateOrgRepoOptions{Name: "team-repo"})

// 文件 CRUD(content 需 base64 编码)
content, _ := client.GetRepositoryContent(ctx, "owner", "repo", "README.md", "main")
res, _ := client.CreateFile(ctx, "owner", "repo", "a.txt", gitcode.CreateFileOptions{
    Message: "add a",
    Content: base64.StdEncoding.EncodeToString([]byte("hi")),
    Branch:  "main",
})
client.UpdateFile(ctx, "owner", "repo", "a.txt", gitcode.UpdateFileOptions{
    Message: "upd", Content: "...", SHA: res.Content.SHA, Branch: "main",
})

// 原始内容 / Tree / Blob / 贡献者 / 语言
raw, _      := client.GetRawFile(ctx, "owner", "repo", "README.md", "main")
tree, _     := client.GetTree(ctx, "owner", "repo", "main", true)
blob, _     := client.GetBlob(ctx, "owner", "repo", sha)
contribs, _ := client.ListContributors(ctx, "owner", "repo")
langs, _    := client.GetLanguages(ctx, "owner", "repo")

// 标签 / Release
tags, _     := client.ListTags(ctx, "owner", "repo")
release, _  := client.CreateRelease(ctx, "owner", "repo", gitcode.CreateReleaseOptions{
    TagName: "v1.0.0", Title: "v1.0.0", Body: "release notes",
})
client.DeleteRelease(ctx, "owner", "repo", release.TagName)
```

### 仓库归档与统计

```go
// 下载仓库压缩包
tarGz, _ := client.GetRepositoryArchive(ctx, "owner", "repo", "main.tar.gz")
zip, _   := client.GetRepositoryArchive(ctx, "owner", "repo", "main.zip")

// 仓库统计
participation, _ := client.GetRepoParticipation(ctx, "owner", "repo")
codeFreq, _      := client.GetRepoCodeFrequency(ctx, "owner", "repo")
commitActivity, _ := client.GetRepoCommitActivity(ctx, "owner", "repo")
punchCard, _      := client.GetRepoPunchCard(ctx, "owner", "repo")
```

### 仓库主题 (Topics)

```go
// 列出主题
topics, _ := client.ListRepositoryTopics(ctx, "owner", "repo")

// 更新主题
client.UpdateRepositoryTopics(ctx, "owner", "repo", []string{"golang", "api", "sdk"})

// 添加/删除单个主题
client.AddRepositoryTopic(ctx, "owner", "repo", "new-topic")
client.DeleteRepositoryTopic(ctx, "owner", "repo", "old-topic")
```

### Fork 同步 / 远程镜像 / 许可协议 / CLA

```go
// Fork 同步
syncStatus, _ := client.GetForkSyncStatus(ctx, "owner", "fork-repo")
fmt.Printf("behind=%d ahead=%d\n", syncStatus.BehindBy, syncStatus.AheadBy)
client.SyncForkRepository(ctx, "owner", "fork-repo") // 同步源仓库

// 远程镜像
mirror, _ := client.GetRepoRemoteMirror(ctx, "owner", "repo")
mirrors, _ := client.ListPushRemoteMirrors(ctx, "owner", "repo", gitcode.ListOptions{})

// 许可协议
license, _ := client.GetRepoLicense(ctx, "owner", "repo")

// CLA
clas, _ := client.ListRepoCLAs(ctx, "owner", "repo")
client.ConfigureRepoCLA(ctx, "owner", "repo", &gitcode.RepoCLA{Name: "CLA", Content: "...", Enabled: true})
```

### 讨论 (Discussions)

```go
// 仓库讨论
discussions, _ := client.ListDiscussions(ctx, "owner", "repo", gitcode.ListOptions{})
discussion, _  := client.GetDiscussion(ctx, "owner", "repo", 1)
comments, _    := client.ListDiscussionComments(ctx, "owner", "repo", 1, gitcode.ListOptions{})
replies, _     := client.ListDiscussionCommentReplies(ctx, "owner", "repo", 1, commentID, gitcode.ListOptions{})

// 组织讨论
orgDiscussions, _ := client.ListOrgDiscussions(ctx, "my-org", gitcode.ListOptions{})
orgDiscussion, _  := client.GetOrgDiscussion(ctx, "my-org", 1)
```

### 协作者

```go
// 列出协作者
collabs, _ := client.ListCollaborators(ctx, "owner", "repo", gitcode.ListOptions{})

// 添加协作者
client.AddCollaborator(ctx, "owner", "repo", "username", &gitcode.AddCollaboratorOptions{
    Permission: "push", // pull, push, admin
})

// 检查是否为协作者
isCollab, _ := client.IsCollaborator(ctx, "owner", "repo", "username")

// 获取协作者权限
perm, _ := client.GetCollaboratorPermission(ctx, "owner", "repo", "username")

// 移除协作者
client.RemoveCollaborator(ctx, "owner", "repo", "username")
```

### Deploy Keys

```go
// 列出部署密钥
keys, _ := client.ListDeployKeys(ctx, "owner", "repo", gitcode.ListOptions{})

// 创建部署密钥
readOnly := true
key, _ := client.CreateDeployKey(ctx, "owner", "repo", gitcode.CreateDeployKeyOptions{
    Title:    "deploy",
    Key:      "ssh-ed25519 AAAA...",
    ReadOnly: &readOnly,
})

// 获取/删除
client.GetDeployKey(ctx, "owner", "repo", key.ID)
client.DeleteDeployKey(ctx, "owner", "repo", key.ID)
```

### Issue / 标签 / 里程碑

```go
// Issue CRUD
issue, _ := client.CreateIssue(ctx, "owner", "repo", gitcode.CreateIssueOptions{
    Title:  "Bug 报告",
    Body:   "问题描述",
    Labels: []string{"bug"},
})
client.UpdateIssue(ctx, "owner", "repo", int(issue.Number), gitcode.UpdateIssueOptions{State: gitcode.IssueStateClosed})
client.ReopenIssue(ctx, "owner", "repo", int(issue.Number))

// 评论
client.CreateIssueComment(ctx, "owner", "repo", int(issue.Number), "+1")
client.ListIssueComments(ctx, "owner", "repo", int(issue.Number))

// 仓库标签
client.CreateIssueLabel(ctx, "owner", "repo", "bug", "#ee0701")
client.UpdateIssueLabel(ctx, "owner", "repo", "bug", gitcode.UpdateLabelOptions{Color: "#d73a4a"})
client.AddIssueLabels(ctx, "owner", "repo", int(issue.Number), []string{"bug"})
client.ReplaceIssueLabels(ctx, "owner", "repo", int(issue.Number), []string{"bug", "p0"})

// 里程碑
ms, _ := client.CreateMilestone(ctx, "owner", "repo", "v2.0", "下个版本")
client.ListMilestones(ctx, "owner", "repo")
client.UpdateMilestone(ctx, "owner", "repo", int(ms.ID), gitcode.UpdateMilestoneOptions{State: "closed"})

// 操作日志 / 关联 PR
logs, _ := client.GetIssueOperateLogs(ctx, "owner", "repo", int(issue.Number))
prs, _  := client.GetIssueLinkedPRs(ctx, "owner", "repo", int(issue.Number))

// 时间线事件
timeline, _ := client.ListIssueTimelineEvents(ctx, "owner", "repo", int(issue.Number), gitcode.ListOptions{})

// 订阅者
subscribers, _ := client.ListIssueSubscribers(ctx, "owner", "repo", int(issue.Number), gitcode.ListOptions{})
client.SubscribeToIssue(ctx, "owner", "repo", int(issue.Number), "username")
client.UnsubscribeFromIssue(ctx, "owner", "repo", int(issue.Number), "username")

// Issue 依赖关系
deps, _ := client.ListIssueDependencies(ctx, "owner", "repo", int(issue.Number), gitcode.ListOptions{})
client.CreateIssueDependency(ctx, "owner", "repo", int(issue.Number), 42) // 依赖 issue #42
client.DeleteIssueDependency(ctx, "owner", "repo", int(issue.Number), 42)

// 指派人
assignees, _ := client.ListRepoAssignees(ctx, "owner", "repo", gitcode.ListOptions{})
client.AddIssueAssignees(ctx, "owner", "repo", int(issue.Number), []string{"user1", "user2"})
client.RemoveIssueAssignees(ctx, "owner", "repo", int(issue.Number), []string{"user1"})

// 关联分支
branches, _ := client.ListIssueRelatedBranches(ctx, "owner", "repo", int(issue.Number))
client.SetIssueRelatedBranches(ctx, "owner", "repo", int(issue.Number), []string{"feature-branch"})

// 看板字段
client.UpdateIssueKanbanValues(ctx, "owner", "repo", int(issue.Number), []gitcode.KanbanValue{
    {FieldID: 1, FieldName: "priority", ValueID: 2, ValueName: "high"},
})

// 修改历史
history, _ := client.ListIssueModifyHistory(ctx, "owner", "repo", int(issue.Number), gitcode.ListOptions{})
commentHistory, _ := client.ListIssueCommentModifyHistory(ctx, "owner", "repo", commentID, gitcode.ListOptions{})

// 企业 Issue 状态
statuses, _ := client.ListEnterpriseIssueStatuses(ctx, "my-enterprise")
```

### Issue 表情回应 (Reactions)

```go
// 添加表情
client.CreateIssueReaction(ctx, "owner", "repo", int(issue.Number), gitcode.ReactionHeart)
client.CreateIssueReaction(ctx, "owner", "repo", int(issue.Number), gitcode.ReactionPlusOne)

// 列出表情
reactions, _ := client.ListIssueReactions(ctx, "owner", "repo", int(issue.Number), gitcode.ListOptions{})

// 删除表情
client.DeleteIssueReaction(ctx, "owner", "repo", int(issue.Number), reactions[0].ID)

// 评论表情
client.CreateIssueCommentReaction(ctx, "owner", "repo", commentID, gitcode.ReactionRocket)
client.ListIssueCommentReactions(ctx, "owner", "repo", commentID, gitcode.ListOptions{})
```

### Pull Request

```go
pr, _ := client.CreatePullRequest(ctx, "owner", "repo", gitcode.CreatePullRequestOptions{
    Title: "feat: xxx", Head: "feature", Base: "main",
})

// 合并 (Squash)
client.MergePullRequest(ctx, "owner", "repo", pr.Number, &gitcode.MergePullRequestOptions{
    CommitMessage: "merge feat",
    Squash:        true,
})

// 文件 / 提交 / 评论 / 审查
client.ListPullRequestFiles(ctx, "owner", "repo", pr.Number)
client.ListPullRequestCommits(ctx, "owner", "repo", pr.Number)
client.CreatePullRequestReview(ctx, "owner", "repo", pr.Number, "LGTM", "APPROVE")

// 标签 / 审查人 / 测试人
client.AddPullRequestLabels(ctx, "owner", "repo", pr.Number, []string{"review"})
client.AssignPullRequestReviewers(ctx, "owner", "repo", pr.Number, "user1,user2")
client.AssignPullRequestTesters(ctx, "owner", "repo", pr.Number, "qa1")
client.HandlePullRequestReview(ctx, "owner", "repo", pr.Number, false) // force=false

// 关联 Issue / 操作日志
issues, _ := client.GetPullRequestLinkedIssues(ctx, "owner", "repo", pr.Number, gitcode.ListOptions{})
logs, _   := client.GetPullRequestOperateLogs(ctx, "owner", "repo", pr.Number)

// Diff / Patch
diff, _  := client.GetPullRequestDiff(ctx, "owner", "repo", pr.Number)
patch, _ := client.GetPullRequestPatch(ctx, "owner", "repo", pr.Number)

// 审查人管理
reviewers, _ := client.ListPullRequestReviewers(ctx, "owner", "repo", pr.Number)
client.RequestPullRequestReviewers(ctx, "owner", "repo", pr.Number, gitcode.PullRequestReviewRequest{
    Reviewers: []string{"reviewer1", "reviewer2"},
})

// 审查操作
review, _ := client.GetPullRequestReview(ctx, "owner", "repo", pr.Number, reviewID)
client.SubmitPullRequestReview(ctx, "owner", "repo", pr.Number, reviewID, "LGTM", "APPROVE")
client.DismissPullRequestReview(ctx, "owner", "repo", pr.Number, reviewID, "Dismissed")

// PR 表情回应
client.CreatePullRequestCommentReaction(ctx, "owner", "repo", commentID, gitcode.ReactionHooray)

// PR 关联/取消关联 Issue
client.LinkPullRequestIssue(ctx, "owner", "repo", pr.Number, 42)
client.UnlinkPullRequestIssue(ctx, "owner", "repo", pr.Number, 42)

// 取消测试人/审查人
client.UnassignPullRequestTesters(ctx, "owner", "repo", pr.Number, "qa1")
client.UnassignPullRequestReviewers(ctx, "owner", "repo", pr.Number, "user1")

// 可选测试人/审查人列表
availableTesters, _ := client.ListPullRequestAvailableTesters(ctx, "owner", "repo", pr.Number, gitcode.ListOptions{})
availableReviewers, _ := client.ListPullRequestAvailableReviewers(ctx, "owner", "repo", pr.Number, gitcode.ListOptions{})

// 评审人 (approval-reviewers)
client.AssignPullRequestApprovalReviewers(ctx, "owner", "repo", pr.Number, "reviewer1,reviewer2")
client.UnassignPullRequestApprovalReviewers(ctx, "owner", "repo", pr.Number, "reviewer1")

// 讨论回复 / 检视意见
client.ReplyPullRequestComment(ctx, "owner", "repo", pr.Number, "discussion-id", "Reply body")
client.ResolvePullRequestDiscussion(ctx, "owner", "repo", pr.Number, "discussion-id", true)

// 修改历史
prHistory, _ := client.ListPullRequestModifyHistory(ctx, "owner", "repo", pr.Number, gitcode.ListOptions{})
prCommentHistory, _ := client.ListPullRequestCommentModifyHistory(ctx, "owner", "repo", commentID, gitcode.ListOptions{})

// 刷新评论位置
client.RefreshPullRequestCommentPosition(ctx, "owner", "repo", pr.Number)

// 文件变更 JSON
fileChanges, _ := client.ListPullRequestFilesJSON(ctx, "owner", "repo", pr.Number, gitcode.ListOptions{})
```

### Commit Status (CI/CD 集成)

```go
// 创建提交状态
client.CreateCommitStatus(ctx, "owner", "repo", sha, gitcode.CreateCommitStatusOptions{
    State:       "success", // pending, success, error, failure
    TargetURL:   "https://ci.example.com/build/123",
    Description: "Build passed",
    Context:     "ci/build",
})

// 列出提交状态
statuses, _ := client.ListCommitStatuses(ctx, "owner", "repo", sha, gitcode.ListOptions{})

// 获取合并状态(CI 整体状态)
combined, _ := client.GetCombinedStatus(ctx, "owner", "repo", sha)
fmt.Printf("总状态: %d 个检查\n", combined.TotalCount)
```

### Commit 评论

```go
// 创建提交评论
comment, _ := client.CreateCommitComment(ctx, "owner", "repo", sha, gitcode.CreateCommitCommentOptions{
    Body:     "Nice work!",
    Path:     "main.go",
    Position: 10,
})

// 列出/获取/更新/删除
client.ListCommitComments(ctx, "owner", "repo", sha, gitcode.ListOptions{})
client.GetCommitComment(ctx, "owner", "repo", comment.ID)
client.UpdateCommitComment(ctx, "owner", "repo", comment.ID, gitcode.UpdateCommitCommentOptions{Body: "Updated"})
client.DeleteCommitComment(ctx, "owner", "repo", comment.ID)

// 列出仓库所有提交评论
client.ListRepoCommitComments(ctx, "owner", "repo", gitcode.ListOptions{})
```

### 分支与提交

```go
client.CreateBranch(ctx, "owner", "repo", gitcode.CreateBranchOptions{
    BranchName: "dev", Refs: "main",
})
client.DeleteBranch(ctx, "owner", "repo", "dev")

// 保护分支规则
client.CreateBranchProtection(ctx, "owner", "repo", gitcode.CreateBranchProtectionOptions{
    Name:                     "main",
    RequiredApprovingReviews: 2,
    AllowForcePushes:         false,
})
client.ListBranchProtections(ctx, "owner", "repo")
client.DeleteBranchProtection(ctx, "owner", "repo", "main")

// 提交
commits, _ := client.ListCommits(ctx, "owner", "repo", gitcode.ListCommitsOptions{Branch: "main"})
commit, _  := client.GetCommit(ctx, "owner", "repo", commits[0].SHA)
cmp, _     := client.CompareCommits(ctx, "owner", "repo", "main", "dev")
fmt.Printf("ahead=%d behind=%d\n", cmp.AheadBy, cmp.BehindBy)
```

### Git References

```go
// 列出引用
refs, _ := client.ListGitReferences(ctx, "owner", "repo", gitcode.ListOptions{})

// 列出分支引用
heads, _ := client.ListGitRefSubPaths(ctx, "owner", "repo", "heads/")

// 获取引用
ref, _ := client.GetGitReference(ctx, "owner", "repo", "heads/main")

// 创建引用
client.CreateGitReference(ctx, "owner", "repo", gitcode.CreateReferenceOptions{
    Ref: "refs/heads/new-branch",
    SHA: "abc123...",
})

// 更新引用
client.UpdateGitReference(ctx, "owner", "repo", "heads/main", gitcode.UpdateReferenceOptions{
    SHA:   "def456...",
    Force: false,
})

// 删除引用
client.DeleteGitReference(ctx, "owner", "repo", "heads/old-branch")
```

### Git Tags (注释标签)

```go
// 创建注释标签
annotatedTag, _ := client.CreateAnnotatedTag(ctx, "owner", "repo", gitcode.CreateAnnotatedTagOptions{
    Tag:     "v2.0.0",
    Message: "Release v2.0.0",
    Object:  "abc123...",
    Type:    "commit",
})

// 获取注释标签
client.GetAnnotatedTag(ctx, "owner", "repo", annotatedTag.SHA)

// Release 资产
assets, _ := client.ListReleaseAssets(ctx, "owner", "repo", releaseID, gitcode.ListOptions{})
client.GetReleaseAsset(ctx, "owner", "repo", assetID)
client.DeleteReleaseAsset(ctx, "owner", "repo", assetID)

// 通过标签获取 Release
release, _ := client.GetReleaseByTag(ctx, "owner", "repo", "v1.0.0")

// 更新 Release
client.UpdateRelease(ctx, "owner", "repo", releaseID, gitcode.UpdateReleaseOptions{
    Body: "Updated release notes",
})
```

### Webhook

```go
active := true
hook, _ := client.CreateWebhook(ctx, "owner", "repo", gitcode.CreateWebhookOptions{
    URL:                 "https://example.com/hook",
    Password:            "s3cret", // wire 字段名为 password(非 GitHub 的 secret)
    PushEvents:          gitcode.Bool(true),
    MergeRequestsEvents: gitcode.Bool(true), // pull_request 事件族
    Active:              &active,
})
client.TestWebhook(ctx, "owner", "repo", hook.ID)
client.DeleteWebhook(ctx, "owner", "repo", hook.ID)
```

GitCode 的 webhook 事件订阅是逐族布尔字段(`push_events` /
`tag_push_events` / `issues_events` / `note_events` /
`merge_requests_events`),没有 `events` 数组;更新时 `nil` = 不变,
`gitcode.Bool(false)` = 关闭该族。中性事件名可用
`gitcode.EventFlags` 映射。

接收并解析事件(以 Gin 为例):

```go
payload, _ := io.ReadAll(r.Body)
switch r.Header.Get("X-Gitcode-Event") {
case "push":
    e, _ := client.ParsePushEvent(payload)
    log.Printf("push %s -> %s", e.Before[:8], e.After[:8])
case "pull_request":
    e, _ := client.ParsePullRequestEvent(payload)
    log.Printf("PR #%d %s", e.Number, e.Action)
case "issues":
    e, _ := client.ParseIssueEvent(payload)
    log.Printf("issue #%d %s", int(e.Issue.Number), e.Action)
case "note":
    e, _ := client.ParseNoteEvent(payload)
    log.Printf("note on %s", e.NoteType)
case "tag_push":
    e, _ := client.ParseTagPushEvent(payload)
    log.Printf("tag %s", e.Ref)
}
```

> 还提供了 `PushEvent`/`PullRequestWebhookEvent`/`IssueWebhookEvent`/`NoteWebhookEvent`/`TagPushEvent` 五种事件类型,见 `webhooks.go`。

### 用户 / SSH 公钥 / 邮箱

```go
me, _     := client.GetCurrentUser(ctx)
user, _   := client.GetUser(ctx, "somebody")
emails, _ := client.ListEmails(ctx)

key, _ := client.CreateSSHKey(ctx, gitcode.CreateSSHKeyOptions{
    Title: "mbp", Key: "ssh-ed25519 AAAA...",
})
client.ListSSHKeys(ctx, gitcode.ListOptions{PerPage: 50})
client.DeleteSSHKey(ctx, key.ID)

events, _ := client.GetUserEvents(ctx, me.Login, "2025", "") // 年度动态
starred, _ := client.ListStarredRepositories(ctx, gitcode.ListStarredReposOptions{})
ns, _      := client.GetNamespace(ctx, "somepath")

// Watch 了的仓库
watched, _ := client.ListUserWatchedRepositories(ctx, "username", gitcode.ListOptions{})
myWatched, _ := client.ListCurrentUserWatchedRepositories(ctx, gitcode.ListOptions{})

// 更新个人资料
client.UpdateCurrentUser(ctx, gitcode.UpdateCurrentUserOptions{
    Name: "New Name", Bio: "Go developer",
})

// 用户 PR 列表
myPRs, _ := client.ListUserPullRequests(ctx, gitcode.ListPullRequestsOptions{
    State: gitcode.PullRequestStateOpen,
})
```

### 用户关注

```go
// 列出关注者/被关注者
followers, _ := client.ListUserFollowers(ctx, "username", gitcode.ListOptions{})
following, _ := client.ListUserFollowing(ctx, "username", gitcode.ListOptions{})

// 当前用户的关注关系
myFollowers, _ := client.ListCurrentUserFollowers(ctx, gitcode.ListOptions{})
myFollowing, _ := client.ListCurrentUserFollowing(ctx, gitcode.ListOptions{})

// 关注/取关
client.FollowUser(ctx, "target-user")
client.UnfollowUser(ctx, "target-user")

// 检查是否关注
isFollowing, _ := client.IsFollowing(ctx, "target-user")
```

### 组织 / 企业

```go
// 组织
client.ListUserOrganizations(ctx, "username", gitcode.ListOptions{})
org, _      := client.GetOrgInfo(ctx, "my-org")
client.UpdateOrganization(ctx, "my-org", gitcode.UpdateOrgOptions{Description: "..."})
client.InviteOrgMember(ctx, "my-org", "newbie", gitcode.InviteMemberOptions{Permission: "write"})
client.RemoveOrgMember(ctx, "my-org", "newbie")
client.ListOrgMembers(ctx, "my-org", "admin", gitcode.ListOptions{})

// 企业
client.ListEnterpriseMembers(ctx, "my-ent", "", gitcode.ListOptions{})
client.UpdateEnterpriseMember(ctx, "my-ent", "user", gitcode.UpdateEnterpriseMemberOptions{Role: "admin"})

// 企业 Issue / PR
client.ListEnterpriseIssues(ctx, "my-ent", gitcode.ListUserIssuesOptions{State: "open"})
client.ListEnterprisePullRequests(ctx, "my-ent", gitcode.ListEnterprisePRsOptions{State: "open"})
client.ListEnterpriseLabels(ctx, "my-ent")
client.GetOrgIssueExtendSettings(ctx, "my-org") // 自定义状态扩展
```

### 组织公开成员 / 屏蔽

```go
// 公开成员
publicMembers, _ := client.ListOrgPublicMembers(ctx, "my-org", gitcode.ListOptions{})
client.PublicizeOrgMembership(ctx, "my-org", "username")
client.ConcealOrgMembership(ctx, "my-org", "username")
isPublic, _ := client.IsOrgPublicMember(ctx, "my-org", "username")

// 屏蔽用户
blocked, _ := client.ListOrgBlockedUsers(ctx, "my-org", gitcode.ListOptions{})
client.BlockOrgUser(ctx, "my-org", "spammer")
client.UnblockOrgUser(ctx, "my-org", "spammer")
isBlocked, _ := client.IsOrgBlockedUser(ctx, "my-org", "spammer")

// 组织自定义角色
roles, _ := client.ListOrgCustomizedRoles(ctx, "my-org")
```

### 组织团队

```go
// 创建团队
team, _ := client.CreateTeam(ctx, "my-org", gitcode.CreateTeamOptions{
    Name:       "backend",
    Permission: "write",
    Privacy:    "closed",
})

// 列出/获取/更新/删除团队
teams, _ := client.ListOrgTeams(ctx, "my-org", gitcode.ListOptions{})
client.GetTeam(ctx, team.ID)
client.UpdateTeam(ctx, team.ID, gitcode.UpdateTeamOptions{Name: "backend-team"})
client.DeleteTeam(ctx, team.ID)

// 团队成员管理
client.ListTeamMembers(ctx, team.ID, gitcode.ListOptions{})
client.AddTeamMember(ctx, team.ID, "new-member")
client.RemoveTeamMember(ctx, team.ID, "old-member")

// 团队仓库管理
client.ListTeamRepositories(ctx, team.ID, gitcode.ListOptions{})
client.AddTeamRepository(ctx, team.ID, "my-org", "my-repo")
client.RemoveTeamRepository(ctx, team.ID, "my-org", "my-repo")
```

### 组织 Webhook

```go
// 创建组织 Webhook
active := true
hook, _ := client.CreateOrgWebhook(ctx, "my-org", gitcode.CreateOrgWebhookOptions{
    URL:    "https://example.com/org-hook",
    Events: []string{"push", "repository"},
    Active: &active,
})

// 列出/获取/更新/删除
hooks, _ := client.ListOrgWebhooks(ctx, "my-org", gitcode.ListOptions{})
client.GetOrgWebhook(ctx, "my-org", hook.ID)
client.UpdateOrgWebhook(ctx, "my-org", hook.ID, gitcode.UpdateOrgWebhookOptions{URL: "https://new-url.com"})
client.DeleteOrgWebhook(ctx, "my-org", hook.ID)
```

### 组织标签

```go
// 创建组织标签
label, _ := client.CreateOrgLabel(ctx, "my-org", gitcode.CreateOrgLabelOptions{
    Name:  "priority:high",
    Color: "#ff0000",
})

// 列出/获取/更新/删除
labels, _ := client.ListOrgLabels(ctx, "my-org", gitcode.ListOptions{})
client.GetOrgLabel(ctx, "my-org", label.ID)
client.UpdateOrgLabel(ctx, "my-org", label.ID, gitcode.UpdateOrgLabelOptions{Color: "#cc0000"})
client.DeleteOrgLabel(ctx, "my-org", label.ID)
```

### 搜索

```go
repos, _ := client.SearchRepositories(ctx, gitcode.SearchRepositoriesOptions{
    Query: "gin web", Sort: "stars_count",
})
issues, _ := client.SearchIssues(ctx, gitcode.SearchIssuesOptions{
    Query: "memory leak", Repo: "owner/repo", State: "open",
})
users, _ := client.SearchUsers(ctx, gitcode.SearchUsersOptions{Query: "octocat"})
```

### Star / Watch / 通知

```go
client.StarRepository(ctx, "owner", "repo")
starred, _ := client.IsRepositoryStarred(ctx, "owner", "repo") // bool
client.UnstarRepository(ctx, "owner", "repo")

stargazers, _ := client.ListStargazers(ctx, "owner", "repo", gitcode.ListOptions{})
watchers, _   := client.ListWatchers(ctx, "owner", "repo", gitcode.ListOptions{})
```

### 增强通知

```go
// 列出通知(带筛选)
notifs, _ := client.ListNotificationsWithOptions(ctx, gitcode.ListNotificationsOptions{
    ListOptions: gitcode.ListOptions{PerPage: 50},
    Status:      "unread",
})

// 获取通知线程详情
thread, _ := client.GetNotificationThread(ctx, threadID)

// 标记单个线程已读
client.MarkNotificationThreadAsRead(ctx, threadID)

// 标记所有通知已读
client.MarkNotificationsAsRead(ctx, gitcode.MarkNotificationsOptions{All: true})

// 仓库通知
repoNotifs, _ := client.ListRepoNotifications(ctx, "owner", "repo", gitcode.ListNotificationsOptions{})
client.MarkRepoNotificationsAsRead(ctx, "owner", "repo", gitcode.MarkNotificationsOptions{All: true})
```

### Wiki

```go
// 列出 Wiki 页面
pages, _ := client.ListWikiPages(ctx, "owner", "repo", gitcode.ListOptions{})

// 获取单个页面
page, _ := client.GetWikiPage(ctx, "owner", "repo", "Home")

// 创建页面
content := base64.StdEncoding.EncodeToString([]byte("# Hello Wiki"))
client.CreateWikiPage(ctx, "owner", "repo", gitcode.CreateWikiPageOptions{
    Title:         "New Page",
    ContentBase64: content,
    Message:       "Create new page",
})

// 更新/删除
client.UpdateWikiPage(ctx, "owner", "repo", "New-Page", gitcode.UpdateWikiPageOptions{
    ContentBase64: content,
    Message:       "Update page",
})
client.DeleteWikiPage(ctx, "owner", "repo", "Old-Page")
```

### 仓库邀请

```go
// 列出待处理邀请
invitations, _ := client.ListPendingRepoInvitations(ctx, gitcode.ListOptions{})

// 接受/拒绝邀请
client.AcceptRepoInvitation(ctx, invitationID)
client.DeclineRepoInvitation(ctx, invitationID)
```

### 模板

```go
// Issue 模板
issueTemplates, _ := client.ListIssueTemplates(ctx, "owner", "repo")
client.GetIssueTemplate(ctx, "owner", "repo", "bug_report")

// PR 合并模板
mergeTemplates, _ := client.ListPullRequestMergeTemplates(ctx, "owner", "repo")

// Gitignore 模板
gitignoreTemplates, _ := client.ListGitignoreTemplates(ctx)
client.GetGitignoreTemplate(ctx, "Go")

// License 模板
licenses, _ := client.ListLicenseTemplates(ctx)
client.GetLicenseTemplate(ctx, "mit")

// Label 模板
labelTemplates, _ := client.ListLabelTemplates(ctx)
client.GetLabelTemplate(ctx, "default")
```

### Markdown 渲染

```go
html, _ := client.RenderMarkdown(ctx, "**bold** and *italic*", "gfm", "owner/repo")
html, _ = client.RenderMarkdownRaw(ctx, "# Raw Markdown")
```

### 仓库设置类

```go
client.UpdateRepoSettings(ctx, "owner", "repo", &gitcode.RepoSettings{HasIssues: true})
client.UpdatePushConfig(ctx, "owner", "repo", &gitcode.PushConfig{MaxFileSize: 104857600})
client.UpdatePRSettings(ctx, "owner", "repo", &gitcode.PRSettings{DefaultMergeMethod: "merge"})
client.SetModuleSetting(ctx, "owner", "repo", gitcode.ModuleSetting{Wiki: false})
client.UpdateReviewerConfig(ctx, "owner", "repo", gitcode.ReviewerConfig{MinApprovingReviews: 1})

// 归档 / 转让
client.ArchiveRepository(ctx, "owner", "repo")
client.TransferRepository(ctx, "owner", "repo", gitcode.TransferRepoOptions{NewOwner: "new-owner"})

// 限流
rl, _ := client.GetRateLimit(ctx)
```

## 高级功能

### 结构化错误处理

所有 API 错误都返回结构化类型，可通过类型断言区分：

```go
_, err := client.GetRepository(ctx, "owner", "missing")
if err != nil {
    switch {
    case gitcode.IsNotFound(err):
        fmt.Println("仓库不存在")
    case gitcode.IsUnauthorized(err):
        fmt.Println("认证失败")
    case gitcode.IsRateLimit(err):
        fmt.Println("触发限流")
    case gitcode.IsForbidden(err):
        fmt.Println("权限不足")
    case gitcode.IsConflict(err):
        fmt.Println("资源已存在")
    default:
        fmt.Printf("其他错误: %v\n", err)
    }
}
```

### Rate Limit 自动重试

配置自动重试策略，遇到 429/5xx 自动等待重试：

```go
client.SetRetryPolicy(gitcode.RetryPolicy{
    MaxRetries:           3,
    InitialBackoff:       1 * time.Second,
    MaxBackoff:           30 * time.Second,
    Multiplier:           2.0,
    RetryableStatusCodes: []int{429, 500, 502, 503, 504},
})
```

### 自动分页

自动遍历所有分页，无需手动处理 `Page` 参数：

```go
// 收集所有结果到一个 slice
allIssues, _ := gitcode.CollectAll(ctx, func(opts gitcode.ListOptions) ([]*gitcode.Issue, error) {
    return client.ListIssues(ctx, "owner", "repo", gitcode.ListIssuesOptions{
        ListOptions: opts,
        State:       gitcode.IssueStateOpen,
    })
})

// 使用迭代器逐页处理
for items, err := range gitcode.Paginate(ctx, func(opts gitcode.ListOptions) ([]*gitcode.Repository, error) {
    return client.ListRepositories(ctx, gitcode.ListRepositoriesOptions{ListOptions: opts})
}) {
    if err != nil { log.Fatal(err) }
    for _, repo := range items {
        fmt.Println(repo.FullName)
    }
}
```

### Webhook 签名验证

验证 Webhook 请求的 HMAC-SHA256 签名：

```go
func handleWebhook(w http.ResponseWriter, r *http.Request) {
    payload, _ := io.ReadAll(r.Body)
    signature := r.Header.Get("X-Gitcode-Signature")

    if !gitcode.VerifyWebhookSignature(payload, "your-webhook-secret", signature) {
        http.Error(w, "Invalid signature", http.StatusUnauthorized)
        return
    }

    // 签名有效，处理事件
    event, _ := client.ParsePushEvent(payload)
    fmt.Printf("Push to %s\n", event.Ref)
}

// 计算签名 (用于测试)
sig := gitcode.ComputeWebhookSignature(payload, secret)
```

### 请求/响应中间件 (Hooks)

添加请求前后的拦截器，用于日志、监控、调试：

```go
// 请求前 Hook
client.AddRequestHook(func(req *http.Request) error {
    log.Printf("[REQ] %s %s", req.Method, req.URL.Path)
    req.Header.Set("X-Request-ID", uuid.New().String())
    return nil
})

// 响应后 Hook
client.AddResponseHook(func(resp *http.Response) error {
    log.Printf("[RESP] %d %s", resp.StatusCode, resp.Request.URL.Path)
    return nil
})

// 清除所有 Hook
client.ClearHooks()
```

### Multipart 文件上传

使用 `io.Reader` 上传文件内容：

```go
// 上传字节
result, _ := client.UploadFileBytes(ctx, "owner", "repo", "hello.txt", []byte("Hello"))

// 上传 io.Reader
file, _ := os.Open("local-file.go")
defer file.Close()
result, _ = client.UploadFileReader(ctx, "owner", "repo", "remote-name.go", file)

// 上传图片
img, _ := os.Open("screenshot.png")
defer img.Close()
result, _ = client.UploadImageReader(ctx, "owner", "repo", "screenshot.png", img)
```

### 输入校验

客户端自动校验必填字段，提前返回友好错误：

```go
// owner 或 repo 为空时返回 ValidationErrorField
_, err := client.GetRepository(ctx, "", "repo")
// err: "validation error: owner - owner is required"

if gitcode.IsValidationError(err) {
    fmt.Println(err)
}
```

### 并发安全

`SetAuthStyle`、`SetHTTPClient`、`AddRequestHook`、`AddResponseHook` 等方法均使用 `sync.RWMutex` 保护，可安全在多个 goroutine 中调用。

---

## 通用类型

`types.go` 中定义了若干容错类型,用于处理 GitCode API 返回值类型不稳定的情况:

| 类型 | 用途 |
|---|---|
| `FlexInt` | 同一字段可能是 `123` 或 `"123"` |
| `FlexString` | 同一字段可能是 `"abc"` 或 `123` |
| `NullableTime` | 时间字段可能为空字符串 |
| `Timestamp` | 字符串格式时间(`RFC3339`) |
| `Error` | API 错误响应结构(见下) |

## 错误处理

`Client` 在 HTTP 状态码 ≥ 400 时返回标准 `error`(`fmt.Errorf`),错误信息包含方法、路径、状态码和响应体:

```go
_, err := client.GetRepository(ctx, "owner", "missing")
if err != nil {
    log.Println(err)
    // 输出形如: GitCode API GET /repos/owner/missing returned 404: {"message":"404 Not Found"}
}
```

如需进一步解析错误体,可使用 `types.Error`:

```go
var apiErr gitcode.Error
if json.Unmarshal([]byte(extractBody(err)), &apiErr) == nil {
    log.Printf("message=%s", apiErr.Message)
}
```

## 项目结构

```
gitcode_api/
├── client.go                   # Client 构造、鉴权、HTTP 请求、User、ListOptions、RateLimit
├── oauth.go                    # OAuthClient 授权码流程
├── repos.go                    # 仓库 / 文件 / Tree / Blob / 标签 / Release / 设置 / 归档 / 转让
├── repos_collaborators.go      # 仓库协作者 CRUD、权限查询
├── repos_topics.go             # 仓库主题管理
├── repos_statuses.go           # 提交状态、合并状态 (CI/CD)
├── repos_deploy_keys.go        # 部署密钥 CRUD
├── repos_reactions.go          # Issue / 评论 / PR 表情回应
├── repos_wiki.go               # Wiki 页面 CRUD
├── repos_invitations.go        # 仓库邀请
├── repos_templates.go          # Issue / PR 模板
├── repos_commits_comments.go   # 提交评论
├── repos_archive.go            # 仓库归档下载、统计
├── repos_assignees.go          # 指派人管理
├── repos_reviewers.go          # PR 审查人、Diff、Patch
├── repos_issues_timeline.go    # Issue 时间线、订阅者、依赖关系
├── repos_discussions.go        # 仓库讨论、评论、回复 / Fork 同步 / 远程镜像 / 许可协议 / CLA
├── issues.go                   # Issue CRUD / 评论 / 标签 / 里程碑
├── issues_enhanced.go          # Issue 关联分支 / 看板字段 / 修改历史 / 企业状态 / 表态
├── enterprise_issues.go        # 用户 / 组织 / 企业级 Issue、操作日志、关联 PR
├── pulls.go                    # Pull Request 全套 + 审查 / 测试 / 标签 / 企业 PR
├── pulls_enhanced.go           # PR 关联 Issue / 测试人管理 / 评审人 / 讨论回复 / 修改历史 / 表态
├── branches.go                 # 分支 / 保护分支 / 提交 / 比较
├── webhooks.go                 # Webhook CRUD + 5 类事件解析
├── search.go                   # 搜索仓库 / Issue / 用户
├── users.go                    # SSH 公钥 / 邮箱 / 动态 / Star / Namespace
├── users_enhanced.go           # Watch 仓库 / 更新资料 / 用户 PR 列表
├── user_followers.go           # 用户关注 / 取关
├── orgs.go                     # 组织 / 企业成员管理 / Issue 扩展配置
├── orgs_enhanced.go            # 组织自定义角色 / 组织讨论
├── org_teams.go                # 组织团队 CRUD / 成员 / 仓库
├── org_hooks.go                # 组织 Webhook CRUD
├── org_labels.go               # 组织标签 CRUD
├── org_members_enhanced.go     # 公开成员 / 屏蔽用户 / 创建删除组织
├── milestones.go               # 里程碑(带选项版本)
├── labels.go                   # 仓库标签更新 / 替换 / 企业标签
├── git_refs.go                 # Git 引用 CRUD
├── git_tags.go                 # 注释标签 / Release 资产
├── notifications_enhanced.go   # 增强通知操作
├── misc.go                     # Gitignore/License/Label 模板 / Markdown / 用户仓库
├── types.go                    # FlexInt/FlexString/NullableTime/Error 等通用类型 + Star/通知 API
├── errors.go                   # 结构化错误类型: NotFound/Unauthorized/RateLimit/Conflict/Validation
├── retry.go                    # Rate Limit 自动重试 + 指数退避
├── pagination.go               # 自动分页迭代器 (泛型)
├── hooks.go                    # 请求/响应中间件
├── validate.go                 # 输入校验
├── webhook_verify.go           # Webhook HMAC-SHA256 签名验证
├── upload.go                   # Multipart 文件上传
├── gitcode_test.go             # 单元测试 + 真实 API 集成测试
├── client_enhanced_test.go     # 新功能单元测试
└── examples/
    └── main.go                 # 完整使用示例
```

## API 覆盖率

本库覆盖了 GitCode 官方文档 ([docs.gitcode.com/docs/apis](https://docs.gitcode.com/docs/apis/)) 中的绝大部分 API:

| 模块 | API 数量 | 状态 |
|---|---|---|
| 认证 & 用户 | 20+ | ✅ 完整 |
| 仓库 CRUD & 文件 | 40+ | ✅ 完整 |
| 讨论 (Discussions) | 10 | ✅ 完整 |
| Fork 同步 / 远程镜像 / CLA | 7 | ✅ 完整 |
| Issue / 标签 / 里程碑 | 30+ | ✅ 完整 |
| Issue 增强 (分支/看板/历史) | 8 | ✅ 完整 |
| Pull Request | 40+ | ✅ 完整 |
| PR 增强 (关联/评审/讨论/历史) | 15 | ✅ 完整 |
| 分支 & 提交 | 17 | ✅ 完整 |
| 保护标签 | 5 | ✅ 完整 |
| Release 增强 | 8 | ✅ 完整 |
| Webhook | 5 CRUD + 5 解析 | ✅ 完整 |
| 组织 / 企业 | 40+ | ✅ 完整 |
| 企业里程碑 / 标签 / 自定义字段 | 14 | ✅ 完整 |
| 看板 (Dashboard) | 7 | ✅ 完整 |
| Actions/CI (工作流/Runner/Artifact) | 23 | ✅ 完整 |
| 组织团队 | 12 | ✅ 完整 |
| 搜索 | 3 | ✅ 完整 |
| 协作者 | 5 | ✅ 完整 |
| Topics | 4 | ✅ 完整 |
| Commit Status | 3 | ✅ 完整 |
| Deploy Keys | 4 | ✅ 完整 |
| Git References | 6 | ✅ 完整 |
| Reactions | 12 | ✅ 完整 |
| Wiki | 5 | ✅ 完整 |
| 通知 | 6 | ✅ 完整 |
| 模板 | 7 | ✅ 完整 |
| **总计** | **428** | ✅ |

## 测试

测试分为两类:纯函数测试(不需要网络)和 **真实 API 集成测试**(会在你的账户下创建/删除临时仓库)。

```bash
# 默认运行(需在 gitcode_test.go 顶部填入有效 Token)
GITCODE_TOKEN="your-token" go test -v ./...

# 跳过需要联网的集成测试
go test -short ./...
```

> 集成测试会真实创建仓库(`test-api-<timestamp>`),测试结束后自动清理。请使用测试账号的 Token。

## 与 gitcode-cli 的关系

本项目参考了 [gitcode-cli](https://gitcode.com/gitcode-cli/cli) 的设计理念,提供 Go 语言的 API 客户端实现:

| gitcode-cli (`gc`) | `gitcode_api` |
|---|---|
| `gc auth login` | `NewClient(token)` / `OAuthClient` |
| `gc repo list` | `ListRepositories()` |
| `gc issue create` | `CreateIssue()` |
| `gc pr create` | `CreatePullRequest()` |

## 相关项目

- [gitcode-cli](https://gitcode.com/gitcode-cli/cli) — GitCode 官方 CLI 工具
- [git-platform-sdk](https://github.com/yi-nology/git-platform-sdk) — 多平台 Git SDK 统一接口(GitHub/GitLab/Gitea/Forgejo/GitCode)

## 许可证

MIT。

# GitCode wire-provenance ledger（SDK wire 形态来源分类）

> 生成于 2026-11-02 的全量审计：docs.gitcode.com sitemap 的 316 个 API 文档页
> 与 SDK 全部请求路径机械 diff 后的逐文件分类。三个 wire 方言在本 SDK 内并存，
> 分类决定"改/留/删"的口径。上游文档变更时重新跑 diff（脚本见本文件尾部）。

## 方言 A：v5 文档化（docs.gitcode.com 有 v5 页）—— 契约即文档

形状以 docs 页为准；拼写（连字符段、单复数）与 Gitee v5 的差异以文档为
准（v0.12.0 起 pull-requests 链接已从下划线改为文档拼写）。

| 面 | 文件 | 说明 |
|---|---|---|
| 仓库/分支/文件/commits/compare | repos.go, branches.go, files.go, commits.go, commits_enhanced.go, branches.go(compare) | 主干 CRUD；commits/{sha}/diff、/patch 为文档段形式 |
| Issues 全族 | issues.go, issues_enhanced.go, enterprise_issues.go(用户/组织部分) | 评论列表文档化 page/per_page；modify-history、user-reactions、related-branches、operate-logs、pull-requests 链接均为文档拼写 |
| Pulls 全族 | pulls.go, pulls_enhanced.go | approval-reviewers/testers、files-json、discussions 评论、modify-history、operate-logs |
| Labels/Milestones | labels.go, milestones.go | |
| Webhooks | webhooks.go | v0.8.0 起按文档：password + 五族布尔旗标 |
| Releases | releases.go, releases_enhanced.go, git_tags.go(releases 部分) | |
| Users/Orgs/Search | users.go, users_enhanced.go, orgs.go, search.go | /user/orgs 不被服务（测试 skip 实证）；org 归属走 /users/{username}/orgs 文档页 |
| Notifications（仓库域） | notifications_enhanced.go | 文档只有仓库域 GET/PUT /repos/{owner}/{repo}/notifications |
| Dashboard kanban | dashboard_kanban.go | v5 文档分类存在（get-api-v-5-org-owner-kanban-list 族） |
| Discussion | repos_discussions.go(discuss 部分) | v5 文档分类存在 |
| Enterprise（v8） | enterprise.go | v8-only 分类；数字企业 id 寻址；label 端点文档响应示例损坏，模型取最小集待 live 验证 |
| Enterprise（v5 遗存） | enterprise_v5.go | 仅 v5 文档存在的 4 端点：单 Issue 查、评论/标签/状态列表 |
| Actions（v8） | actions_v8.go | v8-only 分类；stages→jobs→steps 模型；shared/K8S runner 变体沿用 host-runners 信封（文档未单列 schema，已注释标注） |

## 方言 B：未文档化但 Gitea 血统可辩护 —— 保留 + 注释，live 验证后再定

GitCode 引擎为 Gitea 系（webhook 头 X-Gitea-Event 实证）。以下端点在
Gitea API 中存在、GitCode 文档未收录。**没有真 token 的 live 验证前不改
不删**——删除"可能真实"的能力与保留"可能虚构"的能力同样错误。

| 面 | 文件 | 备注 |
|---|---|---|
| Commit status | repos_statuses.go | 主仓消费（CreateCommitStatus/CombinedStatus）；文档无页 |
| Deploy keys | repos_deploy_keys.go | 主仓消费；文档只有 /user/keys 个人 SSH 键 |
| Reactions | repos_reactions.go | 主仓消费；文档只有 user-reactions 的 GET 列表页（issues_enhanced.go 已按文档形状另配）；create/delete 反应无文档页 |
| Issues timeline/dependencies/blocks/subscribers | repos_issues_timeline.go | Gitea 全套 |
| Wiki | repos_wiki.go | Gitea /wiki/page 形状 |
| Git refs/tags | git_refs.go, git_tags.go | Gitea /git/refs、/git/tags |
| Issue/label/merge 模板 | repos_templates.go | Gitea 形状 |
| Repo assignees | repos_assignees.go | Gitea 形状 |
| Reviewers（reviews/dismissals/requested_reviewers） | repos_reviewers.go, pulls.go(reviews 部分) | Gitea 形状；approval-reviewers 族另有 v5 文档页 |
| PR diff/patch（.diff/.patch 后缀） | repos_reviewers.go | Gitee v5 形状；文档无页；主仓消费 GetPullRequestDiff |
| Notifications（全局域） | notifications_enhanced.go | Gitea 有 /notifications 族；文档只有仓库域；主仓双用（全局+仓库域），仓库域已文档化 |
| /user/orgs | types.go | **例外：live 实证不被服务**（ListOrganizations 测试 skip 注明），保留方法但勿依赖 |
| Forks/Sync/Mirror/Transfer/Archive/统计/事件/设置族 | repos.go, repos_archive.go, repos_discussions.go(mirror/sync 部分) | Gitee v5 有多个同形（forks/sync-repo/push-config/transfer/events）；stats/* 与 download_statistics、file_list 为 GitCode 特有；archive 为 Gitea 形状 |

## 方言 C：确认虚构（已处理存档）

| 面 | 处置 | 版本 |
|---|---|---|
| /orgs/{org}/hooks 组织级 webhook | 整文件删除（v5/v8/AtomGit 均无文档） | v0.9.0 |
| Actions /api/v5 GitHub 形状 | 按 v8 文档整面重写 | v0.10.0 |
| Enterprise /api/v5 名字寻址 | 按 v8 文档整面重写；v5-only 四端点回迁 enterprise_v5.go | v0.11.0 / v0.12.0 |
| webhook secret/events 数组 | 按文档改 password + 五族布尔 | v0.8.0 |
| milestone due_date/id 寻址 | 按文档改 due_on/number+id 兜底 | v0.8.0 |
| Issue 链接 pull_requests→pull-requests | 按文档拼写改（repos/enterprise/org 三处 + Issue 链接） | v0.12.0 |

## 复跑审计

```bash
curl -s https://docs.gitcode.com/sitemap.xml -o /tmp/sitemap.xml
# 提取 /docs/apis/ 的 <loc>，按本文件开头的归一化规则（动词+版本剥除、
# 多词参数保护、参数词表）与 SDK 路径 diff；白名单：actions_v8.go、
# enterprise.go（v8 已核对）。
```

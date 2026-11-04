# GitCode wire-provenance ledger（SDK wire 形态来源分类）

> 生成于 2026-11-02 的全量审计（docs.gitcode.com sitemap 316 个 API 文档页
> vs SDK 全部请求路径机械 diff），**2026-11-03 以有效 token 对真实服务做了
> 只读 live 探针复核**——live 结果覆盖文档与推测，冲突处以 live 为准。
> 复跑方式见文末。

## live 探针结论（2026-11-03，token: 项目自带测试凭据）

方法：对真实仓库（含真实 commit SHA 与真实 issue #1）逐端点只读 GET，
文档化端点作对照（200 证实路由存在），未文档化端点 404 即路由不存在
（资源本身存在——issue/commit 均真实——故 404 可归因路由）。

| 结论 | 端点 |
|---|---|
| **活（未文档化但 200）** | `/repos/{o}/{r}/issues/{n}/pull_requests`（下划线；连字符 404）、`/issues/{n}/operate_logs`（下划线；连字符 404） |
| **死（404，live 实证）** | `/repos/{o}/{r}/statuses/{sha}`、`/repos/{o}/{r}/keys`（deploy keys）、`/issues/{n}/reactions`、`/issues/{n}/user-reactions`（文档形也死）、`/issues/comments/{id}/reactions`、`/pulls/comments/{id}/reactions`、`/issues/{n}/timeline`、`/issues/{n}/subscribers`、`/issues/{n}/dependencies`、`/repos/{o}/{r}/topics`、`/repos/{o}/{r}/git/refs`、`/repos/{o}/{r}/git/tags`、`/repos/{o}/{r}/wiki/pages`、`/repos/{o}/{r}/issue_templates`、`/repos/{o}/{r}/assignees`、全局 `/notifications`、`/user/orgs` |
| **活（文档化，对照证实）** | `/repos/{o}/{r}/commits`、`/issues`、`/pulls`、`/repos/{o}/{r}/notifications`、`/users/{login}/orgs`、`/user/repos` |

**关键教训：文档的 slug 拼写（连字符）与文档分类（user-reactions）都会错；
唯一权威是 live 服务。** 第二轮探针（2026-11-04）补充方法论：GitCode 的
404 有三种 body——框架级 `{"error_msg":"404 Route Not Found"}`（路由不存
在）、GitCode 业务错误 `{"error_code":404,...}`（**路由存在**）、Spring 形
`{"timestamp":...}`（歧义，需 verb 矩阵排除）——body 本身就是路由存在性
的判据；文档化路由同样会死（/protect-branches 有页但 404）。

### 第二轮探针补充（2026-11-04）

| 结论 | 端点 |
|---|---|
| **活（业务错误路由实证）** | `PUT/DELETE /repos/{o}/{r}/branches/{name}/setting`（分支保护更新/移除，要求已有保护；body 形状未 live 校验）、`/org/{org}/kanban/list`（计数信封）、`/org/{org}/kanban/{id}/detail`、`/org/{org}/kanban/{id}/item_list`、`/org/{org}/kanban/{id}/add_item`、`/org/{org}/kanban/{id}/state`、`/org/{org}/kanban/{id}/remove_item`（看板全族，KANBAN_NON_EXISTENT 业务路由实证；id 是字符串） |
| **死（live 404，含文档化）** | `/protect-branches`（有文档页但 GET/POST 全死）、`/branches/{name}/setting/new`、`/branches/{name}/wildcard/setting`、`/commits/{sha}/status`（combined）、全局 `/notifications` |
| **活（文档化，live 200）** | `/compare/{base}...{head}`、`/git/trees/{sha}`、`/git/blobs/{blob_sha}`（用 commit SHA 探会 404，需真实 blob SHA）、`/search/repositories`、`/raw/{path}?ref=`、`/contents/{path}?ref=`、`/orgs/{org}/discuss`、`/pulls`、`/repos/{o}/{r}/notifications` | 三处文档谎言已实证：pull-requests（真 wire 是
下划线）、user-reactions 列表（真 wire 无此路由）、Gitea 血统面（引擎
并非照单全收 Gitea API）。

## 方言 A：v5/v8 文档化且 live 证实或高置信 —— 契约即文档

| 面 | 文件 | 说明 |
|---|---|---|
| 仓库/分支/文件/commits/compare | repos.go, branches.go, files.go, commits.go, commits_enhanced.go | |
| Issues 全族（文档化部分） | issues.go, issues_enhanced.go, enterprise_issues.go | 评论列表分页、modify-history、related-branches、pull_requests 链接（下划线，live 证实）、operate_logs（下划线，live 证实） |
| Pulls 全族 | pulls.go, pulls_enhanced.go | approval-reviewers/testers、files-json、discussions 评论 |
| Labels/Milestones | labels.go, milestones.go | |
| Webhooks | webhooks.go | v0.8.0：password + 五族布尔旗标 |
| Releases | releases.go, releases_enhanced.go | |
| Users/Orgs/Search | users.go, orgs.go, search.go | org 归属走 /users/{login}/orgs（live 200）；/user/orgs 死 |
| Notifications（仓库域） | notifications_enhanced.go | live 200；**全局域死（live 404）**，全局方法保留但调用会 404 |
| Dashboard kanban / Discussion | dashboard_kanban.go, repos_discussions.go | v5 文档分类存在（未 live 验证） |
| Enterprise（v8） | enterprise.go | v8-only；数字 id 寻址；label 端点文档响应损坏已标注 |
| Enterprise（v5 遗存） | enterprise_v5.go | 仅 v5 文档存在的 4 端点 |
| Actions（v8） | actions_v8.go | v8-only；shared/K8S runner 变体沿用信封（已注释标注） |

## 方言 B：live 实证死路径（2026-11-03）—— 能力已从主仓 gitcode 后端移除

以下 SDK 方法保留（无害），但**不得在主仓 gitcode 后端宣称对应能力**；
新增消费前必须先 live 复核。主仓侧处置见 go-git-platform v0.101.0
CHANGELOG。

| 面 | 文件 | 主仓消费 |
|---|---|---|
| Commit status 族 | repos_statuses.go | **是** → CommitStatusManager 能力已移除 |
| Deploy keys | repos_deploy_keys.go | **是** → DeployKeyManager 能力已移除 |
| Reactions 族 | repos_reactions.go | **是** → ReactionManager 能力已移除 |
| 全局 notifications | notifications_enhanced.go | **是** → 仓库域保留、全局方法移除 |
| Issues timeline/subscribers/dependencies | repos_issues_timeline.go | 否 |
| Wiki | repos_wiki.go | 否 |
| Git refs/tags | git_refs.go, git_tags.go | 否（仓库 tags 列表走 repos.go 的文档化端点） |
| Issue/label/merge 模板 | repos_templates.go | 否 |
| Repo assignees | repos_assignees.go | 否 |
| /user/orgs | types.go | 否 |

## 方言 C：确认虚构（处置存档）

| 面 | 处置 | 版本 |
|---|---|---|
| /orgs/{org}/hooks 组织级 webhook | 整文件删除（v5/v8/AtomGit 均无文档） | v0.9.0 |
| Actions /api/v5 GitHub 形状 | 按 v8 文档整面重写 | v0.10.0 |
| Enterprise /api/v5 名字寻址 | 按 v8 文档整面重写；v5-only 四端点回迁 enterprise_v5.go | v0.11.0 / v0.12.0 |
| webhook secret/events 数组 | 按文档改 password + 五族布尔 | v0.8.0 |
| milestone due_date/id 寻址 | 按文档改 due_on/number+id 兜底 | v0.8.0 |
| Issue 链接 pull_requests 拼写 | **回滚**：v0.12.0 按文档改连字符，live 实证下划线才是真 wire，v0.13.0 回滚 | v0.12.0→v0.13.0 |

## 复跑审计

```bash
curl -s https://docs.gitcode.com/sitemap.xml -o /tmp/sitemap.xml
# 提取 /docs/apis/ 的 <loc>，按归一化规则（动词+版本剥除、多词参数保护、
# 参数词表）与 SDK 路径 diff；白名单：actions_v8.go、enterprise.go。
# live 探针：以真实仓库+真实 commit SHA/issue number 逐端点只读 GET，
# 文档化端点作对照；token 用项目自带测试凭据（git 历史 v0.7.2 tag）。
```

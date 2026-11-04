package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	gitcode "github.com/yi-nology/go-gitcode"
)

func main() {
	// ============================================================
	// 1. 基础用法
	// ============================================================
	client := gitcode.NewClient("your-gitcode-token")
	ctx := context.Background()

	// 获取当前用户
	user, err := client.GetCurrentUser(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("当前用户: %s (%s)\n", user.Name, user.Login)

	// ============================================================
	// 2. 认证方式切换
	// ============================================================
	client.SetAuthStyle(gitcode.AuthStylePrivateToken) // PRIVATE-TOKEN Header
	client.SetAuthStyle(gitcode.AuthStyleAccessToken)  // access_token Query
	client.SetAuthStyle(gitcode.AuthStyleBearer)       // Bearer (默认)

	// ============================================================
	// 3. 自定义 HTTP 客户端 & 超时
	// ============================================================
	client.SetHTTPClient(&http.Client{Timeout: 60 * time.Second})

	// 私有部署 — 三种方式
	// 方式一：构造时指定
	// client := gitcode.NewClientWithBaseURL("https://your-gitcode.com/api/v5", "token")
	// 方式二：构造后修改
	// client.SetBaseURL("https://your-gitcode.com/api/v5")
	// 方式三：环境变量 (推荐用于 CLI / CI)
	// export GITCODE_URL="https://your-gitcode.com/api/v5"
	// export GITCODE_TOKEN="your-token"
	// client := gitcode.NewClientFromEnv()

	// ============================================================
	// 4. Rate Limit 重试策略
	// ============================================================
	client.SetRetryPolicy(gitcode.RetryPolicy{
		MaxRetries:           3,
		InitialBackoff:       1 * time.Second,
		MaxBackoff:           30 * time.Second,
		Multiplier:           2.0,
		RetryableStatusCodes: []int{429, 500, 502, 503, 504},
	})

	// ============================================================
	// 5. 请求/响应中间件 (Hooks)
	// ============================================================
	client.AddRequestHook(func(req *http.Request) error {
		fmt.Printf("[REQ] %s %s\n", req.Method, req.URL.Path)
		return nil
	})

	client.AddResponseHook(func(resp *http.Response) error {
		fmt.Printf("[RESP] %d %s\n", resp.StatusCode, resp.Request.URL.Path)
		return nil
	})

	// ============================================================
	// 6. 结构化错误处理
	// ============================================================
	_, err = client.GetRepository(ctx, "owner", "nonexistent")
	if err != nil {
		if gitcode.IsNotFound(err) {
			fmt.Println("仓库不存在")
		} else if gitcode.IsUnauthorized(err) {
			fmt.Println("认证失败，请检查 Token")
		} else if gitcode.IsRateLimit(err) {
			fmt.Println("触发限流，请稍后重试")
		} else if gitcode.IsForbidden(err) {
			fmt.Println("权限不足")
		} else if gitcode.IsConflict(err) {
			fmt.Println("资源已存在")
		} else {
			fmt.Printf("其他错误: %v\n", err)
		}
	}

	// ============================================================
	// 7. 自动分页遍历
	// ============================================================
	allIssues, err := gitcode.CollectAll(ctx, func(opts gitcode.ListOptions) ([]*gitcode.Issue, error) {
		return client.ListIssues(ctx, "owner", "repo", gitcode.ListIssuesOptions{
			ListOptions: opts,
			State:       gitcode.IssueStateOpen,
		})
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\n所有 Open Issues: %d 个\n", len(allIssues))

	// 使用迭代器逐页处理
	for items, err := range gitcode.Paginate(ctx, func(opts gitcode.ListOptions) ([]*gitcode.Repository, error) {
		return client.ListRepositories(ctx, gitcode.ListRepositoriesOptions{ListOptions: opts})
	}) {
		if err != nil {
			log.Fatal(err)
		}
		for _, repo := range items {
			fmt.Printf("  - %s\n", repo.FullName)
		}
	}

	// ============================================================
	// 8. Webhook 签名验证
	// ============================================================
	// 在 HTTP handler 中使用:
	// func handleWebhook(w http.ResponseWriter, r *http.Request) {
	//     payload, _ := io.ReadAll(r.Body)
	//     signature := r.Header.Get("X-Gitcode-Signature")
	//     if !gitcode.VerifyWebhookSignature(payload, "your-webhook-secret", signature) {
	//         http.Error(w, "Invalid signature", http.StatusUnauthorized)
	//         return
	//     }
	//     event, _ := client.ParsePushEvent(payload)
	//     fmt.Printf("Push to %s\n", event.Ref)
	// }

	// 计算签名 (用于测试)
	sig := gitcode.ComputeWebhookSignature([]byte("payload"), "secret")
	fmt.Printf("\nWebhook 签名: %s\n", sig)

	// ============================================================
	// 9. 文件上传 (Multipart)
	// ============================================================
	// 上传文件内容
	result, err := client.UploadFileBytes(ctx, "owner", "repo", "hello.txt", []byte("Hello World"))
	if err != nil {
		fmt.Printf("上传失败: %v\n", err)
	} else {
		fmt.Printf("上传成功: %s\n", result.FilePath)
	}

	// ============================================================
	// 10. OAuth 2.0 流程
	// ============================================================
	oauth := gitcode.NewOAuthClient("client-id", "client-secret", "https://app.example.com/callback")

	// 1. 生成授权 URL
	authURL := oauth.AuthorizeURL("user_info projects", "random-state")
	fmt.Printf("\n请访问: %s\n", authURL)

	// 2. 用 code 换 token (在回调 handler 中)
	// token, err := oauth.ExchangeToken(ctx, code)
	// if err != nil { log.Fatal(err) }

	// 3. 用 token 创建客户端
	// client := gitcode.NewClientFromOAuthToken(token)

	// 4. 刷新 token
	// token, err = oauth.RefreshToken(ctx, token.RefreshToken)

	// ============================================================
	// 11. Actions/CI 示例（v8 流水线 wire）
	// ============================================================
	workflows, err := client.ListActionWorkflows(ctx, "owner", "repo", gitcode.ListOptions{})
	if err != nil {
		fmt.Printf("获取工作流失败: %v\n", err)
	} else {
		fmt.Printf("\n工作流列表: %d 个\n", workflows.TotalCount)
		for _, wf := range workflows.Workflows {
			fmt.Printf("  - %s (%s)\n", wf.Name, wf.State)
		}
	}

	// 获取运行记录
	runs, err := client.ListActionRuns(ctx, "owner", "repo", gitcode.ListWorkflowRunsOptions{
		ListOptions: gitcode.ListOptions{PerPage: 5},
		Status:      "COMPLETED",
	})
	if err != nil {
		fmt.Printf("获取运行记录失败: %v\n", err)
	} else {
		fmt.Printf("\n最近运行记录: %d 个\n", runs.TotalCount)
		for _, run := range runs.WorkflowRuns {
			fmt.Printf("  - %s %s (%s)\n", run.WorkflowRunID, run.Title, run.Status)
		}
	}

	// ============================================================
	// 12. 企业 API 示例
	// ============================================================
	enterpriseID := "12345" // v8 企业面按数字 id 寻址（GetOrgEnterprise 可查组织关联企业的 id）

	// 企业里程碑
	milestones, err := client.ListEnterpriseMilestones(ctx, enterpriseID, gitcode.ListOptions{})
	if err != nil {
		fmt.Printf("获取企业里程碑失败: %v\n", err)
	} else {
		fmt.Printf("\n企业里程碑: %d 个\n", len(milestones))
	}

	// 企业自定义角色
	roles, err := client.ListEnterpriseCustomizedRoles(ctx, enterpriseID)
	if err != nil {
		fmt.Printf("获取企业角色失败: %v\n", err)
	} else {
		fmt.Printf("\n企业自定义角色: %d 个\n", len(roles))
	}

	// ============================================================
	// 13. 看板 API 示例
	// ============================================================
	org := "my-org"
	boards, err := client.ListKanbanBoards(ctx, org, gitcode.ListKanbanBoardsOptions{})
	if err != nil {
		fmt.Printf("获取看板失败: %v\n", err)
	} else {
		fmt.Printf("\n看板列表: %d 个\n", boards.AllCount)
		for _, board := range boards.Content {
			fmt.Printf("  - %s (ID: %s)\n", board.Name, board.ID)
		}
	}

	fmt.Println("\n示例完成！")
}

package gitcode

import (
	"context"
	"net/http"
)

// GetCommitDiff returns the diff of a commit.
//
// GET /repos/{owner}/{repo}/commits/{sha}.diff
func (c *Client) GetCommitDiff(ctx context.Context, owner, repo, sha string) ([]byte, error) {
	return c.doRawRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/commits/%s.diff", owner, repo, sha))
}

// GetCommitPatch returns the patch of a commit.
//
// GET /repos/{owner}/{repo}/commits/{sha}.patch
func (c *Client) GetCommitPatch(ctx context.Context, owner, repo, sha string) ([]byte, error) {
	return c.doRawRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/commits/%s.patch", owner, repo, sha))
}

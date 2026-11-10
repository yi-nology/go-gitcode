package gitcode

import (
	"context"
	"net/http"
)

// GetLatestRelease gets the latest release for a repository.
//
// GET /repos/{owner}/{repo}/releases/latest
func (c *Client) GetLatestRelease(ctx context.Context, owner, repo string) (*Release, error) {
	var release Release
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/releases/latest", owner, repo), nil, &release)
	if err != nil {
		return nil, err
	}
	return &release, nil
}

// ReleaseUploadURL represents the upload URL for a release asset.
type ReleaseUploadURL struct {
	UploadURL string `json:"upload_url"`
	AssetID   int64  `json:"asset_id,omitempty"`
}

// GetReleaseUploadURL gets the upload URL for a release asset.
//
// POST /repos/{owner}/{repo}/releases/{id}/assets/upload_url
func (c *Client) GetReleaseUploadURL(ctx context.Context, owner, repo string, releaseID int64, name, label string) (*ReleaseUploadURL, error) {
	var result ReleaseUploadURL
	body := map[string]string{"name": name}
	if label != "" {
		body["label"] = label
	}
	err := c.doRequest(ctx, http.MethodPost, pathf("/repos/%s/%s/releases/%d/assets/upload_url", owner, repo, releaseID), body, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DownloadReleaseAsset downloads a release asset by ID.
//
// GET /repos/{owner}/{repo}/releases/assets/{id}
func (c *Client) DownloadReleaseAsset(ctx context.Context, owner, repo string, assetID int64) ([]byte, error) {
	return c.doRawRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/releases/assets/%d", owner, repo, assetID))
}

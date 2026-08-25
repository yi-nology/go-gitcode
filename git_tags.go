package gitcode

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// AnnotatedTag represents an annotated git tag.
type AnnotatedTag struct {
	SHA     string        `json:"sha"`
	URL     string        `json:"url"`
	Tag     string        `json:"tag"`
	Message string        `json:"message"`
	Object  *GitObject    `json:"object"`
	Tagger  *CommitAuthor `json:"tagger"`
}

// CreateAnnotatedTagOptions specifies options for creating an annotated tag.
type CreateAnnotatedTagOptions struct {
	Tag     string `json:"tag"`
	Message string `json:"message"`
	Object  string `json:"object"` // SHA of the object to tag
	Type    string `json:"type"`   // commit, tree, blob
	Tagger  *struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Date  string `json:"date"`
	} `json:"tagger,omitempty"`
}

// GetAnnotatedTag gets an annotated tag by SHA.
//
// GET /repos/{owner}/{repo}/git/tags/{sha}
func (c *Client) GetAnnotatedTag(ctx context.Context, owner, repo, sha string) (*AnnotatedTag, error) {
	var tag AnnotatedTag
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/git/tags/%s", owner, repo, sha), nil, &tag)
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// CreateAnnotatedTag creates a new annotated tag.
//
// POST /repos/{owner}/{repo}/git/tags
func (c *Client) CreateAnnotatedTag(ctx context.Context, owner, repo string, opts CreateAnnotatedTagOptions) (*AnnotatedTag, error) {
	var tag AnnotatedTag
	err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/git/tags", owner, repo), opts, &tag)
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// TagWithRelease represents a tag with its associated release info.
type TagWithRelease struct {
	Name    string `json:"name"`
	Message string `json:"message,omitempty"`
	Commit  *struct {
		SHA string `json:"sha"`
		URL string `json:"url"`
	} `json:"commit"`
	ZipballURL string    `json:"zipball_url,omitempty"`
	TarballURL string    `json:"tarball_url,omitempty"`
	CreatedAt  time.Time `json:"created_at,omitempty"`
}

// ListTagsWithOptions lists all tags with pagination options.
//
// GET /repos/{owner}/{repo}/tags
func (c *Client) ListTagsWithOptions(ctx context.Context, owner, repo string, opts ListOptions) ([]*Tag, error) {
	var tags []*Tag
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/tags?%s", owner, repo, opts.toQuery()), nil, &tags)
	if err != nil {
		return nil, err
	}
	return tags, nil
}

// GetTag gets a single tag by name.
//
// GET /repos/{owner}/{repo}/tags/{tag}
func (c *Client) GetTag(ctx context.Context, owner, repo, tagName string) (*Tag, error) {
	var tag Tag
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/tags/%s", owner, repo, tagName), nil, &tag)
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// CreateTag creates a new lightweight tag.
//
// POST /repos/{owner}/{repo}/tags
func (c *Client) CreateTag(ctx context.Context, owner, repo string, opts CreateTagOptions) (*Tag, error) {
	var tag Tag
	err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/tags", owner, repo), opts, &tag)
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// CreateTagOptions specifies options for creating a lightweight tag.
type CreateTagOptions struct {
	TagName string `json:"tag_name"`
	Target  string `json:"target"` // SHA
	Message string `json:"message,omitempty"`
}

// GetReleaseByTag gets a release by its tag name.
//
// GET /repos/{owner}/{repo}/releases/tags/{tag}
func (c *Client) GetReleaseByTag(ctx context.Context, owner, repo, tag string) (*Release, error) {
	var release Release
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/releases/tags/%s", owner, repo, tag), nil, &release)
	if err != nil {
		return nil, err
	}
	return &release, nil
}

// UpdateRelease updates a release.
//
// PATCH /repos/{owner}/{repo}/releases/{id}
func (c *Client) UpdateRelease(ctx context.Context, owner, repo string, releaseID int64, opts UpdateReleaseOptions) (*Release, error) {
	var release Release
	err := c.doRequest(ctx, http.MethodPatch, fmt.Sprintf("/repos/%s/%s/releases/%d", owner, repo, releaseID), opts, &release)
	if err != nil {
		return nil, err
	}
	return &release, nil
}

// UpdateReleaseOptions specifies options for updating a release.
type UpdateReleaseOptions struct {
	TagName         string `json:"tag_name,omitempty"`
	TargetCommitish string `json:"target_commitish,omitempty"`
	Name            string `json:"name,omitempty"`
	Body            string `json:"body,omitempty"`
	Draft           *bool  `json:"draft,omitempty"`
	Prerelease      *bool  `json:"prerelease,omitempty"`
}

// GetRelease gets a single release by ID.
//
// GET /repos/{owner}/{repo}/releases/{id}
func (c *Client) GetRelease(ctx context.Context, owner, repo string, releaseID int64) (*Release, error) {
	var release Release
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/releases/%d", owner, repo, releaseID), nil, &release)
	if err != nil {
		return nil, err
	}
	return &release, nil
}

// DeleteReleaseByID deletes a release by ID.
//
// DELETE /repos/{owner}/{repo}/releases/{id}
func (c *Client) DeleteReleaseByID(ctx context.Context, owner, repo string, releaseID int64) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/repos/%s/%s/releases/%d", owner, repo, releaseID), nil, nil)
}

// ReleaseAsset represents a release asset.
type ReleaseAsset struct {
	ID                 int64     `json:"id"`
	Name               string    `json:"name"`
	ContentType        string    `json:"content_type"`
	Size               int64     `json:"size"`
	DownloadCount      int       `json:"download_count"`
	BrowserDownloadURL string    `json:"browser_download_url"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// ListReleaseAssets lists all assets for a release.
//
// GET /repos/{owner}/{repo}/releases/{id}/assets
func (c *Client) ListReleaseAssets(ctx context.Context, owner, repo string, releaseID int64, opts ListOptions) ([]*ReleaseAsset, error) {
	var assets []*ReleaseAsset
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/releases/%d/assets?%s", owner, repo, releaseID, opts.toQuery()), nil, &assets)
	if err != nil {
		return nil, err
	}
	return assets, nil
}

// GetReleaseAsset gets a single release asset.
//
// GET /repos/{owner}/{repo}/releases/assets/{id}
func (c *Client) GetReleaseAsset(ctx context.Context, owner, repo string, assetID int64) (*ReleaseAsset, error) {
	var asset ReleaseAsset
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/releases/assets/%d", owner, repo, assetID), nil, &asset)
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

// DeleteReleaseAsset deletes a release asset.
//
// DELETE /repos/{owner}/{repo}/releases/assets/{id}
func (c *Client) DeleteReleaseAsset(ctx context.Context, owner, repo string, assetID int64) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/repos/%s/%s/releases/assets/%d", owner, repo, assetID), nil, nil)
}

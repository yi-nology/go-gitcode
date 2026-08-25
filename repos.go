package gitcode

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type Repository struct {
	ID        int64  `json:"id"`
	FullName  string `json:"full_name"`
	Name      string `json:"name"`
	Path      string `json:"path,omitempty"`
	Owner     *User  `json:"owner"`
	Namespace *struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
		Path string `json:"path"`
	} `json:"namespace,omitempty"`
	Description     string     `json:"description"`
	CloneURL        string     `json:"http_url_to_repo"`
	SSHURL          string     `json:"ssh_url_to_repo"`
	HTMLURL         string     `json:"web_url"`
	DefaultBranch   string     `json:"default_branch"`
	Private         bool       `json:"private"`
	Public          bool       `json:"public,omitempty"`
	Fork            bool       `json:"fork"`
	StarsCount      int        `json:"stargazers_count"`
	ForksCount      int        `json:"forks_count"`
	WatchersCount   int        `json:"watchers_count"`
	OpenIssuesCount int        `json:"open_issues_count"`
	Language        string     `json:"language,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	PushedAt        *time.Time `json:"pushed_at,omitempty"`
}

type CreateRepositoryOptions struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Private     *bool  `json:"private,omitempty"`
	AutoInit    *bool  `json:"auto_init,omitempty"`
}

type UpdateRepositoryOptions struct {
	Name          string `json:"name,omitempty"`
	Description   string `json:"description,omitempty"`
	DefaultBranch string `json:"default_branch,omitempty"`
	Private       *bool  `json:"private,omitempty"`
}

type ListRepositoriesOptions struct {
	ListOptions
	Owner string `json:"owner,omitempty"`
	Type  string `json:"type,omitempty"`
	Sort  string `json:"sort,omitempty"`
}

func (c *Client) ListRepositories(ctx context.Context, opts ListRepositoriesOptions) ([]*Repository, error) {
	var repos []*Repository
	path := fmt.Sprintf("/user/repos?%s", opts.toQuery())
	if opts.Owner != "" {
		path = fmt.Sprintf("/users/%s/repos?%s", opts.Owner, opts.toQuery())
	}
	err := c.doRequest(ctx, http.MethodGet, path, nil, &repos)
	if err != nil {
		return nil, err
	}
	return repos, nil
}

func (c *Client) GetRepository(ctx context.Context, owner, repo string) (*Repository, error) {
	var r Repository
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s", owner, repo), nil, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (c *Client) CreateRepository(ctx context.Context, opts CreateRepositoryOptions) (*Repository, error) {
	var r Repository
	err := c.doRequest(ctx, http.MethodPost, "/user/repos", opts, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (c *Client) UpdateRepository(ctx context.Context, owner, repo string, opts UpdateRepositoryOptions) (*Repository, error) {
	var r Repository
	err := c.doRequest(ctx, http.MethodPatch, fmt.Sprintf("/repos/%s/%s", owner, repo), opts, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (c *Client) DeleteRepository(ctx context.Context, owner, repo string) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/repos/%s/%s", owner, repo), nil, nil)
}

func (c *Client) ForkRepository(ctx context.Context, owner, repo string, opts *CreateRepositoryOptions) (*Repository, error) {
	var r Repository
	err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/forks", owner, repo), opts, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

type RepositoryContent struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	Type     string `json:"type"`
	Content  string `json:"content,omitempty"`
	Encoding string `json:"encoding,omitempty"`
	SHA      string `json:"sha"`
	Links    struct {
		Self string `json:"self"`
		Git  string `json:"git"`
	} `json:"_links"`
}

func (c *Client) GetRepositoryContent(ctx context.Context, owner, repo, path, ref string) (*RepositoryContent, error) {
	var content RepositoryContent
	url := fmt.Sprintf("/repos/%s/%s/contents/%s", owner, repo, path)
	if ref != "" {
		url += "?ref=" + ref
	}
	err := c.doRequest(ctx, http.MethodGet, url, nil, &content)
	if err != nil {
		return nil, err
	}
	return &content, nil
}

func (c *Client) ListRepositoryContents(ctx context.Context, owner, repo, path, ref string) ([]*RepositoryContent, error) {
	var contents []*RepositoryContent
	url := fmt.Sprintf("/repos/%s/%s/contents/%s", owner, repo, path)
	if ref != "" {
		url += "?ref=" + ref
	}
	err := c.doRequest(ctx, http.MethodGet, url, nil, &contents)
	if err != nil {
		return nil, err
	}
	return contents, nil
}

type CreateFileOptions struct {
	Message string `json:"message"`
	Content string `json:"content"`
	Branch  string `json:"branch,omitempty"`
}

type UpdateFileOptions struct {
	Message string `json:"message"`
	Content string `json:"content"`
	SHA     string `json:"sha"`
	Branch  string `json:"branch,omitempty"`
}

type DeleteFileOptions struct {
	Message string `json:"message"`
	SHA     string `json:"sha"`
	Branch  string `json:"branch,omitempty"`
}

type FileResult struct {
	Content *RepositoryContent `json:"content"`
	Commit  *Commit            `json:"commit"`
}

func (c *Client) CreateFile(ctx context.Context, owner, repo, path string, opts CreateFileOptions) (*FileResult, error) {
	var result FileResult
	err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/contents/%s", owner, repo, path), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateFile(ctx context.Context, owner, repo, path string, opts UpdateFileOptions) (*FileResult, error) {
	var result FileResult
	err := c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/repos/%s/%s/contents/%s", owner, repo, path), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteFile(ctx context.Context, owner, repo, path string, opts DeleteFileOptions) (*FileResult, error) {
	var result FileResult
	err := c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/repos/%s/%s/contents/%s", owner, repo, path), opts, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

type Tag struct {
	Name   string `json:"name"`
	Commit struct {
		SHA string `json:"sha"`
	} `json:"commit"`
}

func (c *Client) ListTags(ctx context.Context, owner, repo string) ([]*Tag, error) {
	var tags []*Tag
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/tags?per_page=100", owner, repo), nil, &tags)
	if err != nil {
		return nil, err
	}
	return tags, nil
}

type Release struct {
	ID              int64     `json:"id"`
	TagName         string    `json:"tag_name"`
	TargetCommitish string    `json:"target_commitish,omitempty"`
	Name            string    `json:"name"`
	Body            string    `json:"body"`
	HTMLURL         string    `json:"html_url,omitempty"`
	Draft           bool      `json:"draft"`
	Prerelease      bool      `json:"prerelease"`
	CreatedAt       time.Time `json:"created_at"`
	PublishedAt     time.Time `json:"published_at,omitempty"`
}

type CreateReleaseOptions struct {
	TagName    string `json:"tag_name"`
	Target     string `json:"target_commitish,omitempty"`
	Title      string `json:"name"`
	Body       string `json:"body,omitempty"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
}

func (c *Client) ListReleases(ctx context.Context, owner, repo string) ([]*Release, error) {
	var releases []*Release
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/releases?per_page=100", owner, repo), nil, &releases)
	if err != nil {
		return nil, err
	}
	return releases, nil
}

func (c *Client) CreateRelease(ctx context.Context, owner, repo string, opts CreateReleaseOptions) (*Release, error) {
	var r Release
	err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/releases", owner, repo), opts, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (c *Client) DeleteRelease(ctx context.Context, owner, repo string, tagName string) error {
	err := c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/repos/%s/%s/releases/tags/%s", owner, repo, tagName), nil, nil)
	if err == nil {
		return nil
	}
	return c.DeleteTag(ctx, owner, repo, tagName)
}

func (c *Client) DeleteTag(ctx context.Context, owner, repo, tagName string) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/repos/%s/%s/tags/%s", owner, repo, tagName), nil, nil)
}

type Contributor struct {
	ID            string `json:"id"`
	Login         string `json:"login"`
	AvatarURL     string `json:"avatar_url"`
	Contributions int    `json:"contributions"`
}

func (c *Client) ListContributors(ctx context.Context, owner, repo string) ([]*Contributor, error) {
	var contributors []*Contributor
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/contributors?per_page=100", owner, repo), nil, &contributors)
	if err != nil {
		return nil, err
	}
	return contributors, nil
}

type Language map[string]int

func (c *Client) GetLanguages(ctx context.Context, owner, repo string) (Language, error) {
	var lang Language
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/languages", owner, repo), nil, &lang)
	if err != nil {
		return nil, err
	}
	return lang, nil
}

func (c *Client) GetRawFile(ctx context.Context, owner, repo, path, ref string) ([]byte, error) {
	url := fmt.Sprintf("/repos/%s/%s/raw/%s", owner, repo, path)
	if ref != "" {
		url += "?ref=" + ref
	}
	return c.doRawRequest(ctx, http.MethodGet, url)
}

type GitTree struct {
	SHA       string          `json:"sha"`
	URL       string          `json:"url"`
	Truncated bool            `json:"truncated"`
	Tree      []*GitTreeEntry `json:"tree"`
}

type GitTreeEntry struct {
	Path string `json:"path"`
	Mode string `json:"mode"`
	Type string `json:"type"`
	SHA  string `json:"sha"`
	Size int64  `json:"size"`
	URL  string `json:"url"`
}

func (c *Client) GetTree(ctx context.Context, owner, repo, sha string, recursive bool) (*GitTree, error) {
	var tree GitTree
	url := fmt.Sprintf("/repos/%s/%s/git/trees/%s", owner, repo, sha)
	if recursive {
		url += "?recursive=1"
	}
	err := c.doRequest(ctx, http.MethodGet, url, nil, &tree)
	if err != nil {
		return nil, err
	}
	return &tree, nil
}

type GitBlob struct {
	SHA      string `json:"sha"`
	Size     int64  `json:"size"`
	URL      string `json:"url"`
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
}

func (c *Client) GetBlob(ctx context.Context, owner, repo, sha string) (*GitBlob, error) {
	var blob GitBlob
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/git/blobs/%s", owner, repo, sha), nil, &blob)
	if err != nil {
		return nil, err
	}
	return &blob, nil
}

func (c *Client) ListForks(ctx context.Context, owner, repo string, opts ListOptions) ([]*Repository, error) {
	var forks []*Repository
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/forks?%s", owner, repo, opts.toQuery()), nil, &forks)
	if err != nil {
		return nil, err
	}
	return forks, nil
}

func (c *Client) ListWatchers(ctx context.Context, owner, repo string, opts ListOptions) ([]*User, error) {
	var watchers []*User
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/subscribers?%s", owner, repo, opts.toQuery()), nil, &watchers)
	if err != nil {
		return nil, err
	}
	return watchers, nil
}

func (c *Client) ListStargazers(ctx context.Context, owner, repo string, opts ListOptions) ([]*User, error) {
	var stargazers []*User
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/stargazers?%s", owner, repo, opts.toQuery()), nil, &stargazers)
	if err != nil {
		return nil, err
	}
	return stargazers, nil
}

type ContributorStatistic struct {
	Author *User `json:"author"`
	Total  int   `json:"total"`
	Weeks  []*struct {
		W string `json:"w"`
		A int    `json:"a"`
		D int    `json:"d"`
		C int    `json:"c"`
	} `json:"weeks"`
}

func (c *Client) GetContributorStatistics(ctx context.Context, owner, repo string) ([]*ContributorStatistic, error) {
	var stats []*ContributorStatistic
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/contributors/statistic", owner, repo), nil, &stats)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

type RepoEvent struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"`
	CreatedAt string `json:"created_at"`
	PushData  *struct {
		Ref string `json:"ref"`
	} `json:"push_data"`
}

func (c *Client) ListRepoEvents(ctx context.Context, owner, repo string, opts ListOptions) ([]*RepoEvent, error) {
	var events []*RepoEvent
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/events?%s", owner, repo, opts.toQuery()), nil, &events)
	if err != nil {
		return nil, err
	}
	return events, nil
}

type RepoSettings struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	DefaultBranch string `json:"default_branch"`
	HasIssues     bool   `json:"has_issues"`
	HasWiki       bool   `json:"has_wiki"`
	CanComment    bool   `json:"can_comment"`
	Private       bool   `json:"private"`
}

func (c *Client) GetRepoSettings(ctx context.Context, owner, repo string) (*RepoSettings, error) {
	var settings RepoSettings
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/repo_settings", owner, repo), nil, &settings)
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

func (c *Client) UpdateRepoSettings(ctx context.Context, owner, repo string, opts *RepoSettings) (*RepoSettings, error) {
	var settings RepoSettings
	err := c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/repos/%s/%s/repo_settings", owner, repo), opts, &settings)
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

type CreateOrgRepoOptions struct {
	Name              string `json:"name"`
	Description       string `json:"description,omitempty"`
	Homepage          string `json:"homepage,omitempty"`
	HasIssues         *bool  `json:"has_issues,omitempty"`
	HasWiki           *bool  `json:"has_wiki,omitempty"`
	CanComment        *bool  `json:"can_comment,omitempty"`
	Public            *int   `json:"public,omitempty"`
	Private           *bool  `json:"private,omitempty"`
	AutoInit          *bool  `json:"auto_init,omitempty"`
	GitignoreTemplate string `json:"gitignore_template,omitempty"`
	LicenseTemplate   string `json:"license_template,omitempty"`
	Path              string `json:"path,omitempty"`
	DefaultBranch     string `json:"default_branch,omitempty"`
}

func (c *Client) CreateOrgRepository(ctx context.Context, org string, opts CreateOrgRepoOptions) (*Repository, error) {
	var r Repository
	err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/orgs/%s/repos", org), opts, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (c *Client) ListOrgRepositories(ctx context.Context, org, repoType string, opts ListOptions) ([]*Repository, error) {
	var repos []*Repository
	query := opts.toQuery()
	if repoType != "" {
		query += "&type=" + repoType
	}
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/orgs/%s/repos?%s", org, query), nil, &repos)
	if err != nil {
		return nil, err
	}
	return repos, nil
}

type TransferRepoOptions struct {
	NewOwner string `json:"new_owner"`
}

func (c *Client) TransferRepository(ctx context.Context, owner, repo string, opts TransferRepoOptions) error {
	return c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/transfer", owner, repo), opts, nil)
}

type ArchiveStatus struct {
	Archived bool `json:"archived"`
}

func (c *Client) GetArchiveStatus(ctx context.Context, owner, repo string) (*ArchiveStatus, error) {
	var status ArchiveStatus
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/transition", owner, repo), nil, &status)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (c *Client) ArchiveRepository(ctx context.Context, owner, repo string) error {
	return c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/repos/%s/%s/transition", owner, repo), nil, nil)
}

type UpdateMemberOptions struct {
	Permission string `json:"permission"`
}

func (c *Client) UpdateRepoMember(ctx context.Context, owner, repo, username string, opts UpdateMemberOptions) error {
	return c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/repos/%s/%s/members/%s", owner, repo, username), opts, nil)
}

type FileUploadResult struct {
	FilePath string `json:"file_path"`
}

func (c *Client) UploadFile(ctx context.Context, owner, repo string, filePath string) (*FileUploadResult, error) {
	var result FileUploadResult
	body := map[string]string{"file": filePath}
	err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/file/upload", owner, repo), body, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UploadImage(ctx context.Context, owner, repo string, filePath string) (*FileUploadResult, error) {
	var result FileUploadResult
	body := map[string]string{"file": filePath}
	err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/repos/%s/%s/img/upload", owner, repo), body, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

type PushConfig struct {
	MaxFileSize        int      `json:"max_file_size"`
	ProhibitedFiles    []string `json:"prohibited_files"`
	CommitMessageRegex string   `json:"commit_message_regex"`
}

func (c *Client) GetPushConfig(ctx context.Context, owner, repo string) (*PushConfig, error) {
	var config PushConfig
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/push_config", owner, repo), nil, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (c *Client) UpdatePushConfig(ctx context.Context, owner, repo string, config *PushConfig) error {
	return c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/repos/%s/%s/push_config", owner, repo), config, nil)
}

type PRSettings struct {
	DefaultMergeMethod string `json:"default_merge_method"`
	AutoCloseIssues    bool   `json:"auto_close_issues"`
}

func (c *Client) GetPRSettings(ctx context.Context, owner, repo string) (*PRSettings, error) {
	var settings PRSettings
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/pull_request_settings", owner, repo), nil, &settings)
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

func (c *Client) UpdatePRSettings(ctx context.Context, owner, repo string, settings *PRSettings) error {
	return c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/repos/%s/%s/pull_request_settings", owner, repo), settings, nil)
}

type FileListEntry struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Type string `json:"type"`
}

func (c *Client) ListFiles(ctx context.Context, owner, repo string) ([]*FileListEntry, error) {
	var files []*FileListEntry
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/file_list", owner, repo), nil, &files)
	if err != nil {
		return nil, err
	}
	return files, nil
}

type ModuleSetting struct {
	Issues   bool `json:"issues"`
	Wiki     bool `json:"wiki"`
	Releases bool `json:"releases"`
}

func (c *Client) SetModuleSetting(ctx context.Context, owner, repo string, setting ModuleSetting) error {
	return c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/repos/%s/%s/module/setting", owner, repo), setting, nil)
}

type CustomizedRole struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (c *Client) GetCustomizedRoles(ctx context.Context, owner, repo string) ([]*CustomizedRole, error) {
	var roles []*CustomizedRole
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/customized_roles", owner, repo), nil, &roles)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

type ReviewerConfig struct {
	MinApprovingReviews int  `json:"min_approving_reviews"`
	RequireCodeOwner    bool `json:"require_code_owner"`
}

func (c *Client) UpdateReviewerConfig(ctx context.Context, owner, repo string, config ReviewerConfig) error {
	return c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/repos/%s/%s/reviewer", owner, repo), config, nil)
}

type DownloadStatistic struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func (c *Client) GetDownloadStatistics(ctx context.Context, owner, repo string) ([]*DownloadStatistic, error) {
	var stats []*DownloadStatistic
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/download_statistics", owner, repo), nil, &stats)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func (c *Client) UpdateOrgRepoStatus(ctx context.Context, org, repo, status string) error {
	return c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/org/%s/repo/%s/status", org, repo), map[string]string{"status": status}, nil)
}

func (c *Client) TransferToOrg(ctx context.Context, org, repo, newOwner string) error {
	return c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/org/%s/projects/%s/transfer", org, repo), map[string]string{"new_owner": newOwner}, nil)
}

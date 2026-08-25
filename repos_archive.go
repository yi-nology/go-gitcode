package gitcode

import (
	"context"
	"fmt"
	"net/http"
)

// GetRepositoryArchive downloads a repository archive (tar.gz, zip, etc.).
//
// GET /repos/{owner}/{repo}/archive/{archive}
func (c *Client) GetRepositoryArchive(ctx context.Context, owner, repo, archive string) ([]byte, error) {
	// archive can be: master.tar.gz, main.zip, v1.0.0.tar.gz, etc.
	return c.doRawRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/archive/%s", owner, repo, archive))
}

// DownloadRepository downloads a repository as a zip or tarball.
//
// GET /repos/{owner}/{repo}/{format}
func (c *Client) DownloadRepository(ctx context.Context, owner, repo, format string) ([]byte, error) {
	// format can be: repository/archive.zip, repository/archive.tar.gz
	return c.doRawRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/%s", owner, repo, format))
}

// GenerateRepositoryArchive generates a repository archive.
//
// GET /repos/{owner}/{repo}/archive
func (c *Client) GenerateRepositoryArchive(ctx context.Context, owner, repo, ref, format string) ([]byte, error) {
	path := fmt.Sprintf("/repos/%s/%s/archive", owner, repo)
	if ref != "" || format != "" {
		path += "?"
		if ref != "" {
			path += "ref=" + ref
		}
		if format != "" {
			if ref != "" {
				path += "&"
			}
			path += "format=" + format
		}
	}
	return c.doRawRequest(ctx, http.MethodGet, path)
}

// RepoStats represents repository statistics.
type RepoStats struct {
	Participation *ParticipationStats `json:"participation,omitempty"`
}

// ParticipationStats represents participation statistics (weekly commits).
type ParticipationStats struct {
	All   []int `json:"all"`
	Owner []int `json:"owner"`
}

// GetRepoParticipation returns the total commit counts for the owner and all contributors.
//
// GET /repos/{owner}/{repo}/stats/participation
func (c *Client) GetRepoParticipation(ctx context.Context, owner, repo string) (*ParticipationStats, error) {
	var stats ParticipationStats
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/stats/participation", owner, repo), nil, &stats)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// CodeFrequencyEntry represents a [additions, deletions, timestamp] tuple.
type CodeFrequencyEntry []int

// GetRepoCodeFrequency returns a weekly aggregate of the number of additions and deletions.
//
// GET /repos/{owner}/{repo}/stats/code_frequency
func (c *Client) GetRepoCodeFrequency(ctx context.Context, owner, repo string) ([]CodeFrequencyEntry, error) {
	var entries []CodeFrequencyEntry
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/stats/code_frequency", owner, repo), nil, &entries)
	if err != nil {
		return nil, err
	}
	return entries, nil
}

// CommitActivity represents weekly commit activity.
type CommitActivity struct {
	Days  []int `json:"days"` // Sun-Sat
	Total int   `json:"total"`
	Week  int64 `json:"week"` // Unix timestamp
}

// GetRepoCommitActivity returns a weekly aggregate of the number of commits.
//
// GET /repos/{owner}/{repo}/stats/commit_activity
func (c *Client) GetRepoCommitActivity(ctx context.Context, owner, repo string) ([]*CommitActivity, error) {
	var activity []*CommitActivity
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/stats/commit_activity", owner, repo), nil, &activity)
	if err != nil {
		return nil, err
	}
	return activity, nil
}

// PunchCardEntry represents a [day, hour, commits] tuple.
type PunchCardEntry []int

// GetRepoPunchCard returns the number of commits per hour in each day.
//
// GET /repos/{owner}/{repo}/stats/punch_card
func (c *Client) GetRepoPunchCard(ctx context.Context, owner, repo string) ([]PunchCardEntry, error) {
	var entries []PunchCardEntry
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/repos/%s/%s/stats/punch_card", owner, repo), nil, &entries)
	if err != nil {
		return nil, err
	}
	return entries, nil
}

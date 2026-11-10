package gitcode

import (
	"context"
	"net/http"
)

// RepositoryTopics represents the topics of a repository.
type RepositoryTopics struct {
	Topics []string `json:"topics"`
}

// ListRepositoryTopics lists all topics of a repository.
//
// GET /repos/{owner}/{repo}/topics
func (c *Client) ListRepositoryTopics(ctx context.Context, owner, repo string) ([]string, error) {
	var result RepositoryTopics
	err := c.doRequest(ctx, http.MethodGet, pathf("/repos/%s/%s/topics", owner, repo), nil, &result)
	if err != nil {
		return nil, err
	}
	return result.Topics, nil
}

// UpdateRepositoryTopics replaces all topics of a repository.
//
// PUT /repos/{owner}/{repo}/topics
func (c *Client) UpdateRepositoryTopics(ctx context.Context, owner, repo string, topics []string) ([]string, error) {
	var result RepositoryTopics
	err := c.doRequest(ctx, http.MethodPut, pathf("/repos/%s/%s/topics", owner, repo), map[string]interface{}{"topics": topics}, &result)
	if err != nil {
		return nil, err
	}
	return result.Topics, nil
}

// AddRepositoryTopic adds a single topic to a repository.
//
// PUT /repos/{owner}/{repo}/topics/{topic}
func (c *Client) AddRepositoryTopic(ctx context.Context, owner, repo, topic string) error {
	return c.doRequest(ctx, http.MethodPut, pathf("/repos/%s/%s/topics/%s", owner, repo, topic), nil, nil)
}

// DeleteRepositoryTopic removes a single topic from a repository.
//
// DELETE /repos/{owner}/{repo}/topics/{topic}
func (c *Client) DeleteRepositoryTopic(ctx context.Context, owner, repo, topic string) error {
	return c.doRequest(ctx, http.MethodDelete, pathf("/repos/%s/%s/topics/%s", owner, repo, topic), nil, nil)
}

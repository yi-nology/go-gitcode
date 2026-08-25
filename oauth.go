package gitcode

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	DefaultOAuthBaseURL = "https://gitcode.com"
)

type OAuthClient struct {
	clientID     string
	clientSecret string
	redirectURI  string
	baseURL      string
	httpClient   *http.Client
}

func NewOAuthClient(clientID, clientSecret, redirectURI string) *OAuthClient {
	return &OAuthClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
		baseURL:      DefaultOAuthBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewOAuthClientWithBaseURL creates an OAuthClient for a private GitCode deployment.
func NewOAuthClientWithBaseURL(clientID, clientSecret, redirectURI, baseURL string) *OAuthClient {
	o := NewOAuthClient(clientID, clientSecret, redirectURI)
	o.baseURL = strings.TrimRight(baseURL, "/")
	return o
}

func (o *OAuthClient) SetBaseURL(baseURL string) {
	o.baseURL = strings.TrimRight(baseURL, "/")
}

func (o *OAuthClient) AuthorizeURL(scope, state string) string {
	params := url.Values{}
	params.Set("client_id", o.clientID)
	params.Set("redirect_uri", o.redirectURI)
	params.Set("response_type", "code")
	if scope != "" {
		params.Set("scope", scope)
	}
	if state != "" {
		params.Set("state", state)
	}
	return fmt.Sprintf("%s/oauth/authorize?%s", o.baseURL, params.Encode())
}

type OAuthToken struct {
	AccessToken  string    `json:"access_token"`
	ExpiresIn    int       `json:"expires_in"`
	RefreshToken string    `json:"refresh_token"`
	Scope        string    `json:"scope"`
	CreatedAt    time.Time `json:"created_at"`
}

func (o *OAuthClient) ExchangeToken(ctx context.Context, code string) (*OAuthToken, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("client_id", o.clientID)
	data.Set("client_secret", o.clientSecret)

	return o.doTokenRequest(ctx, data)
}

func (o *OAuthClient) RefreshToken(ctx context.Context, refreshToken string) (*OAuthToken, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	return o.doTokenRequest(ctx, data)
}

func (o *OAuthClient) doTokenRequest(ctx context.Context, data url.Values) (*OAuthToken, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.baseURL+"/oauth/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("OAuth token request failed with %d: %s", resp.StatusCode, string(body))
	}

	var token OAuthToken
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, err
	}
	return &token, nil
}

// NewClient creates a new API Client from an OAuthToken, preserving the OAuthClient's base URL.
// The resulting client uses the same host as the OAuthClient, which is correct for private deployments.
func (o *OAuthClient) NewClient(token *OAuthToken) *Client {
	return NewClientWithBaseURL(o.baseURL+"/api/v5", token.AccessToken)
}

// Deprecated: NewClientFromOAuthToken creates a client from an OAuth token using the default public base URL.
// For private deployments, use OAuthClient.NewClient() instead, which preserves the base URL.
func NewClientFromOAuthToken(token *OAuthToken) *Client {
	return NewClient(token.AccessToken)
}

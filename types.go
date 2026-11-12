package gitcode

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Timestamp struct {
	time.Time
}

func (t *Timestamp) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == "" {
		return nil
	}
	tt, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}
	t.Time = tt
	return nil
}

type NullableTime struct {
	time.Time
	Valid bool
}

func (nt *NullableTime) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == "" {
		nt.Valid = false
		return nil
	}
	tt, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}
	nt.Time = tt
	nt.Valid = true
	return nil
}

type FlexInt int

func (fi *FlexInt) UnmarshalJSON(data []byte) error {
	var n int
	if err := json.Unmarshal(data, &n); err == nil {
		*fi = FlexInt(n)
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	*fi = FlexInt(parsed)
	return nil
}

type FlexString string

func (fs *FlexString) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*fs = FlexString(s)
		return nil
	}
	var n int64
	if err := json.Unmarshal(data, &n); err != nil {
		return err
	}
	*fs = FlexString(strconv.FormatInt(n, 10))
	return nil
}

// FlexBool tolerates the value shapes seen across GitCode deployments for
// boolean flags: JSON true/false, the integers 0/1 (the documented hook
// response schema uses integer flags), and "true"/"false"/"0"/"1" strings.
type FlexBool bool

func (fb *FlexBool) UnmarshalJSON(data []byte) error {
	var b bool
	if err := json.Unmarshal(data, &b); err == nil {
		*fb = FlexBool(b)
		return nil
	}
	var n int
	if err := json.Unmarshal(data, &n); err == nil {
		*fb = n != 0
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "1", "yes", "on":
		*fb = true
	case "false", "0", "no", "off", "":
		*fb = false
	default:
		return fmt.Errorf("invalid boolean value %q", s)
	}
	return nil
}

type Error struct {
	Message string `json:"message"`
	Errors  []struct {
		Resource string `json:"resource"`
		Field    string `json:"field"`
		Code     string `json:"code"`
	} `json:"errors"`
}

func (e *Error) Error() string {
	return e.Message
}

type SearchOptions struct {
	ListOptions
	Query string `json:"q"`
	Order string `json:"order,omitempty"`
}

type SearchResult struct {
	TotalCount int           `json:"total_count"`
	Items      []*Repository `json:"items"`
}

type Notification struct {
	ID      string `json:"id"`
	Reason  string `json:"reason"`
	Unread  bool   `json:"unread"`
	Subject struct {
		Title string `json:"title"`
		URL   string `json:"url"`
		Type  string `json:"type"`
	} `json:"subject"`
	CreatedAt time.Time `json:"created_at"`
}

type Star struct {
	StarredAt time.Time `json:"starred_at"`
}

type Member struct {
	ID    string `json:"id"`
	Login string `json:"login"`
	Role  string `json:"role"`
}

type Organization struct {
	ID          int64  `json:"id"`
	Login       string `json:"login"`
	Name        string `json:"name"`
	Description string `json:"description"`
	AvatarURL   string `json:"avatar_url"`
}

func (c *Client) ListOrganizations(ctx context.Context) ([]*Organization, error) {
	var orgs []*Organization
	err := c.doRequest(ctx, http.MethodGet, "/user/orgs", nil, &orgs)
	if err != nil {
		return nil, fmt.Errorf("ListOrganizations not supported by this GitCode instance: %w", err)
	}
	return orgs, nil
}

func (c *Client) GetOrganization(ctx context.Context, org string) (*Organization, error) {
	var o Organization
	err := c.doRequest(ctx, http.MethodGet, pathf("/orgs/%s", org), nil, &o)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (c *Client) ListOrganizationMembers(ctx context.Context, org string) ([]*Member, error) {
	var members []*Member
	err := c.doRequest(ctx, http.MethodGet, pathf("/orgs/%s/members", org), nil, &members)
	if err != nil {
		return nil, err
	}
	return members, nil
}

func (c *Client) ListNotifications(ctx context.Context) ([]*Notification, error) {
	var notifications []*Notification
	err := c.doRequest(ctx, http.MethodGet, "/notifications", nil, &notifications)
	if err != nil {
		return nil, err
	}
	return notifications, nil
}

func (c *Client) StarRepository(ctx context.Context, owner, repo string) error {
	return c.doRequest(ctx, http.MethodPut, pathf("/user/starred/%s/%s", owner, repo), nil, nil)
}

func (c *Client) UnstarRepository(ctx context.Context, owner, repo string) error {
	return c.doRequest(ctx, http.MethodDelete, pathf("/user/starred/%s/%s", owner, repo), nil, nil)
}

func (c *Client) IsRepositoryStarred(ctx context.Context, owner, repo string) (bool, error) {
	_, err := c.doRawRequest(ctx, http.MethodGet, pathf("/user/starred/%s/%s", owner, repo))
	if err != nil {
		if isNotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

package openim

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Client struct {
	BaseURL     string
	AdminUser   string
	AdminSecret string
	HTTP        *http.Client

	mu          sync.Mutex
	adminToken  string
	adminExpiry time.Time
}

type UserToken struct {
	Token             string
	ExpireTimeSeconds int64
}

type apiEnvelope struct {
	ErrCode int             `json:"errCode"`
	ErrMsg  string          `json:"errMsg"`
	Data    json.RawMessage `json:"data"`
}

type APIError struct {
	Code    int
	Message string
}

func (e *APIError) Error() string { return fmt.Sprintf("OpenIM API error %d: %s", e.Code, e.Message) }

func (c *Client) ProvisionUser(ctx context.Context, userID, displayName string) error {
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(displayName) == "" {
		return errors.New("OpenIM user ID and display name are required")
	}
	adminToken, err := c.adminTokenFor(ctx)
	if err != nil {
		return err
	}
	var existing struct {
		UsersInfo []json.RawMessage `json:"usersInfo"`
		Users     []json.RawMessage `json:"users"`
	}
	err = c.call(ctx, "/user/get_users_info", adminToken, map[string]any{"userIDs": []string{userID}}, &existing)
	if err == nil && (len(existing.UsersInfo) > 0 || len(existing.Users) > 0) {
		return nil
	}
	var apiErr *APIError
	if err != nil && !errors.As(err, &apiErr) {
		return err
	}
	if err != nil && apiErr.Code != 1004 {
		return err
	}

	registerErr := c.call(ctx, "/user/user_register", adminToken, map[string]any{
		"users": []map[string]string{{"userID": userID, "nickname": displayName, "faceURL": ""}},
	}, nil)
	if registerErr == nil {
		return nil
	}
	// A timeout or concurrent retry may have completed the write. Re-read before
	// returning an error so normal retries remain idempotent.
	existing = struct {
		UsersInfo []json.RawMessage `json:"usersInfo"`
		Users     []json.RawMessage `json:"users"`
	}{}
	if readErr := c.call(ctx, "/user/get_users_info", adminToken, map[string]any{"userIDs": []string{userID}}, &existing); readErr == nil && (len(existing.UsersInfo) > 0 || len(existing.Users) > 0) {
		return nil
	}
	return registerErr
}

func (c *Client) GetUserToken(ctx context.Context, userID string, platformID int) (UserToken, error) {
	adminToken, err := c.adminTokenFor(ctx)
	if err != nil {
		return UserToken{}, err
	}
	var data struct {
		Token             string `json:"token"`
		ExpireTimeSeconds int64  `json:"expireTimeSeconds"`
	}
	if err := c.call(ctx, "/auth/get_user_token", adminToken, map[string]any{"platformID": platformID, "userID": userID}, &data); err != nil {
		return UserToken{}, err
	}
	if data.Token == "" {
		return UserToken{}, errors.New("OpenIM returned an empty user token")
	}
	return UserToken{Token: data.Token, ExpireTimeSeconds: data.ExpireTimeSeconds}, nil
}

func (c *Client) ForceLogout(ctx context.Context, userID string, platformID int) error {
	adminToken, err := c.adminTokenFor(ctx)
	if err != nil {
		return err
	}
	return c.call(ctx, "/auth/force_logout", adminToken, map[string]any{"platformID": platformID, "userID": userID}, nil)
}

func (c *Client) adminTokenFor(ctx context.Context) (string, error) {
	c.mu.Lock()
	if c.adminToken != "" && time.Now().Before(c.adminExpiry.Add(-time.Minute)) {
		token := c.adminToken
		c.mu.Unlock()
		return token, nil
	}
	c.mu.Unlock()
	var data struct {
		Token             string `json:"token"`
		ExpireTimeSeconds int64  `json:"expireTimeSeconds"`
	}
	if err := c.call(ctx, "/auth/get_admin_token", "", map[string]any{"secret": c.AdminSecret, "userID": c.AdminUser}, &data); err != nil {
		return "", err
	}
	if data.Token == "" {
		return "", errors.New("OpenIM returned an empty admin token")
	}
	expires := time.Now().Add(time.Duration(data.ExpireTimeSeconds) * time.Second)
	c.mu.Lock()
	c.adminToken, c.adminExpiry = data.Token, expires
	c.mu.Unlock()
	return data.Token, nil
}

func (c *Client) call(ctx context.Context, path, token string, payload any, data any) error {
	if c.HTTP == nil {
		return errors.New("OpenIM HTTP client is not configured")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode OpenIM request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.BaseURL, "/")+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create OpenIM request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("operationID", operationID())
	if token != "" {
		req.Header.Set("token", token)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("call OpenIM API: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("OpenIM API returned HTTP %d", resp.StatusCode)
	}
	var envelope apiEnvelope
	if err := json.NewDecoder(io.LimitReader(resp.Body, 2<<20)).Decode(&envelope); err != nil {
		return fmt.Errorf("decode OpenIM response: %w", err)
	}
	if envelope.ErrCode != 0 {
		return &APIError{Code: envelope.ErrCode, Message: envelope.ErrMsg}
	}
	if data != nil && len(envelope.Data) > 0 && string(envelope.Data) != "null" {
		if err := json.Unmarshal(envelope.Data, data); err != nil {
			return fmt.Errorf("decode OpenIM response data: %w", err)
		}
	}
	return nil
}

func operationID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("op-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

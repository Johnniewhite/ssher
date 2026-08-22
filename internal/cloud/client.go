package cloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type APIError struct {
	Status        int
	Code, Message string
}

func (e *APIError) Error() string { return e.Message }

type Client struct {
	BaseURL, Token string
	HTTP           *http.Client
}

func NewClient(baseURL, token string) *Client {
	return &Client{BaseURL: strings.TrimRight(baseURL, "/"), Token: token, HTTP: &http.Client{Timeout: 25 * time.Second}}
}

func (c *Client) request(ctx context.Context, method, path string, input, output any) (int, error) {
	var body io.Reader
	if input != nil {
		b, err := json.Marshal(input)
		if err != nil {
			return 0, err
		}
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, body)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Accept", "application/json")
	if input != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return 0, fmt.Errorf("SSHer Cloud request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var payload struct {
			Error struct{ Code, Message string } `json:"error"`
		}
		_ = json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&payload)
		if payload.Error.Message == "" {
			payload.Error.Message = resp.Status
		}
		return resp.StatusCode, &APIError{Status: resp.StatusCode, Code: payload.Error.Code, Message: payload.Error.Message}
	}
	if output != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(io.LimitReader(resp.Body, 2<<20)).Decode(output); err != nil {
			return resp.StatusCode, err
		}
	}
	return resp.StatusCode, nil
}

type DeviceCode struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresAt       string `json:"expires_at"`
	Interval        int    `json:"interval"`
}
type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}
type Device struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Name      string `json:"name"`
	Platform  string `json:"platform"`
	PublicKey []byte `json:"public_key"`
}
type OrganizationDevice struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Name        string     `json:"name"`
	Platform    string     `json:"platform"`
	PublicKey   []byte     `json:"public_key"`
	HasEnvelope bool       `json:"has_envelope"`
	RevokedAt   *time.Time `json:"revoked_at,omitempty"`
}
type DeviceToken struct {
	Status  string `json:"status"`
	User    User   `json:"user"`
	Device  Device `json:"device"`
	Session struct {
		Token     string `json:"token"`
		ExpiresAt string `json:"expires_at"`
	} `json:"session"`
}
type Organization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Role string `json:"role"`
}
type Envelope struct {
	OrganizationID     string `json:"organization_id"`
	DeviceID           string `json:"device_id"`
	Generation         int    `json:"generation"`
	EphemeralPublicKey []byte `json:"ephemeral_public_key"`
	Nonce              []byte `json:"nonce"`
	Ciphertext         []byte `json:"ciphertext"`
}
type EncryptedServer struct {
	ID                  string            `json:"id"`
	OrganizationID      string            `json:"organization_id"`
	Ciphertext          []byte            `json:"ciphertext"`
	Nonce               []byte            `json:"nonce"`
	Revision            int64             `json:"revision"`
	TeamIDs             []string          `json:"team_ids"`
	TeamAccessExpiresAt map[string]string `json:"team_access_expires_at,omitempty"`
	DeletedAt           *time.Time        `json:"deleted_at,omitempty"`
}

func (c *Client) CreateDeviceCode(ctx context.Context, name, platform string, publicKey []byte) (DeviceCode, error) {
	var out DeviceCode
	_, err := c.request(ctx, http.MethodPost, "/v1/auth/device-codes", map[string]any{"device_name": name, "platform": platform, "public_key": publicKey}, &out)
	return out, err
}
func (c *Client) ExchangeDeviceCode(ctx context.Context, code string) (DeviceToken, bool, error) {
	var out DeviceToken
	status, err := c.request(ctx, http.MethodPost, "/v1/auth/device-codes/token", map[string]string{"device_code": code}, &out)
	return out, status == http.StatusAccepted, err
}
func (c *Client) Organizations(ctx context.Context) ([]Organization, error) {
	var out struct {
		Organizations []Organization `json:"organizations"`
	}
	_, err := c.request(ctx, http.MethodGet, "/v1/organizations", nil, &out)
	return out.Organizations, err
}
func (c *Client) WorkspaceEnvelope(ctx context.Context, orgID, deviceID string) (Envelope, error) {
	var out Envelope
	_, err := c.request(ctx, http.MethodGet, "/v1/organizations/"+orgID+"/workspace-key-envelopes/"+deviceID, nil, &out)
	return out, err
}
func (c *Client) OrganizationDevices(ctx context.Context, orgID string) ([]OrganizationDevice, error) {
	var out struct {
		Devices []OrganizationDevice `json:"devices"`
	}
	_, err := c.request(ctx, http.MethodGet, "/v1/organizations/"+orgID+"/devices", nil, &out)
	return out.Devices, err
}
func (c *Client) PutWorkspaceEnvelope(ctx context.Context, orgID, deviceID string, envelope Envelope) error {
	_, err := c.request(ctx, http.MethodPut, "/v1/organizations/"+orgID+"/workspace-key-envelopes/"+deviceID, map[string]any{
		"generation": envelope.Generation, "ephemeral_public_key": envelope.EphemeralPublicKey,
		"nonce": envelope.Nonce, "ciphertext": envelope.Ciphertext,
	}, nil)
	return err
}
func (c *Client) Servers(ctx context.Context, orgID string) ([]EncryptedServer, error) {
	var out struct {
		Servers []EncryptedServer `json:"servers"`
	}
	_, err := c.request(ctx, http.MethodGet, "/v1/organizations/"+orgID+"/servers?include_deleted=true", nil, &out)
	return out.Servers, err
}
func (c *Client) CreateServer(ctx context.Context, orgID string, in EncryptedServer) (EncryptedServer, error) {
	var out EncryptedServer
	_, err := c.request(ctx, http.MethodPost, "/v1/organizations/"+orgID+"/servers", map[string]any{"id": in.ID, "ciphertext": in.Ciphertext, "nonce": in.Nonce, "team_ids": in.TeamIDs, "team_access_expires_at": in.TeamAccessExpiresAt}, &out)
	return out, err
}
func (c *Client) UpdateServer(ctx context.Context, orgID string, in EncryptedServer, expected int64) (EncryptedServer, error) {
	var out EncryptedServer
	_, err := c.request(ctx, http.MethodPut, "/v1/organizations/"+orgID+"/servers/"+in.ID, map[string]any{"ciphertext": in.Ciphertext, "nonce": in.Nonce, "expected_revision": expected, "team_ids": in.TeamIDs, "team_access_expires_at": in.TeamAccessExpiresAt}, &out)
	return out, err
}

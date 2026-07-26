// Package store owns the ssher data model and the JSON document that lives
// inside the encrypted vault. It is intentionally I/O-free: callers in the
// vault package shuttle bytes in and out, store handles structure.
package store

import "time"

// PortForward maps a local listening address to a remote target.
//
// JSON shape matches the Python:
//
//	{"local_port": 8080, "remote_host": "127.0.0.1", "remote_port": 80}
type PortForward struct {
	LocalPort  int    `json:"local_port"`
	RemoteHost string `json:"remote_host,omitempty"`
	RemotePort int    `json:"remote_port"`
}

type AuthType string

const (
	AuthKey      AuthType = "key"
	AuthPassword AuthType = "password"
)

// Server is the central record in the vault.
type Server struct {
	Name                string            `json:"name"`
	Host                string            `json:"host"`
	User                string            `json:"user"`
	Port                int               `json:"port"`
	AuthType            AuthType          `json:"auth_type"`
	Password            string            `json:"password,omitempty"`
	KeyPath             string            `json:"key_path,omitempty"`
	Group               string            `json:"group,omitempty"`
	Tags                []string          `json:"tags,omitempty"`
	Notes               string            `json:"notes,omitempty"`
	JumpHost            string            `json:"jump_host,omitempty"`
	LocalForwards       []PortForward     `json:"local_forwards,omitempty"`
	RemoteForwards      []PortForward     `json:"remote_forwards,omitempty"`
	X11Forward          bool              `json:"x11_forward,omitempty"`
	KeepAlive           int               `json:"keep_alive"`
	ConnectionTimeout   int               `json:"connection_timeout"`
	CustomOptions       map[string]string `json:"custom_options,omitempty"`
	CreatedAt           string            `json:"created_at,omitempty"`
	LastConnected       string            `json:"last_connected,omitempty"`
	ConnectionCount     int               `json:"connection_count,omitempty"`
	IsFavorite          bool              `json:"is_favorite,omitempty"`
	PasswordExpires     string            `json:"password_expires,omitempty"`
	Aliases             []string          `json:"aliases,omitempty"`
	Profile             string            `json:"profile,omitempty"`
	AutoReconnect       bool              `json:"auto_reconnect,omitempty"`
	MaxReconnectRetries int               `json:"max_reconnect_retries,omitempty"`
}

// HistoryEntry records one connection attempt.
type HistoryEntry struct {
	ServerName   string `json:"server_name"`
	Host         string `json:"host"`
	User         string `json:"user"`
	Timestamp    string `json:"timestamp"`
	Duration     int    `json:"duration,omitempty"`
	Success      bool   `json:"success"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// Profile is a reusable bundle of SSH options applied on top of a server.
type Profile struct {
	Name                string            `json:"name"`
	ConnectionTimeout   int               `json:"connection_timeout,omitempty"`
	KeepAlive           int               `json:"keep_alive,omitempty"`
	X11Forward          bool              `json:"x11_forward,omitempty"`
	CustomOptions       map[string]string `json:"custom_options,omitempty"`
	AutoReconnect       bool              `json:"auto_reconnect,omitempty"`
	MaxReconnectRetries int               `json:"max_reconnect_retries,omitempty"`
}

// Vault is the document round-tripped through encryption.
type Vault struct {
	Version  int               `json:"version"`
	Servers  []Server          `json:"servers"`
	Profiles []Profile         `json:"profiles,omitempty"`
	Aliases  map[string]string `json:"aliases,omitempty"` // alias -> canonical server name
	History  []HistoryEntry    `json:"history,omitempty"`
}

const CurrentVaultVersion = 1

// New constructs an empty vault at the current version.
func New() *Vault {
	return &Vault{
		Version: CurrentVaultVersion,
		Aliases: map[string]string{},
	}
}

// Touch initialises sensible defaults on a server before saving.
func (s *Server) Touch() {
	if s.CreatedAt == "" {
		s.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	if s.Port == 0 {
		s.Port = 22
	}
	if s.AuthType == "" {
		s.AuthType = AuthKey
	}
	if s.Group == "" {
		s.Group = "default"
	}
	if s.ConnectionTimeout == 0 {
		s.ConnectionTimeout = 30
	}
}

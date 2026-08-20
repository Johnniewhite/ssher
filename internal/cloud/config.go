package cloud

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/johnniewhite/ssher/internal/paths"
)

const (
	DefaultAPIURL = "https://api.getssher.com"
	DefaultAppURL = "https://cloud.getssher.com"
)

var ErrNotLoggedIn = errors.New("not signed in to SSHer Cloud; run 'ssher cloud login'")

type Config struct {
	APIURL           string `json:"api_url"`
	AppURL           string `json:"app_url"`
	SessionToken     string `json:"session_token"`
	ExpiresAt        string `json:"expires_at"`
	UserID           string `json:"user_id"`
	UserEmail        string `json:"user_email"`
	DeviceID         string `json:"device_id"`
	OrganizationID   string `json:"organization_id,omitempty"`
	OrganizationName string `json:"organization_name,omitempty"`
}

func LoadConfig() (*Config, error) {
	path, err := paths.CloudConfigFile()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrNotLoggedIn
	}
	if err != nil {
		return nil, fmt.Errorf("read cloud config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("decode cloud config: %w", err)
	}
	if cfg.SessionToken == "" || cfg.DeviceID == "" {
		return nil, ErrNotLoggedIn
	}
	if expiry, err := time.Parse(time.RFC3339, cfg.ExpiresAt); err == nil && time.Now().After(expiry) {
		return nil, errors.New("SSHer Cloud session expired; run 'ssher cloud login' again")
	}
	if cfg.APIURL == "" {
		cfg.APIURL = DefaultAPIURL
	}
	if cfg.AppURL == "" {
		cfg.AppURL = DefaultAppURL
	}
	return &cfg, nil
}

func SaveConfig(cfg *Config) error {
	dir, err := paths.EnsureConfigDir()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, "cloud.json")
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".cloud-*.json")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(paths.FileMode); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

func RemoveConfig() error {
	path, err := paths.CloudConfigFile()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

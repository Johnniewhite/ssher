// Package telemetry records a privacy-safe, one-time installation signal.
// It never sends commands, server data, usernames, hostnames, or vault data.
package telemetry

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/johnniewhite/ssher/internal/paths"
)

const DefaultEndpoint = "https://api.getssher.com/v1/telemetry/installations"

type state struct {
	InstallationID  string `json:"installation_id"`
	ReportedVersion string `json:"reported_version,omitempty"`
}

type event struct {
	InstallationID string `json:"installation_id"`
	Platform       string `json:"platform"`
	Architecture   string `json:"architecture"`
	Version        string `json:"version"`
	Source         string `json:"source"`
}

// MaybeReport sends at most one event for a released version. Failure is
// intentionally silent and never prevents an ssher command from running.
func MaybeReport(version string) {
	if disabled() || version == "" || strings.Contains(version, "dev") {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1200*time.Millisecond)
	defer cancel()
	_ = report(ctx, DefaultEndpoint, version, "cli", http.DefaultClient)
}

func disabled() bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv("SSHER_DISABLE_TELEMETRY")))
	return value == "1" || value == "true" || value == "yes"
}

func report(ctx context.Context, endpoint, version, source string, client *http.Client) error {
	path, err := paths.TelemetryFile()
	if err != nil {
		return err
	}
	current := state{}
	if body, readErr := os.ReadFile(path); readErr == nil {
		_ = json.Unmarshal(body, &current)
	}
	if current.ReportedVersion == version {
		return nil
	}
	if current.InstallationID == "" {
		current.InstallationID, err = randomUUID()
		if err != nil {
			return err
		}
	}
	payload, err := json.Marshal(event{
		InstallationID: current.InstallationID, Platform: runtime.GOOS,
		Architecture: runtime.GOARCH, Version: version, Source: source,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "ssher/"+version)
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return &httpError{status: res.Status}
	}
	current.ReportedVersion = version
	if _, err := paths.EnsureConfigDir(); err != nil {
		return err
	}
	body, err := json.MarshalIndent(current, "", "  ")
	if err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".telemetry-*.json")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(paths.FileMode); err != nil {
		temporary.Close()
		return err
	}
	if _, err := temporary.Write(body); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return paths.ReplaceFile(temporaryPath, path)
}

func randomUUID() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	encoded := hex.EncodeToString(value)
	return encoded[0:8] + "-" + encoded[8:12] + "-" + encoded[12:16] + "-" + encoded[16:20] + "-" + encoded[20:32], nil
}

type httpError struct{ status string }

func (e *httpError) Error() string { return "telemetry endpoint returned " + e.status }

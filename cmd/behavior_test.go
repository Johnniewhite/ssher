package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/johnniewhite/ssher/internal/store"
)

func TestRootAcceptsBareConnectFlags(t *testing.T) {
	if err := rootCmd.ParseFlags([]string{"--reconnect", "--max-retries", "7", "--record"}); err != nil {
		t.Fatalf("parse root connect flags: %v", err)
	}
	if !connectReconnect || connectMaxRetries != 7 || !connectRecord {
		t.Fatalf("flags not applied: reconnect=%v retries=%d record=%v",
			connectReconnect, connectMaxRetries, connectRecord)
	}
}

func TestConnectionPolicyUsesServerDefaultsAndExplicitOverrides(t *testing.T) {
	server := &store.Server{AutoReconnect: true, MaxReconnectRetries: 9}
	reconnect, retries, err := resolveConnectionPolicy(server, false, false, 0, false)
	if err != nil {
		t.Fatalf("server policy: %v", err)
	}
	if !reconnect || retries != 9 {
		t.Fatalf("server policy = (%v, %d), want (true, 9)", reconnect, retries)
	}

	reconnect, retries, err = resolveConnectionPolicy(server, false, true, 0, true)
	if err != nil {
		t.Fatalf("override policy: %v", err)
	}
	if reconnect || retries != 0 {
		t.Fatalf("override policy = (%v, %d), want (false, 0)", reconnect, retries)
	}

	if _, _, err := resolveConnectionPolicy(server, true, true, -1, true); err == nil {
		t.Fatal("negative retry override unexpectedly succeeded")
	}
}

func TestParseForwardsValidatesPorts(t *testing.T) {
	if _, err := parseForwards("0:localhost:80"); err == nil {
		t.Fatal("accepted port zero")
	}
	if _, err := parseForwards("8080:localhost:70000"); err == nil {
		t.Fatal("accepted out-of-range remote port")
	}
	got, err := parseForwards("8080:localhost:80")
	if err != nil {
		t.Fatalf("parse valid forward: %v", err)
	}
	if len(got) != 1 || got[0].RemoteHost != "localhost" {
		t.Fatalf("unexpected forward: %+v", got)
	}
}

func TestCanonicalAliasesDriveSearchAndConfigExport(t *testing.T) {
	server := store.Server{
		Name: "production", Host: "example.com", User: "deploy",
		Port: 22, AuthType: store.AuthKey,
	}
	aliases := map[string]string{"prod": "production"}
	filtered := filterServers([]store.Server{server}, aliases, "", false, "prod")
	if len(filtered) != 1 {
		t.Fatalf("alias search returned %d servers", len(filtered))
	}
	config := renderSSHConfig([]store.Server{server}, aliases)
	if !strings.Contains(config, "Host production prod\n    HostName example.com\n") {
		t.Fatalf("export omitted canonical alias:\n%s", config)
	}
}

func TestOpenPrivateOutputTightensExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "servers.json")
	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatalf("fixture: %v", err)
	}
	f, err := openPrivateOutput(path)
	if err != nil {
		t.Fatalf("openPrivateOutput: %v", err)
	}
	if _, err := f.WriteString("new"); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	// Windows reports synthetic Unix mode bits; confidentiality comes from
	// the inherited user-profile ACL rather than chmod.
	if got := info.Mode().Perm(); runtime.GOOS != "windows" && got != 0o600 {
		t.Fatalf("mode = %o, want 600", got)
	}
}

func TestCSVCellHandlesShortRows(t *testing.T) {
	row := []string{"name-only"}
	if _, ok := csvCell(row, 3); ok {
		t.Fatal("out-of-range CSV cell unexpectedly exists")
	}
	if got, ok := csvCell(row, 0); !ok || got != "name-only" {
		t.Fatalf("valid cell = %q, %v", got, ok)
	}
}

func TestPrintExecResultsCountsRemoteFailures(t *testing.T) {
	results := []execResult{
		{Server: store.Server{Name: "ok"}, Exit: 0},
		{Server: store.Server{Name: "bad-exit"}, Exit: 2},
		{Server: store.Server{Name: "dial-error"}, Err: os.ErrDeadlineExceeded},
	}
	if got := printExecResults(results); got != 2 {
		t.Fatalf("failed count = %d, want 2", got)
	}
}

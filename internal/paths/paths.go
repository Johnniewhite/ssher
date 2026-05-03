// Package paths resolves all on-disk locations used by ssher and exposes
// the canonical permission modes for files and directories.
package paths

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	DirMode  os.FileMode = 0o700
	FileMode os.FileMode = 0o600
)

func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".ssher"), nil
}

func EnsureConfigDir() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	for _, sub := range []string{"", "backups", "recordings"} {
		p := filepath.Join(dir, sub)
		if err := os.MkdirAll(p, DirMode); err != nil {
			return "", fmt.Errorf("create %s: %w", p, err)
		}
		if err := os.Chmod(p, DirMode); err != nil {
			return "", fmt.Errorf("chmod %s: %w", p, err)
		}
	}
	return dir, nil
}

func VaultFile() (string, error)      { return joinConfig("vault.bin") }
func SessionFile() (string, error)    { return joinConfig(".session") }
func BackupsDir() (string, error)     { return joinConfig("backups") }
func RecordingsDir() (string, error)  { return joinConfig("recordings") }
func LegacyServers() (string, error)  { return joinConfig("servers.enc") }
func LegacySalt() (string, error)     { return joinConfig(".salt") }
func LegacyKey() (string, error)      { return joinConfig(".key") }
func LegacyHistory() (string, error)  { return joinConfig("history.json") }
func LegacyProfiles() (string, error) { return joinConfig("profiles.json") }
func LegacyAliases() (string, error)  { return joinConfig("aliases.json") }

func joinConfig(name string) (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name), nil
}

// HostFingerprint returns a stable, machine-bound 32-byte identifier used to
// derive a wrapping key for the session file. Defense-in-depth: an attacker
// who exfiltrates ~/.ssher/.session alone still needs the local machine-id
// (or its equivalent) to unwrap the cached vault key.
func HostFingerprint() ([]byte, error) {
	raw, err := readMachineID()
	if err != nil {
		// Fallback so non-Linux/non-macOS boxes still work, just weaker.
		host, _ := os.Hostname()
		raw = host + ":" + fmt.Sprint(os.Getuid())
	}
	sum := sha256.Sum256([]byte("ssher.host.fingerprint.v1|" + raw))
	return sum[:], nil
}

func readMachineID() (string, error) {
	switch runtime.GOOS {
	case "linux":
		for _, p := range []string{"/etc/machine-id", "/var/lib/dbus/machine-id"} {
			b, err := os.ReadFile(p)
			if err == nil {
				return strings.TrimSpace(string(b)), nil
			}
		}
		return "", fmt.Errorf("no machine-id file")
	case "darwin":
		out, err := exec.Command("/usr/sbin/ioreg", "-rd1", "-c", "IOPlatformExpertDevice").Output()
		if err != nil {
			return "", err
		}
		const marker = "IOPlatformUUID"
		idx := strings.Index(string(out), marker)
		if idx < 0 {
			return "", fmt.Errorf("IOPlatformUUID not found")
		}
		// Format: ... "IOPlatformUUID" = "ABCD-1234-..."
		seg := string(out)[idx+len(marker):]
		q1 := strings.Index(seg, `"`)
		if q1 < 0 {
			return "", fmt.Errorf("malformed ioreg output")
		}
		seg = seg[q1+1:]
		q2 := strings.Index(seg, `"`)
		if q2 < 0 {
			return "", fmt.Errorf("malformed ioreg output")
		}
		return seg[:q2], nil
	default:
		return "", fmt.Errorf("unsupported os: %s", runtime.GOOS)
	}
}

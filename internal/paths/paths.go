// Package paths resolves all on-disk locations used by ssher and exposes
// the canonical permission modes for files and directories.
package paths

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
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
	for _, sub := range []string{"", "backups", "recordings", "cloud-keys"} {
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

func VaultFile() (string, error)       { return joinConfig("vault.bin") }
func SessionFile() (string, error)     { return joinConfig(".session") }
func BackupsDir() (string, error)      { return joinConfig("backups") }
func RecordingsDir() (string, error)   { return joinConfig("recordings") }
func LegacyServers() (string, error)   { return joinConfig("servers.enc") }
func LegacySalt() (string, error)      { return joinConfig(".salt") }
func LegacyKey() (string, error)       { return joinConfig(".key") }
func LegacyHistory() (string, error)   { return joinConfig("history.json") }
func LegacyProfiles() (string, error)  { return joinConfig("profiles.json") }
func LegacyAliases() (string, error)   { return joinConfig("aliases.json") }
func CloudConfigFile() (string, error) { return joinConfig("cloud.json") }
func CloudKeyFile() (string, error)    { return joinConfig("cloud-device-key.pem") }
func CloudKeysDir() (string, error)    { return joinConfig("cloud-keys") }

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
		// Fallback for unusual platforms or locked-down hosts. It is weaker
		// than the platform machine ID, but still keeps a copied session file
		// from being directly reusable on a differently named machine.
		host, _ := os.Hostname()
		raw = host + ":" + fmt.Sprint(os.Getuid())
	}
	sum := sha256.Sum256([]byte("ssher.host.fingerprint.v1|" + raw))
	return sum[:], nil
}

// ReplaceFile atomically moves source over target. os.Rename has replace
// semantics on Unix but not consistently on Windows, so the platform-specific
// implementation uses MoveFileEx there.
func ReplaceFile(source, target string) error { return replaceFile(source, target) }

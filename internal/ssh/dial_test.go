package ssh

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	gossh "golang.org/x/crypto/ssh"

	"github.com/johnniewhite/ssher/internal/store"
)

func TestAuthMethodsFindsDefaultEd25519KeyWithoutAgent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("SSH_AUTH_SOCK", "")
	if err := os.MkdirAll(filepath.Join(home, ".ssh"), 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	block, err := gossh.MarshalPrivateKey(privateKey, "test")
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}
	keyPath := filepath.Join(home, ".ssh", "id_ed25519")
	if err := os.WriteFile(keyPath, pem.EncodeToMemory(block), 0o600); err != nil {
		t.Fatalf("write key: %v", err)
	}

	methods, closers, err := authMethods(&store.Server{AuthType: store.AuthKey})
	if err != nil {
		t.Fatalf("authMethods: %v", err)
	}
	if len(methods) != 1 {
		t.Fatalf("methods = %d, want 1", len(methods))
	}
	if len(closers) != 0 {
		t.Fatalf("unexpected agent closers: %d", len(closers))
	}
}

func TestAuthMethodsReportsMissingExplicitKey(t *testing.T) {
	t.Setenv("SSH_AUTH_SOCK", "")
	_, _, err := authMethods(&store.Server{
		AuthType: store.AuthKey,
		KeyPath:  filepath.Join(t.TempDir(), "missing"),
	})
	if err == nil {
		t.Fatal("missing explicit key unexpectedly succeeded")
	}
}

func TestRemoteForwardUsesRemotePortAsListenerAndHostAsLocalTarget(t *testing.T) {
	forward := store.PortForward{
		LocalPort:  9000,
		RemoteHost: "database.internal",
		RemotePort: 5432,
	}
	if got := remoteListenAddress(forward); got != "127.0.0.1:9000" {
		t.Fatalf("listen address = %q", got)
	}
	if got := remoteTargetAddress(forward); got != "database.internal:5432" {
		t.Fatalf("target address = %q", got)
	}
}

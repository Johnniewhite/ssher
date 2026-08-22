package openssh

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestImportDiscoversIncludesAndUsesResolvedOpenSSHValues(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	sshDir := filepath.Join(home, ".ssh")
	if err := os.MkdirAll(filepath.Join(sshDir, "conf.d"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sshDir, "config"), []byte("Include conf.d/*.conf\nHost prod *.internal !blocked\n  HostName ignored\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sshDir, "conf.d", "team.conf"), []byte("Host staging quoted-host # comment\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	keyPath := filepath.Join(sshDir, "id_prod")
	if err := os.WriteFile(keyPath, []byte("key"), 0o600); err != nil {
		t.Fatal(err)
	}
	resolver := func(_ context.Context, _ string, alias string) ([]byte, error) {
		return []byte(fmt.Sprintf("hostname %s.example.com\nuser deploy\nport 2222\nidentityfile %s\nproxyjump bastion\nserveraliveinterval 30\n", alias, keyPath)), nil
	}
	result, err := Import(context.Background(), filepath.Join(sshDir, "config"), resolver)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Servers) != 3 {
		t.Fatalf("servers = %#v", result.Servers)
	}
	if result.Servers[0].Name != "staging" || result.Servers[2].Name != "prod" {
		t.Fatalf("order/names = %#v", result.Servers)
	}
	if result.Servers[0].Port != 2222 || result.Servers[0].JumpHost != "bastion" || result.Servers[0].KeyPath != keyPath {
		t.Fatalf("resolved = %#v", result.Servers[0])
	}
}

func TestParseResolvedRequiresDestination(t *testing.T) {
	_, err := ParseResolved("broken", strings.NewReader("port 22\n"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseResolvedFallsBackForInvalidPort(t *testing.T) {
	server, err := ParseResolved("prod", strings.NewReader("hostname example.com\nuser deploy\nport 0\n"))
	if err != nil {
		t.Fatal(err)
	}
	if server.Port != 22 {
		t.Fatalf("port = %d, want 22", server.Port)
	}
}

func TestSplitFieldsHonorsQuotesAndComments(t *testing.T) {
	fields, err := splitFields(`Host "production web" prod # ignored`)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(fields, "|") != "Host|production web|prod" {
		t.Fatalf("fields = %#v", fields)
	}
}

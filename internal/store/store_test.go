package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/johnniewhite/ssher/internal/vault"
)

func TestSavedPreservesActiveKDFParameters(t *testing.T) {
	password := []byte("test password")
	params := vault.Argon2idParams{Time: 1, MemKiB: 1024, Threads: 1}
	configureTestVault(t, password, params, params, vault.Argon2idParams{Time: 1, MemKiB: 2048, Threads: 1})

	saved, err := LoadWithPassword(password)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if saved.Params != params {
		t.Fatalf("loaded params = %+v, want %+v", saved.Params, params)
	}
	if err := saved.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}

	_, header, _, err := vault.LoadFile(password)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if header.Params != params {
		t.Fatalf("saved header params = %+v, want %+v", header.Params, params)
	}
}

func TestPasswordUnlockRatchetsWeakKDF(t *testing.T) {
	password := []byte("test password")
	low := vault.Argon2idParams{Time: 1, MemKiB: 1024, Threads: 1}
	high := vault.Argon2idParams{Time: 1, MemKiB: 2048, Threads: 1}
	configureTestVault(t, password, low, high, high)

	saved, err := LoadWithPassword(password)
	if err != nil {
		t.Fatalf("load and ratchet: %v", err)
	}
	if saved.Params != high {
		t.Fatalf("ratcheted params = %+v, want %+v", saved.Params, high)
	}
	if err := saved.SaveAndRefreshSession(); err != nil {
		t.Fatalf("save ratcheted vault: %v", err)
	}
	if err := vault.ClearSession(); err != nil {
		t.Fatalf("clear session: %v", err)
	}
	if _, err := LoadWithPassword(password); err != nil {
		t.Fatalf("password unlock after ratchet: %v", err)
	}
}

func TestUnmarshalMigratesServerAliases(t *testing.T) {
	v, err := unmarshal([]byte(`{
		"version": 1,
		"servers": [{
			"name": "production",
			"host": "example.com",
			"user": "deploy",
			"port": 22,
			"auth_type": "key",
			"aliases": ["prod"]
		}]
	}`))
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := v.Aliases["prod"]; got != "production" {
		t.Fatalf("alias target = %q, want production", got)
	}
	if len(v.Servers[0].Aliases) != 0 {
		t.Fatalf("legacy server aliases were not cleared")
	}
	target, _, err := v.ResolveTarget("pro")
	if err != nil {
		t.Fatalf("fuzzy alias resolve: %v", err)
	}
	if target.Name != "production" {
		t.Fatalf("resolved %q, want production", target.Name)
	}
}

func TestAddServerValidatesRequiredConnectionFields(t *testing.T) {
	v := New()
	tests := []Server{
		{Name: "missing-host", User: "root", Port: 22, AuthType: AuthKey},
		{Name: "missing-user", Host: "example.com", Port: 22, AuthType: AuthKey},
		{Name: "bad-port", Host: "example.com", User: "root", Port: 70000, AuthType: AuthKey},
		{Name: "bad-auth", Host: "example.com", User: "root", Port: 22, AuthType: "other"},
	}
	for _, server := range tests {
		if err := v.AddServer(server); err == nil {
			t.Fatalf("AddServer(%s) unexpectedly succeeded", server.Name)
		}
	}
}

func TestSavedRefusesToGuessMissingKDFParameters(t *testing.T) {
	saved := &Saved{Vault: New(), Key: make([]byte, 32)}
	if err := saved.Save(); err == nil {
		t.Fatal("save without KDF parameters unexpectedly succeeded")
	}
}

func TestRekeyPreservesCompleteVault(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	oldPassword := []byte("old test password")
	newPassword := []byte("new test password")
	saved, err := Initialise(&Vault{
		Version: CurrentVaultVersion,
		Servers: []Server{{
			Name: "production", Host: "example.com", User: "deploy",
			Port: 22, AuthType: AuthPassword, Password: "server-secret",
		}},
		Aliases: map[string]string{"prod": "production"},
	}, oldPassword)
	if err != nil {
		t.Fatalf("initialise: %v", err)
	}
	rekeyed, err := saved.Rekey(newPassword)
	if err != nil {
		t.Fatalf("rekey: %v", err)
	}
	if len(rekeyed.Vault.Servers) != 1 || rekeyed.Vault.Servers[0].Password != "server-secret" {
		t.Fatalf("rekey lost server data: %+v", rekeyed.Vault.Servers)
	}
	if rekeyed.Vault.Aliases["prod"] != "production" {
		t.Fatalf("rekey lost aliases: %+v", rekeyed.Vault.Aliases)
	}
	if _, err := LoadWithPassword(oldPassword); err == nil {
		t.Fatal("old password still unlocks rekeyed vault")
	}
	if _, err := LoadWithPassword(newPassword); err != nil {
		t.Fatalf("new password does not unlock rekeyed vault: %v", err)
	}
}

func configureTestVault(t *testing.T, password []byte, fileParams, minParams, defaultParams vault.Argon2idParams) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	oldMin, oldDefault := vault.MinParams, vault.DefaultParams
	vault.MinParams, vault.DefaultParams = minParams, defaultParams
	t.Cleanup(func() {
		vault.MinParams, vault.DefaultParams = oldMin, oldDefault
	})

	blob, err := vault.Encrypt([]byte(`{"version":1,"servers":[]}`), password, fileParams)
	if err != nil {
		t.Fatalf("encrypt fixture: %v", err)
	}
	dir := filepath.Join(os.Getenv("HOME"), ".ssher")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("mkdir fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "vault.bin"), blob, 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
}

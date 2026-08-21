package cloud

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"testing"

	"golang.org/x/crypto/hkdf"

	"github.com/johnniewhite/ssher/internal/store"
)

func TestDeviceKeyEnvelopeAndServerRoundTrip(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	deviceKey, err := LoadOrCreateDeviceKey()
	if err != nil {
		t.Fatalf("create device key: %v", err)
	}
	reloaded, err := LoadOrCreateDeviceKey()
	if err != nil {
		t.Fatalf("reload device key: %v", err)
	}
	if string(deviceKey.Bytes()) != string(reloaded.Bytes()) {
		t.Fatal("device key did not persist")
	}

	const orgID = "d82a0ab7-18e7-493f-b621-110d7da4897c"
	const deviceID = "fba770d3-da08-4a35-af10-a3e57bc48dd9"
	ephemeral, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	secret, err := ephemeral.ECDH(deviceKey.PublicKey())
	if err != nil {
		t.Fatal(err)
	}
	reader := hkdf.New(sha256.New, secret, []byte(orgID), []byte("ssher-cloud-workspace-key-v1"))
	wrappingKey := make([]byte, 32)
	if _, err := io.ReadFull(reader, wrappingKey); err != nil {
		t.Fatal(err)
	}
	block, _ := aes.NewCipher(wrappingKey)
	gcm, _ := cipher.NewGCM(block)
	workspaceKey := make([]byte, 32)
	if _, err := rand.Read(workspaceKey); err != nil {
		t.Fatal(err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		t.Fatal(err)
	}
	envelope := Envelope{OrganizationID: orgID, DeviceID: deviceID, EphemeralPublicKey: ephemeral.PublicKey().Bytes(), Nonce: nonce}
	envelope.Ciphertext = gcm.Seal(nil, nonce, workspaceKey, []byte(fmt.Sprintf("ssher-cloud-envelope:%s:%s", orgID, deviceID)))
	unwrapped, err := UnwrapWorkspaceKey(envelope, deviceKey)
	if err != nil {
		t.Fatalf("unwrap: %v", err)
	}
	if string(unwrapped) != string(workspaceKey) {
		t.Fatal("workspace key mismatch")
	}

	payload := ServerPayload{Server: store.Server{Name: "prod", Host: "10.0.0.4", User: "deploy", Port: 22, AuthType: store.AuthPassword, Password: "secret"}}
	record, err := EncryptServer(payload, workspaceKey, orgID, "4e87e01d-75e3-467d-a03f-20d44e215343", 7)
	if err != nil {
		t.Fatalf("encrypt server: %v", err)
	}
	decoded, err := DecryptServer(record, workspaceKey)
	if err != nil {
		t.Fatalf("decrypt server: %v", err)
	}
	if decoded.Name != payload.Name || decoded.Password != payload.Password {
		t.Fatalf("payload mismatch: %+v", decoded)
	}
}

func TestServerHashIgnoresCloudSyncMetadata(t *testing.T) {
	server := store.Server{Name: "prod", Host: "example.com", User: "root", Port: 22, AuthType: store.AuthKey}
	want := ServerHash(server)
	server.CloudID = "1"
	server.CloudRevision = 99
	server.CloudSyncedHash = "old"
	server.CloudTeamIDs = []string{"team"}
	if got := ServerHash(server); got != want {
		t.Fatalf("metadata changed hash: %s != %s", got, want)
	}
	server.Host = "new.example.com"
	if got := ServerHash(server); got == want {
		t.Fatal("server change did not change hash")
	}
}

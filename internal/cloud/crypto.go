package cloud

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/hkdf"

	"github.com/johnniewhite/ssher/internal/paths"
	"github.com/johnniewhite/ssher/internal/store"
)

type ServerPayload struct {
	store.Server
	PrivateKey string `json:"private_key,omitempty"`
}

func LoadOrCreateDeviceKey() (*ecdh.PrivateKey, error) {
	path, err := paths.CloudKeyFile()
	if err != nil {
		return nil, err
	}
	if b, err := os.ReadFile(path); err == nil {
		block, _ := pem.Decode(b)
		if block == nil {
			return nil, fmt.Errorf("decode cloud device key")
		}
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		switch privateKey := key.(type) {
		case *ecdh.PrivateKey:
			return privateKey, nil
		case *ecdsa.PrivateKey:
			return privateKey.ECDH()
		default:
			return nil, fmt.Errorf("cloud device key has unexpected type")
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	if _, err := paths.EnsureConfigDir(); err != nil {
		return nil, err
	}
	key, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}), paths.FileMode); err != nil {
		return nil, err
	}
	return key, nil
}

func workspaceWrappingKey(privateKey *ecdh.PrivateKey, publicBytes []byte, orgID string) ([]byte, error) {
	publicKey, err := ecdh.P256().NewPublicKey(publicBytes)
	if err != nil {
		return nil, err
	}
	secret, err := privateKey.ECDH(publicKey)
	if err != nil {
		return nil, err
	}
	reader := hkdf.New(sha256.New, secret, []byte(orgID), []byte("ssher-cloud-workspace-key-v1"))
	key := make([]byte, 32)
	if _, err := io.ReadFull(reader, key); err != nil {
		return nil, err
	}
	return key, nil
}

func UnwrapWorkspaceKey(envelope Envelope, privateKey *ecdh.PrivateKey) ([]byte, error) {
	key, err := workspaceWrappingKey(privateKey, envelope.EphemeralPublicKey, envelope.OrganizationID)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	aad := []byte(fmt.Sprintf("ssher-cloud-envelope:%s:%s", envelope.OrganizationID, envelope.DeviceID))
	return gcm.Open(nil, envelope.Nonce, envelope.Ciphertext, aad)
}

// WrapWorkspaceKey encrypts a workspace key for one authorized device. The
// server stores only this envelope and never receives the plaintext key.
func WrapWorkspaceKey(workspaceKey, recipientPublicKey []byte, orgID, deviceID string) (Envelope, error) {
	ephemeral, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		return Envelope{}, err
	}
	key, err := workspaceWrappingKey(ephemeral, recipientPublicKey, orgID)
	if err != nil {
		return Envelope{}, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return Envelope{}, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return Envelope{}, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return Envelope{}, err
	}
	aad := []byte(fmt.Sprintf("ssher-cloud-envelope:%s:%s", orgID, deviceID))
	return Envelope{
		OrganizationID:     orgID,
		DeviceID:           deviceID,
		Generation:         1,
		EphemeralPublicKey: ephemeral.PublicKey().Bytes(),
		Nonce:              nonce,
		Ciphertext:         gcm.Seal(nil, nonce, workspaceKey, aad),
	}, nil
}

func EncryptServer(payload ServerPayload, key []byte, orgID, serverID string, revision int64) (EncryptedServer, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return EncryptedServer{}, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return EncryptedServer{}, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return EncryptedServer{}, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return EncryptedServer{}, err
	}
	aad := []byte(fmt.Sprintf("ssher-cloud-server:%s:%s:v%d", orgID, serverID, revision))
	return EncryptedServer{ID: serverID, OrganizationID: orgID, Nonce: nonce, Ciphertext: gcm.Seal(nil, nonce, b, aad), Revision: revision}, nil
}

func DecryptServer(record EncryptedServer, key []byte) (ServerPayload, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return ServerPayload{}, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return ServerPayload{}, err
	}
	aad := []byte(fmt.Sprintf("ssher-cloud-server:%s:%s:v%d", record.OrganizationID, record.ID, record.Revision))
	b, err := gcm.Open(nil, record.Nonce, record.Ciphertext, aad)
	if err != nil {
		return ServerPayload{}, err
	}
	var payload ServerPayload
	if err := json.Unmarshal(b, &payload); err != nil {
		return ServerPayload{}, err
	}
	return payload, nil
}

func PayloadForServer(server store.Server) ServerPayload {
	server.CloudID = ""
	server.CloudRevision = 0
	server.CloudSyncedHash = ""
	server.CloudSyncedAt = ""
	server.CloudTeamIDs = nil
	server.CloudTeamAccessExpiresAt = nil
	return ServerPayload{Server: server}
}

func ServerHash(server store.Server) string {
	b, _ := json.Marshal(PayloadForServer(server))
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

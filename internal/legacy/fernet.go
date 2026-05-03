// Package legacy reads the Python ssher v3.x vault format on a one-shot
// basis. We do NOT write Fernet — the only goal here is to migrate forward
// and then forget the format.
//
// Python uses cryptography.Fernet, which is:
//
//	signing key   = key_bytes[:16]   (HMAC-SHA256)
//	encryption key = key_bytes[16:]   (AES-128-CBC)
//
// where key_bytes is 32 bytes derived via PBKDF2-SHA256(password, salt,
// 480_000 iterations). The token is base64url:
//
//	0x80 | timestamp(8B BE) | IV(16B) | ciphertext(variable, PKCS7) | HMAC(32B)
//
// We re-implement decryption here in ~80 lines so we don't need a Fernet
// dependency that we'd carry forever for this one-shot.
package legacy

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"

	"golang.org/x/crypto/pbkdf2"
)

const (
	pythonIterations = 480_000
	fernetKeyLen     = 32
)

var (
	ErrTokenTooShort = errors.New("legacy: fernet token too short")
	ErrBadVersion    = errors.New("legacy: unsupported fernet version")
	ErrBadHMAC       = errors.New("legacy: HMAC verification failed (wrong password)")
	ErrBadPadding    = errors.New("legacy: PKCS7 padding invalid")
)

// DeriveKey runs the same PBKDF2 the Python tool used.
func DeriveKey(password, salt []byte) []byte {
	return pbkdf2.Key(password, salt, pythonIterations, fernetKeyLen, sha256.New)
}

// Decrypt opens a Fernet token using the given 32-byte key.
//
// We don't enforce the optional time-to-live; ssher tokens are not
// time-limited by the writer.
func Decrypt(token, key []byte) ([]byte, error) {
	if len(key) != fernetKeyLen {
		return nil, fmt.Errorf("legacy: bad key length %d", len(key))
	}
	signingKey := key[:16]
	encKey := key[16:]

	// Strip whitespace; the Python writer doesn't add any but conscientious
	// tools sometimes do.
	token = bytes.TrimSpace(token)

	raw, err := base64.URLEncoding.DecodeString(string(token))
	if err != nil {
		return nil, fmt.Errorf("legacy: base64: %w", err)
	}
	// 1 (ver) + 8 (ts) + 16 (iv) + at least 16 (one block) + 32 (hmac) = 73
	if len(raw) < 73 {
		return nil, ErrTokenTooShort
	}
	if raw[0] != 0x80 {
		return nil, ErrBadVersion
	}

	mac := raw[len(raw)-32:]
	body := raw[:len(raw)-32]

	h := hmac.New(sha256.New, signingKey)
	h.Write(body)
	if !hmac.Equal(h.Sum(nil), mac) {
		return nil, ErrBadHMAC
	}

	iv := body[9:25]
	ct := body[25:]
	if len(ct)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("legacy: ciphertext length %d not a multiple of %d", len(ct), aes.BlockSize)
	}

	block, err := aes.NewCipher(encKey)
	if err != nil {
		return nil, fmt.Errorf("legacy: aes: %w", err)
	}
	pt := make([]byte, len(ct))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(pt, ct)

	return stripPKCS7(pt)
}

func stripPKCS7(b []byte) ([]byte, error) {
	if len(b) == 0 {
		return nil, ErrBadPadding
	}
	pad := int(b[len(b)-1])
	if pad < 1 || pad > aes.BlockSize || pad > len(b) {
		return nil, ErrBadPadding
	}
	for i := len(b) - pad; i < len(b); i++ {
		if int(b[i]) != pad {
			return nil, ErrBadPadding
		}
	}
	return b[:len(b)-pad], nil
}

// FernetKey returns the base64url Fernet key that the Python `cryptography`
// library would have used. Exposed for tests / interop debugging — most
// callers should just use DeriveKey + Decrypt directly.
func FernetKey(password, salt []byte) string {
	return base64.URLEncoding.EncodeToString(DeriveKey(password, salt))
}

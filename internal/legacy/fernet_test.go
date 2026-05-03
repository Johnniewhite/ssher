package legacy

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"testing"
	"time"
)

// fernetEncrypt is a *test-only* writer for round-tripping our reader. The
// production code never writes Fernet — this exists so we can verify
// Decrypt against tokens we synthesise here in the same format the Python
// `cryptography` library produces.
func fernetEncrypt(t *testing.T, plaintext, key []byte) []byte {
	t.Helper()
	if len(key) != fernetKeyLen {
		t.Fatalf("bad key length %d", len(key))
	}
	signingKey, encKey := key[:16], key[16:]

	iv := make([]byte, 16)
	if _, err := rand.Read(iv); err != nil {
		t.Fatalf("rand: %v", err)
	}

	pad := aes.BlockSize - len(plaintext)%aes.BlockSize
	padded := append(append([]byte(nil), plaintext...), bytes.Repeat([]byte{byte(pad)}, pad)...)

	block, err := aes.NewCipher(encKey)
	if err != nil {
		t.Fatalf("aes: %v", err)
	}
	ct := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(ct, padded)

	body := make([]byte, 0, 1+8+16+len(ct))
	body = append(body, 0x80)
	var ts [8]byte
	binary.BigEndian.PutUint64(ts[:], uint64(time.Now().Unix()))
	body = append(body, ts[:]...)
	body = append(body, iv...)
	body = append(body, ct...)

	h := hmac.New(sha256.New, signingKey)
	h.Write(body)
	body = append(body, h.Sum(nil)...)

	enc := make([]byte, base64.URLEncoding.EncodedLen(len(body)))
	base64.URLEncoding.Encode(enc, body)
	return enc
}

func TestFernetRoundTrip(t *testing.T) {
	salt := []byte("0123456789abcdef")
	key := DeriveKey([]byte("hunter2"), salt)
	pt := []byte(`{"servers":[{"name":"prod","host":"203.0.113.4"}]}`)

	token := fernetEncrypt(t, pt, key)
	got, err := Decrypt(token, key)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if !bytes.Equal(got, pt) {
		t.Fatalf("plaintext mismatch:\n  got  %q\n  want %q", got, pt)
	}
}

func TestFernetWrongPassword(t *testing.T) {
	salt := []byte("0123456789abcdef")
	right := DeriveKey([]byte("right"), salt)
	wrong := DeriveKey([]byte("wrong"), salt)

	token := fernetEncrypt(t, []byte("payload"), right)
	if _, err := Decrypt(token, wrong); err != ErrBadHMAC {
		t.Fatalf("got %v, want ErrBadHMAC", err)
	}
}

func TestFernetTampered(t *testing.T) {
	salt := []byte("0123456789abcdef")
	key := DeriveKey([]byte("pw"), salt)
	token := fernetEncrypt(t, []byte("payload"), key)

	// Decode, flip a byte in the ciphertext region, re-encode.
	raw, err := base64.URLEncoding.DecodeString(string(token))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	raw[26] ^= 0xff
	enc := make([]byte, base64.URLEncoding.EncodedLen(len(raw)))
	base64.URLEncoding.Encode(enc, raw)

	if _, err := Decrypt(enc, key); err != ErrBadHMAC {
		t.Fatalf("got %v, want ErrBadHMAC after tamper", err)
	}
}

func TestFernetTooShort(t *testing.T) {
	if _, err := Decrypt([]byte("YQ=="), make([]byte, fernetKeyLen)); err != ErrTokenTooShort {
		t.Fatalf("got %v, want ErrTokenTooShort", err)
	}
}

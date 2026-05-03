package vault

import (
	"bytes"
	"testing"
)

// Cheap params so the test suite doesn't pay the production cost on every
// `go test`. We're testing format invariants, not Argon2id itself.
var testParams = Argon2idParams{Time: 1, MemKiB: 8 * 1024, Threads: 1}

func TestEncryptRoundTrip(t *testing.T) {
	pt := []byte(`{"servers":[{"name":"prod","host":"203.0.113.4"}]}`)
	pw := []byte("correct horse battery staple")

	blob, err := Encrypt(pt, pw, testParams)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	got, h, err := Decrypt(blob, pw)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if !bytes.Equal(got, pt) {
		t.Fatalf("plaintext mismatch:\n  got  %q\n  want %q", got, pt)
	}
	if h.Version != formatV1 {
		t.Fatalf("version: got %d want %d", h.Version, formatV1)
	}
	if h.KDFID != kdfArgon2id {
		t.Fatalf("kdf id: got %d want %d", h.KDFID, kdfArgon2id)
	}
	if h.Params != testParams {
		t.Fatalf("params: got %+v want %+v", h.Params, testParams)
	}
}

func TestDecryptWrongPassword(t *testing.T) {
	blob, err := Encrypt([]byte("hi"), []byte("right"), testParams)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if _, _, err := Decrypt(blob, []byte("wrong")); err != ErrBadPassword {
		t.Fatalf("got %v, want ErrBadPassword", err)
	}
}

func TestDecryptCorrupted(t *testing.T) {
	blob, err := Encrypt([]byte("hi"), []byte("pw"), testParams)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	// Flip a ciphertext byte (past the header) to trigger GCM auth failure.
	blob[len(blob)-1] ^= 0xff
	if _, _, err := Decrypt(blob, []byte("pw")); err != ErrBadPassword {
		t.Fatalf("got %v, want ErrBadPassword on tamper", err)
	}
}

func TestDecryptBadMagic(t *testing.T) {
	bogus := make([]byte, headerLen+32)
	if _, _, err := Decrypt(bogus, []byte("pw")); err != ErrBadMagic {
		t.Fatalf("got %v, want ErrBadMagic", err)
	}
}

func TestDecryptTruncated(t *testing.T) {
	if _, _, err := Decrypt([]byte("SSHV"), []byte("pw")); err != ErrTruncated {
		t.Fatalf("got %v, want ErrTruncated", err)
	}
}

func TestEncryptWithKeyMatches(t *testing.T) {
	// Decrypt via password, then re-encrypt via the derived key with the
	// SAME salt, then decrypt again -- should round-trip identically.
	pt := []byte("payload")
	pw := []byte("pw")
	blob, err := Encrypt(pt, pw, testParams)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	out, h, err := Decrypt(blob, pw)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if !bytes.Equal(out, pt) {
		t.Fatalf("first decrypt mismatch")
	}
	key := DeriveKey(pw, h.Salt[:], h.Params)
	rebuilt, err := EncryptWithKey([]byte("new payload"), key, h.Salt, h.Params)
	if err != nil {
		t.Fatalf("encrypt with key: %v", err)
	}
	got, _, err := Decrypt(rebuilt, pw)
	if err != nil {
		t.Fatalf("decrypt rebuilt: %v", err)
	}
	if string(got) != "new payload" {
		t.Fatalf("got %q want %q", got, "new payload")
	}
}

func TestParamsRatchet(t *testing.T) {
	low := Argon2idParams{Time: 1, MemKiB: 1024, Threads: 1}
	h := Header{Params: low}
	if !h.BelowMinParams() {
		t.Fatalf("low params should be below floor")
	}
	h2 := Header{Params: DefaultParams}
	if h2.BelowMinParams() {
		t.Fatalf("default params should not be below floor")
	}
}

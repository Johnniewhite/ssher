package vault

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/johnniewhite/ssher/internal/paths"
)

// Exists reports whether a vault file is present on disk.
func Exists() (bool, error) {
	p, err := paths.VaultFile()
	if err != nil {
		return false, err
	}
	_, err = os.Stat(p)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return err == nil, err
}

// LoadFile reads and decrypts the vault. Returns the plaintext payload,
// the header (so callers can ratchet KDF params), and the derived key (so
// callers can re-encrypt without paying Argon2id again on the same call).
func LoadFile(password []byte) ([]byte, Header, []byte, error) {
	p, err := paths.VaultFile()
	if err != nil {
		return nil, Header{}, nil, err
	}
	blob, err := os.ReadFile(p)
	if err != nil {
		return nil, Header{}, nil, fmt.Errorf("read vault: %w", err)
	}
	pt, h, err := Decrypt(blob, password)
	if err != nil {
		return nil, h, nil, err
	}
	key := DeriveKey(password, h.Salt[:], h.Params)
	return pt, h, key, nil
}

// OpenBlobWithKey reads the vault file and decrypts using a pre-derived key,
// skipping the Argon2id work. Used when a session is active and we already
// have the derived key cached.
func OpenBlobWithKey(key []byte) ([]byte, Header, error) {
	p, err := paths.VaultFile()
	if err != nil {
		return nil, Header{}, err
	}
	blob, err := os.ReadFile(p)
	if err != nil {
		return nil, Header{}, fmt.Errorf("read vault: %w", err)
	}
	pt, h, derr := DecryptWithKey(blob, key)
	return pt, h, derr
}


// SaveFile atomically writes a fresh vault using the supplied derived key and
// header salt. Pass DefaultParams unless you have a reason not to.
func SaveFile(plaintext, key []byte, salt [saltLen]byte, params Argon2idParams) error {
	if _, err := paths.EnsureConfigDir(); err != nil {
		return err
	}
	blob, err := EncryptWithKey(plaintext, key, salt, params)
	if err != nil {
		return err
	}
	return atomicWrite(blob)
}

// SaveFileFromPassword is the cold-start path: derives a fresh key from the
// password using DefaultParams. Used on first-time setup and change-password.
func SaveFileFromPassword(plaintext, password []byte) error {
	if _, err := paths.EnsureConfigDir(); err != nil {
		return err
	}
	blob, err := Encrypt(plaintext, password, DefaultParams)
	if err != nil {
		return err
	}
	return atomicWrite(blob)
}

func atomicWrite(blob []byte) error {
	p, err := paths.VaultFile()
	if err != nil {
		return err
	}
	dir := filepath.Dir(p)
	tmp, err := os.CreateTemp(dir, "vault.bin.*.tmp")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpPath := tmp.Name()
	cleanup := func() { _ = os.Remove(tmpPath) }

	if err := os.Chmod(tmpPath, paths.FileMode); err != nil {
		_ = tmp.Close()
		cleanup()
		return fmt.Errorf("chmod temp: %w", err)
	}
	if _, err := tmp.Write(blob); err != nil {
		_ = tmp.Close()
		cleanup()
		return fmt.Errorf("write temp: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		cleanup()
		return fmt.Errorf("sync temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return fmt.Errorf("close temp: %w", err)
	}
	if err := os.Rename(tmpPath, p); err != nil {
		cleanup()
		return fmt.Errorf("rename temp: %w", err)
	}
	return nil
}

// HeaderSalt is a small accessor so callers don't need to pull saltLen
// constants from this package when re-saving.
func (h Header) SaltBytes() [saltLen]byte { return h.Salt }

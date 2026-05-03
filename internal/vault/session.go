package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"golang.org/x/crypto/hkdf"

	"github.com/johnniewhite/ssher/internal/paths"
)

// SessionTimeout matches the Python version's 30-minute idle window.
const SessionTimeout = 30 * time.Minute

const (
	sessMagic   = "SSES"
	sessVersion = 1
	// header (4+1+8+saltLen) + nonce + cipherbody (key+tag).
	sessHeaderLen = 4 + 1 + 8 + saltLen
)

var ErrNoSession = errors.New("vault: no active session")

// SaveSession encrypts the unlocked vault key against a host-bound wrapping
// key and writes it to ~/.ssher/.session with an absolute expiry.
//
// The vaultSalt is also stored so subsequent commands can re-encrypt without
// generating a new salt (which would invalidate other cached keys).
func SaveSession(vaultKey []byte, vaultSalt [saltLen]byte) error {
	if len(vaultKey) != keyLen {
		return fmt.Errorf("vault: bad key length %d", len(vaultKey))
	}
	if _, err := paths.EnsureConfigDir(); err != nil {
		return err
	}

	wrap, err := wrappingKey()
	if err != nil {
		return err
	}
	defer wipe(wrap)

	var nonce [nonceLen]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return fmt.Errorf("rand nonce: %w", err)
	}
	block, err := aes.NewCipher(wrap)
	if err != nil {
		return fmt.Errorf("aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("gcm: %w", err)
	}
	ct := gcm.Seal(nil, nonce[:], vaultKey, nil)

	expires := time.Now().Add(SessionTimeout).UnixNano()

	out := make([]byte, 0, sessHeaderLen+nonceLen+len(ct))
	out = append(out, []byte(sessMagic)...)
	out = append(out, sessVersion)
	var u64 [8]byte
	binary.BigEndian.PutUint64(u64[:], uint64(expires))
	out = append(out, u64[:]...)
	out = append(out, vaultSalt[:]...)
	out = append(out, nonce[:]...)
	out = append(out, ct...)

	p, err := paths.SessionFile()
	if err != nil {
		return err
	}
	return writeFileAtomic(p, out)
}

// LoadSession returns the cached vault key and matching vault salt if the
// session is still valid. Returns ErrNoSession otherwise (callers should
// treat this as "prompt for password").
func LoadSession() (key []byte, salt [saltLen]byte, err error) {
	p, perr := paths.SessionFile()
	if perr != nil {
		return nil, salt, perr
	}
	blob, rerr := os.ReadFile(p)
	if errors.Is(rerr, os.ErrNotExist) {
		return nil, salt, ErrNoSession
	}
	if rerr != nil {
		return nil, salt, fmt.Errorf("read session: %w", rerr)
	}
	if len(blob) < sessHeaderLen+nonceLen {
		return nil, salt, ErrNoSession
	}
	if string(blob[:4]) != sessMagic {
		return nil, salt, ErrNoSession
	}
	if blob[4] != sessVersion {
		return nil, salt, ErrNoSession
	}
	expires := int64(binary.BigEndian.Uint64(blob[5:13]))
	if time.Now().UnixNano() > expires {
		_ = ClearSession()
		return nil, salt, ErrNoSession
	}
	copy(salt[:], blob[13:13+saltLen])
	nonce := blob[13+saltLen : 13+saltLen+nonceLen]
	ct := blob[13+saltLen+nonceLen:]

	wrap, werr := wrappingKey()
	if werr != nil {
		return nil, salt, werr
	}
	defer wipe(wrap)

	block, cerr := aes.NewCipher(wrap)
	if cerr != nil {
		return nil, salt, fmt.Errorf("aes cipher: %w", cerr)
	}
	gcm, gerr := cipher.NewGCM(block)
	if gerr != nil {
		return nil, salt, fmt.Errorf("gcm: %w", gerr)
	}
	pt, oerr := gcm.Open(nil, nonce, ct, nil)
	if oerr != nil {
		// Tampered or wrong host — treat as no session.
		return nil, salt, ErrNoSession
	}
	if len(pt) != keyLen {
		return nil, salt, ErrNoSession
	}
	return pt, salt, nil
}

// ClearSession removes the cached session file.
func ClearSession() error {
	p, err := paths.SessionFile()
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove session: %w", err)
	}
	return nil
}

// SessionInfo reports whether a session is active and how long until it
// expires. Used by `ssher vault status`.
func SessionInfo() (active bool, remaining time.Duration, err error) {
	p, perr := paths.SessionFile()
	if perr != nil {
		return false, 0, perr
	}
	blob, rerr := os.ReadFile(p)
	if errors.Is(rerr, os.ErrNotExist) {
		return false, 0, nil
	}
	if rerr != nil {
		return false, 0, fmt.Errorf("read session: %w", rerr)
	}
	if len(blob) < 13 || string(blob[:4]) != sessMagic {
		return false, 0, nil
	}
	expires := int64(binary.BigEndian.Uint64(blob[5:13]))
	now := time.Now().UnixNano()
	if now > expires {
		return false, 0, nil
	}
	return true, time.Duration(expires - now), nil
}

func wrappingKey() ([]byte, error) {
	hf, err := paths.HostFingerprint()
	if err != nil {
		return nil, err
	}
	r := hkdf.New(sha256.New, hf, []byte("ssher.session.salt.v1"), []byte("ssher session wrap"))
	out := make([]byte, keyLen)
	if _, err := io.ReadFull(r, out); err != nil {
		return nil, fmt.Errorf("hkdf: %w", err)
	}
	return out, nil
}

func writeFileAtomic(path string, data []byte) error {
	tmp, err := os.CreateTemp(pathDir(path), "session.*.tmp")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpPath := tmp.Name()
	if err := os.Chmod(tmpPath, paths.FileMode); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return os.Rename(tmpPath, path)
}

func pathDir(p string) string {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' {
			return p[:i]
		}
	}
	return "."
}

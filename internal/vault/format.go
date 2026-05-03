// Package vault implements the on-disk encrypted vault for ssher.
//
// File format (vault.bin):
//
//	[ magic       "SSHV" 4B          ]
//	[ version     uint8  1B          ] -- format version
//	[ kdf_id      uint8  1B          ] -- 1 = argon2id
//	[ argon2_time uint32 4B BE       ]
//	[ argon2_mem  uint32 4B BE (KiB) ]
//	[ argon2_par  uint8  1B          ]
//	[ salt        16B                ]
//	[ nonce       12B                ] -- AES-GCM nonce
//	[ ciphertext  var                ] -- AES-256-GCM(gzip(json)) + 16B tag
//
// The password is derived to a 32-byte AES key via Argon2id with the parameters
// stored in the header so we can ratchet costs upward over time without
// breaking existing vaults.
package vault

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

const (
	magic       = "SSHV"
	formatV1    = 1
	kdfArgon2id = 1

	saltLen  = 16
	nonceLen = 12
	keyLen   = 32

	headerLen = 4 + 1 + 1 + 4 + 4 + 1 + saltLen + nonceLen
)

// Argon2idParams controls the cost of password-to-key derivation.
type Argon2idParams struct {
	Time    uint32 // iterations
	MemKiB  uint32 // memory in KiB
	Threads uint8  // parallelism
}

// DefaultParams matches the OWASP "use Argon2id with m=64MiB, t=3, p=4"
// guidance as of the 2025 cutoff. Surfaced in the header so vaults written by
// older binaries continue to open after we raise the floor.
var DefaultParams = Argon2idParams{Time: 3, MemKiB: 64 * 1024, Threads: 4}

// MinParams is the floor applied on unlock; vaults below this are rewritten at
// DefaultParams during the next save.
var MinParams = Argon2idParams{Time: 2, MemKiB: 32 * 1024, Threads: 2}

var (
	ErrBadMagic    = errors.New("vault: bad magic (not a ssher vault file)")
	ErrBadVersion  = errors.New("vault: unsupported format version")
	ErrBadKDF      = errors.New("vault: unsupported KDF")
	ErrBadPassword = errors.New("vault: incorrect master password")
	ErrTruncated   = errors.New("vault: file truncated")
)

// Header is the parsed leading metadata of a vault file.
type Header struct {
	Version uint8
	KDFID   uint8
	Params  Argon2idParams
	Salt    [saltLen]byte
	Nonce   [nonceLen]byte
}

func (h Header) BelowMinParams() bool {
	return h.Params.Time < MinParams.Time ||
		h.Params.MemKiB < MinParams.MemKiB ||
		h.Params.Threads < MinParams.Threads
}

func parseHeader(b []byte) (Header, []byte, error) {
	if len(b) < headerLen {
		return Header{}, nil, ErrTruncated
	}
	if string(b[:4]) != magic {
		return Header{}, nil, ErrBadMagic
	}
	h := Header{
		Version: b[4],
		KDFID:   b[5],
	}
	if h.Version != formatV1 {
		return Header{}, nil, ErrBadVersion
	}
	if h.KDFID != kdfArgon2id {
		return Header{}, nil, ErrBadKDF
	}
	h.Params.Time = binary.BigEndian.Uint32(b[6:10])
	h.Params.MemKiB = binary.BigEndian.Uint32(b[10:14])
	h.Params.Threads = b[14]
	copy(h.Salt[:], b[15:15+saltLen])
	copy(h.Nonce[:], b[15+saltLen:15+saltLen+nonceLen])
	return h, b[headerLen:], nil
}

func writeHeader(buf *bytes.Buffer, h Header) {
	buf.WriteString(magic)
	buf.WriteByte(h.Version)
	buf.WriteByte(h.KDFID)
	var u32 [4]byte
	binary.BigEndian.PutUint32(u32[:], h.Params.Time)
	buf.Write(u32[:])
	binary.BigEndian.PutUint32(u32[:], h.Params.MemKiB)
	buf.Write(u32[:])
	buf.WriteByte(h.Params.Threads)
	buf.Write(h.Salt[:])
	buf.Write(h.Nonce[:])
}

// DeriveKey runs Argon2id on the password with the given salt/params.
func DeriveKey(password []byte, salt []byte, p Argon2idParams) []byte {
	return argon2.IDKey(password, salt, p.Time, p.MemKiB, p.Threads, keyLen)
}

// Encrypt produces a self-contained vault blob. Plaintext is gzipped before
// encryption; this both shrinks realistic vaults (lots of repeated JSON keys)
// and obscures plaintext length signals.
func Encrypt(plaintext, password []byte, params Argon2idParams) ([]byte, error) {
	var salt [saltLen]byte
	if _, err := rand.Read(salt[:]); err != nil {
		return nil, fmt.Errorf("rand salt: %w", err)
	}
	var nonce [nonceLen]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return nil, fmt.Errorf("rand nonce: %w", err)
	}

	key := DeriveKey(password, salt[:], params)
	defer wipe(key)

	var gz bytes.Buffer
	gw := gzip.NewWriter(&gz)
	if _, err := gw.Write(plaintext); err != nil {
		return nil, fmt.Errorf("gzip write: %w", err)
	}
	if err := gw.Close(); err != nil {
		return nil, fmt.Errorf("gzip close: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm: %w", err)
	}
	ct := gcm.Seal(nil, nonce[:], gz.Bytes(), nil)

	var out bytes.Buffer
	out.Grow(headerLen + len(ct))
	writeHeader(&out, Header{
		Version: formatV1,
		KDFID:   kdfArgon2id,
		Params:  params,
		Salt:    salt,
		Nonce:   nonce,
	})
	out.Write(ct)
	return out.Bytes(), nil
}

// Decrypt reverses Encrypt. Returns ErrBadPassword on auth failure (caller can
// distinguish wrong-password from corruption).
func Decrypt(blob, password []byte) ([]byte, Header, error) {
	h, ct, err := parseHeader(blob)
	if err != nil {
		return nil, Header{}, err
	}
	key := DeriveKey(password, h.Salt[:], h.Params)
	defer wipe(key)
	pt, err := openCiphertext(h, ct, key)
	return pt, h, err
}

// DecryptWithKey opens a vault blob using a pre-derived key, skipping
// Argon2id. Used when a session is active. Returns ErrBadPassword on auth
// failure (the key was derived for a different password / vault).
func DecryptWithKey(blob, key []byte) ([]byte, Header, error) {
	h, ct, err := parseHeader(blob)
	if err != nil {
		return nil, Header{}, err
	}
	pt, err := openCiphertext(h, ct, key)
	return pt, h, err
}

func openCiphertext(h Header, ct, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm: %w", err)
	}
	gz, err := gcm.Open(nil, h.Nonce[:], ct, nil)
	if err != nil {
		return nil, ErrBadPassword
	}
	gr, err := gzip.NewReader(bytes.NewReader(gz))
	if err != nil {
		return nil, fmt.Errorf("gunzip: %w", err)
	}
	defer gr.Close()
	pt, err := io.ReadAll(gr)
	if err != nil {
		return nil, fmt.Errorf("gunzip read: %w", err)
	}
	return pt, nil
}

// EncryptWithKey re-encrypts using a pre-derived key (no Argon2id work). Used
// when re-saving a vault we just decrypted: avoids paying the KDF cost twice
// in the same command.
func EncryptWithKey(plaintext, key []byte, salt [saltLen]byte, params Argon2idParams) ([]byte, error) {
	if len(key) != keyLen {
		return nil, fmt.Errorf("vault: bad key length %d", len(key))
	}
	var nonce [nonceLen]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return nil, fmt.Errorf("rand nonce: %w", err)
	}

	var gz bytes.Buffer
	gw := gzip.NewWriter(&gz)
	if _, err := gw.Write(plaintext); err != nil {
		return nil, fmt.Errorf("gzip write: %w", err)
	}
	if err := gw.Close(); err != nil {
		return nil, fmt.Errorf("gzip close: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm: %w", err)
	}
	ct := gcm.Seal(nil, nonce[:], gz.Bytes(), nil)

	var out bytes.Buffer
	out.Grow(headerLen + len(ct))
	writeHeader(&out, Header{
		Version: formatV1,
		KDFID:   kdfArgon2id,
		Params:  params,
		Salt:    salt,
		Nonce:   nonce,
	})
	out.Write(ct)
	return out.Bytes(), nil
}

func wipe(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

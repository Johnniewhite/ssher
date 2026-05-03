// Package pwgen produces cryptographically secure passwords.
package pwgen

import (
	"crypto/rand"
	"errors"
	"math/big"
)

const (
	Lower   = "abcdefghijklmnopqrstuvwxyz"
	Upper   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Digits  = "0123456789"
	Symbols = "!@#$%^&*()-_=+[]{};:,.<>?/"
)

// Options controls character classes and length.
type Options struct {
	Length        int
	IncludeUpper  bool
	IncludeLower  bool
	IncludeDigits bool
	IncludeSymbol bool
	Exclude       string // chars to exclude entirely (e.g. ambiguous: 0Ol1I)
}

// Default returns sane defaults: 20 chars, all classes, no exclusions.
func Default() Options {
	return Options{
		Length: 20, IncludeUpper: true, IncludeLower: true,
		IncludeDigits: true, IncludeSymbol: true,
	}
}

// Generate produces one password matching opts. Returns an error if no
// character classes are selected or length < 4.
func Generate(opts Options) (string, error) {
	if opts.Length < 4 {
		return "", errors.New("password length must be at least 4")
	}
	pool := buildPool(opts)
	if len(pool) == 0 {
		return "", errors.New("no character classes selected")
	}
	out := make([]byte, opts.Length)
	max := big.NewInt(int64(len(pool)))
	for i := range out {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		out[i] = pool[n.Int64()]
	}
	return string(out), nil
}

func buildPool(opts Options) []byte {
	var p []byte
	if opts.IncludeLower {
		p = append(p, Lower...)
	}
	if opts.IncludeUpper {
		p = append(p, Upper...)
	}
	if opts.IncludeDigits {
		p = append(p, Digits...)
	}
	if opts.IncludeSymbol {
		p = append(p, Symbols...)
	}
	if opts.Exclude == "" {
		return p
	}
	excl := map[byte]bool{}
	for i := 0; i < len(opts.Exclude); i++ {
		excl[opts.Exclude[i]] = true
	}
	out := p[:0]
	for _, c := range p {
		if !excl[c] {
			out = append(out, c)
		}
	}
	return out
}

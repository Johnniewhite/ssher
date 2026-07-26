package pwgen

import (
	"strings"
	"testing"
)

func TestGenerateHonorsExclusions(t *testing.T) {
	opts := Default()
	opts.Length = 64
	opts.Exclude = Lower + Upper + Digits
	password, err := Generate(opts)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for _, r := range password {
		if !strings.ContainsRune(Symbols, r) {
			t.Fatalf("generated excluded character %q", r)
		}
	}
}

func TestGenerateRejectsEmptyPool(t *testing.T) {
	opts := Options{Length: 20}
	if _, err := Generate(opts); err == nil {
		t.Fatal("empty character pool unexpectedly succeeded")
	}
}

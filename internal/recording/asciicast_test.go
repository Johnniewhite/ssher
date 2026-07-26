package recording

import (
	"os"
	"testing"
)

func TestNewWriterDoesNotCollide(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	first, firstPath, err := NewWriter("production", 80, 24)
	if err != nil {
		t.Fatalf("first writer: %v", err)
	}
	if err := first.Close(); err != nil {
		t.Fatalf("close first: %v", err)
	}
	second, secondPath, err := NewWriter("production", 80, 24)
	if err != nil {
		t.Fatalf("second writer: %v", err)
	}
	if err := second.Close(); err != nil {
		t.Fatalf("close second: %v", err)
	}
	if firstPath == secondPath {
		t.Fatalf("recordings collided at %s", firstPath)
	}
	for _, path := range []string{firstPath, secondPath} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat %s: %v", path, err)
		}
		if info.Mode().Perm() != 0o600 {
			t.Fatalf("%s mode = %o, want 600", path, info.Mode().Perm())
		}
	}
}

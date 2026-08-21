package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReplaceFileOverwritesExistingTarget(t *testing.T) {
	directory := t.TempDir()
	source := filepath.Join(directory, "source")
	target := filepath.Join(directory, "target")
	if err := os.WriteFile(source, []byte("new"), FileMode); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("old"), FileMode); err != nil {
		t.Fatal(err)
	}
	if err := ReplaceFile(source, target); err != nil {
		t.Fatalf("ReplaceFile: %v", err)
	}
	contents, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "new" {
		t.Fatalf("target contents = %q, want new", contents)
	}
	if _, err := os.Stat(source); !os.IsNotExist(err) {
		t.Fatalf("source still exists after replacement: %v", err)
	}
}

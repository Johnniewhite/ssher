package transfer

import "testing"

func TestShellQuoteProtectsRsyncSSHKeyPath(t *testing.T) {
	got := shellQuote("/tmp/key path/owner's key")
	want := `'/tmp/key path/owner'\''s key'`
	if got != want {
		t.Fatalf("shellQuote = %q, want %q", got, want)
	}
}

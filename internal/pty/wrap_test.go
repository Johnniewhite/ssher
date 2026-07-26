package pty

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestRunWithIONoPasswordPromptStillForwardsInput(t *testing.T) {
	var stdout bytes.Buffer
	done := make(chan error, 1)
	go func() {
		exit, err := runWithIO(
			[]string{"/bin/sh", "-c", `printf 'ready: '; read line; printf 'received:%s\n' "$line"`},
			"unused",
			DefaultPromptPatterns,
			strings.NewReader("continue\n"),
			&stdout,
		)
		if err == nil && exit != 0 {
			err = &unexpectedExitError{exit: exit}
		}
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("runWithIO: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("no-prompt command hung instead of receiving stdin")
	}
	if got := stdout.String(); !strings.Contains(got, "received:continue") {
		t.Fatalf("output did not contain forwarded input: %q", got)
	}
}

type unexpectedExitError struct {
	exit int
}

func (e *unexpectedExitError) Error() string {
	return "unexpected child exit"
}

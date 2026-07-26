package pty

import (
	"bytes"
	"io"
	"strings"
	"syscall"
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

func TestWatchAndInjectTreatsLinuxPTYCloseAsEOF(t *testing.T) {
	var stdout bytes.Buffer
	stream := &eioReadWriter{Reader: strings.NewReader("key auth complete\n")}

	if err := watchAndInject(stream, "unused", DefaultPromptPatterns, &stdout); err != nil {
		t.Fatalf("watchAndInject: %v", err)
	}
	if got := stdout.String(); got != "key auth complete\n" {
		t.Fatalf("stdout = %q", got)
	}
}

type eioReadWriter struct {
	io.Reader
}

func (rw *eioReadWriter) Read(p []byte) (int, error) {
	n, err := rw.Reader.Read(p)
	if err == io.EOF {
		return n, syscall.EIO
	}
	return n, err
}

func (rw *eioReadWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

//go:build windows

package pty

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestRunWithIOUsesWindowsConPTY(t *testing.T) {
	var output bytes.Buffer
	assertWindowsPTYRun(t, func() (int, error) {
		return runWithIO(
			[]string{"cmd.exe", "/D", "/Q", "/V:ON", "/C", `set /p line=ready: & echo received:!line!`},
			"unused",
			DefaultPromptPatterns,
			strings.NewReader("continue\r"),
			&output,
		)
	})
	if got := output.String(); !strings.Contains(got, "received:continue") {
		t.Fatalf("ConPTY did not forward input: %q", got)
	}
}

func TestRunWithIOInjectsPasswordThroughWindowsConPTY(t *testing.T) {
	var output bytes.Buffer
	assertWindowsPTYRun(t, func() (int, error) {
		return runWithIO(
			[]string{"cmd.exe", "/D", "/Q", "/V:ON", "/C", `set /p secret=Password: & echo received:!secret!`},
			"cloud-secret",
			DefaultPromptPatterns,
			strings.NewReader(""),
			&output,
		)
	})
	if got := output.String(); !strings.Contains(got, "received:cloud-secret") {
		t.Fatalf("ConPTY did not inject the prompt response: %q", got)
	}
}

func assertWindowsPTYRun(t *testing.T, run func() (int, error)) {
	t.Helper()
	done := make(chan error, 1)
	go func() {
		exit, err := run()
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
	case <-time.After(10 * time.Second):
		t.Fatal("Windows ConPTY command timed out")
	}
}

type unexpectedExitError struct{ exit int }

func (e *unexpectedExitError) Error() string { return "unexpected child exit" }

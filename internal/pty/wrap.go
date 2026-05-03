// Package pty wraps creack/pty for the one place in ssher where pseudoterminal
// password injection is unavoidable: the `wrap` subcommand, which is a
// drop-in for sshpass and must work with arbitrary user-supplied SSH-shaped
// commands (ssh, scp, rsync, git+ssh, ...).
//
// Everywhere else in the codebase we use native SSH auth via x/crypto/ssh —
// there is NO PTY/pexpect-equivalent code outside this package. Don't grow
// one elsewhere; route through here.
package pty

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/creack/pty"
	"golang.org/x/term"
)

// DefaultPromptPatterns mirrors what sshpass watches for. Lowercased before
// comparison so we can match "Password:", "password:", "Enter password:" etc.
var DefaultPromptPatterns = []string{
	"password:",
	"passphrase",
}

// Run executes argv with a PTY, watches its output for a password prompt, and
// injects the password once. After injection (or if no prompt is seen before
// the child exits), the PTY is wired bidirectionally to the local terminal
// so the user's session continues normally.
//
// Returns the child's exit code.
func Run(argv []string, password string, promptPatterns []string) (int, error) {
	if len(argv) == 0 {
		return -1, errors.New("wrap: no command supplied")
	}
	if len(promptPatterns) == 0 {
		promptPatterns = DefaultPromptPatterns
	}

	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Env = os.Environ() // forward env verbatim

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return -1, fmt.Errorf("start pty: %w", err)
	}
	defer ptmx.Close()

	// Mirror local TTY size into the PTY, and re-mirror on SIGWINCH.
	resizeCh := make(chan os.Signal, 1)
	signal.Notify(resizeCh, syscall.SIGWINCH)
	go func() {
		for range resizeCh {
			_ = pty.InheritSize(os.Stdin, ptmx)
		}
	}()
	resizeCh <- syscall.SIGWINCH
	defer signal.Stop(resizeCh)

	// Local raw mode so user keystrokes pass through cleanly post-injection.
	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		state, err := term.MakeRaw(fd)
		if err != nil {
			return -1, fmt.Errorf("raw mode: %w", err)
		}
		defer term.Restore(fd, state)
	}

	if err := watchAndInject(ptmx, password, promptPatterns); err != nil {
		return -1, err
	}

	// After injection, hand the PTY off to the user. Two goroutines: stdin
	// -> PTY and PTY -> stdout. The PTY -> stdout side terminates when the
	// child closes its side; we wait for that to know the session ended.
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); _, _ = io.Copy(ptmx, os.Stdin) }()
	go func() { defer wg.Done(); _, _ = io.Copy(os.Stdout, ptmx) }()

	werr := cmd.Wait()
	// Closing the PTY unblocks the io.Copy goroutines.
	_ = ptmx.Close()
	wg.Wait()

	if werr == nil {
		return 0, nil
	}
	var ee *exec.ExitError
	if errors.As(werr, &ee) {
		if status, ok := ee.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus(), nil
		}
		return 1, nil
	}
	return -1, werr
}

// watchAndInject reads PTY output, copying it to stdout as it streams, until
// it detects a prompt pattern; at that point it writes the password (with a
// trailing newline) to the PTY and returns. If the child exits before a
// prompt appears, returns nil (let the caller handle the exit code).
func watchAndInject(ptmx io.ReadWriter, password string, patterns []string) error {
	buf := make([]byte, 4096)
	var seen bytes.Buffer
	const maxScan = 64 * 1024 // never scan more than this for a prompt

	for {
		n, err := ptmx.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			_, _ = os.Stdout.Write(chunk)
			seen.Write(chunk)
			if seen.Len() > maxScan {
				// Prompt clearly didn't appear; stop scanning, treat as
				// no-prompt-needed (e.g. key auth succeeded).
				return nil
			}
			low := strings.ToLower(seen.String())
			for _, pat := range patterns {
				if strings.Contains(low, strings.ToLower(pat)) {
					if _, werr := io.WriteString(ptmx.(io.Writer), password+"\n"); werr != nil {
						return fmt.Errorf("write password: %w", werr)
					}
					return nil
				}
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
	}
}

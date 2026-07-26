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
	return runWithIO(argv, password, promptPatterns, os.Stdin, os.Stdout)
}

func runWithIO(argv []string, password string, promptPatterns []string, stdin io.Reader, stdout io.Writer) (int, error) {
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
	if stdinFile, ok := stdin.(*os.File); ok {
		resizeCh := make(chan os.Signal, 1)
		resizeDone := make(chan struct{})
		signal.Notify(resizeCh, syscall.SIGWINCH)
		go func() {
			for {
				select {
				case <-resizeCh:
					_ = pty.InheritSize(stdinFile, ptmx)
				case <-resizeDone:
					return
				}
			}
		}()
		resizeCh <- syscall.SIGWINCH
		defer func() {
			signal.Stop(resizeCh)
			close(resizeDone)
		}()

		// Local raw mode so user keystrokes pass through cleanly.
		fd := int(stdinFile.Fd())
		if term.IsTerminal(fd) {
			state, err := term.MakeRaw(fd)
			if err != nil {
				return -1, fmt.Errorf("raw mode: %w", err)
			}
			defer term.Restore(fd, state)
		}
	}

	// Forward stdin immediately. Key-authenticated commands may never display
	// a password prompt; waiting for one before forwarding input deadlocks an
	// otherwise successful interactive session.
	go func() { _, _ = io.Copy(ptmx, stdin) }()

	if err := watchAndInject(ptmx, password, promptPatterns, stdout); err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return -1, err
	}

	// The scanner hands PTY output off after injection (or after its bounded
	// scan). Only this direction is awaited: a terminal read from os.Stdin
	// cannot be portably cancelled and must not delay returning the child code.
	outputDone := make(chan struct{})
	go func() {
		defer close(outputDone)
		_, _ = io.Copy(stdout, ptmx)
	}()

	werr := cmd.Wait()
	// Closing the PTY unblocks the output copier.
	_ = ptmx.Close()
	<-outputDone

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
func watchAndInject(ptmx io.ReadWriter, password string, patterns []string, stdout io.Writer) error {
	buf := make([]byte, 4096)
	var seen bytes.Buffer
	const maxScan = 64 * 1024 // never scan more than this for a prompt

	for {
		n, err := ptmx.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			_, _ = stdout.Write(chunk)
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
			// Linux PTY masters return EIO when the slave side closes. It is
			// the PTY equivalent of EOF, not a failed wrapped command.
			if errors.Is(err, io.EOF) || errors.Is(err, syscall.EIO) {
				return nil
			}
			return err
		}
	}
}

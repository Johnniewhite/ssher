package ssh

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"sync"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"

	"github.com/johnniewhite/ssher/internal/store"
)

// InteractiveOptions tunes the interactive session.
type InteractiveOptions struct {
	// TeeOutput, if non-nil, receives a copy of every byte the remote PTY
	// emits. Used by `connect --record` to capture asciicast output.
	TeeOutput io.Writer
}

// Interactive runs a full PTY shell on the remote host. Stdin/stdout are
// piped through and the local terminal is put into raw mode so escape
// sequences (Ctrl-C, arrow keys, etc.) are forwarded verbatim.
//
// Port forwarding is set up *before* the shell starts and torn down on exit.
//
// This is the path the bare `ssher <name>` command takes.
func Interactive(c *Client, opts InteractiveOptions) error {
	sess, err := c.SSH.NewSession()
	if err != nil {
		return fmt.Errorf("new session: %w", err)
	}
	defer sess.Close()

	// Mirror the local terminal size into the remote PTY.
	fd := int(os.Stdin.Fd())
	width, height, err := term.GetSize(fd)
	if err != nil {
		width, height = 80, 24
	}
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	termType := os.Getenv("TERM")
	if termType == "" {
		termType = "xterm-256color"
	}
	if err := sess.RequestPty(termType, height, width, modes); err != nil {
		return fmt.Errorf("request pty: %w", err)
	}

	cancelForwards, err := setupForwards(c)
	if err != nil {
		return err
	}
	defer cancelForwards()

	sess.Stdin = os.Stdin
	if opts.TeeOutput != nil {
		sess.Stdout = io.MultiWriter(os.Stdout, opts.TeeOutput)
		sess.Stderr = io.MultiWriter(os.Stderr, opts.TeeOutput)
	} else {
		sess.Stdout = os.Stdout
		sess.Stderr = os.Stderr
	}

	// Local raw mode so the remote shell sees keystrokes immediately.
	if term.IsTerminal(fd) {
		state, err := term.MakeRaw(fd)
		if err != nil {
			return fmt.Errorf("raw mode: %w", err)
		}
		defer term.Restore(fd, state)
	}

	if err := sess.Shell(); err != nil {
		return fmt.Errorf("start shell: %w", err)
	}
	return sess.Wait()
}

// Run executes a non-interactive command on the remote host and returns its
// combined output and exit code. Used by `ssher exec`.
func Run(ctx context.Context, c *Client, cmd string) (output string, exitCode int, err error) {
	sess, serr := c.SSH.NewSession()
	if serr != nil {
		return "", -1, fmt.Errorf("new session: %w", serr)
	}
	defer sess.Close()

	type result struct {
		out []byte
		err error
	}
	done := make(chan result, 1)
	go func() {
		out, err := sess.CombinedOutput(cmd)
		done <- result{out: out, err: err}
	}()

	var out []byte
	var runErr error
	select {
	case res := <-done:
		out, runErr = res.out, res.err
	case <-ctx.Done():
		_ = sess.Close()
		return "", -1, ctx.Err()
	}
	exit := 0
	if runErr != nil {
		var ee *ssh.ExitError
		if isExitError(runErr, &ee) {
			exit = ee.ExitStatus()
			runErr = nil // exit-code != 0 isn't a Go error for our callers
		}
	}
	return string(out), exit, runErr
}

func isExitError(err error, into **ssh.ExitError) bool {
	for cur := err; cur != nil; {
		if ee, ok := cur.(*ssh.ExitError); ok {
			*into = ee
			return true
		}
		type unwrapper interface{ Unwrap() error }
		u, ok := cur.(unwrapper)
		if !ok {
			return false
		}
		cur = u.Unwrap()
	}
	return false
}

// setupForwards installs local and remote port forwards as configured on the
// server. Returns a cleanup func that closes the listeners.
func setupForwards(c *Client) (func(), error) {
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	var listeners []net.Listener

	for _, fwd := range c.Server.LocalForwards {
		l, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(fwd.LocalPort))
		if err != nil {
			cancel()
			closeAll(listeners)
			return nil, fmt.Errorf("local forward listen :%d: %w", fwd.LocalPort, err)
		}
		listeners = append(listeners, l)
		wg.Add(1)
		go acceptLocal(ctx, &wg, c.SSH, l, fwd)
	}

	for _, fwd := range c.Server.RemoteForwards {
		l, err := c.SSH.Listen("tcp", remoteListenAddress(fwd))
		if err != nil {
			cancel()
			closeAll(listeners)
			return nil, fmt.Errorf("remote forward listen :%d: %w", fwd.LocalPort, err)
		}
		listeners = append(listeners, l)
		wg.Add(1)
		go acceptRemote(ctx, &wg, l, fwd)
	}

	cleanup := func() {
		cancel()
		closeAll(listeners)
		wg.Wait()
	}
	return cleanup, nil
}

func closeAll(ls []net.Listener) {
	for _, l := range ls {
		_ = l.Close()
	}
}

func acceptLocal(ctx context.Context, wg *sync.WaitGroup, client *ssh.Client, l net.Listener, fwd store.PortForward) {
	defer wg.Done()
	for {
		local, err := l.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			return
		}
		go func() {
			defer local.Close()
			remote, err := client.Dial("tcp", net.JoinHostPort(fwd.RemoteHost, strconv.Itoa(fwd.RemotePort)))
			if err != nil {
				return
			}
			defer remote.Close()
			pipeBoth(local, remote)
		}()
	}
}

func acceptRemote(ctx context.Context, wg *sync.WaitGroup, l net.Listener, fwd store.PortForward) {
	defer wg.Done()
	for {
		remote, err := l.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			return
		}
		go func() {
			defer remote.Close()
			local, err := net.Dial("tcp", remoteTargetAddress(fwd))
			if err != nil {
				return
			}
			defer local.Close()
			pipeBoth(local, remote)
		}()
	}
}

func remoteListenAddress(fwd store.PortForward) string {
	return net.JoinHostPort("127.0.0.1", strconv.Itoa(fwd.LocalPort))
}

func remoteTargetAddress(fwd store.PortForward) string {
	host := fwd.RemoteHost
	if host == "" {
		host = "127.0.0.1"
	}
	return net.JoinHostPort(host, strconv.Itoa(fwd.RemotePort))
}

func pipeBoth(a, b net.Conn) {
	done := make(chan struct{}, 2)
	go func() { _, _ = io.Copy(a, b); done <- struct{}{} }()
	go func() { _, _ = io.Copy(b, a); done <- struct{}{} }()
	<-done
}

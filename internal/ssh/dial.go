// Package ssh wraps golang.org/x/crypto/ssh to provide ssher's connection
// surface: dial, jump-host chains, port forwarding, interactive sessions,
// non-interactive command exec.
//
// We intentionally do NOT shell out to /usr/bin/ssh. That choice lets us
// authenticate natively (no pexpect, no password injection), at the cost of
// not honouring exotic ~/.ssh/config directives. Documented in CLAUDE.md.
package ssh

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"

	"github.com/johnniewhite/ssher/internal/store"
)

// PassphrasePrompt, when set, is called to obtain the passphrase for an
// encrypted private key. It's a hook (nil by default) so this package stays
// free of a hard dependency on the interactive UI layer and remains testable
// without a TTY. cmd/root.go wires it to the ui password prompt.
var PassphrasePrompt func(keyPath string) ([]byte, error)
var knownHostsMu sync.Mutex

// Client wraps an *ssh.Client and remembers the *underlying* network conn so
// callers chaining jump hosts can close the whole stack on shutdown.
type Client struct {
	SSH    *ssh.Client
	Server *store.Server
	closer func() error

	closeOnce     sync.Once
	closeErr      error
	keepaliveStop chan struct{}
	keepaliveDone chan struct{}
}

func (c *Client) Close() error {
	if c == nil {
		return nil
	}
	c.closeOnce.Do(func() {
		if c.keepaliveStop != nil {
			close(c.keepaliveStop)
		}
		c.closeErr = c.SSH.Close()
		if c.closer != nil {
			if cerr := c.closer(); c.closeErr == nil {
				c.closeErr = cerr
			}
		}
		if c.keepaliveDone != nil {
			<-c.keepaliveDone
		}
	})
	return c.closeErr
}

// Dial connects to a server, walking the JumpHost chain if any.
//
// The chain is single-hop in the original Python (server.JumpHost names ONE
// other saved server). We honour that and don't attempt to traverse arbitrary
// chains — the data model doesn't support it and OpenSSH ProxyJump-style
// chains aren't represented in the vault.
func Dial(v *store.Vault, target *store.Server) (*Client, error) {
	cfg, authCleanup, err := buildClientConfig(target)
	if err != nil {
		return nil, err
	}
	defer authCleanup()
	addr := net.JoinHostPort(target.Host, strconv.Itoa(target.Port))

	// No jump host: direct dial.
	if target.JumpHost == "" {
		client, err := ssh.Dial("tcp", addr, cfg)
		if err != nil {
			return nil, fmt.Errorf("dial %s: %w", addr, err)
		}
		return newClient(client, target, nil), nil
	}

	// Jump host: dial the bastion first, then tunnel TCP to the target.
	jumpServer, _, err := v.ResolveTarget(target.JumpHost)
	if err != nil {
		return nil, fmt.Errorf("jump host %q: %w", target.JumpHost, err)
	}
	jumpCfg, jumpAuthCleanup, err := buildClientConfig(jumpServer)
	if err != nil {
		return nil, fmt.Errorf("jump host %q: %w", target.JumpHost, err)
	}
	defer jumpAuthCleanup()
	jumpAddr := net.JoinHostPort(jumpServer.Host, strconv.Itoa(jumpServer.Port))
	jumpClient, err := ssh.Dial("tcp", jumpAddr, jumpCfg)
	if err != nil {
		return nil, fmt.Errorf("dial jump host %s: %w", jumpAddr, err)
	}

	tunnel, err := jumpClient.Dial("tcp", addr)
	if err != nil {
		_ = jumpClient.Close()
		return nil, fmt.Errorf("tunnel through %s -> %s: %w", jumpAddr, addr, err)
	}

	conn, chans, reqs, err := ssh.NewClientConn(tunnel, addr, cfg)
	if err != nil {
		_ = tunnel.Close()
		_ = jumpClient.Close()
		return nil, fmt.Errorf("ssh handshake to %s via %s: %w", addr, jumpAddr, err)
	}
	client := ssh.NewClient(conn, chans, reqs)
	return newClient(client, target, func() error {
		_ = jumpClient.Close()
		return nil
	}), nil
}

func buildClientConfig(s *store.Server) (*ssh.ClientConfig, func(), error) {
	auths, authClosers, err := authMethods(s)
	if err != nil {
		return nil, func() {}, err
	}
	cleanup := func() {
		for _, closer := range authClosers {
			_ = closer.Close()
		}
	}
	hostKeyCallback, err := buildHostKeyCallback()
	if err != nil {
		cleanup()
		return nil, func() {}, err
	}
	timeout := time.Duration(s.ConnectionTimeout) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &ssh.ClientConfig{
		User:            s.User,
		Auth:            auths,
		HostKeyCallback: hostKeyCallback,
		Timeout:         timeout,
	}, cleanup, nil
}

func authMethods(s *store.Server) ([]ssh.AuthMethod, []io.Closer, error) {
	switch s.AuthType {
	case store.AuthPassword:
		if s.Password == "" {
			return nil, nil, errors.New("server has password auth but no password set")
		}
		return []ssh.AuthMethod{ssh.Password(s.Password)}, nil, nil

	case store.AuthKey:
		var methods []ssh.AuthMethod
		var closers []io.Closer
		var keyErrors []error
		var agentMethod ssh.AuthMethod
		var agentCloser io.Closer

		if agentAuth, closer, err := agentAuthMethod(); err == nil {
			agentMethod = agentAuth
			agentCloser = closer
		}

		paths := []string{s.KeyPath}
		if s.KeyPath == "" {
			home, err := os.UserHomeDir()
			if err == nil {
				paths = []string{
					filepath.Join(home, ".ssh", "id_ed25519"),
					filepath.Join(home, ".ssh", "id_ecdsa"),
					filepath.Join(home, ".ssh", "id_rsa"),
				}
			}
		}
		for _, path := range paths {
			if path == "" {
				continue
			}
			signer, err := signerFromFile(path)
			if err == nil {
				methods = append(methods, ssh.PublicKeys(signer))
				continue
			}
			if s.KeyPath != "" || !errors.Is(err, os.ErrNotExist) {
				keyErrors = append(keyErrors, err)
			}
		}
		if agentMethod != nil {
			methods = append(methods, agentMethod)
			closers = append(closers, agentCloser)
		}

		if len(methods) > 0 {
			return methods, closers, nil
		}
		for _, closer := range closers {
			_ = closer.Close()
		}
		if len(keyErrors) > 0 {
			return nil, nil, errors.Join(keyErrors...)
		}
		return nil, nil, errors.New("no usable SSH key or ssh-agent identity found")

	default:
		return nil, nil, fmt.Errorf("unknown auth type %q", s.AuthType)
	}
}

func signerFromFile(path string) (ssh.Signer, error) {
	key, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read key %s: %w", path, err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err == nil {
		return signer, nil
	}
	var missing *ssh.PassphraseMissingError
	if !errors.As(err, &missing) || PassphrasePrompt == nil {
		return nil, fmt.Errorf("parse key %s: %w", path, err)
	}
	passphrase, promptErr := PassphrasePrompt(path)
	if promptErr != nil {
		return nil, fmt.Errorf("read passphrase for %s: %w", path, promptErr)
	}
	defer func() {
		for i := range passphrase {
			passphrase[i] = 0
		}
	}()
	signer, err = ssh.ParsePrivateKeyWithPassphrase(key, passphrase)
	if err != nil {
		return nil, fmt.Errorf("parse key %s: %w", path, err)
	}
	return signer, nil
}

func newClient(client *ssh.Client, server *store.Server, closer func() error) *Client {
	c := &Client{SSH: client, Server: server, closer: closer}
	if server.KeepAlive <= 0 {
		return c
	}
	c.keepaliveStop = make(chan struct{})
	c.keepaliveDone = make(chan struct{})
	go func() {
		defer close(c.keepaliveDone)
		ticker := time.NewTicker(time.Duration(server.KeepAlive) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if _, _, err := client.SendRequest("keepalive@openssh.com", true, nil); err != nil {
					_ = client.Close()
					return
				}
			case <-c.keepaliveStop:
				return
			}
		}
	}()
	return c
}

// buildHostKeyCallback uses ~/.ssh/known_hosts when it exists. If not, we
// install an "accept on first use, persist" callback that mirrors OpenSSH's
// strict-host-key-checking=accept-new behaviour.
func buildHostKeyCallback() (ssh.HostKeyCallback, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve home for known_hosts: %w", err)
	}
	khPath := filepath.Join(home, ".ssh", "known_hosts")
	if _, err := os.Stat(khPath); errors.Is(err, os.ErrNotExist) {
		// Touch the file with restrictive perms so knownhosts.New works.
		if err := os.MkdirAll(filepath.Dir(khPath), 0o700); err != nil {
			return nil, fmt.Errorf("create ~/.ssh: %w", err)
		}
		f, err := os.OpenFile(khPath, os.O_CREATE|os.O_WRONLY, 0o600)
		if err != nil {
			return nil, fmt.Errorf("create known_hosts: %w", err)
		}
		_ = f.Close()
	}
	verify, err := knownhosts.New(khPath)
	if err != nil {
		return nil, fmt.Errorf("load known_hosts: %w", err)
	}
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		err := verify(hostname, remote, key)
		if err == nil {
			return nil
		}
		// On unknown host: append and accept (TOFU). On mismatched host:
		// fail hard -- this is the man-in-the-middle protection.
		var ke *knownhosts.KeyError
		if !errors.As(err, &ke) {
			return err
		}
		if len(ke.Want) > 0 {
			return fmt.Errorf("host key mismatch for %s: %w", hostname, err)
		}
		// Unknown host: serialize TOFU decisions and reload the file inside the
		// lock. Parallel fleet connections must not accept two different first
		// keys for the same hostname.
		knownHostsMu.Lock()
		defer knownHostsMu.Unlock()
		current, reloadErr := knownhosts.New(khPath)
		if reloadErr != nil {
			return fmt.Errorf("reload known_hosts: %w", reloadErr)
		}
		if currentErr := current(hostname, remote, key); currentErr == nil {
			return nil
		} else {
			var currentKeyErr *knownhosts.KeyError
			if !errors.As(currentErr, &currentKeyErr) || len(currentKeyErr.Want) > 0 {
				return fmt.Errorf("host key mismatch for %s: %w", hostname, currentErr)
			}
		}

		// Still unknown: append and accept.
		f, ferr := os.OpenFile(khPath, os.O_APPEND|os.O_WRONLY, 0o600)
		if ferr != nil {
			return fmt.Errorf("append known_hosts: %w", ferr)
		}
		defer f.Close()
		line := knownhosts.Line([]string{knownhosts.Normalize(hostname)}, key) + "\n"
		if _, werr := f.WriteString(line); werr != nil {
			return fmt.Errorf("write known_hosts: %w", werr)
		}
		fmt.Fprintf(os.Stderr, "warning: added %s to known_hosts (TOFU)\n", hostname)
		return nil
	}, nil
}

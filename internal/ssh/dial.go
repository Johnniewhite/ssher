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
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"

	"github.com/johnniewhite/ssher/internal/store"
)

// Client wraps an *ssh.Client and remembers the *underlying* network conn so
// callers chaining jump hosts can close the whole stack on shutdown.
type Client struct {
	SSH    *ssh.Client
	Server *store.Server
	closer func() error
}

func (c *Client) Close() error {
	if c == nil {
		return nil
	}
	if err := c.SSH.Close(); err != nil {
		_ = c.closer()
		return err
	}
	if c.closer != nil {
		return c.closer()
	}
	return nil
}

// Dial connects to a server, walking the JumpHost chain if any.
//
// The chain is single-hop in the original Python (server.JumpHost names ONE
// other saved server). We honour that and don't attempt to traverse arbitrary
// chains — the data model doesn't support it and OpenSSH ProxyJump-style
// chains aren't represented in the vault.
func Dial(v *store.Vault, target *store.Server) (*Client, error) {
	cfg, err := buildClientConfig(target)
	if err != nil {
		return nil, err
	}
	addr := net.JoinHostPort(target.Host, strconv.Itoa(target.Port))

	// No jump host: direct dial.
	if target.JumpHost == "" {
		client, err := ssh.Dial("tcp", addr, cfg)
		if err != nil {
			return nil, fmt.Errorf("dial %s: %w", addr, err)
		}
		return &Client{SSH: client, Server: target}, nil
	}

	// Jump host: dial the bastion first, then tunnel TCP to the target.
	jumpServer, _, err := v.ResolveTarget(target.JumpHost)
	if err != nil {
		return nil, fmt.Errorf("jump host %q: %w", target.JumpHost, err)
	}
	jumpCfg, err := buildClientConfig(jumpServer)
	if err != nil {
		return nil, fmt.Errorf("jump host %q: %w", target.JumpHost, err)
	}
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
	return &Client{
		SSH:    client,
		Server: target,
		closer: func() error {
			_ = jumpClient.Close()
			return nil
		},
	}, nil
}

func buildClientConfig(s *store.Server) (*ssh.ClientConfig, error) {
	auths, err := authMethods(s)
	if err != nil {
		return nil, err
	}
	hostKeyCallback, err := buildHostKeyCallback()
	if err != nil {
		return nil, err
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
	}, nil
}

func authMethods(s *store.Server) ([]ssh.AuthMethod, error) {
	switch s.AuthType {
	case store.AuthPassword:
		if s.Password == "" {
			return nil, errors.New("server has password auth but no password set")
		}
		return []ssh.AuthMethod{ssh.Password(s.Password)}, nil

	case store.AuthKey:
		path := s.KeyPath
		if path == "" {
			home, _ := os.UserHomeDir()
			path = filepath.Join(home, ".ssh", "id_rsa")
		}
		key, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read key %s: %w", path, err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			// Common case: passphrase-protected key. We don't currently
			// prompt for the passphrase -- documented limitation, easy to
			// extend later by checking for *ssh.PassphraseMissingError.
			return nil, fmt.Errorf("parse key %s: %w", path, err)
		}
		// Also try the SSH agent if available; lots of users keep their
		// keys agent-resident. Best-effort -- failure is silent.
		methods := []ssh.AuthMethod{ssh.PublicKeys(signer)}
		if agentAuth, err := agentAuthMethod(); err == nil {
			methods = append(methods, agentAuth)
		}
		return methods, nil

	default:
		return nil, fmt.Errorf("unknown auth type %q", s.AuthType)
	}
}

// buildHostKeyCallback uses ~/.ssh/known_hosts when it exists. If not, we
// install an "accept on first use, persist" callback that mirrors OpenSSH's
// strict-host-key-checking=accept-new behaviour.
func buildHostKeyCallback() (ssh.HostKeyCallback, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return ssh.InsecureIgnoreHostKey(), nil
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
		// Unknown host: append.
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

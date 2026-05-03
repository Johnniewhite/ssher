package ssh

import (
	"errors"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// agentAuthMethod returns an ssh.AuthMethod that delegates to a running
// ssh-agent reachable via $SSH_AUTH_SOCK. Returns an error (not silently nil)
// if the env var isn't set or the socket is unreachable, so callers can decide
// whether to surface or ignore.
func agentAuthMethod() (ssh.AuthMethod, error) {
	sock := os.Getenv("SSH_AUTH_SOCK")
	if sock == "" {
		return nil, errors.New("SSH_AUTH_SOCK not set")
	}
	conn, err := net.Dial("unix", sock)
	if err != nil {
		return nil, err
	}
	a := agent.NewClient(conn)
	return ssh.PublicKeysCallback(a.Signers), nil
}

package ssh

import (
	"errors"
	"io"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// agentAuthMethod returns an ssh.AuthMethod that delegates to a running
// ssh-agent reachable via $SSH_AUTH_SOCK. Returns an error (not silently nil)
// if the env var isn't set or the socket is unreachable, so callers can decide
// whether to surface or ignore.
func agentAuthMethod() (ssh.AuthMethod, io.Closer, error) {
	sock := os.Getenv("SSH_AUTH_SOCK")
	if sock == "" {
		return nil, nil, errors.New("SSH_AUTH_SOCK not set")
	}
	conn, err := net.Dial("unix", sock)
	if err != nil {
		return nil, nil, err
	}
	a := agent.NewClient(conn)
	return ssh.PublicKeysCallback(a.Signers), conn, nil
}

//go:build windows

package ssh

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/Microsoft/go-winio"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

const windowsOpenSSHAgentPipe = `\\.\pipe\openssh-ssh-agent`

// agentAuthMethod connects to the OpenSSH Authentication Agent service that
// ships with modern Windows. OpenSSH exposes it as a named pipe rather than a
// Unix-domain socket.
func agentAuthMethod() (ssh.AuthMethod, io.Closer, error) {
	pipe := strings.TrimSpace(os.Getenv("SSH_AUTH_SOCK"))
	if !strings.HasPrefix(strings.ToLower(pipe), `\\.\pipe\`) {
		pipe = windowsOpenSSHAgentPipe
	}
	timeout := 3 * time.Second
	conn, err := winio.DialPipe(pipe, &timeout)
	if err != nil {
		return nil, nil, err
	}
	client := agent.NewClient(conn)
	return ssh.PublicKeysCallback(client.Signers), conn, nil
}

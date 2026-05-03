package transfer

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/johnniewhite/ssher/internal/store"
)

// RsyncMode controls direction.
type RsyncMode int

const (
	RsyncPush RsyncMode = iota
	RsyncPull
)

// RsyncOptions tunes the shelled-out rsync invocation.
type RsyncOptions struct {
	Mode    RsyncMode
	DryRun  bool
	Delete  bool
	Verbose bool
	Args    []string // extra args appended verbatim
}

// Rsync shells out to /usr/bin/rsync. Only key-based auth is supported here
// because we don't currently inject passwords into rsync's child SSH process.
// Password-auth servers will see rsync prompt and fail.
func Rsync(s *store.Server, local, remote string, opts RsyncOptions) error {
	if s.AuthType == store.AuthPassword {
		return errors.New("rsync currently requires SSH key auth; password-auth servers are not supported (use scp/sftp instead)")
	}
	if _, err := exec.LookPath("rsync"); err != nil {
		return errors.New("rsync binary not found in PATH")
	}

	sshOpts := "ssh -p " + strconv.Itoa(s.Port) +
		" -o StrictHostKeyChecking=accept-new" +
		" -o ConnectTimeout=" + strconv.Itoa(maxInt(s.ConnectionTimeout, 30))

	if s.KeyPath != "" {
		sshOpts += " -i " + s.KeyPath
	} else {
		home, _ := os.UserHomeDir()
		sshOpts += " -i " + filepath.Join(home, ".ssh", "id_rsa")
	}

	args := []string{"-az", "--info=progress2", "-e", sshOpts}
	if opts.DryRun {
		args = append(args, "--dry-run")
	}
	if opts.Delete {
		args = append(args, "--delete")
	}
	if opts.Verbose {
		args = append(args, "-v")
	}
	args = append(args, opts.Args...)

	remoteSpec := fmt.Sprintf("%s@%s:%s", s.User, s.Host, remote)
	switch opts.Mode {
	case RsyncPush:
		args = append(args, local, remoteSpec)
	case RsyncPull:
		args = append(args, remoteSpec, local)
	}

	cmd := exec.Command("rsync", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

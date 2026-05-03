// Package transfer covers SCP, SFTP, and rsync. SFTP is the workhorse for
// our upload/download paths because it's a real protocol over the SSH
// channel; SCP is offered for compatibility but SFTP is preferred whenever
// we control both ends.
package transfer

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/pkg/sftp"

	sshc "github.com/johnniewhite/ssher/internal/ssh"
)

// Upload sends a local file or directory to the remote host via SFTP.
// Recursive when localPath is a directory.
func Upload(c *sshc.Client, localPath, remotePath string) error {
	client, err := sftp.NewClient(c.SSH)
	if err != nil {
		return fmt.Errorf("sftp: %w", err)
	}
	defer client.Close()

	st, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("stat %s: %w", localPath, err)
	}
	if st.IsDir() {
		return uploadDir(client, localPath, remotePath)
	}
	return uploadFile(client, localPath, remotePath, st.Mode())
}

func uploadFile(c *sftp.Client, src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open %s: %w", src, err)
	}
	defer in.Close()

	if err := c.MkdirAll(path.Dir(dst)); err != nil {
		return fmt.Errorf("mkdir %s: %w", path.Dir(dst), err)
	}
	out, err := c.Create(dst)
	if err != nil {
		return fmt.Errorf("create remote %s: %w", dst, err)
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copy: %w", err)
	}
	return c.Chmod(dst, mode)
}

func uploadDir(c *sftp.Client, src, dst string) error {
	if err := c.MkdirAll(dst); err != nil {
		return fmt.Errorf("mkdir %s: %w", dst, err)
	}
	return filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		remote := path.Join(dst, filepath.ToSlash(rel))
		if info.IsDir() {
			if rel == "." {
				return nil
			}
			return c.MkdirAll(remote)
		}
		return uploadFile(c, p, remote, info.Mode())
	})
}

// Download fetches a remote file or directory to the local filesystem.
func Download(c *sshc.Client, remotePath, localPath string) error {
	client, err := sftp.NewClient(c.SSH)
	if err != nil {
		return fmt.Errorf("sftp: %w", err)
	}
	defer client.Close()

	st, err := client.Stat(remotePath)
	if err != nil {
		return fmt.Errorf("stat %s: %w", remotePath, err)
	}
	if st.IsDir() {
		return downloadDir(client, remotePath, localPath)
	}
	return downloadFile(client, remotePath, localPath, st.Mode())
}

func downloadFile(c *sftp.Client, src, dst string, mode os.FileMode) error {
	in, err := c.Open(src)
	if err != nil {
		return fmt.Errorf("open remote %s: %w", src, err)
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(dst), err)
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("create local %s: %w", dst, err)
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func downloadDir(c *sftp.Client, src, dst string) error {
	walker := c.Walk(src)
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	for walker.Step() {
		if err := walker.Err(); err != nil {
			return err
		}
		rel, err := filepath.Rel(src, walker.Path())
		if err != nil {
			return err
		}
		local := filepath.Join(dst, rel)
		info := walker.Stat()
		if info.IsDir() {
			if rel == "." {
				continue
			}
			if err := os.MkdirAll(local, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := downloadFile(c, walker.Path(), local, info.Mode()); err != nil {
			return err
		}
	}
	return nil
}

// SFTPInteractive runs an interactive SFTP session. We don't reimplement the
// full sftp(1) REPL — we shell out to the system sftp binary instead, since
// users expect its exact behaviour and tab completion. This is the only
// transfer path that shells out under normal operation.
//
// (Rsync also shells out, by necessity — there's no Go-native rsync.)
func SFTPInteractive(c *sshc.Client) error {
	return errors.New("interactive sftp REPL is provided by `ssher exec sftp` for now; use `ssher upload`/`download` for scripted transfers")
}

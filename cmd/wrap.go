package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/johnniewhite/ssher/internal/pty"
)

var (
	wrapEnv      bool
	wrapFile     string
	wrapFD       int
	wrapPrompt   string
	wrapPwInline string
)

var wrapCmd = &cobra.Command{
	Use:   "wrap [-e | -f FILE | -d FD | -p PASSWORD] [-P PROMPT] -- <command>...",
	Short: "Run a command with a password injected at its prompt (sshpass replacement)",
	Long: `wrap is a drop-in sshpass replacement. The wrapped command runs in a
pseudoterminal; ssher watches its output for a password prompt and injects
the supplied password, then hands the terminal off to the user.

Example:
  ssher wrap -e ssh user@host                # password from $SSHPASS
  ssher wrap -f /etc/secret/pw scp file user@host:
  ssher wrap -d 3 ssh user@host              # password from fd 3
  ssher wrap -P "Enter passphrase:" git push # custom prompt`,
	Args:               cobra.MinimumNArgs(1),
	DisableFlagParsing: false,
	RunE: func(c *cobra.Command, args []string) error {
		password, err := readWrapPassword()
		if err != nil {
			return err
		}
		patterns := pty.DefaultPromptPatterns
		if wrapPrompt != "" {
			patterns = []string{wrapPrompt}
		}
		exit, err := pty.Run(args, password, patterns)
		if err != nil {
			return err
		}
		if exit != 0 {
			os.Exit(exit)
		}
		return nil
	},
}

func init() {
	wrapCmd.Flags().BoolVarP(&wrapEnv, "env", "e", false, "read password from $SSHPASS")
	wrapCmd.Flags().StringVarP(&wrapFile, "file", "f", "", "read password from FILE (first line)")
	wrapCmd.Flags().IntVarP(&wrapFD, "fd", "d", 0, "read password from file descriptor")
	wrapCmd.Flags().StringVarP(&wrapPwInline, "password", "p", "", "password on the command line (insecure)")
	wrapCmd.Flags().StringVarP(&wrapPrompt, "prompt", "P", "", "custom prompt substring to match")
	rootCmd.AddCommand(wrapCmd)
}

func readWrapPassword() (string, error) {
	chosen := 0
	if wrapEnv {
		chosen++
	}
	if wrapFile != "" {
		chosen++
	}
	if wrapFD != 0 {
		chosen++
	}
	if wrapPwInline != "" {
		chosen++
	}
	if chosen == 0 {
		return "", errors.New("one of -e/-f/-d/-p is required")
	}
	if chosen > 1 {
		return "", errors.New("-e, -f, -d, -p are mutually exclusive")
	}

	switch {
	case wrapEnv:
		pw := os.Getenv("SSHPASS")
		if pw == "" {
			return "", errors.New("SSHPASS env var is empty")
		}
		return pw, nil
	case wrapFile != "":
		b, err := os.ReadFile(wrapFile)
		if err != nil {
			return "", fmt.Errorf("read password file: %w", err)
		}
		return strings.SplitN(strings.TrimRight(string(b), "\r\n"), "\n", 2)[0], nil
	case wrapFD != 0:
		f := os.NewFile(uintptr(wrapFD), fmt.Sprintf("fd-%d", wrapFD))
		if f == nil {
			return "", fmt.Errorf("invalid file descriptor %d", wrapFD)
		}
		defer f.Close()
		b, err := readAllAtMost(f, 4096)
		if err != nil {
			return "", fmt.Errorf("read fd %d: %w", wrapFD, err)
		}
		return strings.SplitN(strings.TrimRight(string(b), "\r\n"), "\n", 2)[0], nil
	default:
		return wrapPwInline, nil
	}
}

func readAllAtMost(f *os.File, max int) ([]byte, error) {
	buf := make([]byte, 0, 256)
	tmp := make([]byte, 256)
	for len(buf) < max {
		n, err := f.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return buf, err
		}
		if n == 0 {
			break
		}
	}
	return buf, nil
}

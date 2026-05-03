// Package clipboard writes text to the system clipboard. We shell out to
// platform-native tools (pbcopy on macOS, xclip / wl-copy on Linux) rather
// than pulling in a CGO clipboard library — keeps the Go binary CGO-free
// and the dependency story simple.
package clipboard

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Copy writes text to the system clipboard.
func Copy(text string) error {
	cmd, err := chooseCopyCommand()
	if err != nil {
		return err
	}
	cmd.Stdin = strings.NewReader(text)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("clipboard tool failed: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func chooseCopyCommand() (*exec.Cmd, error) {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("pbcopy"), nil
	case "linux":
		if _, err := exec.LookPath("wl-copy"); err == nil {
			return exec.Command("wl-copy"), nil
		}
		if _, err := exec.LookPath("xclip"); err == nil {
			return exec.Command("xclip", "-selection", "clipboard"), nil
		}
		if _, err := exec.LookPath("xsel"); err == nil {
			return exec.Command("xsel", "--clipboard", "--input"), nil
		}
		return nil, errors.New("no clipboard tool found (install wl-copy, xclip, or xsel)")
	default:
		return nil, fmt.Errorf("clipboard not supported on %s", runtime.GOOS)
	}
}

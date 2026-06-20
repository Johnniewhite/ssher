package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/johnniewhite/ssher/internal/clipboard"
)

// clipClearTTL is how long a copied password lingers before auto-clear. It's a
// var so `copy`/`generate-password` can expose a --clear-after override.
var clipClearTTL = 20 * time.Second

// clipClearCmd is an internal, hidden helper. `copy` and `generate-password`
// re-exec ssher with this subcommand, detached, so the parent can exit
// immediately while a background process waits out the TTL and then clears the
// clipboard — but only if it still holds the secret (sha256-compared, so we
// never clobber something the user copied in the meantime). The secret is
// never passed in argv; only its hash is.
var clipClearCmd = &cobra.Command{
	Use:    "__clipboard-clear <ttl-seconds> <sha256-hex>",
	Hidden: true,
	Args:   cobra.ExactArgs(2),
	RunE: func(c *cobra.Command, args []string) error {
		secs, err := strconv.Atoi(args[0])
		if err != nil || secs < 0 {
			return nil
		}
		want := args[1]
		time.Sleep(time.Duration(secs) * time.Second)
		cur, err := clipboard.Paste()
		if err != nil {
			return nil
		}
		sum := sha256.Sum256([]byte(cur))
		if hex.EncodeToString(sum[:]) == want {
			_ = clipboard.Clear()
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(clipClearCmd)
}

// spawnClipboardClear launches the detached clear helper for secret. A ttl of
// zero (or a non-positive value) disables auto-clear. Best-effort: any failure
// to spawn is swallowed — we never want clipboard hygiene to break a copy.
func spawnClipboardClear(secret string, ttl time.Duration) {
	if ttl <= 0 || secret == "" {
		return
	}
	exe, err := os.Executable()
	if err != nil {
		return
	}
	sum := sha256.Sum256([]byte(secret))
	cmd := exec.Command(exe, "__clipboard-clear",
		strconv.Itoa(int(ttl.Seconds())), hex.EncodeToString(sum[:]))
	// Setsid detaches the child from our process group so it survives our exit.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	_ = cmd.Start()
}

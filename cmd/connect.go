package cmd

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/johnniewhite/ssher/internal/recording"
	sshc "github.com/johnniewhite/ssher/internal/ssh"
	"github.com/johnniewhite/ssher/internal/store"
	"github.com/johnniewhite/ssher/internal/ui"
)

func termGetSize() (int, int, error) {
	return term.GetSize(int(os.Stdin.Fd()))
}

var (
	connectReconnect    bool
	connectMaxRetries   int
	connectRecord       bool
	connectRetryBackoff = []time.Duration{
		2 * time.Second, 5 * time.Second, 10 * time.Second,
		20 * time.Second, 30 * time.Second,
	}
)

var connectCmd = &cobra.Command{
	Use:   "connect [target]",
	Short: "Connect to a saved server",
	Args:  cobra.ExactArgs(1),
	RunE:  runConnect,
}

func init() {
	connectCmd.Flags().BoolVar(&connectReconnect, "reconnect", false, "auto-reconnect on disconnect")
	connectCmd.Flags().IntVar(&connectMaxRetries, "max-retries", 0, "override max reconnection attempts")
	connectCmd.Flags().BoolVar(&connectRecord, "record", false, "record the session to ~/.ssher/recordings")
	// Mirror connect flags on the root command so the documented shorthand
	// `ssher prod --reconnect` behaves like `ssher connect prod --reconnect`.
	rootCmd.Flags().BoolVar(&connectReconnect, "reconnect", false, "auto-reconnect on disconnect")
	rootCmd.Flags().IntVar(&connectMaxRetries, "max-retries", 0, "override max reconnection attempts")
	rootCmd.Flags().BoolVar(&connectRecord, "record", false, "record the session to ~/.ssher/recordings")
	rootCmd.AddCommand(connectCmd)
}

func runConnect(c *cobra.Command, args []string) error {
	saved, err := ui.LoadVault()
	if err != nil {
		return err
	}
	target, how, err := saved.Vault.ResolveTarget(args[0])
	if err != nil {
		return err
	}
	if how != "by name" && how != "by index 1" {
		fmt.Println(ui.Muted.Render(fmt.Sprintf("(matched %s)", how)))
	}

	reconnectSet := c != nil && c.Flags().Changed("reconnect")
	retriesSet := c != nil && c.Flags().Changed("max-retries")
	reconnect, maxRetries, err := resolveConnectionPolicy(
		target, connectReconnect, reconnectSet, connectMaxRetries, retriesSet,
	)
	if err != nil {
		return err
	}

	for attempt := 0; ; attempt++ {
		err := connectOnce(saved, target)
		if err == nil {
			return nil
		}
		if !reconnect || attempt >= maxRetries {
			return err
		}
		backoff := connectRetryBackoff[min(attempt, len(connectRetryBackoff)-1)]
		fmt.Println(ui.Warnf(fmt.Sprintf("disconnected: %v -- retry %d/%d in %s", err, attempt+1, maxRetries, backoff)))
		time.Sleep(backoff)
	}
}

func resolveConnectionPolicy(target *store.Server, reconnectValue, reconnectSet bool, retriesValue int, retriesSet bool) (bool, int, error) {
	reconnect := target.AutoReconnect
	if reconnectSet {
		reconnect = reconnectValue
	}
	maxRetries := target.MaxReconnectRetries
	if retriesSet {
		if retriesValue < 0 {
			return false, 0, fmt.Errorf("--max-retries cannot be negative")
		}
		maxRetries = retriesValue
	} else if maxRetries <= 0 {
		maxRetries = 3
	}
	return reconnect, maxRetries, nil
}

func connectOnce(saved *store.Saved, target *store.Server) error {
	start := time.Now()
	client, err := sshc.Dial(saved.Vault, target)
	if err != nil {
		historyErr := recordHistory(saved, target, time.Since(start), false, false, err.Error())
		return errors.Join(err, historyErr)
	}
	defer client.Close()

	fmt.Println(ui.Successf(fmt.Sprintf("connected to %s (%s@%s:%d)", target.Name, target.User, target.Host, target.Port)))
	if target.X11Forward {
		fmt.Println(ui.Warnf("native X11 forwarding is unsupported; X11 settings apply only to `export-config`"))
	}

	opts := sshc.InteractiveOptions{}
	if connectRecord {
		w, ttyW, ttyH := openRecording(target.Name)
		if w != nil {
			defer w.Close()
			opts.TeeOutput = w
			fmt.Println(ui.Infof(fmt.Sprintf("recording session (%dx%d) to ~/.ssher/recordings/", ttyW, ttyH)))
		}
	}

	err = sshc.Interactive(client, opts)
	dur := time.Since(start)
	historyErr := recordHistory(saved, target, dur, true, err == nil, errString(err))
	return errors.Join(err, historyErr)
}

func openRecording(title string) (*recording.Writer, int, int) {
	w, h := ui.TerminalWidth(), 24
	if termH := terminalHeight(); termH > 0 {
		h = termH
	}
	rec, _, err := recording.NewWriter(title, w, h)
	if err != nil {
		fmt.Println(ui.Warnf(fmt.Sprintf("could not start recording: %v", err)))
		return nil, w, h
	}
	return rec, w, h
}

func terminalHeight() int {
	_, h, err := termGetSize()
	if err != nil {
		return 0
	}
	return h
}

func recordHistory(saved *store.Saved, s *store.Server, dur time.Duration, established, ok bool, msg string) error {
	saved.Vault.RecordHistory(store.HistoryEntry{
		ServerName:   s.Name,
		Host:         s.Host,
		User:         s.User,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		Duration:     int(dur.Seconds()),
		Success:      ok,
		ErrorMessage: msg,
	})
	if established {
		if err := saved.Vault.UpdateServer(s.Name, func(srv *store.Server) {
			srv.LastConnected = time.Now().UTC().Format(time.RFC3339)
			srv.ConnectionCount++
		}); err != nil {
			return err
		}
	}
	return saved.SaveAndRefreshSession()
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

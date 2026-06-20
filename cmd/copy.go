package cmd

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/johnniewhite/ssher/internal/clipboard"
	"github.com/johnniewhite/ssher/internal/ui"
)

var copyField string
var copyClearAfter time.Duration

var copyCmd = &cobra.Command{
	Use:   "copy <target>",
	Short: "Copy a server field to the clipboard",
	Long: `Fields:
  host       remote hostname
  user       login user
  password   plaintext password (if stored)
  command    full ssh command line

Default field is "command".`,
	Args: cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		saved, err := ui.LoadVault()
		if err != nil {
			return err
		}
		target, _, err := saved.Vault.ResolveTarget(args[0])
		if err != nil {
			return err
		}

		var text string
		switch copyField {
		case "", "command":
			text = fmt.Sprintf("ssh -p %d %s@%s", target.Port, target.User, target.Host)
		case "host":
			text = target.Host
		case "user":
			text = target.User
		case "password":
			if target.Password == "" {
				return errors.New("server has no stored password")
			}
			text = target.Password
		default:
			return fmt.Errorf("unknown field %q", copyField)
		}
		if err := clipboard.Copy(text); err != nil {
			return err
		}
		msg := fmt.Sprintf("copied %s of %s to clipboard", copyField, target.Name)
		// Don't let a stored password linger on the clipboard.
		if copyField == "password" && copyClearAfter > 0 {
			spawnClipboardClear(text, copyClearAfter)
			msg += fmt.Sprintf(" (clears in %s)", copyClearAfter)
		}
		fmt.Println(ui.Successf(msg))
		return nil
	},
}

func init() {
	copyCmd.Flags().StringVar(&copyField, "field", "command", "host|user|password|command")
	copyCmd.Flags().DurationVar(&copyClearAfter, "clear-after", clipClearTTL, "auto-clear a copied password after this long (0 disables)")
	rootCmd.AddCommand(copyCmd)
}

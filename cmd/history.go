package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/johnniewhite/ssher/internal/ui"
)

var historyLimit int

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show connection history",
	RunE: func(c *cobra.Command, args []string) error {
		saved, err := ui.LoadVault()
		if err != nil {
			return err
		}
		entries := saved.Vault.History
		if len(entries) == 0 {
			fmt.Println(ui.Muted.Render("no history yet"))
			return nil
		}
		if historyLimit > 0 && len(entries) > historyLimit {
			entries = entries[len(entries)-historyLimit:]
		}
		for _, e := range entries {
			status := ui.Success.Render("ok")
			if !e.Success {
				status = ui.ErrorSty.Render("fail")
			}
			fmt.Printf("%s  %s  %s@%s  %s  %ds\n",
				e.Timestamp, status, e.User, e.Host, e.ServerName, e.Duration)
			if e.ErrorMessage != "" {
				fmt.Printf("    %s\n", ui.Muted.Render(e.ErrorMessage))
			}
		}
		return nil
	},
}

func init() {
	historyCmd.Flags().IntVarP(&historyLimit, "limit", "n", 50, "show last N entries (0 for all)")
	rootCmd.AddCommand(historyCmd)
}

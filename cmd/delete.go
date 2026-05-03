package cmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/johnniewhite/ssher/internal/ui"
)

var deleteForce bool

var deleteCmd = &cobra.Command{
	Use:     "delete [target]",
	Aliases: []string{"rm", "remove"},
	Short:   "Delete a saved server",
	Args:    cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		saved, err := ui.LoadVault()
		if err != nil {
			return err
		}
		target, _, err := saved.Vault.ResolveTarget(args[0])
		if err != nil {
			return err
		}

		if !deleteForce {
			var ok bool
			if err := huh.NewConfirm().
				Title(fmt.Sprintf("Delete %s (%s@%s)?", target.Name, target.User, target.Host)).
				Affirmative("Delete").
				Negative("Cancel").
				Value(&ok).
				Run(); err != nil {
				return err
			}
			if !ok {
				fmt.Println(ui.Muted.Render("cancelled"))
				return nil
			}
		}

		name := target.Name
		if err := saved.Vault.DeleteServer(name); err != nil {
			return err
		}
		if err := saved.SaveAndRefreshSession(); err != nil {
			return err
		}
		fmt.Println(ui.Successf(fmt.Sprintf("deleted %s", name)))
		return nil
	},
}

func init() {
	deleteCmd.Flags().BoolVarP(&deleteForce, "force", "f", false, "skip confirmation")
	rootCmd.AddCommand(deleteCmd)
}

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/johnniewhite/ssher/internal/store"
	"github.com/johnniewhite/ssher/internal/ui"
)

var favoriteCmd = &cobra.Command{
	Use:     "favorite [target]",
	Aliases: []string{"fav", "star"},
	Short:   "Toggle favorite status on a server",
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
		var newState bool
		if err := saved.Vault.UpdateServer(target.Name, func(s *store.Server) {
			s.IsFavorite = !s.IsFavorite
			newState = s.IsFavorite
		}); err != nil {
			return err
		}
		if err := saved.SaveAndRefreshSession(); err != nil {
			return err
		}
		state := "removed from favorites"
		if newState {
			state = "marked favorite"
		}
		fmt.Println(ui.Successf(fmt.Sprintf("%s %s", target.Name, state)))
		return nil
	},
}

func init() { rootCmd.AddCommand(favoriteCmd) }

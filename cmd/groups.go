package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/johnniewhite/ssher/internal/ui"
)

var groupsCmd = &cobra.Command{
	Use:   "groups",
	Short: "List groups and the servers in each",
	RunE: func(c *cobra.Command, args []string) error {
		saved, err := ui.LoadVault()
		if err != nil {
			return err
		}
		groups := saved.Vault.Groups()
		if len(groups) == 0 {
			fmt.Println(ui.Muted.Render("no servers yet"))
			return nil
		}
		for _, g := range groups {
			members := saved.Vault.ServersInGroup(g)
			fmt.Println(ui.Heading.Render(fmt.Sprintf("%s (%d)", g, len(members))))
			for _, s := range members {
				fmt.Printf("  %s %s@%s\n", ui.Title.Render(s.Name), s.User, s.Host)
			}
			fmt.Println()
		}
		return nil
	},
}

func init() { rootCmd.AddCommand(groupsCmd) }

package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/johnniewhite/ssher/internal/ui"
)

var aliasCmd = &cobra.Command{
	Use:   "alias",
	Short: "Manage server aliases",
}

var aliasListCmd = &cobra.Command{
	Use:   "list",
	Short: "List aliases",
	RunE: func(c *cobra.Command, args []string) error {
		saved, err := ui.LoadVault()
		if err != nil {
			return err
		}
		if len(saved.Vault.Aliases) == 0 {
			fmt.Println(ui.Muted.Render("no aliases"))
			return nil
		}
		keys := make([]string, 0, len(saved.Vault.Aliases))
		for k := range saved.Vault.Aliases {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Printf("  %s -> %s\n", ui.Title.Render(k), saved.Vault.Aliases[k])
		}
		return nil
	},
}

var aliasAddCmd = &cobra.Command{
	Use:   "add <server> <alias>",
	Short: "Add an alias for a server",
	Args:  cobra.ExactArgs(2),
	RunE: func(c *cobra.Command, args []string) error {
		saved, err := ui.LoadVault()
		if err != nil {
			return err
		}
		target, _, err := saved.Vault.ResolveTarget(args[0])
		if err != nil {
			return err
		}
		alias := args[1]
		if existing, ok := saved.Vault.Aliases[alias]; ok {
			return fmt.Errorf("alias %q already maps to %s", alias, existing)
		}
		saved.Vault.Aliases[alias] = target.Name
		if err := saved.SaveAndRefreshSession(); err != nil {
			return err
		}
		fmt.Println(ui.Successf(fmt.Sprintf("%s -> %s", alias, target.Name)))
		return nil
	},
}

var aliasRemoveCmd = &cobra.Command{
	Use:     "remove <alias>",
	Aliases: []string{"rm", "delete"},
	Short:   "Remove an alias",
	Args:    cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		saved, err := ui.LoadVault()
		if err != nil {
			return err
		}
		alias := args[0]
		if _, ok := saved.Vault.Aliases[alias]; !ok {
			return fmt.Errorf("alias %q not found", alias)
		}
		delete(saved.Vault.Aliases, alias)
		if err := saved.SaveAndRefreshSession(); err != nil {
			return err
		}
		fmt.Println(ui.Successf(fmt.Sprintf("removed %s", alias)))
		return nil
	},
}

func init() {
	aliasCmd.AddCommand(aliasListCmd, aliasAddCmd, aliasRemoveCmd)
	rootCmd.AddCommand(aliasCmd)
}

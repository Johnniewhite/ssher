package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/johnniewhite/ssher/internal/store"
	"github.com/johnniewhite/ssher/internal/ui"
)

var listGroup string
var listFavoritesOnly bool
var listSearch string

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List saved servers",
	RunE: func(c *cobra.Command, args []string) error {
		saved, err := ui.LoadVault()
		if err != nil {
			return err
		}
		view := saved.Vault.SortedServers()
		view = filterServers(view, listGroup, listFavoritesOnly, listSearch)

		fmt.Println(ui.Title.Render(fmt.Sprintf("Saved servers (%d)", len(view))))
		fmt.Println(ui.RenderServerTable(view))
		return nil
	},
}

func init() {
	listCmd.Flags().StringVarP(&listGroup, "group", "g", "", "filter by group")
	listCmd.Flags().BoolVarP(&listFavoritesOnly, "favorites", "f", false, "favorites only")
	listCmd.Flags().StringVarP(&listSearch, "search", "s", "", "case-insensitive substring match on name/host/tags")
	rootCmd.AddCommand(listCmd)
}

func filterServers(in []store.Server, group string, favoritesOnly bool, search string) []store.Server {
	if group == "" && !favoritesOnly && search == "" {
		return in
	}
	q := strings.ToLower(search)
	out := make([]store.Server, 0, len(in))
	for _, s := range in {
		if group != "" && s.Group != group {
			continue
		}
		if favoritesOnly && !s.IsFavorite {
			continue
		}
		if q != "" && !matchesSearch(s, q) {
			continue
		}
		out = append(out, s)
	}
	return out
}

func matchesSearch(s store.Server, q string) bool {
	if strings.Contains(strings.ToLower(s.Name), q) {
		return true
	}
	if strings.Contains(strings.ToLower(s.Host), q) {
		return true
	}
	for _, t := range s.Tags {
		if strings.Contains(strings.ToLower(t), q) {
			return true
		}
	}
	for _, a := range s.Aliases {
		if strings.Contains(strings.ToLower(a), q) {
			return true
		}
	}
	return false
}

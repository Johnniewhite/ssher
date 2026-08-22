package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/johnniewhite/ssher/internal/openssh"
	"github.com/johnniewhite/ssher/internal/store"
	"github.com/johnniewhite/ssher/internal/ui"
)

var (
	openSSHConfigPath  string
	openSSHReplace     bool
	openSSHDryRun      bool
	openSSHPush        bool
	openSSHIncludeKeys bool
)

var importOpenSSHCmd = &cobra.Command{
	Use:     "import-ssh-config [path]",
	Aliases: []string{"import-openssh"},
	Short:   "Import concrete hosts from OpenSSH configuration",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		path := openSSHConfigPath
		if len(args) == 1 {
			path = args[0]
		}
		_, err := importOpenSSH(c.Context(), path, openSSHReplace, openSSHDryRun, openSSHPush, openSSHIncludeKeys)
		return err
	},
}

var cloudImportOpenSSHCmd = &cobra.Command{
	Use:   "import-ssh-config [path]",
	Short: "Import OpenSSH hosts and immediately push them to Cloud",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		path := openSSHConfigPath
		if len(args) == 1 {
			path = args[0]
		}
		_, err := importOpenSSH(c.Context(), path, openSSHReplace, openSSHDryRun, !openSSHDryRun, openSSHIncludeKeys)
		return err
	},
}

func importOpenSSH(ctx context.Context, path string, replace, dryRun, push, includeKeys bool) (int, error) {
	result, err := openssh.Import(ctx, path, nil)
	for _, warning := range result.Warnings {
		fmt.Fprintln(os.Stderr, ui.Warnf(warning))
	}
	if err != nil {
		return 0, err
	}
	if len(result.Servers) == 0 {
		fmt.Println(ui.Muted.Render("no concrete OpenSSH Host aliases found; wildcard-only entries are not importable servers"))
		return 0, nil
	}
	saved, err := ui.LoadVault()
	if err != nil {
		return 0, err
	}
	added, updated, skipped := 0, 0, 0
	for _, imported := range result.Servers {
		index := vaultServerIndexByName(saved.Vault, imported.Name)
		if index >= 0 {
			if !replace {
				skipped++
				continue
			}
			merged := mergeOpenSSHServer(saved.Vault.Servers[index], imported)
			saved.Vault.Servers[index] = merged
			updated++
			continue
		}
		if err := saved.Vault.AddServer(imported); err != nil {
			return added + updated, err
		}
		added++
	}
	changed := added + updated
	if dryRun {
		fmt.Printf("OpenSSH import preview: %d new, %d updated, %d unchanged\n", added, updated, skipped)
		for _, server := range result.Servers {
			fmt.Printf("  %-24s %s@%s:%d\n", server.Name, server.User, server.Host, server.Port)
		}
		return changed, nil
	}
	if changed > 0 {
		if err := saved.SaveAndRefreshSession(); err != nil {
			return 0, err
		}
	}
	fmt.Println(ui.Successf(fmt.Sprintf("OpenSSH import complete (%d new, %d updated, %d unchanged)", added, updated, skipped)))
	if push {
		cloudIncludeKeys = includeKeys
		cloudServerTarget = ""
		pushed, err := pushCloud(ctx)
		if err != nil {
			return changed, fmt.Errorf("import saved locally, but Cloud push failed: %w", err)
		}
		fmt.Println(ui.Successf(fmt.Sprintf("OpenSSH to Cloud complete (%d imported, %d encrypted records pushed)", changed, pushed)))
	}
	return changed, nil
}

func mergeOpenSSHServer(existing, imported store.Server) store.Server {
	existing.Host = imported.Host
	existing.User = imported.User
	existing.Port = imported.Port
	existing.AuthType = imported.AuthType
	if imported.KeyPath != "" {
		existing.KeyPath = imported.KeyPath
	}
	existing.JumpHost = imported.JumpHost
	existing.LocalForwards = imported.LocalForwards
	existing.RemoteForwards = imported.RemoteForwards
	existing.X11Forward = imported.X11Forward
	if imported.KeepAlive > 0 {
		existing.KeepAlive = imported.KeepAlive
	}
	if imported.ConnectionTimeout > 0 {
		existing.ConnectionTimeout = imported.ConnectionTimeout
	}
	if existing.Group == "" || existing.Group == "default" {
		existing.Group = imported.Group
	}
	if existing.CustomOptions == nil {
		existing.CustomOptions = map[string]string{}
	}
	for key, value := range imported.CustomOptions {
		existing.CustomOptions[key] = value
	}
	return existing
}

func vaultServerIndexByName(vault *store.Vault, name string) int {
	for index := range vault.Servers {
		if strings.EqualFold(vault.Servers[index].Name, name) {
			return index
		}
	}
	return -1
}

func init() {
	for _, command := range []*cobra.Command{importOpenSSHCmd, cloudImportOpenSSHCmd} {
		command.Flags().StringVar(&openSSHConfigPath, "config", "", "OpenSSH config path (default ~/.ssh/config)")
		command.Flags().BoolVar(&openSSHReplace, "replace", false, "refresh matching server connection fields")
		command.Flags().BoolVar(&openSSHDryRun, "dry-run", false, "preview without changing the vault or Cloud")
		command.Flags().BoolVar(&openSSHIncludeKeys, "include-keys", false, "include private-key contents in the encrypted Cloud payload")
	}
	importOpenSSHCmd.Flags().BoolVar(&openSSHPush, "push", false, "push the imported inventory to the linked Cloud workspace")
	rootCmd.AddCommand(importOpenSSHCmd)
}

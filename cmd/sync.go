package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/johnniewhite/ssher/internal/transfer"
	"github.com/johnniewhite/ssher/internal/ui"
)

var (
	syncServer  string
	syncPull    bool
	syncDelete  bool
	syncDryRun  bool
	syncVerbose bool
)

var syncCmd = &cobra.Command{
	Use:   "sync <local> <remote>",
	Short: "Rsync between local and remote (key auth only)",
	Args:  cobra.ExactArgs(2),
	RunE: func(c *cobra.Command, args []string) error {
		if syncServer == "" {
			return fmt.Errorf("--server is required")
		}
		saved, err := ui.LoadVault()
		if err != nil {
			return err
		}
		target, _, err := saved.Vault.ResolveTarget(syncServer)
		if err != nil {
			return err
		}
		mode := transfer.RsyncPush
		if syncPull {
			mode = transfer.RsyncPull
		}
		opts := transfer.RsyncOptions{
			Mode:    mode,
			DryRun:  syncDryRun,
			Delete:  syncDelete,
			Verbose: syncVerbose,
		}
		fmt.Println(ui.Infof(fmt.Sprintf("rsync %s <-> %s:%s", args[0], target.Name, args[1])))
		return transfer.Rsync(target, args[0], args[1], opts)
	},
}

func init() {
	syncCmd.Flags().StringVarP(&syncServer, "server", "s", "", "target server")
	syncCmd.Flags().BoolVar(&syncPull, "pull", false, "remote -> local instead of local -> remote")
	syncCmd.Flags().BoolVar(&syncDelete, "delete", false, "delete files at destination not present at source")
	syncCmd.Flags().BoolVar(&syncDryRun, "dry-run", false, "show what would be transferred")
	syncCmd.Flags().BoolVarP(&syncVerbose, "verbose", "v", false, "rsync verbose output")
	rootCmd.AddCommand(syncCmd)
}

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	sshc "github.com/johnniewhite/ssher/internal/ssh"
	"github.com/johnniewhite/ssher/internal/transfer"
	"github.com/johnniewhite/ssher/internal/ui"
)

var downloadServer string

var downloadCmd = &cobra.Command{
	Use:   "download <remote> <local>",
	Short: "Download a file or directory via SFTP",
	Args:  cobra.ExactArgs(2),
	RunE: func(c *cobra.Command, args []string) error {
		if downloadServer == "" {
			return fmt.Errorf("--server is required")
		}
		saved, err := ui.LoadVault()
		if err != nil {
			return err
		}
		target, _, err := saved.Vault.ResolveTarget(downloadServer)
		if err != nil {
			return err
		}
		client, err := sshc.Dial(saved.Vault, target)
		if err != nil {
			return err
		}
		defer client.Close()

		fmt.Println(ui.Infof(fmt.Sprintf("downloading %s:%s -> %s", target.Name, args[0], args[1])))
		if err := transfer.Download(client, args[0], args[1]); err != nil {
			return err
		}
		fmt.Println(ui.Successf("download complete"))
		return nil
	},
}

func init() {
	downloadCmd.Flags().StringVarP(&downloadServer, "server", "s", "", "source server (name, alias, or index)")
	rootCmd.AddCommand(downloadCmd)
}

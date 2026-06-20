package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	sshc "github.com/johnniewhite/ssher/internal/ssh"
	"github.com/johnniewhite/ssher/internal/transfer"
	"github.com/johnniewhite/ssher/internal/ui"
)

var uploadServer string

var uploadCmd = &cobra.Command{
	Use:   "upload <local> <remote>",
	Short: "Upload a file or directory via SFTP",
	Args:  cobra.ExactArgs(2),
	RunE: func(c *cobra.Command, args []string) error {
		saved, err := ui.LoadVault()
		if err != nil {
			return err
		}
		target, _, err := saved.Vault.ResolveTarget(uploadServer)
		if err != nil {
			return err
		}
		client, err := sshc.Dial(saved.Vault, target)
		if err != nil {
			return err
		}
		defer client.Close()

		fmt.Println(ui.Infof(fmt.Sprintf("uploading %s -> %s:%s", args[0], target.Name, args[1])))
		if err := transfer.Upload(client, args[0], args[1]); err != nil {
			return err
		}
		fmt.Println(ui.Successf("upload complete"))
		return nil
	},
}

func init() {
	uploadCmd.Flags().StringVarP(&uploadServer, "server", "s", "", "target server (name, alias, or index)")
	_ = uploadCmd.MarkFlagRequired("server")
	rootCmd.AddCommand(uploadCmd)
}

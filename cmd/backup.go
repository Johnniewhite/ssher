package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/johnniewhite/ssher/internal/paths"
	"github.com/johnniewhite/ssher/internal/ui"
	"github.com/johnniewhite/ssher/internal/vault"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Snapshot the encrypted vault to ~/.ssher/backups",
	RunE: func(c *cobra.Command, args []string) error {
		exists, err := vault.Exists()
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("no vault to back up")
		}
		src, err := paths.VaultFile()
		if err != nil {
			return err
		}
		dir, err := paths.BackupsDir()
		if err != nil {
			return err
		}
		stamp := time.Now().UTC().Format("20060102-150405")
		dst := filepath.Join(dir, fmt.Sprintf("vault-%s.bin", stamp))
		if err := copyFile(src, dst); err != nil {
			return err
		}
		fmt.Println(ui.Successf(fmt.Sprintf("backed up to %s", dst)))
		return nil
	},
}

func init() { rootCmd.AddCommand(backupCmd) }

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, paths.FileMode)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

package cmd

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/johnniewhite/ssher/internal/store"
	"github.com/johnniewhite/ssher/internal/ui"
	"github.com/johnniewhite/ssher/internal/vault"
)

var vaultCmd = &cobra.Command{
	Use:   "vault",
	Short: "Manage the encrypted vault",
}

var vaultUnlockCmd = &cobra.Command{
	Use:   "unlock",
	Short: "Authenticate and create a session",
	RunE: func(c *cobra.Command, args []string) error {
		exists, err := vault.Exists()
		if err != nil {
			return err
		}
		if !exists {
			return errors.New("no vault yet -- run `ssher add` to create one")
		}
		if _, err := ui.LoadVault(); err != nil {
			return err
		}
		fmt.Println(ui.Successf("vault unlocked"))
		return nil
	},
}

var vaultLockCmd = &cobra.Command{
	Use:   "lock",
	Short: "Clear the active session",
	RunE: func(c *cobra.Command, args []string) error {
		if err := vault.ClearSession(); err != nil {
			return err
		}
		fmt.Println(ui.Successf("vault locked"))
		return nil
	},
}

var vaultStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show vault state",
	RunE: func(c *cobra.Command, args []string) error {
		exists, err := vault.Exists()
		if err != nil {
			return err
		}
		fmt.Println(ui.Title.Render("Vault status"))
		if !exists {
			fmt.Println(ui.Muted.Render("  vault file:  not yet created"))
			return nil
		}
		fmt.Println(ui.Muted.Render("  vault file:  present"))

		active, remaining, err := vault.SessionInfo()
		if err != nil {
			return err
		}
		if active {
			rem := remaining.Round(time.Second)
			fmt.Printf("  session:     %s (%s remaining)\n",
				ui.Success.Render("unlocked"), rem)
		} else {
			fmt.Printf("  session:     %s\n", ui.Warn.Render("locked"))
		}
		fmt.Printf("  timeout:     %s\n", vault.SessionTimeout)
		return nil
	},
}

var vaultChangePasswordCmd = &cobra.Command{
	Use:   "change-password",
	Short: "Re-encrypt the vault with a new master password",
	RunE: func(c *cobra.Command, args []string) error {
		// Require the *current* password explicitly. Even if a session is
		// active we want fresh proof of intent before re-keying the vault.
		oldPw, err := ui.PromptPassword("Current master password: ")
		if err != nil {
			return err
		}
		saved, err := store.LoadWithPassword([]byte(oldPw))
		if err != nil {
			if errors.Is(err, vault.ErrBadPassword) {
				return errors.New("incorrect master password")
			}
			return err
		}

		newPw, err := ui.PromptNewPassword("New master password")
		if err != nil {
			return err
		}

		// Re-marshal and re-encrypt with fresh salt+key.
		fresh, err := store.InitialiseEmpty([]byte(newPw))
		if err != nil {
			return err
		}
		// Carry servers/profiles/aliases/history forward.
		fresh.Vault.Servers = saved.Vault.Servers
		fresh.Vault.Profiles = saved.Vault.Profiles
		fresh.Vault.Aliases = saved.Vault.Aliases
		fresh.Vault.History = saved.Vault.History
		if err := fresh.SaveAndRefreshSession(); err != nil {
			return err
		}
		fmt.Println(ui.Successf("master password changed"))
		return nil
	},
}

func init() {
	vaultCmd.AddCommand(vaultUnlockCmd, vaultLockCmd, vaultStatusCmd, vaultChangePasswordCmd)
	rootCmd.AddCommand(vaultCmd)
}

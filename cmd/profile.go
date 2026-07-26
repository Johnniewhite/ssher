package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/johnniewhite/ssher/internal/store"
	"github.com/johnniewhite/ssher/internal/ui"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage reusable connection profiles",
}

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List profiles",
	RunE: func(c *cobra.Command, args []string) error {
		saved, err := ui.LoadVault()
		if err != nil {
			return err
		}
		if len(saved.Vault.Profiles) == 0 {
			fmt.Println(ui.Muted.Render("no profiles yet"))
			return nil
		}
		for _, p := range saved.Vault.Profiles {
			fmt.Printf("%s  timeout=%d keepalive=%d x11(export-only)=%v reconnect=%v\n",
				ui.Title.Render(p.Name),
				p.ConnectionTimeout, p.KeepAlive, p.X11Forward, p.AutoReconnect)
		}
		return nil
	},
}

var (
	profileAddTimeout   int
	profileAddKeepAlive int
	profileAddX11       bool
	profileAddReconnect bool
	profileAddRetries   int
)

var profileAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Create a profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		saved, err := ui.LoadVault()
		if err != nil {
			return err
		}
		name := args[0]
		for _, p := range saved.Vault.Profiles {
			if p.Name == name {
				return fmt.Errorf("profile %q already exists", name)
			}
		}
		saved.Vault.Profiles = append(saved.Vault.Profiles, store.Profile{
			Name:                name,
			ConnectionTimeout:   profileAddTimeout,
			KeepAlive:           profileAddKeepAlive,
			X11Forward:          profileAddX11,
			AutoReconnect:       profileAddReconnect,
			MaxReconnectRetries: profileAddRetries,
		})
		if err := saved.SaveAndRefreshSession(); err != nil {
			return err
		}
		fmt.Println(ui.Successf(fmt.Sprintf("profile %q created", name)))
		return nil
	},
}

var profileRemoveCmd = &cobra.Command{
	Use:     "remove <name>",
	Aliases: []string{"rm", "delete"},
	Short:   "Remove a profile",
	Args:    cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		saved, err := ui.LoadVault()
		if err != nil {
			return err
		}
		name := args[0]
		filtered := saved.Vault.Profiles[:0]
		found := false
		for _, p := range saved.Vault.Profiles {
			if p.Name == name {
				found = true
				continue
			}
			filtered = append(filtered, p)
		}
		if !found {
			return fmt.Errorf("profile %q not found", name)
		}
		saved.Vault.Profiles = filtered
		if err := saved.SaveAndRefreshSession(); err != nil {
			return err
		}
		fmt.Println(ui.Successf(fmt.Sprintf("profile %q removed", name)))
		return nil
	},
}

var (
	profileApplyServer string
)

var profileApplyCmd = &cobra.Command{
	Use:   "apply <profile>",
	Short: "Apply a profile to a server",
	Args:  cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		saved, err := ui.LoadVault()
		if err != nil {
			return err
		}
		profileName := args[0]
		var profile *store.Profile
		for i := range saved.Vault.Profiles {
			if saved.Vault.Profiles[i].Name == profileName {
				profile = &saved.Vault.Profiles[i]
				break
			}
		}
		if profile == nil {
			return fmt.Errorf("profile %q not found", profileName)
		}
		target, _, err := saved.Vault.ResolveTarget(profileApplyServer)
		if err != nil {
			return err
		}
		if err := saved.Vault.UpdateServer(target.Name, func(s *store.Server) {
			s.Profile = profile.Name
			if profile.ConnectionTimeout > 0 {
				s.ConnectionTimeout = profile.ConnectionTimeout
			}
			if profile.KeepAlive > 0 {
				s.KeepAlive = profile.KeepAlive
			}
			s.X11Forward = profile.X11Forward
			s.AutoReconnect = profile.AutoReconnect
			if profile.MaxReconnectRetries > 0 {
				s.MaxReconnectRetries = profile.MaxReconnectRetries
			}
			if len(profile.CustomOptions) > 0 {
				if s.CustomOptions == nil {
					s.CustomOptions = map[string]string{}
				}
				for k, v := range profile.CustomOptions {
					s.CustomOptions[k] = v
				}
			}
		}); err != nil {
			return err
		}
		if err := saved.SaveAndRefreshSession(); err != nil {
			return err
		}
		fmt.Println(ui.Successf(fmt.Sprintf("applied profile %q to %s", profileName, target.Name)))
		return nil
	},
}

func init() {
	profileAddCmd.Flags().IntVar(&profileAddTimeout, "timeout", 30, "connection timeout in seconds")
	profileAddCmd.Flags().IntVar(&profileAddKeepAlive, "keepalive", 60, "keepalive in seconds (0 to disable)")
	profileAddCmd.Flags().BoolVar(&profileAddX11, "x11", false, "enable X11 forwarding")
	_ = profileAddCmd.Flags().MarkDeprecated("x11", "native X11 forwarding is unsupported; this setting is export-config-only")
	profileAddCmd.Flags().BoolVar(&profileAddReconnect, "reconnect", false, "auto-reconnect by default")
	profileAddCmd.Flags().IntVar(&profileAddRetries, "max-retries", 3, "max reconnect attempts")

	profileApplyCmd.Flags().StringVarP(&profileApplyServer, "server", "s", "", "target server")
	_ = profileApplyCmd.MarkFlagRequired("server")

	profileCmd.AddCommand(profileListCmd, profileAddCmd, profileRemoveCmd, profileApplyCmd)
	rootCmd.AddCommand(profileCmd)
}

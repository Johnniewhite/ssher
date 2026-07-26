package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/johnniewhite/ssher/internal/store"
	"github.com/johnniewhite/ssher/internal/ui"
)

var editCmd = &cobra.Command{
	Use:   "edit [target]",
	Short: "Edit a saved server",
	Args:  cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		saved, err := ui.LoadVault()
		if err != nil {
			return err
		}
		target, _, err := saved.Vault.ResolveTarget(args[0])
		if err != nil {
			return err
		}
		updated, err := promptEditServer(*target)
		if err != nil {
			return err
		}
		if err := normalizeJumpHost(saved.Vault, target.Name, &updated); err != nil {
			return err
		}
		if err := saved.Vault.UpdateServer(target.Name, func(s *store.Server) {
			*s = updated
		}); err != nil {
			return err
		}
		if err := saved.SaveAndRefreshSession(); err != nil {
			return err
		}
		fmt.Println(ui.Successf(fmt.Sprintf("updated %s", updated.Name)))
		return nil
	},
}

func normalizeJumpHost(v *store.Vault, serverName string, server *store.Server) error {
	if server.JumpHost == "" {
		return nil
	}
	jump, _, err := v.ResolveTarget(server.JumpHost)
	if err != nil {
		return fmt.Errorf("jump host: %w", err)
	}
	if jump.Name == serverName {
		return fmt.Errorf("a server cannot use itself as a jump host")
	}
	server.JumpHost = jump.Name
	return nil
}

func init() {
	rootCmd.AddCommand(editCmd)
}

func promptEditServer(s store.Server) (store.Server, error) {
	portStr := strconv.Itoa(s.Port)
	tagsStr := joinCSV(s.Tags)
	localForwards := formatForwards(s.LocalForwards)
	remoteForwards := formatForwards(s.RemoteForwards)
	authChoice := string(s.AuthType)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Host or IP").Value(&s.Host).Validate(nonEmpty("host")),
			huh.NewInput().Title("User").Value(&s.User).Validate(nonEmpty("user")),
			huh.NewInput().Title("Port").Value(&portStr).Validate(validPort),
			huh.NewSelect[string]().
				Title("Authentication").
				Options(
					huh.NewOption("SSH key", "key"),
					huh.NewOption("Password", "password"),
				).
				Value(&authChoice),
			huh.NewInput().Title("Group").Value(&s.Group),
			huh.NewInput().Title("Tags (comma-separated)").Value(&tagsStr),
			huh.NewInput().Title("Notes").Value(&s.Notes),
			huh.NewInput().Title("Jump host").Value(&s.JumpHost),
			huh.NewInput().Title("Local forwards").
				Description("local:host:remote, comma-separated").
				Value(&localForwards),
			huh.NewInput().Title("Remote forwards").
				Description("remote:local-host:local-port, comma-separated").
				Value(&remoteForwards),
			huh.NewConfirm().Title("Favorite?").Value(&s.IsFavorite),
		),
	)
	if err := form.Run(); err != nil {
		return s, err
	}
	port, _ := strconv.Atoi(portStr)
	s.Port = port
	s.AuthType = store.AuthType(authChoice)
	s.Tags = splitCSV(tagsStr)
	var err error
	if s.LocalForwards, err = parseForwards(localForwards); err != nil {
		return s, fmt.Errorf("local forwards: %w", err)
	}
	if s.RemoteForwards, err = parseForwards(remoteForwards); err != nil {
		return s, fmt.Errorf("remote forwards: %w", err)
	}

	switch s.AuthType {
	case store.AuthPassword:
		pwForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().Title("Password").
					EchoMode(huh.EchoModePassword).
					Value(&s.Password).Validate(nonEmpty("password")),
			),
		)
		if err := pwForm.Run(); err != nil {
			return s, err
		}
		s.KeyPath = ""
	case store.AuthKey:
		keyForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().Title("Key path").
					Description("Leave blank to try ssh-agent and standard key files").
					Value(&s.KeyPath),
			),
		)
		if err := keyForm.Run(); err != nil {
			return s, err
		}
		s.Password = ""
	}
	return s, nil
}

func joinCSV(xs []string) string {
	return strings.Join(xs, ",")
}

func formatForwards(forwards []store.PortForward) string {
	parts := make([]string, 0, len(forwards))
	for _, fwd := range forwards {
		parts = append(parts, fmt.Sprintf("%d:%s:%d", fwd.LocalPort, fwd.RemoteHost, fwd.RemotePort))
	}
	return strings.Join(parts, ",")
}

package cmd

import (
	"fmt"
	"strconv"

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

func init() {
	rootCmd.AddCommand(editCmd)
}

func promptEditServer(s store.Server) (store.Server, error) {
	portStr := strconv.Itoa(s.Port)
	tagsStr := joinCSV(s.Tags)
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
			huh.NewConfirm().Title("X11 forwarding?").Value(&s.X11Forward),
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

	if s.AuthType == store.AuthPassword && s.Password == "" {
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
	}
	return s, nil
}

func joinCSV(xs []string) string {
	out := ""
	for i, x := range xs {
		if i > 0 {
			out += ","
		}
		out += x
	}
	return out
}

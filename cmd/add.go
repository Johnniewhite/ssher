package cmd

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/johnniewhite/ssher/internal/store"
	"github.com/johnniewhite/ssher/internal/ui"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new server (interactive)",
	RunE: func(c *cobra.Command, args []string) error {
		saved, err := ui.LoadVault()
		if err != nil {
			return err
		}

		s, err := promptNewServer(saved.Vault)
		if err != nil {
			return err
		}
		if err := saved.Vault.AddServer(s); err != nil {
			return err
		}
		if err := saved.SaveAndRefreshSession(); err != nil {
			return err
		}
		fmt.Println(ui.Successf(fmt.Sprintf("added %s (%s@%s:%d)", s.Name, s.User, s.Host, s.Port)))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}

func promptNewServer(v *store.Vault) (store.Server, error) {
	var s store.Server
	var portStr = "22"
	var tagsStr, forwardsLocalStr, forwardsRemoteStr string
	var authChoice string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Name").Value(&s.Name).
				Validate(uniqueServerName(v)),
			huh.NewInput().Title("Host or IP").Value(&s.Host).
				Validate(nonEmpty("host")),
			huh.NewInput().Title("User").Value(&s.User).
				Validate(nonEmpty("user")),
			huh.NewInput().Title("Port").Value(&portStr).
				Validate(validPort),
			huh.NewSelect[string]().
				Title("Authentication").
				Options(
					huh.NewOption("SSH key", "key"),
					huh.NewOption("Password", "password"),
				).
				Value(&authChoice),
		),
	)
	if err := form.Run(); err != nil {
		return s, err
	}
	port, _ := strconv.Atoi(portStr)
	s.Port = port
	s.AuthType = store.AuthType(authChoice)

	switch s.AuthType {
	case store.AuthKey:
		keyForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().Title("Key path").
					Description("Leave blank for ~/.ssh/id_rsa").
					Value(&s.KeyPath),
			),
		)
		if err := keyForm.Run(); err != nil {
			return s, err
		}
	case store.AuthPassword:
		pwForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().Title("Password").
					EchoMode(huh.EchoModePassword).
					Value(&s.Password).
					Validate(nonEmpty("password")),
			),
		)
		if err := pwForm.Run(); err != nil {
			return s, err
		}
	}

	extras := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Group").Description("Leave blank for 'default'").Value(&s.Group),
			huh.NewInput().Title("Tags (comma-separated)").Value(&tagsStr),
			huh.NewInput().Title("Notes").Value(&s.Notes),
			huh.NewInput().Title("Jump host (server name)").Value(&s.JumpHost),
			huh.NewInput().Title("Local forwards").
				Description("e.g. 8080:127.0.0.1:80,5432:db:5432 — blank for none").
				Value(&forwardsLocalStr),
			huh.NewInput().Title("Remote forwards").
				Description("Same syntax — blank for none").
				Value(&forwardsRemoteStr),
			huh.NewConfirm().Title("Mark as favorite?").Value(&s.IsFavorite),
		),
	)
	if err := extras.Run(); err != nil {
		return s, err
	}

	if s.Group == "" {
		s.Group = "default"
	}
	if s.JumpHost != "" {
		jump, _, err := v.ResolveTarget(s.JumpHost)
		if err != nil {
			return s, fmt.Errorf("jump host: %w", err)
		}
		if jump.Name == s.Name {
			return s, errors.New("a server cannot use itself as a jump host")
		}
		s.JumpHost = jump.Name
	}
	if t := splitCSV(tagsStr); len(t) > 0 {
		s.Tags = t
	}
	if fw, err := parseForwards(forwardsLocalStr); err != nil {
		return s, fmt.Errorf("local forwards: %w", err)
	} else {
		s.LocalForwards = fw
	}
	if fw, err := parseForwards(forwardsRemoteStr); err != nil {
		return s, fmt.Errorf("remote forwards: %w", err)
	} else {
		s.RemoteForwards = fw
	}
	return s, nil
}

func uniqueServerName(v *store.Vault) func(string) error {
	return func(s string) error {
		if strings.TrimSpace(s) == "" {
			return errors.New("name is required")
		}
		for _, existing := range v.Servers {
			if existing.Name == s {
				return fmt.Errorf("server %q already exists", s)
			}
		}
		return nil
	}
}

func nonEmpty(field string) func(string) error {
	return func(s string) error {
		if strings.TrimSpace(s) == "" {
			return fmt.Errorf("%s is required", field)
		}
		return nil
	}
}

func validPort(s string) error {
	n, err := strconv.Atoi(s)
	if err != nil {
		return errors.New("port must be a number")
	}
	if n < 1 || n > 65535 {
		return errors.New("port must be 1..65535")
	}
	return nil
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// parseForwards accepts forms like "8080:127.0.0.1:80" or "8080:80" (host
// defaulting to 127.0.0.1), comma-separated.
func parseForwards(s string) ([]store.PortForward, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	parts := strings.Split(s, ",")
	out := make([]store.PortForward, 0, len(parts))
	for _, raw := range parts {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		segs := strings.Split(raw, ":")
		var pf store.PortForward
		switch len(segs) {
		case 2:
			lp, err := parseForwardPort(segs[0])
			if err != nil {
				return nil, fmt.Errorf("bad local port in %q", raw)
			}
			rp, err := parseForwardPort(segs[1])
			if err != nil {
				return nil, fmt.Errorf("bad remote port in %q", raw)
			}
			pf = store.PortForward{LocalPort: lp, RemoteHost: "127.0.0.1", RemotePort: rp}
		case 3:
			lp, err := parseForwardPort(segs[0])
			if err != nil {
				return nil, fmt.Errorf("bad local port in %q", raw)
			}
			if strings.TrimSpace(segs[1]) == "" {
				return nil, fmt.Errorf("empty remote host in %q", raw)
			}
			rp, err := parseForwardPort(segs[2])
			if err != nil {
				return nil, fmt.Errorf("bad remote port in %q", raw)
			}
			pf = store.PortForward{LocalPort: lp, RemoteHost: segs[1], RemotePort: rp}
		default:
			return nil, fmt.Errorf("forward %q must be local:remote or local:host:remote", raw)
		}
		out = append(out, pf)
	}
	return out, nil
}

func parseForwardPort(s string) (int, error) {
	port, err := strconv.Atoi(s)
	if err != nil || port < 1 || port > 65535 {
		return 0, errors.New("port must be 1..65535")
	}
	return port, nil
}

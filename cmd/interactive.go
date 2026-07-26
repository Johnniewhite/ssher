package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	sshc "github.com/johnniewhite/ssher/internal/ssh"
	"github.com/johnniewhite/ssher/internal/store"
	"github.com/johnniewhite/ssher/internal/transfer"
	"github.com/johnniewhite/ssher/internal/ui"
)

// runInteractive is the default `ssher` (no-args) experience: print the
// server list and dispatch from a menu. We deliberately keep this simple
// rather than building a full bubbletea TUI — the muscle memory from the
// Python tool is "type a number or a letter command".
func runInteractive(c *cobra.Command, args []string) error {
	saved, err := ui.LoadVault()
	if err != nil {
		return err
	}

	for {
		clearScreen()
		printHeader()
		fmt.Println(ui.Title.Render(fmt.Sprintf("Saved servers (%d)", len(saved.Vault.Servers))))
		fmt.Println(ui.RenderServerTable(saved.Vault.SortedServers()))
		fmt.Println()
		fmt.Println(ui.Muted.Render("[number] connect    [a] add    [e] edit    [d] delete    [f] toggle favorite"))
		fmt.Println(ui.Muted.Render("[g] groups         [s] search [t] transfer [x] exec       [p] ping all"))
		fmt.Println(ui.Muted.Render("[c] copy           [v] vault  [h] history  [q] quit       [?] help"))
		fmt.Println()

		var input string
		if err := huh.NewInput().
			Title(">").
			Value(&input).
			Run(); err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				return nil
			}
			return err
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Number → connect by index.
		if isAllDigits(input) {
			if err := runConnect(c, []string{input}); err != nil {
				printErr(err)
				pause()
			}
			continue
		}

		switch strings.ToLower(input) {
		case "q", "quit", "exit":
			return nil
		case "a", "add":
			err := runAddInteractive(saved)
			if err != nil {
				printErr(err)
				pause()
			}
			// Reload — add() saved already.
			saved, err = ui.LoadVault()
			if err != nil {
				return err
			}
		case "e", "edit":
			if err := handleEditFlow(saved); err != nil {
				printErr(err)
				pause()
			}
			saved, err = ui.LoadVault()
			if err != nil {
				return err
			}
		case "d", "del", "delete":
			if err := handleDeleteFlow(saved); err != nil {
				printErr(err)
				pause()
			}
			saved, err = ui.LoadVault()
			if err != nil {
				return err
			}
		case "f", "fav":
			if err := handleFavoriteFlow(saved); err != nil {
				printErr(err)
				pause()
			}
			saved, err = ui.LoadVault()
			if err != nil {
				return err
			}
		case "g", "groups":
			if err := groupsCmd.RunE(c, nil); err != nil {
				printErr(err)
			}
			pause()
		case "s", "search":
			if err := handleSearchFlow(saved); err != nil {
				printErr(err)
				pause()
			}
		case "x", "exec":
			if err := handleExecFlow(saved); err != nil {
				printErr(err)
				pause()
			}
		case "p", "ping":
			if err := pingCmd.RunE(c, nil); err != nil {
				printErr(err)
			}
			pause()
		case "c", "copy":
			if err := handleCopyFlow(saved); err != nil {
				printErr(err)
				pause()
			}
		case "v", "vault":
			if err := vaultStatusCmd.RunE(c, nil); err != nil {
				printErr(err)
			}
			pause()
		case "h", "history":
			if err := historyCmd.RunE(c, nil); err != nil {
				printErr(err)
			}
			pause()
		case "?", "help":
			printHelpReference()
			pause()
		case "t", "transfer":
			handleTransferFlow(saved)
		default:
			// Try as fuzzy server target.
			if err := runConnect(c, []string{input}); err != nil {
				printErr(err)
				pause()
			}
		}
	}
}

func runAddInteractive(saved *store.Saved) error {
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
	fmt.Println(ui.Successf(fmt.Sprintf("added %s", s.Name)))
	pause()
	return nil
}

func handleEditFlow(saved *store.Saved) error {
	target, err := pickServer(saved, "Edit which server?")
	if err != nil || target == nil {
		return err
	}
	updated, err := promptEditServer(*target)
	if err != nil {
		return err
	}
	if err := normalizeJumpHost(saved.Vault, target.Name, &updated); err != nil {
		return err
	}
	if err := saved.Vault.UpdateServer(target.Name, func(s *store.Server) { *s = updated }); err != nil {
		return err
	}
	if err := saved.SaveAndRefreshSession(); err != nil {
		return err
	}
	fmt.Println(ui.Successf("updated"))
	pause()
	return nil
}

func handleDeleteFlow(saved *store.Saved) error {
	target, err := pickServer(saved, "Delete which server?")
	if err != nil || target == nil {
		return err
	}
	var ok bool
	if err := huh.NewConfirm().
		Title(fmt.Sprintf("Delete %s?", target.Name)).
		Affirmative("Delete").
		Negative("Cancel").
		Value(&ok).
		Run(); err != nil {
		return err
	}
	if !ok {
		return nil
	}
	if err := saved.Vault.DeleteServer(target.Name); err != nil {
		return err
	}
	if err := saved.SaveAndRefreshSession(); err != nil {
		return err
	}
	fmt.Println(ui.Successf("deleted"))
	pause()
	return nil
}

func handleFavoriteFlow(saved *store.Saved) error {
	target, err := pickServer(saved, "Toggle favorite on which server?")
	if err != nil || target == nil {
		return err
	}
	if err := saved.Vault.UpdateServer(target.Name, func(s *store.Server) { s.IsFavorite = !s.IsFavorite }); err != nil {
		return err
	}
	if err := saved.SaveAndRefreshSession(); err != nil {
		return err
	}
	pause()
	return nil
}

func handleSearchFlow(saved *store.Saved) error {
	var query string
	if err := huh.NewInput().Title("Search").Value(&query).Run(); err != nil {
		return err
	}
	if query == "" {
		return nil
	}
	view := filterServers(saved.Vault.SortedServers(), saved.Vault.Aliases, "", false, query)
	fmt.Println(ui.RenderServerTable(view))
	pause()
	return nil
}

func handleExecFlow(saved *store.Saved) error {
	var cmdline string
	if err := huh.NewInput().Title("Command to run").Value(&cmdline).Run(); err != nil {
		return err
	}
	if cmdline == "" {
		return nil
	}
	results := runOnAll(saved.Vault, saved.Vault.SortedServers(), cmdline, defaultExecParallel, defaultExecTimeout)
	_ = printExecResults(results)
	pause()
	return nil
}

func handleTransferFlow(saved *store.Saved) {
	target, err := pickServer(saved, "Transfer to/from which server?")
	if err != nil || target == nil {
		return
	}

	direction := "upload"
	if err := huh.NewSelect[string]().
		Title("Direction").
		Options(
			huh.NewOption("upload (local → remote)", "upload"),
			huh.NewOption("download (remote → local)", "download"),
		).
		Value(&direction).
		Run(); err != nil {
		return
	}

	var local, remote string
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Local path").Value(&local),
			huh.NewInput().Title("Remote path").Value(&remote),
		),
	).Run(); err != nil {
		return
	}
	if strings.TrimSpace(local) == "" || strings.TrimSpace(remote) == "" {
		fmt.Println(ui.Warnf("both local and remote paths are required"))
		pause()
		return
	}

	client, err := sshc.Dial(saved.Vault, target)
	if err != nil {
		printErr(err)
		pause()
		return
	}
	defer client.Close()

	if direction == "upload" {
		fmt.Println(ui.Infof(fmt.Sprintf("uploading %s -> %s:%s", local, target.Name, remote)))
		err = transfer.Upload(client, local, remote)
	} else {
		fmt.Println(ui.Infof(fmt.Sprintf("downloading %s:%s -> %s", target.Name, remote, local)))
		err = transfer.Download(client, remote, local)
	}
	if err != nil {
		printErr(err)
	} else {
		fmt.Println(ui.Successf("transfer complete"))
	}
	pause()
}

func handleCopyFlow(saved *store.Saved) error {
	target, err := pickServer(saved, "Copy details from which server?")
	if err != nil || target == nil {
		return err
	}
	field := "command"
	if err := huh.NewSelect[string]().
		Title("Field").
		Options(
			huh.NewOption("ssh command", "command"),
			huh.NewOption("host", "host"),
			huh.NewOption("user", "user"),
			huh.NewOption("password", "password"),
		).Value(&field).Run(); err != nil {
		return err
	}
	copyField = field
	if err := copyCmd.RunE(nil, []string{target.Name}); err != nil {
		return err
	}
	pause()
	return nil
}

func pickServer(saved *store.Saved, title string) (*store.Server, error) {
	servers := saved.Vault.SortedServers()
	if len(servers) == 0 {
		fmt.Println(ui.Muted.Render("no servers yet"))
		pause()
		return nil, nil
	}
	options := make([]huh.Option[string], 0, len(servers))
	for _, s := range servers {
		label := fmt.Sprintf("%s — %s@%s", s.Name, s.User, s.Host)
		options = append(options, huh.NewOption(label, s.Name))
	}
	var pick string
	if err := huh.NewSelect[string]().
		Title(title).
		Options(options...).
		Value(&pick).
		Run(); err != nil {
		return nil, err
	}
	for i := range saved.Vault.Servers {
		if saved.Vault.Servers[i].Name == pick {
			return &saved.Vault.Servers[i], nil
		}
	}
	return nil, fmt.Errorf("server %q vanished", pick)
}

func clearScreen() {
	// ANSI: clear screen + move cursor home. Cheap and works everywhere
	// the rest of our lipgloss output works.
	fmt.Print("\033[2J\033[H")
}

func printHeader() {
	fmt.Println(ui.Title.Render("ssher ") + ui.Muted.Render(Version))
	fmt.Println()
}

func printHelpReference() {
	fmt.Println(ui.Heading.Render("Commands"))
	fmt.Println("  [number]   connect to server N in the list")
	fmt.Println("  [name]     connect by name (also exact alias)")
	fmt.Println("  a / add    add a new server")
	fmt.Println("  e / edit   edit a server")
	fmt.Println("  d / del    delete a server")
	fmt.Println("  f          toggle favorite on a server")
	fmt.Println("  g          list groups")
	fmt.Println("  s          search servers")
	fmt.Println("  t          transfer files (upload/download)")
	fmt.Println("  x          run a command on all servers")
	fmt.Println("  p          TCP-ping all servers")
	fmt.Println("  c          copy server details to clipboard")
	fmt.Println("  v          show vault status")
	fmt.Println("  h          connection history")
	fmt.Println("  q          quit")
}

func pause() {
	fmt.Println()
	fmt.Println(ui.Muted.Render("press enter to continue"))
	var s string
	_, _ = fmt.Scanln(&s)
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

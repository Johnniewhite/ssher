package cmd

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	sshc "github.com/johnniewhite/ssher/internal/ssh"
	"github.com/johnniewhite/ssher/internal/store"
	"github.com/johnniewhite/ssher/internal/ui"
)

var (
	execAll      bool
	execGroup    string
	execServers  []string
	execParallel int
)

var execCmd = &cobra.Command{
	Use:   "exec <command>",
	Short: "Run a command on multiple servers in parallel",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		saved, err := ui.LoadVault()
		if err != nil {
			return err
		}
		targets, err := selectExecTargets(saved.Vault)
		if err != nil {
			return err
		}
		if len(targets) == 0 {
			return fmt.Errorf("no servers selected (use --all, --group, or --servers)")
		}

		cmdline := strings.Join(args, " ")
		fmt.Println(ui.Infof(fmt.Sprintf("running on %d server(s): %s", len(targets), cmdline)))

		results := runOnAll(saved.Vault, targets, cmdline, execParallel)
		printExecResults(results)
		return nil
	},
}

func init() {
	execCmd.Flags().BoolVar(&execAll, "all", false, "run on all servers")
	execCmd.Flags().StringVarP(&execGroup, "group", "g", "", "run on servers in this group")
	execCmd.Flags().StringSliceVarP(&execServers, "servers", "s", nil, "comma-separated server names/indices")
	execCmd.Flags().IntVarP(&execParallel, "parallel", "p", 8, "max concurrent connections")
	rootCmd.AddCommand(execCmd)
}

func selectExecTargets(v *store.Vault) ([]store.Server, error) {
	switch {
	case execAll:
		return v.SortedServers(), nil
	case execGroup != "":
		return v.ServersInGroup(execGroup), nil
	case len(execServers) > 0:
		out := make([]store.Server, 0, len(execServers))
		for _, raw := range execServers {
			s, _, err := v.ResolveTarget(strings.TrimSpace(raw))
			if err != nil {
				return nil, err
			}
			out = append(out, *s)
		}
		return out, nil
	default:
		return nil, nil
	}
}

type execResult struct {
	Server  store.Server
	Output  string
	Exit    int
	Err     error
	Elapsed time.Duration
}

func runOnAll(v *store.Vault, targets []store.Server, cmdline string, parallelism int) []execResult {
	if parallelism <= 0 {
		parallelism = 4
	}
	sem := make(chan struct{}, parallelism)
	results := make([]execResult, len(targets))
	var wg sync.WaitGroup
	for i, t := range targets {
		wg.Add(1)
		t := t
		i := i
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			start := time.Now()
			res := execResult{Server: t}
			defer func() { res.Elapsed = time.Since(start); results[i] = res }()
			client, err := sshc.Dial(v, &t)
			if err != nil {
				res.Err = err
				return
			}
			defer client.Close()
			out, exit, err := sshc.Run(client, cmdline)
			res.Output = out
			res.Exit = exit
			res.Err = err
		}()
	}
	wg.Wait()
	return results
}

func printExecResults(results []execResult) {
	var ok, failed int
	for _, r := range results {
		header := fmt.Sprintf("[%s] %s@%s (%s)",
			r.Server.Name, r.Server.User, r.Server.Host, r.Elapsed.Round(time.Millisecond))
		switch {
		case r.Err != nil:
			fmt.Println(ui.ErrorSty.Render(header))
			fmt.Println(ui.Muted.Render("  " + r.Err.Error()))
			failed++
		case r.Exit != 0:
			fmt.Println(ui.Warn.Render(fmt.Sprintf("%s [exit %d]", header, r.Exit)))
			fmt.Println(indent(r.Output, "  "))
			failed++
		default:
			fmt.Println(ui.Success.Render(header))
			fmt.Println(indent(r.Output, "  "))
			ok++
		}
	}
	fmt.Println()
	fmt.Println(ui.Heading.Render(fmt.Sprintf("%d ok / %d failed", ok, failed)))
}

func indent(s, prefix string) string {
	if s == "" {
		return ""
	}
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	for i, l := range lines {
		lines[i] = prefix + l
	}
	return strings.Join(lines, "\n")
}

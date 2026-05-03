// Package cmd contains the cobra command tree for ssher. main.go calls
// Execute and that's the entire entry-point machinery.
//
// Each command lives in its own file (add.go, list.go, ...). Commands that
// need vault access call ui.LoadVault() in their RunE -- this handles
// session caching and first-time setup transparently.
package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

const Version = "0.1.0-dev"

var rootCmd = &cobra.Command{
	Use:           "ssher [target]",
	Short:         "Ultimate SSH Configuration Manager",
	Long:          "ssher manages SSH server configurations in an encrypted vault and provides a drop-in sshpass replacement.",
	Version:       Version,
	SilenceUsage:  true,
	SilenceErrors: true,
	// `ssher`         -> interactive mode
	// `ssher prod`    -> connect to server "prod"
	// `ssher 3`       -> connect to server #3 in the sorted list
	Args: cobra.ArbitraryArgs,
	RunE: func(c *cobra.Command, args []string) error {
		switch len(args) {
		case 0:
			return runInteractive(c, args)
		case 1:
			// Pure-positional fall-through to `connect`. Matches Python's
			// `ssher <name-or-number>` behaviour.
			return runConnect(c, args)
		default:
			// Multiple positional args with no subcommand is almost certainly
			// a typo. Show usage rather than guessing.
			return fmt.Errorf("unknown command: %v\n\nrun 'ssher --help' for usage", args)
		}
	},
}

func Execute() error {
	return rootCmd.Execute()
}

// IsLikelyServerTarget is a small heuristic used by RunE dispatchers: a bare
// argument that's a positive integer or doesn't start with "-" is treated as
// a server target.
func IsLikelyServerTarget(s string) bool {
	if s == "" || s[0] == '-' {
		return false
	}
	if _, err := strconv.Atoi(s); err == nil {
		return true
	}
	return true
}

// printErr renders an error to stderr through ui styling. Used by command
// implementations that want to short-circuit cobra's default handling.
func printErr(err error) {
	fmt.Fprintln(os.Stderr, err)
}

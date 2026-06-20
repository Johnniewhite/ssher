package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/johnniewhite/ssher/internal/clipboard"
	"github.com/johnniewhite/ssher/internal/pwgen"
	"github.com/johnniewhite/ssher/internal/ui"
)

var (
	genLen        int
	genNoUpper    bool
	genNoLower    bool
	genNoDigits   bool
	genNoSymbols  bool
	genExclude    string
	genCount      int
	genCopy       bool
	genClearAfter time.Duration
)

var genpassCmd = &cobra.Command{
	Use:     "generate-password",
	Aliases: []string{"genpass"},
	Short:   "Generate a cryptographically secure password",
	RunE: func(c *cobra.Command, args []string) error {
		opts := pwgen.Options{
			Length:        genLen,
			IncludeUpper:  !genNoUpper,
			IncludeLower:  !genNoLower,
			IncludeDigits: !genNoDigits,
			IncludeSymbol: !genNoSymbols,
			Exclude:       genExclude,
		}
		if genCount < 1 {
			genCount = 1
		}
		var last string
		for i := 0; i < genCount; i++ {
			pw, err := pwgen.Generate(opts)
			if err != nil {
				return err
			}
			fmt.Println(pw)
			last = pw
		}
		if genCopy {
			if err := clipboard.Copy(last); err != nil {
				return err
			}
			msg := "last password copied to clipboard"
			if genClearAfter > 0 {
				spawnClipboardClear(last, genClearAfter)
				msg += fmt.Sprintf(" (clears in %s)", genClearAfter)
			}
			fmt.Println(ui.Successf(msg))
		}
		return nil
	},
}

func init() {
	genpassCmd.Flags().IntVarP(&genLen, "length", "l", 20, "password length")
	genpassCmd.Flags().BoolVar(&genNoUpper, "no-upper", false, "exclude uppercase")
	genpassCmd.Flags().BoolVar(&genNoLower, "no-lower", false, "exclude lowercase")
	genpassCmd.Flags().BoolVar(&genNoDigits, "no-digits", false, "exclude digits")
	genpassCmd.Flags().BoolVar(&genNoSymbols, "no-symbols", false, "exclude symbols")
	genpassCmd.Flags().StringVar(&genExclude, "exclude", "", "characters to exclude")
	genpassCmd.Flags().IntVarP(&genCount, "count", "n", 1, "generate N passwords")
	genpassCmd.Flags().BoolVarP(&genCopy, "clipboard", "c", false, "copy last generated password to clipboard")
	genpassCmd.Flags().DurationVar(&genClearAfter, "clear-after", clipClearTTL, "auto-clear the copied password after this long (0 disables)")
	rootCmd.AddCommand(genpassCmd)
}

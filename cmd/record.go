package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/johnniewhite/ssher/internal/recording"
	"github.com/johnniewhite/ssher/internal/ui"
)

var recordCmd = &cobra.Command{
	Use:   "record",
	Short: "List or replay session recordings",
}

var recordListCmd = &cobra.Command{
	Use:   "list",
	Short: "List session recordings",
	RunE: func(c *cobra.Command, args []string) error {
		paths, err := recording.List()
		if err != nil {
			return err
		}
		if len(paths) == 0 {
			fmt.Println(ui.Muted.Render("no recordings yet"))
			return nil
		}
		for _, p := range paths {
			info, err := os.Stat(p)
			if err != nil {
				continue
			}
			fmt.Printf("  %s  %d bytes  %s\n",
				ui.Title.Render(filepath.Base(p)),
				info.Size(),
				info.ModTime().Format("2006-01-02 15:04"))
		}
		return nil
	},
}

var recordReplaySpeed float64

var recordReplayCmd = &cobra.Command{
	Use:   "replay <file>",
	Short: "Replay a session recording (asciicast v2)",
	Args:  cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		return recording.Replay(args[0], os.Stdout, recordReplaySpeed)
	},
}

func init() {
	recordReplayCmd.Flags().Float64VarP(&recordReplaySpeed, "speed", "s", 1.0, "playback speed multiplier")
	recordCmd.AddCommand(recordListCmd, recordReplayCmd)
	rootCmd.AddCommand(recordCmd)
}

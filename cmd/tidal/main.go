package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/rustyguts/tidal/internal/version"
)

var rootCmd = &cobra.Command{
	Use:           "tidal",
	Short:         "Tidal — video transcoding pipeline",
	Long:          "Tidal queues, monitors, and automates ffmpeg-based video transcoding.",
	SilenceUsage:  true,
	SilenceErrors: true,
	Version:       fmt.Sprintf("%s (commit %s, built %s)", version.Version, version.Commit, version.Date),
}

func main() {
	rootCmd.AddCommand(serverCmd())
	rootCmd.AddCommand(workerCmd())
	rootCmd.AddCommand(runjobCmd())
	rootCmd.AddCommand(migrateCmd())
	rootCmd.AddCommand(jobCmd())
	rootCmd.AddCommand(versionCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

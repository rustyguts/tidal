package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/rustyguts/tidal/internal/version"
)

func versionCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		RunE: func(_ *cobra.Command, _ []string) error {
			info := version.Get()
			if asJSON {
				return json.NewEncoder(os.Stdout).Encode(info)
			}
			fmt.Printf("tidal %s (commit %s, built %s)\n", info.Version, info.Commit, info.Date)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "emit JSON")
	return cmd
}

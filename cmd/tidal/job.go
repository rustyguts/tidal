package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/rustyguts/tidal/internal/client"
)

func jobCmd() *cobra.Command {
	var serverURL string
	cmd := &cobra.Command{
		Use:   "job",
		Short: "Job management CLI (enqueue, list, get, cancel)",
	}
	cmd.PersistentFlags().StringVar(&serverURL, "server", envOr("TIDAL_SERVER_URL", "http://localhost:8080"), "Tidal server URL")

	var preset, source, output, cache, sourceMove string
	enqueue := &cobra.Command{
		Use:   "enqueue",
		Short: "Enqueue a transcode job (use --preset name OR --preset-id uuid)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if source == "" {
				return fmt.Errorf("--source required")
			}
			c := client.New(serverURL)
			pid, err := uuid.Parse(preset)
			if err != nil {
				p, err := c.PresetByName(cmd.Context(), preset)
				if err != nil {
					return err
				}
				pid = p.ID
			}
			j, err := c.EnqueueJob(cmd.Context(), client.EnqueueInput{
				PresetID:       pid.String(),
				SourcePath:     source,
				OutputPath:     output,
				CachePath:      cache,
				SourceMovePath: sourceMove,
			})
			if err != nil {
				return err
			}
			fmt.Printf("queued: %s status=%s output=%s\n", j.ID, j.Status, j.OutputPath)
			return nil
		},
	}
	enqueue.Flags().StringVar(&preset, "preset", "", "preset name or UUID")
	enqueue.Flags().StringVar(&source, "source", "", "absolute path to source file")
	enqueue.Flags().StringVar(&output, "output", "", "output path (default: alongside source)")
	enqueue.Flags().StringVar(&cache, "cache", "", "tidal cache path (default /var/cache/tidal in worker)")
	enqueue.Flags().StringVar(&sourceMove, "source-move", "", "move source here on success (file or directory)")
	cmd.AddCommand(enqueue)

	var status string
	var limit int
	list := &cobra.Command{
		Use:   "list",
		Short: "List recent jobs",
		RunE: func(cmd *cobra.Command, _ []string) error {
			c := client.New(serverURL)
			jobs, err := c.ListJobs(cmd.Context(), status, limit)
			if err != nil {
				return err
			}
			tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "ID\tSTATUS\tPCT\tSOURCE\tOUTPUT")
			for _, j := range jobs {
				fmt.Fprintf(tw, "%s\t%s\t%.1f%%\t%s\t%s\n", j.ID, j.Status, j.Progress.Percent, j.SourcePath, j.OutputPath)
			}
			return tw.Flush()
		},
	}
	list.Flags().StringVar(&status, "status", "", "filter by status")
	list.Flags().IntVar(&limit, "limit", 50, "max rows")
	cmd.AddCommand(list)

	get := &cobra.Command{
		Use:   "get [id]",
		Short: "Show one job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := uuid.Parse(args[0])
			if err != nil {
				return err
			}
			c := client.New(serverURL)
			j, err := c.GetJob(cmd.Context(), id)
			if err != nil {
				return err
			}
			fmt.Printf("id:        %s\nstatus:    %s\npreset:    %s\nsource:    %s\noutput:    %s\nprogress:  %.1f%% (frame=%d speed=%.2fx)\nerror:     %s\n",
				j.ID, j.Status, j.PresetID, j.SourcePath, j.OutputPath,
				j.Progress.Percent, j.Progress.Frame, j.Progress.Speed, j.Error)
			return nil
		},
	}
	cmd.AddCommand(get)

	cancel := &cobra.Command{
		Use:   "cancel [id]",
		Short: "Cancel a running job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := uuid.Parse(args[0])
			if err != nil {
				return err
			}
			return client.New(serverURL).CancelJob(cmd.Context(), id)
		},
	}
	cmd.AddCommand(cancel)

	return cmd
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

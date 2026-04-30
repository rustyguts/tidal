package main

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/rustyguts/tidal/internal/config"
	"github.com/rustyguts/tidal/internal/db"
	"github.com/rustyguts/tidal/internal/logging"
)

func migrateCmd() *cobra.Command {
	var dbURL string
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Apply database migrations",
	}
	cmd.PersistentFlags().StringVar(&dbURL, "db-url", "", "Postgres URL (defaults to TIDAL_DB_URL)")

	resolve := func() (string, error) {
		if dbURL != "" {
			return dbURL, nil
		}
		cfg, err := config.Load()
		if err != nil {
			return "", err
		}
		logging.Setup(logging.Options{Level: cfg.LogLevel, Pretty: cfg.LogPretty, Service: "migrate"})
		return cfg.DBURL, nil
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "up",
		Short: "Apply all pending up migrations",
		RunE: func(_ *cobra.Command, _ []string) error {
			url, err := resolve()
			if err != nil {
				return err
			}
			if err := db.Up(url); err != nil {
				return err
			}
			log.Info().Msg("migrations applied")
			return nil
		},
	})
	var steps int
	downCmd := &cobra.Command{
		Use:   "down",
		Short: "Roll back N migrations (default 1)",
		RunE: func(_ *cobra.Command, _ []string) error {
			url, err := resolve()
			if err != nil {
				return err
			}
			if err := db.Down(url, steps); err != nil {
				return err
			}
			log.Info().Int("steps", max(steps, 1)).Msg("migrations rolled back")
			return nil
		},
	}
	downCmd.Flags().IntVar(&steps, "steps", 1, "number of migrations to roll back")
	cmd.AddCommand(downCmd)

	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show current migration version",
		RunE: func(_ *cobra.Command, _ []string) error {
			url, err := resolve()
			if err != nil {
				return err
			}
			s, err := db.CurrentVersion(url)
			if err != nil {
				return err
			}
			fmt.Printf("version=%d dirty=%v\n", s.Version, s.Dirty)
			return nil
		},
	})

	return cmd
}

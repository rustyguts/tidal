package main

import (
	"context"
	"errors"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/rustyguts/tidal/internal/automations"
	"github.com/rustyguts/tidal/internal/config"
	"github.com/rustyguts/tidal/internal/db"
	"github.com/rustyguts/tidal/internal/jobs"
	"github.com/rustyguts/tidal/internal/logging"
	"github.com/rustyguts/tidal/internal/presets"
	"github.com/rustyguts/tidal/internal/queue"
	"github.com/rustyguts/tidal/internal/realtime"
	"github.com/rustyguts/tidal/internal/server"
)

func serverCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "server",
		Short: "Run the Tidal HTTP server (REST API + Vue SPA)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			logging.Setup(logging.Options{Level: cfg.LogLevel, Pretty: cfg.LogPretty, Service: "server"})

			ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			if cfg.AutoMigrate {
				if err := db.Up(cfg.DBURL); err != nil {
					return err
				}
				log.Info().Msg("auto-migrate complete")
			}

			pool, err := db.Connect(ctx, db.DefaultConfig(cfg.DBURL))
			if err != nil {
				return err
			}
			defer pool.Close()

			presetSvc := presets.New(pool)
			planted, err := presetSvc.Seed(ctx)
			if err != nil {
				return err
			}
			if len(planted) > 0 {
				log.Info().Strs("planted", planted).Msg("seeded builtin presets")
			}

			redisOpt, err := queue.ParseRedisURL(cfg.RedisURL)
			if err != nil {
				return err
			}
			redisCopy := redisOpt
			qc := queue.NewClient(redisOpt)
			defer qc.Close()

			hub := realtime.NewHub()
			jobSvc := jobs.NewService(pool, presetSvc, hub)
			jobSvc.SetEnqueuer(qc)

			autoSvc := automations.NewService(pool, presetSvc)
			jobSvc.SetArchiver(autoSvc)

			// Note: the automation scheduler runs in the worker pod (see
			// `tidal worker`), not here. Server stays stateless so it can scale
			// horizontally with zero-downtime rolling updates.

			srv := server.New(server.Deps{
				Config:      cfg,
				Pool:        pool,
				Presets:     presetSvc,
				Jobs:        jobSvc,
				Automations: autoSvc,
				Hub:         hub,
				RedisOpt:    &redisCopy,
			})

			errCh := make(chan error, 1)
			go func() { errCh <- srv.Start() }()

			select {
			case <-ctx.Done():
				log.Info().Msg("shutdown signal received")
			case err := <-errCh:
				if err != nil && !errors.Is(err, context.Canceled) {
					return err
				}
			}

			shutCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
			defer cancel()
			if err := srv.Shutdown(shutCtx); err != nil {
				return err
			}
			// Give in-flight goroutines a moment to drain.
			<-time.After(50 * time.Millisecond)
			return nil
		},
	}
}

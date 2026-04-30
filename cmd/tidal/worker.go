package main

import (
	"fmt"
	"os/signal"
	"syscall"

	corev1 "k8s.io/api/core/v1"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/rustyguts/tidal/internal/automations"
	"github.com/rustyguts/tidal/internal/config"
	"github.com/rustyguts/tidal/internal/db"
	"github.com/rustyguts/tidal/internal/jobs"
	"github.com/rustyguts/tidal/internal/k8s"
	"github.com/rustyguts/tidal/internal/logging"
	"github.com/rustyguts/tidal/internal/presets"
	"github.com/rustyguts/tidal/internal/queue"
	"github.com/rustyguts/tidal/internal/realtime"
	"github.com/rustyguts/tidal/internal/worker"
)

func workerCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "worker",
		Short: "Run the asynq worker (consumes the transcode queue)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			logging.Setup(logging.Options{Level: cfg.LogLevel, Pretty: cfg.LogPretty, Service: "worker"})

			ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			pool, err := db.Connect(ctx, db.DefaultConfig(cfg.DBURL))
			if err != nil {
				return err
			}
			defer pool.Close()

			redisOpt, err := queue.ParseRedisURL(cfg.RedisURL)
			if err != nil {
				return err
			}
			qc := queue.NewClient(redisOpt)
			defer qc.Close()

			hub := realtime.NewHub()
			presetSvc := presets.New(pool)
			jobSvc := jobs.NewService(pool, presetSvc, hub)
			jobSvc.SetEnqueuer(qc)
			autoSvc := automations.NewService(pool, presetSvc)
			jobSvc.SetArchiver(autoSvc)

			runner, err := buildRunner(cfg, jobSvc)
			if err != nil {
				return err
			}

			w := worker.New(worker.Config{
				RedisOpt:    redisOpt,
				Concurrency: cfg.WorkerConcurrency,
			}, jobSvc, runner)

			if err := w.Start(); err != nil {
				return err
			}

			<-ctx.Done()
			log.Info().Msg("shutdown signal received")
			w.Shutdown()
			return nil
		},
	}
}

func buildRunner(cfg config.Config, jobSvc *jobs.Service) (worker.Runner, error) {
	switch cfg.Dispatcher {
	case config.DispatcherLocal:
		return worker.NewLocalRunner(jobSvc), nil
	case config.DispatcherK8s:
		cli, err := k8s.NewClient()
		if err != nil {
			return nil, fmt.Errorf("k8s client: %w", err)
		}
		proto := k8s.JobSpec{
			Namespace:         cfg.JobNamespace,
			Image:             cfg.JobImage,
			ImagePullPolicy:   corev1.PullIfNotPresent,
			ServiceAccount:    cfg.JobServiceAccount,
			ServerInternalURL: cfg.ServerInternalURL,
			CallbackSecretRef: "tidal-callback-secret",
			MediaPVC:          cfg.JobMediaPVC,
			MediaHostPath:     cfg.JobMediaHostPath,
			MediaMountPath:    cfg.JobMediaMountPath,
		}
		return k8s.NewDispatcher(cli, proto), nil
	default:
		return nil, fmt.Errorf("unknown dispatcher mode %q", cfg.Dispatcher)
	}
}

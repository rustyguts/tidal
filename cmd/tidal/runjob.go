package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/rustyguts/tidal/internal/domain"
	"github.com/rustyguts/tidal/internal/ffmpeg"
	"github.com/rustyguts/tidal/internal/logging"
)

func runjobCmd() *cobra.Command {
	var (
		jobID          string
		serverURL      string
		callbackHeader string
	)
	cmd := &cobra.Command{
		Use:   "runjob",
		Short: "Run a single transcoding job (entrypoint inside Kubernetes Job pods)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			logging.Setup(logging.Options{Level: "info", Pretty: false, Service: "runjob"})

			secret := os.Getenv("TIDAL_CALLBACK_SECRET")
			if secret == "" {
				return fmt.Errorf("TIDAL_CALLBACK_SECRET env required")
			}

			ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			cb := ffmpeg.NewCallbackClient(serverURL, callbackHeader, secret)
			spec, err := cb.Spec(ctx, jobID)
			if err != nil {
				return fmt.Errorf("fetch spec: %w", err)
			}
			log.Info().
				Str("job", jobID).
				Str("source", spec.SourcePath).
				Str("output", spec.OutputPath).
				Msg("runjob start")

			probe, err := ffmpeg.Probe(ctx, spec.SourcePath)
			if err != nil {
				_ = cb.Status(context.Background(), jobID, string(domain.JobFailed), "probe: "+err.Error())
				return err
			}

			if err := cb.Status(ctx, jobID, string(domain.JobRunning), ""); err != nil {
				log.Warn().Err(err).Msg("post running status")
			}

			progSink := newProgressSink(cb, jobID)
			defer progSink.Close()

			runErr := ffmpeg.Run(ctx, ffmpeg.RunInput{
				Preset:     spec.Preset,
				SourcePath: spec.SourcePath,
				OutputPath: spec.OutputPath,
				DurationMs: probe.DurationMs,
			}, ffmpeg.Hooks{
				OnProgress: progSink.OnProgress,
				OnLog:      progSink.OnLog,
			})

			progSink.Flush()

			final := context.Background()
			switch {
			case runErr == nil:
				return cb.Status(final, jobID, string(domain.JobSucceeded), "")
			case errors.Is(runErr, context.Canceled):
				return cb.Status(final, jobID, string(domain.JobCancelled), "")
			default:
				return cb.Status(final, jobID, string(domain.JobFailed), runErr.Error())
			}
		},
	}
	cmd.Flags().StringVar(&jobID, "job-id", "", "ID of the job to run")
	cmd.Flags().StringVar(&serverURL, "server-url", "", "Tidal server base URL")
	cmd.Flags().StringVar(&callbackHeader, "callback-header", "X-Tidal-Callback-Secret", "header used for the shared secret")
	_ = cmd.MarkFlagRequired("job-id")
	_ = cmd.MarkFlagRequired("server-url")
	return cmd
}

// progressSink batches log lines and rate-limits progress posts so the
// callback channel doesn't drown the server.
type progressSink struct {
	cb    *ffmpeg.CallbackClient
	jobID string

	mu       sync.Mutex
	pending  ffmpeg.LogBatch
	lastProg time.Time
	stopCh   chan struct{}
	once     sync.Once
}

func newProgressSink(cb *ffmpeg.CallbackClient, jobID string) *progressSink {
	s := &progressSink{cb: cb, jobID: jobID, stopCh: make(chan struct{})}
	go s.flushLoop()
	return s
}

func (s *progressSink) OnProgress(p domain.FFmpegProgress) {
	s.mu.Lock()
	now := time.Now()
	if !s.lastProg.IsZero() && now.Sub(s.lastProg) < 500*time.Millisecond && p.Percent < 100 {
		s.mu.Unlock()
		return
	}
	s.lastProg = now
	s.mu.Unlock()
	if err := s.cb.Progress(context.Background(), s.jobID, p); err != nil {
		log.Warn().Err(err).Msg("post progress")
	}
}

func (s *progressSink) OnLog(l ffmpeg.LogLine) {
	s.mu.Lock()
	s.pending.Lines = append(s.pending.Lines, ffmpeg.LogBatchLine{Stream: l.Stream, Line: l.Line})
	flushNow := len(s.pending.Lines) >= 100
	s.mu.Unlock()
	if flushNow {
		s.Flush()
	}
}

func (s *progressSink) Flush() {
	s.mu.Lock()
	if len(s.pending.Lines) == 0 {
		s.mu.Unlock()
		return
	}
	batch := s.pending
	s.pending = ffmpeg.LogBatch{}
	s.mu.Unlock()
	if err := s.cb.Logs(context.Background(), s.jobID, batch); err != nil {
		log.Warn().Err(err).Int("lines", len(batch.Lines)).Msg("post logs")
	}
}

func (s *progressSink) Close() {
	s.once.Do(func() { close(s.stopCh) })
	s.Flush()
}

func (s *progressSink) flushLoop() {
	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-s.stopCh:
			return
		case <-t.C:
			s.Flush()
		}
	}
}

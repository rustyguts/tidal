package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"

	"github.com/rustyguts/tidal/internal/domain"
	"github.com/rustyguts/tidal/internal/jobs"
	"github.com/rustyguts/tidal/internal/queue"
)

// Runner is the abstraction the worker calls to actually execute a transcode.
// In local mode it's a thin wrapper over jobs.Service.Run; in k8s mode it
// creates a Kubernetes Job and waits for it to land.
type Runner interface {
	Run(ctx context.Context, jobID domain.JobID) error
}

type Config struct {
	RedisOpt    asynq.RedisClientOpt
	Concurrency int
}

type Worker struct {
	srv     *asynq.Server
	mux     *asynq.ServeMux
	jobsSvc *jobs.Service
	runner  Runner
}

func New(cfg Config, jobsSvc *jobs.Service, runner Runner) *Worker {
	c := cfg.Concurrency
	if c <= 0 {
		c = 4
	}
	srv := asynq.NewServer(cfg.RedisOpt, asynq.Config{
		Concurrency: c,
		Queues: map[string]int{
			queue.QueueTranscode: 6,
			queue.QueueScan:      3,
			queue.QueueDefault:   1,
		},
		Logger: zerologAdapter{},
	})
	mux := asynq.NewServeMux()
	w := &Worker{srv: srv, mux: mux, jobsSvc: jobsSvc, runner: runner}
	mux.HandleFunc(queue.TaskTypeTranscode, w.handleTranscode)
	mux.HandleFunc(queue.TaskTypeScan, w.handleScan)
	mux.HandleFunc(queue.TaskTypeCleanup, w.handleCleanup)
	return w
}

func (w *Worker) Start() error {
	log.Info().Msg("worker starting")
	if err := w.srv.Start(w.mux); err != nil {
		return fmt.Errorf("worker start: %w", err)
	}
	return nil
}

func (w *Worker) Shutdown() {
	log.Info().Msg("worker shutting down")
	w.srv.Shutdown()
}

func (w *Worker) Stop() {
	w.srv.Stop()
}

func (w *Worker) handleTranscode(ctx context.Context, t *asynq.Task) error {
	var p queue.TranscodePayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal transcode: %w", err)
	}
	return w.runner.Run(ctx, p.JobID)
}

func (w *Worker) handleScan(ctx context.Context, t *asynq.Task) error {
	// Server-side scheduler currently drives scans directly; the asynq path
	// is reserved for moving scheduling to the queue (Phase 9 follow-up).
	log.Debug().Msg("scan task (stub)")
	_ = ctx
	_ = t
	return nil
}

func (w *Worker) handleCleanup(ctx context.Context, t *asynq.Task) error {
	var p queue.CleanupPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("unmarshal cleanup: %w", err)
	}
	older := time.Duration(p.OlderThanHours) * time.Hour
	if older <= 0 {
		older = 30 * 24 * time.Hour
	}
	switch p.Kind {
	case "", "jobs":
		n, err := w.jobsSvc.PruneOldTerminal(ctx, older)
		if err != nil {
			return err
		}
		log.Info().Int64("deleted", n).Dur("older_than", older).Msg("pruned old jobs")
	default:
		return fmt.Errorf("unknown cleanup kind %q", p.Kind)
	}
	return nil
}

// LocalRunner satisfies Runner by invoking jobs.Service.Run in-process.
type LocalRunner struct {
	svc *jobs.Service
}

func NewLocalRunner(svc *jobs.Service) *LocalRunner { return &LocalRunner{svc: svc} }

func (r *LocalRunner) Run(ctx context.Context, jobID domain.JobID) error {
	return r.svc.Run(ctx, jobID)
}

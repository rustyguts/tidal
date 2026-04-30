package jobs

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/rustyguts/tidal/internal/domain"
	"github.com/rustyguts/tidal/internal/ffmpeg"
	"github.com/rustyguts/tidal/internal/metrics"
	"github.com/rustyguts/tidal/internal/presets"
	"github.com/rustyguts/tidal/internal/realtime"
)

// Enqueuer abstracts the asynq client so jobs.Service stays decoupled from
// the queue package (and can be unit-tested with a fake).
type Enqueuer interface {
	EnqueueTranscode(ctx context.Context, jobID domain.JobID) (string, error)
	CancelProcessing(asynqID string) error
}

// SourceArchiver moves a source file into an automation's archive_dir after
// a successful transcode. Optional; jobs.Service skips archive when nil.
type SourceArchiver interface {
	ArchiveSource(ctx context.Context, automationID domain.AutomationID, sourcePath string) error
}

type Service struct {
	repo     *repo
	presets  *presets.Service
	queue    Enqueuer
	archiver SourceArchiver
	hub      *realtime.Hub
}

func NewService(pool *pgxpool.Pool, presetSvc *presets.Service, hub *realtime.Hub) *Service {
	return &Service{
		repo:    newRepo(pool),
		presets: presetSvc,
		hub:     hub,
	}
}

// SetEnqueuer wires the queue client. Must be called before Create is invoked.
func (s *Service) SetEnqueuer(e Enqueuer) { s.queue = e }

// SetArchiver wires the post-transcode source archiver. Optional.
func (s *Service) SetArchiver(a SourceArchiver) { s.archiver = a }

// DefaultCachePath is the in-container directory used for ffmpeg working
// files when the caller doesn't override it. Mounted via a volume in the
// container; clean for each pod restart.
const DefaultCachePath = "/var/cache/tidal"

type CreateInput struct {
	PresetID       domain.PresetID
	SourcePath     string
	OutputPath     string // optional; derived from preset suffix if empty
	CachePath      string // optional; defaults to DefaultCachePath
	SourceMovePath string // optional; if set, source moves here on success
	AutomationID   *domain.AutomationID
}

// Create validates the inputs, persists the job row, and enqueues an asynq
// transcode task. The worker picks the task up and calls Run.
func (s *Service) Create(ctx context.Context, in CreateInput) (domain.Job, error) {
	preset, err := s.presets.Get(ctx, in.PresetID)
	if err != nil {
		return domain.Job{}, err
	}

	out := strings.TrimSpace(in.OutputPath)
	if out == "" {
		out = derivedOutputPath(in.SourcePath, preset)
	}
	cache := strings.TrimSpace(in.CachePath)
	if cache == "" {
		cache = DefaultCachePath
	}

	job := domain.Job{
		ID:             uuid.New(),
		PresetID:       in.PresetID,
		SourcePath:     in.SourcePath,
		OutputPath:     out,
		CachePath:      cache,
		SourceMovePath: strings.TrimSpace(in.SourceMovePath),
		Status:         domain.JobQueued,
		AutomationID:   in.AutomationID,
		CreatedAt:      time.Now().UTC(),
	}
	if err := s.repo.insert(ctx, job); err != nil {
		return domain.Job{}, err
	}

	if s.queue == nil {
		return domain.Job{}, errors.New("no queue client wired")
	}
	asynqID, err := s.queue.EnqueueTranscode(ctx, job.ID)
	if err != nil {
		_ = s.repo.updateStatus(ctx, job.ID, domain.JobFailed, "enqueue: "+err.Error(), true)
		s.publishStatus(job.ID, domain.JobFailed, err.Error())
		return domain.Job{}, err
	}
	if err := s.repo.setAsynqID(ctx, job.ID, asynqID); err != nil {
		log.Warn().Err(err).Str("job", job.ID.String()).Msg("set asynq id")
	}

	s.publishStatus(job.ID, domain.JobQueued, "")
	return s.Get(ctx, job.ID)
}

func (s *Service) Get(ctx context.Context, id domain.JobID) (domain.Job, error) {
	return s.repo.get(ctx, id)
}

type ListInput struct {
	Status   string
	PresetID *uuid.UUID
	Limit    int
}

func (s *Service) List(ctx context.Context, in ListInput) ([]domain.Job, error) {
	return s.repo.list(ctx, listFilter{
		Status:   in.Status,
		PresetID: in.PresetID,
		Limit:    in.Limit,
	})
}

func (s *Service) ListLogs(ctx context.Context, id domain.JobID, fromSeq int64, limit int) ([]domain.JobLog, error) {
	if _, err := s.repo.get(ctx, id); err != nil {
		return nil, err
	}
	return s.repo.listLogs(ctx, id, fromSeq, limit)
}

// Cancel marks the job as cancelling and signals the asynq task. The worker
// observes ctx.Done and finishes via Cancelled.
func (s *Service) Cancel(ctx context.Context, id domain.JobID) error {
	j, err := s.repo.get(ctx, id)
	if err != nil {
		return err
	}
	if j.Status.Terminal() {
		return nil
	}
	if err := s.repo.updateStatus(ctx, id, domain.JobCancelling, "", false); err != nil {
		return err
	}
	s.publishStatus(id, domain.JobCancelling, "")
	if s.queue != nil {
		if err := s.queue.CancelProcessing(j.AsynqID); err != nil {
			log.Warn().Err(err).Str("job", id.String()).Msg("cancel processing")
		}
	}
	return nil
}

// Run executes a job in-process. Called by the worker's transcode handler in
// local-dispatcher mode. The worker's ctx is cancelled when CancelProcessing is
// invoked, so SIGTERM propagates to ffmpeg via exec.CommandContext.
func (s *Service) Run(ctx context.Context, id domain.JobID) error {
	j, err := s.repo.get(ctx, id)
	if err != nil {
		return err
	}
	preset, err := s.presets.Get(ctx, j.PresetID)
	if err != nil {
		s.Failed(ctx, id, fmt.Errorf("load preset: %w", err))
		return err
	}

	probe, err := ffmpeg.Probe(ctx, j.SourcePath)
	if err != nil {
		s.Failed(ctx, id, fmt.Errorf("probe: %w", err))
		return nil // do not asynq-retry; failure recorded
	}

	s.Started(ctx, id)

	last := time.Time{}
	hooks := ffmpeg.Hooks{
		OnProgress: func(p domain.FFmpegProgress) {
			now := time.Now()
			if !last.IsZero() && now.Sub(last) < 500*time.Millisecond && p.Percent < 100 {
				return
			}
			last = now
			s.Progress(ctx, id, p)
		},
		OnLog: func(l ffmpeg.LogLine) {
			s.AppendLog(ctx, id, l.Stream, l.Line)
		},
	}

	runErr := ffmpeg.Run(ctx, ffmpeg.RunInput{
		Preset:     preset.Spec,
		SourcePath: j.SourcePath,
		OutputPath: j.OutputPath,
		DurationMs: probe.DurationMs,
	}, hooks)

	switch {
	case runErr == nil:
		s.Succeeded(ctx, id)
	case errors.Is(runErr, context.Canceled):
		s.Cancelled(ctx, id)
	default:
		s.Failed(ctx, id, fmt.Errorf("ffmpeg: %w", runErr))
	}
	return nil
}

// --- ProgressSink-style methods used by Run + Phase 5 callbacks ---

func (s *Service) Started(ctx context.Context, id domain.JobID) {
	if err := s.repo.updateStatus(ctx, id, domain.JobRunning, "", false); err != nil {
		log.Error().Err(err).Msg("started: update status")
	}
	s.publishStatus(id, domain.JobRunning, "")
}

func (s *Service) Progress(ctx context.Context, id domain.JobID, p domain.FFmpegProgress) {
	p.UpdatedAt = time.Now().UTC()
	if err := s.repo.updateProgress(ctx, id, p); err != nil {
		log.Error().Err(err).Msg("progress: update")
	}
	s.hub.Publish(topicJob(id), realtime.EventProgress, p)
}

func (s *Service) AppendLog(ctx context.Context, id domain.JobID, stream, line string) {
	if err := s.repo.appendLog(ctx, id, stream, line); err != nil {
		log.Warn().Err(err).Msg("append log")
		return
	}
	s.hub.Publish(topicJob(id), realtime.EventLog, domain.JobLog{
		JobID: id, Stream: stream, Line: line, EmittedAt: time.Now().UTC(),
	})
}

func (s *Service) Succeeded(ctx context.Context, id domain.JobID) {
	if err := s.repo.updateStatus(ctx, id, domain.JobSucceeded, "", true); err != nil {
		log.Error().Err(err).Msg("succeeded: update")
	}
	s.publishStatus(id, domain.JobSucceeded, "")
	s.recordTerminalMetric(ctx, id, domain.JobSucceeded)
	s.maybeMoveSource(ctx, id)
	s.maybeArchive(ctx, id)
}

func (s *Service) maybeMoveSource(ctx context.Context, id domain.JobID) {
	j, err := s.repo.get(ctx, id)
	if err != nil || j.SourceMovePath == "" {
		return
	}
	if err := moveFile(j.SourcePath, j.SourceMovePath); err != nil {
		log.Warn().Err(err).Str("job", id.String()).Str("dest", j.SourceMovePath).Msg("move source")
	}
}

func (s *Service) recordTerminalMetric(ctx context.Context, id domain.JobID, status domain.JobStatus) {
	metrics.JobsTotal.WithLabelValues(string(status)).Inc()
	j, err := s.repo.get(ctx, id)
	if err != nil || j.StartedAt == nil || j.FinishedAt == nil {
		return
	}
	metrics.JobDuration.WithLabelValues(string(status)).Observe(j.FinishedAt.Sub(*j.StartedAt).Seconds())
}

// PruneOldTerminal deletes finished/failed/cancelled jobs older than `older`.
// ON DELETE CASCADE on job_logs cleans those up too.
func (s *Service) PruneOldTerminal(ctx context.Context, older time.Duration) (int64, error) {
	return s.repo.PruneOldTerminal(ctx, older)
}

func (s *Service) maybeArchive(ctx context.Context, id domain.JobID) {
	if s.archiver == nil {
		return
	}
	j, err := s.repo.get(ctx, id)
	if err != nil || j.AutomationID == nil {
		return
	}
	if err := s.archiver.ArchiveSource(ctx, *j.AutomationID, j.SourcePath); err != nil {
		log.Warn().Err(err).Str("job", id.String()).Msg("archive source")
	}
}

func (s *Service) Failed(ctx context.Context, id domain.JobID, jobErr error) {
	msg := jobErr.Error()
	if err := s.repo.updateStatus(ctx, id, domain.JobFailed, msg, true); err != nil {
		log.Error().Err(err).Msg("failed: update")
	}
	s.publishStatus(id, domain.JobFailed, msg)
	s.recordTerminalMetric(ctx, id, domain.JobFailed)
}

func (s *Service) Cancelled(ctx context.Context, id domain.JobID) {
	if err := s.repo.updateStatus(ctx, id, domain.JobCancelled, "", true); err != nil {
		log.Error().Err(err).Msg("cancelled: update")
	}
	s.publishStatus(id, domain.JobCancelled, "")
	s.recordTerminalMetric(ctx, id, domain.JobCancelled)
}

func (s *Service) publishStatus(id domain.JobID, status domain.JobStatus, errMsg string) {
	s.hub.Publish(topicJob(id), realtime.EventStatus, map[string]any{
		"jobId":  id,
		"status": status,
		"error":  errMsg,
	})
}

func topicJob(id domain.JobID) string { return "job:" + id.String() }

func derivedOutputPath(source string, p domain.Preset) string {
	dir := filepath.Dir(source)
	base := strings.TrimSuffix(filepath.Base(source), filepath.Ext(source))
	suffix := p.Spec.OutputSuffix
	if suffix == "" {
		suffix = "_" + p.Name
	}
	ext := strings.ToLower(p.Spec.Container)
	if ext == "" {
		ext = "mp4"
	}
	return filepath.Join(dir, base+suffix+"."+ext)
}

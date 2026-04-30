package automations

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/rustyguts/tidal/internal/domain"
	"github.com/rustyguts/tidal/internal/jobs"
)

// Scanner handles a single scan tick for one automation: enumerate the watch
// directory, match the glob, debounce recently-modified files, and enqueue a
// transcode for each match. Outcome is recorded in automation_runs.
type Scanner struct {
	autos *Service
	jobs  *jobs.Service
}

func NewScanner(autos *Service, jobsSvc *jobs.Service) *Scanner {
	return &Scanner{autos: autos, jobs: jobsSvc}
}

// ScanOnce runs a single scan tick. Idempotent: if a source has already been
// enqueued (per the automation_runs unique index), insertRun coerces the row
// to "skipped_dupe".
func (s *Scanner) ScanOnce(ctx context.Context, a domain.Automation) error {
	entries, err := os.ReadDir(a.WatchDir)
	if err != nil {
		_ = s.autos.RecordRun(ctx, a.ID, a.WatchDir, domain.OutcomeError, err.Error(), nil)
		return fmt.Errorf("read %s: %w", a.WatchDir, err)
	}
	debounce := time.Duration(a.DebounceMs) * time.Millisecond
	now := time.Now()

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		match, err := filepath.Match(a.Glob, e.Name())
		if err != nil || !match {
			continue
		}
		full := filepath.Join(a.WatchDir, e.Name())
		fi, err := e.Info()
		if err != nil {
			continue
		}
		if debounce > 0 && now.Sub(fi.ModTime()) < debounce {
			// Source still being written; wait for next tick.
			continue
		}

		out := filepath.Join(a.OutputDir, outputName(e.Name(), a))

		dupe, err := s.autos.HasEnqueued(ctx, a.ID, full)
		if err != nil {
			log.Warn().Str("source", full).Err(err).Msg("dedupe check")
			continue
		}
		if dupe {
			continue
		}

		automationID := a.ID
		j, err := s.jobs.Create(ctx, jobs.CreateInput{
			PresetID:     a.PresetID,
			SourcePath:   full,
			OutputPath:   out,
			AutomationID: &automationID,
		})
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			_ = s.autos.RecordRun(ctx, a.ID, full, domain.OutcomeError, err.Error(), nil)
			log.Warn().Str("automation", a.Name).Str("source", full).Err(err).Msg("enqueue failed")
			continue
		}
		jid := j.ID
		_ = s.autos.RecordRun(ctx, a.ID, full, domain.OutcomeEnqueued, "", &jid)
		log.Info().Str("automation", a.Name).Str("source", full).Str("job", j.ID.String()).Msg("automation enqueued job")
	}
	return nil
}

func outputName(filename string, a domain.Automation) string {
	stem := filename[:len(filename)-len(filepath.Ext(filename))]
	// We don't have the preset spec here; rely on jobs.Service.Create
	// fallback path which derives extension from the preset container.
	// Just append a `.out` placeholder so it's distinct; the service will
	// override the path because we pass OutputPath explicitly. So instead
	// keep the original extension and let jobs.Service Trust the caller.
	return stem + filepath.Ext(filename)
}

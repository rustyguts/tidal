package workflows

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/rustyguts/tidal/internal/domain"
	"github.com/rustyguts/tidal/internal/jobs"
)

// Executor runs the configured actions of a workflow against a single trigger
// event. v1 supports a single `enqueue_transcode` action per workflow.
type Executor struct {
	wf   *Service
	jobs *jobs.Service
}

func NewExecutor(wf *Service, jobsSvc *jobs.Service) *Executor {
	return &Executor{wf: wf, jobs: jobsSvc}
}

// Execute fans out the workflow's actions for the given trigger context.
// Records workflow_runs row on enqueue + bumps the runs counter once per
// successful enqueue.
func (e *Executor) Execute(ctx context.Context, w domain.Workflow, tc TriggerContext) error {
	dupe, err := e.wf.HasEnqueued(ctx, w.ID, tc.Path)
	if err != nil {
		return fmt.Errorf("dedupe check: %w", err)
	}
	if dupe {
		return nil
	}

	for i, a := range w.Actions {
		if a.Type != domain.ActionEnqueueTranscode {
			log.Warn().Str("workflow", w.Name).Str("action", a.Type).Msg("unsupported action type, skipping")
			continue
		}
		if err := e.enqueueTranscode(ctx, w, a, tc); err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			_ = e.wf.RecordRun(ctx, w.ID, tc.Path, domain.WorkflowOutcomeError,
				fmt.Sprintf("action[%d]: %v", i, err), nil)
			log.Warn().Str("workflow", w.Name).Str("source", tc.Path).Err(err).Msg("action failed")
			return err
		}
	}
	return nil
}

func (e *Executor) enqueueTranscode(ctx context.Context, w domain.Workflow, a domain.Action, tc TriggerContext) error {
	vars := tc.Vars()
	out := Render(a.OutputPath, vars)
	move := Render(a.SourceMovePath, vars)
	cache := Render(a.CachePath, vars)

	pid, err := uuid.Parse(a.PresetID)
	if err != nil {
		return fmt.Errorf("preset uuid: %w", err)
	}

	wfID := w.ID
	j, err := e.jobs.Create(ctx, jobs.CreateInput{
		PresetID:       pid,
		SourcePath:     tc.Path,
		OutputPath:     out,
		SourceMovePath: move,
		CachePath:      cache,
		WorkflowID:     &wfID,
	})
	if err != nil {
		return err
	}

	jid := j.ID
	if err := e.wf.RecordRun(ctx, w.ID, tc.Path, domain.WorkflowOutcomeEnqueued, "", &jid); err != nil {
		log.Warn().Err(err).Msg("record workflow run")
	}
	if err := e.wf.IncrementRuns(ctx, w.ID); err != nil {
		log.Warn().Err(err).Msg("increment runs")
	}
	log.Info().
		Str("workflow", w.Name).
		Str("source", tc.Path).
		Str("job", j.ID.String()).
		Time("at", time.Now().UTC()).
		Msg("workflow fired")
	return nil
}

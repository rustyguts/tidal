package workflows

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rustyguts/tidal/internal/domain"
	"github.com/rustyguts/tidal/internal/presets"
)

type Service struct {
	repo    *repo
	presets *presets.Service
}

func NewService(pool *pgxpool.Pool, presetSvc *presets.Service) *Service {
	return &Service{repo: newRepo(pool), presets: presetSvc}
}

type CreateInput struct {
	Name              string
	Enabled           bool
	Trigger           domain.Trigger
	Actions           []domain.Action
	PollIntervalMs    int
	StableThresholdMs int
}

func (s *Service) List(ctx context.Context) ([]domain.Workflow, error) {
	return s.repo.list(ctx)
}

func (s *Service) ListEnabled(ctx context.Context) ([]domain.Workflow, error) {
	return s.repo.listEnabled(ctx)
}

func (s *Service) Get(ctx context.Context, id domain.WorkflowID) (domain.Workflow, error) {
	return s.repo.get(ctx, id)
}

func (s *Service) Create(ctx context.Context, in CreateInput) (domain.Workflow, error) {
	if err := s.validate(ctx, in.Trigger, in.Actions, in.Name); err != nil {
		return domain.Workflow{}, err
	}
	w := domain.Workflow{
		ID:                uuid.New(),
		Name:              strings.TrimSpace(in.Name),
		Enabled:           in.Enabled,
		Trigger:           in.Trigger,
		Actions:           in.Actions,
		PollIntervalMs:    defaultIfZero(in.PollIntervalMs, 30000),
		StableThresholdMs: defaultIfZero(in.StableThresholdMs, 60000),
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}
	if err := s.repo.insert(ctx, w); err != nil {
		return domain.Workflow{}, err
	}
	return s.repo.get(ctx, w.ID)
}

type UpdateInput struct {
	Name              *string
	Enabled           *bool
	Trigger           *domain.Trigger
	Actions           *[]domain.Action
	PollIntervalMs    *int
	StableThresholdMs *int
}

func (s *Service) Update(ctx context.Context, id domain.WorkflowID, in UpdateInput) (domain.Workflow, error) {
	cur, err := s.repo.get(ctx, id)
	if err != nil {
		return domain.Workflow{}, err
	}
	if in.Name != nil {
		cur.Name = strings.TrimSpace(*in.Name)
	}
	if in.Enabled != nil {
		cur.Enabled = *in.Enabled
	}
	if in.Trigger != nil {
		cur.Trigger = *in.Trigger
	}
	if in.Actions != nil {
		cur.Actions = *in.Actions
	}
	if in.PollIntervalMs != nil {
		cur.PollIntervalMs = *in.PollIntervalMs
	}
	if in.StableThresholdMs != nil {
		cur.StableThresholdMs = *in.StableThresholdMs
	}
	if err := s.validate(ctx, cur.Trigger, cur.Actions, cur.Name); err != nil {
		return domain.Workflow{}, err
	}
	if err := s.repo.update(ctx, cur); err != nil {
		return domain.Workflow{}, err
	}
	return s.repo.get(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id domain.WorkflowID) error {
	return s.repo.delete(ctx, id)
}

func (s *Service) SetEnabled(ctx context.Context, id domain.WorkflowID, enabled bool) error {
	return s.repo.setEnabled(ctx, id, enabled)
}

func (s *Service) ListRuns(ctx context.Context, id domain.WorkflowID, limit int) ([]domain.WorkflowRun, error) {
	if _, err := s.repo.get(ctx, id); err != nil {
		return nil, err
	}
	return s.repo.listRuns(ctx, id, limit)
}

// IncrementRuns increments runs_count + sets last_run_at.
func (s *Service) IncrementRuns(ctx context.Context, id domain.WorkflowID) error {
	return s.repo.incrementRuns(ctx, id)
}

// IncrementSuccess increments success_count.
func (s *Service) IncrementSuccess(ctx context.Context, id domain.WorkflowID) error {
	return s.repo.incrementSuccess(ctx, id)
}

// HasEnqueued reports whether a source path has been recorded as enqueued for
// this workflow. Used by the watcher to dedupe.
func (s *Service) HasEnqueued(ctx context.Context, id domain.WorkflowID, sourcePath string) (bool, error) {
	return s.repo.hasEnqueued(ctx, id, sourcePath)
}

// RecordRun appends a row to workflow_runs.
func (s *Service) RecordRun(ctx context.Context, workflowID domain.WorkflowID, sourcePath, outcome, message string, jobID *domain.JobID) error {
	return s.repo.insertRun(ctx, runInsert{
		WorkflowID: workflowID,
		SourcePath: sourcePath,
		JobID:      jobID,
		Outcome:    outcome,
		Message:    message,
	})
}

func (s *Service) validate(ctx context.Context, trigger domain.Trigger, actions []domain.Action, name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("name required")
	}
	if err := trigger.Validate(); err != nil {
		return fmt.Errorf("trigger: %w", err)
	}
	if len(actions) == 0 {
		return fmt.Errorf("at least one action required")
	}
	for i, a := range actions {
		if err := a.Validate(); err != nil {
			return fmt.Errorf("actions[%d]: %w", i, err)
		}
		if a.Type == domain.ActionEnqueueTranscode {
			pid, err := uuid.Parse(a.PresetID)
			if err != nil {
				return fmt.Errorf("actions[%d].presetId invalid uuid", i)
			}
			if _, err := s.presets.Get(ctx, pid); err != nil {
				return fmt.Errorf("actions[%d].presetId not found", i)
			}
		}
	}
	return nil
}

func defaultIfZero(v, def int) int {
	if v <= 0 {
		return def
	}
	return v
}

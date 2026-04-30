package automations

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
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

func (s *Service) List(ctx context.Context) ([]domain.Automation, error) {
	return s.repo.list(ctx)
}

func (s *Service) ListEnabled(ctx context.Context) ([]domain.Automation, error) {
	return s.repo.listEnabled(ctx)
}

func (s *Service) Get(ctx context.Context, id domain.AutomationID) (domain.Automation, error) {
	return s.repo.get(ctx, id)
}

type CreateInput struct {
	Name           string
	Enabled        bool
	WatchDir       string
	Glob           string
	PresetID       domain.PresetID
	OutputDir      string
	ArchiveDir     string
	PollIntervalMs int
	DebounceMs     int
}

func (s *Service) Create(ctx context.Context, in CreateInput) (domain.Automation, error) {
	if _, err := s.presets.Get(ctx, in.PresetID); err != nil {
		return domain.Automation{}, fmt.Errorf("preset: %w", err)
	}
	if err := validateInput(in); err != nil {
		return domain.Automation{}, err
	}
	a := domain.Automation{
		ID:             uuid.New(),
		Name:           in.Name,
		Enabled:        in.Enabled,
		WatchDir:       in.WatchDir,
		Glob:           in.Glob,
		PresetID:       in.PresetID,
		OutputDir:      in.OutputDir,
		ArchiveDir:     in.ArchiveDir,
		PollIntervalMs: in.PollIntervalMs,
		DebounceMs:     in.DebounceMs,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	if a.PollIntervalMs == 0 {
		a.PollIntervalMs = 30000
	}
	if a.DebounceMs == 0 {
		a.DebounceMs = 5000
	}
	a.PollInterval = time.Duration(a.PollIntervalMs) * time.Millisecond
	if err := s.repo.insert(ctx, a); err != nil {
		return domain.Automation{}, err
	}
	return s.repo.get(ctx, a.ID)
}

type UpdateInput struct {
	Name           *string
	Enabled        *bool
	WatchDir       *string
	Glob           *string
	PresetID       *domain.PresetID
	OutputDir      *string
	ArchiveDir     *string
	PollIntervalMs *int
	DebounceMs     *int
}

func (s *Service) Update(ctx context.Context, id domain.AutomationID, in UpdateInput) (domain.Automation, error) {
	cur, err := s.repo.get(ctx, id)
	if err != nil {
		return domain.Automation{}, err
	}
	if in.Name != nil {
		cur.Name = *in.Name
	}
	if in.Enabled != nil {
		cur.Enabled = *in.Enabled
	}
	if in.WatchDir != nil {
		cur.WatchDir = *in.WatchDir
	}
	if in.Glob != nil {
		cur.Glob = *in.Glob
	}
	if in.PresetID != nil {
		if _, err := s.presets.Get(ctx, *in.PresetID); err != nil {
			return domain.Automation{}, fmt.Errorf("preset: %w", err)
		}
		cur.PresetID = *in.PresetID
	}
	if in.OutputDir != nil {
		cur.OutputDir = *in.OutputDir
	}
	if in.ArchiveDir != nil {
		cur.ArchiveDir = *in.ArchiveDir
	}
	if in.PollIntervalMs != nil {
		cur.PollIntervalMs = *in.PollIntervalMs
	}
	if in.DebounceMs != nil {
		cur.DebounceMs = *in.DebounceMs
	}
	if err := s.repo.update(ctx, cur); err != nil {
		return domain.Automation{}, err
	}
	return s.repo.get(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id domain.AutomationID) error {
	return s.repo.delete(ctx, id)
}

func (s *Service) SetEnabled(ctx context.Context, id domain.AutomationID, enabled bool) error {
	return s.repo.setEnabled(ctx, id, enabled)
}

func (s *Service) ListRuns(ctx context.Context, id domain.AutomationID, limit int) ([]domain.AutomationRun, error) {
	if _, err := s.repo.get(ctx, id); err != nil {
		return nil, err
	}
	return s.repo.listRuns(ctx, id, limit)
}

// HasEnqueued reports whether the source is already recorded as enqueued.
// The scanner consults this before creating a duplicate job row.
func (s *Service) HasEnqueued(ctx context.Context, automationID domain.AutomationID, sourcePath string) (bool, error) {
	return s.repo.hasEnqueued(ctx, automationID, sourcePath)
}

// RecordRun is exposed so the scanner / executor can write run history entries.
func (s *Service) RecordRun(ctx context.Context, automationID domain.AutomationID, sourcePath, outcome, message string, jobID *domain.JobID) error {
	return s.repo.insertRun(ctx, runInsert{
		AutomationID: automationID,
		SourcePath:   sourcePath,
		JobID:        jobID,
		Outcome:      outcome,
		Message:      message,
	})
}

func validateInput(in CreateInput) error {
	for k, v := range map[string]string{
		"name":       in.Name,
		"watchDir":   in.WatchDir,
		"glob":       in.Glob,
		"outputDir":  in.OutputDir,
		"archiveDir": in.ArchiveDir,
	} {
		if strings.TrimSpace(v) == "" {
			return fmt.Errorf("%s required", k)
		}
	}
	if !filepath.IsAbs(in.WatchDir) || !filepath.IsAbs(in.OutputDir) || !filepath.IsAbs(in.ArchiveDir) {
		return errors.New("watchDir, outputDir, and archiveDir must be absolute paths")
	}
	return nil
}

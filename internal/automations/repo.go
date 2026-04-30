package automations

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rustyguts/tidal/internal/domain"
)

var (
	ErrNotFound = errors.New("automation not found")
	ErrConflict = errors.New("automation name already exists")
)

const automationCols = "id, name, enabled, watch_dir, glob, preset_id, output_dir, archive_dir, poll_interval_ms, debounce_ms, created_at, updated_at"

type repo struct {
	pool *pgxpool.Pool
}

func newRepo(pool *pgxpool.Pool) *repo { return &repo{pool: pool} }

func (r *repo) list(ctx context.Context) ([]domain.Automation, error) {
	rows, err := r.pool.Query(ctx, "SELECT "+automationCols+" FROM automations ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("list automations: %w", err)
	}
	defer rows.Close()
	out := []domain.Automation{}
	for rows.Next() {
		a, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *repo) listEnabled(ctx context.Context) ([]domain.Automation, error) {
	rows, err := r.pool.Query(ctx,
		"SELECT "+automationCols+" FROM automations WHERE enabled = true ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Automation{}
	for rows.Next() {
		a, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *repo) get(ctx context.Context, id domain.AutomationID) (domain.Automation, error) {
	row := r.pool.QueryRow(ctx, "SELECT "+automationCols+" FROM automations WHERE id = $1", id)
	a, err := scan(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Automation{}, ErrNotFound
	}
	return a, err
}

func (r *repo) insert(ctx context.Context, a domain.Automation) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO automations (id, name, enabled, watch_dir, glob, preset_id, output_dir, archive_dir, poll_interval_ms, debounce_ms)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		a.ID, a.Name, a.Enabled, a.WatchDir, a.Glob, a.PresetID, a.OutputDir, a.ArchiveDir, a.PollIntervalMs, a.DebounceMs,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrConflict
		}
		return fmt.Errorf("insert automation: %w", err)
	}
	return nil
}

func (r *repo) update(ctx context.Context, a domain.Automation) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE automations SET name=$2, enabled=$3, watch_dir=$4, glob=$5, preset_id=$6,
		 output_dir=$7, archive_dir=$8, poll_interval_ms=$9, debounce_ms=$10, updated_at=now()
		 WHERE id=$1`,
		a.ID, a.Name, a.Enabled, a.WatchDir, a.Glob, a.PresetID, a.OutputDir, a.ArchiveDir, a.PollIntervalMs, a.DebounceMs,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrConflict
		}
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *repo) delete(ctx context.Context, id domain.AutomationID) error {
	tag, err := r.pool.Exec(ctx, "DELETE FROM automations WHERE id = $1", id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *repo) setEnabled(ctx context.Context, id domain.AutomationID, enabled bool) error {
	tag, err := r.pool.Exec(ctx, "UPDATE automations SET enabled=$2, updated_at=now() WHERE id=$1", id, enabled)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *repo) listRuns(ctx context.Context, id domain.AutomationID, limit int) ([]domain.AutomationRun, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id, automation_id, source_path, job_id, outcome, message, occurred_at
		 FROM automation_runs WHERE automation_id = $1
		 ORDER BY occurred_at DESC LIMIT $2`,
		id, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.AutomationRun{}
	for rows.Next() {
		var (
			r domain.AutomationRun
			job *uuid.UUID
		)
		if err := rows.Scan(&r.ID, &r.AutomationID, &r.SourcePath, &job, &r.Outcome, &r.Message, &r.OccurredAt); err != nil {
			return nil, err
		}
		if job != nil {
			j := *job
			r.JobID = &j
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// hasEnqueued reports whether the given source has already been recorded
// as `enqueued` for this automation. Used by the scanner to dedupe before
// creating a new job row.
func (r *repo) hasEnqueued(ctx context.Context, automationID domain.AutomationID, sourcePath string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM automation_runs
		 WHERE automation_id = $1 AND source_path = $2 AND outcome = $3)`,
		automationID, sourcePath, domain.OutcomeEnqueued,
	).Scan(&exists)
	return exists, err
}

type runInsert struct {
	AutomationID domain.AutomationID
	SourcePath   string
	JobID        *domain.JobID
	Outcome      string
	Message      string
}

func (r *repo) insertRun(ctx context.Context, in runInsert) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO automation_runs (automation_id, source_path, job_id, outcome, message)
		 VALUES ($1, $2, $3, $4, $5)`,
		in.AutomationID, in.SourcePath, in.JobID, in.Outcome, in.Message)
	if err != nil && isUniqueViolation(err) && in.Outcome == domain.OutcomeEnqueued {
		// dedupe index: source already enqueued for this automation; surface as skipped
		return r.insertRun(ctx, runInsert{
			AutomationID: in.AutomationID,
			SourcePath:   in.SourcePath,
			Outcome:      domain.OutcomeSkippedDupe,
			Message:      "already enqueued",
		})
	}
	return err
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scan(r rowScanner) (domain.Automation, error) {
	var a domain.Automation
	if err := r.Scan(
		&a.ID, &a.Name, &a.Enabled, &a.WatchDir, &a.Glob, &a.PresetID,
		&a.OutputDir, &a.ArchiveDir, &a.PollIntervalMs, &a.DebounceMs,
		&a.CreatedAt, &a.UpdatedAt,
	); err != nil {
		return domain.Automation{}, err
	}
	a.PollInterval = time.Duration(a.PollIntervalMs) * time.Millisecond
	return a, nil
}

func isUniqueViolation(err error) bool {
	type sqlStater interface{ SQLState() string }
	var s sqlStater
	if errors.As(err, &s) {
		return s.SQLState() == "23505"
	}
	return false
}

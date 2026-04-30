package workflows

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rustyguts/tidal/internal/domain"
)

var (
	ErrNotFound = errors.New("workflow not found")
	ErrConflict = errors.New("workflow name already exists")
)

const cols = "id, name, enabled, trigger, actions, poll_interval_ms, stable_threshold_ms, runs_count, success_count, last_run_at, created_at, updated_at"

type repo struct {
	pool *pgxpool.Pool
}

func newRepo(pool *pgxpool.Pool) *repo { return &repo{pool: pool} }

func (r *repo) list(ctx context.Context) ([]domain.Workflow, error) {
	rows, err := r.pool.Query(ctx, "SELECT "+cols+" FROM workflows ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("list workflows: %w", err)
	}
	defer rows.Close()
	out := []domain.Workflow{}
	for rows.Next() {
		w, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

func (r *repo) listEnabled(ctx context.Context) ([]domain.Workflow, error) {
	rows, err := r.pool.Query(ctx, "SELECT "+cols+" FROM workflows WHERE enabled = true ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Workflow{}
	for rows.Next() {
		w, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

func (r *repo) get(ctx context.Context, id domain.WorkflowID) (domain.Workflow, error) {
	row := r.pool.QueryRow(ctx, "SELECT "+cols+" FROM workflows WHERE id = $1", id)
	w, err := scan(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Workflow{}, ErrNotFound
	}
	return w, err
}

func (r *repo) insert(ctx context.Context, w domain.Workflow) error {
	trigJSON, err := json.Marshal(w.Trigger)
	if err != nil {
		return err
	}
	actsJSON, err := domain.MarshalActions(w.Actions)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx,
		`INSERT INTO workflows (id, name, enabled, trigger, actions, poll_interval_ms, stable_threshold_ms)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		w.ID, w.Name, w.Enabled, trigJSON, actsJSON, w.PollIntervalMs, w.StableThresholdMs,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrConflict
		}
		return fmt.Errorf("insert workflow: %w", err)
	}
	return nil
}

func (r *repo) update(ctx context.Context, w domain.Workflow) error {
	trigJSON, err := json.Marshal(w.Trigger)
	if err != nil {
		return err
	}
	actsJSON, err := domain.MarshalActions(w.Actions)
	if err != nil {
		return err
	}
	tag, err := r.pool.Exec(ctx,
		`UPDATE workflows
		 SET name=$2, enabled=$3, trigger=$4, actions=$5,
		     poll_interval_ms=$6, stable_threshold_ms=$7, updated_at=now()
		 WHERE id=$1`,
		w.ID, w.Name, w.Enabled, trigJSON, actsJSON, w.PollIntervalMs, w.StableThresholdMs,
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

func (r *repo) setEnabled(ctx context.Context, id domain.WorkflowID, enabled bool) error {
	tag, err := r.pool.Exec(ctx,
		"UPDATE workflows SET enabled=$2, updated_at=now() WHERE id=$1", id, enabled)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *repo) delete(ctx context.Context, id domain.WorkflowID) error {
	tag, err := r.pool.Exec(ctx, "DELETE FROM workflows WHERE id = $1", id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *repo) incrementRuns(ctx context.Context, id domain.WorkflowID) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE workflows SET runs_count = runs_count + 1, last_run_at = now() WHERE id = $1", id)
	return err
}

func (r *repo) incrementSuccess(ctx context.Context, id domain.WorkflowID) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE workflows SET success_count = success_count + 1 WHERE id = $1", id)
	return err
}

func (r *repo) hasEnqueued(ctx context.Context, id domain.WorkflowID, sourcePath string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM workflow_runs
		 WHERE workflow_id = $1 AND source_path = $2 AND outcome = 'enqueued')`,
		id, sourcePath,
	).Scan(&exists)
	return exists, err
}

func (r *repo) listRuns(ctx context.Context, id domain.WorkflowID, limit int) ([]domain.WorkflowRun, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id, workflow_id, source_path, job_id, outcome, message, occurred_at
		 FROM workflow_runs WHERE workflow_id = $1
		 ORDER BY occurred_at DESC LIMIT $2`,
		id, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.WorkflowRun{}
	for rows.Next() {
		var (
			r   domain.WorkflowRun
			job *uuid.UUID
		)
		if err := rows.Scan(&r.ID, &r.WorkflowID, &r.SourcePath, &job, &r.Outcome, &r.Message, &r.OccurredAt); err != nil {
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

type runInsert struct {
	WorkflowID domain.WorkflowID
	SourcePath string
	JobID      *domain.JobID
	Outcome    string
	Message    string
}

func (r *repo) insertRun(ctx context.Context, in runInsert) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO workflow_runs (workflow_id, source_path, job_id, outcome, message)
		 VALUES ($1, $2, $3, $4, $5)`,
		in.WorkflowID, in.SourcePath, in.JobID, in.Outcome, in.Message)
	if err != nil && isUniqueViolation(err) && (in.Outcome == domain.WorkflowOutcomeEnqueued || in.Outcome == domain.WorkflowOutcomeTriggered) {
		return r.insertRun(ctx, runInsert{
			WorkflowID: in.WorkflowID,
			SourcePath: in.SourcePath,
			Outcome:    domain.WorkflowOutcomeSkippedDupe,
			Message:    "already enqueued",
		})
	}
	return err
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scan(r rowScanner) (domain.Workflow, error) {
	var (
		w        domain.Workflow
		trigRaw  []byte
		actsRaw  []byte
		lastRun  *time.Time
	)
	err := r.Scan(
		&w.ID, &w.Name, &w.Enabled, &trigRaw, &actsRaw,
		&w.PollIntervalMs, &w.StableThresholdMs,
		&w.RunsCount, &w.SuccessCount, &lastRun,
		&w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		return domain.Workflow{}, err
	}
	if err := json.Unmarshal(trigRaw, &w.Trigger); err != nil {
		return domain.Workflow{}, fmt.Errorf("unmarshal trigger for %s: %w", w.ID, err)
	}
	w.Actions, err = domain.UnmarshalActions(actsRaw)
	if err != nil {
		return domain.Workflow{}, fmt.Errorf("unmarshal actions for %s: %w", w.ID, err)
	}
	if lastRun != nil {
		w.LastRunAt = lastRun
	}
	return w, nil
}

func isUniqueViolation(err error) bool {
	type sqlStater interface{ SQLState() string }
	var s sqlStater
	if errors.As(err, &s) {
		return s.SQLState() == "23505"
	}
	return false
}

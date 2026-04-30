package jobs

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
	ErrNotFound = errors.New("job not found")
	ErrInvalid  = errors.New("invalid job state")
)

const jobCols = "id, asynq_id, preset_id, source_path, output_path, cache_path, source_move_path, status, k8s_job_name, automation_id, progress, error, created_at, started_at, finished_at"

type repo struct {
	pool *pgxpool.Pool
}

func newRepo(pool *pgxpool.Pool) *repo {
	return &repo{pool: pool}
}

func (r *repo) insert(ctx context.Context, j domain.Job) error {
	progress, _ := json.Marshal(j.Progress)
	_, err := r.pool.Exec(ctx,
		`INSERT INTO jobs (id, asynq_id, preset_id, source_path, output_path, cache_path, source_move_path, status, automation_id, progress, error, created_at)
		 VALUES ($1, NULLIF($2, ''), $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		j.ID, j.AsynqID, j.PresetID, j.SourcePath, j.OutputPath, j.CachePath, j.SourceMovePath, string(j.Status),
		j.AutomationID, progress, j.Error, j.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert job: %w", err)
	}
	return nil
}

func (r *repo) get(ctx context.Context, id domain.JobID) (domain.Job, error) {
	row := r.pool.QueryRow(ctx, "SELECT "+jobCols+" FROM jobs WHERE id = $1", id)
	j, err := scanJob(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Job{}, ErrNotFound
	}
	return j, err
}

type listFilter struct {
	Status   string
	PresetID *uuid.UUID
	Limit    int
}

func (r *repo) list(ctx context.Context, f listFilter) ([]domain.Job, error) {
	q := "SELECT " + jobCols + " FROM jobs WHERE 1=1"
	args := []any{}
	if f.Status != "" {
		args = append(args, f.Status)
		q += fmt.Sprintf(" AND status = $%d", len(args))
	}
	if f.PresetID != nil {
		args = append(args, *f.PresetID)
		q += fmt.Sprintf(" AND preset_id = $%d", len(args))
	}
	q += " ORDER BY created_at DESC"
	limit := f.Limit
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	args = append(args, limit)
	q += fmt.Sprintf(" LIMIT $%d", len(args))

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list jobs: %w", err)
	}
	defer rows.Close()
	out := []domain.Job{}
	for rows.Next() {
		j, err := scanJob(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

func (r *repo) setAsynqID(ctx context.Context, id domain.JobID, asynqID string) error {
	_, err := r.pool.Exec(ctx, "UPDATE jobs SET asynq_id = $2 WHERE id = $1", id, asynqID)
	return err
}

func (r *repo) hardDelete(ctx context.Context, id domain.JobID) error {
	tag, err := r.pool.Exec(ctx, "DELETE FROM jobs WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete job: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *repo) updateStatus(ctx context.Context, id domain.JobID, status domain.JobStatus, errMsg string, finished bool) error {
	q := "UPDATE jobs SET status = $2, error = $3"
	args := []any{id, string(status), errMsg}
	if status == domain.JobRunning {
		q += ", started_at = COALESCE(started_at, now())"
	}
	if finished {
		q += ", finished_at = now()"
	}
	q += " WHERE id = $1"
	tag, err := r.pool.Exec(ctx, q, args...)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *repo) updateProgress(ctx context.Context, id domain.JobID, p domain.FFmpegProgress) error {
	pj, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("marshal progress: %w", err)
	}
	_, err = r.pool.Exec(ctx, "UPDATE jobs SET progress = $2 WHERE id = $1", id, pj)
	return err
}

// PruneOldTerminal deletes jobs in terminal states whose finished_at is older
// than `older`. Returns the number of deleted rows. ON DELETE CASCADE on
// job_logs cleans those up too.
func (r *repo) PruneOldTerminal(ctx context.Context, older time.Duration) (int64, error) {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM jobs
		 WHERE status IN ('succeeded','failed','cancelled')
		   AND finished_at IS NOT NULL
		   AND finished_at < now() - $1::interval`,
		fmt.Sprintf("%d milliseconds", older.Milliseconds()))
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (r *repo) appendLog(ctx context.Context, id domain.JobID, stream, line string) error {
	// Use a per-job sequence: pull max(seq), then insert. Single-process safe;
	// fine for the local dispatcher. Multi-process callers should use a sequence.
	_, err := r.pool.Exec(ctx,
		`INSERT INTO job_logs (job_id, seq, stream, line)
		 VALUES ($1,
		         COALESCE((SELECT MAX(seq) FROM job_logs WHERE job_id = $1), 0) + 1,
		         $2, $3)`,
		id, stream, line)
	return err
}

func (r *repo) listLogs(ctx context.Context, id domain.JobID, fromSeq int64, limit int) ([]domain.JobLog, error) {
	if limit <= 0 || limit > 1000 {
		limit = 200
	}
	rows, err := r.pool.Query(ctx,
		`SELECT job_id, seq, stream, line, emitted_at
		 FROM job_logs WHERE job_id = $1 AND seq > $2
		 ORDER BY seq ASC LIMIT $3`,
		id, fromSeq, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.JobLog{}
	for rows.Next() {
		var l domain.JobLog
		if err := rows.Scan(&l.JobID, &l.Seq, &l.Stream, &l.Line, &l.EmittedAt); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanJob(r rowScanner) (domain.Job, error) {
	var (
		j           domain.Job
		asynq       *string
		k8sName     *string
		automation  *uuid.UUID
		progressRaw []byte
		started     *time.Time
		finished    *time.Time
	)
	err := r.Scan(&j.ID, &asynq, &j.PresetID, &j.SourcePath, &j.OutputPath,
		&j.CachePath, &j.SourceMovePath,
		&j.Status, &k8sName, &automation, &progressRaw, &j.Error,
		&j.CreatedAt, &started, &finished)
	if err != nil {
		return domain.Job{}, err
	}
	if asynq != nil {
		j.AsynqID = *asynq
	}
	if k8sName != nil {
		j.K8sJobName = *k8sName
	}
	if automation != nil {
		j.AutomationID = automation
	}
	if started != nil {
		j.StartedAt = started
	}
	if finished != nil {
		j.FinishedAt = finished
	}
	if len(progressRaw) > 0 {
		_ = json.Unmarshal(progressRaw, &j.Progress)
	}
	return j, nil
}

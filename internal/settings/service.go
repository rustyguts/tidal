// Package settings stores runtime-tunable configuration in the DB so the
// operator can change things like worker concurrency without redeploying.
package settings

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	// KeyTranscodeConcurrency caps the number of ffmpeg jobs running
	// simultaneously across this worker process. Tuned via the Settings page.
	KeyTranscodeConcurrency = "transcode.concurrency"
)

type Service struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Service { return &Service{pool: pool} }

// Get returns the raw string value for key, or "" if unset.
func (s *Service) Get(ctx context.Context, key string) (string, error) {
	var v string
	err := s.pool.QueryRow(ctx, `SELECT value FROM settings WHERE key = $1`, key).Scan(&v)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("get setting %s: %w", key, err)
	}
	return v, nil
}

func (s *Service) GetInt(ctx context.Context, key string, def int) (int, error) {
	v, err := s.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	if v == "" {
		return def, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def, nil
	}
	return n, nil
}

func (s *Service) Set(ctx context.Context, key, value string) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO settings (key, value) VALUES ($1, $2)
		 ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = now()`,
		key, value)
	if err != nil {
		return fmt.Errorf("set setting %s: %w", key, err)
	}
	return nil
}

func (s *Service) SetInt(ctx context.Context, key string, n int) error {
	return s.Set(ctx, key, strconv.Itoa(n))
}

// All returns every setting row. Used by the settings handler to render the page.
func (s *Service) All(ctx context.Context) (map[string]string, error) {
	rows, err := s.pool.Query(ctx, `SELECT key, value FROM settings`)
	if err != nil {
		return nil, fmt.Errorf("list settings: %w", err)
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		out[k] = v
	}
	return out, rows.Err()
}

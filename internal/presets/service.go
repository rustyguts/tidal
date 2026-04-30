package presets

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
	ErrNotFound = errors.New("preset not found")
	ErrConflict = errors.New("preset name already exists")
)

type Service struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

const (
	colSelect = "id, name, description, builtin, spec, created_at, updated_at"
)

func (s *Service) List(ctx context.Context) ([]domain.Preset, error) {
	rows, err := s.pool.Query(ctx, "SELECT "+colSelect+" FROM presets ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("list presets: %w", err)
	}
	defer rows.Close()

	out := []domain.Preset{}
	for rows.Next() {
		p, err := scanPreset(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Service) Get(ctx context.Context, id domain.PresetID) (domain.Preset, error) {
	row := s.pool.QueryRow(ctx, "SELECT "+colSelect+" FROM presets WHERE id = $1", id)
	p, err := scanPreset(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Preset{}, ErrNotFound
	}
	return p, err
}

func (s *Service) GetByName(ctx context.Context, name string) (domain.Preset, error) {
	row := s.pool.QueryRow(ctx, "SELECT "+colSelect+" FROM presets WHERE name = $1", name)
	p, err := scanPreset(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Preset{}, ErrNotFound
	}
	return p, err
}

type CreateInput struct {
	Name        string
	Description string
	Builtin     bool
	Spec        domain.PresetSpec
}

func (s *Service) Create(ctx context.Context, in CreateInput) (domain.Preset, error) {
	if err := in.Spec.Validate(); err != nil {
		return domain.Preset{}, err
	}
	specJSON, err := json.Marshal(in.Spec)
	if err != nil {
		return domain.Preset{}, fmt.Errorf("marshal spec: %w", err)
	}
	id := uuid.New()
	now := time.Now().UTC()

	_, err = s.pool.Exec(ctx,
		`INSERT INTO presets (id, name, description, builtin, spec, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $6)`,
		id, in.Name, in.Description, in.Builtin, specJSON, now)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.Preset{}, ErrConflict
		}
		return domain.Preset{}, fmt.Errorf("insert preset: %w", err)
	}
	return s.Get(ctx, id)
}

type UpdateInput struct {
	Name        *string
	Description *string
	Spec        *domain.PresetSpec
}

func (s *Service) Update(ctx context.Context, id domain.PresetID, in UpdateInput) (domain.Preset, error) {
	cur, err := s.Get(ctx, id)
	if err != nil {
		return domain.Preset{}, err
	}
	if in.Name != nil {
		cur.Name = *in.Name
	}
	if in.Description != nil {
		cur.Description = *in.Description
	}
	if in.Spec != nil {
		if err := in.Spec.Validate(); err != nil {
			return domain.Preset{}, err
		}
		cur.Spec = *in.Spec
	}
	specJSON, err := json.Marshal(cur.Spec)
	if err != nil {
		return domain.Preset{}, fmt.Errorf("marshal spec: %w", err)
	}
	now := time.Now().UTC()
	tag, err := s.pool.Exec(ctx,
		`UPDATE presets SET name = $2, description = $3, spec = $4, updated_at = $5 WHERE id = $1`,
		id, cur.Name, cur.Description, specJSON, now)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.Preset{}, ErrConflict
		}
		return domain.Preset{}, fmt.Errorf("update preset: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.Preset{}, ErrNotFound
	}
	return s.Get(ctx, id)
}

// Delete removes a preset. Builtin presets may be deleted; the seeded_presets
// marker keeps the name from being resurrected on next boot.
func (s *Service) Delete(ctx context.Context, id domain.PresetID) error {
	tag, err := s.pool.Exec(ctx, "DELETE FROM presets WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete preset: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Duplicate clones the preset with a new name (caller-provided).
func (s *Service) Duplicate(ctx context.Context, id domain.PresetID, newName string) (domain.Preset, error) {
	cur, err := s.Get(ctx, id)
	if err != nil {
		return domain.Preset{}, err
	}
	return s.Create(ctx, CreateInput{
		Name:        newName,
		Description: cur.Description,
		Builtin:     false,
		Spec:        cur.Spec,
	})
}

// rowScanner abstracts pgx.Row vs pgx.Rows.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanPreset(r rowScanner) (domain.Preset, error) {
	var (
		p        domain.Preset
		specRaw  []byte
		descNull *string
	)
	err := r.Scan(&p.ID, &p.Name, &descNull, &p.Builtin, &specRaw, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return domain.Preset{}, err
	}
	if descNull != nil {
		p.Description = *descNull
	}
	if err := json.Unmarshal(specRaw, &p.Spec); err != nil {
		return domain.Preset{}, fmt.Errorf("unmarshal spec for %s: %w", p.ID, err)
	}
	return p, nil
}

func isUniqueViolation(err error) bool {
	// pgx returns *pgconn.PgError; check SQLSTATE 23505.
	type sqlStater interface{ SQLState() string }
	var s sqlStater
	if errors.As(err, &s) {
		return s.SQLState() == "23505"
	}
	return false
}

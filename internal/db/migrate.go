package db

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// migrationsRoot points at the subdir inside the embed.
const migrationsRoot = "migrations"

func newMigrate(dbURL string) (*migrate.Migrate, error) {
	sub, err := fs.Sub(migrationsFS, migrationsRoot)
	if err != nil {
		return nil, fmt.Errorf("migrations sub-fs: %w", err)
	}
	src, err := iofs.New(sub, ".")
	if err != nil {
		return nil, fmt.Errorf("migrations source: %w", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", src, migrateURL(dbURL))
	if err != nil {
		return nil, fmt.Errorf("migrate init: %w", err)
	}
	return m, nil
}

// migrateURL rewrites a libpq-style postgres:// URL to the pgx5:// scheme expected
// by the golang-migrate pgx/v5 driver. Other schemes pass through unchanged.
func migrateURL(dbURL string) string {
	for _, prefix := range []string{"postgres://", "postgresql://"} {
		if strings.HasPrefix(dbURL, prefix) {
			return "pgx5://" + strings.TrimPrefix(dbURL, prefix)
		}
	}
	return dbURL
}

func Up(dbURL string) error {
	m, err := newMigrate(dbURL)
	if err != nil {
		return err
	}
	defer closeMigrate(m)
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}

func Down(dbURL string, steps int) error {
	m, err := newMigrate(dbURL)
	if err != nil {
		return err
	}
	defer closeMigrate(m)
	if steps <= 0 {
		steps = 1
	}
	if err := m.Steps(-steps); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate down: %w", err)
	}
	return nil
}

type Status struct {
	Version uint
	Dirty   bool
}

func CurrentVersion(dbURL string) (Status, error) {
	m, err := newMigrate(dbURL)
	if err != nil {
		return Status{}, err
	}
	defer closeMigrate(m)
	v, dirty, err := m.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			return Status{}, nil
		}
		return Status{}, fmt.Errorf("version: %w", err)
	}
	return Status{Version: v, Dirty: dirty}, nil
}

func closeMigrate(m *migrate.Migrate) {
	srcErr, dbErr := m.Close()
	_ = srcErr
	_ = dbErr
}

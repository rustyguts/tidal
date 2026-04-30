package automations

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"

	"github.com/rustyguts/tidal/internal/domain"
)

// ArchiveSource moves the source file to the automation's archive_dir. Used by
// the post-transcode hook in jobs.Service when a job has automation_id set.
// On rename failure across mounts, falls back to copy + delete.
func (s *Service) ArchiveSource(ctx context.Context, automationID domain.AutomationID, sourcePath string) error {
	a, err := s.repo.get(ctx, automationID)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(a.ArchiveDir, 0o755); err != nil {
		return fmt.Errorf("mkdir archive: %w", err)
	}
	dest := filepath.Join(a.ArchiveDir, filepath.Base(sourcePath))

	if err := os.Rename(sourcePath, dest); err != nil {
		// Cross-device fallback: copy + delete.
		var linkErr *os.LinkError
		if !errors.As(err, &linkErr) {
			return fmt.Errorf("rename: %w", err)
		}
		if err := copyFile(sourcePath, dest); err != nil {
			return err
		}
		if err := os.Remove(sourcePath); err != nil {
			return fmt.Errorf("remove source after copy: %w", err)
		}
	}
	log.Info().Str("automation", a.Name).Str("source", sourcePath).Str("dest", dest).Msg("archived")
	_ = s.RecordRun(ctx, a.ID, sourcePath, domain.OutcomeArchived, dest, nil)
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open src: %w", err)
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("open dst: %w", err)
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copy: %w", err)
	}
	return out.Sync()
}

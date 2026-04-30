package presets

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/rustyguts/tidal/internal/domain"
)

// Builtin returns the canonical list of seeded presets.
// Adding to this list reseeds on next boot for any name not yet planted.
// Removing or renaming a builtin does NOT delete existing rows — names
// stay in seeded_presets so the user can curate as they wish.
func Builtin() []CreateInput {
	res1080 := &domain.Resolution{Width: 1920, Height: 1080}
	res4k := &domain.Resolution{Width: 3840, Height: 2160}

	return []CreateInput{
		{
			Name:        "h264-1080p",
			Description: "H.264 1080p, slow preset, AAC 192k — broad compatibility",
			Builtin:     true,
			Spec: domain.PresetSpec{
				Container:    "mp4",
				VideoCodec:   "libx264",
				VideoPreset:  "slow",
				CRF:          20,
				AudioCodec:   "aac",
				AudioBitrate: "192k",
				OutputSuffix: "_1080p",
				Resolution:   res1080,
			},
		},
		{
			Name:        "h264-4k",
			Description: "H.264 2160p (4K), slow preset, AAC 192k",
			Builtin:     true,
			Spec: domain.PresetSpec{
				Container:    "mp4",
				VideoCodec:   "libx264",
				VideoPreset:  "slow",
				CRF:          20,
				AudioCodec:   "aac",
				AudioBitrate: "192k",
				OutputSuffix: "_4k",
				Resolution:   res4k,
			},
		},
		{
			Name:        "h265-1080p",
			Description: "HEVC 1080p, medium preset, AAC 160k — smaller files",
			Builtin:     true,
			Spec: domain.PresetSpec{
				Container:    "mkv",
				VideoCodec:   "libx265",
				VideoPreset:  "medium",
				CRF:          22,
				AudioCodec:   "aac",
				AudioBitrate: "160k",
				OutputSuffix: "_1080p_h265",
				Resolution:   res1080,
			},
		},
		{
			Name:        "av1-1080p",
			Description: "AV1 (svt-av1) 1080p, preset 4, AAC 160k, MP4 container",
			Builtin:     true,
			Spec: domain.PresetSpec{
				Container:    "mp4",
				VideoCodec:   "libsvtav1",
				VideoPreset:  "4",
				CRF:          30,
				AudioCodec:   "aac",
				AudioBitrate: "160k",
				OutputSuffix: "_1080p_av1",
				Resolution:   res1080,
			},
		},
		{
			Name:        "audio-only-aac",
			Description: "Strip video, output AAC audio at 192k",
			Builtin:     true,
			Spec: domain.PresetSpec{
				Container:    "mp4",
				VideoCodec:   "copy",
				CRF:          0,
				AudioCodec:   "aac",
				AudioBitrate: "192k",
				ExtraArgs:    []string{"-vn"},
				OutputSuffix: "_audio",
			},
		},
	}
}

// Seed inserts any builtin preset whose name is not yet recorded in seeded_presets.
// Already-planted names are skipped (even if the preset row was deleted by the user)
// so deletions are persistent across restarts.
func (s *Service) Seed(ctx context.Context) (planted []string, err error) {
	for _, b := range Builtin() {
		ok, err := s.markIfAbsent(ctx, b.Name)
		if err != nil {
			return planted, fmt.Errorf("mark %s: %w", b.Name, err)
		}
		if !ok {
			continue
		}
		if _, err := s.Create(ctx, b); err != nil {
			if errors.Is(err, ErrConflict) {
				// Row already exists despite missing marker — record marker, move on.
				continue
			}
			return planted, fmt.Errorf("create builtin %s: %w", b.Name, err)
		}
		planted = append(planted, b.Name)
	}
	return planted, nil
}

// markIfAbsent inserts into seeded_presets if name is missing. Returns true if
// inserted (caller should plant the preset).
func (s *Service) markIfAbsent(ctx context.Context, name string) (bool, error) {
	tag, err := s.pool.Exec(ctx,
		`INSERT INTO seeded_presets (name) VALUES ($1) ON CONFLICT (name) DO NOTHING`,
		name)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}

// RestoreDefaults removes seeded_presets markers for builtin names whose preset row
// is missing, then re-seeds. Returns the names that were re-planted.
func (s *Service) RestoreDefaults(ctx context.Context) ([]string, error) {
	builtins := Builtin()
	missingNames := make([]string, 0, len(builtins))
	for _, b := range builtins {
		_, err := s.GetByName(ctx, b.Name)
		if errors.Is(err, ErrNotFound) {
			missingNames = append(missingNames, b.Name)
			continue
		}
		if err != nil {
			return nil, err
		}
	}
	if len(missingNames) == 0 {
		return nil, nil
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, "DELETE FROM seeded_presets WHERE name = ANY($1)", missingNames); err != nil {
		return nil, fmt.Errorf("clear markers: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.Seed(ctx)
}

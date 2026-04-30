package builder

import "github.com/rustyguts/tidal/internal/domain"

func subtitles(_ Context, spec domain.PresetSpecV2) ([]string, error) {
	switch spec.Subtitles.Mode {
	case domain.SubtitleStrip:
		return []string{"-sn"}, nil
	case domain.SubtitleCopy:
		return []string{"-c:s", "copy"}, nil
	case domain.SubtitleBurn:
		// Burn-in is implemented as a `subtitles=` video filter the user
		// wires into spec.Filters.Video — this section emits nothing.
		return nil, nil
	}
	return nil, nil
}

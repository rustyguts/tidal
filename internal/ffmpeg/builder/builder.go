// Package builder composes the ffmpeg argv from a PresetSpec. Each section
// (input, video, audio, filters...) is a small pure function. The top-level
// Compose runs them in a fixed order. Validation is a separate concern (see
// internal/domain/preset_v2_validate.go) — Compose trusts its input.
package builder

import (
	"github.com/rustyguts/tidal/internal/domain"
)

// Context is the per-invocation environment for argv composition: paths,
// progress destination, and two-pass state.
type Context struct {
	InputPath     string
	OutputPath    string
	ProgressURL   string // if empty, no -progress flag
	PassLogPrefix string // base path for ffmpeg2pass logs (used when TwoPass)
	Pass          int    // 0 = single pass, 1 or 2 = two-pass run number
}

// Section emits a slice of argv tokens for one logical group of flags. Sections
// returning errors abort the whole compose.
type Section func(ctx Context, spec domain.PresetSpec) ([]string, error)

// orderedSections is the canonical compose order. Section ordering matters:
// global flags must precede -i, -i must precede output options, output options
// (-c:v, -vf, -map ...) must precede the output path.
var orderedSections = []Section{
	global,
	hwaccelInput,
	input,
	mapping,
	filters,
	video,
	audio,
	subtitles,
	container,
	rawExtras,
	output,
}

// Compose runs every section in order and concatenates the resulting tokens.
func Compose(ctx Context, spec domain.PresetSpec) ([]string, error) {
	out := make([]string, 0, 64)
	for _, s := range orderedSections {
		toks, err := s(ctx, spec)
		if err != nil {
			return nil, err
		}
		out = append(out, toks...)
	}
	return out, nil
}

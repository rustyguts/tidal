package builder

import (
	"strconv"

	"github.com/rustyguts/tidal/internal/domain"
)

func global(ctx Context, spec domain.PresetSpecV2) ([]string, error) {
	out := []string{"-y", "-hide_banner", "-nostats", "-loglevel", "info"}
	if spec.Threading.Threads > 0 {
		out = append(out, "-threads", strconv.Itoa(spec.Threading.Threads))
	}
	if spec.Threading.FilterThreads > 0 {
		out = append(out, "-filter_threads", strconv.Itoa(spec.Threading.FilterThreads))
	}
	if ctx.ProgressURL != "" {
		out = append(out, "-progress", ctx.ProgressURL)
	}
	return out, nil
}

package builder

import (
	"os"

	"github.com/rustyguts/tidal/internal/domain"
)

func output(ctx Context, spec domain.PresetSpecV2) ([]string, error) {
	// Pass 1 of two-pass writes to /dev/null with the null muxer.
	if spec.Video.TwoPass && ctx.Pass == 1 {
		return []string{"-an", "-f", "null", os.DevNull}, nil
	}
	if ctx.OutputPath == "" {
		return nil, nil
	}
	return []string{ctx.OutputPath}, nil
}

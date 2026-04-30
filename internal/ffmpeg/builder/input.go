package builder

import (
	"fmt"
	"strings"

	"github.com/rustyguts/tidal/internal/domain"
)

func input(ctx Context, spec domain.PresetSpecV2) ([]string, error) {
	out := []string{}
	if ss := strings.TrimSpace(spec.Input.SeekStart); ss != "" {
		out = append(out, "-ss", ss)
	}
	if d := strings.TrimSpace(spec.Input.Duration); d != "" {
		out = append(out, "-t", d)
	}
	if spec.Input.ReadRateLimit > 0 {
		out = append(out, "-readrate", fmt.Sprintf("%v", spec.Input.ReadRateLimit))
	}
	if len(spec.Input.ProtocolWhitelist) > 0 {
		out = append(out, "-protocol_whitelist", strings.Join(spec.Input.ProtocolWhitelist, ","))
	}
	if ctx.InputPath != "" {
		out = append(out, "-i", ctx.InputPath)
	}
	return out, nil
}

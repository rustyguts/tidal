package builder

import (
	"fmt"
	"strings"

	"github.com/rustyguts/tidal/internal/domain"
)

func container(_ Context, spec domain.PresetSpecV2) ([]string, error) {
	out := []string{}
	flags := uniqueMovFlags(spec.Container.Faststart, spec.Container.MovFlags)
	if len(flags) > 0 && (spec.Container.Format == "mp4" || spec.Container.Format == "mov") {
		out = append(out, "-movflags", strings.Join(flags, ""))
	}
	if spec.Container.FragmentDurMs > 0 {
		out = append(out, "-frag_duration", fmt.Sprintf("%d", spec.Container.FragmentDurMs*1000))
	}
	return out, nil
}

func uniqueMovFlags(faststart bool, extra []string) []string {
	seen := map[string]struct{}{}
	out := []string{}
	add := func(f string) {
		if f == "" {
			return
		}
		if _, ok := seen[f]; ok {
			return
		}
		seen[f] = struct{}{}
		out = append(out, f)
	}
	if faststart {
		add("+faststart")
	}
	for _, f := range extra {
		add(f)
	}
	return out
}

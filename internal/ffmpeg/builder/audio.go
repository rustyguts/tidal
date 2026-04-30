package builder

import (
	"strconv"
	"strings"

	"github.com/rustyguts/tidal/internal/domain"
)

func audio(_ Context, spec domain.PresetSpec) ([]string, error) {
	if spec.Audio.Disabled {
		return []string{"-an"}, nil
	}
	if spec.Audio.Codec == "" {
		return nil, nil
	}
	out := []string{"-c:a", spec.Audio.Codec}
	if spec.Audio.Codec == "copy" {
		return out, nil
	}
	if b := strings.TrimSpace(spec.Audio.Bitrate); b != "" {
		out = append(out, "-b:a", b)
	}
	if spec.Audio.SampleRate > 0 {
		out = append(out, "-ar", strconv.Itoa(spec.Audio.SampleRate))
	}
	if spec.Audio.Channels > 0 {
		out = append(out, "-ac", strconv.Itoa(spec.Audio.Channels))
	}
	if p := strings.TrimSpace(spec.Audio.Profile); p != "" {
		out = append(out, "-profile:a", p)
	}
	if spec.Audio.VBRQuality != nil {
		// libopus uses -vbr on; libmp3lame/aac use -q:a. Caller chose codec;
		// emit -q:a as the broadly-applicable form. Codec-specific VBR
		// switches can be added via rawExtras.
		out = append(out, "-q:a", strconv.Itoa(*spec.Audio.VBRQuality))
	}
	return out, nil
}

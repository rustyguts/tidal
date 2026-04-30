package builder

import "github.com/rustyguts/tidal/internal/domain"

// mapping emits explicit -map tokens. Defaults to ffmpeg's automatic stream
// selection when nothing is specified — that's the desired behavior for the
// majority of presets.
func mapping(_ Context, spec domain.PresetSpecV2) ([]string, error) {
	if len(spec.Mapping.Streams) == 0 && spec.Mapping.Video == "" && spec.Mapping.Audio == "" && spec.Mapping.Subtitle == "" {
		return nil, nil
	}
	out := []string{}
	for _, s := range spec.Mapping.Streams {
		out = append(out, "-map", s)
	}
	if v := spec.Mapping.Video; v != "" && v != "auto" {
		if v == "none" {
			out = append(out, "-map", "-0:v")
		} else {
			out = append(out, "-map", v)
		}
	}
	if a := spec.Mapping.Audio; a != "" && a != "auto" {
		if a == "none" {
			out = append(out, "-map", "-0:a")
		} else {
			out = append(out, "-map", a)
		}
	}
	if s := spec.Mapping.Subtitle; s != "" && s != "auto" {
		if s == "none" {
			out = append(out, "-map", "-0:s")
		} else {
			out = append(out, "-map", s)
		}
	}
	return out, nil
}

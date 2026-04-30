package builder

import (
	"sort"
	"strconv"
	"strings"

	"github.com/rustyguts/tidal/internal/domain"
)

// filters merges implicit (resolution, frameRate) and explicit filter chains
// into a single -vf and/or -af argument. ffmpeg only honors one -vf per
// output, so synthesizing into a single chain is mandatory.
func filters(_ Context, spec domain.PresetSpec) ([]string, error) {
	vchain := buildVideoChain(spec)
	achain := buildAudioChain(spec)
	out := []string{}
	if rendered := renderChain(vchain); rendered != "" {
		out = append(out, "-vf", rendered)
	}
	if rendered := renderChain(achain); rendered != "" {
		out = append(out, "-af", rendered)
	}
	return out, nil
}

func buildVideoChain(spec domain.PresetSpec) []domain.FilterStep {
	if spec.Video.Codec == "copy" {
		return nil
	}
	chain := []domain.FilterStep{}
	if spec.Video.Resolution != nil {
		chain = append(chain, domain.FilterStep{
			Name: "scale",
			Args: map[string]string{
				"w": strconv.Itoa(spec.Video.Resolution.Width),
				"h": strconv.Itoa(spec.Video.Resolution.Height),
			},
		})
	}
	if fr := strings.TrimSpace(spec.Video.FrameRate); fr != "" {
		chain = append(chain, domain.FilterStep{Name: "fps", Args: map[string]string{"fps": fr}})
	}
	for _, s := range spec.Filters.Video {
		if s.IsEnabled() {
			chain = append(chain, s)
		}
	}
	return chain
}

func buildAudioChain(spec domain.PresetSpec) []domain.FilterStep {
	if spec.Audio.Codec == "copy" || spec.Audio.Disabled {
		return nil
	}
	chain := []domain.FilterStep{}
	for _, s := range spec.Filters.Audio {
		if s.IsEnabled() {
			chain = append(chain, s)
		}
	}
	return chain
}

func renderChain(chain []domain.FilterStep) string {
	if len(chain) == 0 {
		return ""
	}
	parts := make([]string, 0, len(chain))
	for _, step := range chain {
		parts = append(parts, renderFilter(step))
	}
	return strings.Join(parts, ",")
}

func renderFilter(step domain.FilterStep) string {
	// scale with only w+h: render positionally so the output matches the
	// idiomatic ffmpeg form (scale=W:H) rather than the named form
	// (scale=w=W:h=H). Both are valid; positional is the more common
	// hand-written form.
	if step.Name == "scale" && len(step.Args) > 0 && len(step.Args) <= 2 {
		w, hasW := step.Args["w"]
		h, hasH := step.Args["h"]
		if hasW && hasH && len(step.Args) == 2 {
			return "scale=" + escapeFilterArg(w) + ":" + escapeFilterArg(h)
		}
	}
	if len(step.Args) == 0 {
		return step.Name
	}
	keys := make([]string, 0, len(step.Args))
	for k := range step.Args {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+escapeFilterArg(step.Args[k]))
	}
	return step.Name + "=" + strings.Join(parts, ":")
}

// escapeFilterArg quotes a value that contains characters with special meaning
// in ffmpeg's filter graph language. ffmpeg uses a layered escape scheme: the
// graph parser uses `\` for escaping in unquoted values, and `'`-quoted values
// escape inner `'` as `\'`. We use single-quotes when needed.
func escapeFilterArg(v string) string {
	if !needsFilterQuote(v) {
		return v
	}
	// Escape backslash first, then single quote.
	v = strings.ReplaceAll(v, `\`, `\\`)
	v = strings.ReplaceAll(v, `'`, `\'`)
	return "'" + v + "'"
}

func needsFilterQuote(v string) bool {
	for _, r := range v {
		switch r {
		case ':', ',', '\\', '\'', '[', ']', ';':
			return true
		}
	}
	return false
}

package domain

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/rustyguts/tidal/internal/ffmpeg/catalog"
)

// ValidateOpts adjusts validator strictness. The default (zero value) is
// strict: raw extras must match the allowlist.
type ValidateOpts struct {
	// PermissiveRawExtras switches the raw-extras check from allowlist mode
	// to denylist-only — any token not explicitly forbidden is accepted.
	// Intended for power users running self-hosted instances; controlled by
	// TIDAL_PRESET_RAW_EXTRAS_PERMISSIVE.
	PermissiveRawExtras bool
}

// ValidateV2 checks every field of a PresetSpecV2 against the catalog and
// raw-extras allow/deny rules. Errors are returned as a flat joined message;
// callers may wrap as needed.
func ValidateV2(s PresetSpecV2, cat *catalog.Catalog, opts ValidateOpts) error {
	if cat == nil {
		return errors.New("validate: catalog is nil")
	}
	var errs []string

	errs = append(errs, validateContainer(s.Container, cat)...)
	errs = append(errs, validateVideo(s.Video, cat)...)
	errs = append(errs, validateAudio(s.Audio, cat)...)
	errs = append(errs, validateHwaccel(s.Hwaccel, s.Video, cat)...)
	errs = append(errs, validateFilters(s.Filters, cat)...)
	errs = append(errs, validateSubtitles(s.Subtitles)...)
	errs = append(errs, validateRawExtras(s.RawExtras, cat, opts)...)

	if len(errs) == 0 {
		return nil
	}
	return errors.New(strings.Join(errs, "; "))
}

func validateContainer(c ContainerSpec, cat *catalog.Catalog) []string {
	if strings.TrimSpace(c.Format) == "" {
		return []string{"container.format required"}
	}
	if _, ok := cat.Container(c.Format); !ok {
		return []string{fmt.Sprintf("container.format %q not in catalog", c.Format)}
	}
	return nil
}

func validateVideo(v VideoSpec, cat *catalog.Catalog) []string {
	if v.Disabled {
		return nil
	}
	if strings.TrimSpace(v.Codec) == "" {
		return []string{"video.codec required (or set video.disabled)"}
	}
	codec, ok := cat.VideoCodec(v.Codec)
	if !ok {
		return []string{fmt.Sprintf("video.codec %q not in catalog", v.Codec)}
	}
	if codec.Family == catalog.FamilyPassthru {
		// Skip remaining checks for stream-copy.
		return nil
	}

	var errs []string
	if v.Preset != "" && len(codec.Presets) > 0 && !contains(codec.Presets, v.Preset) {
		errs = append(errs, fmt.Sprintf("video.preset %q not valid for codec %q", v.Preset, v.Codec))
	}
	if v.Tune != "" && len(codec.Tunes) > 0 && !contains(codec.Tunes, v.Tune) {
		errs = append(errs, fmt.Sprintf("video.tune %q not valid for codec %q", v.Tune, v.Codec))
	}
	if v.Profile != "" && len(codec.Profiles) > 0 && !contains(codec.Profiles, v.Profile) {
		errs = append(errs, fmt.Sprintf("video.profile %q not valid for codec %q", v.Profile, v.Codec))
	}
	if v.Level != "" && len(codec.Levels) > 0 && !contains(codec.Levels, v.Level) {
		errs = append(errs, fmt.Sprintf("video.level %q not valid for codec %q", v.Level, v.Codec))
	}
	if v.PixelFormat != "" {
		if len(codec.PixelFormats) > 0 && !contains(codec.PixelFormats, v.PixelFormat) {
			errs = append(errs, fmt.Sprintf("video.pixelFormat %q not valid for codec %q", v.PixelFormat, v.Codec))
		} else if len(codec.PixelFormats) == 0 && !contains(cat.PixelFormats, v.PixelFormat) {
			errs = append(errs, fmt.Sprintf("video.pixelFormat %q not in catalog", v.PixelFormat))
		}
	}

	if v.Rate.Mode != "" && v.Rate.Mode != RateModeNone {
		mode := catalog.RateMode(v.Rate.Mode)
		valid := false
		for _, m := range codec.RateModes {
			if m == mode {
				valid = true
				break
			}
		}
		if !valid {
			errs = append(errs, fmt.Sprintf("video.rate.mode %q not supported by codec %q", v.Rate.Mode, v.Codec))
		}
	}
	if v.Rate.CRF != nil && codec.CRFRange != nil {
		if *v.Rate.CRF < codec.CRFRange.Min || *v.Rate.CRF > codec.CRFRange.Max {
			errs = append(errs, fmt.Sprintf("video.rate.crf %d out of range %d..%d for %q", *v.Rate.CRF, codec.CRFRange.Min, codec.CRFRange.Max, v.Codec))
		}
	}
	if v.Rate.QP != nil && codec.QPRange != nil {
		if *v.Rate.QP < codec.QPRange.Min || *v.Rate.QP > codec.QPRange.Max {
			errs = append(errs, fmt.Sprintf("video.rate.qp %d out of range %d..%d for %q", *v.Rate.QP, codec.QPRange.Min, codec.QPRange.Max, v.Codec))
		}
	}

	if v.TwoPass && !codec.AllowsTwoPass {
		errs = append(errs, fmt.Sprintf("video.twoPass not supported by codec %q", v.Codec))
	}

	if v.Resolution != nil {
		if v.Resolution.Width <= 0 || v.Resolution.Width%2 != 0 {
			errs = append(errs, "video.resolution.width must be positive and even")
		}
		if v.Resolution.Height <= 0 || v.Resolution.Height%2 != 0 {
			errs = append(errs, "video.resolution.height must be positive and even")
		}
	}
	return errs
}

func validateAudio(a AudioSpec, cat *catalog.Catalog) []string {
	if a.Disabled || a.Codec == "" {
		return nil
	}
	codec, ok := cat.AudioCodec(a.Codec)
	if !ok {
		return []string{fmt.Sprintf("audio.codec %q not in catalog", a.Codec)}
	}
	if a.Codec == "copy" {
		return nil
	}
	var errs []string
	if a.SampleRate > 0 && len(codec.SampleRates) > 0 {
		ok := false
		for _, sr := range codec.SampleRates {
			if sr == a.SampleRate {
				ok = true
				break
			}
		}
		if !ok {
			errs = append(errs, fmt.Sprintf("audio.sampleRate %d not valid for codec %q", a.SampleRate, a.Codec))
		}
	}
	if a.Channels > 0 && len(codec.Channels) > 0 {
		ok := false
		for _, ch := range codec.Channels {
			if ch == a.Channels {
				ok = true
				break
			}
		}
		if !ok {
			errs = append(errs, fmt.Sprintf("audio.channels %d not valid for codec %q", a.Channels, a.Codec))
		}
	}
	if a.Profile != "" && len(codec.Profiles) > 0 && !contains(codec.Profiles, a.Profile) {
		errs = append(errs, fmt.Sprintf("audio.profile %q not valid for codec %q", a.Profile, a.Codec))
	}
	if a.VBRQuality != nil && codec.VBRQuality != nil {
		if *a.VBRQuality < codec.VBRQuality.Min || *a.VBRQuality > codec.VBRQuality.Max {
			errs = append(errs, fmt.Sprintf("audio.vbrQuality %d out of range %d..%d for %q", *a.VBRQuality, codec.VBRQuality.Min, codec.VBRQuality.Max, a.Codec))
		}
	}
	return errs
}

func validateHwaccel(h *HwaccelSpec, v VideoSpec, cat *catalog.Catalog) []string {
	if h == nil || h.Type == "" || h.Type == "none" {
		return nil
	}
	hw, ok := cat.Hwaccel(h.Type)
	if !ok {
		return []string{fmt.Sprintf("hwaccel.type %q not in catalog", h.Type)}
	}
	if v.Codec == "" || v.Disabled {
		return nil
	}
	if len(hw.CodecNames) > 0 && !contains(hw.CodecNames, v.Codec) {
		return []string{fmt.Sprintf("hwaccel %q does not pair with video.codec %q", h.Type, v.Codec)}
	}
	return nil
}

func validateFilters(f FilterChain, cat *catalog.Catalog) []string {
	var errs []string
	for _, step := range f.Video {
		if errs = append(errs, validateFilterStep(step, catalog.FilterVideo, cat)...); len(errs) > 50 {
			break
		}
	}
	for _, step := range f.Audio {
		if errs = append(errs, validateFilterStep(step, catalog.FilterAudio, cat)...); len(errs) > 50 {
			break
		}
	}
	return errs
}

func validateFilterStep(step FilterStep, scope catalog.FilterScope, cat *catalog.Catalog) []string {
	if strings.TrimSpace(step.Name) == "" {
		return []string{"filter step missing name"}
	}
	spec, ok := cat.Filter(step.Name)
	if !ok {
		return []string{fmt.Sprintf("filter %q not in catalog", step.Name)}
	}
	if spec.Scope != scope {
		return []string{fmt.Sprintf("filter %q has scope %q, used in %q chain", step.Name, spec.Scope, scope)}
	}
	var errs []string
	known := make(map[string]catalog.FilterArg, len(spec.Args))
	for _, a := range spec.Args {
		known[a.Name] = a
	}
	for k := range step.Args {
		if _, ok := known[k]; !ok {
			errs = append(errs, fmt.Sprintf("filter %q: unknown arg %q", step.Name, k))
		}
	}
	for _, a := range spec.Args {
		if !a.Required {
			continue
		}
		if _, present := step.Args[a.Name]; !present {
			errs = append(errs, fmt.Sprintf("filter %q: required arg %q missing", step.Name, a.Name))
		}
	}
	return errs
}

func validateSubtitles(s SubtitleSpec) []string {
	switch s.Mode {
	case "", SubtitleCopy, SubtitleBurn, SubtitleStrip:
		return nil
	}
	return []string{fmt.Sprintf("subtitles.mode %q invalid", s.Mode)}
}

func validateRawExtras(extras []string, cat *catalog.Catalog, opts ValidateOpts) []string {
	if len(extras) == 0 {
		return nil
	}
	allow, errAllow := compileRegexes(cat.RawExtrasAllow)
	deny, errDeny := compileRegexes(cat.RawExtrasDeny)
	if errAllow != nil || errDeny != nil {
		return []string{"rawExtras: catalog regex compile error"}
	}
	valueDeny, _ := compileRegexes(catalog.ValueDeny)

	var errs []string
	expectingValue := false
	for _, tok := range extras {
		if expectingValue {
			expectingValue = false
			if len(tok) > catalog.ValueMaxLen {
				errs = append(errs, fmt.Sprintf("rawExtras value too long (>%d): %q", catalog.ValueMaxLen, truncate(tok, 32)))
				continue
			}
			for _, re := range valueDeny {
				if re.MatchString(tok) {
					errs = append(errs, fmt.Sprintf("rawExtras value contains forbidden character: %q", truncate(tok, 64)))
					break
				}
			}
			continue
		}
		if !strings.HasPrefix(tok, "-") {
			errs = append(errs, fmt.Sprintf("rawExtras: %q is not a flag (expected leading -)", truncate(tok, 32)))
			continue
		}
		denied := false
		for _, re := range deny {
			if re.MatchString(tok) {
				errs = append(errs, fmt.Sprintf("rawExtras: flag %q is denylisted", tok))
				denied = true
				break
			}
		}
		if denied {
			expectingValue = true
			continue
		}
		if !opts.PermissiveRawExtras {
			matched := false
			for _, re := range allow {
				if re.MatchString(tok) {
					matched = true
					break
				}
			}
			if !matched {
				errs = append(errs, fmt.Sprintf("rawExtras: flag %q not in allowlist", tok))
			}
		}
		expectingValue = true
	}
	return errs
}

func compileRegexes(patterns []string) ([]*regexp.Regexp, error) {
	out := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			return nil, fmt.Errorf("compile %q: %w", p, err)
		}
		out = append(out, re)
	}
	return out, nil
}

func contains(haystack []string, needle string) bool {
	for _, x := range haystack {
		if x == needle {
			return true
		}
	}
	return false
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

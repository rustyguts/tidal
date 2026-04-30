package builder

import (
	"os"
	"strconv"
	"strings"

	"github.com/rustyguts/tidal/internal/domain"
)

func video(ctx Context, spec domain.PresetSpecV2) ([]string, error) {
	if spec.Video.Disabled {
		return []string{"-vn"}, nil
	}
	if spec.Video.Codec == "" {
		return nil, nil
	}
	out := []string{"-c:v", spec.Video.Codec}
	if spec.Video.Codec == "copy" {
		// Stream copy: ignore all encode-time params.
		return out, nil
	}
	if v := strings.TrimSpace(spec.Video.Preset); v != "" {
		out = append(out, "-preset", v)
	}
	if v := strings.TrimSpace(spec.Video.Tune); v != "" {
		out = append(out, "-tune", v)
	}
	if v := strings.TrimSpace(spec.Video.Profile); v != "" {
		out = append(out, "-profile:v", v)
	}
	if v := strings.TrimSpace(spec.Video.Level); v != "" {
		out = append(out, "-level", v)
	}
	if v := strings.TrimSpace(spec.Video.PixelFormat); v != "" {
		out = append(out, "-pix_fmt", v)
	}

	out = append(out, rateFlags(spec.Video.Rate)...)
	out = append(out, gopFlags(spec.Video.GOP)...)
	if v := strings.TrimSpace(spec.Video.FrameRate); v != "" {
		out = append(out, "-r", v)
	}

	if c := spec.Video.Color; c != nil {
		if c.Range != "" {
			out = append(out, "-color_range", c.Range)
		}
		if c.Primaries != "" {
			out = append(out, "-color_primaries", c.Primaries)
		}
		if c.Transfer != "" {
			out = append(out, "-color_trc", c.Transfer)
		}
		if c.Matrix != "" {
			out = append(out, "-colorspace", c.Matrix)
		}
	}

	for _, kv := range spec.Video.CodecExtra {
		out = append(out, kv.Key, kv.Value)
	}

	// Two-pass: stable prefix derived from output path. Pass 1 forces no audio
	// and null muxer; pass 2 produces the actual file.
	if spec.Video.TwoPass && ctx.Pass > 0 {
		prefix := ctx.PassLogPrefix
		if prefix == "" {
			prefix = ctx.OutputPath + ".ffpass"
		}
		out = append(out, "-pass", strconv.Itoa(ctx.Pass), "-passlogfile", prefix)
	}

	return out, nil
}

func rateFlags(r domain.VideoRate) []string {
	switch r.Mode {
	case domain.RateModeCRF:
		out := []string{}
		if r.CRF != nil {
			out = append(out, "-crf", strconv.Itoa(*r.CRF))
		}
		if r.MaxRate != "" {
			out = append(out, "-maxrate", r.MaxRate)
		}
		if r.BufSize != "" {
			out = append(out, "-bufsize", r.BufSize)
		}
		return out
	case domain.RateModeCBR:
		out := []string{}
		if r.Bitrate != "" {
			out = append(out, "-b:v", r.Bitrate, "-minrate", firstNonEmpty(r.MinRate, r.Bitrate), "-maxrate", firstNonEmpty(r.MaxRate, r.Bitrate))
		}
		if r.BufSize != "" {
			out = append(out, "-bufsize", r.BufSize)
		}
		return out
	case domain.RateModeVBR, domain.RateModeABR:
		out := []string{}
		if r.Bitrate != "" {
			out = append(out, "-b:v", r.Bitrate)
		}
		if r.MaxRate != "" {
			out = append(out, "-maxrate", r.MaxRate)
		}
		if r.MinRate != "" {
			out = append(out, "-minrate", r.MinRate)
		}
		if r.BufSize != "" {
			out = append(out, "-bufsize", r.BufSize)
		}
		return out
	case domain.RateModeQP:
		if r.QP != nil {
			return []string{"-qp", strconv.Itoa(*r.QP)}
		}
	}
	return nil
}

func gopFlags(g domain.GopSpec) []string {
	out := []string{}
	if g.KeyintMin > 0 {
		out = append(out, "-keyint_min", strconv.Itoa(g.KeyintMin))
	}
	if g.SceneCut != nil {
		v := "1"
		if !*g.SceneCut {
			v = "0"
		}
		out = append(out, "-sc_threshold", v)
	}
	if g.BFrames != nil {
		out = append(out, "-bf", strconv.Itoa(*g.BFrames))
	}
	if g.RefFrames != nil {
		out = append(out, "-refs", strconv.Itoa(*g.RefFrames))
	}
	// keyintSec is intentionally not emitted as -g here: -g is in frames, and
	// converting requires fps (which can be source-dependent). The codec
	// usually picks a sensible default; users who need exact GOP can set
	// keyintMin or supply a -force_key_frames expression via rawExtras.
	return out
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// PassPrefixForOutput derives a stable ffmpeg2pass log prefix for a given
// output path. Exposed for the runner so it can clean up logs after pass 2.
func PassPrefixForOutput(out string) string {
	return out + ".ffpass"
}

// CleanupPassLogs removes the ffmpeg2pass-NN.log files created by two-pass
// encoding. Called by the runner after a successful pass 2.
func CleanupPassLogs(prefix string) {
	for _, suffix := range []string{"-0.log", "-0.log.mbtree", ".cutree"} {
		_ = os.Remove(prefix + suffix)
	}
}


package builder

import (
	"strings"
	"testing"

	"github.com/rustyguts/tidal/internal/domain"
)

func mustCompose(t *testing.T, ctx Context, spec domain.PresetSpec) []string {
	t.Helper()
	args, err := Compose(ctx, spec)
	if err != nil {
		t.Fatalf("Compose: %v", err)
	}
	return args
}

func TestCompose_h264_basic(t *testing.T) {
	spec := domain.PresetSpec{
		Container: domain.ContainerSpec{Format: "mp4", Faststart: true},
		Video: domain.VideoSpec{
			Codec: "libx264", Preset: "slow",
			Rate: domain.VideoRate{Mode: domain.RateModeCRF, CRF: intPtr(20)},
		},
		Audio: domain.AudioSpec{Codec: "aac", Bitrate: "192k"},
	}
	args := mustCompose(t, Context{InputPath: "/in.mp4", OutputPath: "/out.mp4"}, spec)
	requirePair(t, args, "-c:v", "libx264")
	requirePair(t, args, "-preset", "slow")
	requirePair(t, args, "-crf", "20")
	requirePair(t, args, "-c:a", "aac")
	requirePair(t, args, "-b:a", "192k")
	requirePair(t, args, "-movflags", "+faststart")
	requireToken(t, args, "/out.mp4")
}

func TestCompose_copy_passthrough(t *testing.T) {
	spec := domain.PresetSpec{
		Container: domain.ContainerSpec{Format: "mp4"},
		Video: domain.VideoSpec{
			Codec: "copy",
			Rate:  domain.VideoRate{Mode: domain.RateModeNone},
		},
		Audio: domain.AudioSpec{Codec: "copy", Bitrate: "192k"},
	}
	args := mustCompose(t, Context{InputPath: "/in.mp4", OutputPath: "/out.mp4"}, spec)
	requirePair(t, args, "-c:v", "copy")
	requirePair(t, args, "-c:a", "copy")
	rejectToken(t, args, "-preset")
	rejectToken(t, args, "-crf")
	rejectToken(t, args, "-b:a")
}

func TestCompose_resolution_emitsScale(t *testing.T) {
	spec := domain.PresetSpec{
		Container: domain.ContainerSpec{Format: "mp4"},
		Video: domain.VideoSpec{
			Codec:      "libx264",
			Resolution: &domain.Resolution{Width: 1920, Height: 1080},
			Rate:       domain.VideoRate{Mode: domain.RateModeCRF, CRF: intPtr(23)},
		},
		Audio: domain.AudioSpec{Codec: "aac"},
	}
	args := mustCompose(t, Context{InputPath: "in", OutputPath: "out"}, spec)
	requirePair(t, args, "-vf", "scale=1920:1080")
}

func TestCompose_filters_chain(t *testing.T) {
	spec := domain.PresetSpec{
		Container: domain.ContainerSpec{Format: "mp4"},
		Video: domain.VideoSpec{
			Codec:      "libx264",
			Resolution: &domain.Resolution{Width: 1280, Height: 720},
			Rate:       domain.VideoRate{Mode: domain.RateModeCRF, CRF: intPtr(23)},
		},
		Filters: domain.FilterChain{
			Video: []domain.FilterStep{
				{Name: "yadif", Args: map[string]string{"mode": "1"}},
				{Name: "unsharp"},
			},
		},
		Audio: domain.AudioSpec{Codec: "aac"},
	}
	args := mustCompose(t, Context{InputPath: "in", OutputPath: "out"}, spec)
	vfIdx := indexOf(args, "-vf")
	if vfIdx < 0 {
		t.Fatalf("missing -vf in %v", args)
	}
	chain := args[vfIdx+1]
	if !strings.HasPrefix(chain, "scale=1280:720,") {
		t.Errorf("chain should start with synthesized scale, got %q", chain)
	}
	if !strings.Contains(chain, "yadif=mode=1") {
		t.Errorf("chain missing yadif: %q", chain)
	}
	if !strings.Contains(chain, "unsharp") {
		t.Errorf("chain missing unsharp: %q", chain)
	}
}

func TestCompose_filterEscape_pathColons(t *testing.T) {
	// Subtitle filenames with `:` must be single-quoted.
	step := domain.FilterStep{Name: "subtitles", Args: map[string]string{"filename": "/media/sub:1.srt"}}
	got := renderFilter(step)
	if !strings.Contains(got, `'/media/sub:1.srt'`) {
		t.Errorf("got %q, expected quoted filename", got)
	}
}

func TestCompose_hwaccel_mismatch_errors(t *testing.T) {
	spec := domain.PresetSpec{
		Container: domain.ContainerSpec{Format: "mp4"},
		Hwaccel:   &domain.HwaccelSpec{Type: "nvdec"},
		Video: domain.VideoSpec{
			Codec: "libx264", // CPU codec, not NVENC
			Rate:  domain.VideoRate{Mode: domain.RateModeCRF, CRF: intPtr(23)},
		},
		Audio: domain.AudioSpec{Codec: "aac"},
	}
	if _, err := Compose(Context{InputPath: "in", OutputPath: "out"}, spec); err == nil {
		t.Fatal("expected error for nvdec + libx264 mismatch")
	}
}

func TestCompose_hwaccel_nvdec_pairing(t *testing.T) {
	spec := domain.PresetSpec{
		Container: domain.ContainerSpec{Format: "mp4"},
		Hwaccel:   &domain.HwaccelSpec{Type: "nvdec"},
		Video: domain.VideoSpec{
			Codec: "h264_nvenc",
			Rate:  domain.VideoRate{Mode: domain.RateModeCBR, Bitrate: "5M"},
		},
		Audio: domain.AudioSpec{Codec: "aac"},
	}
	args := mustCompose(t, Context{InputPath: "in", OutputPath: "out"}, spec)
	requirePair(t, args, "-hwaccel", "cuda")
	requirePair(t, args, "-c:v", "h264_nvenc")
	requirePair(t, args, "-b:v", "5M")
}

func TestCompose_twoPass_pass1_emitsNullMuxer(t *testing.T) {
	spec := domain.PresetSpec{
		Container: domain.ContainerSpec{Format: "mp4"},
		Video: domain.VideoSpec{
			Codec:   "libx264",
			TwoPass: true,
			Rate:    domain.VideoRate{Mode: domain.RateModeCBR, Bitrate: "5M"},
		},
		Audio: domain.AudioSpec{Codec: "aac"},
	}
	args1 := mustCompose(t, Context{InputPath: "in", OutputPath: "out", Pass: 1, PassLogPrefix: "/tmp/p"}, spec)
	requirePair(t, args1, "-pass", "1")
	requirePair(t, args1, "-passlogfile", "/tmp/p")
	requirePair(t, args1, "-f", "null")
	requireToken(t, args1, "-an")

	args2 := mustCompose(t, Context{InputPath: "in", OutputPath: "out", Pass: 2, PassLogPrefix: "/tmp/p"}, spec)
	requirePair(t, args2, "-pass", "2")
	requireToken(t, args2, "out")
	rejectToken(t, args2, "-f")
}

func TestCompose_disabled_video(t *testing.T) {
	spec := domain.PresetSpec{
		Container: domain.ContainerSpec{Format: "mp4"},
		Video:     domain.VideoSpec{Disabled: true},
		Audio:     domain.AudioSpec{Codec: "aac", Bitrate: "192k"},
	}
	args := mustCompose(t, Context{InputPath: "in", OutputPath: "out"}, spec)
	requireToken(t, args, "-vn")
	rejectToken(t, args, "-c:v")
}

func TestCompose_audio_disabled(t *testing.T) {
	spec := domain.PresetSpec{
		Container: domain.ContainerSpec{Format: "mp4"},
		Video: domain.VideoSpec{
			Codec: "libx264",
			Rate:  domain.VideoRate{Mode: domain.RateModeCRF, CRF: intPtr(23)},
		},
		Audio: domain.AudioSpec{Disabled: true},
	}
	args := mustCompose(t, Context{InputPath: "in", OutputPath: "out"}, spec)
	requireToken(t, args, "-an")
	rejectToken(t, args, "-c:a")
}

func TestCompose_audio_extras(t *testing.T) {
	spec := domain.PresetSpec{
		Container: domain.ContainerSpec{Format: "mp4"},
		Video: domain.VideoSpec{
			Codec: "libx264",
			Rate:  domain.VideoRate{Mode: domain.RateModeCRF, CRF: intPtr(23)},
		},
		Audio: domain.AudioSpec{Codec: "aac", Bitrate: "192k", SampleRate: 48000, Channels: 2},
	}
	args := mustCompose(t, Context{InputPath: "in", OutputPath: "out"}, spec)
	requirePair(t, args, "-ar", "48000")
	requirePair(t, args, "-ac", "2")
}

func TestCompose_input_seek_duration(t *testing.T) {
	spec := domain.PresetSpec{
		Container: domain.ContainerSpec{Format: "mp4"},
		Input:     domain.InputSpec{SeekStart: "00:01:30", Duration: "30"},
		Video: domain.VideoSpec{
			Codec: "libx264",
			Rate:  domain.VideoRate{Mode: domain.RateModeCRF, CRF: intPtr(23)},
		},
		Audio: domain.AudioSpec{Codec: "aac"},
	}
	args := mustCompose(t, Context{InputPath: "in", OutputPath: "out"}, spec)
	requirePair(t, args, "-ss", "00:01:30")
	requirePair(t, args, "-t", "30")
	// -ss must come before -i.
	if idxOf(args, "-ss") > idxOf(args, "-i") {
		t.Errorf("-ss must precede -i; args = %v", args)
	}
}

func intPtr(v int) *int { return &v }

func indexOf(args []string, want string) int {
	for i, a := range args {
		if a == want {
			return i
		}
	}
	return -1
}

func idxOf(args []string, want string) int { return indexOf(args, want) }

func requirePair(t *testing.T, args []string, flag, val string) {
	t.Helper()
	for i := 0; i < len(args)-1; i++ {
		if args[i] == flag && args[i+1] == val {
			return
		}
	}
	t.Errorf("missing %s %s in %v", flag, val, args)
}

func requireToken(t *testing.T, args []string, want string) {
	t.Helper()
	if indexOf(args, want) < 0 {
		t.Errorf("missing token %q in %v", want, args)
	}
}

func rejectToken(t *testing.T, args []string, unwanted string) {
	t.Helper()
	if indexOf(args, unwanted) >= 0 {
		t.Errorf("unexpected token %q in %v", unwanted, args)
	}
}

package builder

import (
	"os"
	"strings"
	"testing"

	"github.com/rustyguts/tidal/internal/domain"
)

// Cover gaps in mapping.go.
func TestMapping_explicitStreamsAndSelectors(t *testing.T) {
	spec := domain.PresetSpecV2{
		Container: domain.ContainerSpec{Format: "mp4"},
		Video:     domain.VideoSpec{Codec: "libx264", Rate: domain.VideoRate{Mode: domain.RateModeCRF, CRF: intPtr(23)}},
		Audio:     domain.AudioSpec{Codec: "aac"},
		Mapping: domain.MappingSpec{
			Streams:  []string{"0:v:0", "0:a:0"},
			Video:    "0:v:1",
			Audio:    "none",
			Subtitle: "0:s:0",
		},
	}
	args := mustCompose(t, Context{InputPath: "in", OutputPath: "out"}, spec)
	requirePair(t, args, "-map", "0:v:0")
	requirePair(t, args, "-map", "0:a:0")
	requirePair(t, args, "-map", "0:v:1")
	requirePair(t, args, "-map", "-0:a")
	requirePair(t, args, "-map", "0:s:0")
}

func TestMapping_noneVideoNoneSubtitle(t *testing.T) {
	spec := domain.PresetSpecV2{
		Container: domain.ContainerSpec{Format: "mp4"},
		Video:     domain.VideoSpec{Codec: "libx264", Rate: domain.VideoRate{Mode: domain.RateModeCRF, CRF: intPtr(23)}},
		Audio:     domain.AudioSpec{Codec: "aac"},
		Mapping: domain.MappingSpec{
			Video:    "none",
			Subtitle: "none",
		},
	}
	args := mustCompose(t, Context{InputPath: "in", OutputPath: "out"}, spec)
	requirePair(t, args, "-map", "-0:v")
	requirePair(t, args, "-map", "-0:s")
}

// Hwaccel branches: vaapi, qsv, videotoolbox, unknown.
func TestHwaccel_vaapi(t *testing.T) {
	spec := domain.PresetSpecV2{
		Container: domain.ContainerSpec{Format: "mp4"},
		Hwaccel:   &domain.HwaccelSpec{Type: "vaapi", Device: "/dev/dri/renderD129", OutputFormat: "vaapi"},
		Video:     domain.VideoSpec{Codec: "h264_vaapi", Rate: domain.VideoRate{Mode: domain.RateModeCBR, Bitrate: "5M"}},
		Audio:     domain.AudioSpec{Codec: "aac"},
	}
	args := mustCompose(t, Context{InputPath: "in", OutputPath: "out"}, spec)
	requirePair(t, args, "-vaapi_device", "/dev/dri/renderD129")
	requirePair(t, args, "-hwaccel", "vaapi")
	requirePair(t, args, "-hwaccel_output_format", "vaapi")
}

func TestHwaccel_vaapi_defaultDevice(t *testing.T) {
	spec := domain.PresetSpecV2{
		Container: domain.ContainerSpec{Format: "mp4"},
		Hwaccel:   &domain.HwaccelSpec{Type: "vaapi"},
		Video:     domain.VideoSpec{Codec: "h264_vaapi", Rate: domain.VideoRate{Mode: domain.RateModeCBR, Bitrate: "5M"}},
		Audio:     domain.AudioSpec{Codec: "aac"},
	}
	args := mustCompose(t, Context{InputPath: "in", OutputPath: "out"}, spec)
	requirePair(t, args, "-vaapi_device", "/dev/dri/renderD128")
	requirePair(t, args, "-hwaccel_output_format", "vaapi")
}

func TestHwaccel_qsv(t *testing.T) {
	spec := domain.PresetSpecV2{
		Container: domain.ContainerSpec{Format: "mp4"},
		Hwaccel:   &domain.HwaccelSpec{Type: "qsv"},
		Video:     domain.VideoSpec{Codec: "h264_qsv", Rate: domain.VideoRate{Mode: domain.RateModeCBR, Bitrate: "5M"}},
		Audio:     domain.AudioSpec{Codec: "aac"},
	}
	args := mustCompose(t, Context{InputPath: "in", OutputPath: "out"}, spec)
	requirePair(t, args, "-hwaccel", "qsv")
	requirePair(t, args, "-qsv_device", "/dev/dri/renderD128")
}

func TestHwaccel_qsv_explicitOutputFormat(t *testing.T) {
	spec := domain.PresetSpecV2{
		Container: domain.ContainerSpec{Format: "mp4"},
		Hwaccel:   &domain.HwaccelSpec{Type: "qsv", OutputFormat: "nv12"},
		Video:     domain.VideoSpec{Codec: "h264_qsv", Rate: domain.VideoRate{Mode: domain.RateModeCBR, Bitrate: "5M"}},
		Audio:     domain.AudioSpec{Codec: "aac"},
	}
	args := mustCompose(t, Context{InputPath: "in", OutputPath: "out"}, spec)
	requirePair(t, args, "-hwaccel_output_format", "nv12")
}

func TestHwaccel_videotoolbox(t *testing.T) {
	spec := domain.PresetSpecV2{
		Container: domain.ContainerSpec{Format: "mp4"},
		Hwaccel:   &domain.HwaccelSpec{Type: "videotoolbox", OutputFormat: "videotoolbox_vld"},
		Video:     domain.VideoSpec{Codec: "h264_videotoolbox", Rate: domain.VideoRate{Mode: domain.RateModeCBR, Bitrate: "5M"}},
		Audio:     domain.AudioSpec{Codec: "aac"},
	}
	args, err := Compose(Context{InputPath: "in", OutputPath: "out"}, spec)
	if err != nil {
		t.Fatalf("Compose: %v", err)
	}
	requirePair(t, args, "-hwaccel", "videotoolbox")
	requirePair(t, args, "-hwaccel_output_format", "videotoolbox_vld")
}

func TestHwaccel_unknownType_errors(t *testing.T) {
	spec := domain.PresetSpecV2{
		Container: domain.ContainerSpec{Format: "mp4"},
		Hwaccel:   &domain.HwaccelSpec{Type: "magic"},
		Video:     domain.VideoSpec{Codec: "libx264", Rate: domain.VideoRate{Mode: domain.RateModeCRF, CRF: intPtr(23)}},
		Audio:     domain.AudioSpec{Codec: "aac"},
	}
	if _, err := Compose(Context{InputPath: "in", OutputPath: "out"}, spec); err == nil {
		t.Fatal("expected unknown hwaccel error")
	}
}

func TestHwaccel_vaapi_codecMismatch(t *testing.T) {
	spec := domain.PresetSpecV2{
		Container: domain.ContainerSpec{Format: "mp4"},
		Hwaccel:   &domain.HwaccelSpec{Type: "vaapi"},
		Video:     domain.VideoSpec{Codec: "libx264", Rate: domain.VideoRate{Mode: domain.RateModeCRF, CRF: intPtr(23)}},
		Audio:     domain.AudioSpec{Codec: "aac"},
	}
	if _, err := Compose(Context{InputPath: "in", OutputPath: "out"}, spec); err == nil {
		t.Fatal("expected vaapi/codec mismatch")
	}
}

func TestHwaccel_qsv_codecMismatch(t *testing.T) {
	spec := domain.PresetSpecV2{
		Container: domain.ContainerSpec{Format: "mp4"},
		Hwaccel:   &domain.HwaccelSpec{Type: "qsv"},
		Video:     domain.VideoSpec{Codec: "libx264", Rate: domain.VideoRate{Mode: domain.RateModeCRF, CRF: intPtr(23)}},
		Audio:     domain.AudioSpec{Codec: "aac"},
	}
	if _, err := Compose(Context{InputPath: "in", OutputPath: "out"}, spec); err == nil {
		t.Fatal("expected qsv/codec mismatch")
	}
}

// Subtitle modes.
func TestSubtitles_copy(t *testing.T) {
	spec := domain.PresetSpecV2{
		Container: domain.ContainerSpec{Format: "mkv"},
		Video:     domain.VideoSpec{Codec: "libx264", Rate: domain.VideoRate{Mode: domain.RateModeCRF, CRF: intPtr(23)}},
		Audio:     domain.AudioSpec{Codec: "aac"},
		Subtitles: domain.SubtitleSpec{Mode: domain.SubtitleCopy},
	}
	args := mustCompose(t, Context{InputPath: "in", OutputPath: "out"}, spec)
	requirePair(t, args, "-c:s", "copy")
}

func TestSubtitles_strip(t *testing.T) {
	spec := domain.PresetSpecV2{
		Container: domain.ContainerSpec{Format: "mp4"},
		Video:     domain.VideoSpec{Codec: "libx264", Rate: domain.VideoRate{Mode: domain.RateModeCRF, CRF: intPtr(23)}},
		Audio:     domain.AudioSpec{Codec: "aac"},
		Subtitles: domain.SubtitleSpec{Mode: domain.SubtitleStrip},
	}
	args := mustCompose(t, Context{InputPath: "in", OutputPath: "out"}, spec)
	requireToken(t, args, "-sn")
}

func TestSubtitles_burn_emitsNoFlag(t *testing.T) {
	spec := domain.PresetSpecV2{
		Container: domain.ContainerSpec{Format: "mp4"},
		Video:     domain.VideoSpec{Codec: "libx264", Rate: domain.VideoRate{Mode: domain.RateModeCRF, CRF: intPtr(23)}},
		Audio:     domain.AudioSpec{Codec: "aac"},
		Subtitles: domain.SubtitleSpec{Mode: domain.SubtitleBurn},
	}
	args := mustCompose(t, Context{InputPath: "in", OutputPath: "out"}, spec)
	rejectToken(t, args, "-sn")
	rejectToken(t, args, "-c:s")
}

// Rate flag branches.
func TestRateFlags_CBR_minMax(t *testing.T) {
	spec := domain.PresetSpecV2{
		Container: domain.ContainerSpec{Format: "mp4"},
		Video: domain.VideoSpec{
			Codec: "libx264",
			Rate:  domain.VideoRate{Mode: domain.RateModeCBR, Bitrate: "4M", BufSize: "8M", MinRate: "3M", MaxRate: "5M"},
		},
		Audio: domain.AudioSpec{Codec: "aac"},
	}
	args := mustCompose(t, Context{InputPath: "in", OutputPath: "out"}, spec)
	requirePair(t, args, "-b:v", "4M")
	requirePair(t, args, "-bufsize", "8M")
	// CBR forces minrate==maxrate to bitrate when not provided; here MinRate is set so it's used as-is.
	requirePair(t, args, "-minrate", "3M")
	requirePair(t, args, "-maxrate", "5M")
}

func TestRateFlags_CBR_fallbackToBitrate(t *testing.T) {
	spec := domain.PresetSpecV2{
		Container: domain.ContainerSpec{Format: "mp4"},
		Video: domain.VideoSpec{
			Codec: "libx264",
			Rate:  domain.VideoRate{Mode: domain.RateModeCBR, Bitrate: "4M"},
		},
		Audio: domain.AudioSpec{Codec: "aac"},
	}
	args := mustCompose(t, Context{InputPath: "in", OutputPath: "out"}, spec)
	requirePair(t, args, "-minrate", "4M")
	requirePair(t, args, "-maxrate", "4M")
}

func TestRateFlags_VBR_full(t *testing.T) {
	spec := domain.PresetSpecV2{
		Container: domain.ContainerSpec{Format: "mp4"},
		Video: domain.VideoSpec{
			Codec: "libx264",
			Rate:  domain.VideoRate{Mode: domain.RateModeVBR, Bitrate: "3M", MaxRate: "5M", MinRate: "1M", BufSize: "6M"},
		},
		Audio: domain.AudioSpec{Codec: "aac"},
	}
	args := mustCompose(t, Context{InputPath: "in", OutputPath: "out"}, spec)
	requirePair(t, args, "-b:v", "3M")
	requirePair(t, args, "-maxrate", "5M")
	requirePair(t, args, "-minrate", "1M")
	requirePair(t, args, "-bufsize", "6M")
}

func TestRateFlags_QP(t *testing.T) {
	qp := 22
	spec := domain.PresetSpecV2{
		Container: domain.ContainerSpec{Format: "mp4"},
		Video: domain.VideoSpec{
			Codec: "h264_nvenc",
			Rate:  domain.VideoRate{Mode: domain.RateModeQP, QP: &qp},
		},
		Audio:   domain.AudioSpec{Codec: "aac"},
		Hwaccel: nil,
	}
	args := mustCompose(t, Context{InputPath: "in", OutputPath: "out"}, spec)
	requirePair(t, args, "-qp", "22")
}

func TestRateFlags_CRF_withMaxAndBuf(t *testing.T) {
	spec := domain.PresetSpecV2{
		Container: domain.ContainerSpec{Format: "mp4"},
		Video: domain.VideoSpec{
			Codec: "libx264",
			Rate:  domain.VideoRate{Mode: domain.RateModeCRF, CRF: intPtr(20), MaxRate: "5M", BufSize: "10M"},
		},
		Audio: domain.AudioSpec{Codec: "aac"},
	}
	args := mustCompose(t, Context{InputPath: "in", OutputPath: "out"}, spec)
	requirePair(t, args, "-crf", "20")
	requirePair(t, args, "-maxrate", "5M")
	requirePair(t, args, "-bufsize", "10M")
}

// GOP flags.
func TestGopFlags_full(t *testing.T) {
	scene := false
	bf := 3
	refs := 4
	spec := domain.PresetSpecV2{
		Container: domain.ContainerSpec{Format: "mp4"},
		Video: domain.VideoSpec{
			Codec: "libx264",
			Rate:  domain.VideoRate{Mode: domain.RateModeCRF, CRF: intPtr(20)},
			GOP:   domain.GopSpec{KeyintMin: 24, SceneCut: &scene, BFrames: &bf, RefFrames: &refs},
		},
		Audio: domain.AudioSpec{Codec: "aac"},
	}
	args := mustCompose(t, Context{InputPath: "in", OutputPath: "out"}, spec)
	requirePair(t, args, "-keyint_min", "24")
	requirePair(t, args, "-sc_threshold", "0")
	requirePair(t, args, "-bf", "3")
	requirePair(t, args, "-refs", "4")
}

func TestGopFlags_sceneCutTrue(t *testing.T) {
	scene := true
	spec := domain.PresetSpecV2{
		Container: domain.ContainerSpec{Format: "mp4"},
		Video: domain.VideoSpec{
			Codec: "libx264",
			Rate:  domain.VideoRate{Mode: domain.RateModeCRF, CRF: intPtr(20)},
			GOP:   domain.GopSpec{SceneCut: &scene},
		},
		Audio: domain.AudioSpec{Codec: "aac"},
	}
	args := mustCompose(t, Context{InputPath: "in", OutputPath: "out"}, spec)
	requirePair(t, args, "-sc_threshold", "1")
}

// Color and codec extras.
func TestVideo_color_and_codecExtra(t *testing.T) {
	spec := domain.PresetSpecV2{
		Container: domain.ContainerSpec{Format: "mp4"},
		Video: domain.VideoSpec{
			Codec: "libx264",
			Tune:  "film",
			Profile: "high",
			Level: "4.1",
			Rate:  domain.VideoRate{Mode: domain.RateModeCRF, CRF: intPtr(20)},
			Color: &domain.ColorSpec{Range: "tv", Primaries: "bt709", Transfer: "bt709", Matrix: "bt709"},
			CodecExtra: []domain.KV{{Key: "-x264-params", Value: "keyint=240"}},
		},
		Audio: domain.AudioSpec{Codec: "aac"},
	}
	args := mustCompose(t, Context{InputPath: "in", OutputPath: "out"}, spec)
	requirePair(t, args, "-tune", "film")
	requirePair(t, args, "-profile:v", "high")
	requirePair(t, args, "-level", "4.1")
	requirePair(t, args, "-color_range", "tv")
	requirePair(t, args, "-color_primaries", "bt709")
	requirePair(t, args, "-color_trc", "bt709")
	requirePair(t, args, "-colorspace", "bt709")
	requirePair(t, args, "-x264-params", "keyint=240")
}

// Global flags: threading + progressURL.
func TestGlobal_threading_progress(t *testing.T) {
	spec := domain.PresetSpecV2{
		Container: domain.ContainerSpec{Format: "mp4"},
		Video:     domain.VideoSpec{Codec: "libx264", Rate: domain.VideoRate{Mode: domain.RateModeCRF, CRF: intPtr(20)}},
		Audio:     domain.AudioSpec{Codec: "aac"},
		Threading: domain.ThreadingSpec{Threads: 4, FilterThreads: 2},
	}
	args := mustCompose(t, Context{InputPath: "in", OutputPath: "out", ProgressURL: "pipe:1"}, spec)
	requirePair(t, args, "-threads", "4")
	requirePair(t, args, "-filter_threads", "2")
	requirePair(t, args, "-progress", "pipe:1")
}

// Audio extras: profile, vbrQuality, sample rate, channels.
func TestAudio_profile_vbr(t *testing.T) {
	q := 5
	spec := domain.PresetSpecV2{
		Container: domain.ContainerSpec{Format: "mp4"},
		Video:     domain.VideoSpec{Codec: "libx264", Rate: domain.VideoRate{Mode: domain.RateModeCRF, CRF: intPtr(20)}},
		Audio:     domain.AudioSpec{Codec: "libopus", SampleRate: 48000, Channels: 2, Profile: "voip", VBRQuality: &q},
	}
	args := mustCompose(t, Context{InputPath: "in", OutputPath: "out"}, spec)
	requirePair(t, args, "-profile:a", "voip")
	requirePair(t, args, "-q:a", "5")
}

// Container: explicit movflags + fragment duration.
func TestContainer_explicitMovflags(t *testing.T) {
	spec := domain.PresetSpecV2{
		Container: domain.ContainerSpec{
			Format: "mp4", Faststart: true,
			MovFlags: []string{"+faststart", "+frag_keyframe"},
			FragmentDurMs: 4000,
		},
		Video: domain.VideoSpec{Codec: "libx264", Rate: domain.VideoRate{Mode: domain.RateModeCRF, CRF: intPtr(20)}},
		Audio: domain.AudioSpec{Codec: "aac"},
	}
	args := mustCompose(t, Context{InputPath: "in", OutputPath: "out"}, spec)
	movIdx := indexOf(args, "-movflags")
	if movIdx < 0 {
		t.Fatalf("missing -movflags in %v", args)
	}
	value := args[movIdx+1]
	if !strings.Contains(value, "+faststart") || !strings.Contains(value, "+frag_keyframe") {
		t.Errorf("movflags = %q", value)
	}
	requirePair(t, args, "-frag_duration", "4000000")
}

func TestContainer_mkv_skipsMovflags(t *testing.T) {
	spec := domain.PresetSpecV2{
		Container: domain.ContainerSpec{Format: "mkv", Faststart: true, MovFlags: []string{"+frag_keyframe"}},
		Video:     domain.VideoSpec{Codec: "libx264", Rate: domain.VideoRate{Mode: domain.RateModeCRF, CRF: intPtr(20)}},
		Audio:     domain.AudioSpec{Codec: "aac"},
	}
	args := mustCompose(t, Context{InputPath: "in", OutputPath: "out"}, spec)
	rejectToken(t, args, "-movflags")
}

// Filter chain: audio chain + escaping coverage.
func TestAudioFilterChain_renders(t *testing.T) {
	spec := domain.PresetSpecV2{
		Container: domain.ContainerSpec{Format: "mp4"},
		Video:     domain.VideoSpec{Codec: "libx264", Rate: domain.VideoRate{Mode: domain.RateModeCRF, CRF: intPtr(20)}},
		Audio:     domain.AudioSpec{Codec: "aac"},
		Filters: domain.FilterChain{
			Audio: []domain.FilterStep{
				{Name: "loudnorm", Args: map[string]string{"I": "-16"}},
				{Name: "volume", Args: map[string]string{"volume": "2.0"}},
			},
		},
	}
	args := mustCompose(t, Context{InputPath: "in", OutputPath: "out"}, spec)
	afIdx := indexOf(args, "-af")
	if afIdx < 0 {
		t.Fatalf("missing -af in %v", args)
	}
	if !strings.Contains(args[afIdx+1], "loudnorm=") || !strings.Contains(args[afIdx+1], "volume=") {
		t.Errorf("audio chain = %q", args[afIdx+1])
	}
}

func TestEscapeFilterArg_nonSpecialPassthrough(t *testing.T) {
	if got := escapeFilterArg("plain"); got != "plain" {
		t.Errorf("escapeFilterArg(plain) = %q", got)
	}
}

func TestEscapeFilterArg_backslashAndQuote(t *testing.T) {
	got := escapeFilterArg(`a\b'c:d`)
	if got != `'a\\b\'c:d'` {
		t.Errorf("escapeFilterArg = %q", got)
	}
}

// Disabled filter step is skipped.
func TestFilterStep_disabled_skipped(t *testing.T) {
	disabled := false
	spec := domain.PresetSpecV2{
		Container: domain.ContainerSpec{Format: "mp4"},
		Video: domain.VideoSpec{
			Codec: "libx264",
			Rate:  domain.VideoRate{Mode: domain.RateModeCRF, CRF: intPtr(20)},
		},
		Audio: domain.AudioSpec{Codec: "aac"},
		Filters: domain.FilterChain{
			Video: []domain.FilterStep{{Name: "yadif", Enabled: &disabled}},
		},
	}
	args := mustCompose(t, Context{InputPath: "in", OutputPath: "out"}, spec)
	for _, a := range args {
		if strings.Contains(a, "yadif") {
			t.Errorf("disabled yadif should not appear in args: %v", args)
		}
	}
}

// Pass log helpers.
func TestPassPrefixForOutput(t *testing.T) {
	got := PassPrefixForOutput("/x/out.mp4")
	if got != "/x/out.mp4.ffpass" {
		t.Errorf("got %q", got)
	}
}

func TestCleanupPassLogs_doesNotPanicOnMissing(t *testing.T) {
	prefix := "/tmp/_tidal_pass_test_unlikely_path"
	CleanupPassLogs(prefix) // should be safe even when files don't exist
}

func TestCleanupPassLogs_removesFiles(t *testing.T) {
	dir := t.TempDir()
	prefix := dir + "/p"
	for _, suffix := range []string{"-0.log", "-0.log.mbtree", ".cutree"} {
		if err := os.WriteFile(prefix+suffix, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	CleanupPassLogs(prefix)
	for _, suffix := range []string{"-0.log", "-0.log.mbtree", ".cutree"} {
		if _, err := os.Stat(prefix + suffix); err == nil {
			t.Errorf("%s still exists", prefix+suffix)
		}
	}
}

// uniqueMovFlags dedup.
func TestUniqueMovFlags_dedupes(t *testing.T) {
	got := uniqueMovFlags(true, []string{"+faststart", "+frag_keyframe"})
	if len(got) != 2 {
		t.Errorf("expected dedup, got %v", got)
	}
}

// firstNonEmpty.
func TestFirstNonEmpty(t *testing.T) {
	if got := firstNonEmpty("", "", "x", "y"); got != "x" {
		t.Errorf("got %q", got)
	}
	if got := firstNonEmpty("", ""); got != "" {
		t.Errorf("got %q", got)
	}
}

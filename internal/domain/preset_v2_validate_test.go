package domain

import (
	"strings"
	"testing"

	"github.com/rustyguts/tidal/internal/ffmpeg/catalog"
)

func validV2() PresetSpecV2 {
	crf := 23
	return PresetSpecV2{
		SchemaVersion: SchemaVersionV2,
		Container:     ContainerSpec{Format: "mp4", Faststart: true},
		Video: VideoSpec{
			Codec: "libx264", Preset: "slow",
			Rate:        VideoRate{Mode: RateModeCRF, CRF: &crf},
			PixelFormat: "yuv420p",
		},
		Audio: AudioSpec{Codec: "aac", Bitrate: "192k", SampleRate: 48000, Channels: 2},
	}
}

func TestValidateV2_happyPath(t *testing.T) {
	cat := catalog.Default()
	if err := ValidateV2(validV2(), cat, ValidateOpts{}); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
}

func TestValidateV2_unknownContainer(t *testing.T) {
	s := validV2()
	s.Container.Format = "wmv"
	err := ValidateV2(s, catalog.Default(), ValidateOpts{})
	mustErrContain(t, err, "container.format")
}

func TestValidateV2_unknownCodec(t *testing.T) {
	s := validV2()
	s.Video.Codec = "vc1"
	err := ValidateV2(s, catalog.Default(), ValidateOpts{})
	mustErrContain(t, err, "video.codec")
}

func TestValidateV2_invalidPresetForCodec(t *testing.T) {
	s := validV2()
	s.Video.Preset = "p7" // NVENC preset, not valid for libx264
	err := ValidateV2(s, catalog.Default(), ValidateOpts{})
	mustErrContain(t, err, "video.preset")
}

func TestValidateV2_crfOutOfRange(t *testing.T) {
	s := validV2()
	huge := 99
	s.Video.Rate.CRF = &huge
	err := ValidateV2(s, catalog.Default(), ValidateOpts{})
	mustErrContain(t, err, "video.rate.crf")
}

func TestValidateV2_twoPassUnsupported(t *testing.T) {
	s := validV2()
	s.Video.Codec = "h264_nvenc"
	s.Video.Preset = "p4"
	s.Video.Rate = VideoRate{Mode: RateModeCBR, Bitrate: "5M"}
	s.Video.TwoPass = true
	err := ValidateV2(s, catalog.Default(), ValidateOpts{})
	mustErrContain(t, err, "video.twoPass")
}

func TestValidateV2_hwaccelMismatch(t *testing.T) {
	s := validV2()
	s.Hwaccel = &HwaccelSpec{Type: "nvdec"}
	err := ValidateV2(s, catalog.Default(), ValidateOpts{})
	mustErrContain(t, err, "hwaccel")
}

func TestValidateV2_resolutionOdd(t *testing.T) {
	s := validV2()
	s.Video.Resolution = &Resolution{Width: 1921, Height: 1080}
	err := ValidateV2(s, catalog.Default(), ValidateOpts{})
	mustErrContain(t, err, "resolution.width")
}

func TestValidateV2_filterUnknown(t *testing.T) {
	s := validV2()
	s.Filters.Video = []FilterStep{{Name: "frobnicate"}}
	err := ValidateV2(s, catalog.Default(), ValidateOpts{})
	mustErrContain(t, err, "frobnicate")
}

func TestValidateV2_filterScopeMismatch(t *testing.T) {
	s := validV2()
	// loudnorm is audio-scope; using it in video chain must fail.
	s.Filters.Video = []FilterStep{{Name: "loudnorm"}}
	err := ValidateV2(s, catalog.Default(), ValidateOpts{})
	mustErrContain(t, err, "scope")
}

func TestValidateV2_filterMissingRequired(t *testing.T) {
	s := validV2()
	s.Filters.Video = []FilterStep{{Name: "scale"}} // w, h required
	err := ValidateV2(s, catalog.Default(), ValidateOpts{})
	mustErrContain(t, err, "scale")
	mustErrContain(t, err, "required")
}

func TestValidateV2_rawExtras_allowed(t *testing.T) {
	s := validV2()
	s.RawExtras = []string{"-map", "0:v:0", "-map", "0:a:0"}
	if err := ValidateV2(s, catalog.Default(), ValidateOpts{}); err != nil {
		t.Errorf("expected valid map allowlist, got %v", err)
	}
}

func TestValidateV2_rawExtras_deniedFlag(t *testing.T) {
	s := validV2()
	s.RawExtras = []string{"-i", "/etc/passwd"}
	err := ValidateV2(s, catalog.Default(), ValidateOpts{})
	mustErrContain(t, err, "denylisted")
}

func TestValidateV2_rawExtras_shellInjection(t *testing.T) {
	s := validV2()
	s.RawExtras = []string{"-map", "0:v:0; rm -rf /"}
	err := ValidateV2(s, catalog.Default(), ValidateOpts{})
	mustErrContain(t, err, "forbidden character")
}

func TestValidateV2_rawExtras_outsideAllowlist(t *testing.T) {
	s := validV2()
	s.RawExtras = []string{"-some-random-flag", "value"}
	err := ValidateV2(s, catalog.Default(), ValidateOpts{})
	mustErrContain(t, err, "allowlist")
}

func TestValidateV2_rawExtras_permissiveSkipsAllowlist(t *testing.T) {
	s := validV2()
	s.RawExtras = []string{"-some-random-flag", "value"}
	if err := ValidateV2(s, catalog.Default(), ValidateOpts{PermissiveRawExtras: true}); err != nil {
		t.Errorf("permissive mode should skip allowlist, got %v", err)
	}
}

func TestValidateV2_rawExtras_permissiveStillEnforcesDeny(t *testing.T) {
	s := validV2()
	s.RawExtras = []string{"-i", "/etc/passwd"}
	err := ValidateV2(s, catalog.Default(), ValidateOpts{PermissiveRawExtras: true})
	mustErrContain(t, err, "denylisted")
}

func TestValidateV2_v1upgrade_allBuiltinsValid(t *testing.T) {
	cat := catalog.Default()
	builtins := []PresetSpec{
		{Container: "mp4", VideoCodec: "libx264", VideoPreset: "slow", CRF: 20, AudioCodec: "aac", AudioBitrate: "192k", Resolution: &Resolution{Width: 1920, Height: 1080}},
		{Container: "mp4", VideoCodec: "libx264", VideoPreset: "slow", CRF: 20, AudioCodec: "aac", AudioBitrate: "192k", Resolution: &Resolution{Width: 3840, Height: 2160}},
		{Container: "mkv", VideoCodec: "libx265", VideoPreset: "medium", CRF: 22, AudioCodec: "aac", AudioBitrate: "160k", Resolution: &Resolution{Width: 1920, Height: 1080}},
		{Container: "mp4", VideoCodec: "libsvtav1", VideoPreset: "4", CRF: 30, AudioCodec: "aac", AudioBitrate: "160k", Resolution: &Resolution{Width: 1920, Height: 1080}},
		{Container: "mp4", VideoCodec: "copy", AudioCodec: "aac", AudioBitrate: "192k", ExtraArgs: []string{"-vn"}},
	}
	for i, v1 := range builtins {
		v2 := UpgradeFromV1(v1)
		if err := ValidateV2(v2, cat, ValidateOpts{}); err != nil {
			t.Errorf("builtin[%d] %q invalid after upgrade: %v", i, v1.VideoCodec, err)
		}
	}
}

func mustErrContain(t *testing.T, err error, sub string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error containing %q, got nil", sub)
	}
	if !strings.Contains(err.Error(), sub) {
		t.Errorf("error %q does not contain %q", err.Error(), sub)
	}
}

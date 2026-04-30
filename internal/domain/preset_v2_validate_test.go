package domain

import (
	"strings"
	"testing"

	"github.com/rustyguts/tidal/internal/ffmpeg/catalog"
)

func validV2() PresetSpec {
	crf := 23
	return PresetSpec{
		Container: ContainerSpec{Format: "mp4", Faststart: true},
		Video: VideoSpec{
			Codec: "libx264", Preset: "slow",
			Rate:        VideoRate{Mode: RateModeCRF, CRF: &crf},
			PixelFormat: "yuv420p",
		},
		Audio: AudioSpec{Codec: "aac", Bitrate: "192k", SampleRate: 48000, Channels: 2},
	}
}

func TestValidate_happyPath(t *testing.T) {
	cat := catalog.Default()
	if err := Validate(validV2(), cat, ValidateOpts{}); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
}

func TestValidate_unknownContainer(t *testing.T) {
	s := validV2()
	s.Container.Format = "wmv"
	err := Validate(s, catalog.Default(), ValidateOpts{})
	mustErrContain(t, err, "container.format")
}

func TestValidate_unknownCodec(t *testing.T) {
	s := validV2()
	s.Video.Codec = "vc1"
	err := Validate(s, catalog.Default(), ValidateOpts{})
	mustErrContain(t, err, "video.codec")
}

func TestValidate_invalidPresetForCodec(t *testing.T) {
	s := validV2()
	s.Video.Preset = "p7" // NVENC preset, not valid for libx264
	err := Validate(s, catalog.Default(), ValidateOpts{})
	mustErrContain(t, err, "video.preset")
}

func TestValidate_crfOutOfRange(t *testing.T) {
	s := validV2()
	huge := 99
	s.Video.Rate.CRF = &huge
	err := Validate(s, catalog.Default(), ValidateOpts{})
	mustErrContain(t, err, "video.rate.crf")
}

func TestValidate_twoPassUnsupported(t *testing.T) {
	s := validV2()
	s.Video.Codec = "h264_nvenc"
	s.Video.Preset = "p4"
	s.Video.Rate = VideoRate{Mode: RateModeCBR, Bitrate: "5M"}
	s.Video.TwoPass = true
	err := Validate(s, catalog.Default(), ValidateOpts{})
	mustErrContain(t, err, "video.twoPass")
}

func TestValidate_hwaccelMismatch(t *testing.T) {
	s := validV2()
	s.Hwaccel = &HwaccelSpec{Type: "nvdec"}
	err := Validate(s, catalog.Default(), ValidateOpts{})
	mustErrContain(t, err, "hwaccel")
}

func TestValidate_resolutionOdd(t *testing.T) {
	s := validV2()
	s.Video.Resolution = &Resolution{Width: 1921, Height: 1080}
	err := Validate(s, catalog.Default(), ValidateOpts{})
	mustErrContain(t, err, "resolution.width")
}

func TestValidate_filterUnknown(t *testing.T) {
	s := validV2()
	s.Filters.Video = []FilterStep{{Name: "frobnicate"}}
	err := Validate(s, catalog.Default(), ValidateOpts{})
	mustErrContain(t, err, "frobnicate")
}

func TestValidate_filterScopeMismatch(t *testing.T) {
	s := validV2()
	// loudnorm is audio-scope; using it in video chain must fail.
	s.Filters.Video = []FilterStep{{Name: "loudnorm"}}
	err := Validate(s, catalog.Default(), ValidateOpts{})
	mustErrContain(t, err, "scope")
}

func TestValidate_filterMissingRequired(t *testing.T) {
	s := validV2()
	s.Filters.Video = []FilterStep{{Name: "scale"}} // w, h required
	err := Validate(s, catalog.Default(), ValidateOpts{})
	mustErrContain(t, err, "scale")
	mustErrContain(t, err, "required")
}

func TestValidate_rawExtras_allowed(t *testing.T) {
	s := validV2()
	s.RawExtras = []string{"-map", "0:v:0", "-map", "0:a:0"}
	if err := Validate(s, catalog.Default(), ValidateOpts{}); err != nil {
		t.Errorf("expected valid map allowlist, got %v", err)
	}
}

func TestValidate_rawExtras_deniedFlag(t *testing.T) {
	s := validV2()
	s.RawExtras = []string{"-i", "/etc/passwd"}
	err := Validate(s, catalog.Default(), ValidateOpts{})
	mustErrContain(t, err, "denylisted")
}

func TestValidate_rawExtras_shellInjection(t *testing.T) {
	s := validV2()
	s.RawExtras = []string{"-map", "0:v:0; rm -rf /"}
	err := Validate(s, catalog.Default(), ValidateOpts{})
	mustErrContain(t, err, "forbidden character")
}

func TestValidate_rawExtras_outsideAllowlist(t *testing.T) {
	s := validV2()
	s.RawExtras = []string{"-some-random-flag", "value"}
	err := Validate(s, catalog.Default(), ValidateOpts{})
	mustErrContain(t, err, "allowlist")
}

func TestValidate_rawExtras_permissiveSkipsAllowlist(t *testing.T) {
	s := validV2()
	s.RawExtras = []string{"-some-random-flag", "value"}
	if err := Validate(s, catalog.Default(), ValidateOpts{PermissiveRawExtras: true}); err != nil {
		t.Errorf("permissive mode should skip allowlist, got %v", err)
	}
}

func TestValidate_rawExtras_permissiveStillEnforcesDeny(t *testing.T) {
	s := validV2()
	s.RawExtras = []string{"-i", "/etc/passwd"}
	err := Validate(s, catalog.Default(), ValidateOpts{PermissiveRawExtras: true})
	mustErrContain(t, err, "denylisted")
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

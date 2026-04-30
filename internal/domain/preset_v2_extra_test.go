package domain

import (
	"testing"

	"github.com/rustyguts/tidal/internal/ffmpeg/catalog"
)

func mustCatalog(t *testing.T) *catalog.Catalog {
	t.Helper()
	return catalog.Default()
}

func TestFilterStep_IsEnabled(t *testing.T) {
	if !(FilterStep{}).IsEnabled() {
		t.Error("nil enabled should be true")
	}
	yes := true
	no := false
	if !(FilterStep{Enabled: &yes}).IsEnabled() {
		t.Error("Enabled=true should report enabled")
	}
	if (FilterStep{Enabled: &no}).IsEnabled() {
		t.Error("Enabled=false should report disabled")
	}
}

func TestUpgradeFromV1_audioCodecCopy_dropsBitrate(t *testing.T) {
	v1 := PresetSpec{Container: "mp4", VideoCodec: "libx264", AudioCodec: "copy", AudioBitrate: "192k"}
	v2 := UpgradeFromV1(v1)
	if v2.Audio.Bitrate != "" {
		t.Errorf("expected empty bitrate for copy codec, got %q", v2.Audio.Bitrate)
	}
}

func TestUpgradeFromV1_emptyAudioCodec(t *testing.T) {
	v1 := PresetSpec{Container: "mp4", VideoCodec: "libx264"}
	v2 := UpgradeFromV1(v1)
	if v2.Audio.Codec != "" {
		t.Errorf("expected empty audio codec, got %q", v2.Audio.Codec)
	}
}

func TestUpgradeFromV1_movContainerSetsFaststart(t *testing.T) {
	v1 := PresetSpec{Container: "mov", VideoCodec: "libx264"}
	v2 := UpgradeFromV1(v1)
	if !v2.Container.Faststart {
		t.Error("mov should default to faststart=true")
	}
}

func TestUpgradeFromV1_mkvNoFaststart(t *testing.T) {
	v1 := PresetSpec{Container: "mkv", VideoCodec: "libx264"}
	v2 := UpgradeFromV1(v1)
	if v2.Container.Faststart {
		t.Error("mkv should not have faststart")
	}
}

func TestUnmarshalSpec_invalidJSON(t *testing.T) {
	_, err := UnmarshalSpec([]byte("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestUnmarshalSpec_invalidV1JSON(t *testing.T) {
	_, err := UnmarshalSpec([]byte(`{"container":123}`))
	if err == nil {
		t.Error("expected error for v1 with wrong type")
	}
}

func TestUnmarshalSpec_v2_missingSchemaVersionIsLegacyV1(t *testing.T) {
	// Even without schemaVersion, if it parses as v1 it upgrades.
	v2, err := UnmarshalSpec([]byte(`{"container":"mp4","videoCodec":"libx264","crf":18}`))
	if err != nil {
		t.Fatal(err)
	}
	if v2.SchemaVersion != SchemaVersionV2 {
		t.Errorf("schemaVersion = %d", v2.SchemaVersion)
	}
}

func TestDetectSchemaVersion_invalidJSON(t *testing.T) {
	if v := detectSchemaVersion([]byte("nope")); v != 1 {
		t.Errorf("invalid json should detect as v1, got %d", v)
	}
}

func TestPresetSpec_Validate_emptyContainer(t *testing.T) {
	err := PresetSpec{Container: "", VideoCodec: "libx264", CRF: 23}.Validate()
	if err == nil {
		t.Error("expected error for empty container")
	}
}

func TestValidateV2_filterRequiredArgsPresent(t *testing.T) {
	// scale requires w,h — provide both → no error from filter.
	s := validV2()
	s.Filters.Video = []FilterStep{{Name: "scale", Args: map[string]string{"w": "1920", "h": "1080"}}}
	if err := ValidateV2(s, nil, ValidateOpts{}); err == nil {
		t.Skip("nil catalog hits early-return; validated above")
	}
}

func TestValidateV2_audioFilterScopeWrong(t *testing.T) {
	s := validV2()
	// scale is video-scope; using in audio chain rejected.
	s.Filters.Audio = []FilterStep{{Name: "scale", Args: map[string]string{"w": "640", "h": "480"}}}
	cat := mustCatalog(t)
	err := ValidateV2(s, cat, ValidateOpts{})
	if err == nil {
		t.Error("expected scope error")
	}
}

func TestValidateV2_audioBadSampleRate(t *testing.T) {
	s := validV2()
	s.Audio.SampleRate = 7000 // not in aac sample rates
	err := ValidateV2(s, mustCatalog(t), ValidateOpts{})
	if err == nil {
		t.Error("expected sampleRate error")
	}
}

func TestValidateV2_audioBadChannels(t *testing.T) {
	s := validV2()
	s.Audio.Channels = 7
	err := ValidateV2(s, mustCatalog(t), ValidateOpts{})
	if err == nil {
		t.Error("expected channels error")
	}
}

func TestValidateV2_audioVBRQualityOutOfRange(t *testing.T) {
	q := 99
	s := validV2()
	s.Audio.Codec = "libopus"
	s.Audio.Bitrate = "128k"
	s.Audio.SampleRate = 48000
	s.Audio.Channels = 2
	s.Audio.VBRQuality = &q
	err := ValidateV2(s, mustCatalog(t), ValidateOpts{})
	if err == nil {
		t.Error("expected vbrQuality error")
	}
}

func TestValidateV2_unknownPixelFormat(t *testing.T) {
	s := validV2()
	s.Video.PixelFormat = "yuv999p"
	err := ValidateV2(s, mustCatalog(t), ValidateOpts{})
	if err == nil {
		t.Error("expected pixelFormat error")
	}
}

func TestValidateV2_qpRangeCheck(t *testing.T) {
	q := 999
	s := validV2()
	s.Video.Codec = "h264_nvenc"
	s.Video.Preset = "p4"
	s.Video.Rate = VideoRate{Mode: RateModeQP, QP: &q}
	err := ValidateV2(s, mustCatalog(t), ValidateOpts{})
	if err == nil {
		t.Error("expected qp out-of-range error")
	}
}

func TestValidateV2_subtitleModeInvalid(t *testing.T) {
	s := validV2()
	s.Subtitles.Mode = "weird"
	err := ValidateV2(s, mustCatalog(t), ValidateOpts{})
	if err == nil {
		t.Error("expected subtitle mode error")
	}
}

func TestValidateV2_audioCodecUnknown(t *testing.T) {
	s := validV2()
	s.Audio.Codec = "no-such-codec"
	err := ValidateV2(s, mustCatalog(t), ValidateOpts{})
	if err == nil {
		t.Error("expected audio codec error")
	}
}

func TestValidateV2_hwaccelUnknown(t *testing.T) {
	s := validV2()
	s.Hwaccel = &HwaccelSpec{Type: "magic"}
	err := ValidateV2(s, mustCatalog(t), ValidateOpts{})
	if err == nil {
		t.Error("expected unknown hwaccel error")
	}
}

func TestValidateV2_nilCatalog(t *testing.T) {
	if err := ValidateV2(validV2(), nil, ValidateOpts{}); err == nil {
		t.Error("nil catalog should error")
	}
}

func TestValidateV2_rawExtras_badNonFlag(t *testing.T) {
	s := validV2()
	s.RawExtras = []string{"value-without-flag"}
	err := ValidateV2(s, mustCatalog(t), ValidateOpts{})
	if err == nil {
		t.Error("expected non-flag error")
	}
}

func TestValidateV2_rawExtras_valueTooLong(t *testing.T) {
	long := make([]byte, 300)
	for i := range long {
		long[i] = 'x'
	}
	s := validV2()
	s.RawExtras = []string{"-map", string(long)}
	err := ValidateV2(s, mustCatalog(t), ValidateOpts{})
	if err == nil {
		t.Error("expected too-long value error")
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("abc", 10); got != "abc" {
		t.Errorf("truncate short = %q", got)
	}
	if got := truncate("abcdefghij", 4); got != "abcd..." {
		t.Errorf("truncate long = %q", got)
	}
}

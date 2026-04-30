package domain

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestUpgradeFromV1_h264_1080p(t *testing.T) {
	v1 := PresetSpec{
		Container:    "mp4",
		VideoCodec:   "libx264",
		VideoPreset:  "slow",
		CRF:          20,
		AudioCodec:   "aac",
		AudioBitrate: "192k",
		OutputSuffix: "_1080p",
		Resolution:   &Resolution{Width: 1920, Height: 1080},
	}

	v2 := UpgradeFromV1(v1)

	if v2.SchemaVersion != SchemaVersionV2 {
		t.Fatalf("schemaVersion=%d, want %d", v2.SchemaVersion, SchemaVersionV2)
	}
	if v2.Container.Format != "mp4" || !v2.Container.Faststart {
		t.Errorf("container = %+v; want mp4 + faststart", v2.Container)
	}
	if v2.Video.Codec != "libx264" || v2.Video.Preset != "slow" {
		t.Errorf("video = %+v", v2.Video)
	}
	if v2.Video.Rate.Mode != RateModeCRF || v2.Video.Rate.CRF == nil || *v2.Video.Rate.CRF != 20 {
		t.Errorf("rate = %+v", v2.Video.Rate)
	}
	if v2.Video.Resolution == nil || v2.Video.Resolution.Width != 1920 {
		t.Errorf("resolution = %+v", v2.Video.Resolution)
	}
	if v2.Audio.Codec != "aac" || v2.Audio.Bitrate != "192k" {
		t.Errorf("audio = %+v", v2.Audio)
	}
	if v2.OutputSuffix != "_1080p" {
		t.Errorf("suffix = %q", v2.OutputSuffix)
	}
}

func TestUpgradeFromV1_audio_only(t *testing.T) {
	v1 := PresetSpec{
		Container:    "mp4",
		VideoCodec:   "copy",
		CRF:          0,
		AudioCodec:   "aac",
		AudioBitrate: "192k",
		ExtraArgs:    []string{"-vn"},
		OutputSuffix: "_audio",
	}

	v2 := UpgradeFromV1(v1)

	if v2.Video.Codec != "copy" {
		t.Errorf("video codec = %q", v2.Video.Codec)
	}
	if v2.Video.Rate.Mode != RateModeNone {
		t.Errorf("rate mode = %q, want none for copy", v2.Video.Rate.Mode)
	}
	if !reflect.DeepEqual(v2.RawExtras, []string{"-vn"}) {
		t.Errorf("rawExtras = %v", v2.RawExtras)
	}
}

func TestUnmarshalSpec_v1(t *testing.T) {
	raw := []byte(`{"container":"mkv","videoCodec":"libx265","crf":22,"audioCodec":"aac","audioBitrate":"160k","resolution":{"width":1920,"height":1080}}`)

	v2, err := UnmarshalSpec(raw)
	if err != nil {
		t.Fatal(err)
	}
	if v2.SchemaVersion != SchemaVersionV2 {
		t.Fatalf("schemaVersion=%d", v2.SchemaVersion)
	}
	if v2.Video.Codec != "libx265" {
		t.Errorf("codec = %q", v2.Video.Codec)
	}
	if v2.Video.Rate.CRF == nil || *v2.Video.Rate.CRF != 22 {
		t.Errorf("crf = %+v", v2.Video.Rate.CRF)
	}
}

func TestUnmarshalSpec_v2(t *testing.T) {
	v2 := PresetSpecV2{
		SchemaVersion: SchemaVersionV2,
		Container:     ContainerSpec{Format: "mp4", Faststart: true},
		Video: VideoSpec{
			Codec: "libx264",
			Rate:  VideoRate{Mode: RateModeCRF, CRF: intPtr(23)},
		},
		Audio: AudioSpec{Codec: "aac", Bitrate: "192k"},
	}

	raw, err := json.Marshal(v2)
	if err != nil {
		t.Fatal(err)
	}
	got, err := UnmarshalSpec(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got.Video.Codec != "libx264" || got.Container.Format != "mp4" {
		t.Errorf("round-trip mismatch: %+v", got)
	}
}

func TestUnmarshalSpec_unknownFutureVersion(t *testing.T) {
	// schemaVersion > current must still parse — forward compat for rolling
	// upgrades. Unknown fields are ignored.
	raw := []byte(`{"schemaVersion":3,"container":{"format":"mp4"},"video":{"codec":"libx264","rate":{"mode":"crf","crf":23}},"audio":{"codec":"aac"}}`)
	v2, err := UnmarshalSpec(raw)
	if err != nil {
		t.Fatal(err)
	}
	if v2.Video.Codec != "libx264" {
		t.Errorf("codec = %q", v2.Video.Codec)
	}
}

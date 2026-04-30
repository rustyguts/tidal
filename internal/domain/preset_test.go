package domain

import (
	"encoding/json"
	"testing"
)

func intPtr(v int) *int { return &v }

func TestUnmarshalSpec_roundTrip(t *testing.T) {
	spec := PresetSpec{
		Container: ContainerSpec{Format: "mp4", Faststart: true},
		Video: VideoSpec{
			Codec: "libx264",
			Rate:  VideoRate{Mode: RateModeCRF, CRF: intPtr(23)},
		},
		Audio: AudioSpec{Codec: "aac", Bitrate: "192k"},
	}

	raw, err := json.Marshal(spec)
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

func TestUnmarshalSpec_ignoresUnknownFields(t *testing.T) {
	// Legacy DB rows carry "schemaVersion": 2 plus other historical keys that
	// are no longer part of the struct; json.Unmarshal must drop them silently.
	raw := []byte(`{"schemaVersion":99,"unknownField":"ignored","container":{"format":"mp4"},"video":{"codec":"libx264","rate":{"mode":"crf","crf":23}},"audio":{"codec":"aac"}}`)
	spec, err := UnmarshalSpec(raw)
	if err != nil {
		t.Fatal(err)
	}
	if spec.Video.Codec != "libx264" {
		t.Errorf("codec = %q", spec.Video.Codec)
	}
}

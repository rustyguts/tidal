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

func TestUnmarshalSpec_legacyContainerString(t *testing.T) {
	// Old presets stored container as a plain string like "mp4" instead of
	// an object like {"format":"mp4"}. UnmarshalSpec must handle both.
	raw := []byte(`{"container":"mp4","video":{"codec":"libx264","rate":{"mode":"crf","crf":23}},"audio":{"codec":"aac"}}`)
	spec, err := UnmarshalSpec(raw)
	if err != nil {
		t.Fatal(err)
	}
	if spec.Container.Format != "mp4" {
		t.Errorf("container.format = %q, want %q", spec.Container.Format, "mp4")
	}
}

func TestContainerSpec_unmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantFmt string
		wantErr bool
	}{
		{"object form", `{"format":"mkv","faststart":true}`, "mkv", false},
		{"legacy string", `"webm"`, "webm", false},
		{"invalid", `123`, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var c ContainerSpec
			err := json.Unmarshal([]byte(tt.raw), &c)
			if (err != nil) != tt.wantErr {
				t.Errorf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && c.Format != tt.wantFmt {
				t.Errorf("format = %q, want %q", c.Format, tt.wantFmt)
			}
		})
	}
}

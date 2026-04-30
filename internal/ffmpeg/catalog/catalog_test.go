package catalog

import (
	"encoding/json"
	"testing"
)

func TestDefault_lookups(t *testing.T) {
	c := Default()

	for _, want := range []string{"mp4", "mkv", "webm", "mov"} {
		if _, ok := c.Container(want); !ok {
			t.Errorf("container %q missing", want)
		}
	}
	for _, want := range []string{
		"libx264", "libx265", "libsvtav1", "libvpx-vp9", "libaom-av1",
		"h264_nvenc", "hevc_nvenc", "av1_nvenc",
		"h264_vaapi", "hevc_vaapi", "h264_qsv", "hevc_qsv", "copy",
	} {
		if _, ok := c.VideoCodec(want); !ok {
			t.Errorf("video codec %q missing", want)
		}
	}
	for _, want := range []string{"aac", "libopus", "libmp3lame", "flac", "ac3", "eac3", "copy"} {
		if _, ok := c.AudioCodec(want); !ok {
			t.Errorf("audio codec %q missing", want)
		}
	}
	for _, want := range []string{"none", "nvdec", "vaapi", "qsv", "videotoolbox"} {
		if _, ok := c.Hwaccel(want); !ok {
			t.Errorf("hwaccel %q missing", want)
		}
	}
	for _, want := range []string{"scale", "crop", "yadif", "loudnorm", "subtitles"} {
		if _, ok := c.Filter(want); !ok {
			t.Errorf("filter %q missing", want)
		}
	}
}

// Every video codec used by the seeded builtins (internal/presets/seed.go) must
// appear in the catalog so validation succeeds without manual catalog edits
// when builtins change.
func TestDefault_coversBuiltinCodecs(t *testing.T) {
	c := Default()
	want := []string{"libx264", "libx265", "libsvtav1", "copy"}
	for _, name := range want {
		if _, ok := c.VideoCodec(name); !ok {
			t.Errorf("seeded builtin codec %q missing from catalog", name)
		}
	}
}

func TestJSONSchema_marshallable(t *testing.T) {
	c := Default()
	raw := JSONSchema(c)
	if len(raw) == 0 {
		t.Fatal("empty schema")
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("schema is not valid JSON: %v", err)
	}
	if doc["title"] != "PresetSpec" {
		t.Errorf("title = %v", doc["title"])
	}
	props, ok := doc["properties"].(map[string]any)
	if !ok || props["video"] == nil {
		t.Errorf("video properties missing in schema")
	}
}

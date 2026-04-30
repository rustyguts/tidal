package ffmpeg

import (
	"strings"
	"testing"

	"github.com/rustyguts/tidal/internal/domain"
)

func TestBuildArgs_Simple(t *testing.T) {
	in := RunInput{
		SourcePath: "/media/input.mp4",
		OutputPath: "/media/output.mp4",
		Preset: domain.PresetSpec{
			Container:  "mp4",
			VideoCodec: "libx264",
			CRF:        23,
		},
	}
	args := BuildArgs(in, "")
	assertContains(t, args, "-y")
	assertContains(t, args, "-i")
	assertContains(t, args, "/media/input.mp4")
	assertContains(t, args, "-c:v", "libx264")
	assertContains(t, args, "-crf", "23")
	assertContains(t, args, "/media/output.mp4")
	assertNotContains(t, args, "-progress")
}

func TestBuildArgs_WithProgress(t *testing.T) {
	in := RunInput{
		SourcePath: "in.mp4",
		OutputPath: "out.mp4",
		Preset: domain.PresetSpec{
			Container:  "mp4",
			VideoCodec: "libx264",
			CRF:        23,
		},
	}
	args := BuildArgs(in, "http://localhost:8080/progress")
	assertContains(t, args, "-progress", "http://localhost:8080/progress")
}

func TestBuildArgs_WithResolution(t *testing.T) {
	in := RunInput{
		SourcePath: "in.mp4",
		OutputPath: "out.mp4",
		Preset: domain.PresetSpec{
			Container:  "mp4",
			VideoCodec: "libx264",
			CRF:        23,
			Resolution: &domain.Resolution{Width: 1920, Height: 1080},
		},
	}
	args := BuildArgs(in, "")
	assertContains(t, args, "-vf", "scale=1920:1080")
}

func TestBuildArgs_PresetOptions(t *testing.T) {
	in := RunInput{
		SourcePath: "in.mp4",
		OutputPath: "out.mp4",
		Preset: domain.PresetSpec{
			Container:    "mp4",
			VideoCodec:   "libx265",
			VideoPreset:  "slow",
			CRF:          28,
			PixelFormat:  "yuv420p",
			AudioCodec:   "aac",
			AudioBitrate: "192k",
			ExtraArgs:    []string{"-tag:v", "hvc1"},
		},
	}
	args := BuildArgs(in, "")
	assertContains(t, args, "-c:v", "libx265")
	assertContains(t, args, "-preset", "slow")
	assertContains(t, args, "-crf", "28")
	assertContains(t, args, "-pix_fmt", "yuv420p")
	assertContains(t, args, "-c:a", "aac")
	assertContains(t, args, "-b:a", "192k")
	assertContains(t, args, "-tag:v", "hvc1")
}

func TestBuildArgs_VideoCodecCopy(t *testing.T) {
	in := RunInput{
		SourcePath: "in.mp4",
		OutputPath: "out.mp4",
		Preset: domain.PresetSpec{
			Container:   "mp4",
			VideoCodec:  "copy",
			VideoPreset: "slow",
			CRF:         23,
			PixelFormat: "yuv420p",
		},
	}
	args := BuildArgs(in, "")
	assertContains(t, args, "-c:v", "copy")
	assertNotContains(t, args, "-preset")
	assertNotContains(t, args, "-crf")
	assertNotContains(t, args, "-pix_fmt")
}

func TestBuildArgs_AudioCodecCopyNoBitrate(t *testing.T) {
	in := RunInput{
		SourcePath: "in.mp4",
		OutputPath: "out.mp4",
		Preset: domain.PresetSpec{
			Container:   "mp4",
			VideoCodec:  "libx264",
			CRF:         23,
			AudioCodec:  "copy",
			AudioBitrate: "192k",
		},
	}
	args := BuildArgs(in, "")
	assertContains(t, args, "-c:a", "copy")
	assertNotContains(t, args, "-b:a")
}

func TestParseBitrate(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"1000kbits", 1_000_000},
		{"1.5Mbits", 1_500_000},
		{"N/A", 0},
		{"", 0},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := parseBitrate(tt.input); got != tt.want {
				t.Errorf("parseBitrate(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseSpeed(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"1.5x", 1.5},
		{"0.5x", 0.5},
		{"N/A", 0},
		{"", 0},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := parseSpeed(tt.input); got != tt.want {
				t.Errorf("parseSpeed(%q) = %f, want %f", tt.input, got, tt.want)
			}
		})
	}
}

func TestReadProgress_ParsesFields(t *testing.T) {
	input := `frame=100
fps=30.0
out_time_us=5000000
bitrate=1000kbits
total_size=65536
speed=1.5x
progress=continue

frame=200
out_time_ms=10000000
progress=end
`

	var progresses []domain.FFmpegProgress
	onProgress := func(p domain.FFmpegProgress) {
		progresses = append(progresses, p)
	}

	readProgress(strings.NewReader(input), 20000, onProgress)

	if len(progresses) != 2 {
		t.Fatalf("got %d progress updates, want 2", len(progresses))
	}

	p1 := progresses[0]
	if p1.Frame != 100 {
		t.Errorf("Frame = %d, want 100", p1.Frame)
	}
	if p1.FPS != 30.0 {
		t.Errorf("FPS = %f, want 30.0", p1.FPS)
	}
	if p1.TimeMs != 5000 {
		t.Errorf("TimeMs = %d, want 5000", p1.TimeMs)
	}
	if p1.Bitrate != 1_000_000 {
		t.Errorf("Bitrate = %d, want 1000000", p1.Bitrate)
	}
	if p1.SizeBytes != 65536 {
		t.Errorf("SizeBytes = %d, want 65536", p1.SizeBytes)
	}
	if p1.Speed != 1.5 {
		t.Errorf("Speed = %f, want 1.5", p1.Speed)
	}
	if p1.Percent == 0 {
		t.Error("Percent should be non-zero when duration provided")
	}

	p2 := progresses[1]
	if p2.Frame != 200 {
		t.Errorf("Frame = %d, want 200", p2.Frame)
	}
}

func TestReadProgress_PercentClamped(t *testing.T) {
	// time_ms exceeds duration_ms
	input := "out_time_us=100000000\nprogress=continue\n"
	var p domain.FFmpegProgress
	readProgress(strings.NewReader(input), 1000, func(prog domain.FFmpegProgress) {
		p = prog
	})
	if p.Percent != 100 {
		t.Errorf("Percent = %f, want 100", p.Percent)
	}
}

func TestReadProgress_NilCallbackDiscards(t *testing.T) {
	readProgress(strings.NewReader("frame=1\nprogress=continue\n"), 1000, nil)
}

func assertContains(t *testing.T, args []string, expected ...string) {
	t.Helper()
	for _, e := range expected {
		found := false
		for _, a := range args {
			if a == e {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("args %v missing %q", args, e)
		}
	}
}

func assertNotContains(t *testing.T, args []string, unexpected ...string) {
	t.Helper()
	for _, u := range unexpected {
		for _, a := range args {
			if a == u {
				t.Errorf("args %v contains unexpected %q", args, u)
			}
		}
	}
}

func TestReadLogs_NilCallbackDiscards(t *testing.T) {
	readLogs(strings.NewReader("some log line\nanother line\n"), "stderr", nil)
}

func TestReadLogs_WithCallback(t *testing.T) {
	input := "line1\nline2\n"
	var logs []LogLine
	readLogs(strings.NewReader(input), "stderr", func(l LogLine) {
		logs = append(logs, l)
	})
	if len(logs) != 2 {
		t.Fatalf("got %d log lines, want 2", len(logs))
	}
	if logs[0].Stream != "stderr" {
		t.Errorf("Stream = %q, want stderr", logs[0].Stream)
	}
	if logs[0].Line != "line1" {
		t.Errorf("Line = %q, want line1", logs[0].Line)
	}
}

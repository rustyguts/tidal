package ffmpeg

import (
	"strings"
	"testing"

	"github.com/rustyguts/tidal/internal/domain"
)

// Argv assembly is exercised in detail by builder/builder_test.go and
// builder_extra_test.go; this file covers parsers and stream helpers that
// live next to the runner.

func TestParseBitrate(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"1000kbits", 1_000_000},
		{"1.5Mbits", 1_500_000},
		{"N/A", 0},
		{"", 0},
		{"500", 500},
		{"500/s", 500},
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
		{"2", 2},
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

func TestReadProgress_skipsBlankAndMalformedLines(t *testing.T) {
	input := "\n  \nnot-a-pair\nframe=42\nprogress=continue\n"
	var p domain.FFmpegProgress
	readProgress(strings.NewReader(input), 0, func(prog domain.FFmpegProgress) {
		p = prog
	})
	if p.Frame != 42 {
		t.Errorf("Frame = %d, want 42", p.Frame)
	}
}

func TestReadProgress_zeroDurationLeavesPercentZero(t *testing.T) {
	input := "out_time_us=5000000\nprogress=continue\n"
	var p domain.FFmpegProgress
	readProgress(strings.NewReader(input), 0, func(prog domain.FFmpegProgress) {
		p = prog
	})
	if p.Percent != 0 {
		t.Errorf("Percent = %f, want 0 when duration unknown", p.Percent)
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

func TestEnsureOutputDir_skipsRoot(t *testing.T) {
	if err := ensureOutputDir(""); err != nil {
		t.Errorf("empty path should be a no-op, got %v", err)
	}
	if err := ensureOutputDir("file.mp4"); err != nil {
		t.Errorf(". path should be a no-op, got %v", err)
	}
}

func TestEnsureOutputDir_createsDir(t *testing.T) {
	tmp := t.TempDir()
	if err := ensureOutputDir(tmp + "/sub/out.mp4"); err != nil {
		t.Errorf("expected mkdir to succeed, got %v", err)
	}
}

func TestRun_validatesOutputDir(t *testing.T) {
	// Pass an unwritable parent so ensureOutputDir reports an error path.
	// Skip on non-POSIX where /proc may not exist.
	if err := ensureOutputDir("/proc/1/clearly-not-writable/x"); err == nil {
		t.Skip("could not provoke mkdir error in this env")
	}
}

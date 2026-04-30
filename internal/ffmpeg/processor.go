package ffmpeg

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rustyguts/tidal/internal/domain"
)

// LogLine is a single stderr/stdout line emitted by ffmpeg.
type LogLine struct {
	Stream string // "stderr" | "stdout" | "system"
	Line   string
	At     time.Time
}

// Hooks lets a caller observe progress + logs without coupling to a transport.
type Hooks struct {
	OnProgress func(domain.FFmpegProgress)
	OnLog      func(LogLine)
}

type RunInput struct {
	Preset     domain.PresetSpec
	SourcePath string
	OutputPath string
	DurationMs int64 // optional; if 0, percent stays at 0
}

// BuildArgs assembles the ffmpeg argv. Caller passes already-validated input.
func BuildArgs(in RunInput, progressURL string) []string {
	args := []string{
		"-y",
		"-hide_banner",
		"-nostats",
		"-loglevel", "info",
	}
	if progressURL != "" {
		args = append(args, "-progress", progressURL)
	}
	args = append(args, "-i", in.SourcePath)

	if in.Preset.Resolution != nil {
		args = append(args, "-vf", fmt.Sprintf("scale=%d:%d", in.Preset.Resolution.Width, in.Preset.Resolution.Height))
	}
	if in.Preset.VideoCodec != "" {
		args = append(args, "-c:v", in.Preset.VideoCodec)
	}
	if in.Preset.VideoCodec != "copy" {
		if in.Preset.VideoPreset != "" {
			args = append(args, "-preset", in.Preset.VideoPreset)
		}
		if in.Preset.CRF > 0 {
			args = append(args, "-crf", strconv.Itoa(in.Preset.CRF))
		}
		if in.Preset.PixelFormat != "" {
			args = append(args, "-pix_fmt", in.Preset.PixelFormat)
		}
	}
	if in.Preset.AudioCodec != "" {
		args = append(args, "-c:a", in.Preset.AudioCodec)
	}
	if in.Preset.AudioCodec != "" && in.Preset.AudioCodec != "copy" && in.Preset.AudioBitrate != "" {
		args = append(args, "-b:a", in.Preset.AudioBitrate)
	}
	args = append(args, in.Preset.ExtraArgs...)
	args = append(args, in.OutputPath)
	return args
}

// Run executes ffmpeg, streaming progress to hooks. Cancels on ctx.Done.
func Run(ctx context.Context, in RunInput, hooks Hooks) error {
	if err := ensureOutputDir(in.OutputPath); err != nil {
		return err
	}

	args := BuildArgs(in, "pipe:1")
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	cmd.Cancel = func() error { return cmd.Process.Signal(interruptSignal()) }

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr pipe: %w", err)
	}

	if hooks.OnLog != nil {
		hooks.OnLog(LogLine{Stream: "system", Line: "ffmpeg " + strings.Join(args, " "), At: time.Now()})
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start ffmpeg: %w", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		readProgress(stdout, in.DurationMs, hooks.OnProgress)
	}()
	go func() {
		defer wg.Done()
		readLogs(stderr, "stderr", hooks.OnLog)
	}()

	wg.Wait()
	if err := cmd.Wait(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("ffmpeg exit %d", exitErr.ExitCode())
		}
		return fmt.Errorf("ffmpeg wait: %w", err)
	}
	return nil
}

func ensureOutputDir(out string) error {
	dir := filepath.Dir(out)
	if dir == "" || dir == "." {
		return nil
	}
	return mkdirAll(dir, 0o755)
}

func readLogs(r io.Reader, stream string, onLog func(LogLine)) {
	if onLog == nil {
		_, _ = io.Copy(io.Discard, r)
		return
	}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		onLog(LogLine{Stream: stream, Line: sc.Text(), At: time.Now()})
	}
}

// readProgress parses the key=value stream emitted by `-progress pipe:1`.
// ffmpeg writes a block ending in `progress=continue` (or `progress=end`)
// roughly every 0.5–1s by default.
func readProgress(r io.Reader, durationMs int64, onProgress func(domain.FFmpegProgress)) {
	if onProgress == nil {
		_, _ = io.Copy(io.Discard, r)
		return
	}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 32*1024), 256*1024)

	cur := domain.FFmpegProgress{}
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		v = strings.TrimSpace(v)
		switch k {
		case "frame":
			if n, err := strconv.ParseInt(v, 10, 64); err == nil {
				cur.Frame = n
			}
		case "fps":
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				cur.FPS = f
			}
		case "out_time_us", "out_time_ms":
			// ffmpeg reports microseconds in this key (despite the "_ms" name).
			if n, err := strconv.ParseInt(v, 10, 64); err == nil {
				cur.TimeMs = n / 1000
			}
		case "bitrate":
			cur.Bitrate = parseBitrate(v)
		case "total_size":
			if n, err := strconv.ParseInt(v, 10, 64); err == nil {
				cur.SizeBytes = n
			}
		case "speed":
			cur.Speed = parseSpeed(v)
		case "progress":
			cur.UpdatedAt = time.Now()
			if durationMs > 0 {
				p := float64(cur.TimeMs) / float64(durationMs) * 100
				if p > 100 {
					p = 100
				}
				cur.Percent = p
			}
			onProgress(cur)
			if v == "end" {
				return
			}
		}
	}
}

func parseBitrate(s string) int64 {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "/s")
	switch {
	case strings.HasSuffix(s, "kbits"):
		n, _ := strconv.ParseFloat(strings.TrimSuffix(s, "kbits"), 64)
		return int64(n * 1000)
	case strings.HasSuffix(s, "Mbits"):
		n, _ := strconv.ParseFloat(strings.TrimSuffix(s, "Mbits"), 64)
		return int64(n * 1_000_000)
	}
	if s == "N/A" {
		return 0
	}
	n, _ := strconv.ParseFloat(s, 64)
	return int64(n)
}

func parseSpeed(s string) float64 {
	s = strings.TrimSpace(strings.TrimSuffix(s, "x"))
	if s == "" || s == "N/A" {
		return 0
	}
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

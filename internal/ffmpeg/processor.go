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
	"github.com/rustyguts/tidal/internal/ffmpeg/builder"
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

// RunInput is the input to Run. Spec is the structured preset;
// SourcePath/OutputPath/DurationMs identify the source + sink.
type RunInput struct {
	Spec       domain.PresetSpec
	SourcePath string
	OutputPath string
	DurationMs int64
}

// Run executes ffmpeg from a structured preset spec, handling two-pass when
// requested. Cancels on ctx.Done.
func Run(ctx context.Context, in RunInput, hooks Hooks) error {
	if err := ensureOutputDir(in.OutputPath); err != nil {
		return err
	}

	if in.Spec.Video.TwoPass {
		prefix := builder.PassPrefixForOutput(in.OutputPath)
		for pass := 1; pass <= 2; pass++ {
			args, err := builder.Compose(builder.Context{
				InputPath:     in.SourcePath,
				OutputPath:    in.OutputPath,
				ProgressURL:   "pipe:1",
				PassLogPrefix: prefix,
				Pass:          pass,
			}, in.Spec)
			if err != nil {
				return fmt.Errorf("compose pass %d: %w", pass, err)
			}
			if err := execFfmpeg(ctx, args, in.DurationMs, hooks); err != nil {
				return err
			}
		}
		builder.CleanupPassLogs(prefix)
		return nil
	}

	args, err := builder.Compose(builder.Context{
		InputPath:   in.SourcePath,
		OutputPath:  in.OutputPath,
		ProgressURL: "pipe:1",
	}, in.Spec)
	if err != nil {
		return fmt.Errorf("compose: %w", err)
	}
	return execFfmpeg(ctx, args, in.DurationMs, hooks)
}

// execFfmpeg runs ffmpeg with the supplied argv, wires progress + log hooks,
// and propagates ctx cancellation. Reused by Run and each two-pass iteration.
func execFfmpeg(ctx context.Context, args []string, durationMs int64, hooks Hooks) error {
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
		readProgress(stdout, durationMs, hooks.OnProgress)
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

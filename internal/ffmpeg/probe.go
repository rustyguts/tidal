package ffmpeg

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"time"
)

type ProbeResult struct {
	DurationMs int64    `json:"durationMs"`
	Width      int      `json:"width,omitempty"`
	Height     int      `json:"height,omitempty"`
	VideoCodec string   `json:"videoCodec,omitempty"`
	AudioCodec string   `json:"audioCodec,omitempty"`
	Bitrate    int64    `json:"bitrate,omitempty"`
	Streams    int      `json:"streams"`
	Raw        []byte   `json:"-"`
}

type rawProbe struct {
	Format struct {
		Duration string `json:"duration"`
		BitRate  string `json:"bit_rate"`
	} `json:"format"`
	Streams []struct {
		CodecType string `json:"codec_type"`
		CodecName string `json:"codec_name"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
	} `json:"streams"`
}

func Probe(ctx context.Context, path string) (ProbeResult, error) {
	cctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	out, err := exec.CommandContext(cctx, "ffprobe",
		"-v", "error",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		path,
	).Output()
	if err != nil {
		return ProbeResult{}, fmt.Errorf("ffprobe %s: %w", path, err)
	}
	var rp rawProbe
	if err := json.Unmarshal(out, &rp); err != nil {
		return ProbeResult{}, fmt.Errorf("ffprobe parse: %w", err)
	}
	res := ProbeResult{Raw: out, Streams: len(rp.Streams)}
	if d, err := strconv.ParseFloat(rp.Format.Duration, 64); err == nil {
		res.DurationMs = int64(d * 1000)
	}
	if br, err := strconv.ParseInt(rp.Format.BitRate, 10, 64); err == nil {
		res.Bitrate = br
	}
	for _, s := range rp.Streams {
		switch s.CodecType {
		case "video":
			if res.VideoCodec == "" {
				res.VideoCodec = s.CodecName
				res.Width = s.Width
				res.Height = s.Height
			}
		case "audio":
			if res.AudioCodec == "" {
				res.AudioCodec = s.CodecName
			}
		}
	}
	return res, nil
}

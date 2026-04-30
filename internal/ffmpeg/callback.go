package ffmpeg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rustyguts/tidal/internal/domain"
)

// JobSpecResponse mirrors the Go server's `/api/internal/jobs/:id/spec` body.
type JobSpecResponse struct {
	JobID      string              `json:"jobId"`
	Spec       domain.PresetSpec `json:"spec"`
	SourcePath string              `json:"sourcePath"`
	OutputPath string              `json:"outputPath"`
	CachePath  string              `json:"cachePath,omitempty"`
}

// CallbackClient is used by `tidal runjob` to fetch the spec and post status,
// progress, and log batches back to the Tidal server. The /api/internal/*
// endpoints are unauthenticated by design — restrict via NetworkPolicy or
// in-cluster networking.
type CallbackClient struct {
	base string
	hc   *http.Client
}

func NewCallbackClient(serverURL string) *CallbackClient {
	return &CallbackClient{
		base: strings.TrimRight(serverURL, "/"),
		hc:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *CallbackClient) Spec(ctx context.Context, jobID string) (JobSpecResponse, error) {
	var out JobSpecResponse
	return out, c.do(ctx, http.MethodGet, "/api/internal/jobs/"+jobID+"/spec", nil, &out)
}

func (c *CallbackClient) Status(ctx context.Context, jobID, status, errMsg string) error {
	return c.do(ctx, http.MethodPost, "/api/internal/jobs/"+jobID+"/status",
		map[string]any{"status": status, "error": errMsg}, nil)
}

func (c *CallbackClient) Progress(ctx context.Context, jobID string, p domain.FFmpegProgress) error {
	return c.do(ctx, http.MethodPost, "/api/internal/jobs/"+jobID+"/progress", p, nil)
}

type LogBatch struct {
	Lines []LogBatchLine `json:"lines"`
}
type LogBatchLine struct {
	Stream string `json:"stream"`
	Line   string `json:"line"`
}

func (c *CallbackClient) Logs(ctx context.Context, jobID string, batch LogBatch) error {
	return c.do(ctx, http.MethodPost, "/api/internal/jobs/"+jobID+"/log", batch, nil)
}

func (c *CallbackClient) do(ctx context.Context, method, path string, body, out any) error {
	var rb io.Reader
	if body != nil {
		buf := &bytes.Buffer{}
		if err := json.NewEncoder(buf).Encode(body); err != nil {
			return err
		}
		rb = buf
	}
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, rb)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.hc.Do(req)
	if err != nil {
		return fmt.Errorf("%s %s: %w", method, path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s %s: %d %s", method, path, resp.StatusCode, strings.TrimSpace(string(b)))
	}
	if out != nil && resp.StatusCode != http.StatusNoContent {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

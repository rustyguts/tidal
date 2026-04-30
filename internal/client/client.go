package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/rustyguts/tidal/internal/domain"
)

// Client is a thin REST client used by the `tidal job` CLI and the per-Job pod
// for `/api/internal/*` callbacks (Phase 5).
type Client struct {
	base string
	hc   *http.Client
}

func New(baseURL string) *Client {
	base := strings.TrimRight(baseURL, "/")
	return &Client{base: base, hc: &http.Client{Timeout: 30 * time.Second}}
}

type EnqueueInput struct {
	PresetID       string `json:"presetId"`
	SourcePath     string `json:"sourcePath"`
	OutputPath     string `json:"outputPath,omitempty"`
	CachePath      string `json:"cachePath,omitempty"`
	SourceMovePath string `json:"sourceMovePath,omitempty"`
}

func (c *Client) ListPresets(ctx context.Context) ([]domain.Preset, error) {
	var out []domain.Preset
	return out, c.do(ctx, http.MethodGet, "/api/presets", nil, &out)
}

func (c *Client) PresetByName(ctx context.Context, name string) (domain.Preset, error) {
	all, err := c.ListPresets(ctx)
	if err != nil {
		return domain.Preset{}, err
	}
	for _, p := range all {
		if p.Name == name {
			return p, nil
		}
	}
	return domain.Preset{}, fmt.Errorf("preset %q not found", name)
}

func (c *Client) EnqueueJob(ctx context.Context, in EnqueueInput) (domain.Job, error) {
	var out domain.Job
	return out, c.do(ctx, http.MethodPost, "/api/jobs", in, &out)
}

func (c *Client) ListJobs(ctx context.Context, status string, limit int) ([]domain.Job, error) {
	q := ""
	if status != "" || limit > 0 {
		q = "?"
		if status != "" {
			q += "status=" + status
		}
		if limit > 0 {
			if !strings.HasSuffix(q, "?") {
				q += "&"
			}
			q += fmt.Sprintf("limit=%d", limit)
		}
	}
	var out []domain.Job
	return out, c.do(ctx, http.MethodGet, "/api/jobs"+q, nil, &out)
}

func (c *Client) GetJob(ctx context.Context, id uuid.UUID) (domain.Job, error) {
	var out domain.Job
	return out, c.do(ctx, http.MethodGet, "/api/jobs/"+id.String(), nil, &out)
}

func (c *Client) CancelJob(ctx context.Context, id uuid.UUID) error {
	return c.do(ctx, http.MethodDelete, "/api/jobs/"+id.String(), nil, nil)
}

func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	var reqBody io.Reader
	if body != nil {
		buf := &bytes.Buffer{}
		if err := json.NewEncoder(buf).Encode(body); err != nil {
			return err
		}
		reqBody = buf
	}
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, reqBody)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		return err
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

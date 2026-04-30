package ffmpeg

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rustyguts/tidal/internal/domain"
)

func newTestServer(t *testing.T, h http.Handler) (*httptest.Server, *CallbackClient) {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return srv, NewCallbackClient(srv.URL)
}

func TestCallback_Spec(t *testing.T) {
	srv, c := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/internal/jobs/abc/spec" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(JobSpecResponse{
			JobID: "abc", SourcePath: "/in", OutputPath: "/out",
			Spec: domain.PresetSpec{Container: domain.ContainerSpec{Format: "mp4"}, Video: domain.VideoSpec{Codec: "libx264"}, Audio: domain.AudioSpec{Codec: "aac"}},
		})
	}))
	_ = srv

	got, err := c.Spec(context.Background(), "abc")
	if err != nil {
		t.Fatal(err)
	}
	if got.JobID != "abc" || got.SourcePath != "/in" {
		t.Errorf("got %+v", got)
	}
	if got.Spec.Video.Codec != "libx264" {
		t.Errorf("spec codec = %q", got.Spec.Video.Codec)
	}
}

func TestCallback_StatusProgressLogs(t *testing.T) {
	type recv struct {
		method string
		path   string
		body   string
	}
	calls := []recv{}
	srv, c := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		calls = append(calls, recv{method: r.Method, path: r.URL.Path, body: string(b)})
		w.WriteHeader(http.StatusNoContent)
	}))
	_ = srv

	if err := c.Status(context.Background(), "j", "running", ""); err != nil {
		t.Fatal(err)
	}
	if err := c.Progress(context.Background(), "j", domain.FFmpegProgress{Frame: 42}); err != nil {
		t.Fatal(err)
	}
	if err := c.Logs(context.Background(), "j", LogBatch{Lines: []LogBatchLine{{Stream: "stderr", Line: "x"}}}); err != nil {
		t.Fatal(err)
	}

	if len(calls) != 3 {
		t.Fatalf("expected 3 calls, got %d", len(calls))
	}
	if calls[0].path != "/api/internal/jobs/j/status" {
		t.Errorf("status path = %q", calls[0].path)
	}
	if !strings.Contains(calls[0].body, `"status":"running"`) {
		t.Errorf("status body = %s", calls[0].body)
	}
	if calls[1].path != "/api/internal/jobs/j/progress" {
		t.Errorf("progress path = %q", calls[1].path)
	}
	if !strings.Contains(calls[1].body, `"frame":42`) {
		t.Errorf("progress body = %s", calls[1].body)
	}
	if calls[2].path != "/api/internal/jobs/j/log" {
		t.Errorf("log path = %q", calls[2].path)
	}
}

func TestCallback_4xx_returnsError(t *testing.T) {
	srv, c := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid"))
	}))
	_ = srv
	err := c.Status(context.Background(), "j", "x", "")
	if err == nil || !strings.Contains(err.Error(), "400") {
		t.Errorf("expected 400 error, got %v", err)
	}
}

func TestCallback_5xx_returnsError(t *testing.T) {
	srv, c := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	_ = srv
	_, err := c.Spec(context.Background(), "j")
	if err == nil {
		t.Error("expected error on 500")
	}
}

func TestCallback_baseURLNormalized(t *testing.T) {
	c := NewCallbackClient("http://example.com/")
	if !strings.HasSuffix(c.base, "example.com") {
		t.Errorf("trailing slash should be trimmed; base = %q", c.base)
	}
}

func TestCallback_invalidServerURL(t *testing.T) {
	c := NewCallbackClient("http://127.0.0.1:1") // unreachable port
	err := c.Status(context.Background(), "j", "x", "")
	if err == nil {
		t.Error("expected connection error")
	}
}

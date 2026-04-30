package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/rustyguts/tidal/internal/domain"
)

func postPreview(t *testing.T, body string) *previewResponse {
	t.Helper()
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/presets/preview", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewPresetPreview(nil, domain.ValidateOpts{})
	if err := h.Post(c); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var resp previewResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	return &resp
}

func TestPreview_validSpec(t *testing.T) {
	body := `{"spec":{"container":{"format":"mp4"},"video":{"codec":"libx264","preset":"slow","rate":{"mode":"crf","crf":20}},"audio":{"codec":"aac","bitrate":"192k"}}}`
	resp := postPreview(t, body)
	if len(resp.Errors) != 0 {
		t.Errorf("unexpected errors: %v", resp.Errors)
	}
	if !containsArg(resp.Argv, "-c:v", "libx264") {
		t.Errorf("argv missing libx264: %v", resp.Argv)
	}
}

func TestPreview_invalidCRF(t *testing.T) {
	body := `{"spec":{"container":{"format":"mp4"},"video":{"codec":"libx264","rate":{"mode":"crf","crf":99}},"audio":{"codec":"aac"}}}`
	resp := postPreview(t, body)
	if len(resp.Errors) == 0 {
		t.Errorf("expected validation errors for crf=99")
	}
}

func TestPreview_unknownCodec(t *testing.T) {
	body := `{"spec":{"container":{"format":"mp4"},"video":{"codec":"vc1","rate":{"mode":"crf","crf":20}},"audio":{"codec":"aac"}}}`
	resp := postPreview(t, body)
	if len(resp.Errors) == 0 {
		t.Errorf("expected error for unknown codec")
	}
}

func TestPreview_explicitPaths(t *testing.T) {
	body := `{"spec":{"container":{"format":"mp4"},"video":{"codec":"libx264","rate":{"mode":"crf","crf":23}},"audio":{"codec":"aac"}},"inputPath":"/in.mov","outputPath":"/out.mp4"}`
	resp := postPreview(t, body)
	if !containsArg(resp.Argv, "-i", "/in.mov") {
		t.Errorf("expected -i /in.mov in argv: %v", resp.Argv)
	}
}

func TestPreview_invalidBody(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/presets/preview", strings.NewReader(`not json`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := NewPresetPreview(nil, domain.ValidateOpts{})
	err := h.Post(c)
	if err == nil {
		t.Error("expected 400 on bad body")
	}
}

func TestPreview_emptySpec(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/presets/preview", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := NewPresetPreview(nil, domain.ValidateOpts{})
	err := h.Post(c)
	if err == nil {
		t.Error("expected error on missing spec")
	}
}

func TestPreview_unparseableSpec(t *testing.T) {
	body := `{"spec":"not-an-object"}`
	resp := postPreview(t, body)
	if len(resp.Errors) == 0 {
		t.Error("expected parse errors")
	}
}

func TestPreview_placeholderPaths(t *testing.T) {
	body := `{"spec":{"container":{"format":"mp4"},"video":{"codec":"libx264","rate":{"mode":"crf","crf":23}},"audio":{"codec":"aac"}}}`
	resp := postPreview(t, body)
	hasInput := false
	hasOutput := false
	for _, a := range resp.Argv {
		if a == "<input>" {
			hasInput = true
		}
		if a == "<output>" {
			hasOutput = true
		}
	}
	if !hasInput || !hasOutput {
		t.Errorf("placeholder paths missing in argv: %v", resp.Argv)
	}
}

func containsArg(args []string, flag, val string) bool {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == flag && args[i+1] == val {
			return true
		}
	}
	return false
}

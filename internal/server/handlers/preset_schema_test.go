package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestPresetSchema_Get(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/presets/schema", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewPresetSchema(nil)
	if err := h.Get(c); err != nil {
		t.Fatalf("handler: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var doc map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &doc); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if doc["schemaVersion"].(float64) != 2 {
		t.Errorf("schemaVersion = %v", doc["schemaVersion"])
	}
	if doc["catalog"] == nil {
		t.Errorf("catalog missing")
	}
	if doc["schema"] == nil {
		t.Errorf("schema missing")
	}
}

func TestPresetSchema_cachesPayload(t *testing.T) {
	h := NewPresetSchema(nil)
	e := echo.New()
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/presets/schema", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		if err := h.Get(c); err != nil {
			t.Fatalf("call %d: %v", i, err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("call %d status %d", i, rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "PresetSpecV2") {
			t.Errorf("call %d body missing schema title", i)
		}
	}
}

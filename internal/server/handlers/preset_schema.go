package handlers

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"

	"github.com/rustyguts/tidal/internal/ffmpeg/catalog"
)

// PresetSchema serves the catalog + JSON Schema describing PresetSpecV2.
// Frontend caches the response and uses it to drive the preset editor's
// codec-aware dropdowns and field tooltips.
type PresetSchema struct {
	cat *catalog.Catalog

	once    sync.Once
	payload []byte
	err     error
}

func NewPresetSchema(cat *catalog.Catalog) *PresetSchema {
	if cat == nil {
		cat = catalog.Default()
	}
	return &PresetSchema{cat: cat}
}

type schemaResponse struct {
	SchemaVersion int              `json:"schemaVersion"`
	Catalog       *catalog.Catalog `json:"catalog"`
	Schema        json.RawMessage  `json:"schema"`
}

func (h *PresetSchema) Get(c echo.Context) error {
	h.once.Do(func() {
		body, err := json.Marshal(schemaResponse{
			SchemaVersion: 2,
			Catalog:       h.cat,
			Schema:        catalog.JSONSchema(h.cat),
		})
		if err != nil {
			h.err = err
			return
		}
		h.payload = body
	})
	if h.err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, h.err.Error())
	}
	return c.JSONBlob(http.StatusOK, h.payload)
}

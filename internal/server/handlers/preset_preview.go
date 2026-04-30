package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/rustyguts/tidal/internal/domain"
	"github.com/rustyguts/tidal/internal/ffmpeg/builder"
	"github.com/rustyguts/tidal/internal/ffmpeg/catalog"
)

// PresetPreview validates a candidate preset spec and returns the ffmpeg argv
// the builder would emit for it. Never executes ffmpeg. Used by the editor
// for live "what will the encode command look like" feedback.
type PresetPreview struct {
	cat  *catalog.Catalog
	opts domain.ValidateOpts
}

func NewPresetPreview(cat *catalog.Catalog, opts domain.ValidateOpts) *PresetPreview {
	if cat == nil {
		cat = catalog.Default()
	}
	return &PresetPreview{cat: cat, opts: opts}
}

type previewRequest struct {
	Spec       json.RawMessage `json:"spec"`
	InputPath  string          `json:"inputPath,omitempty"`  // placeholder; default <input>
	OutputPath string          `json:"outputPath,omitempty"` // placeholder; default <output>
}

type previewResponse struct {
	Argv     []string `json:"argv"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
	// Spec is the spec after V1→V2 upgrade so the UI sees the canonical V2
	// shape it should round-trip on save.
	Spec domain.PresetSpec `json:"spec"`
}

func (h *PresetPreview) Post(c echo.Context) error {
	var req previewRequest
	if err := c.Bind(&req); err != nil || len(req.Spec) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "spec required")
	}
	v2, err := domain.UnmarshalSpec(req.Spec)
	if err != nil {
		return c.JSON(http.StatusOK, previewResponse{
			Errors: []string{"spec parse: " + err.Error()},
		})
	}
	resp := previewResponse{Spec: v2}
	if err := domain.Validate(v2, h.cat, h.opts); err != nil {
		resp.Errors = append(resp.Errors, err.Error())
	}

	in := req.InputPath
	if in == "" {
		in = "<input>"
	}
	out := req.OutputPath
	if out == "" {
		out = "<output>"
	}
	args, buildErr := builder.Compose(builder.Context{
		InputPath:  in,
		OutputPath: out,
	}, v2)
	if buildErr != nil {
		resp.Errors = append(resp.Errors, "compose: "+buildErr.Error())
	}
	resp.Argv = args
	return c.JSON(http.StatusOK, resp)
}

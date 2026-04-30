package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/rustyguts/tidal/internal/domain"
	"github.com/rustyguts/tidal/internal/jobs"
	"github.com/rustyguts/tidal/internal/presets"
)

// JobCallbacks handles `/api/internal/*` endpoints called by `tidal runjob`
// pods to report job progress, logs, and status transitions. Mounted behind
// the shared-secret middleware in router.go.
type JobCallbacks struct {
	jobs    *jobs.Service
	presets *presets.Service
}

func NewJobCallbacks(jobsSvc *jobs.Service, presetSvc *presets.Service) *JobCallbacks {
	return &JobCallbacks{jobs: jobsSvc, presets: presetSvc}
}

type jobSpecResponse struct {
	JobID      uuid.UUID         `json:"jobId"`
	Preset     domain.PresetSpec `json:"preset"`
	SourcePath string            `json:"sourcePath"`
	OutputPath string            `json:"outputPath"`
	CachePath  string            `json:"cachePath,omitempty"`
}

// Spec returns the resolved JobSpec needed by `tidal runjob`. The pod has no
// DB access; it pulls everything it needs through this endpoint.
func (h *JobCallbacks) Spec(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	j, err := h.jobs.Get(c.Request().Context(), id)
	if err != nil {
		return mapJobErr(err)
	}
	p, err := h.presets.Get(c.Request().Context(), j.PresetID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, jobSpecResponse{
		JobID:      j.ID,
		Preset:     p.Spec,
		SourcePath: j.SourcePath,
		OutputPath: j.OutputPath,
		CachePath:  j.CachePath,
	})
}

type statusRequest struct {
	Status domain.JobStatus `json:"status"`
	Error  string           `json:"error,omitempty"`
}

func (h *JobCallbacks) Status(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var req statusRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	switch req.Status {
	case domain.JobRunning:
		h.jobs.Started(c.Request().Context(), id)
	case domain.JobSucceeded:
		h.jobs.Succeeded(c.Request().Context(), id)
	case domain.JobFailed:
		h.jobs.Failed(c.Request().Context(), id, errorMsg(req.Error))
	case domain.JobCancelled:
		h.jobs.Cancelled(c.Request().Context(), id)
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "invalid status")
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *JobCallbacks) Progress(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var p domain.FFmpegProgress
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	h.jobs.Progress(c.Request().Context(), id, p)
	return c.NoContent(http.StatusNoContent)
}

type logRequest struct {
	Lines []logLine `json:"lines"`
}
type logLine struct {
	Stream string `json:"stream"`
	Line   string `json:"line"`
}

func (h *JobCallbacks) Log(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var req logRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	for _, l := range req.Lines {
		h.jobs.AppendLog(c.Request().Context(), id, l.Stream, l.Line)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *JobCallbacks) Heartbeat(c echo.Context) error {
	// Reserved: future use for stuck-pod detection. No body needed.
	return c.NoContent(http.StatusNoContent)
}

type errMsg string

func (e errMsg) Error() string { return string(e) }
func errorMsg(s string) error  { return errMsg(s) }

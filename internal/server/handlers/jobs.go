package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/rustyguts/tidal/internal/domain"
	"github.com/rustyguts/tidal/internal/jobs"
	"github.com/rustyguts/tidal/internal/presets"
)

type Jobs struct {
	svc *jobs.Service
}

func NewJobs(svc *jobs.Service) *Jobs {
	return &Jobs{svc: svc}
}

type jobCreateRequest struct {
	PresetID       string `json:"presetId"`
	SourcePath     string `json:"sourcePath"`
	OutputPath     string `json:"outputPath,omitempty"`
	CachePath      string `json:"cachePath,omitempty"`
	SourceMovePath string `json:"sourceMovePath,omitempty"`
}

func (h *Jobs) List(c echo.Context) error {
	in := jobs.ListInput{
		Status: c.QueryParam("status"),
		Limit:  parseInt(c.QueryParam("limit"), 100),
	}
	if p := c.QueryParam("preset"); p != "" {
		id, err := uuid.Parse(p)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid preset")
		}
		in.PresetID = &id
	}
	out, err := h.svc.List(c.Request().Context(), in)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

func (h *Jobs) Get(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	j, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return mapJobErr(err)
	}
	return c.JSON(http.StatusOK, j)
}

func (h *Jobs) Create(c echo.Context) error {
	var req jobCreateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	pid, err := uuid.Parse(req.PresetID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid presetId")
	}
	if req.SourcePath == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "sourcePath required")
	}
	j, err := h.svc.Create(c.Request().Context(), jobs.CreateInput{
		PresetID:       pid,
		SourcePath:     req.SourcePath,
		OutputPath:     req.OutputPath,
		CachePath:      req.CachePath,
		SourceMovePath: req.SourceMovePath,
	})
	if err != nil {
		if errors.Is(err, presets.ErrNotFound) {
			return echo.NewHTTPError(http.StatusBadRequest, "preset not found")
		}
		return err
	}
	return c.JSON(http.StatusCreated, j)
}

func (h *Jobs) Cancel(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.Cancel(c.Request().Context(), id); err != nil {
		return mapJobErr(err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Jobs) Logs(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	from := int64(parseInt(c.QueryParam("from"), 0))
	limit := parseInt(c.QueryParam("limit"), 200)
	logs, err := h.svc.ListLogs(c.Request().Context(), id, from, limit)
	if err != nil {
		return mapJobErr(err)
	}
	return c.JSON(http.StatusOK, logs)
}

func parseInt(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func mapJobErr(err error) error {
	switch {
	case errors.Is(err, jobs.ErrNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	case errors.Is(err, jobs.ErrInvalid):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	default:
		return err
	}
}

// JobIDParam exposes the parsed :id from the route for SSE handlers.
func JobIDParam(c echo.Context) (domain.JobID, error) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return domain.JobID{}, echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	return id, nil
}

package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/rustyguts/tidal/internal/automations"
	"github.com/rustyguts/tidal/internal/domain"
)

type Automations struct {
	svc *automations.Service
	on  func()
}

// NewAutomations registers the automations REST handlers. The optional `notify`
// callback fires on any change so the scheduler can reconcile.
func NewAutomations(svc *automations.Service, notify func()) *Automations {
	return &Automations{svc: svc, on: notify}
}

func (h *Automations) notify() {
	if h.on != nil {
		h.on()
	}
}

func (h *Automations) List(c echo.Context) error {
	out, err := h.svc.List(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

func (h *Automations) Get(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	a, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return mapErr(err)
	}
	return c.JSON(http.StatusOK, a)
}

type automationCreate struct {
	Name           string `json:"name"`
	Enabled        bool   `json:"enabled"`
	WatchDir       string `json:"watchDir"`
	Glob           string `json:"glob"`
	PresetID       string `json:"presetId"`
	OutputDir      string `json:"outputDir"`
	ArchiveDir     string `json:"archiveDir"`
	PollIntervalMs int    `json:"pollIntervalMs"`
	DebounceMs     int    `json:"debounceMs"`
}

func (h *Automations) Create(c echo.Context) error {
	var req automationCreate
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	pid, err := uuid.Parse(req.PresetID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid presetId")
	}
	a, err := h.svc.Create(c.Request().Context(), automations.CreateInput{
		Name:           req.Name,
		Enabled:        req.Enabled,
		WatchDir:       req.WatchDir,
		Glob:           req.Glob,
		PresetID:       pid,
		OutputDir:      req.OutputDir,
		ArchiveDir:     req.ArchiveDir,
		PollIntervalMs: req.PollIntervalMs,
		DebounceMs:     req.DebounceMs,
	})
	if err != nil {
		return mapErr(err)
	}
	h.notify()
	return c.JSON(http.StatusCreated, a)
}

type automationUpdate struct {
	Name           *string `json:"name,omitempty"`
	Enabled        *bool   `json:"enabled,omitempty"`
	WatchDir       *string `json:"watchDir,omitempty"`
	Glob           *string `json:"glob,omitempty"`
	PresetID       *string `json:"presetId,omitempty"`
	OutputDir      *string `json:"outputDir,omitempty"`
	ArchiveDir     *string `json:"archiveDir,omitempty"`
	PollIntervalMs *int    `json:"pollIntervalMs,omitempty"`
	DebounceMs     *int    `json:"debounceMs,omitempty"`
}

func (h *Automations) Update(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var req automationUpdate
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	in := automations.UpdateInput{
		Name: req.Name, Enabled: req.Enabled,
		WatchDir: req.WatchDir, Glob: req.Glob,
		OutputDir: req.OutputDir, ArchiveDir: req.ArchiveDir,
		PollIntervalMs: req.PollIntervalMs, DebounceMs: req.DebounceMs,
	}
	if req.PresetID != nil {
		pid, err := uuid.Parse(*req.PresetID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid presetId")
		}
		in.PresetID = &pid
	}
	a, err := h.svc.Update(c.Request().Context(), id, in)
	if err != nil {
		return mapErr(err)
	}
	h.notify()
	return c.JSON(http.StatusOK, a)
}

func (h *Automations) Delete(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.Delete(c.Request().Context(), id); err != nil {
		return mapErr(err)
	}
	h.notify()
	return c.NoContent(http.StatusNoContent)
}

func (h *Automations) Enable(c echo.Context) error  { return h.toggle(c, true) }
func (h *Automations) Disable(c echo.Context) error { return h.toggle(c, false) }

func (h *Automations) toggle(c echo.Context, enabled bool) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.SetEnabled(c.Request().Context(), id, enabled); err != nil {
		return mapErr(err)
	}
	h.notify()
	return c.NoContent(http.StatusNoContent)
}

func (h *Automations) ListRuns(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	out, err := h.svc.ListRuns(c.Request().Context(), id, limit)
	if err != nil {
		return mapErr(err)
	}
	return c.JSON(http.StatusOK, out)
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, automations.ErrNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	case errors.Is(err, automations.ErrConflict):
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	default:
		return err
	}
}

// shared helper exposed for completeness — automations handlers don't currently
// need to look up by id beyond the existing routes.
var _ domain.AutomationID

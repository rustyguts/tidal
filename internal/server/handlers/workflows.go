package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/rustyguts/tidal/internal/domain"
	"github.com/rustyguts/tidal/internal/workflows"
)

type Workflows struct {
	svc *workflows.Service
}

func NewWorkflows(svc *workflows.Service) *Workflows { return &Workflows{svc: svc} }

type workflowCreate struct {
	Name              string          `json:"name"`
	Enabled           bool            `json:"enabled"`
	Trigger           domain.Trigger  `json:"trigger"`
	Actions           []domain.Action `json:"actions"`
	PollIntervalMs    int             `json:"pollIntervalMs"`
	StableThresholdMs int             `json:"stableThresholdMs"`
}

type workflowUpdate struct {
	Name              *string          `json:"name,omitempty"`
	Enabled           *bool            `json:"enabled,omitempty"`
	Trigger           *domain.Trigger  `json:"trigger,omitempty"`
	Actions           *[]domain.Action `json:"actions,omitempty"`
	PollIntervalMs    *int             `json:"pollIntervalMs,omitempty"`
	StableThresholdMs *int             `json:"stableThresholdMs,omitempty"`
}

func (h *Workflows) List(c echo.Context) error {
	out, err := h.svc.List(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

func (h *Workflows) Get(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	w, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return mapWfErr(err)
	}
	return c.JSON(http.StatusOK, w)
}

func (h *Workflows) Create(c echo.Context) error {
	var req workflowCreate
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	w, err := h.svc.Create(c.Request().Context(), workflows.CreateInput{
		Name:              req.Name,
		Enabled:           req.Enabled,
		Trigger:           req.Trigger,
		Actions:           req.Actions,
		PollIntervalMs:    req.PollIntervalMs,
		StableThresholdMs: req.StableThresholdMs,
	})
	if err != nil {
		return mapWfErr(err)
	}
	return c.JSON(http.StatusCreated, w)
}

func (h *Workflows) Update(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var req workflowUpdate
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	w, err := h.svc.Update(c.Request().Context(), id, workflows.UpdateInput{
		Name:              req.Name,
		Enabled:           req.Enabled,
		Trigger:           req.Trigger,
		Actions:           req.Actions,
		PollIntervalMs:    req.PollIntervalMs,
		StableThresholdMs: req.StableThresholdMs,
	})
	if err != nil {
		return mapWfErr(err)
	}
	return c.JSON(http.StatusOK, w)
}

func (h *Workflows) Delete(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.Delete(c.Request().Context(), id); err != nil {
		return mapWfErr(err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Workflows) Enable(c echo.Context) error  { return h.toggle(c, true) }
func (h *Workflows) Disable(c echo.Context) error { return h.toggle(c, false) }

func (h *Workflows) toggle(c echo.Context, enabled bool) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.SetEnabled(c.Request().Context(), id, enabled); err != nil {
		return mapWfErr(err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Workflows) ListRuns(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	out, err := h.svc.ListRuns(c.Request().Context(), id, limit)
	if err != nil {
		return mapWfErr(err)
	}
	return c.JSON(http.StatusOK, out)
}

func mapWfErr(err error) error {
	switch {
	case errors.Is(err, workflows.ErrNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	case errors.Is(err, workflows.ErrConflict):
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	default:
		return err
	}
}

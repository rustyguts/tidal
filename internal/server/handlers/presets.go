package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/rustyguts/tidal/internal/domain"
	"github.com/rustyguts/tidal/internal/presets"
)

type Presets struct {
	svc *presets.Service
}

func NewPresets(svc *presets.Service) *Presets {
	return &Presets{svc: svc}
}

type presetCreateRequest struct {
	Name           string            `json:"name"`
	Description    string            `json:"description"`
	Spec           domain.PresetSpec `json:"spec"`
	OutputPath     string            `json:"outputPath,omitempty"`
	CachePath      string            `json:"cachePath,omitempty"`
	SourceMovePath string            `json:"sourceMovePath,omitempty"`
}

type presetUpdateRequest struct {
	Name           *string            `json:"name,omitempty"`
	Description    *string            `json:"description,omitempty"`
	Spec           *domain.PresetSpec `json:"spec,omitempty"`
	OutputPath     *string            `json:"outputPath,omitempty"`
	CachePath      *string            `json:"cachePath,omitempty"`
	SourceMovePath *string            `json:"sourceMovePath,omitempty"`
}

type duplicateRequest struct {
	Name string `json:"name"`
}

func (h *Presets) List(c echo.Context) error {
	out, err := h.svc.List(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

func (h *Presets) Get(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	p, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, p)
}

func (h *Presets) Create(c echo.Context) error {
	var req presetCreateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	p, err := h.svc.Create(c.Request().Context(), presets.CreateInput{
		Name:           req.Name,
		Description:    req.Description,
		Spec:           req.Spec,
		OutputPath:     req.OutputPath,
		CachePath:      req.CachePath,
		SourceMovePath: req.SourceMovePath,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, p)
}

func (h *Presets) Update(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var req presetUpdateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	p, err := h.svc.Update(c.Request().Context(), id, presets.UpdateInput{
		Name:           req.Name,
		Description:    req.Description,
		Spec:           req.Spec,
		OutputPath:     req.OutputPath,
		CachePath:      req.CachePath,
		SourceMovePath: req.SourceMovePath,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, p)
}

func (h *Presets) Delete(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.Delete(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Presets) Duplicate(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var req duplicateRequest
	if err := c.Bind(&req); err != nil || req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name required")
	}
	p, err := h.svc.Duplicate(c.Request().Context(), id, req.Name)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, p)
}

type restoreResponse struct {
	Restored []string `json:"restored"`
}

func (h *Presets) RestoreDefaults(c echo.Context) error {
	planted, err := h.svc.RestoreDefaults(c.Request().Context())
	if err != nil {
		return err
	}
	if planted == nil {
		planted = []string{}
	}
	return c.JSON(http.StatusOK, restoreResponse{Restored: planted})
}

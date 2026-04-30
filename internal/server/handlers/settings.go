package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/rustyguts/tidal/internal/settings"
)

type Settings struct {
	svc *settings.Service
}

func NewSettings(svc *settings.Service) *Settings {
	return &Settings{svc: svc}
}

type settingsResponse struct {
	TranscodeConcurrency int `json:"transcodeConcurrency"`
}

type settingsUpdateRequest struct {
	TranscodeConcurrency *int `json:"transcodeConcurrency,omitempty"`
}

const (
	defaultTranscodeConcurrency = 4
	maxTranscodeConcurrency     = 64
)

func (h *Settings) Get(c echo.Context) error {
	n, err := h.svc.GetInt(c.Request().Context(), settings.KeyTranscodeConcurrency, defaultTranscodeConcurrency)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, settingsResponse{TranscodeConcurrency: n})
}

func (h *Settings) Update(c echo.Context) error {
	var req settingsUpdateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if req.TranscodeConcurrency != nil {
		n := *req.TranscodeConcurrency
		if n < 1 || n > maxTranscodeConcurrency {
			return echo.NewHTTPError(http.StatusBadRequest, "transcodeConcurrency out of range (1..64)")
		}
		if err := h.svc.SetInt(c.Request().Context(), settings.KeyTranscodeConcurrency, n); err != nil {
			return err
		}
	}
	return h.Get(c)
}

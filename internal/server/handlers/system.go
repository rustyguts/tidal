package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/rustyguts/tidal/internal/config"
	"github.com/rustyguts/tidal/internal/version"
)

type System struct {
	cfg config.Config
}

func NewSystem(cfg config.Config) *System {
	return &System{cfg: cfg}
}

type infoResponse struct {
	Version    version.Info `json:"version"`
	Env        string       `json:"env"`
	Dispatcher string       `json:"dispatcher"`
	MediaRoots []string     `json:"mediaRoots"`
}

func (s *System) Info(c echo.Context) error {
	return c.JSON(http.StatusOK, infoResponse{
		Version:    version.Get(),
		Env:        s.cfg.Env,
		Dispatcher: string(s.cfg.Dispatcher),
		MediaRoots: s.cfg.MediaRoots,
	})
}

func (s *System) Healthz(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (s *System) Readyz(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ready"})
}

package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"

	"github.com/hibiken/asynq"

	"github.com/rustyguts/tidal/internal/config"
	"github.com/rustyguts/tidal/internal/ffmpeg/catalog"
	"github.com/rustyguts/tidal/internal/jobs"
	"github.com/rustyguts/tidal/internal/presets"
	"github.com/rustyguts/tidal/internal/realtime"
	"github.com/rustyguts/tidal/internal/settings"
	"github.com/rustyguts/tidal/internal/workflows"
)

type Deps struct {
	Config    config.Config
	Pool      *pgxpool.Pool
	Presets   *presets.Service
	Catalog   *catalog.Catalog
	Jobs      *jobs.Service
	Workflows *workflows.Service
	Settings  *settings.Service
	Hub       *realtime.Hub
	RedisOpt  *asynq.RedisClientOpt // when set, asynqmon is mounted at /asynq
}

type Server struct {
	cfg config.Config
	e   *echo.Echo
	srv *http.Server
}

func New(deps Deps) *Server {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())
	e.Use(requestLogger())
	e.HTTPErrorHandler = httpErrorHandler

	mountRoutes(e, deps)

	srv := &http.Server{
		Addr:              deps.Config.HTTPAddr,
		Handler:           e,
		ReadHeaderTimeout: 10 * time.Second,
	}
	return &Server{cfg: deps.Config, e: e, srv: srv}
}

func (s *Server) Start() error {
	log.Info().Str("addr", s.cfg.HTTPAddr).Msg("server listening")
	if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Info().Msg("server shutdown")
	return s.srv.Shutdown(ctx)
}

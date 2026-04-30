package server

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/rustyguts/tidal/internal/presets"
)

func requestLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			latency := time.Since(start)
			req := c.Request()
			res := c.Response()
			ev := log.Info()
			if res.Status >= 500 || err != nil {
				ev = log.Error()
			}
			ev.Str("method", req.Method).
				Str("path", req.URL.Path).
				Int("status", res.Status).
				Dur("latency", latency).
				Str("rid", res.Header().Get(echo.HeaderXRequestID)).
				Err(err).
				Msg("http")
			return err
		}
	}
}

type errorBody struct {
	Error string `json:"error"`
}

func httpErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}
	status := http.StatusInternalServerError
	msg := "internal error"

	var he *echo.HTTPError
	switch {
	case errors.As(err, &he):
		status = he.Code
		if m, ok := he.Message.(string); ok {
			msg = m
		} else {
			msg = http.StatusText(status)
		}
	case errors.Is(err, presets.ErrNotFound):
		status = http.StatusNotFound
		msg = err.Error()
	case errors.Is(err, presets.ErrConflict):
		status = http.StatusConflict
		msg = err.Error()
	default:
		msg = err.Error()
	}

	if err := c.JSON(status, errorBody{Error: msg}); err != nil {
		log.Error().Err(err).Msg("write error response")
	}
}

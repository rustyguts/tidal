package server

import (
	"net/http"
	"strings"

	"github.com/hibiken/asynqmon"
	"github.com/labstack/echo/v4"
)

const asynqRoot = "/asynq"

// mountAsynqmon attaches the official asynqmon SPA + API at /asynq. The
// handler is a standard http.Handler; we wrap it with echo.WrapHandler.
func mountAsynqmon(e *echo.Echo, deps Deps) {
	if deps.RedisOpt == nil {
		return
	}
	mon := asynqmon.New(asynqmon.Options{
		RootPath:     asynqRoot,
		RedisConnOpt: *deps.RedisOpt,
	})
	// Map Echo route /asynq/* to the asynqmon handler. Echo strips the matched
	// prefix params for us; we need the full path including /asynq, so use the
	// raw request URL inside a wrapped HandlerFunc.
	h := echo.WrapHandler(asynqmonHandler(mon))
	e.Any(asynqRoot, redirectToTrailing)
	e.Any(asynqRoot+"/*", h)
}

func redirectToTrailing(c echo.Context) error {
	return c.Redirect(http.StatusFound, asynqRoot+"/")
}

// asynqmonHandler returns an http.Handler that ensures the asynqmon root
// matches its configured RootPath when calls come in via Echo's routing.
func asynqmonHandler(mon http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// asynqmon expects req.URL.Path to start with RootPath; Echo passes the
		// full path here, so this normally just works. Keep this wrapper as a
		// hook point for future header manipulation.
		if !strings.HasPrefix(r.URL.Path, asynqRoot) {
			r.URL.Path = asynqRoot + r.URL.Path
		}
		mon.ServeHTTP(w, r)
	})
}

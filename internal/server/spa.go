package server

import (
	"errors"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/rustyguts/tidal/ui"
)

// reservedPrefixes never resolve to the SPA fallback. Requests that miss the
// API/asynq routers should still 404 instead of returning index.html. Any
// path starting with one of these is owned by another router.
var reservedPrefixes = []string{
	"/api/",
	"/asynq",
	"/healthz",
	"/readyz",
	"/metrics",
}

// mountSPA registers the embedded Vue SPA at the root, with history-mode
// fallback to index.html for any client-side route. If no UI was embedded
// (default `go build` without `-tags=embed`) it serves a small placeholder
// instead.
func mountSPA(e *echo.Echo) {
	if !ui.Available || ui.Dist == nil {
		e.GET("/*", placeholderSPA)
		return
	}
	dist := ui.Dist
	e.GET("/*", spaHandler(dist))
}

func placeholderSPA(c echo.Context) error {
	for _, p := range reservedPrefixes {
		if strings.HasPrefix(c.Request().URL.Path, p) {
			return echo.NewHTTPError(http.StatusNotFound, "not found")
		}
	}
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	return c.String(http.StatusOK,
		`<!doctype html><meta charset=utf-8><title>Tidal</title>
<style>html,body{background:#020617;color:#cbd5e1;font:14px/1.5 system-ui;margin:0;padding:48px}</style>
<h1 style="color:#fff">Tidal</h1>
<p>The Vue SPA was not embedded into this binary. Build with <code>-tags=embed</code> after <code>make ui</code>, or run <code>make ui-dev</code> on :5173 with a Vite proxy to this server.</p>
<p>API is live at <a href="/api/system/info" style="color:#7dd3fc">/api/system/info</a>.</p>`)
}

func spaHandler(dist fs.FS) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := c.Request()
		urlPath := req.URL.Path
		for _, p := range reservedPrefixes {
			if strings.HasPrefix(urlPath, p) {
				return echo.NewHTTPError(http.StatusNotFound, "not found")
			}
		}

		clean := strings.TrimPrefix(path.Clean(urlPath), "/")
		if clean == "" {
			return serveFile(c, dist, "index.html")
		}
		f, err := dist.Open(clean)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				// SPA fallback: serve index.html so vue-router can handle the route
				return serveFile(c, dist, "index.html")
			}
			return err
		}
		defer f.Close()
		stat, err := f.Stat()
		if err != nil {
			return err
		}
		if stat.IsDir() {
			return serveFile(c, dist, "index.html")
		}
		return serveFile(c, dist, clean)
	}
}

func serveFile(c echo.Context, dist fs.FS, name string) error {
	f, err := dist.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	c.Response().Header().Set(echo.HeaderContentType, contentType(name))
	if _, err := io.Copy(c.Response().Writer, f); err != nil {
		return err
	}
	return nil
}

func contentType(name string) string {
	switch ext := strings.ToLower(path.Ext(name)); ext {
	case ".html":
		return echo.MIMETextHTMLCharsetUTF8
	case ".js", ".mjs":
		return "application/javascript; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".json":
		return echo.MIMEApplicationJSON
	case ".svg":
		return "image/svg+xml"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".ico":
		return "image/x-icon"
	case ".woff", ".woff2":
		return "font/" + ext[1:]
	default:
		return "application/octet-stream"
	}
}

package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/rustyguts/tidal/internal/realtime"
)

// streamEvents writes SSE events from the given hub topic until the client disconnects.
func streamEvents(c echo.Context, hub *realtime.Hub, topic string) error {
	w := c.Response()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	flusher, ok := w.Writer.(http.Flusher)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "streaming unsupported")
	}

	ch, cancel := hub.Subscribe(topic)
	defer cancel()

	// initial comment to flush headers immediately
	fmt.Fprint(w, ": ok\n\n")
	flusher.Flush()

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	ctx := c.Request().Context()
	enc := json.NewEncoder(w)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-heartbeat.C:
			fmt.Fprint(w, ": ping\n\n")
			flusher.Flush()
		case ev, ok := <-ch:
			if !ok {
				return nil
			}
			fmt.Fprintf(w, "event: %s\n", ev.Kind)
			fmt.Fprint(w, "data: ")
			if err := enc.Encode(ev); err != nil {
				return err
			}
			fmt.Fprint(w, "\n")
			flusher.Flush()
		}
	}
}

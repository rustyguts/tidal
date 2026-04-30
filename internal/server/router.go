package server

import (
	"github.com/labstack/echo/v4"

	"github.com/rustyguts/tidal/internal/domain"
	"github.com/rustyguts/tidal/internal/metrics"
	"github.com/rustyguts/tidal/internal/realtime"
	"github.com/rustyguts/tidal/internal/server/handlers"
)

func mountRoutes(e *echo.Echo, deps Deps) {
	api := e.Group("/api")

	sys := handlers.NewSystem(deps.Config)
	api.GET("/system/info", sys.Info)
	api.GET("/system/healthz", sys.Healthz)
	api.GET("/system/readyz", sys.Readyz)
	e.GET("/healthz", sys.Healthz)
	e.GET("/readyz", sys.Readyz)

	pres := handlers.NewPresets(deps.Presets)
	api.GET("/presets", pres.List)
	api.POST("/presets", pres.Create)
	// /presets/schema and /presets/preview are matched first and not parsed
	// as :id because Echo's router is path-prefix sensitive.
	presSchema := handlers.NewPresetSchema(deps.Catalog)
	presPreview := handlers.NewPresetPreview(deps.Catalog, domain.ValidateOpts{
		PermissiveRawExtras: deps.Config.PresetRawExtrasPermissive,
	})
	api.GET("/presets/schema", presSchema.Get)
	api.POST("/presets/preview", presPreview.Post)
	api.GET("/presets/:id", pres.Get)
	api.PATCH("/presets/:id", pres.Update)
	api.DELETE("/presets/:id", pres.Delete)
	api.POST("/presets/:id/duplicate", pres.Duplicate)
	api.POST("/presets/restore-defaults", pres.RestoreDefaults)

	jobsHandler := handlers.NewJobs(deps.Jobs)
	api.GET("/jobs", jobsHandler.List)
	api.POST("/jobs", jobsHandler.Create)
	api.GET("/jobs/:id", jobsHandler.Get)
	api.DELETE("/jobs/:id", jobsHandler.Cancel)
	api.GET("/jobs/:id/logs", jobsHandler.Logs)

	if deps.Workflows != nil {
		wf := handlers.NewWorkflows(deps.Workflows)
		api.GET("/workflows", wf.List)
		api.POST("/workflows", wf.Create)
		api.GET("/workflows/:id", wf.Get)
		api.PATCH("/workflows/:id", wf.Update)
		api.DELETE("/workflows/:id", wf.Delete)
		api.POST("/workflows/:id/enable", wf.Enable)
		api.POST("/workflows/:id/disable", wf.Disable)
		api.GET("/workflows/:id/runs", wf.ListRuns)
	}

	api.GET("/jobs/events", func(c echo.Context) error {
		return streamEvents(c, deps.Hub, realtime.TopicJobsFirehose)
	})
	api.GET("/jobs/:id/events", func(c echo.Context) error {
		id, err := handlers.JobIDParam(c)
		if err != nil {
			return err
		}
		return streamEvents(c, deps.Hub, "job:"+id.String())
	})

	e.GET("/metrics", echo.WrapHandler(metrics.Handler()))

	mountAsynqmon(e, deps)

	// Per-job pod callbacks. Unauthenticated by design — internal endpoints
	// stay on the cluster network. Tighten with NetworkPolicy if needed.
	callbacks := handlers.NewJobCallbacks(deps.Jobs, deps.Presets)
	internal := api.Group("/internal")
	internal.GET("/jobs/:id/spec", callbacks.Spec)
	internal.POST("/jobs/:id/status", callbacks.Status)
	internal.POST("/jobs/:id/progress", callbacks.Progress)
	internal.POST("/jobs/:id/log", callbacks.Log)
	internal.POST("/jobs/:id/heartbeat", callbacks.Heartbeat)

	mountSPA(e)
}

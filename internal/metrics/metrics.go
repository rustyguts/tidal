package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Registry is the prom Registry used by the server. Reset() rewires for tests.
var Registry = prometheus.NewRegistry()

var (
	JobsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "tidal_jobs_total",
		Help: "Number of transcode jobs by terminal status.",
	}, []string{"status"})

	JobDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "tidal_job_duration_seconds",
		Help:    "Wall-clock duration of completed transcode jobs.",
		Buckets: prometheus.ExponentialBuckets(1, 2, 12), // 1s … ~70m
	}, []string{"status"})

	AutomationScans = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "tidal_automation_scans_total",
		Help: "Automation scan ticks by outcome.",
	}, []string{"outcome"})
)

func init() {
	Registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		JobsTotal, JobDuration, AutomationScans,
	)
}

// Handler returns the prom HTTP handler.
func Handler() http.Handler {
	return promhttp.HandlerFor(Registry, promhttp.HandlerOpts{Registry: Registry})
}

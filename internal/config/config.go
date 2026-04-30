package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
)

const envPrefix = "TIDAL"

type DispatcherMode string

const (
	DispatcherLocal DispatcherMode = "local"
	DispatcherK8s   DispatcherMode = "k8s"
)

type Config struct {
	// Service identity / logging
	Env      string `envconfig:"ENV" default:"development"`
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`
	LogPretty bool  `envconfig:"LOG_PRETTY" default:"true"`

	// HTTP server
	HTTPAddr        string        `envconfig:"HTTP_ADDR" default:":8080"`
	ShutdownTimeout time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"15s"`

	// AutoMigrate runs `migrate up` on server startup. Convenient for
	// docker-compose dev; production should use the Helm pre-install Job.
	AutoMigrate bool `envconfig:"AUTO_MIGRATE" default:"false"`

	// Storage
	DBURL    string `envconfig:"DB_URL" default:"postgres://tidal:tidal@localhost:5432/tidal?sslmode=disable"`
	RedisURL string `envconfig:"REDIS_URL" default:"redis://localhost:6379/0"`

	// Worker / dispatcher
	Dispatcher DispatcherMode `envconfig:"DISPATCHER" default:"local"`
	WorkerConcurrency int      `envconfig:"WORKER_CONCURRENCY" default:"4"`

	// Job pod (k8s dispatcher only)
	JobImage          string `envconfig:"DISPATCHER_JOB_IMAGE"`
	JobNamespace      string `envconfig:"DISPATCHER_NAMESPACE" default:"default"`
	JobServiceAccount string `envconfig:"DISPATCHER_JOB_SERVICE_ACCOUNT"`
	JobMediaPVC       string `envconfig:"DISPATCHER_MEDIA_PVC"`
	JobMediaHostPath  string `envconfig:"DISPATCHER_MEDIA_HOSTPATH"`
	JobMediaMountPath string `envconfig:"DISPATCHER_MEDIA_MOUNT_PATH" default:"/media"`

	// Per-job pod resource hints (passed straight to corev1.ResourceRequirements).
	// Empty string = unset (k8s allows partial Requests/Limits).
	JobRequestCPU    string `envconfig:"DISPATCHER_JOB_REQUEST_CPU"`
	JobRequestMemory string `envconfig:"DISPATCHER_JOB_REQUEST_MEMORY"`
	JobLimitCPU      string `envconfig:"DISPATCHER_JOB_LIMIT_CPU"`
	JobLimitMemory   string `envconfig:"DISPATCHER_JOB_LIMIT_MEMORY"`

	ServerInternalURL string `envconfig:"SERVER_INTERNAL_URL" default:"http://tidal-server.default.svc.cluster.local:8080"`

	// Media browsing (server-side allow-list)
	MediaRoots StringSlice `envconfig:"MEDIA_ROOTS" default:"/media"`
}

// StringSlice supports comma-separated values via env.
type StringSlice []string

func (s *StringSlice) Decode(value string) error {
	if value == "" {
		*s = nil
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	*s = out
	return nil
}

func Load() (Config, error) {
	var c Config
	if err := envconfig.Process(envPrefix, &c); err != nil {
		return Config{}, fmt.Errorf("load config: %w", err)
	}
	if err := c.validate(); err != nil {
		return Config{}, err
	}
	return c, nil
}

func (c Config) validate() error {
	switch c.Dispatcher {
	case DispatcherLocal, DispatcherK8s:
	default:
		return fmt.Errorf("invalid TIDAL_DISPATCHER %q (want local|k8s)", c.Dispatcher)
	}
	if c.Dispatcher == DispatcherK8s {
		if c.JobImage == "" {
			return fmt.Errorf("TIDAL_DISPATCHER_JOB_IMAGE required when DISPATCHER=k8s")
		}
	}
	return nil
}

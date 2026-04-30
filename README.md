# Tidal

**Tidal** is an open-source video transcoding service. Queue, monitor, and manage ffmpeg jobs through a modern web UI or REST API. It scales from a single Docker Compose instance on a homelab to a full Kubernetes deployment.

## Features

- **Web UI** — Vue 3 SPA with real-time job progress via Server-Sent Events
- **REST API** — Echo-based HTTP API for programmatic control
- **Preset management** — Create, edit, and reuse encoding presets (H.264, H.265, AV1, and more)
- **Workflow automations** — Watch folders and automatically transcode files with configurable source/output paths
- **Two dispatcher modes** — Run ffmpeg in-process or dispatch to Kubernetes Jobs for isolation and scale
- **Helm chart** — Ready-to-deploy on any Kubernetes cluster

## Quick Start

### Docker Compose

```bash
make dev            # start server, worker, UI, postgres, redis
```

Then open `http://localhost:8090` in your browser.

### Local Development

```bash
make dev-db         # start postgres + redis only
make ui-dev         # start Vite dev server (proxies API to :8080)
make build          # build the Go binary
./bin/tidal server  # start the API server
./bin/tidal worker  # start the worker
```

### Kubernetes

```bash
helm install tidal deploy/helm/tidal
```

See the Helm chart for configuration options including dispatcher mode, resource limits, and storage.

## Architecture

Tidal is a single Go binary with multiple subcommands:

| Command | Role |
|---------|------|
| `tidal server` | Echo HTTP API + serves the Vue SPA |
| `tidal worker` | Job consumer and ffmpeg runner (local mode) or dispatcher (k8s mode) |
| `tidal runjob` | Entry point for per-job Kubernetes pods |
| `tidal migrate` | Database migrations |

**Tech stack:**

| Layer | Technology |
|-------|------------|
| Backend | Go 1.26, Echo, Cobra |
| Queue | asynq (Redis-backed) |
| Database | PostgreSQL |
| Frontend | Vue 3, Vite, Tailwind CSS v4 |
| Package manager | Bun |
| Deployment | Docker Compose, Helm (Kubernetes) |

## Project Layout

```
cmd/tidal/          # CLI entry point (cobra subcommands)
internal/           # Private Go packages
  server/           # HTTP handlers and SSE
  worker/           # Job consumer
  k8s/              # Kubernetes dispatcher
  jobs/             # Job service and domain logic
  workflows/        # Folder watcher and automation executor
  presets/          # Encoding preset management
  ffmpeg/           # ffmpeg command builder
  queue/            # asynq task definitions
  realtime/         # SSE hub
  db/               # Database layer + migrations
ui/                 # Vue 3 SPA
deploy/helm/tidal/  # Helm chart
```

## Commands

```bash
make dev            # Full stack via Docker Compose
make dev-db         # Postgres + Redis only
make dev-down       # Stop Docker Compose
make ui-dev         # Vite dev server
make build          # Build Go binary
make test           # Run Go tests
make test-ui        # Run UI tests
make test-all       # Run all tests
make migrate-up     # Apply database migrations
make migrate-down   # Roll back one migration
```

## License

[MIT](LICENSE)

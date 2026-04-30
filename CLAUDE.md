# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

Tidal is a video transcoding service: queue/monitor/manage ffmpeg jobs through a web UI or REST API. Single Go binary (`./tidal`) with cobra subcommands, plus a Vue 3 SPA. Production runs on k3s via the helm chart in `deploy/helm/tidal`; local dev runs via docker compose with the Vite UI proxy.

## Common commands

```bash
make dev-db                # postgres + redis only (compose services)
make dev                   # full stack via docker compose (server + worker + ui)
make dev-down              # stop docker compose
make dev-logs              # follow docker compose logs
make ui-dev                # bun-run Vite dev server (proxies /api,/asynq,/healthz to :8080)
make build                 # build ./bin/tidal binary
make test                  # go test ./... -race -count=1
make test-ui               # cd ui && bun run test (vitest + happy-dom)
make test-all              # both
make migrate-up            # run DB migrations against $DB_URL
make migrate-create NAME=x # new migration pair under internal/db/migrations
helm lint deploy/helm/tidal
helm template tidal deploy/helm/tidal --set image.tag=test
```

Run a single Go test:
```bash
go test ./internal/jobs -run TestServiceCancel -race -count=1 -v
```

`make test` does NOT start postgres/redis. Run `make dev-db` first. Tests also require `ffmpeg` on `$PATH`.

## Architecture

### Single binary, multiple modes

`cmd/tidal/main.go` wires cobra subcommands. Each is a separate process role:

| Subcommand | Role |
|------------|------|
| `server` | Echo HTTP/SSE API + serves the embedded Vue SPA at `/` |
| `worker` | asynq consumer. In `local` dispatcher mode runs ffmpeg in-process; in `k8s` mode it's the **dispatcher** that creates per-job K8s Jobs |
| `runjob` | Entrypoint baked into per-job K8s Job pods. Fetches its spec from the server, runs ffmpeg, posts progress/logs over the REST callback API |
| `migrate up\|down` | DB migrations (golang-migrate; SQL files embedded via `//go:embed migrations/*.sql`) |

`worker` and `dispatcher` are the same code path; the helm chart deploys them as separate Deployments named `tidal-worker` (only active when `dispatcher.mode=worker`) and `tidal-dispatcher` (active when `dispatcher.mode=batch`).

### Dispatcher modes

| Mode | `TIDAL_DISPATCHER` | Where ffmpeg runs | Wired in |
|------|--------------------|-------------------|----------|
| docker compose | `local` | `worker` container, in-process | `worker.NewLocalRunner` |
| Helm `worker` | `local` | `tidal-worker` Deployment, in-process | same |
| Helm `batch` (chart default) | `k8s` | Per-job pod via `tidal runjob` | `k8s.Dispatcher` |

In k8s mode the dispatcher creates a `batch/v1.Job` named `tidal-job-<jobID>`, watches it, and is the **authoritative writer for terminal status** (cancelled/failed) in the DB. `runjob` only posts `JobRunning`, progress, logs, and `JobSucceeded`. On pod SIGTERM (eviction, rolling deploy, dispatcher-initiated delete) `runjob` exits silently — pod termination is NOT user cancellation. See `internal/k8s/dispatcher.go` and `cmd/tidal/runjob.go`.

The dispatcher polls the DB via the `JobCoordinator` interface (`IsCancelRequested` / `Cancelled` / `Failed`). User cancel flow: server marks `status=cancelling` → dispatcher observes, deletes K8s Job, calls `Cancelled()`, returns nil so asynq doesn't retry.

### Job lifecycle

```
Create → JobQueued → asynq enqueue → worker.handleTranscode → runner.Run
  ├─ local: jobs.Service.Run → ffmpeg.Run → Started/Progress/Log → Succeeded|Failed|Cancelled
  └─ k8s:   k8s.Dispatcher.Run → Create batch/v1.Job → waitForCompletion (poll DB cancel + Job condition)
            └─ pod runs `tidal runjob` → callback API: Spec, Status(running), Progress, Log, Status(succeeded)
```

Server handlers split into user-facing routes (`/api/...`) and internal callback routes (`/api/internal/...`) consumed by `runjob`. See `internal/server/handlers/jobs_callbacks.go`.

### Real-time

Server-Sent Events via an in-memory `realtime.Hub` (`internal/realtime`). Topics are per-job (`job:<id>`) and global (jobs list, presets). Multi-replica deploys are stateless — no Redis fanout — so each browser connects to one server pod and receives events only from that pod's local activity. Events surface via `jobs.Service.publish*` and the `Hub.Publish` calls inside ffmpeg progress hooks.

### Workflows

`internal/workflows` (replaces the old "Automations" — see commit `3a0a999`). A `Watcher` polls media roots; matched files trigger an `Executor` that creates a job with the workflow's preset + move paths. Scheduling lives in the worker pod (NOT server) so multiple server replicas don't double-enqueue scans. Worker is a singleton; scaling >1 requires leader election.

### Helm chart + ArgoCD

The chart in `deploy/helm/tidal` is consumed by ArgoCD from the houston repo (`k3s/tidal/values.yaml` + `bootstrap/argocd/apps/tidal.yaml`).

Two non-obvious chart contracts to preserve when editing:

1. **`tidal-migrate` is an ArgoCD `Sync` hook**, not a tracked resource. Job specs are immutable, so leaving it as a regular resource caused argocd selfHeal to retry forever each time the image tag changed, taking out the dispatcher Deployment (and therefore in-flight transcode pods) on every retry. The hook annotation + `BeforeHookCreation` delete-policy means each sync deletes-then-creates.

2. **Dispatcher-created `tidal-job-*` Jobs must NOT carry the `argocd.argoproj.io/instance` (or `app.kubernetes.io/instance`) label** — that's how argocd identifies tracked resources. `internal/k8s/jobspec.go` deliberately uses `app.kubernetes.io/managed-by: tidal-dispatcher` instead. If argocd ever adopted these jobs it would prune them on the next sync.

The dispatcher Deployment uses `strategy: Recreate` (replicaCount=1; rolling update would briefly run two dispatchers and create duplicate K8s Jobs). This is fine as long as syncs only fire on real spec changes — the houston `Application` sets `ApplyOutOfSyncOnly=true,ServerSideApply=true,RespectIgnoreDifferences=true` to enforce that.

### UI

Vue 3 + Vite + Tailwind v4, in `ui/`. Bun is the package manager (NOT npm). Production builds compile with the `embed` build tag so `ui/embed_real.go` embeds `ui/dist/` via `//go:embed all:dist`. Without `embed`, the server returns 404 for `/`. Vite proxies `/api`, `/asynq`, `/healthz` to `VITE_API_PROXY` (default `http://localhost:8080`; in compose `http://server:8080`).

## Key environment variables

| Variable | Default | Notes |
|----------|---------|-------|
| `TIDAL_DB_URL` | `postgres://tidal:tidal@localhost:5432/tidal` | |
| `TIDAL_REDIS_URL` | `redis://localhost:6379/0` | |
| `TIDAL_DISPATCHER` | `local` | `local` or `k8s` |
| `TIDAL_AUTO_MIGRATE` | `false` | `true` runs migrations on server startup (compose only) |
| `TIDAL_MEDIA_ROOTS` | `/media` | Comma-separated allowed source paths |
| `TIDAL_HTTP_ADDR` | `:8080` | |
| `TIDAL_PRESET_RAW_EXTRAS_PERMISSIVE` | `false` | Allow unrecognised ffmpeg flags in presets |

K8s-dispatcher-only: `TIDAL_DISPATCHER_NAMESPACE`, `TIDAL_DISPATCHER_JOB_IMAGE`, `TIDAL_DISPATCHER_JOB_SERVICE_ACCOUNT`, `TIDAL_DISPATCHER_MEDIA_PVC` / `TIDAL_DISPATCHER_MEDIA_HOSTPATH`, `TIDAL_SERVER_INTERNAL_URL`, `TIDAL_DISPATCHER_JOB_REQUEST_CPU` / `_REQUEST_MEMORY` / `_LIMIT_CPU` / `_LIMIT_MEMORY`.

## Gotchas

- **Tests need live postgres/redis + ffmpeg on PATH.** No mocked DB.
- **`make gen` is wired but unused.** No `//go:generate` directives currently exist; `mockgen` install is for future use.
- **No `golangci-lint` config and no `sqlc.yaml`** — both intentionally removed. CI runs `go vet` + `go test`.
- **Backoff limit on transcode K8s Jobs is 0.** Pod failures don't auto-retry at the K8s level; asynq retries the task instead, and the dispatcher re-attaches via deterministic Job name.
- **TTL on transcode K8s Jobs is 600s.** Old Jobs auto-clean themselves 10 min after terminal state.

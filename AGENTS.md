# Tidal — Agent Notes

## Developer Commands

```bash
make dev-db          # postgres + redis only (docker compose services)
make dev             # full stack via docker compose (builds server + worker + ui)
make dev-down        # stop docker compose
make dev-logs        # follow docker compose logs
make dev-server      # run Go server binary (needs postgres + redis running)
make ui-dev          # bun-run Vite dev server (proxies /api to :8080)
make build           # build ./bin/tidal binary
make test            # go test ./... -race -count=1
make test-ui         # cd ui && bun run test (vitest + happy-dom)
make test-all        # both Go + UI tests
make migrate-up      # run DB migrations
make migrate-down    # rollback one migration
make migrate-create NAME=foo  # create new migration pair
make helm-lint       # helm lint + helm template validation
```

## Multi-Package Layout

| Area | Path | Entrypoint | Notes |
|------|------|------------|-------|
| Backend | `cmd/tidal/` | `main.go` → cobra subcommands | Single binary, multiple modes |
| Backend libs | `internal/` | — | All private; domain logic lives here |
| Frontend | `ui/` | `src/main.ts` | Vue 3 + Vite + Tailwind v4 + bun |
| Helm chart | `deploy/helm/tidal/` | — | Depends on Bitnami postgresql + redis |

Backend cobra subcommands:
- `server` — Echo REST API + serves Vue SPA
- `worker` — asynq consumer (local or k8s dispatch)
- `runjob` — K8s Job pod entrypoint (fetches spec from server, runs ffmpeg)
- `migrate up/down` — DB migrations

## Dev Environment Gotchas

**Package manager.** Use `bun` for the UI, not `npm`. `make ui-dev` runs `cd ui && bun install && bun run dev`. CI installs with `bun install` but the `ui/` job in `ci.yml` currently uses `npm ci` and `npm run build` — that path works because the `package.json` satisfies both.

**Vite proxy.** The dev server proxies `/api`, `/asynq`, and `/healthz` to `VITE_API_PROXY` (default `http://localhost:8080`). In docker compose the `ui` service sets `VITE_API_PROXY: http://server:8080`.

**Go testing requires postgres + redis on localhost.** `make test` does not start Docker services; run `make dev-db` first. Tests also expect `ffmpeg` on `$PATH`.

**Migrations.** DB migrations are embedded via `//go:embed migrations/*.sql` in `internal/db/migrate.go`. In local docker compose, `TIDAL_AUTO_MIGRATE=true` triggers `migrate up` on server startup. In production (Helm), a pre-install `Job` runs migrations instead.

**No active codegen.** `sqlc.yaml` and `.golangci.yml` were removed. The Makefile still has a `make gen` target referencing `mockgen`, but there are no `//go:generate` directives in the codebase yet.

**Build tag for UI embedding.** Production builds of the Go binary compile `ui/embed_real.go` (tag `embed`) to embed `ui/dist/` via `//go:embed all:dist`. Without the `embed` tag, the UI is not included and the server returns 404 for `/`.

## Docker Compose vs Helm Dispatch Modes

| Mode | Dispatch | `TIDAL_DISPATCHER` | Where ffmpeg runs |
|------|----------|--------------------|-------------------|
| docker compose | inline | `local` | `worker` container (always on) |
| Helm `batch` (default) | K8s Jobs | `k8s` | Per-job pod (`tidal runjob`) |
| Helm `worker` | inline | `local` | `worker` Deployment (always on) |

The Helm chart default is `dispatcher.mode: batch` (K8s Jobs). Switch to `worker` mode by passing `--set dispatcher.mode=worker`.

## CI / Testing

- `ci.yml` runs on `main` and `dev` branches and PRs. Jobs: `go` (vet, test, lint), `ui` (build), `helm` (lint + template validation).
- `test.yml` runs on every push and PR. Jobs: `test-go`, `test-ui`, `lint-go`.
- `publish.yml` builds + pushes Docker image to GHCR. Triggered on push to `main`/`dev` and releases.
- `helm-publish.yml` packages + pushes the Helm chart to `oci://ghcr.io/rustyguts/charts` on releases.

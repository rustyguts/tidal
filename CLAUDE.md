# Tidal

Tidal is an application that simplifies video encoding by allowing users to queue, monitor, manage, and automate video transcoding.

## Quick Start

Run the full stack with Docker Compose:

```bash
docker compose up --build
```

Then open http://localhost:5173 in your browser.

## Development

### Prerequisites

- Go 1.26+
- Bun 1.0+
- Docker & Docker Compose (for postgres + redis)

### Start dependencies

```bash
make dev-db
```

### Start backend server

```bash
make dev-server
```

### Start frontend (Vite dev server)

```bash
make ui-dev
```

The Vite dev server proxies API calls to the Go backend.

## Testing

### Backend

```bash
make test        # run all Go tests
make test-go     # alias
```

### Frontend

```bash
cd ui && bun run test       # run component tests
```

### All

```bash
make test-all
```

## Project Structure

```
cmd/tidal/          # CLI entrypoints (cobra)
internal/
  auth/             # callback secret middleware
  client/           # REST API client
  config/           # env-based config
  db/               # migrations, connection pool
  domain/           # models + validation
  ffmpeg/           # ffmpeg arg builder + execution
  jobs/             # job business logic + repository
  logging/          # zerolog setup
  presets/          # preset CRUD + seeding
  queue/            # asynq wrapper
  realtime/         # in-memory SSE hub
  server/           # Echo HTTP server + handlers + SSE
  version/          # build-time version info
  worker/           # asynq worker
ui/                 # Vue 3 SPA (bun + Vite + Tailwind)
```

## Architecture

- **Backend**: Go/Echo REST API with asynq (Redis) for job queueing
- **Worker**: asynq worker that runs ffmpeg transcodes locally
- **Database**: PostgreSQL for state (jobs, presets)
- **Queue**: Redis + asynq for task dispatch
- **Frontend**: Vue 3 SPA served by Vite dev server in development, or embedded into Go binary in production
- **Real-time**: Server-Sent Events via in-memory pub/sub hub

## Docker Compose Services

| Service | Image | Purpose |
|---------|-------|---------|
| postgres | postgres:18-alpine | Application database |
| redis | redis:8-alpine | Task queue + pub/sub |
| tidal-server | Dockerfile | REST API + SSE |
| tidal-worker | Dockerfile | asynq worker (local mode) |
| tidal-ui | oven/bun:1 | Vite dev server |

## Key Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `TIDAL_DB_URL` | postgres://tidal:tidal@localhost:5432/tidal | PostgreSQL connection |
| `TIDAL_REDIS_URL` | redis://localhost:6379/0 | Redis connection |
| `TIDAL_DISPATCHER` | local | `local` or `k8s` |
| `TIDAL_HTTP_ADDR` | :8080 | Server listen address |
| `TIDAL_MEDIA_ROOTS` | /media | Comma-separated media paths |

## CLI Commands

```bash
./tidal server      # start HTTP server
./tidal worker      # start asynq worker
./tidal migrate up  # run database migrations
./tidal job list    # list jobs
./tidal version     # show version
```

## License

MIT

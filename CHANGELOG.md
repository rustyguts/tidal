# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- Switched UI package manager from `npm`/`pnpm` to `bun`.
- Replaced `docker-compose.yml` with full-stack compose including `tidal-server`, `tidal-worker`, and `tidal-ui` services.
- `tidal-worker` now runs in its own container using `DISPATCHER=local` (no Kubernetes batch jobs in Docker Compose).
- `tidal-ui` service runs Vite dev server via `oven/bun:1` image with hot-reload.
- `vite.config.ts` now reads `VITE_API_PROXY` env var for dynamic backend target.
- Updated `Makefile` to use `bun` instead of `pnpm`.

### Added
- Created `Dockerfile` with multi-stage build (Go builder + alpine with ffmpeg).
- Added backend unit tests for `domain`, `config`, `version`, `realtime`, `auth`, and `ffmpeg` packages.
- Added frontend component tests for `StatusBadge`, `ProgressBar`, `Button`, and `EmptyState`.
- Added `vitest` + `@vue/test-utils` + `happy-dom` for Vue component testing.
- Added `.github/workflows/test.yml` CI workflow for Go + UI tests.
- Added `CLAUDE.md` with project overview, dev commands, and architecture.
- Added `CHANGELOG.md`.

### Removed
- Removed `.golangci.yml`.
- Removed `sqlc.yaml` and empty `internal/db/sqlcgen/` / `internal/db/queries/` directories.
- Removed `package-lock.json` in favor of `bun.lock`.

### Fixed
- Updated `.gitignore` to be appropriate for Go + Bun project (removed Python template).
- Updated `.dockerignore` to exclude build artifacts and node_modules.
- Fixed `TestHub_ConcurrentPublish` race condition in `realtime/hub_test.go`.
- Fixed `TestJobStatus_Terminal` to use correct `Job*` constants (not `JobStatus*`).

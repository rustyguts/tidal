.DEFAULT_GOAL := help

SHELL := /usr/bin/env bash
GO ?= go
BUN ?= bun

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -s -w \
	-X github.com/rustyguts/tidal/internal/version.Version=$(VERSION) \
	-X github.com/rustyguts/tidal/internal/version.Commit=$(COMMIT) \
	-X github.com/rustyguts/tidal/internal/version.Date=$(DATE)

BIN := bin/tidal

DB_URL ?= postgres://tidal:tidal@localhost:5432/tidal?sslmode=disable
REDIS_URL ?= redis://localhost:6379/0

## help: print this help
help:
	@awk 'BEGIN { FS = ":.*?## " } /^[a-zA-Z0-9_.-]+:.*?## / { printf "  %-22s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

## gen: run code generators (mockgen, etc.)
gen:
	$(GO) generate ./...

## deps: install Go dependencies
deps:
	$(GO) mod tidy

## build: build the tidal binary
build:
	CGO_ENABLED=0 $(GO) build -ldflags="$(LDFLAGS)" -o $(BIN) ./cmd/tidal

## dev-db: start postgres + redis via docker compose
dev-db:
	docker compose up -d postgres redis

## dev-down: stop docker compose services
dev-down:
	docker compose down

## dev: run full app via docker compose
dev:
	docker compose up --build -d

## dev-logs: follow logs from docker compose
dev-logs:
	docker compose logs -f

## ui-dev: start Vite dev server with bun
ui-dev:
	cd ui && $(BUN) install && $(BUN) run dev

## migrate-up: run all up migrations against $(DB_URL)
migrate-up:
	$(GO) run ./cmd/tidal migrate up --db-url="$(DB_URL)"

## migrate-down: roll back one migration
migrate-down:
	$(GO) run ./cmd/tidal migrate down --db-url="$(DB_URL)"

## migrate-create: create a new migration pair (use NAME=foo)
migrate-create:
	@test -n "$(NAME)" || (echo "NAME=foo required"; exit 1)
	migrate create -ext sql -dir internal/db/migrations -seq $(NAME)

## test: run all Go unit tests with race detector
test:
	$(GO) test ./... -race -count=1

## test-go: run Go tests (alias for test)
test-go:
	$(GO) test ./... -race -count=1

## test-ui: run Vue component tests
test-ui:
	cd ui && $(BUN) run test

## test-all: run Go + UI tests
test-all: test-go test-ui

## docker: build the container image
docker:
	docker build -t tidal:dev .

## clean: remove build artifacts
clean:
	rm -rf bin ui/dist coverage.* dist

## tools: install pinned dev tools for Go
tools:
	$(GO) install go.uber.org/mock/mockgen@latest

.PHONY: help deps gen build dev-db dev-down dev dev-logs ui-dev \
	migrate-up migrate-down migrate-create test test-go test-ui test-all \
	docker clean tools
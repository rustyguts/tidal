## syntax=docker/dockerfile:1.7

# ---- ui build stage ----
FROM oven/bun:1 AS ui
WORKDIR /ui
COPY ui/package.json ui/bun.lock ./
COPY ui/ ./
RUN bun install --frozen-lockfile && bun run build

# ---- go build stage ----
FROM golang:1.26-alpine AS builder
RUN apk add --no-cache git
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=ui /ui/dist ./ui/dist
RUN CGO_ENABLED=0 go build -tags=embed -o /usr/local/bin/tidal ./cmd/tidal

# ---- runtime ----
FROM alpine:3.21
RUN apk add --no-cache ffmpeg ca-certificates tzdata
COPY --from=builder /usr/local/bin/tidal /usr/local/bin/tidal
EXPOSE 8080
ENTRYPOINT ["tidal"]
CMD ["server"]
